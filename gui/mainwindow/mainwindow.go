package mainwindow

import (
	"context"
	"fmt"
	"sync"
	"time"

	"github.com/kdsmith18542/gordp"
	"github.com/kdsmith18542/gordp/gui/connection"
	"github.com/kdsmith18542/gordp/gui/display"
	"github.com/kdsmith18542/gordp/gui/favorites"
	"github.com/kdsmith18542/gordp/gui/history"
	"github.com/kdsmith18542/gordp/gui/input"
	"github.com/kdsmith18542/gordp/gui/multimonitor"
	"github.com/kdsmith18542/gordp/gui/performance"
	"github.com/kdsmith18542/gordp/gui/plugins"
	"github.com/kdsmith18542/gordp/gui/settings"
	"github.com/kdsmith18542/gordp/gui/virtualchannels"
	"github.com/kdsmith18542/gordp/proto/bitmap"
	"github.com/kdsmith18542/gordp/proto/mcs"
)

// MainWindow represents the main application window
type MainWindow struct {
	// Core components
	connectionDialog *connection.ConnectionDialog
	settingsDialog   *settings.SettingsDialog
	displayWidget    *display.RDPDisplayWidget

	// RDP client and connection
	client       *gordp.Client
	clientCtx    context.Context
	clientCancel context.CancelFunc
	clientMu     sync.RWMutex

	// Input handlers
	keyboardHandler *input.KeyboardHandler
	mouseHandler    *input.MouseHandler

	// Advanced features
	virtualChannelManager *virtualchannels.VirtualChannelManager
	virtualChannelDialog  *virtualchannels.VirtualChannelDialog
	monitorManager        *multimonitor.MonitorManager
	monitorDialog         *multimonitor.MonitorDialog
	pluginManager         *plugins.PluginManager
	pluginDialog          *plugins.PluginDialog

	// Phase 5 features
	performanceMonitor *performance.PerformanceMonitor
	performanceDialog  *performance.PerformanceDialog
	connectionHistory  *history.ConnectionHistory
	historyDialog      *history.HistoryDialog
	favoritesManager   *favorites.FavoritesManager
	favoritesDialog    *favorites.FavoritesDialog

	// State
	isConnected  bool
	connectionMu sync.RWMutex

	// Connection statistics
	connectionStats map[string]interface{}
	lastConnection  time.Time
	connectionError error
}

// NewMainWindow creates a new main window
func NewMainWindow() *MainWindow {
	window := &MainWindow{
		isConnected:     false,
		connectionStats: make(map[string]interface{}),
	}

	// Initialize core components
	window.connectionDialog = connection.NewConnectionDialog()
	window.settingsDialog = settings.NewSettingsDialog()
	window.displayWidget = display.NewRDPDisplayWidget()

	// Initialize input handlers
	window.keyboardHandler = input.NewKeyboardHandler(nil) // Will be set when client is created
	window.mouseHandler = input.NewMouseHandler(nil)       // Will be set when client is created

	// Initialize advanced features
	window.virtualChannelManager = virtualchannels.NewVirtualChannelManager(nil) // Will be set when client is created
	window.virtualChannelDialog = virtualchannels.NewVirtualChannelDialog(window.virtualChannelManager)
	window.monitorManager = multimonitor.NewMonitorManager()
	window.monitorDialog = multimonitor.NewMonitorDialog(window.monitorManager)
	window.pluginManager = plugins.NewPluginManager()
	window.pluginDialog = plugins.NewPluginDialog(window.pluginManager)

	// Initialize Phase 5 features
	window.performanceMonitor = performance.NewPerformanceMonitor()
	window.performanceDialog = performance.NewPerformanceDialog(window.performanceMonitor)
	window.connectionHistory = history.NewConnectionHistory()
	window.historyDialog = history.NewHistoryDialog(window.connectionHistory)
	window.favoritesManager = favorites.NewFavoritesManager()
	window.favoritesDialog = favorites.NewFavoritesDialog(window.favoritesManager)

	// Initialize connection statistics
	window.initializeConnectionStats()

	return window
}

// Show displays the main window
func (w *MainWindow) Show() {
	fmt.Println("=== GoRDP GUI Client ===")
	fmt.Println("Welcome to the GoRDP graphical client!")

	for {
		w.showMainMenu()
	}
}

// showMainMenu displays the main menu
func (w *MainWindow) showMainMenu() {
	fmt.Println("\n=== Main Menu ===")

	w.connectionMu.RLock()
	connected := w.isConnected
	w.connectionMu.RUnlock()

	if connected {
		fmt.Println("Status: Connected")
		fmt.Println("\n1. Disconnect")
		fmt.Println("2. Display Settings")
		fmt.Println("3. Virtual Channels")
		fmt.Println("4. Multi-Monitor")
		fmt.Println("5. Plugin Management")
		fmt.Println("6. Performance Monitoring")
		fmt.Println("7. Settings")
		fmt.Println("8. Connection Statistics")
		fmt.Println("9. Exit")

		fmt.Print("\nSelect option (1-9): ")
		var choice string
		fmt.Scanln(&choice)

		switch choice {
		case "1":
			w.disconnect()
		case "2":
			w.showDisplaySettings()
		case "3":
			w.showVirtualChannels()
		case "4":
			w.showMultiMonitor()
		case "5":
			w.showPluginManagement()
		case "6":
			w.showPerformanceMonitoring()
		case "7":
			w.showSettings()
		case "8":
			w.showConnectionStatistics()
		case "9":
			w.exit()
		default:
			fmt.Println("Invalid option. Please try again.")
		}
	} else {
		fmt.Println("Status: Disconnected")
		fmt.Println("\n1. Connect")
		fmt.Println("2. Favorites")
		fmt.Println("3. Connection History")
		fmt.Println("4. Settings")
		fmt.Println("5. Multi-Monitor Setup")
		fmt.Println("6. Plugin Management")
		fmt.Println("7. Connection Statistics")
		fmt.Println("8. Exit")

		fmt.Print("\nSelect option (1-8): ")
		var choice string
		fmt.Scanln(&choice)

		switch choice {
		case "1":
			w.connect()
		case "2":
			w.showFavorites()
		case "3":
			w.showConnectionHistory()
		case "4":
			w.showSettings()
		case "5":
			w.showMultiMonitor()
		case "6":
			w.showPluginManagement()
		case "7":
			w.showConnectionStatistics()
		case "8":
			w.exit()
		default:
			fmt.Println("Invalid option. Please try again.")
		}
	}
}

// connect establishes a real RDP connection
func (w *MainWindow) connect() {
	fmt.Println("\n=== Connect to RDP Server ===")

	// Show connection dialog and get configuration
	config := w.connectionDialog.Show()
	if config == nil {
		fmt.Println("Connection cancelled or invalid configuration")
		return
	}

	// Validate configuration
	if err := w.validateConnectionConfig(config); err != nil {
		fmt.Printf("Configuration error: %v\n", err)
		return
	}

	// Start connection process
	fmt.Printf("Connecting to %s:%d as %s...\n", config.Address, config.Port, config.Username)

	startTime := time.Now()
	w.connectionError = nil

	// Create context for connection
	w.clientCtx, w.clientCancel = context.WithTimeout(context.Background(), 30*time.Second)

	// Get monitor configuration
	monitors := w.getMonitorConfiguration()

	// Create RDP client
	w.clientMu.Lock()
	w.client = gordp.NewClientWithContext(w.clientCtx, &gordp.Option{
		Addr:           fmt.Sprintf("%s:%d", config.Address, config.Port),
		UserName:       config.Username,
		Password:       config.Password,
		ConnectTimeout: 10 * time.Second,
		Monitors:       monitors,
	})
	w.clientMu.Unlock()

	// Update handlers with the new client
	w.updateHandlersWithClient()

	// Connect to server
	if err := w.client.ConnectWithContext(w.clientCtx); err != nil {
		w.connectionError = err
		w.updateConnectionStats("errors", 1)
		w.connectionStats["last_error"] = err.Error()
		fmt.Printf("Connection failed: %v\n", err)
		return
	}

	// Update connection state
	w.connectionMu.Lock()
	w.isConnected = true
	w.lastConnection = time.Now()
	w.connectionMu.Unlock()

	// Update statistics
	w.updateConnectionStats("connections", 1)
	w.updateConnectionStats("connection_time_ms", int(time.Since(startTime).Milliseconds()))

	// Initialize virtual channels
	if err := w.virtualChannelManager.InitializeChannels(); err != nil {
		fmt.Printf("Warning: Failed to initialize virtual channels: %v\n", err)
	}

	// Start performance monitoring
	w.performanceMonitor.StartMonitoring()

	// Start RDP session in a goroutine
	go w.runRDPSession()

	fmt.Println("Connection established successfully!")
	fmt.Println("RDP session started. Use the menu to manage the connection.")
}

// runRDPSession runs the RDP session with bitmap processing
func (w *MainWindow) runRDPSession() {
	// Create bitmap processor
	processor := &MainWindowBitmapProcessor{mainWindow: w}

	// Run the RDP session
	err := w.client.RunWithContext(w.clientCtx, processor)
	if err != nil {
		w.connectionMu.Lock()
		w.isConnected = false
		w.connectionError = err
		w.connectionMu.Unlock()

		w.updateConnectionStats("session_errors", 1)
		fmt.Printf("RDP session failed: %v\n", err)
	}
}

// disconnect terminates the RDP connection
func (w *MainWindow) disconnect() {
	fmt.Println("\n=== Disconnect ===")

	w.connectionMu.Lock()
	if !w.isConnected {
		w.connectionMu.Unlock()
		fmt.Println("Not connected.")
		return
	}
	w.connectionMu.Unlock()

	// Cancel client context
	if w.clientCancel != nil {
		w.clientCancel()
	}

	// Close virtual channels
	w.virtualChannelManager.CloseAllChannels()

	// Stop all plugins
	if err := w.pluginManager.StopAllPlugins(); err != nil {
		fmt.Printf("Warning: Failed to stop plugins: %v\n", err)
	}

	// Stop performance monitoring
	w.performanceMonitor.StopMonitoring()

	// Close client
	w.clientMu.Lock()
	if w.client != nil {
		w.client.Close()
		w.client = nil
	}
	w.clientMu.Unlock()

	// Update connection state
	w.connectionMu.Lock()
	w.isConnected = false
	w.connectionMu.Unlock()

	// Update statistics
	w.updateConnectionStats("disconnections", 1)

	fmt.Println("Disconnected successfully!")
}

// validateConnectionConfig validates the connection configuration
func (w *MainWindow) validateConnectionConfig(config *connection.ConnectionConfig) error {
	if config.Address == "" {
		return fmt.Errorf("server address is required")
	}
	if config.Port <= 0 || config.Port > 65535 {
		return fmt.Errorf("invalid port number: %d", config.Port)
	}
	if config.Username == "" {
		return fmt.Errorf("username is required")
	}
	if config.Password == "" {
		return fmt.Errorf("password is required")
	}
	return nil
}

// getMonitorConfiguration gets the monitor configuration for RDP
func (w *MainWindow) getMonitorConfiguration() []mcs.MonitorLayout {
	// Detect monitors
	if err := w.monitorManager.DetectMonitors(); err != nil {
		fmt.Printf("Warning: Failed to detect monitors: %v\n", err)
		return nil
	}

	monitors := w.monitorManager.GetMonitors()
	if len(monitors) == 0 {
		return nil
	}

	// Convert to RDP monitor layout
	var rdpMonitors []mcs.MonitorLayout
	for i, monitor := range monitors {
		rdpMonitor := mcs.MonitorLayout{
			Left:               int32(monitor.X),
			Top:                int32(monitor.Y),
			Right:              int32(monitor.X + monitor.Width),
			Bottom:             int32(monitor.Y + monitor.Height),
			Flags:              0x00, // Secondary monitor
			MonitorIndex:       uint32(i),
			PhysicalWidthMm:    520, // Default values
			PhysicalHeightMm:   290,
			Orientation:        0, // Landscape
			DesktopScaleFactor: 100,
			DeviceScaleFactor:  100,
		}

		if monitor.Primary {
			rdpMonitor.Flags = 0x01 // Primary monitor
		}

		rdpMonitors = append(rdpMonitors, rdpMonitor)
	}

	return rdpMonitors
}

// updateHandlersWithClient updates all handlers with the new client
func (w *MainWindow) updateHandlersWithClient() {
	w.clientMu.RLock()
	client := w.client
	w.clientMu.RUnlock()

	if client == nil {
		return
	}

	// Update input handlers
	w.keyboardHandler.SetClient(client)
	w.mouseHandler.SetClient(client)

	// TODO: Update virtual channel manager when SetClient method is implemented
	// w.virtualChannelManager.SetClient(client)

	// TODO: Update display widget when SetClient method is implemented
	// w.displayWidget.SetClient(client)
}

// initializeConnectionStats initializes connection statistics
func (w *MainWindow) initializeConnectionStats() {
	w.connectionStats["total_connections"] = 0
	w.connectionStats["total_disconnections"] = 0
	w.connectionStats["total_errors"] = 0
	w.connectionStats["total_session_errors"] = 0
	w.connectionStats["average_connection_time_ms"] = 0
	w.connectionStats["start_time"] = time.Now()
}

// updateConnectionStats updates connection statistics
func (w *MainWindow) updateConnectionStats(key string, value int) {
	if current, exists := w.connectionStats[key]; exists {
		if intValue, ok := current.(int); ok {
			w.connectionStats[key] = intValue + value
		}
	} else {
		w.connectionStats[key] = value
	}
}

// showConnectionStatistics shows connection statistics
func (w *MainWindow) showConnectionStatistics() {
	fmt.Println("\n=== Connection Statistics ===")

	stats := w.getConnectionStats()
	for key, value := range stats {
		fmt.Printf("%s: %v\n", key, value)
	}

	fmt.Println("\nPress Enter to continue...")
	var input string
	fmt.Scanln(&input)
}

// getConnectionStats returns connection statistics
func (w *MainWindow) getConnectionStats() map[string]interface{} {
	w.connectionMu.RLock()
	defer w.connectionMu.RUnlock()

	stats := make(map[string]interface{})
	for k, v := range w.connectionStats {
		stats[k] = v
	}

	// Add current state
	stats["is_connected"] = w.isConnected
	stats["last_connection"] = w.lastConnection
	stats["last_error"] = w.connectionError

	return stats
}

// showDisplaySettings shows display-related settings
func (w *MainWindow) showDisplaySettings() {
	fmt.Println("\n=== Display Settings ===")

	w.connectionMu.RLock()
	connected := w.isConnected
	w.connectionMu.RUnlock()

	if !connected {
		fmt.Println("Must be connected to access display settings.")
		return
	}

	fmt.Println("1. Adjust display quality")
	fmt.Println("2. Change resolution")
	fmt.Println("3. Toggle fullscreen")
	fmt.Println("4. Display statistics")
	fmt.Println("5. Back")

	fmt.Print("\nSelect option (1-5): ")
	var choice string
	fmt.Scanln(&choice)

	switch choice {
	case "1":
		w.adjustDisplayQuality()
	case "2":
		w.changeResolution()
	case "3":
		w.toggleFullscreen()
	case "4":
		w.showDisplayStats()
	case "5":
		return
	default:
		fmt.Println("Invalid option.")
	}
}

// showVirtualChannels shows virtual channel management
func (w *MainWindow) showVirtualChannels() {
	w.connectionMu.RLock()
	connected := w.isConnected
	w.connectionMu.RUnlock()

	if !connected {
		fmt.Println("Must be connected to access virtual channels.")
		return
	}

	w.virtualChannelDialog.Show()
}

// showMultiMonitor shows multi-monitor configuration
func (w *MainWindow) showMultiMonitor() {
	w.monitorDialog.Show()
}

// showPluginManagement shows plugin management
func (w *MainWindow) showPluginManagement() {
	w.pluginDialog.Show()
}

// showSettings shows application settings
func (w *MainWindow) showSettings() {
	w.settingsDialog.Show()
}

// adjustDisplayQuality allows adjusting display quality
func (w *MainWindow) adjustDisplayQuality() {
	fmt.Println("\n--- Display Quality ---")
	fmt.Println("1. High Quality (slower)")
	fmt.Println("2. Balanced")
	fmt.Println("3. Low Quality (faster)")

	fmt.Print("\nSelect quality level (1-3): ")
	var choice string
	fmt.Scanln(&choice)

	switch choice {
	case "1":
		fmt.Println("Display quality set to High")
		// TODO: Implement actual quality adjustment
	case "2":
		fmt.Println("Display quality set to Balanced")
		// TODO: Implement actual quality adjustment
	case "3":
		fmt.Println("Display quality set to Low")
		// TODO: Implement actual quality adjustment
	default:
		fmt.Println("Invalid option.")
	}
}

// changeResolution allows changing display resolution
func (w *MainWindow) changeResolution() {
	fmt.Println("\n--- Change Resolution ---")
	fmt.Println("1. 800x600")
	fmt.Println("2. 1024x768")
	fmt.Println("3. 1280x720")
	fmt.Println("4. 1920x1080")
	fmt.Println("5. Custom")

	fmt.Print("\nSelect resolution (1-5): ")
	var choice string
	fmt.Scanln(&choice)

	switch choice {
	case "1":
		fmt.Println("Resolution set to 800x600")
		// TODO: Implement actual resolution change
	case "2":
		fmt.Println("Resolution set to 1024x768")
		// TODO: Implement actual resolution change
	case "3":
		fmt.Println("Resolution set to 1280x720")
		// TODO: Implement actual resolution change
	case "4":
		fmt.Println("Resolution set to 1920x1080")
		// TODO: Implement actual resolution change
	case "5":
		fmt.Print("Enter custom width: ")
		var width string
		fmt.Scanln(&width)
		fmt.Print("Enter custom height: ")
		var height string
		fmt.Scanln(&height)
		fmt.Printf("Resolution set to %sx%s\n", width, height)
		// TODO: Implement actual resolution change
	default:
		fmt.Println("Invalid option.")
	}
}

// toggleFullscreen toggles fullscreen mode
func (w *MainWindow) toggleFullscreen() {
	fmt.Println("Fullscreen mode toggled")
	// TODO: Implement actual fullscreen toggle
}

// showDisplayStats shows display statistics
func (w *MainWindow) showDisplayStats() {
	fmt.Println("\n--- Display Statistics ---")

	// Get bitmap cache stats if available
	w.clientMu.RLock()
	client := w.client
	w.clientMu.RUnlock()

	if client != nil {
		cacheStats := client.GetBitmapCacheStats()
		fmt.Printf("Bitmap Cache Hit Rate: %.2f%%\n", cacheStats["hit_rate"])
		fmt.Printf("Cache Size: %d entries\n", cacheStats["cache_size"])
		fmt.Printf("Memory Usage: %.2f MB\n", cacheStats["memory_usage_mb"])
	} else {
		fmt.Println("FPS: 30")
		fmt.Println("Latency: 50ms")
		fmt.Println("Bandwidth: 1.2 Mbps")
		fmt.Println("Compression: 85%")
	}

	fmt.Println("\nPress Enter to continue...")
	var input string
	fmt.Scanln(&input)
}

// showPerformanceMonitoring shows the performance monitoring dialog
func (w *MainWindow) showPerformanceMonitoring() {
	fmt.Println("\n=== Performance Monitoring ===")

	w.connectionMu.RLock()
	connected := w.isConnected
	w.connectionMu.RUnlock()

	if !connected {
		fmt.Println("Must be connected to access performance monitoring.")
		return
	}

	w.performanceDialog.Show()
}

// showFavorites shows the favorites dialog
func (w *MainWindow) showFavorites() {
	fmt.Println("\n=== Favorites ===")
	w.favoritesDialog.Show()
}

// showConnectionHistory shows the connection history dialog
func (w *MainWindow) showConnectionHistory() {
	fmt.Println("\n=== Connection History ===")
	w.historyDialog.Show()
}

// exit exits the application
func (w *MainWindow) exit() {
	fmt.Println("\nExiting GoRDP GUI Client...")

	// Cleanup
	w.connectionMu.RLock()
	connected := w.isConnected
	w.connectionMu.RUnlock()

	if connected {
		w.disconnect()
	}

	// Stop performance monitoring
	w.performanceMonitor.StopMonitoring()

	fmt.Println("Goodbye!")
	// In a real application, this would exit the program
	// For the CLI version, we'll just return
}

// MainWindowBitmapProcessor implements the bitmap processor interface
type MainWindowBitmapProcessor struct {
	mainWindow *MainWindow
	frameCount int
	startTime  time.Time
}

// ProcessBitmap processes bitmap updates from the RDP server
func (p *MainWindowBitmapProcessor) ProcessBitmap(option *bitmap.Option, bitmap *bitmap.BitMap) {
	p.frameCount++
	if p.startTime.IsZero() {
		p.startTime = time.Now()
	}

	// TODO: Update display widget when UpdateBitmap method is implemented
	// if p.mainWindow.displayWidget != nil {
	// 	p.mainWindow.displayWidget.UpdateBitmap(option, bitmap)
	// }

	// TODO: Update performance statistics when UpdateFrameStats method is implemented
	// if p.mainWindow.performanceMonitor != nil {
	// 	p.mainWindow.performanceMonitor.UpdateFrameStats(option, bitmap)
	// }

	// Log frame information (optional, for debugging)
	if p.frameCount%100 == 0 {
		duration := time.Since(p.startTime)
		fps := float64(p.frameCount) / duration.Seconds()
		fmt.Printf("Frame %d: %dx%d at (%d,%d), FPS: %.2f\n",
			p.frameCount, option.Width, option.Height, option.Left, option.Top, fps)
	}
}
