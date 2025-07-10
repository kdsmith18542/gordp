package plugin

import (
	"context"
	"fmt"
	"sync"
	"time"

	"github.com/kdsmith18542/gordp/proto/bitmap"
	"github.com/kdsmith18542/gordp/proto/clipboard"
	"github.com/kdsmith18542/gordp/proto/device"
)

// PluginType represents the type of plugin
type PluginType string

const (
	PluginTypeInput          PluginType = "input"
	PluginTypeDisplay        PluginType = "display"
	PluginTypeAudio          PluginType = "audio"
	PluginTypeClipboard      PluginType = "clipboard"
	PluginTypeDevice         PluginType = "device"
	PluginTypeSecurity       PluginType = "security"
	PluginTypePerformance    PluginType = "performance"
	PluginTypeVirtualChannel PluginType = "virtual_channel"
	PluginTypeCustom         PluginType = "custom"
)

// PluginStatus represents the status of a plugin
type PluginStatus string

const (
	PluginStatusUnloaded PluginStatus = "unloaded"
	PluginStatusLoading  PluginStatus = "loading"
	PluginStatusLoaded   PluginStatus = "loaded"
	PluginStatusRunning  PluginStatus = "running"
	PluginStatusError    PluginStatus = "error"
	PluginStatusStopped  PluginStatus = "stopped"
)

// PluginInfo contains information about a plugin
type PluginInfo struct {
	Name        string                 `json:"name"`
	Version     string                 `json:"version"`
	Type        PluginType             `json:"type"`
	Description string                 `json:"description"`
	Author      string                 `json:"author"`
	License     string                 `json:"license"`
	Config      map[string]interface{} `json:"config"`
}

// Plugin represents a plugin interface
type Plugin interface {
	// Info returns plugin information
	Info() *PluginInfo

	// Initialize initializes the plugin
	Initialize(config map[string]interface{}) error

	// Start starts the plugin
	Start(ctx context.Context) error

	// Stop stops the plugin
	Stop() error

	// Status returns the current status
	Status() PluginStatus
}

// InputPlugin handles input-related functionality
type InputPlugin interface {
	Plugin

	// OnKeyPress is called when a key is pressed
	OnKeyPress(keyCode uint8, modifiers interface{}) error

	// OnMouseMove is called when the mouse moves
	OnMouseMove(x, y uint16) error

	// OnMouseClick is called when the mouse is clicked
	OnMouseClick(button interface{}, x, y uint16) error
}

// DisplayPlugin handles display-related functionality
type DisplayPlugin interface {
	Plugin

	// OnBitmapReceived is called when a bitmap is received
	OnBitmapReceived(option *bitmap.Option, bitmap *bitmap.BitMap) error

	// OnDisplayChanged is called when display settings change
	OnDisplayChanged(width, height int) error
}

// AudioPlugin handles audio-related functionality
type AudioPlugin interface {
	Plugin

	// OnAudioData is called when audio data is received
	OnAudioData(formatID uint16, data []byte, timestamp uint32) error

	// OnAudioFormatChanged is called when audio format changes
	OnAudioFormatChanged(format interface{}) error
}

// ClipboardPlugin handles clipboard-related functionality
type ClipboardPlugin interface {
	Plugin

	// OnClipboardData is called when clipboard data is received
	OnClipboardData(format clipboard.ClipboardFormat, data []byte) error

	// OnClipboardRequest is called when clipboard data is requested
	OnClipboardRequest(format clipboard.ClipboardFormat) error
}

// DevicePlugin handles device-related functionality
type DevicePlugin interface {
	Plugin

	// OnDeviceAnnounce is called when a device is announced
	OnDeviceAnnounce(device *device.DeviceAnnounce) error

	// OnDeviceIORequest is called when a device I/O request is made
	OnDeviceIORequest(request *device.DeviceIORequest) (*device.DeviceIOCompletion, error)
}

// VirtualChannelPlugin handles virtual channel functionality
type VirtualChannelPlugin interface {
	Plugin

	// OnChannelData is called when data is received on a virtual channel
	OnChannelData(channelID uint32, data []byte) error

	// OnChannelCreated is called when a virtual channel is created
	OnChannelCreated(channelID uint32, channelName string) error
}

// SecurityPlugin handles security-related functionality
type SecurityPlugin interface {
	Plugin

	// OnAuthentication is called during authentication
	OnAuthentication(username, password string) error

	// OnCertificateValidation is called during certificate validation
	OnCertificateValidation(certificate interface{}) error
}

// PerformancePlugin handles performance-related functionality
type PerformancePlugin interface {
	Plugin

	// OnPerformanceMetrics is called with performance metrics
	OnPerformanceMetrics(metrics map[string]interface{}) error

	// OnCacheHit is called when a cache hit occurs
	OnCacheHit(cacheType string, hitRate float64) error
}

// PluginManager manages all plugins
type PluginManager struct {
	plugins map[string]Plugin
	mutex   sync.RWMutex
	ctx     context.Context
	cancel  context.CancelFunc
}

// NewPluginManager creates a new plugin manager
func NewPluginManager() *PluginManager {
	ctx, cancel := context.WithCancel(context.Background())
	return &PluginManager{
		plugins: make(map[string]Plugin),
		ctx:     ctx,
		cancel:  cancel,
	}
}

// RegisterPlugin registers a plugin
func (pm *PluginManager) RegisterPlugin(plugin Plugin) error {
	pm.mutex.Lock()
	defer pm.mutex.Unlock()

	info := plugin.Info()
	if info == nil {
		return fmt.Errorf("plugin info is nil")
	}

	if _, exists := pm.plugins[info.Name]; exists {
		return fmt.Errorf("plugin %s already registered", info.Name)
	}

	pm.plugins[info.Name] = plugin
	return nil
}

// UnregisterPlugin unregisters a plugin
func (pm *PluginManager) UnregisterPlugin(name string) error {
	pm.mutex.Lock()
	defer pm.mutex.Unlock()

	plugin, exists := pm.plugins[name]
	if !exists {
		return fmt.Errorf("plugin %s not found", name)
	}

	// Stop the plugin if it's running
	if plugin.Status() == PluginStatusRunning {
		if err := plugin.Stop(); err != nil {
			return fmt.Errorf("failed to stop plugin %s: %w", name, err)
		}
	}

	delete(pm.plugins, name)
	return nil
}

// GetPlugin returns a plugin by name
func (pm *PluginManager) GetPlugin(name string) (Plugin, bool) {
	pm.mutex.RLock()
	defer pm.mutex.RUnlock()

	plugin, exists := pm.plugins[name]
	return plugin, exists
}

// GetPluginsByType returns all plugins of a specific type
func (pm *PluginManager) GetPluginsByType(pluginType PluginType) []Plugin {
	pm.mutex.RLock()
	defer pm.mutex.RUnlock()

	var plugins []Plugin
	for _, plugin := range pm.plugins {
		if plugin.Info().Type == pluginType {
			plugins = append(plugins, plugin)
		}
	}
	return plugins
}

// ListPlugins returns all registered plugins
func (pm *PluginManager) ListPlugins() []Plugin {
	pm.mutex.RLock()
	defer pm.mutex.RUnlock()

	plugins := make([]Plugin, 0, len(pm.plugins))
	for _, plugin := range pm.plugins {
		plugins = append(plugins, plugin)
	}
	return plugins
}

// InitializePlugin initializes a plugin
func (pm *PluginManager) InitializePlugin(name string, config map[string]interface{}) error {
	plugin, exists := pm.GetPlugin(name)
	if !exists {
		return fmt.Errorf("plugin %s not found", name)
	}

	return plugin.Initialize(config)
}

// StartPlugin starts a plugin
func (pm *PluginManager) StartPlugin(name string) error {
	plugin, exists := pm.GetPlugin(name)
	if !exists {
		return fmt.Errorf("plugin %s not found", name)
	}

	return plugin.Start(pm.ctx)
}

// StopPlugin stops a plugin
func (pm *PluginManager) StopPlugin(name string) error {
	plugin, exists := pm.GetPlugin(name)
	if !exists {
		return fmt.Errorf("plugin %s not found", name)
	}

	return plugin.Stop()
}

// StartAllPlugins starts all plugins
func (pm *PluginManager) StartAllPlugins() error {
	pm.mutex.RLock()
	plugins := make([]Plugin, 0, len(pm.plugins))
	for _, plugin := range pm.plugins {
		plugins = append(plugins, plugin)
	}
	pm.mutex.RUnlock()

	for _, plugin := range plugins {
		if err := plugin.Start(pm.ctx); err != nil {
			return fmt.Errorf("failed to start plugin %s: %w", plugin.Info().Name, err)
		}
	}

	return nil
}

// StopAllPlugins stops all plugins
func (pm *PluginManager) StopAllPlugins() error {
	pm.mutex.RLock()
	plugins := make([]Plugin, 0, len(pm.plugins))
	for _, plugin := range pm.plugins {
		plugins = append(plugins, plugin)
	}
	pm.mutex.RUnlock()

	for _, plugin := range plugins {
		if err := plugin.Stop(); err != nil {
			return fmt.Errorf("failed to stop plugin %s: %w", plugin.Info().Name, err)
		}
	}

	return nil
}

// Close closes the plugin manager
func (pm *PluginManager) Close() error {
	pm.cancel()
	return pm.StopAllPlugins()
}

// PluginStats contains statistics about plugins
type PluginStats struct {
	TotalPlugins    int                    `json:"total_plugins"`
	RunningPlugins  int                    `json:"running_plugins"`
	ErrorPlugins    int                    `json:"error_plugins"`
	PluginDetails   map[string]PluginInfo  `json:"plugin_details"`
	PerformanceData map[string]interface{} `json:"performance_data"`
}

// GetStats returns statistics about all plugins
func (pm *PluginManager) GetStats() *PluginStats {
	pm.mutex.RLock()
	defer pm.mutex.RUnlock()

	stats := &PluginStats{
		TotalPlugins:    len(pm.plugins),
		PluginDetails:   make(map[string]PluginInfo),
		PerformanceData: make(map[string]interface{}),
	}

	for name, plugin := range pm.plugins {
		info := plugin.Info()
		if info != nil {
			stats.PluginDetails[name] = *info
		}

		status := plugin.Status()
		switch status {
		case PluginStatusRunning:
			stats.RunningPlugins++
		case PluginStatusError:
			stats.ErrorPlugins++
		}
	}

	return stats
}

// PluginEvent represents a plugin event
type PluginEvent struct {
	PluginName string      `json:"plugin_name"`
	EventType  string      `json:"event_type"`
	Timestamp  time.Time   `json:"timestamp"`
	Data       interface{} `json:"data"`
}

// PluginEventHandler handles plugin events
type PluginEventHandler func(event *PluginEvent) error

// PluginEventManager manages plugin events
type PluginEventManager struct {
	handlers map[string][]PluginEventHandler
	mutex    sync.RWMutex
}

// NewPluginEventManager creates a new plugin event manager
func NewPluginEventManager() *PluginEventManager {
	return &PluginEventManager{
		handlers: make(map[string][]PluginEventHandler),
	}
}

// RegisterEventHandler registers an event handler
func (pem *PluginEventManager) RegisterEventHandler(eventType string, handler PluginEventHandler) {
	pem.mutex.Lock()
	defer pem.mutex.Unlock()

	pem.handlers[eventType] = append(pem.handlers[eventType], handler)
}

// EmitEvent emits a plugin event
func (pem *PluginEventManager) EmitEvent(event *PluginEvent) error {
	pem.mutex.RLock()
	handlers := pem.handlers[event.EventType]
	pem.mutex.RUnlock()

	for _, handler := range handlers {
		if err := handler(event); err != nil {
			return fmt.Errorf("event handler failed: %w", err)
		}
	}

	return nil
}
