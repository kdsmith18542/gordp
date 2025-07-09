package virtualchannel

import (
	"bytes"
	"fmt"
	"io"
	"sync"

	"github.com/GoFeGroup/gordp/core"
	"github.com/GoFeGroup/gordp/glog"
)

// VirtualChannel represents a virtual channel in RDP
type VirtualChannel struct {
	ID       uint16
	Name     string
	Priority uint8
	Flags    uint32
}

// VirtualChannelManager manages virtual channels
type VirtualChannelManager struct {
	channels map[uint16]*VirtualChannel
	mutex    sync.RWMutex
}

// NewVirtualChannelManager creates a new virtual channel manager
func NewVirtualChannelManager() *VirtualChannelManager {
	return &VirtualChannelManager{
		channels: make(map[uint16]*VirtualChannel),
	}
}

// RegisterChannel registers a virtual channel
func (m *VirtualChannelManager) RegisterChannel(channel *VirtualChannel) error {
	m.mutex.Lock()
	defer m.mutex.Unlock()

	if _, exists := m.channels[channel.ID]; exists {
		return fmt.Errorf("virtual channel with ID %d already exists", channel.ID)
	}

	m.channels[channel.ID] = channel
	glog.Debugf("Registered virtual channel: %s (ID: %d)", channel.Name, channel.ID)
	return nil
}

// GetChannel retrieves a virtual channel by ID
func (m *VirtualChannelManager) GetChannel(id uint16) (*VirtualChannel, bool) {
	m.mutex.RLock()
	defer m.mutex.RUnlock()

	channel, exists := m.channels[id]
	return channel, exists
}

// GetChannelByName retrieves a virtual channel by its name
func (m *VirtualChannelManager) GetChannelByName(name string) (*VirtualChannel, bool) {
	m.mutex.RLock()
	defer m.mutex.RUnlock()
	for _, ch := range m.channels {
		if ch.Name == name {
			return ch, true
		}
	}
	return nil, false
}

// ListChannels returns all registered channels
func (m *VirtualChannelManager) ListChannels() []*VirtualChannel {
	m.mutex.RLock()
	defer m.mutex.RUnlock()

	channels := make([]*VirtualChannel, 0, len(m.channels))
	for _, channel := range m.channels {
		channels = append(channels, channel)
	}
	return channels
}

// VirtualChannelData represents data sent over a virtual channel
type VirtualChannelData struct {
	ChannelID uint16
	Data      []byte
	Flags     uint32
}

// VirtualChannelHandler handles virtual channel data
type VirtualChannelHandler interface {
	HandleData(channelID uint16, data []byte) error
	OnChannelOpen(channelID uint16, channelName string) error
	OnChannelClose(channelID uint16) error
}

// DefaultVirtualChannelHandler provides a default implementation
type DefaultVirtualChannelHandler struct {
	manager *VirtualChannelManager
}

// NewDefaultVirtualChannelHandler creates a new default handler
func NewDefaultVirtualChannelHandler(manager *VirtualChannelManager) *DefaultVirtualChannelHandler {
	return &DefaultVirtualChannelHandler{
		manager: manager,
	}
}

// HandleData handles incoming virtual channel data
func (h *DefaultVirtualChannelHandler) HandleData(channelID uint16, data []byte) error {
	channel, exists := h.manager.GetChannel(channelID)
	if !exists {
		return fmt.Errorf("unknown virtual channel ID: %d", channelID)
	}

	glog.Debugf("Received data on virtual channel %s (ID: %d): %d bytes",
		channel.Name, channelID, len(data))

	// Default implementation just logs the data
	// Subclasses can override this to handle specific channel types
	return nil
}

// OnChannelOpen handles channel open events
func (h *DefaultVirtualChannelHandler) OnChannelOpen(channelID uint16, channelName string) error {
	glog.Debugf("Virtual channel opened: %s (ID: %d)", channelName, channelID)
	return nil
}

// OnChannelClose handles channel close events
func (h *DefaultVirtualChannelHandler) OnChannelClose(channelID uint16) error {
	glog.Debugf("Virtual channel closed: ID: %d", channelID)
	return nil
}

// VirtualChannelPacket represents a virtual channel packet header
type VirtualChannelPacket struct {
	Length    uint32
	Flags     uint32
	ChannelID uint16
	Data      []byte
}

// ReadVirtualChannelPacket reads a virtual channel packet from the stream
func ReadVirtualChannelPacket(r io.Reader) (*VirtualChannelPacket, error) {
	packet := &VirtualChannelPacket{}

	// Read packet header
	if err := core.ReadLE(r, &packet.Length); err != nil {
		return nil, fmt.Errorf("failed to read packet length: %v", err)
	}

	if err := core.ReadLE(r, &packet.Flags); err != nil {
		return nil, fmt.Errorf("failed to read packet flags: %v", err)
	}

	if err := core.ReadLE(r, &packet.ChannelID); err != nil {
		return nil, fmt.Errorf("failed to read channel ID: %v", err)
	}

	// Read packet data
	if packet.Length > 0 {
		packet.Data = make([]byte, packet.Length)
		if _, err := io.ReadFull(r, packet.Data); err != nil {
			return nil, fmt.Errorf("failed to read packet data: %w", err)
		}
	}

	return packet, nil
}

// Serialize serializes the virtual channel packet
func (p *VirtualChannelPacket) Serialize() []byte {
	buf := new(bytes.Buffer)
	core.WriteLE(buf, p.Length)
	core.WriteLE(buf, p.Flags)
	core.WriteLE(buf, p.ChannelID)
	if len(p.Data) > 0 {
		buf.Write(p.Data)
	}
	return buf.Bytes()
}

// VirtualChannelFlags
const (
	CHANNEL_FLAG_FIRST             = 0x00000001
	CHANNEL_FLAG_LAST              = 0x00000002
	CHANNEL_FLAG_SHOW_PROTOCOL     = 0x00000010
	CHANNEL_FLAG_SUSPEND           = 0x00000020
	CHANNEL_FLAG_RESUME            = 0x00000040
	CHANNEL_FLAG_SHADOW_PERSISTENT = 0x00000080
	CHANNEL_FLAG_TUNNEL_CREATE     = 0x00000100
	CHANNEL_FLAG_TUNNEL_DATA       = 0x00000200
	CHANNEL_FLAG_TUNNEL_CLOSE      = 0x00000400
)

// Common virtual channel names
const (
	CHANNEL_NAME_CLIPRDR = "cliprdr" // Clipboard redirection
	CHANNEL_NAME_RDPDR   = "rdpdr"   // Device redirection
	CHANNEL_NAME_RDPSND  = "rdpsnd"  // Audio redirection
	CHANNEL_NAME_DRDYNVC = "drdynvc" // Dynamic virtual channels
)
