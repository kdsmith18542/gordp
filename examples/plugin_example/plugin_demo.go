package main

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/kdsmith18542/gordp"
	"github.com/kdsmith18542/gordp/config"
	"github.com/kdsmith18542/gordp/plugin"
	"github.com/kdsmith18542/gordp/plugin/examples"
	"github.com/kdsmith18542/gordp/proto/bitmap"
)

// dummyProcessor implements the Processor interface for the example
type dummyProcessor struct{}

func (dp *dummyProcessor) ProcessBitmap(option *bitmap.Option, bitmap *bitmap.BitMap) {
	// This is just a dummy implementation for the example
}

func main() {
	fmt.Println("GoRDP Plugin System Example")
	fmt.Println("===========================")

	// Create plugin manager
	pluginManager := plugin.NewPluginManager()
	defer pluginManager.Close()

	// Create and register plugins
	loggerPlugin := examples.NewLoggerPlugin()
	performancePlugin := examples.NewPerformancePlugin()

	// Register plugins
	if err := pluginManager.RegisterPlugin(loggerPlugin); err != nil {
		log.Fatalf("Failed to register logger plugin: %v", err)
	}

	if err := pluginManager.RegisterPlugin(performancePlugin); err != nil {
		log.Fatalf("Failed to register performance plugin: %v", err)
	}

	// Initialize plugins with configuration
	loggerConfig := map[string]interface{}{
		"log_file":  "rdp_plugin_example.log",
		"log_level": "debug",
	}

	performanceConfig := map[string]interface{}{
		"update_interval": 500, // 500ms
		"max_history":     50,
		"enable_alerts":   true,
	}

	if err := pluginManager.InitializePlugin("logger", loggerConfig); err != nil {
		log.Fatalf("Failed to initialize logger plugin: %v", err)
	}

	if err := pluginManager.InitializePlugin("performance", performanceConfig); err != nil {
		log.Fatalf("Failed to initialize performance plugin: %v", err)
	}

	// Start plugins
	if err := pluginManager.StartAllPlugins(); err != nil {
		log.Fatalf("Failed to start plugins: %v", err)
	}

	// Create RDP client configuration
	cfg := &config.Config{
		Connection: config.ConnectionConfig{
			Address: "localhost",
			Port:    3389,
		},
		Authentication: config.AuthConfig{
			Username: "testuser",
			Password: "testpass",
		},
		Display: config.DisplayConfig{
			Width:      1024,
			Height:     768,
			ColorDepth: 16,
		},
		Security: config.SecurityConfig{
			EncryptionLevel: "high",
		},
		Performance: config.PerformanceConfig{
			BitmapCache: true,
		},
	}

	// Create RDP client with context
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	client := gordp.NewClient(&gordp.Option{
		Addr:           cfg.Connection.Address + ":" + fmt.Sprintf("%d", cfg.Connection.Port),
		UserName:       cfg.Authentication.Username,
		Password:       cfg.Authentication.Password,
		ConnectTimeout: cfg.Connection.ConnectTimeout,
	})

	// Set up signal handling for graceful shutdown
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)

	// Start connection in a goroutine
	go func() {
		fmt.Println("Connecting to RDP server...")

		// Log connection attempt
		loggerPlugin.LogConnection(cfg.Connection.Address, cfg.Connection.Port, false)

		if err := client.ConnectWithContext(ctx); err != nil {
			loggerPlugin.LogError("Connection failed", err)
			log.Printf("Connection failed: %v", err)
			return
		}

		// Log successful connection
		loggerPlugin.LogConnection(cfg.Connection.Address, cfg.Connection.Port, true)
		loggerPlugin.LogInfo("RDP connection established", map[string]interface{}{
			"host": cfg.Connection.Address,
			"port": cfg.Connection.Port,
		})

		// Start the RDP session with a dummy processor
		if err := client.RunWithContext(ctx, &dummyProcessor{}); err != nil {
			loggerPlugin.LogError("RDP session failed", err)
			log.Printf("RDP session failed: %v", err)
			return
		}
	}()

	// Start performance monitoring goroutine
	go func() {
		ticker := time.NewTicker(2 * time.Second)
		defer ticker.Stop()

		for {
			select {
			case <-ctx.Done():
				return
			case <-ticker.C:
				// Update performance metrics
				performancePlugin.UpdateConnectionLatency(50 * time.Millisecond)
				performancePlugin.UpdateBandwidthUsage(1024*1024, 512*1024) // 1MB sent, 512KB received
				performancePlugin.UpdateCachePerformance("bitmap", 85, 15)  // 85% hit rate
				performancePlugin.UpdateCachePerformance("glyph", 90, 10)   // 90% hit rate
				performancePlugin.UpdateInputLatency(10 * time.Millisecond)
				performancePlugin.UpdateDisplayLatency(25 * time.Millisecond)
				performancePlugin.UpdateAudioLatency(15 * time.Millisecond)
				performancePlugin.UpdateCPUUsage(25.5) // 25.5% CPU usage

				// Log performance summary
				summary := performancePlugin.GetPerformanceSummary()
				loggerPlugin.LogPerformance(summary)

				// Print performance summary to console
				fmt.Printf("\n=== Performance Summary ===\n")
				fmt.Printf("Frame Rate: %.2f fps\n", summary["current_frame_rate"])
				fmt.Printf("CPU Usage: %.1f%%\n", summary["current_cpu_usage"])
				fmt.Printf("Connection Latency: %s\n", summary["connection_latency"])
				fmt.Printf("Input Latency: %s\n", summary["input_latency"])
				fmt.Printf("Display Latency: %s\n", summary["display_latency"])
				fmt.Printf("Audio Latency: %s\n", summary["audio_latency"])
				fmt.Printf("Bandwidth In: %.2f Mbps\n", summary["bandwidth_in_mbps"])
				fmt.Printf("Bandwidth Out: %.2f Mbps\n", summary["bandwidth_out_mbps"])
				fmt.Printf("Bitmap Cache Hit Rate: %.1f%%\n", summary["bitmap_cache_hit_rate"].(float64)*100)
				fmt.Printf("Glyph Cache Hit Rate: %.1f%%\n", summary["glyph_cache_hit_rate"].(float64)*100)
				fmt.Printf("Virtual Channels: %d\n", summary["virtual_channels"])
				fmt.Printf("Total Errors: %d\n", summary["total_errors"])
				fmt.Printf("Total Warnings: %d\n", summary["total_warnings"])
				fmt.Printf("===========================\n")
			}
		}
	}()

	// Start plugin statistics monitoring
	go func() {
		ticker := time.NewTicker(5 * time.Second)
		defer ticker.Stop()

		for {
			select {
			case <-ctx.Done():
				return
			case <-ticker.C:
				// Get plugin statistics
				stats := pluginManager.GetStats()

				fmt.Printf("\n=== Plugin Statistics ===\n")
				fmt.Printf("Total Plugins: %d\n", stats.TotalPlugins)
				fmt.Printf("Running Plugins: %d\n", stats.RunningPlugins)
				fmt.Printf("Error Plugins: %d\n", stats.ErrorPlugins)

				for name, info := range stats.PluginDetails {
					if plugin, exists := pluginManager.GetPlugin(name); exists {
						fmt.Printf("Plugin: %s (Type: %s, Status: %s)\n",
							name, info.Type, plugin.Status())
					}
				}
				fmt.Printf("=======================\n")
			}
		}
	}()

	// Simulate some RDP events
	go func() {
		time.Sleep(3 * time.Second)

		// Simulate input events
		loggerPlugin.LogInput("keyboard", map[string]interface{}{
			"key_code":  65,
			"key_name":  "A",
			"modifiers": []string{"shift"},
		})

		loggerPlugin.LogInput("mouse", map[string]interface{}{
			"x":      100,
			"y":      200,
			"button": "left",
			"action": "click",
		})

		// Simulate display events
		loggerPlugin.LogDisplay(1024, 768, "RDP6")

		// Simulate security events
		loggerPlugin.LogSecurity("authentication", map[string]interface{}{
			"method": "NTLM",
			"user":   cfg.Authentication.Username,
		})

		// Simulate virtual channel events
		loggerPlugin.LogVirtualChannel(1, "cliprdr", "created")
		loggerPlugin.LogVirtualChannel(2, "rdpsnd", "created")

		// Simulate device events
		loggerPlugin.LogDevice("printer", "HP LaserJet", "announced")
		loggerPlugin.LogDevice("usb", "USB Flash Drive", "connected")

		// Simulate clipboard events
		loggerPlugin.LogClipboard("text", 256, "server_to_client")

		// Simulate audio events
		loggerPlugin.LogAudio(1, 1024, uint32(time.Now().Unix()))

		// Simulate cache events
		loggerPlugin.LogCache("bitmap", 0.85, 1024*1024)
		loggerPlugin.LogCache("glyph", 0.90, 512*1024)

		// Simulate network events
		loggerPlugin.LogNetwork(1024*1024, 512*1024, 50*time.Millisecond)

		// Simulate session events
		loggerPlugin.LogSession("session-123", "user_logged_in", map[string]interface{}{
			"username": cfg.Authentication.Username,
			"host":     cfg.Connection.Address,
		})

		// Simulate user events
		loggerPlugin.LogUser(cfg.Authentication.Username, "application_launched", map[string]interface{}{
			"app_name": "notepad.exe",
		})

		// Simulate system events
		loggerPlugin.LogSystem("display_changed", map[string]interface{}{
			"old_resolution": "800x600",
			"new_resolution": "1024x768",
		})

		// Add some performance errors and warnings
		performancePlugin.AddError("network_timeout", "Connection timeout", "warning", "network")
		performancePlugin.AddError("cache_miss", "Bitmap cache miss", "info", "cache")
		performancePlugin.AddWarning("high_latency", "Input latency above threshold", "input")

		// Update virtual channel metrics
		performancePlugin.UpdateVirtualChannelMetrics(1, "cliprdr", 1024, 2048)
		performancePlugin.UpdateVirtualChannelMetrics(2, "rdpsnd", 512, 1024)
	}()

	// Wait for interrupt signal
	<-sigChan
	fmt.Println("\nShutting down...")

	// Cancel context to stop all goroutines
	cancel()

	// Stop all plugins
	if err := pluginManager.StopAllPlugins(); err != nil {
		log.Printf("Error stopping plugins: %v", err)
	}

	// Get final statistics
	finalStats := pluginManager.GetStats()
	fmt.Printf("\n=== Final Plugin Statistics ===\n")
	fmt.Printf("Total Plugins: %d\n", finalStats.TotalPlugins)
	fmt.Printf("Running Plugins: %d\n", finalStats.RunningPlugins)
	fmt.Printf("Error Plugins: %d\n", finalStats.ErrorPlugins)

	// Get final performance summary
	finalSummary := performancePlugin.GetPerformanceSummary()
	fmt.Printf("\n=== Final Performance Summary ===\n")
	summaryJSON, _ := json.MarshalIndent(finalSummary, "", "  ")
	fmt.Println(string(summaryJSON))

	// Get log statistics
	logStats := loggerPlugin.GetLogStats()
	fmt.Printf("\n=== Log Statistics ===\n")
	logStatsJSON, _ := json.MarshalIndent(logStats, "", "  ")
	fmt.Println(string(logStatsJSON))

	fmt.Println("\nPlugin example completed successfully!")
}
