// Package config provides comprehensive configuration management for the GoRDP library.
// This package supports loading configuration from multiple sources:
//   - JSON and YAML files
//   - Environment variables
//   - Default values
//
// The configuration covers all aspects of RDP connections including:
//   - Connection settings (timeouts, retries, keep-alive)
//   - Authentication (credentials, NLA, SSL)
//   - Display settings (resolution, monitors, DPI)
//   - Performance tuning (caching, compression, bandwidth)
//   - Security settings (certificates, encryption, FIPS)
//   - Input handling (keyboard, mouse, IME)
//   - Virtual channels (clipboard, audio, devices)
//   - Logging configuration
//
// Example usage:
//
//	config := config.DefaultConfig()
//	config.Connection.Address = "192.168.1.100"
//	config.Authentication.Username = "user"
//	config.Authentication.Password = "pass"
//
//	// Or load from file
//	config, err := config.LoadFromFile("config.json")
package config

import (
	"encoding/json"
	"fmt"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/kdsmith18542/gordp/proto/mcs"
	"gopkg.in/yaml.v3"
)

// Config represents the complete configuration for GoRDP
type Config struct {
	// Connection settings
	Connection ConnectionConfig `json:"connection" yaml:"connection"`

	// Authentication settings
	Authentication AuthConfig `json:"authentication" yaml:"authentication"`

	// Display settings
	Display DisplayConfig `json:"display" yaml:"display"`

	// Performance settings
	Performance PerformanceConfig `json:"performance" yaml:"performance"`

	// Security settings
	Security SecurityConfig `json:"security" yaml:"security"`

	// Input settings
	Input InputConfig `json:"input" yaml:"input"`

	// Virtual channel settings
	VirtualChannels VirtualChannelConfig `json:"virtual_channels" yaml:"virtual_channels"`

	// Logging settings
	Logging LoggingConfig `json:"logging" yaml:"logging"`
}

// ConnectionConfig contains connection-related settings
type ConnectionConfig struct {
	Address         string        `json:"address" yaml:"address"`
	Port            int           `json:"port" yaml:"port"`
	ConnectTimeout  time.Duration `json:"connect_timeout" yaml:"connect_timeout"`
	ReadTimeout     time.Duration `json:"read_timeout" yaml:"read_timeout"`
	WriteTimeout    time.Duration `json:"write_timeout" yaml:"write_timeout"`
	KeepAlive       bool          `json:"keep_alive" yaml:"keep_alive"`
	KeepAlivePeriod time.Duration `json:"keep_alive_period" yaml:"keep_alive_period"`
	MaxRetries      int           `json:"max_retries" yaml:"max_retries"`
	RetryDelay      time.Duration `json:"retry_delay" yaml:"retry_delay"`
}

// AuthConfig contains authentication settings
type AuthConfig struct {
	Username        string `json:"username" yaml:"username"`
	Password        string `json:"password" yaml:"password"`
	Domain          string `json:"domain" yaml:"domain"`
	SmartCard       bool   `json:"smart_card" yaml:"smart_card"`
	SmartCardReader string `json:"smart_card_reader" yaml:"smart_card_reader"`
	UseNLA          bool   `json:"use_nla" yaml:"use_nla"`
	UseSSL          bool   `json:"use_ssl" yaml:"use_ssl"`
}

// DisplayConfig contains display-related settings
type DisplayConfig struct {
	Width          int                 `json:"width" yaml:"width"`
	Height         int                 `json:"height" yaml:"height"`
	ColorDepth     int                 `json:"color_depth" yaml:"color_depth"`
	Monitors       []mcs.MonitorLayout `json:"monitors" yaml:"monitors"`
	HighDPI        bool                `json:"high_dpi" yaml:"high_dpi"`
	DesktopScale   int                 `json:"desktop_scale" yaml:"desktop_scale"`
	DeviceScale    int                 `json:"device_scale" yaml:"device_scale"`
	Wallpaper      bool                `json:"wallpaper" yaml:"wallpaper"`
	Themes         bool                `json:"themes" yaml:"themes"`
	FontSmoothing  bool                `json:"font_smoothing" yaml:"font_smoothing"`
	FullWindowDrag bool                `json:"full_window_drag" yaml:"full_window_drag"`
	MenuAnimations bool                `json:"menu_animations" yaml:"menu_animations"`
}

// PerformanceConfig contains performance-related settings
type PerformanceConfig struct {
	BitmapCache      bool   `json:"bitmap_cache" yaml:"bitmap_cache"`
	BitmapCacheSize  int    `json:"bitmap_cache_size" yaml:"bitmap_cache_size"`
	Compression      bool   `json:"compression" yaml:"compression"`
	CompressionLevel int    `json:"compression_level" yaml:"compression_level"`
	NetworkAutoTune  bool   `json:"network_auto_tune" yaml:"network_auto_tune"`
	BandwidthLimit   int    `json:"bandwidth_limit" yaml:"bandwidth_limit"`
	FrameRate        int    `json:"frame_rate" yaml:"frame_rate"`
	QualityLevel     string `json:"quality_level" yaml:"quality_level"`
}

// SecurityConfig contains security-related settings
type SecurityConfig struct {
	FIPSCompliance        bool   `json:"fips_compliance" yaml:"fips_compliance"`
	CertificateValidation bool   `json:"certificate_validation" yaml:"certificate_validation"`
	CertificatePath       string `json:"certificate_path" yaml:"certificate_path"`
	PrivateKeyPath        string `json:"private_key_path" yaml:"private_key_path"`
	TrustedCAs            string `json:"trusted_cas" yaml:"trusted_cas"`
	EncryptionLevel       string `json:"encryption_level" yaml:"encryption_level"`
}

// InputConfig contains input-related settings
type InputConfig struct {
	KeyboardLayout    string `json:"keyboard_layout" yaml:"keyboard_layout"`
	MouseSensitivity  int    `json:"mouse_sensitivity" yaml:"mouse_sensitivity"`
	MouseAcceleration bool   `json:"mouse_acceleration" yaml:"mouse_acceleration"`
	UnicodeInput      bool   `json:"unicode_input" yaml:"unicode_input"`
	IMEInput          bool   `json:"ime_input" yaml:"ime_input"`
	InputBufferSize   int    `json:"input_buffer_size" yaml:"input_buffer_size"`
}

// VirtualChannelConfig contains virtual channel settings
type VirtualChannelConfig struct {
	Clipboard      bool     `json:"clipboard" yaml:"clipboard"`
	Audio          bool     `json:"audio" yaml:"audio"`
	Device         bool     `json:"device" yaml:"device"`
	Printer        bool     `json:"printer" yaml:"printer"`
	Drive          bool     `json:"drive" yaml:"drive"`
	Port           bool     `json:"port" yaml:"port"`
	SmartCard      bool     `json:"smart_card" yaml:"smart_card"`
	CustomChannels []string `json:"custom_channels" yaml:"custom_channels"`
}

// LoggingConfig contains logging settings
type LoggingConfig struct {
	Level      string `json:"level" yaml:"level"`
	Format     string `json:"format" yaml:"format"`
	Output     string `json:"output" yaml:"output"`
	File       string `json:"file" yaml:"file"`
	MaxSize    int    `json:"max_size" yaml:"max_size"`
	MaxBackups int    `json:"max_backups" yaml:"max_backups"`
	MaxAge     int    `json:"max_age" yaml:"max_age"`
	Compress   bool   `json:"compress" yaml:"compress"`
}

// DefaultConfig returns a default configuration
func DefaultConfig() *Config {
	return &Config{
		Connection: ConnectionConfig{
			Port:            3389,
			ConnectTimeout:  10 * time.Second,
			ReadTimeout:     30 * time.Second,
			WriteTimeout:    30 * time.Second,
			KeepAlive:       true,
			KeepAlivePeriod: 30 * time.Second,
			MaxRetries:      3,
			RetryDelay:      1 * time.Second,
		},
		Authentication: AuthConfig{
			UseNLA: true,
			UseSSL: true,
		},
		Display: DisplayConfig{
			Width:          1920,
			Height:         1080,
			ColorDepth:     24,
			HighDPI:        false,
			DesktopScale:   100,
			DeviceScale:    100,
			Wallpaper:      false,
			Themes:         false,
			FontSmoothing:  true,
			FullWindowDrag: true,
			MenuAnimations: false,
		},
		Performance: PerformanceConfig{
			BitmapCache:      true,
			BitmapCacheSize:  1000,
			Compression:      true,
			CompressionLevel: 6,
			NetworkAutoTune:  true,
			FrameRate:        30,
			QualityLevel:     "high",
		},
		Security: SecurityConfig{
			FIPSCompliance:        false,
			CertificateValidation: true,
			EncryptionLevel:       "high",
		},
		Input: InputConfig{
			KeyboardLayout:    "en-US",
			MouseSensitivity:  100,
			MouseAcceleration: false,
			UnicodeInput:      true,
			IMEInput:          true,
			InputBufferSize:   1024,
		},
		VirtualChannels: VirtualChannelConfig{
			Clipboard: true,
			Audio:     true,
			Device:    true,
			Printer:   false,
			Drive:     false,
			Port:      false,
			SmartCard: false,
		},
		Logging: LoggingConfig{
			Level:      "info",
			Format:     "json",
			Output:     "stdout",
			MaxSize:    100,
			MaxBackups: 3,
			MaxAge:     28,
			Compress:   true,
		},
	}
}

// LoadFromFile loads configuration from a JSON or YAML file
func LoadFromFile(filename string) (*Config, error) {
	data, err := os.ReadFile(filename)
	if err != nil {
		return nil, fmt.Errorf("failed to read config file: %w", err)
	}

	config := DefaultConfig()

	if strings.HasSuffix(filename, ".json") {
		if err := json.Unmarshal(data, config); err != nil {
			return nil, fmt.Errorf("failed to parse JSON config: %w", err)
		}
	} else if strings.HasSuffix(filename, ".yaml") || strings.HasSuffix(filename, ".yml") {
		if err := yaml.Unmarshal(data, config); err != nil {
			return nil, fmt.Errorf("failed to parse YAML config: %w", err)
		}
	} else {
		return nil, fmt.Errorf("unsupported config file format")
	}

	return config, nil
}

// LoadFromEnvironment loads configuration from environment variables
func LoadFromEnvironment() *Config {
	config := DefaultConfig()

	// Connection settings
	if addr := os.Getenv("RDP_ADDRESS"); addr != "" {
		config.Connection.Address = addr
	}
	if port := os.Getenv("RDP_PORT"); port != "" {
		if p, err := strconv.Atoi(port); err == nil {
			config.Connection.Port = p
		}
	}
	if timeout := os.Getenv("RDP_CONNECT_TIMEOUT"); timeout != "" {
		if t, err := time.ParseDuration(timeout); err == nil {
			config.Connection.ConnectTimeout = t
		}
	}

	// Authentication settings
	if username := os.Getenv("RDP_USERNAME"); username != "" {
		config.Authentication.Username = username
	}
	if password := os.Getenv("RDP_PASSWORD"); password != "" {
		config.Authentication.Password = password
	}
	if domain := os.Getenv("RDP_DOMAIN"); domain != "" {
		config.Authentication.Domain = domain
	}

	// Display settings
	if width := os.Getenv("RDP_WIDTH"); width != "" {
		if w, err := strconv.Atoi(width); err == nil {
			config.Display.Width = w
		}
	}
	if height := os.Getenv("RDP_HEIGHT"); height != "" {
		if h, err := strconv.Atoi(height); err == nil {
			config.Display.Height = h
		}
	}

	// Performance settings
	if cache := os.Getenv("RDP_BITMAP_CACHE"); cache != "" {
		config.Performance.BitmapCache = cache == "true"
	}
	if compression := os.Getenv("RDP_COMPRESSION"); compression != "" {
		config.Performance.Compression = compression == "true"
	}

	// Security settings
	if fips := os.Getenv("RDP_FIPS_COMPLIANCE"); fips != "" {
		config.Security.FIPSCompliance = fips == "true"
	}

	// Logging settings
	if level := os.Getenv("RDP_LOG_LEVEL"); level != "" {
		config.Logging.Level = level
	}

	return config
}

// Merge merges another configuration into this one
func (c *Config) Merge(other *Config) {
	if other == nil {
		return
	}

	// Merge connection settings
	if other.Connection.Address != "" {
		c.Connection.Address = other.Connection.Address
	}
	if other.Connection.Port != 0 {
		c.Connection.Port = other.Connection.Port
	}
	if other.Connection.ConnectTimeout != 0 {
		c.Connection.ConnectTimeout = other.Connection.ConnectTimeout
	}

	// Merge authentication settings
	if other.Authentication.Username != "" {
		c.Authentication.Username = other.Authentication.Username
	}
	if other.Authentication.Password != "" {
		c.Authentication.Password = other.Authentication.Password
	}
	if other.Authentication.Domain != "" {
		c.Authentication.Domain = other.Authentication.Domain
	}

	// Merge display settings
	if other.Display.Width != 0 {
		c.Display.Width = other.Display.Width
	}
	if other.Display.Height != 0 {
		c.Display.Height = other.Display.Height
	}
	if len(other.Display.Monitors) > 0 {
		c.Display.Monitors = other.Display.Monitors
	}

	// Merge other settings as needed...
}

// Validate validates the configuration
func (c *Config) Validate() error {
	if c.Connection.Address == "" {
		return fmt.Errorf("connection address is required")
	}
	if c.Connection.Port <= 0 || c.Connection.Port > 65535 {
		return fmt.Errorf("invalid port number: %d", c.Connection.Port)
	}
	if c.Authentication.Username == "" {
		return fmt.Errorf("username is required")
	}
	if c.Display.Width <= 0 || c.Display.Height <= 0 {
		return fmt.Errorf("invalid display dimensions: %dx%d", c.Display.Width, c.Display.Height)
	}
	if c.Performance.CompressionLevel < 0 || c.Performance.CompressionLevel > 9 {
		return fmt.Errorf("invalid compression level: %d", c.Performance.CompressionLevel)
	}
	return nil
}

// ToMap converts the configuration to a map for easy access
func (c *Config) ToMap() map[string]interface{} {
	data, _ := json.Marshal(c)
	var result map[string]interface{}
	json.Unmarshal(data, &result)
	return result
}

// GetString returns a string value from the configuration
func (c *Config) GetString(path string) (string, error) {
	parts := strings.Split(path, ".")
	current := c.ToMap()

	for i, part := range parts {
		if i == len(parts)-1 {
			if val, ok := current[part].(string); ok {
				return val, nil
			}
			return "", fmt.Errorf("path %s does not point to a string value", path)
		}

		if next, ok := current[part].(map[string]interface{}); ok {
			current = next
		} else {
			return "", fmt.Errorf("invalid path: %s", path)
		}
	}

	return "", fmt.Errorf("path not found: %s", path)
}

// GetInt returns an integer value from the configuration
func (c *Config) GetInt(path string) (int, error) {
	parts := strings.Split(path, ".")
	current := c.ToMap()

	for i, part := range parts {
		if i == len(parts)-1 {
			if val, ok := current[part].(float64); ok {
				return int(val), nil
			}
			return 0, fmt.Errorf("path %s does not point to a numeric value", path)
		}

		if next, ok := current[part].(map[string]interface{}); ok {
			current = next
		} else {
			return 0, fmt.Errorf("invalid path: %s", path)
		}
	}

	return 0, fmt.Errorf("path not found: %s", path)
}

// GetBool returns a boolean value from the configuration
func (c *Config) GetBool(path string) (bool, error) {
	parts := strings.Split(path, ".")
	current := c.ToMap()

	for i, part := range parts {
		if i == len(parts)-1 {
			if val, ok := current[part].(bool); ok {
				return val, nil
			}
			return false, fmt.Errorf("path %s does not point to a boolean value", path)
		}

		if next, ok := current[part].(map[string]interface{}); ok {
			current = next
		} else {
			return false, fmt.Errorf("invalid path: %s", path)
		}
	}

	return false, fmt.Errorf("path not found: %s", path)
}
