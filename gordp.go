// Package gordp provides a production-grade implementation of the Remote Desktop Protocol (RDP) client in Go.
// This library supports full RDP protocol features including input handling, clipboard integration,
// audio redirection, device redirection, and multi-monitor support.
//
// Example usage:
//
//	client := gordp.NewClient(&gordp.Option{
//		Addr:     "192.168.1.100:3389",
//		UserName: "username",
//		Password: "password",
//	})
//	err := client.Connect()
//	if err != nil {
//		log.Fatal(err)
//	}
//	defer client.Close()
//
//	processor := &MyProcessor{}
//	err = client.Run(processor)
package gordp

import (
	"bytes"
	"context"
	"fmt"
	"time"

	"github.com/kdsmith18542/gordp/core"
	"github.com/kdsmith18542/gordp/glog"
	"github.com/kdsmith18542/gordp/proto/bitmap"
	"github.com/kdsmith18542/gordp/proto/clipboard"
	"github.com/kdsmith18542/gordp/proto/device"
	"github.com/kdsmith18542/gordp/proto/drdynvc"
	"github.com/kdsmith18542/gordp/proto/mcs"
	"github.com/kdsmith18542/gordp/proto/t128"
	"github.com/kdsmith18542/gordp/proto/virtualchannel"
)

type Option struct {
	Addr     string
	UserName string
	Password string

	ConnectTimeout time.Duration

	// Multi-monitor configuration (optional)
	Monitors []mcs.MonitorLayout
}

type Processor interface {
	ProcessBitmap(*bitmap.Option, *bitmap.BitMap)
}

type Client struct {
	option Option

	// Context for cancellation and timeout
	ctx    context.Context
	cancel context.CancelFunc

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

	// Dynamic virtual channel custom handlers
	dvcHandlers map[string]drdynvc.DynamicVirtualChannelHandler

	// Bitmap cache and compression support
	bitmapCacheManager *t128.BitmapCacheManager

	// Offscreen bitmap support
	offscreenBitmapManager *t128.OffscreenBitmapManager

	clipboardManager *clipboard.ClipboardManager

	// Device redirection support
	deviceManager *device.DeviceManager

	// Multi-monitor configuration
	monitors []mcs.MonitorLayout
}

func NewClient(opt *Option) *Client {
	ctx, cancel := context.WithCancel(context.Background())
	c := &Client{
		option: Option{
			Addr:           opt.Addr,
			UserName:       opt.UserName,
			Password:       opt.Password,
			ConnectTimeout: opt.ConnectTimeout,
			Monitors:       opt.Monitors,
		},
		ctx:      ctx,
		cancel:   cancel,
		monitors: opt.Monitors,
	}
	if c.option.ConnectTimeout == 0 {
		c.option.ConnectTimeout = 5 * time.Second
	}
	c.vcManager = virtualchannel.NewVirtualChannelManager()
	c.vcHandlers = make(map[string]virtualchannel.VirtualChannelHandler)
	c.dvcManager = drdynvc.NewDynamicVirtualChannelManager()
	c.dvcHandlers = make(map[string]drdynvc.DynamicVirtualChannelHandler)
	c.bitmapCacheManager = t128.NewBitmapCacheManager()
	c.offscreenBitmapManager = t128.NewOffscreenBitmapManager(7680, 100) // Default values
	c.clipboardManager = clipboard.NewClipboardManager(nil)
	c.deviceManager = device.NewDeviceManager(nil)

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
	_ = c.vcManager.RegisterChannel(&virtualchannel.VirtualChannel{
		ID:    4,
		Name:  virtualchannel.CHANNEL_NAME_RDPDR,
		Flags: virtualchannel.CHANNEL_FLAG_FIRST | virtualchannel.CHANNEL_FLAG_LAST,
	})

	return c
}

// NewClientWithContext creates a new client with a custom context
func NewClientWithContext(ctx context.Context, opt *Option) *Client {
	ctx, cancel := context.WithCancel(ctx)
	c := &Client{
		option: Option{
			Addr:           opt.Addr,
			UserName:       opt.UserName,
			Password:       opt.Password,
			ConnectTimeout: opt.ConnectTimeout,
			Monitors:       opt.Monitors,
		},
		ctx:      ctx,
		cancel:   cancel,
		monitors: opt.Monitors,
	}
	if c.option.ConnectTimeout == 0 {
		c.option.ConnectTimeout = 5 * time.Second
	}
	c.vcManager = virtualchannel.NewVirtualChannelManager()
	c.vcHandlers = make(map[string]virtualchannel.VirtualChannelHandler)
	c.dvcManager = drdynvc.NewDynamicVirtualChannelManager()
	c.dvcHandlers = make(map[string]drdynvc.DynamicVirtualChannelHandler)
	c.bitmapCacheManager = t128.NewBitmapCacheManager()
	c.offscreenBitmapManager = t128.NewOffscreenBitmapManager(7680, 100) // Default values
	c.clipboardManager = clipboard.NewClipboardManager(nil)
	c.deviceManager = device.NewDeviceManager(nil)

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
	_ = c.vcManager.RegisterChannel(&virtualchannel.VirtualChannel{
		ID:    4,
		Name:  virtualchannel.CHANNEL_NAME_RDPDR,
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
		// Check if context is cancelled
		select {
		case <-c.ctx.Done():
			core.ThrowError(c.ctx.Err())
		default:
		}

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

// ConnectWithContext connects with a custom context
func (c *Client) ConnectWithContext(ctx context.Context) error {
	return core.Try(func() {
		// Check if context is cancelled
		select {
		case <-ctx.Done():
			core.ThrowError(ctx.Err())
		default:
		}

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
	c.cancel() // Cancel the context
	c.stream.Close()
}

// Context returns the client's context
func (c *Client) Context() context.Context {
	return c.ctx
}

// Done returns a channel that's closed when the client is cancelled
func (c *Client) Done() <-chan struct{} {
	return c.ctx.Done()
}

// Cancel cancels the client context
func (c *Client) Cancel() {
	c.cancel()
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
			// Check if context is cancelled
			select {
			case <-c.ctx.Done():
				core.ThrowError(c.ctx.Err())
			default:
			}

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

						// Retrieve cached bitmap from cache manager
						cachedBitmap := c.bitmapCacheManager.GetCachedBitmap(uint16(v.CacheId), v.CacheIndex, v.Key1, v.Key2)
						if cachedBitmap != nil {
							option := &bitmap.Option{
								Top:         int(v.DestTop),
								Left:        int(v.DestLeft),
								Width:       int(cachedBitmap.Width),
								Height:      int(cachedBitmap.Height),
								BitPerPixel: int(cachedBitmap.BitsPerPixel),
								Data:        cachedBitmap.BitmapDataStream,
							}
							processor.ProcessBitmap(option, bitmap.NewBitmapFromRLE(option))
							glog.Debugf("Retrieved cached bitmap: %dx%d", option.Width, option.Height)
						} else {
							glog.Warnf("Cached bitmap not found: cache=%d, index=%d", v.CacheId, v.CacheIndex)
						}
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

// RunWithContext runs the RDP session with a custom context
func (c *Client) RunWithContext(ctx context.Context, processor Processor) error {
	return core.Try(func() {
		for {
			// Check if context is cancelled
			select {
			case <-ctx.Done():
				core.ThrowError(ctx.Err())
			case <-c.ctx.Done():
				core.ThrowError(c.ctx.Err())
			default:
			}

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

						// Retrieve cached bitmap from cache manager
						cachedBitmap := c.bitmapCacheManager.GetCachedBitmap(uint16(v.CacheId), v.CacheIndex, v.Key1, v.Key2)
						if cachedBitmap != nil {
							option := &bitmap.Option{
								Top:         int(v.DestTop),
								Left:        int(v.DestLeft),
								Width:       int(cachedBitmap.Width),
								Height:      int(cachedBitmap.Height),
								BitPerPixel: int(cachedBitmap.BitsPerPixel),
								Data:        cachedBitmap.BitmapDataStream,
							}
							processor.ProcessBitmap(option, bitmap.NewBitmapFromRLE(option))
							glog.Debugf("Retrieved cached bitmap: %dx%d", option.Width, option.Height)
						} else {
							glog.Warnf("Cached bitmap not found: cache=%d, index=%d", v.CacheId, v.CacheIndex)
						}
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
	ch, ok := c.vcManager.GetChannel(packet.ChannelID)
	if !ok {
		return
	}
	if ch.Name == virtualchannel.CHANNEL_NAME_CLIPRDR {
		// Route to clipboard manager
		msg, err := clipboard.ReadClipboardMessage(bytes.NewReader(packet.Data))
		if err == nil {
			glog.GetStructuredLogger().InfoStructured("Received clipboard message", map[string]interface{}{
				"type":   msg.MessageType,
				"length": msg.DataLength,
			})
			_ = c.clipboardManager.ProcessMessage(msg)
		}
		return
	}
	if ch.Name == virtualchannel.CHANNEL_NAME_RDPDR {
		// Route to device manager
		msg, err := device.ReadDeviceMessage(bytes.NewReader(packet.Data))
		if err == nil {
			glog.GetStructuredLogger().InfoStructured("Received device message", map[string]interface{}{
				"component_id": msg.ComponentID,
				"packet_id":    msg.PacketID,
				"data_length":  len(msg.Data),
			})
			_ = c.deviceManager.ProcessMessage(msg)
		}
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
	glog.GetStructuredLogger().InfoStructured("DVC create request", map[string]interface{}{
		"channel_name": req.ChannelName,
		"channel_id":   req.ChannelId,
	})
	// Use custom handler if registered
	handler, ok := c.dvcHandlers[req.ChannelName]
	if !ok {
		handler = drdynvc.NewDefaultDynamicVirtualChannelHandler()
	}
	err = c.dvcManager.RegisterChannelWithID(req.ChannelId, req.ChannelName, handler)
	if err != nil {
		glog.GetStructuredLogger().ErrorStructured("Failed to register DVC", err, map[string]interface{}{
			"channel_name": req.ChannelName,
			"channel_id":   req.ChannelId,
		})
	}
	if handler != nil {
		handler.OnChannelCreated(req.ChannelId, req.ChannelName)
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
	glog.GetStructuredLogger().InfoStructured("DVC open response", map[string]interface{}{
		"channel_id": resp.ChannelId,
		"status":     resp.Status,
	})
	if resp.Status == drdynvc.DVCOPEN_SUCCESS {
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
	glog.GetStructuredLogger().InfoStructured("DVC close response", map[string]interface{}{
		"channel_id": resp.ChannelId,
		"status":     resp.Status,
	})
	if resp.Status == drdynvc.DVCCLOSE_SUCCESS {
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

// RegisterDynamicVirtualChannelHandler allows users to register a custom handler for a DVC by name
func (c *Client) RegisterDynamicVirtualChannelHandler(channelName string, handler drdynvc.DynamicVirtualChannelHandler) error {
	if channelName == "" || handler == nil {
		return fmt.Errorf("channel name and handler must be non-nil")
	}
	c.dvcHandlers[channelName] = handler
	glog.GetStructuredLogger().InfoStructured("Registered DVC handler", map[string]interface{}{
		"channel_name": channelName,
	})
	return nil
}

// ListDynamicVirtualChannels returns a list of currently open DVCs
func (c *Client) ListDynamicVirtualChannels() []string {
	channels := []string{}
	for _, ch := range c.dvcManager.Channels {
		channels = append(channels, ch.ChannelName)
	}
	return channels
}

// RegisterClipboardHandler allows users to register a custom clipboard handler
func (c *Client) RegisterClipboardHandler(handler clipboard.ClipboardHandler) error {
	if handler == nil {
		return fmt.Errorf("clipboard handler must be non-nil")
	}
	c.clipboardManager = clipboard.NewClipboardManager(handler)
	glog.GetStructuredLogger().InfoStructured("Registered clipboard handler", map[string]interface{}{})
	return nil
}

// IsClipboardChannelOpen returns true if the cliprdr channel is registered and ready
func (c *Client) IsClipboardChannelOpen() bool {
	_, ok := c.vcManager.GetChannelByName(virtualchannel.CHANNEL_NAME_CLIPRDR)
	return ok
}

// RegisterDeviceHandler allows users to register a custom device handler
func (c *Client) RegisterDeviceHandler(handler device.DeviceHandler) error {
	if handler == nil {
		return fmt.Errorf("device handler must be non-nil")
	}
	c.deviceManager = device.NewDeviceManager(handler)
	glog.GetStructuredLogger().InfoStructured("Registered device handler", map[string]interface{}{})
	return nil
}

// IsDeviceChannelOpen returns true if the rdpdr channel is registered and ready
func (c *Client) IsDeviceChannelOpen() bool {
	_, ok := c.vcManager.GetChannelByName(virtualchannel.CHANNEL_NAME_RDPDR)
	return ok
}

// GetDeviceCount returns the number of currently redirected devices
func (c *Client) GetDeviceCount() int {
	return c.deviceManager.GetDeviceCount()
}

// ListDevices returns all currently redirected devices
func (c *Client) ListDevices() []*device.DeviceAnnounce {
	return c.deviceManager.ListDevices()
}

// GetDeviceStats returns statistics about device usage
func (c *Client) GetDeviceStats() map[string]interface{} {
	return c.deviceManager.GetDeviceStats()
}

// SendDeviceMessage sends a device message on the rdpdr channel
func (c *Client) SendDeviceMessage(msg *device.DeviceMessage) error {
	return c.SendVirtualChannelData(virtualchannel.CHANNEL_NAME_RDPDR, msg.Serialize(), 0)
}

// AnnounceDevice announces a new device for redirection
func (c *Client) AnnounceDevice(deviceType device.DeviceType, preferredDosName, deviceData string) error {
	msg := c.deviceManager.CreateDeviceAnnounceMessage(deviceType, preferredDosName, deviceData)
	return c.SendDeviceMessage(msg)
}

// SendPrinterData sends printer data to the server
func (c *Client) SendPrinterData(jobID uint32, data []byte, flags uint32) error {
	msg := c.deviceManager.CreatePrinterDataMessage(jobID, data, flags)
	return c.SendDeviceMessage(msg)
}

// SetMonitors sets the multi-monitor layout for the client
func (c *Client) SetMonitors(monitors []mcs.MonitorLayout) {
	c.monitors = monitors
}

// GetMonitors returns the current monitor layout
func (c *Client) GetMonitors() []mcs.MonitorLayout {
	return c.monitors
}
