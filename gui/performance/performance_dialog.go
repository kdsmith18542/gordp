package performance

import (
	"fmt"
	"strconv"
	"time"
)

// PerformanceDialog manages the performance monitoring dialog
type PerformanceDialog struct {
	monitor *PerformanceMonitor
}

// NewPerformanceDialog creates a new performance dialog
func NewPerformanceDialog(monitor *PerformanceMonitor) *PerformanceDialog {
	return &PerformanceDialog{
		monitor: monitor,
	}
}

// Show displays the performance monitoring dialog
func (d *PerformanceDialog) Show() {
	fmt.Println("\n=== Performance Monitoring ===")

	for {
		fmt.Println("\n1. Start/Stop Monitoring")
		fmt.Println("2. Current Performance")
		fmt.Println("3. Performance History")
		fmt.Println("4. Performance Thresholds")
		fmt.Println("5. Performance Report")
		fmt.Println("6. Clear History")
		fmt.Println("7. Back to Main Menu")

		fmt.Print("\nSelect option (1-7): ")
		var choice string
		fmt.Scanln(&choice)

		switch choice {
		case "1":
			d.toggleMonitoring()
		case "2":
			d.showCurrentPerformance()
		case "3":
			d.showPerformanceHistory()
		case "4":
			d.configureThresholds()
		case "5":
			d.showPerformanceReport()
		case "6":
			d.clearHistory()
		case "7":
			return
		default:
			fmt.Println("Invalid option. Please try again.")
		}
	}
}

// toggleMonitoring starts or stops performance monitoring
func (d *PerformanceDialog) toggleMonitoring() {
	if d.monitor.IsMonitoring() {
		d.monitor.StopMonitoring()
		fmt.Println("Performance monitoring stopped")
	} else {
		d.monitor.StartMonitoring()
		fmt.Println("Performance monitoring started")
	}
}

// showCurrentPerformance displays current performance metrics
func (d *PerformanceDialog) showCurrentPerformance() {
	metrics := d.monitor.GetCurrentMetrics()

	fmt.Println("\n--- Current Performance ---")
	fmt.Printf("FPS: %.1f\n", metrics.FPS)
	fmt.Printf("Latency: %.0fms\n", metrics.Latency.Milliseconds())
	fmt.Printf("Bandwidth: %.2f Mbps\n", metrics.Bandwidth)
	fmt.Printf("Compression: %.1f%%\n", metrics.Compression)
	fmt.Printf("Memory Usage: %.1f MB\n", float64(metrics.MemoryUsage)/1024/1024)
	fmt.Printf("CPU Usage: %.1f%%\n", metrics.CPUUsage)
	fmt.Printf("Packet Loss: %.2f%%\n", metrics.PacketLoss)
	fmt.Printf("Connection Quality: %s\n", metrics.ConnectionQuality)

	// Calculate cache hit rate
	cacheHitRate := 0.0
	if metrics.CacheHits+metrics.CacheMisses > 0 {
		cacheHitRate = float64(metrics.CacheHits) / float64(metrics.CacheHits+metrics.CacheMisses) * 100
	}
	fmt.Printf("Cache Hit Rate: %.1f%%\n", cacheHitRate)

	// Check for warnings
	warnings := d.monitor.CheckThresholds()
	if len(warnings) > 0 {
		fmt.Println("\n⚠️  Performance Warnings:")
		for _, warning := range warnings {
			fmt.Printf("  - %s\n", warning)
		}
	}

	fmt.Println("\nPress Enter to continue...")
	var input string
	fmt.Scanln(&input)
}

// showPerformanceHistory displays performance history
func (d *PerformanceDialog) showPerformanceHistory() {
	fmt.Println("\n--- Performance History ---")

	// Show averages for different time periods
	avg1min := d.monitor.GetAverageMetrics(1 * time.Minute)
	avg5min := d.monitor.GetAverageMetrics(5 * time.Minute)
	avg10min := d.monitor.GetAverageMetrics(10 * time.Minute)

	fmt.Println("1 Minute Average:")
	fmt.Printf("  FPS: %.1f, Latency: %.0fms, Memory: %.1f MB\n",
		avg1min.FPS, avg1min.Latency.Milliseconds(), float64(avg1min.MemoryUsage)/1024/1024)

	fmt.Println("5 Minute Average:")
	fmt.Printf("  FPS: %.1f, Latency: %.0fms, Memory: %.1f MB\n",
		avg5min.FPS, avg5min.Latency.Milliseconds(), float64(avg5min.MemoryUsage)/1024/1024)

	fmt.Println("10 Minute Average:")
	fmt.Printf("  FPS: %.1f, Latency: %.0fms, Memory: %.1f MB\n",
		avg10min.FPS, avg10min.Latency.Milliseconds(), float64(avg10min.MemoryUsage)/1024/1024)

	fmt.Println("\nPress Enter to continue...")
	var input string
	fmt.Scanln(&input)
}

// configureThresholds allows configuring performance thresholds
func (d *PerformanceDialog) configureThresholds() {
	thresholds := d.monitor.GetThresholds()

	fmt.Println("\n--- Performance Thresholds ---")
	fmt.Printf("Current thresholds:\n")
	fmt.Printf("  Minimum FPS: %.1f\n", thresholds["fps_min"])
	fmt.Printf("  Maximum Latency: %.0fms\n", thresholds["latency_max"])
	fmt.Printf("  Maximum Memory: %.1f MB\n", thresholds["memory_max"])
	fmt.Printf("  Maximum CPU: %.1f%%\n", thresholds["cpu_max"])

	fmt.Println("\n1. Set Minimum FPS")
	fmt.Println("2. Set Maximum Latency")
	fmt.Println("3. Set Maximum Memory")
	fmt.Println("4. Set Maximum CPU")
	fmt.Println("5. Reset to Defaults")
	fmt.Println("6. Back")

	fmt.Print("\nSelect option (1-6): ")
	var choice string
	fmt.Scanln(&choice)

	switch choice {
	case "1":
		d.setThreshold("fps_min", "minimum FPS")
	case "2":
		d.setThreshold("latency_max", "maximum latency (ms)")
	case "3":
		d.setThreshold("memory_max", "maximum memory (MB)")
	case "4":
		d.setThreshold("cpu_max", "maximum CPU usage (%)")
	case "5":
		d.resetThresholds()
	case "6":
		return
	default:
		fmt.Println("Invalid option.")
	}
}

// setThreshold sets a specific threshold
func (d *PerformanceDialog) setThreshold(name, description string) {
	fmt.Printf("Enter %s: ", description)
	var valueStr string
	fmt.Scanln(&valueStr)

	value, err := strconv.ParseFloat(valueStr, 64)
	if err != nil {
		fmt.Println("Invalid value. Please enter a number.")
		return
	}

	d.monitor.SetThreshold(name, value)
	fmt.Printf("Threshold updated: %s = %.1f\n", description, value)
}

// resetThresholds resets thresholds to default values
func (d *PerformanceDialog) resetThresholds() {
	d.monitor.SetThreshold("fps_min", 15.0)
	d.monitor.SetThreshold("latency_max", 200.0)
	d.monitor.SetThreshold("memory_max", 500.0)
	d.monitor.SetThreshold("cpu_max", 80.0)

	fmt.Println("Thresholds reset to default values")
}

// showPerformanceReport displays a comprehensive performance report
func (d *PerformanceDialog) showPerformanceReport() {
	report := d.monitor.GetPerformanceReport()

	fmt.Println("\n=== Performance Report ===")

	// Current metrics
	current := report["current"].(map[string]interface{})
	fmt.Println("Current Metrics:")
	fmt.Printf("  FPS: %.1f\n", current["fps"])
	fmt.Printf("  Latency: %.0fms\n", current["latency_ms"])
	fmt.Printf("  Bandwidth: %.2f Mbps\n", current["bandwidth_mbps"])
	fmt.Printf("  Compression: %.1f%%\n", current["compression_pct"])
	fmt.Printf("  Memory: %.1f MB\n", current["memory_mb"])
	fmt.Printf("  CPU: %.1f%%\n", current["cpu_pct"])
	fmt.Printf("  Packet Loss: %.2f%%\n", current["packet_loss_pct"])
	fmt.Printf("  Connection Quality: %s\n", current["connection_quality"])
	fmt.Printf("  Cache Hit Rate: %.1f%%\n", current["cache_hit_rate_pct"])

	// 1-minute average
	avg1min := report["average_1min"].(map[string]interface{})
	fmt.Println("\n1-Minute Average:")
	fmt.Printf("  FPS: %.1f, Latency: %.0fms, Memory: %.1f MB\n",
		avg1min["fps"], avg1min["latency_ms"], avg1min["memory_mb"])

	// 5-minute average
	avg5min := report["average_5min"].(map[string]interface{})
	fmt.Println("5-Minute Average:")
	fmt.Printf("  FPS: %.1f, Latency: %.0fms, Memory: %.1f MB\n",
		avg5min["fps"], avg5min["latency_ms"], avg5min["memory_mb"])

	// System info
	fmt.Printf("\nUptime: %s\n", report["uptime"])
	fmt.Printf("Monitoring: %v\n", report["monitoring"])

	// Performance recommendations
	d.showPerformanceRecommendations(current)

	fmt.Println("\nPress Enter to continue...")
	var input string
	fmt.Scanln(&input)
}

// showPerformanceRecommendations provides performance recommendations
func (d *PerformanceDialog) showPerformanceRecommendations(current map[string]interface{}) {
	fmt.Println("\nPerformance Recommendations:")

	fps := current["fps"].(float64)
	latency := current["latency_ms"].(float64)
	memory := current["memory_mb"].(float64)
	cpu := current["cpu_pct"].(float64)
	packetLoss := current["packet_loss_pct"].(float64)

	if fps < 15.0 {
		fmt.Println("  ⚠️  Consider reducing display quality or resolution")
	}

	if latency > 100.0 {
		fmt.Println("  ⚠️  High latency detected - check network connection")
	}

	if memory > 400.0 {
		fmt.Println("  ⚠️  High memory usage - consider closing other applications")
	}

	if cpu > 70.0 {
		fmt.Println("  ⚠️  High CPU usage - consider reducing quality settings")
	}

	if packetLoss > 5.0 {
		fmt.Println("  ⚠️  High packet loss - check network stability")
	}

	if fps >= 25.0 && latency < 50.0 && memory < 200.0 && cpu < 50.0 {
		fmt.Println("  ✅ Performance is excellent")
	}
}

// clearHistory clears performance history
func (d *PerformanceDialog) clearHistory() {
	fmt.Print("Are you sure you want to clear performance history? (y/N): ")
	var confirm string
	fmt.Scanln(&confirm)

	if confirm == "y" || confirm == "Y" {
		d.monitor.ClearHistory()
		fmt.Println("Performance history cleared")
	} else {
		fmt.Println("Operation cancelled")
	}
}
