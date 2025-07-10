package main

import (
	"fmt"
	"log"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/kdsmith18542/gordp"
	"github.com/kdsmith18542/gordp/proto/audio"
	"github.com/kdsmith18542/gordp/proto/bitmap"
	"github.com/kdsmith18542/gordp/proto/clipboard"
	"github.com/kdsmith18542/gordp/proto/device"
	"github.com/kdsmith18542/gordp/proto/mcs"
	"github.com/kdsmith18542/gordp/proto/t128"
)

// ComprehensiveRDPClient demonstrates all major features of GoRDP
type ComprehensiveRDPClient struct {
	client     *gordp.Client
	frameCount int
	startTime  time.Time
}

// NewComprehensiveRDPClient creates a new comprehensive RDP client
func NewComprehensiveRDPClient(addr, username, password string) *ComprehensiveRDPClient {
	// Configure multi-monitor setup
	monitors := []mcs.MonitorLayout{
		{
			Left:               0,
			Top:                0,
			Right:              1920,
			Bottom:             1080,
			Flags:              0x01, // Primary monitor
			MonitorIndex:       0,
			PhysicalWidthMm:    520,
			PhysicalHeightMm:   290,
			Orientation:        0, // Landscape
			DesktopScaleFactor: 100,
			DeviceScaleFactor:  100,
		},
		{
			Left:               1920,
			Top:                0,
			Right:              3840,
			Bottom:             1080,
			Flags:              0x00, // Secondary monitor
			MonitorIndex:       1,
			PhysicalWidthMm:    520,
			PhysicalHeightMm:   290,
			Orientation:        0, // Landscape
			DesktopScaleFactor: 100,
			DeviceScaleFactor:  100,
		},
	}

	client := gordp.NewClient(&gordp.Option{
		Addr:           addr,
		UserName:       username,
		Password:       password,
		ConnectTimeout: 10 * time.Second,
		Monitors:       monitors,
	})

	return &ComprehensiveRDPClient{
		client:    client,
		startTime: time.Now(),
	}
}

// Connect establishes the RDP connection
func (c *ComprehensiveRDPClient) Connect() error {
	log.Println("Connecting to RDP server...")

	// Register all handlers before connecting
	c.registerHandlers()

	err := c.client.Connect()
	if err != nil {
		return fmt.Errorf("connection failed: %w", err)
	}

	log.Println("Successfully connected to RDP server")
	return nil
}

// registerHandlers registers all event handlers
func (c *ComprehensiveRDPClient) registerHandlers() {
	// Register clipboard handler
	c.client.RegisterClipboardHandler(&ComprehensiveClipboardHandler{})

	// Register device handler
	c.client.RegisterDeviceHandler(&ComprehensiveDeviceHandler{})

	// Register dynamic virtual channel handler
	c.client.RegisterDynamicVirtualChannelHandler("COMPREHENSIVE_CHANNEL", &ComprehensiveDynamicVirtualChannelHandler{})
}

// Run starts the RDP session
func (c *ComprehensiveRDPClient) Run() error {
	log.Println("Starting RDP session...")

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

	// Run the session with our bitmap processor
	err := c.client.Run(c)
	if err != nil {
		return fmt.Errorf("session failed: %w", err)
	}

	return nil
}

// Close closes the RDP connection
func (c *ComprehensiveRDPClient) Close() {
	if c.client != nil {
		c.client.Close()
	}
	log.Println("RDP connection closed")
}

// ProcessBitmap implements the bitmap processor interface
func (c *ComprehensiveRDPClient) ProcessBitmap(option *bitmap.Option, bitmap *bitmap.BitMap) {
	c.frameCount++

	// Log frame information every 100 frames
	if c.frameCount%100 == 0 {
		duration := time.Since(c.startTime)
		fps := float64(c.frameCount) / duration.Seconds()

		log.Printf("Frame %d: %dx%d bitmap at (%d,%d) - FPS: %.2f",
			c.frameCount, option.Width, option.Height, option.Left, option.Top, fps)

		// Log cache statistics
		stats := c.client.GetBitmapCacheStats()
		log.Printf("Cache stats: %v", stats)
	}

	// Save every 1000th frame for debugging
	if c.frameCount%1000 == 0 {
		data := bitmap.ToPng()
		filename := fmt.Sprintf("frame_%d_%d_%d.png", c.frameCount, option.Left, option.Top)
		if err := os.WriteFile(filename, data, 0644); err != nil {
			log.Printf("Failed to save frame: %v", err)
		} else {
			log.Printf("Saved frame to %s", filename)
		}
	}
}

// DemonstrateInputHandling shows various input capabilities
func (c *ComprehensiveRDPClient) DemonstrateInputHandling() {
	log.Println("Demonstrating input handling...")

	// Wait a moment for the session to stabilize
	time.Sleep(2 * time.Second)

	// Basic keyboard input
	log.Println("Sending basic keyboard input...")
	c.client.SendString("Hello from GoRDP Comprehensive Client!")
	c.client.SendKeyPress(t128.VK_RETURN, t128.ModifierKey{})

	// Special keys
	log.Println("Sending special keys...")
	c.client.SendSpecialKey("F1", t128.ModifierKey{})
	time.Sleep(500 * time.Millisecond)
	c.client.SendSpecialKey("F2", t128.ModifierKey{})

	// Key combinations
	log.Println("Sending key combinations...")
	c.client.SendCtrlKey('a') // Select all
	time.Sleep(200 * time.Millisecond)
	c.client.SendCtrlKey('c') // Copy
	time.Sleep(200 * time.Millisecond)

	// Function keys
	log.Println("Sending function keys...")
	for i := 1; i <= 5; i++ {
		c.client.SendFunctionKey(i, t128.ModifierKey{})
		time.Sleep(200 * time.Millisecond)
	}

	// Arrow keys
	log.Println("Sending arrow keys...")
	directions := []string{"up", "down", "left", "right"}
	for _, dir := range directions {
		c.client.SendArrowKey(dir, t128.ModifierKey{})
		time.Sleep(200 * time.Millisecond)
	}

	// Navigation keys
	log.Println("Sending navigation keys...")
	navKeys := []string{"home", "end", "pageup", "pagedown"}
	for _, key := range navKeys {
		c.client.SendNavigationKey(key, t128.ModifierKey{})
		time.Sleep(200 * time.Millisecond)
	}

	// Unicode input
	log.Println("Sending Unicode input...")
	c.client.SendUnicodeString("Hello 世界! 🌍")
	c.client.SendKeyPress(t128.VK_RETURN, t128.ModifierKey{})

	// Mouse input
	log.Println("Demonstrating mouse input...")
	c.client.SendMouseMoveEvent(500, 300)
	time.Sleep(500 * time.Millisecond)

	c.client.SendMouseClickEvent(t128.MouseButtonLeft, 500, 300)
	time.Sleep(500 * time.Millisecond)

	c.client.SendMouseClickEvent(t128.MouseButtonRight, 600, 400)
	time.Sleep(500 * time.Millisecond)

	// Mouse wheel
	log.Println("Sending mouse wheel events...")
	c.client.SendMouseWheelEvent(120, 500, 300) // Scroll up
	time.Sleep(200 * time.Millisecond)
	c.client.SendMouseWheelEvent(-120, 500, 300) // Scroll down
	time.Sleep(200 * time.Millisecond)

	// Advanced mouse features
	log.Println("Demonstrating advanced mouse features...")
	c.client.SendMouseDoubleClickEvent(t128.MouseButtonLeft, 700, 500)
	time.Sleep(500 * time.Millisecond)

	c.client.SendMouseDragEvent(t128.MouseButtonLeft, 100, 100, 200, 200)
	time.Sleep(500 * time.Millisecond)

	// Key sequences
	log.Println("Sending key sequences...")
	keys := []uint8{t128.VK_H, t128.VK_E, t128.VK_L, t128.VK_L, t128.VK_O}
	c.client.SendKeySequence(keys, t128.ModifierKey{})
	c.client.SendKeyPress(t128.VK_RETURN, t128.ModifierKey{})

	// Key with delay
	log.Println("Sending key with delay...")
	c.client.SendKeyWithDelay('x', 100, t128.ModifierKey{})

	// Key repeat
	log.Println("Sending repeated keys...")
	c.client.SendKeyRepeat('y', 3, t128.ModifierKey{})

	log.Println("Input demonstration completed")
}

// DemonstrateMultiMonitor shows multi-monitor capabilities
func (c *ComprehensiveRDPClient) DemonstrateMultiMonitor() {
	log.Println("Demonstrating multi-monitor support...")

	// Get current monitor configuration
	monitors := c.client.GetMonitors()
	log.Printf("Current monitors: %d", len(monitors))

	for i, monitor := range monitors {
		log.Printf("Monitor %d: (%d,%d) to (%d,%d), flags: 0x%x",
			i, monitor.Left, monitor.Top, monitor.Right, monitor.Bottom, monitor.Flags)
	}

	// Send input to different monitors
	if len(monitors) > 1 {
		// Send input to primary monitor
		c.client.SendMouseMoveEvent(500, 300)
		c.client.SendString("Primary monitor")
		c.client.SendKeyPress(t128.VK_RETURN, t128.ModifierKey{})

		// Send input to secondary monitor
		c.client.SendMouseMoveEvent(2500, 300) // 1920 + 580
		c.client.SendString("Secondary monitor")
		c.client.SendKeyPress(t128.VK_RETURN, t128.ModifierKey{})
	}
}

// DemonstrateVirtualChannels shows virtual channel capabilities
func (c *ComprehensiveRDPClient) DemonstrateVirtualChannels() {
	log.Println("Demonstrating virtual channels...")

	// Check channel status
	if c.client.IsClipboardChannelOpen() {
		log.Println("Clipboard channel is open")
	}

	if c.client.IsDeviceChannelOpen() {
		log.Println("Device channel is open")
	}

	// List dynamic virtual channels
	channels := c.client.ListDynamicVirtualChannels()
	log.Printf("Dynamic virtual channels: %v", channels)

	// Send data to virtual channel
	err := c.client.SendVirtualChannelData("COMPREHENSIVE_CHANNEL", []byte("Hello from comprehensive client!"), 0)
	if err != nil {
		log.Printf("Failed to send virtual channel data: %v", err)
	}
}

// DemonstratePerformanceMonitoring shows performance monitoring capabilities
func (c *ComprehensiveRDPClient) DemonstratePerformanceMonitoring() {
	log.Println("Demonstrating performance monitoring...")

	// Get bitmap cache statistics
	stats := c.client.GetBitmapCacheStats()
	log.Printf("Bitmap cache statistics: %v", stats)

	// Get device statistics
	deviceStats := c.client.GetDeviceStats()
	log.Printf("Device statistics: %v", deviceStats)

	// Monitor performance over time
	go func() {
		ticker := time.NewTicker(5 * time.Second)
		defer ticker.Stop()

		for range ticker.C {
			stats := c.client.GetBitmapCacheStats()
			log.Printf("Performance update - Cache stats: %v", stats)
		}
	}()
}

// ComprehensiveClipboardHandler handles clipboard events
type ComprehensiveClipboardHandler struct{}

func (h *ComprehensiveClipboardHandler) OnFormatList(formats []clipboard.ClipboardFormat) error {
	log.Printf("Clipboard formats available: %d", len(formats))
	for i, format := range formats {
		log.Printf("  Format %d: %d", i, format)
	}
	return nil
}

func (h *ComprehensiveClipboardHandler) OnFormatDataRequest(formatID clipboard.ClipboardFormat) error {
	log.Printf("Clipboard data requested for format: %d", formatID)
	return nil
}

func (h *ComprehensiveClipboardHandler) OnFormatDataResponse(formatID clipboard.ClipboardFormat, data []byte) error {
	log.Printf("Clipboard data received: format=%d, size=%d bytes", formatID, len(data))

	// Handle different clipboard formats
	switch formatID {
	case clipboard.CLIPRDR_FORMAT_UNICODETEXT:
		log.Printf("Text data: %s", string(data))
	case clipboard.CLIPRDR_FORMAT_HTML:
		log.Printf("HTML data: %s", string(data))
	case clipboard.CLIPRDR_FORMAT_RAW_BITMAP:
		log.Printf("Bitmap data: %d bytes", len(data))
	default:
		log.Printf("Unknown format data: %d bytes", len(data))
	}

	return nil
}

func (h *ComprehensiveClipboardHandler) OnFileContentsRequest(streamID uint32, listIndex uint32, dwFlags uint32, nPositionLow uint32, nPositionHigh uint32, cbRequested uint32, clipDataID uint32) error {
	log.Printf("File contents request: streamID=%d, listIndex=%d, flags=%d, position=%d, requested=%d, clipDataID=%d",
		streamID, listIndex, dwFlags, (uint64(nPositionHigh)<<32)|uint64(nPositionLow), cbRequested, clipDataID)
	return nil
}

// ComprehensiveAudioHandler handles audio events
type ComprehensiveAudioHandler struct{}

func (h *ComprehensiveAudioHandler) OnAudioData(formatID uint16, data []byte, timestamp uint32) error {
	log.Printf("Audio data: format=%d, size=%d, timestamp=%d", formatID, len(data), timestamp)
	return nil
}

func (h *ComprehensiveAudioHandler) OnAudioFormatList(formats []audio.AudioFormat) error {
	log.Printf("Audio formats available: %d", len(formats))
	for i, format := range formats {
		log.Printf("  Format %d: %d channels, %d Hz, %d bits",
			i, format.Channels, format.SamplesPerSec, format.BitsPerSample)
	}
	return nil
}

func (h *ComprehensiveAudioHandler) OnAudioFormatConfirm(formatID uint16) error {
	log.Printf("Audio format confirmed: %d", formatID)
	return nil
}

// ComprehensiveDeviceHandler handles device events
type ComprehensiveDeviceHandler struct{}

func (h *ComprehensiveDeviceHandler) OnDeviceAnnounce(device *device.DeviceAnnounce) error {
	log.Printf("Device announced: type=%d, name=%s", device.DeviceType, device.PreferredDosName)
	return nil
}

func (h *ComprehensiveDeviceHandler) OnDeviceIORequest(request *device.DeviceIORequest) (*device.DeviceIOCompletion, error) {
	log.Printf("Device I/O request: device=%d, major=%d, minor=%d",
		request.DeviceID, request.MajorFunction, request.MinorFunction)

	return &device.DeviceIOCompletion{
		DeviceID:     request.DeviceID,
		CompletionID: request.CompletionID,
		IoStatus:     0, // STATUS_SUCCESS
		Data:         []byte{},
	}, nil
}

func (h *ComprehensiveDeviceHandler) OnPrinterData(data *device.PrinterData) error {
	log.Printf("Printer data: job=%d, size=%d bytes", data.JobID, len(data.Data))
	return nil
}

func (h *ComprehensiveDeviceHandler) OnDriveAccess(path string, operation string) error {
	log.Printf("Drive access: %s %s", operation, path)
	return nil
}

func (h *ComprehensiveDeviceHandler) OnPortAccess(portName string, operation string) error {
	log.Printf("Port access: %s %s", operation, portName)
	return nil
}

// ComprehensiveDynamicVirtualChannelHandler handles dynamic virtual channel events
type ComprehensiveDynamicVirtualChannelHandler struct{}

func (h *ComprehensiveDynamicVirtualChannelHandler) OnChannelCreated(channelId uint32, channelName string) error {
	log.Printf("Dynamic virtual channel created: %s (ID: %d)", channelName, channelId)
	return nil
}

func (h *ComprehensiveDynamicVirtualChannelHandler) OnChannelOpened(channelId uint32) error {
	log.Printf("Dynamic virtual channel opened: ID: %d", channelId)
	return nil
}

func (h *ComprehensiveDynamicVirtualChannelHandler) OnChannelClosed(channelId uint32) error {
	log.Printf("Dynamic virtual channel closed: ID: %d", channelId)
	return nil
}

func (h *ComprehensiveDynamicVirtualChannelHandler) OnDataReceived(channelId uint32, data []byte) error {
	log.Printf("Received data on comprehensive channel: %d bytes", len(data))
	return nil
}

func main() {
	// Check command line arguments
	if len(os.Args) != 4 {
		fmt.Println("Usage: comprehensive_client <server:port> <username> <password>")
		fmt.Println("Example: comprehensive_client 192.168.1.100:3389 administrator password")
		os.Exit(1)
	}

	addr := os.Args[1]
	username := os.Args[2]
	password := os.Args[3]

	// Create comprehensive client
	client := NewComprehensiveRDPClient(addr, username, password)
	defer client.Close()

	// Connect to server
	err := client.Connect()
	if err != nil {
		log.Fatalf("Failed to connect: %v", err)
	}

	// Demonstrate features
	go func() {
		time.Sleep(3 * time.Second) // Wait for session to stabilize

		client.DemonstrateInputHandling()
		time.Sleep(2 * time.Second)

		client.DemonstrateMultiMonitor()
		time.Sleep(2 * time.Second)

		client.DemonstrateVirtualChannels()
		time.Sleep(2 * time.Second)

		client.DemonstratePerformanceMonitoring()
	}()

	// Run the session
	log.Println("Starting comprehensive RDP session...")
	err = client.Run()
	if err != nil {
		log.Fatalf("Session failed: %v", err)
	}
}
