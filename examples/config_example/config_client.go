package main

import (
	"context"
	"fmt"
	"log"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/kdsmith18542/gordp"
	"github.com/kdsmith18542/gordp/config"
	"github.com/kdsmith18542/gordp/proto/bitmap"
)

// ConfigClient demonstrates the new configuration system
type ConfigClient struct {
	client *gordp.Client
	config *config.Config
	ctx    context.Context
	cancel context.CancelFunc
}

// NewConfigClient creates a new client with configuration
func NewConfigClient(cfg *config.Config) *ConfigClient {
	ctx, cancel := context.WithCancel(context.Background())

	// Convert config to gordp.Option
	option := &gordp.Option{
		Addr:           fmt.Sprintf("%s:%d", cfg.Connection.Address, cfg.Connection.Port),
		UserName:       cfg.Authentication.Username,
		Password:       cfg.Authentication.Password,
		ConnectTimeout: cfg.Connection.ConnectTimeout,
		Monitors:       cfg.Display.Monitors,
	}

	client := gordp.NewClientWithContext(ctx, option)

	return &ConfigClient{
		client: client,
		config: cfg,
		ctx:    ctx,
		cancel: cancel,
	}
}

// Connect establishes the RDP connection with context
func (c *ConfigClient) Connect() error {
	log.Printf("Connecting to %s with configuration...", c.config.Connection.Address)

	// Validate configuration
	if err := c.config.Validate(); err != nil {
		return fmt.Errorf("configuration validation failed: %w", err)
	}

	// Connect with context
	err := c.client.ConnectWithContext(c.ctx)
	if err != nil {
		return fmt.Errorf("connection failed: %w", err)
	}

	log.Println("Successfully connected to RDP server")
	return nil
}

// Run starts the RDP session with context
func (c *ConfigClient) Run() error {
	log.Println("Starting RDP session with configuration...")

	// Set up signal handling for graceful shutdown
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)

	// Start a goroutine to handle signals
	go func() {
		<-sigChan
		log.Println("Received shutdown signal, closing connection...")
		c.Close()
		os.Exit(0)
	}()

	// Create bitmap processor with configuration
	processor := &ConfigBitmapProcessor{
		config: c.config,
	}

	// Run the session with context
	err := c.client.RunWithContext(c.ctx, processor)
	if err != nil {
		return fmt.Errorf("session failed: %w", err)
	}

	return nil
}

// Close closes the RDP connection
func (c *ConfigClient) Close() {
	c.cancel()
	c.client.Close()
	log.Println("RDP connection closed")
}

// GetConfigValue demonstrates accessing configuration values
func (c *ConfigClient) GetConfigValue(path string) {
	if val, err := c.config.GetString(path); err == nil {
		log.Printf("Config %s: %s", path, val)
	} else if val, err := c.config.GetInt(path); err == nil {
		log.Printf("Config %s: %d", path, val)
	} else if val, err := c.config.GetBool(path); err == nil {
		log.Printf("Config %s: %t", path, val)
	} else {
		log.Printf("Config %s: not found", path)
	}
}

// ConfigBitmapProcessor processes bitmaps with configuration
type ConfigBitmapProcessor struct {
	config     *config.Config
	frameCount int
	startTime  time.Time
}

// ProcessBitmap implements the bitmap processor interface
func (p *ConfigBitmapProcessor) ProcessBitmap(option *bitmap.Option, bitmap *bitmap.BitMap) {
	p.frameCount++

	// Log frame information based on configuration
	if p.frameCount == 1 {
		p.startTime = time.Now()
		log.Printf("First frame received: %dx%d at (%d,%d)",
			option.Width, option.Height, option.Left, option.Top)
	}

	// Log performance information based on configuration
	if p.config.Performance.FrameRate > 0 && p.frameCount%p.config.Performance.FrameRate == 0 {
		duration := time.Since(p.startTime)
		fps := float64(p.frameCount) / duration.Seconds()

		log.Printf("Performance: Frame %d, FPS: %.2f, Quality: %s",
			p.frameCount, fps, p.config.Performance.QualityLevel)
	}

	// Apply performance settings
	if !p.config.Performance.BitmapCache {
		// Skip caching if disabled
		log.Printf("Bitmap caching disabled, processing frame %d", p.frameCount)
	}

	// Apply display settings
	if p.config.Display.HighDPI {
		// Apply high DPI scaling
		log.Printf("High DPI mode: scaling bitmap %dx%d", option.Width, option.Height)
	}
}

func main() {
	// Load configuration from file
	cfg, err := config.LoadFromFile("config.json")
	if err != nil {
		log.Printf("Failed to load config file: %v", err)
		log.Println("Using default configuration...")
		cfg = config.DefaultConfig()

		// Set required fields
		if len(os.Args) >= 4 {
			cfg.Connection.Address = os.Args[1]
			cfg.Authentication.Username = os.Args[2]
			cfg.Authentication.Password = os.Args[3]
		} else {
			log.Fatal("Usage: config_client <server> <username> <password>")
		}
	}

	// Create client with configuration
	client := NewConfigClient(cfg)
	defer client.Close()

	// Demonstrate configuration access
	log.Println("Configuration values:")
	client.GetConfigValue("connection.address")
	client.GetConfigValue("display.width")
	client.GetConfigValue("performance.bitmap_cache")
	client.GetConfigValue("security.fips_compliance")

	// Connect to server
	err = client.Connect()
	if err != nil {
		log.Fatalf("Failed to connect: %v", err)
	}

	// Run the session
	log.Println("Starting configured RDP session...")
	err = client.Run()
	if err != nil {
		log.Fatalf("Session failed: %v", err)
	}
}
