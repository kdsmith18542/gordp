package virtualchannels

import (
	"fmt"
	"sync"

	"github.com/kdsmith18542/gordp"
)

// VirtualChannelManager manages virtual channels for the GUI
type VirtualChannelManager struct {
	client *gordp.Client

	// Channel handlers
	clipboardHandler *ClipboardHandler
	deviceHandler    *DeviceHandler
	audioHandler     *AudioHandler

	// State management
	mu       sync.RWMutex
	channels map[string]bool // Track open channels
}

// NewVirtualChannelManager creates a new virtual channel manager
func NewVirtualChannelManager(client *gordp.Client) *VirtualChannelManager {
	manager := &VirtualChannelManager{
		client:   client,
		channels: make(map[string]bool),
	}

	// Initialize handlers
	manager.clipboardHandler = NewClipboardHandler(manager)
	manager.deviceHandler = NewDeviceHandler(manager)
	manager.audioHandler = NewAudioHandler(manager)

	return manager
}

// InitializeChannels initializes all virtual channels
func (m *VirtualChannelManager) InitializeChannels() error {
	m.mu.Lock()
	defer m.mu.Unlock()

	// Register clipboard handler
	err := m.client.RegisterClipboardHandler(m.clipboardHandler)
	if err != nil {
		return fmt.Errorf("failed to register clipboard handler: %v", err)
	}

	// Register device handler
	err = m.client.RegisterDeviceHandler(m.deviceHandler)
	if err != nil {
		return fmt.Errorf("failed to register device handler: %v", err)
	}

	fmt.Println("Virtual channels initialized")
	return nil
}

// GetClipboardHandler returns the clipboard handler
func (m *VirtualChannelManager) GetClipboardHandler() *ClipboardHandler {
	return m.clipboardHandler
}

// GetDeviceHandler returns the device handler
func (m *VirtualChannelManager) GetDeviceHandler() *DeviceHandler {
	return m.deviceHandler
}

// GetAudioHandler returns the audio handler
func (m *VirtualChannelManager) GetAudioHandler() *AudioHandler {
	return m.audioHandler
}

// IsChannelOpen checks if a specific channel is open
func (m *VirtualChannelManager) IsChannelOpen(channelName string) bool {
	m.mu.RLock()
	defer m.mu.RUnlock()
	return m.channels[channelName]
}

// SetChannelOpen sets the open state of a channel
func (m *VirtualChannelManager) SetChannelOpen(channelName string, open bool) {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.channels[channelName] = open
}

// GetOpenChannels returns a list of open channels
func (m *VirtualChannelManager) GetOpenChannels() []string {
	m.mu.RLock()
	defer m.mu.RUnlock()

	var openChannels []string
	for channel, open := range m.channels {
		if open {
			openChannels = append(openChannels, channel)
		}
	}
	return openChannels
}

// CloseAllChannels closes all virtual channels
func (m *VirtualChannelManager) CloseAllChannels() {
	m.mu.Lock()
	defer m.mu.Unlock()

	for channel := range m.channels {
		m.channels[channel] = false
	}

	fmt.Println("All virtual channels closed")
}
