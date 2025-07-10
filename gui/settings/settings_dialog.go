package settings

import (
	"bufio"
	"fmt"
	"os"
	"strconv"
	"strings"
)

// Settings represents application settings
type Settings struct {
	// Display settings
	DefaultZoom         float64
	EnableSmoothScaling bool

	// Connection settings
	DefaultPort    int
	ConnectTimeout int
	AutoReconnect  bool

	// Performance settings
	EnableBitmapCache bool
	CacheSize         int
	CompressionLevel  int
}

// SettingsDialog represents the settings dialog
type SettingsDialog struct {
	// Current settings
	currentSettings *Settings
}

// NewSettingsDialog creates a new settings dialog
func NewSettingsDialog() *SettingsDialog {
	dialog := &SettingsDialog{
		currentSettings: &Settings{
			DefaultZoom:         1.0,
			EnableSmoothScaling: true,
			DefaultPort:         3389,
			ConnectTimeout:      5,
			AutoReconnect:       false,
			EnableBitmapCache:   true,
			CacheSize:           100,
			CompressionLevel:    6,
		},
	}

	return dialog
}

// Show displays the settings dialog
func (d *SettingsDialog) Show() {
	fmt.Println("=== Settings ===")

	reader := bufio.NewReader(os.Stdin)

	// Display settings
	fmt.Println("\n--- Display Settings ---")

	fmt.Printf("Default Zoom (current: %.1fx): ", d.currentSettings.DefaultZoom)
	zoomStr, _ := reader.ReadString('\n')
	zoomStr = strings.TrimSpace(zoomStr)
	if zoomStr != "" {
		if zoom, err := strconv.ParseFloat(zoomStr, 64); err == nil && zoom >= 0.1 && zoom <= 5.0 {
			d.currentSettings.DefaultZoom = zoom
		}
	}

	fmt.Printf("Enable Smooth Scaling (current: %t) [y/n]: ", d.currentSettings.EnableSmoothScaling)
	scalingStr, _ := reader.ReadString('\n')
	scalingStr = strings.ToLower(strings.TrimSpace(scalingStr))
	if scalingStr == "y" || scalingStr == "yes" {
		d.currentSettings.EnableSmoothScaling = true
	} else if scalingStr == "n" || scalingStr == "no" {
		d.currentSettings.EnableSmoothScaling = false
	}

	// Connection settings
	fmt.Println("\n--- Connection Settings ---")

	fmt.Printf("Default Port (current: %d): ", d.currentSettings.DefaultPort)
	portStr, _ := reader.ReadString('\n')
	portStr = strings.TrimSpace(portStr)
	if portStr != "" {
		if port, err := strconv.Atoi(portStr); err == nil && port >= 1 && port <= 65535 {
			d.currentSettings.DefaultPort = port
		}
	}

	fmt.Printf("Connection Timeout (current: %d seconds): ", d.currentSettings.ConnectTimeout)
	timeoutStr, _ := reader.ReadString('\n')
	timeoutStr = strings.TrimSpace(timeoutStr)
	if timeoutStr != "" {
		if timeout, err := strconv.Atoi(timeoutStr); err == nil && timeout >= 1 && timeout <= 60 {
			d.currentSettings.ConnectTimeout = timeout
		}
	}

	fmt.Printf("Auto-reconnect (current: %t) [y/n]: ", d.currentSettings.AutoReconnect)
	reconnectStr, _ := reader.ReadString('\n')
	reconnectStr = strings.ToLower(strings.TrimSpace(reconnectStr))
	if reconnectStr == "y" || reconnectStr == "yes" {
		d.currentSettings.AutoReconnect = true
	} else if reconnectStr == "n" || reconnectStr == "no" {
		d.currentSettings.AutoReconnect = false
	}

	// Performance settings
	fmt.Println("\n--- Performance Settings ---")

	fmt.Printf("Enable Bitmap Cache (current: %t) [y/n]: ", d.currentSettings.EnableBitmapCache)
	cacheStr, _ := reader.ReadString('\n')
	cacheStr = strings.ToLower(strings.TrimSpace(cacheStr))
	if cacheStr == "y" || cacheStr == "yes" {
		d.currentSettings.EnableBitmapCache = true
	} else if cacheStr == "n" || cacheStr == "no" {
		d.currentSettings.EnableBitmapCache = false
	}

	fmt.Printf("Cache Size (current: %d MB): ", d.currentSettings.CacheSize)
	cacheSizeStr, _ := reader.ReadString('\n')
	cacheSizeStr = strings.TrimSpace(cacheSizeStr)
	if cacheSizeStr != "" {
		if cacheSize, err := strconv.Atoi(cacheSizeStr); err == nil && cacheSize >= 1 && cacheSize <= 1000 {
			d.currentSettings.CacheSize = cacheSize
		}
	}

	fmt.Printf("Compression Level (current: %d, 0-9): ", d.currentSettings.CompressionLevel)
	compressionStr, _ := reader.ReadString('\n')
	compressionStr = strings.TrimSpace(compressionStr)
	if compressionStr != "" {
		if compression, err := strconv.Atoi(compressionStr); err == nil && compression >= 0 && compression <= 9 {
			d.currentSettings.CompressionLevel = compression
		}
	}

	fmt.Println("\nSettings saved!")
}

// GetSettings returns the current settings
func (d *SettingsDialog) GetSettings() *Settings {
	return d.currentSettings
}

// SetSettings sets the current settings
func (d *SettingsDialog) SetSettings(settings *Settings) {
	if settings != nil {
		d.currentSettings = settings
	}
}
