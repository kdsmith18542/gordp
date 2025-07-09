package gordp

import (
	"bytes"
	"fmt"
	"time"

	"github.com/GoFeGroup/gordp/core"
	"github.com/GoFeGroup/gordp/glog"
	"github.com/GoFeGroup/gordp/proto/bitmap"
	"github.com/GoFeGroup/gordp/proto/drdynvc"
	"github.com/GoFeGroup/gordp/proto/mcs"
	"github.com/GoFeGroup/gordp/proto/t128"
	"github.com/GoFeGroup/gordp/proto/virtualchannel"
)

type Option struct {
	Addr     string
	UserName string
	Password string

	ConnectTimeout time.Duration
}

type Processor interface {
	ProcessBitmap(*bitmap.Option, *bitmap.BitMap)
}

type Client struct {
	option Option

	//conn   net.Conn // TCP连接
	stream *core.Stream

	// from negotiation
	selectProtocol uint32 // 协商RDP协议，0:rdp, 1:ssl, 2:hybrid
	userId         uint16
	shareId        uint32
	serverVersion  uint32 // 服务端RDP版本号

	// input state
	modifierKeys t128.ModifierKey

	// Virtual channel support
	vcManager  *virtualchannel.VirtualChannelManager
	vcHandlers map[string]virtualchannel.VirtualChannelHandler

	// Dynamic virtual channel support
	dvcManager *drdynvc.DynamicVirtualChannelManager

	// Bitmap cache and compression support
	bitmapCacheManager *t128.BitmapCacheManager

	// Offscreen bitmap support
	offscreenBitmapManager *t128.OffscreenBitmapManager
}

func NewClient(opt *Option) *Client {
	c := &Client{
		option: Option{
			Addr:           opt.Addr,
			UserName:       opt.UserName,
			Password:       opt.Password,
			ConnectTimeout: opt.ConnectTimeout,
		},
	}
	if c.option.ConnectTimeout == 0 {
		c.option.ConnectTimeout = 5 * time.Second
	}
	c.vcManager = virtualchannel.NewVirtualChannelManager()
	c.vcHandlers = make(map[string]virtualchannel.VirtualChannelHandler)
	c.dvcManager = drdynvc.NewDynamicVirtualChannelManager()
	c.bitmapCacheManager = t128.NewBitmapCacheManager()
	c.offscreenBitmapManager = t128.NewOffscreenBitmapManager(7680, 100) // Default values

	// Register default virtual channels
	_ = c.vcManager.RegisterChannel(&virtualchannel.VirtualChannel{
		ID:    1,
		Name:  virtualchannel.CHANNEL_NAME_CLIPRDR,
		Flags: virtualchannel.CHANNEL_FLAG_FIRST | virtualchannel.CHANNEL_FLAG_LAST,
	})
	_ = c.vcManager.RegisterChannel(&virtualchannel.VirtualChannel{
		ID:    2,
		Name:  virtualchannel.CHANNEL_NAME_RDPSND,
		Flags: virtualchannel.CHANNEL_FLAG_FIRST | virtualchannel.CHANNEL_FLAG_LAST,
	})
	_ = c.vcManager.RegisterChannel(&virtualchannel.VirtualChannel{
		ID:    3,
		Name:  virtualchannel.CHANNEL_NAME_DRDYNVC,
		Flags: virtualchannel.CHANNEL_FLAG_FIRST | virtualchannel.CHANNEL_FLAG_LAST,
	})

	return c
}

//func (c *Client) tcpConnect() {
//	conn, err := net.DialTimeout("tcp", c.option.Addr, c.option.ConnectTimeout)
//	core.ThrowError(err)
//	c.conn = conn
//}

// Connect
// https://www.cyberark.com/resources/threat-research-blog/explain-like-i-m-5-remote-desktop-protocol-rdp
func (c *Client) Connect() error {
	return core.Try(func() {
		c.stream = core.NewStream(c.option.Addr, c.option.ConnectTimeout)
		c.negotiation()
		c.basicSettingsExchange()
		c.channelConnect()
		c.sendClientInfo()
		c.readLicensing()
		c.capabilitiesExchange()
		c.sendClientFinalization()
	})
}

func (c *Client) Close() {
	c.stream.Close()
}

// GetBitmapCacheStats returns statistics about the bitmap cache
func (c *Client) GetBitmapCacheStats() map[string]interface{} {
	return c.bitmapCacheManager.GetCacheStats()
}

// ClearBitmapCache clears all bitmap caches
func (c *Client) ClearBitmapCache() {
	c.bitmapCacheManager.ClearCache()
}

func (c *Client) Run(processor Processor) error {
	return core.Try(func() {
		for {
			pdu := c.readPdu()
			switch p := pdu.(type) {
			case *t128.TsFpUpdatePDU:
				if p.Length == 0 {
					break
				}
				switch pp := p.PDU.(type) {
				case *t128.TsFpUpdateBitmap:
					for _, v := range pp.Rectangles {
						// Process bitmap through cache manager for optimization
						optimizedBitmap, cached := c.bitmapCacheManager.OptimizeBitmapData(&v)

						option := &bitmap.Option{
							Top:         int(optimizedBitmap.DestTop),  // for position
							Left:        int(optimizedBitmap.DestLeft), // for position
							Width:       int(optimizedBitmap.Width),
							Height:      int(optimizedBitmap.Height),
							BitPerPixel: int(optimizedBitmap.BitsPerPixel),
							Data:        optimizedBitmap.BitmapDataStream,
						}

						if cached {
							glog.Debugf("Using cached bitmap: %dx%d", option.Width, option.Height)
						}

						if optimizedBitmap.BitsPerPixel == 32 {
							processor.ProcessBitmap(option, bitmap.NewBitMapFromRDP6(option))
						} else {
							processor.ProcessBitmap(option, bitmap.NewBitmapFromRLE(option))
						}
					}
				case *t128.TsFpUpdateCachedBitmap:
					for _, v := range pp.Rectangles {
						glog.Debugf("Cached bitmap update: cache=%d, index=%d, key=%08X%08X",
							v.CacheId, v.CacheIndex, v.Key1, v.Key2)
						// TODO: Retrieve and display cached bitmap from cache manager
					}
				case *t128.TsFpUpdateSurfaceCommands:
					for _, cmd := range pp.Commands {
						switch sc := cmd.(type) {
						case *t128.TsSetSurfaceBitsCommand:
							option := &bitmap.Option{
								Top:         int(sc.DestTop),
								Left:        int(sc.DestLeft),
								Width:       int(sc.BitmapData.Width),
								Height:      int(sc.BitmapData.Height),
								BitPerPixel: int(sc.BitmapData.BitsPerPixel),
								Data:        sc.BitmapData.BitmapDataStream,
							}
							if sc.BitmapData.BitsPerPixel == 32 {
								processor.ProcessBitmap(option, bitmap.NewBitMapFromRDP6(option))
							} else {
								processor.ProcessBitmap(option, bitmap.NewBitmapFromRLE(option))
							}
						case *t128.TsCreateSurfaceCommand:
							c.offscreenBitmapManager.ProcessOffscreenBitmap(&t128.TsOffscreenBitmapData{
								CacheId:    0, // For now, single cache
								CacheIndex: sc.SurfaceId,
								Width:      sc.Width,
								Height:     sc.Height,
								Bpp:        uint16(sc.PixelFormat),
								Data:       sc.SurfaceData,
							})
							glog.Debugf("CreateSurface: ID=%d, %dx%d", sc.SurfaceId, sc.Width, sc.Height)
						case *t128.TsDeleteSurfaceCommand:
							c.offscreenBitmapManager.RemoveOffscreenBitmap(sc.SurfaceId)
							glog.Debugf("DeleteSurface: ID=%d", sc.SurfaceId)
						default:
							glog.Debugf("Unhandled surface command: %T", sc)
						}
					}
				default:
					glog.Debugf("pdutype2: %T", pp)
				}
			default:
				// Attempt to process as a virtual channel packet
				c.tryHandleVirtualChannelPDU(pdu)
			}
		}
	})
}

// tryHandleVirtualChannelPDU attempts to parse and dispatch a virtual channel packet
func (c *Client) tryHandleVirtualChannelPDU(pdu interface{}) {
	if pdu == nil {
		return
	}
	// Only handle PDUs that are []byte or have a Data field
	var data []byte
	switch v := pdu.(type) {
	case []byte:
		data = v
	case interface{ Data() []byte }:
		data = v.Data()
	default:
		return
	}
	if len(data) < 10 { // Minimum size for a virtual channel packet header
		return
	}
	packet, err := virtualchannel.ReadVirtualChannelPacket(bytes.NewReader(data))
	if err != nil {
		return
	}
	// Find handler by channel name
	ch, ok := c.vcManager.GetChannel(packet.ChannelID)
	if !ok {
		return
	}
	handler, ok := c.vcHandlers[ch.Name]
	if !ok {
		handler = virtualchannel.NewDefaultVirtualChannelHandler(c.vcManager)
	}
	_ = handler.HandleData(packet.ChannelID, packet.Data)
}

// SendVirtualChannelData sends data on a named virtual channel
func (c *Client) SendVirtualChannelData(channelName string, data []byte, flags uint32) error {
	ch, ok := c.vcManager.GetChannelByName(channelName)
	if !ok {
		return fmt.Errorf("unknown virtual channel: %s", channelName)
	}
	packet := &virtualchannel.VirtualChannelPacket{
		Length:    uint32(len(data)),
		Flags:     flags,
		ChannelID: ch.ID,
		Data:      data,
	}
	serialized := packet.Serialize()
	mcsReq := mcs.NewSendDataRequest(c.userId, ch.ID)
	_, err := c.stream.Write(mcsReq.Serialize(serialized))
	return err
}

// Add helper to VirtualChannelManager to get channel by name

func (c *Client) dispatchVirtualChannelData(packet *virtualchannel.VirtualChannelPacket) error {
	// Check if this is a dynamic virtual channel
	if packet.ChannelID == 3 { // drdynvc channel ID
		return c.handleDynamicVirtualChannel(packet.Data)
	}

	// Handle regular virtual channels
	handler, exists := c.vcHandlers[fmt.Sprintf("channel_%d", packet.ChannelID)]
	if !exists {
		// Use default handler
		handler = virtualchannel.NewDefaultVirtualChannelHandler(c.vcManager)
	}

	return handler.HandleData(packet.ChannelID, packet.Data)
}

// handleDynamicVirtualChannel handles dynamic virtual channel messages
func (c *Client) handleDynamicVirtualChannel(data []byte) error {
	msg, err := drdynvc.ReadDynamicVirtualChannelMessage(bytes.NewReader(data))
	if err != nil {
		return fmt.Errorf("failed to read dynamic virtual channel message: %w", err)
	}

	switch msg.MessageType {
	case drdynvc.DVCCREATE_REQ:
		return c.handleCreateRequest(msg.Data)
	case drdynvc.DVCCREATE_RSP:
		return c.handleCreateResponse(msg.Data)
	case drdynvc.DVCOPEN_REQ:
		return c.handleOpenRequest(msg.Data)
	case drdynvc.DVCOPEN_RSP:
		return c.handleOpenResponse(msg.Data)
	case drdynvc.DVCCLOSE_REQ:
		return c.handleCloseRequest(msg.Data)
	case drdynvc.DVCCLOSE_RSP:
		return c.handleCloseResponse(msg.Data)
	case drdynvc.DVCDATA_FIRST, drdynvc.DVCDATA, drdynvc.DVCDATA_LAST, drdynvc.DVCDATA_FIRST_LAST:
		return c.handleDataMessage(msg.Data)
	default:
		glog.Debugf("Unknown dynamic virtual channel message type: 0x%02x", msg.MessageType)
		return nil
	}
}

// handleCreateRequest handles a dynamic virtual channel create request
func (c *Client) handleCreateRequest(data []byte) error {
	req, err := drdynvc.ParseCreateRequest(data)
	if err != nil {
		return fmt.Errorf("failed to parse create request: %w", err)
	}

	glog.Debugf("Dynamic virtual channel create request: %s (ID: %d)", req.ChannelName, req.ChannelId)

	// Register the channel
	err = c.dvcManager.RegisterChannelWithID(req.ChannelId, req.ChannelName, drdynvc.NewDefaultDynamicVirtualChannelHandler())
	if err != nil {
		glog.Errorf("Failed to register dynamic virtual channel: %v", err)
	}

	// Send create response
	resp := &drdynvc.CreateResponse{
		RequestId: req.RequestId,
		ChannelId: req.ChannelId,
		Status:    drdynvc.DVCCREATE_SUCCESS,
	}

	respData := resp.Serialize()
	dvcMsg := &drdynvc.DynamicVirtualChannelMessage{
		MessageType: drdynvc.DVCCREATE_RSP,
		Data:        respData,
	}

	return c.SendVirtualChannelData("drdynvc", dvcMsg.Serialize(), 0)
}

// handleCreateResponse handles a dynamic virtual channel create response
func (c *Client) handleCreateResponse(data []byte) error {
	resp, err := drdynvc.ParseCreateResponse(data)
	if err != nil {
		return fmt.Errorf("failed to parse create response: %w", err)
	}

	glog.Debugf("Dynamic virtual channel create response: ID: %d, Status: %d", resp.ChannelId, resp.Status)

	if resp.Status == drdynvc.DVCCREATE_SUCCESS {
		// Channel created successfully, send open request
		openReq := &drdynvc.OpenRequest{
			RequestId: resp.RequestId,
			ChannelId: resp.ChannelId,
		}

		openReqData := openReq.Serialize()
		dvcMsg := &drdynvc.DynamicVirtualChannelMessage{
			MessageType: drdynvc.DVCOPEN_REQ,
			Data:        openReqData,
		}

		return c.SendVirtualChannelData("drdynvc", dvcMsg.Serialize(), 0)
	}

	return nil
}

// handleOpenRequest handles a dynamic virtual channel open request
func (c *Client) handleOpenRequest(data []byte) error {
	req, err := drdynvc.ParseOpenRequest(data)
	if err != nil {
		return fmt.Errorf("failed to parse open request: %w", err)
	}

	glog.Debugf("Dynamic virtual channel open request: ID: %d", req.ChannelId)

	// Send open response
	resp := &drdynvc.OpenResponse{
		RequestId: req.RequestId,
		ChannelId: req.ChannelId,
		Status:    drdynvc.DVCOPEN_SUCCESS,
	}

	respData := resp.Serialize()
	dvcMsg := &drdynvc.DynamicVirtualChannelMessage{
		MessageType: drdynvc.DVCOPEN_RSP,
		Data:        respData,
	}

	return c.SendVirtualChannelData("drdynvc", dvcMsg.Serialize(), 0)
}

// handleOpenResponse handles a dynamic virtual channel open response
func (c *Client) handleOpenResponse(data []byte) error {
	resp, err := drdynvc.ParseOpenResponse(data)
	if err != nil {
		return fmt.Errorf("failed to parse open response: %w", err)
	}

	glog.Debugf("Dynamic virtual channel open response: ID: %d, Status: %d", resp.ChannelId, resp.Status)

	if resp.Status == drdynvc.DVCOPEN_SUCCESS {
		// Channel opened successfully
		channel, exists := c.dvcManager.GetChannel(resp.ChannelId)
		if exists && channel.Handler != nil {
			return channel.Handler.OnChannelOpened(resp.ChannelId)
		}
	}

	return nil
}

// handleCloseRequest handles a dynamic virtual channel close request
func (c *Client) handleCloseRequest(data []byte) error {
	req, err := drdynvc.ParseCloseRequest(data)
	if err != nil {
		return fmt.Errorf("failed to parse close request: %w", err)
	}

	glog.Debugf("Dynamic virtual channel close request: ID: %d", req.ChannelId)

	// Send close response
	resp := &drdynvc.CloseResponse{
		RequestId: req.RequestId,
		ChannelId: req.ChannelId,
		Status:    drdynvc.DVCCLOSE_SUCCESS,
	}

	respData := resp.Serialize()
	dvcMsg := &drdynvc.DynamicVirtualChannelMessage{
		MessageType: drdynvc.DVCCLOSE_RSP,
		Data:        respData,
	}

	return c.SendVirtualChannelData("drdynvc", dvcMsg.Serialize(), 0)
}

// handleCloseResponse handles a dynamic virtual channel close response
func (c *Client) handleCloseResponse(data []byte) error {
	resp, err := drdynvc.ParseCloseResponse(data)
	if err != nil {
		return fmt.Errorf("failed to parse close response: %w", err)
	}

	glog.Debugf("Dynamic virtual channel close response: ID: %d, Status: %d", resp.ChannelId, resp.Status)

	if resp.Status == drdynvc.DVCCLOSE_SUCCESS {
		// Channel closed successfully
		channel, exists := c.dvcManager.GetChannel(resp.ChannelId)
		if exists && channel.Handler != nil {
			return channel.Handler.OnChannelClosed(resp.ChannelId)
		}
	}

	return nil
}

// handleDataMessage handles a dynamic virtual channel data message
func (c *Client) handleDataMessage(data []byte) error {
	msg, err := drdynvc.ParseDataMessage(data)
	if err != nil {
		return fmt.Errorf("failed to parse data message: %w", err)
	}

	glog.Debugf("Dynamic virtual channel data message: ID: %d, %d bytes", msg.ChannelId, len(msg.Data))

	// Forward data to channel handler
	channel, exists := c.dvcManager.GetChannel(msg.ChannelId)
	if exists && channel.Handler != nil {
		return channel.Handler.OnDataReceived(msg.ChannelId, msg.Data)
	}

	return nil
}
