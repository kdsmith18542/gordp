package plugins

import (
	"context"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
	"regexp"
	"strings"
	"sync"

	"github.com/kdsmith18542/gordp/plugin"
)

// PluginInfo represents information about a plugin
type PluginInfo struct {
	Name         string
	Version      string
	Description  string
	Enabled      bool
	Config       map[string]interface{}
	Path         string
	Author       string
	License      string
	APIVersion   string
	Dependencies []string
}

// PluginManager manages RDP plugins for the GUI
type PluginManager struct {
	mu sync.RWMutex

	// Plugin management
	plugins map[string]*PluginInfo
	manager *plugin.PluginManager

	// State
	isEnabled bool

	// Plugin directories
	pluginDirs []string
}

// NewPluginManager creates a new plugin manager
func NewPluginManager() *PluginManager {
	// Get plugin directories
	pluginDirs := []string{
		"./plugins",                    // Local plugins
		"./gui/plugins",                // GUI plugins
		"./plugin/examples",            // Example plugins
		os.Getenv("GORDP_PLUGIN_PATH"), // Environment variable
	}

	// Filter out empty paths
	var validDirs []string
	for _, dir := range pluginDirs {
		if dir != "" {
			validDirs = append(validDirs, dir)
		}
	}

	return &PluginManager{
		plugins:    make(map[string]*PluginInfo),
		manager:    plugin.NewPluginManager(),
		isEnabled:  true,
		pluginDirs: validDirs,
	}
}

// LoadPlugin loads a plugin from a file path
func (m *PluginManager) LoadPlugin(pluginPath string) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	if !m.isEnabled {
		return fmt.Errorf("plugin system is disabled")
	}

	// Validate plugin path
	if !m.isValidPluginPath(pluginPath) {
		return fmt.Errorf("invalid plugin path: %s", pluginPath)
	}

	// Check if plugin is already loaded
	pluginName := filepath.Base(pluginPath)
	if _, exists := m.plugins[pluginName]; exists {
		return fmt.Errorf("plugin %s is already loaded", pluginName)
	}

	// Get plugin info
	info, err := m.GetPluginInfo(pluginPath)
	if err != nil {
		return fmt.Errorf("failed to get plugin info: %v", err)
	}

	// Try to load the plugin using Go's plugin system
	// Note: This is a simplified implementation. In a real scenario,
	// you would use Go's plugin package or a custom plugin loader
	loadedPlugin, err := m.loadGoPlugin(pluginPath)
	if err != nil {
		return fmt.Errorf("failed to load plugin %s: %v", pluginName, err)
	}

	// Register plugin with core manager
	if err := m.manager.RegisterPlugin(loadedPlugin); err != nil {
		return fmt.Errorf("failed to register plugin %s: %v", pluginName, err)
	}

	// Add to our tracking
	m.plugins[pluginName] = info
	fmt.Printf("Plugin loaded: %s v%s from %s\n", info.Name, info.Version, pluginPath)

	return nil
}

// UnloadPlugin unloads a plugin by name
func (m *PluginManager) UnloadPlugin(pluginName string) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	if !m.isEnabled {
		return fmt.Errorf("plugin system is disabled")
	}

	// Check if plugin is loaded
	info, exists := m.plugins[pluginName]
	if !exists {
		return fmt.Errorf("plugin %s is not loaded", pluginName)
	}

	// Stop plugin if it's running
	if err := m.manager.StopPlugin(pluginName); err != nil {
		fmt.Printf("Warning: failed to stop plugin %s: %v\n", pluginName, err)
	}

	// Unregister plugin from core manager
	if err := m.manager.UnregisterPlugin(pluginName); err != nil {
		return fmt.Errorf("failed to unregister plugin %s: %v", pluginName, err)
	}

	// Remove from our tracking
	delete(m.plugins, pluginName)
	fmt.Printf("Plugin unloaded: %s from %s\n", pluginName, info.Path)

	return nil
}

// GetPluginInfo extracts plugin metadata from a plugin file
func (m *PluginManager) GetPluginInfo(pluginPath string) (*PluginInfo, error) {
	// Try to read metadata from JSON file first
	metadataPath := strings.TrimSuffix(pluginPath, filepath.Ext(pluginPath)) + ".json"
	if metadata, err := m.readPluginMetadata(metadataPath); err == nil {
		return metadata, nil
	}

	// Try to extract metadata from binary file
	if metadata, err := m.extractMetadataFromBinary(pluginPath); err == nil {
		return metadata, nil
	}

	// Fallback to basic info
	pluginName := filepath.Base(pluginPath)
	return &PluginInfo{
		Name:         pluginName,
		Version:      "1.0.0",
		Description:  fmt.Sprintf("Plugin: %s", pluginName),
		Enabled:      false,
		Config:       make(map[string]interface{}),
		Path:         pluginPath,
		Author:       "Unknown",
		License:      "Unknown",
		APIVersion:   "1.0",
		Dependencies: []string{},
	}, nil
}

// readPluginMetadata reads plugin metadata from a JSON file
func (m *PluginManager) readPluginMetadata(metadataPath string) (*PluginInfo, error) {
	data, err := ioutil.ReadFile(metadataPath)
	if err != nil {
		return nil, fmt.Errorf("failed to read metadata file: %v", err)
	}

	var metadata map[string]interface{}
	if err := json.Unmarshal(data, &metadata); err != nil {
		return nil, fmt.Errorf("failed to parse metadata JSON: %v", err)
	}

	// Extract plugin info from metadata
	info := &PluginInfo{
		Config:       make(map[string]interface{}),
		Dependencies: []string{},
	}

	if name, ok := metadata["name"].(string); ok {
		info.Name = name
	}
	if version, ok := metadata["version"].(string); ok {
		info.Version = version
	}
	if description, ok := metadata["description"].(string); ok {
		info.Description = description
	}
	if author, ok := metadata["author"].(string); ok {
		info.Author = author
	}
	if license, ok := metadata["license"].(string); ok {
		info.License = license
	}
	if apiVersion, ok := metadata["api_version"].(string); ok {
		info.APIVersion = apiVersion
	}
	if config, ok := metadata["config"].(map[string]interface{}); ok {
		info.Config = config
	}
	if deps, ok := metadata["dependencies"].([]interface{}); ok {
		for _, dep := range deps {
			if depStr, ok := dep.(string); ok {
				info.Dependencies = append(info.Dependencies, depStr)
			}
		}
	}

	info.Enabled = false // Default to disabled
	info.Path = strings.TrimSuffix(metadataPath, ".json")

	return info, nil
}

// extractMetadataFromBinary extracts metadata from a binary plugin file
func (m *PluginManager) extractMetadataFromBinary(pluginPath string) (*PluginInfo, error) {
	data, err := ioutil.ReadFile(pluginPath)
	if err != nil {
		return nil, fmt.Errorf("failed to read plugin file: %v", err)
	}

	// Look for metadata patterns in the binary
	content := string(data)

	// Extract plugin name from filename
	pluginName := filepath.Base(pluginPath)
	pluginName = strings.TrimSuffix(pluginName, filepath.Ext(pluginName))

	// Look for version pattern
	versionRegex := regexp.MustCompile(`(?i)version["\s]*[:=]["\s]*([0-9]+\.[0-9]+\.[0-9]+)`)
	versionMatch := versionRegex.FindStringSubmatch(content)
	version := "1.0.0"
	if len(versionMatch) > 1 {
		version = versionMatch[1]
	}

	// Look for description pattern
	descRegex := regexp.MustCompile(`(?i)description["\s]*[:=]["\s]*"([^"]+)"`)
	descMatch := descRegex.FindStringSubmatch(content)
	description := fmt.Sprintf("Plugin: %s", pluginName)
	if len(descMatch) > 1 {
		description = descMatch[1]
	}

	// Look for author pattern
	authorRegex := regexp.MustCompile(`(?i)author["\s]*[:=]["\s]*"([^"]+)"`)
	authorMatch := authorRegex.FindStringSubmatch(content)
	author := "Unknown"
	if len(authorMatch) > 1 {
		author = authorMatch[1]
	}

	// Look for license pattern
	licenseRegex := regexp.MustCompile(`(?i)license["\s]*[:=]["\s]*"([^"]+)"`)
	licenseMatch := licenseRegex.FindStringSubmatch(content)
	license := "Unknown"
	if len(licenseMatch) > 1 {
		license = licenseMatch[1]
	}

	return &PluginInfo{
		Name:         pluginName,
		Version:      version,
		Description:  description,
		Author:       author,
		License:      license,
		Enabled:      false,
		Config:       make(map[string]interface{}),
		Path:         pluginPath,
		APIVersion:   "1.0",
		Dependencies: []string{},
	}, nil
}

// loadGoPlugin loads a Go plugin (simplified implementation)
func (m *PluginManager) loadGoPlugin(pluginPath string) (plugin.Plugin, error) {
	// This is a simplified implementation
	// In a real scenario, you would use Go's plugin package or a custom loader

	// For now, we'll create a dummy plugin based on the file path
	// In practice, you would:
	// 1. Use plugin.Open() to load the plugin
	// 2. Look up the plugin symbol
	// 3. Cast it to the appropriate interface

	pluginName := filepath.Base(pluginPath)

	// Create a basic plugin implementation
	basicPlugin := &BasicPlugin{
		info: &plugin.PluginInfo{
			Name:        pluginName,
			Version:     "1.0.0",
			Type:        plugin.PluginTypeCustom,
			Description: fmt.Sprintf("Loaded plugin: %s", pluginName),
			Author:      "GoRDP",
			License:     "MIT",
			Config:      make(map[string]interface{}),
		},
		status: plugin.PluginStatusUnloaded,
	}

	return basicPlugin, nil
}

// isValidPluginPath checks if a plugin path is valid
func (m *PluginManager) isValidPluginPath(pluginPath string) bool {
	// Check if file exists
	if _, err := os.Stat(pluginPath); os.IsNotExist(err) {
		return false
	}

	// Check if it's a plugin file
	ext := strings.ToLower(filepath.Ext(pluginPath))
	validExts := []string{".so", ".dll", ".dylib", ".go"}

	for _, validExt := range validExts {
		if ext == validExt {
			return true
		}
	}

	return false
}

// ScanPluginDirectories scans all plugin directories for available plugins
func (m *PluginManager) ScanPluginDirectories() ([]*PluginInfo, error) {
	var plugins []*PluginInfo

	for _, dir := range m.pluginDirs {
		if dir == "" {
			continue
		}

		dirPlugins, err := m.scanDirectory(dir)
		if err != nil {
			fmt.Printf("Warning: failed to scan directory %s: %v\n", dir, err)
			continue
		}

		plugins = append(plugins, dirPlugins...)
	}

	return plugins, nil
}

// scanDirectory scans a directory for plugin files
func (m *PluginManager) scanDirectory(dir string) ([]*PluginInfo, error) {
	var plugins []*PluginInfo

	entries, err := ioutil.ReadDir(dir)
	if err != nil {
		return nil, err
	}

	for _, entry := range entries {
		if entry.IsDir() {
			continue
		}

		pluginPath := filepath.Join(dir, entry.Name())
		if !m.isValidPluginPath(pluginPath) {
			continue
		}

		info, err := m.GetPluginInfo(pluginPath)
		if err != nil {
			fmt.Printf("Warning: failed to get info for %s: %v\n", pluginPath, err)
			continue
		}

		plugins = append(plugins, info)
	}

	return plugins, nil
}

// BasicPlugin is a basic plugin implementation for testing
type BasicPlugin struct {
	info   *plugin.PluginInfo
	status plugin.PluginStatus
}

func (bp *BasicPlugin) Info() *plugin.PluginInfo {
	return bp.info
}

func (bp *BasicPlugin) Initialize(config map[string]interface{}) error {
	bp.status = plugin.PluginStatusLoaded
	return nil
}

func (bp *BasicPlugin) Start(ctx context.Context) error {
	bp.status = plugin.PluginStatusRunning
	return nil
}

func (bp *BasicPlugin) Stop() error {
	bp.status = plugin.PluginStatusStopped
	return nil
}

func (bp *BasicPlugin) Status() plugin.PluginStatus {
	return bp.status
}

// RegisterPlugin registers a plugin
func (m *PluginManager) RegisterPlugin(plugin plugin.Plugin) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	if !m.isEnabled {
		return fmt.Errorf("plugin system is disabled")
	}

	// Register plugin via the core plugin manager
	if err := m.manager.RegisterPlugin(plugin); err != nil {
		return fmt.Errorf("failed to register plugin: %v", err)
	}

	// Get plugin info
	info := plugin.Info()
	if info == nil {
		return fmt.Errorf("plugin info is nil")
	}

	// Create plugin info for GUI
	guiInfo := &PluginInfo{
		Name:        info.Name,
		Version:     info.Version,
		Description: info.Description,
		Enabled:     true,
		Config:      make(map[string]interface{}),
	}

	m.plugins[guiInfo.Name] = guiInfo
	fmt.Printf("Plugin registered: %s v%s\n", guiInfo.Name, guiInfo.Version)

	return nil
}

// UnregisterPlugin unregisters a plugin by name
func (m *PluginManager) UnregisterPlugin(name string) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	if !m.isEnabled {
		return fmt.Errorf("plugin system is disabled")
	}

	// Unregister plugin via the core plugin manager
	if err := m.manager.UnregisterPlugin(name); err != nil {
		return fmt.Errorf("failed to unregister plugin %s: %v", name, err)
	}

	// Remove from our tracking
	delete(m.plugins, name)
	fmt.Printf("Plugin unregistered: %s\n", name)

	return nil
}

// GetPlugins returns all registered plugins
func (m *PluginManager) GetPlugins() []*PluginInfo {
	m.mu.RLock()
	defer m.mu.RUnlock()

	plugins := make([]*PluginInfo, 0, len(m.plugins))
	for _, plugin := range m.plugins {
		plugins = append(plugins, plugin)
	}
	return plugins
}

// GetPlugin returns a specific plugin by name
func (m *PluginManager) GetPlugin(name string) *PluginInfo {
	m.mu.RLock()
	defer m.mu.RUnlock()
	return m.plugins[name]
}

// EnablePlugin enables or disables a plugin
func (m *PluginManager) EnablePlugin(name string, enabled bool) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	plugin, exists := m.plugins[name]
	if !exists {
		return fmt.Errorf("plugin %s not found", name)
	}

	plugin.Enabled = enabled
	status := "enabled"
	if !enabled {
		status = "disabled"
	}
	fmt.Printf("Plugin %s %s\n", name, status)

	return nil
}

// SetPluginConfig sets configuration for a plugin
func (m *PluginManager) SetPluginConfig(name string, key string, value interface{}) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	plugin, exists := m.plugins[name]
	if !exists {
		return fmt.Errorf("plugin %s not found", name)
	}

	plugin.Config[key] = value
	fmt.Printf("Plugin %s config updated: %s = %v\n", name, key, value)

	return nil
}

// GetPluginConfig gets configuration for a plugin
func (m *PluginManager) GetPluginConfig(name string, key string) (interface{}, bool) {
	m.mu.RLock()
	defer m.mu.RUnlock()

	plugin, exists := m.plugins[name]
	if !exists {
		return nil, false
	}

	value, exists := plugin.Config[key]
	return value, exists
}

// IsEnabled returns whether the plugin system is enabled
func (m *PluginManager) IsEnabled() bool {
	m.mu.RLock()
	defer m.mu.RUnlock()
	return m.isEnabled
}

// SetEnabled enables or disables the plugin system
func (m *PluginManager) SetEnabled(enabled bool) {
	m.mu.Lock()
	defer m.mu.Unlock()

	m.isEnabled = enabled
	status := "enabled"
	if !enabled {
		status = "disabled"
	}
	fmt.Printf("Plugin system %s\n", status)
}

// GetPluginStats returns statistics about plugins
func (m *PluginManager) GetPluginStats() map[string]interface{} {
	m.mu.RLock()
	defer m.mu.RUnlock()

	stats := m.manager.GetStats()

	return map[string]interface{}{
		"total_plugins":   stats.TotalPlugins,
		"running_plugins": stats.RunningPlugins,
		"error_plugins":   stats.ErrorPlugins,
		"system_enabled":  m.isEnabled,
	}
}

// InitializePlugin initializes a plugin
func (m *PluginManager) InitializePlugin(name string, config map[string]interface{}) error {
	m.mu.RLock()
	defer m.mu.RUnlock()

	if !m.isEnabled {
		return fmt.Errorf("plugin system is disabled")
	}

	plugin, exists := m.plugins[name]
	if !exists {
		return fmt.Errorf("plugin %s not found", name)
	}

	if !plugin.Enabled {
		return fmt.Errorf("plugin %s is disabled", name)
	}

	// Initialize plugin via the core plugin manager
	if err := m.manager.InitializePlugin(name, config); err != nil {
		return fmt.Errorf("failed to initialize plugin %s: %v", name, err)
	}

	fmt.Printf("Plugin %s initialized\n", name)
	return nil
}

// StartPlugin starts a plugin
func (m *PluginManager) StartPlugin(name string) error {
	m.mu.RLock()
	defer m.mu.RUnlock()

	if !m.isEnabled {
		return fmt.Errorf("plugin system is disabled")
	}

	plugin, exists := m.plugins[name]
	if !exists {
		return fmt.Errorf("plugin %s not found", name)
	}

	if !plugin.Enabled {
		return fmt.Errorf("plugin %s is disabled", name)
	}

	// Start plugin via the core plugin manager
	if err := m.manager.StartPlugin(name); err != nil {
		return fmt.Errorf("failed to start plugin %s: %v", name, err)
	}

	fmt.Printf("Plugin %s started\n", name)
	return nil
}

// StopPlugin stops a plugin
func (m *PluginManager) StopPlugin(name string) error {
	m.mu.RLock()
	defer m.mu.RUnlock()

	if !m.isEnabled {
		return fmt.Errorf("plugin system is disabled")
	}

	_, exists := m.plugins[name]
	if !exists {
		return fmt.Errorf("plugin %s not found", name)
	}

	// Stop plugin via the core plugin manager
	if err := m.manager.StopPlugin(name); err != nil {
		return fmt.Errorf("failed to stop plugin %s: %v", name, err)
	}

	fmt.Printf("Plugin %s stopped\n", name)
	return nil
}

// StartAllPlugins starts all plugins
func (m *PluginManager) StartAllPlugins() error {
	m.mu.RLock()
	defer m.mu.RUnlock()

	if !m.isEnabled {
		return fmt.Errorf("plugin system is disabled")
	}

	// Start all plugins via the core plugin manager
	if err := m.manager.StartAllPlugins(); err != nil {
		return fmt.Errorf("failed to start all plugins: %v", err)
	}

	fmt.Println("All plugins started")
	return nil
}

// StopAllPlugins stops all plugins
func (m *PluginManager) StopAllPlugins() error {
	m.mu.RLock()
	defer m.mu.RUnlock()

	if !m.isEnabled {
		return fmt.Errorf("plugin system is disabled")
	}

	// Stop all plugins via the core plugin manager
	if err := m.manager.StopAllPlugins(); err != nil {
		return fmt.Errorf("failed to stop all plugins: %v", err)
	}

	fmt.Println("All plugins stopped")
	return nil
}

// GetPluginStatus returns the status of a plugin
func (m *PluginManager) GetPluginStatus(name string) (plugin.PluginStatus, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()

	if !m.isEnabled {
		return plugin.PluginStatusUnloaded, fmt.Errorf("plugin system is disabled")
	}

	// Get plugin from core manager
	p, exists := m.manager.GetPlugin(name)
	if !exists {
		return plugin.PluginStatusUnloaded, fmt.Errorf("plugin %s not found", name)
	}

	return p.Status(), nil
}
