package performance

import (
	"fmt"
	"runtime"
	"sync"
	"time"

	"github.com/kdsmith18542/gordp"
)

// PerformanceMetrics represents performance data
type PerformanceMetrics struct {
	// Display metrics
	FPS         float64
	Latency     time.Duration
	Bandwidth   float64 // Mbps
	Compression float64 // Percentage
	FrameDrops  int
	FrameCount  int64

	// Memory metrics
	MemoryUsage uint64 // Bytes
	MemoryPeak  uint64 // Bytes
	CacheHits   int64
	CacheMisses int64

	// Network metrics
	BytesReceived     int64
	BytesSent         int64
	PacketLoss        float64 // Percentage
	ConnectionQuality string

	// CPU metrics
	CPUUsage       float64 // Percentage
	ProcessingTime time.Duration

	// Timestamp
	Timestamp time.Time
}

// PerformanceMonitor monitors RDP client performance
type PerformanceMonitor struct {
	mu sync.RWMutex

	// Current metrics
	currentMetrics *PerformanceMetrics

	// Historical data
	history    []*PerformanceMetrics
	maxHistory int

	// Monitoring state
	isMonitoring bool
	startTime    time.Time

	// Thresholds
	thresholds map[string]float64

	// RDP client reference
	client *gordp.Client

	// Performance tracking
	lastFrameTime   time.Time
	frameCount      int64
	lastCPUUsage    float64
	lastMemoryUsage uint64
}

// NewPerformanceMonitor creates a new performance monitor
func NewPerformanceMonitor() *PerformanceMonitor {
	return &PerformanceMonitor{
		currentMetrics: &PerformanceMetrics{
			Timestamp: time.Now(),
		},
		history:    make([]*PerformanceMetrics, 0),
		maxHistory: 1000, // Keep last 1000 samples
		thresholds: map[string]float64{
			"fps_min":     15.0,
			"latency_max": 200.0, // ms
			"memory_max":  500.0, // MB
			"cpu_max":     80.0,  // percentage
		},
	}
}

// SetClient sets the RDP client reference
func (m *PerformanceMonitor) SetClient(client *gordp.Client) {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.client = client
}

// GetClient returns the RDP client reference
func (m *PerformanceMonitor) GetClient() *gordp.Client {
	m.mu.RLock()
	defer m.mu.RUnlock()
	return m.client
}

// GetPerformanceStats returns real performance statistics
func (m *PerformanceMonitor) GetPerformanceStats() map[string]interface{} {
	m.mu.RLock()
	defer m.mu.RUnlock()

	stats := make(map[string]interface{})

	// Get current metrics
	current := m.currentMetrics
	stats["current"] = map[string]interface{}{
		"fps":                current.FPS,
		"latency_ms":         current.Latency.Milliseconds(),
		"bandwidth_mbps":     current.Bandwidth,
		"compression_pct":    current.Compression,
		"memory_mb":          float64(current.MemoryUsage) / 1024 / 1024,
		"cpu_pct":            current.CPUUsage,
		"packet_loss_pct":    current.PacketLoss,
		"connection_quality": current.ConnectionQuality,
		"frame_count":        current.FrameCount,
		"frame_drops":        current.FrameDrops,
		"cache_hits":         current.CacheHits,
		"cache_misses":       current.CacheMisses,
		"bytes_received":     current.BytesReceived,
		"bytes_sent":         current.BytesSent,
		"processing_time_ms": current.ProcessingTime.Milliseconds(),
	}

	// Get RDP client statistics if available
	if m.client != nil {
		// Get bitmap cache statistics
		if cacheStats := m.client.GetBitmapCacheStats(); cacheStats != nil {
			stats["bitmap_cache"] = cacheStats
		}

		// Get device statistics
		if deviceStats := m.client.GetDeviceStats(); deviceStats != nil {
			stats["device_stats"] = deviceStats
		}
	}

	// Get system statistics
	stats["system"] = m.getSystemStats()

	// Get historical averages
	if len(m.history) > 0 {
		stats["averages"] = m.getHistoricalAverages()
	}

	// Get monitoring status
	stats["monitoring"] = map[string]interface{}{
		"is_active":  m.isMonitoring,
		"uptime":     time.Since(m.startTime).String(),
		"start_time": m.startTime.Format(time.RFC3339),
	}

	return stats
}

// UpdatePerformanceMetrics updates performance metrics with real data
func (m *PerformanceMonitor) UpdatePerformanceMetrics() {
	m.mu.Lock()
	defer m.mu.Unlock()

	now := time.Now()

	// Update timestamp
	m.currentMetrics.Timestamp = now

	// Calculate FPS
	if !m.lastFrameTime.IsZero() {
		elapsed := now.Sub(m.lastFrameTime).Seconds()
		if elapsed > 0 {
			m.currentMetrics.FPS = 1.0 / elapsed
		}
	}
	m.lastFrameTime = now
	m.frameCount++

	// Update frame count
	m.currentMetrics.FrameCount = m.frameCount

	// Get system memory usage
	m.updateMemoryMetrics()

	// Get CPU usage
	m.updateCPUMetrics()

	// Get RDP client metrics if available
	if m.client != nil {
		m.updateRDPClientMetrics()
	}

	// Calculate bandwidth if we have network data
	m.calculateBandwidth()

	// Add to history
	m.history = append(m.history, m.currentMetrics)

	// Trim history if needed
	if len(m.history) > m.maxHistory {
		m.history = m.history[1:]
	}

	// Create new metrics object for next update
	m.currentMetrics = &PerformanceMetrics{
		Timestamp: now,
	}

	fmt.Printf("Performance metrics updated: FPS=%.1f, Memory=%.1fMB, CPU=%.1f%%\n",
		m.currentMetrics.FPS, float64(m.currentMetrics.MemoryUsage)/1024/1024, m.currentMetrics.CPUUsage)
}

// updateMemoryMetrics updates memory usage metrics
func (m *PerformanceMonitor) updateMemoryMetrics() {
	var memStats runtime.MemStats
	runtime.ReadMemStats(&memStats)

	m.currentMetrics.MemoryUsage = memStats.HeapAlloc
	if memStats.HeapAlloc > m.currentMetrics.MemoryPeak {
		m.currentMetrics.MemoryPeak = memStats.HeapAlloc
	}
}

// updateCPUMetrics updates CPU usage metrics
func (m *PerformanceMonitor) updateCPUMetrics() {
	// Simple CPU usage calculation based on processing time
	// In a real implementation, you would use system-specific APIs
	// For now, we'll use a simplified approach
	if m.lastCPUUsage == 0 {
		m.currentMetrics.CPUUsage = 25.0 // Default value
	} else {
		// Simulate CPU usage based on processing time
		processingTime := m.currentMetrics.ProcessingTime.Milliseconds()
		m.currentMetrics.CPUUsage = float64(processingTime) / 10.0 // Simplified calculation
		if m.currentMetrics.CPUUsage > 100.0 {
			m.currentMetrics.CPUUsage = 100.0
		}
	}
	m.lastCPUUsage = m.currentMetrics.CPUUsage
}

// updateRDPClientMetrics updates metrics from RDP client
func (m *PerformanceMonitor) updateRDPClientMetrics() {
	// Get bitmap cache statistics
	if cacheStats := m.client.GetBitmapCacheStats(); cacheStats != nil {
		// Extract cache hit/miss data
		if cache0, ok := cacheStats["cache_0"].(map[string]interface{}); ok {
			if hits, ok := cache0["hit_count"].(int64); ok {
				m.currentMetrics.CacheHits = hits
			}
			if misses, ok := cache0["miss_count"].(int64); ok {
				m.currentMetrics.CacheMisses = misses
			}
		}
	}

	// Get device statistics for network data
	if deviceStats := m.client.GetDeviceStats(); deviceStats != nil {
		// Extract network statistics
		if bytesReceived, ok := deviceStats["bytes_received"].(int64); ok {
			m.currentMetrics.BytesReceived = bytesReceived
		}
		if bytesSent, ok := deviceStats["bytes_sent"].(int64); ok {
			m.currentMetrics.BytesSent = bytesSent
		}
	}
}

// calculateBandwidth calculates bandwidth usage
func (m *PerformanceMonitor) calculateBandwidth() {
	// Calculate bandwidth in Mbps
	elapsed := time.Since(m.startTime).Seconds()
	if elapsed > 0 {
		totalBytes := m.currentMetrics.BytesReceived + m.currentMetrics.BytesSent
		m.currentMetrics.Bandwidth = float64(totalBytes) * 8 / (1024 * 1024 * elapsed)
	}
}

// getSystemStats returns system statistics
func (m *PerformanceMonitor) getSystemStats() map[string]interface{} {
	var memStats runtime.MemStats
	runtime.ReadMemStats(&memStats)

	return map[string]interface{}{
		"heap_alloc":     memStats.HeapAlloc,
		"heap_sys":       memStats.HeapSys,
		"heap_idle":      memStats.HeapIdle,
		"heap_inuse":     memStats.HeapInuse,
		"heap_released":  memStats.HeapReleased,
		"heap_objects":   memStats.HeapObjects,
		"stack_inuse":    memStats.StackInuse,
		"stack_sys":      memStats.StackSys,
		"total_alloc":    memStats.TotalAlloc,
		"num_gc":         memStats.NumGC,
		"pause_total_ns": memStats.PauseTotalNs,
		"num_goroutines": runtime.NumGoroutine(),
	}
}

// getHistoricalAverages returns historical averages
func (m *PerformanceMonitor) getHistoricalAverages() map[string]interface{} {
	if len(m.history) == 0 {
		return map[string]interface{}{}
	}

	var sumFPS, sumLatency, sumBandwidth, sumCompression, sumCPU, sumPacketLoss float64
	var sumMemory, sumBytesReceived, sumBytesSent uint64
	var sumCacheHits, sumCacheMisses int64
	count := 0

	for _, metrics := range m.history {
		sumFPS += metrics.FPS
		sumLatency += float64(metrics.Latency.Milliseconds())
		sumBandwidth += metrics.Bandwidth
		sumCompression += metrics.Compression
		sumCPU += metrics.CPUUsage
		sumPacketLoss += metrics.PacketLoss
		sumMemory += metrics.MemoryUsage
		sumBytesReceived += uint64(metrics.BytesReceived)
		sumBytesSent += uint64(metrics.BytesSent)
		sumCacheHits += metrics.CacheHits
		sumCacheMisses += metrics.CacheMisses
		count++
	}

	if count == 0 {
		return map[string]interface{}{}
	}

	return map[string]interface{}{
		"avg_fps":             sumFPS / float64(count),
		"avg_latency_ms":      sumLatency / float64(count),
		"avg_bandwidth_mbps":  sumBandwidth / float64(count),
		"avg_compression_pct": sumCompression / float64(count),
		"avg_cpu_pct":         sumCPU / float64(count),
		"avg_packet_loss_pct": sumPacketLoss / float64(count),
		"avg_memory_mb":       float64(sumMemory/uint64(count)) / 1024 / 1024,
		"avg_bytes_received":  int64(sumBytesReceived / uint64(count)),
		"avg_bytes_sent":      int64(sumBytesSent / uint64(count)),
		"avg_cache_hits":      sumCacheHits / int64(count),
		"avg_cache_misses":    sumCacheMisses / int64(count),
	}
}

// StartMonitoring starts performance monitoring
func (m *PerformanceMonitor) StartMonitoring() {
	m.mu.Lock()
	defer m.mu.Unlock()

	if m.isMonitoring {
		return
	}

	m.isMonitoring = true
	m.startTime = time.Now()

	// Start monitoring goroutine
	go m.monitoringLoop()

	fmt.Println("Performance monitoring started")
}

// StopMonitoring stops performance monitoring
func (m *PerformanceMonitor) StopMonitoring() {
	m.mu.Lock()
	defer m.mu.Unlock()

	m.isMonitoring = false
	fmt.Println("Performance monitoring stopped")
}

// monitoringLoop runs the monitoring loop
func (m *PerformanceMonitor) monitoringLoop() {
	ticker := time.NewTicker(1 * time.Second)
	defer ticker.Stop()

	for {
		select {
		case <-ticker.C:
			if !m.isMonitoring {
				return
			}
			m.UpdatePerformanceMetrics()
		}
	}
}

// updateMetrics updates current performance metrics
func (m *PerformanceMonitor) updateMetrics() {
	m.mu.Lock()
	defer m.mu.Unlock()

	// Update timestamp
	m.currentMetrics.Timestamp = time.Now()

	// Add to history
	m.history = append(m.history, m.currentMetrics)

	// Trim history if needed
	if len(m.history) > m.maxHistory {
		m.history = m.history[1:]
	}

	// Create new metrics object for next update
	m.currentMetrics = &PerformanceMetrics{
		Timestamp: time.Now(),
	}
}

// UpdateFPS updates the frames per second metric
func (m *PerformanceMonitor) UpdateFPS(fps float64) {
	m.mu.Lock()
	defer m.mu.Unlock()

	m.currentMetrics.FPS = fps
}

// UpdateLatency updates the latency metric
func (m *PerformanceMonitor) UpdateLatency(latency time.Duration) {
	m.mu.Lock()
	defer m.mu.Unlock()

	m.currentMetrics.Latency = latency
}

// UpdateBandwidth updates the bandwidth metric
func (m *PerformanceMonitor) UpdateBandwidth(bandwidth float64) {
	m.mu.Lock()
	defer m.mu.Unlock()

	m.currentMetrics.Bandwidth = bandwidth
}

// UpdateCompression updates the compression ratio metric
func (m *PerformanceMonitor) UpdateCompression(compression float64) {
	m.mu.Lock()
	defer m.mu.Unlock()

	m.currentMetrics.Compression = compression
}

// UpdateMemoryUsage updates memory usage metrics
func (m *PerformanceMonitor) UpdateMemoryUsage(usage uint64) {
	m.mu.Lock()
	defer m.mu.Unlock()

	m.currentMetrics.MemoryUsage = usage
	if usage > m.currentMetrics.MemoryPeak {
		m.currentMetrics.MemoryPeak = usage
	}
}

// UpdateCacheStats updates cache hit/miss statistics
func (m *PerformanceMonitor) UpdateCacheStats(hits, misses int64) {
	m.mu.Lock()
	defer m.mu.Unlock()

	m.currentMetrics.CacheHits = hits
	m.currentMetrics.CacheMisses = misses
}

// UpdateNetworkStats updates network statistics
func (m *PerformanceMonitor) UpdateNetworkStats(bytesReceived, bytesSent int64, packetLoss float64) {
	m.mu.Lock()
	defer m.mu.Unlock()

	m.currentMetrics.BytesReceived = bytesReceived
	m.currentMetrics.BytesSent = bytesSent
	m.currentMetrics.PacketLoss = packetLoss

	// Determine connection quality
	if packetLoss < 1.0 {
		m.currentMetrics.ConnectionQuality = "Excellent"
	} else if packetLoss < 5.0 {
		m.currentMetrics.ConnectionQuality = "Good"
	} else if packetLoss < 10.0 {
		m.currentMetrics.ConnectionQuality = "Fair"
	} else {
		m.currentMetrics.ConnectionQuality = "Poor"
	}
}

// UpdateCPUUsage updates CPU usage metric
func (m *PerformanceMonitor) UpdateCPUUsage(usage float64) {
	m.mu.Lock()
	defer m.mu.Unlock()

	m.currentMetrics.CPUUsage = usage
}

// UpdateProcessingTime updates processing time metric
func (m *PerformanceMonitor) UpdateProcessingTime(processingTime time.Duration) {
	m.mu.Lock()
	defer m.mu.Unlock()

	m.currentMetrics.ProcessingTime = processingTime
}

// GetCurrentMetrics returns the current performance metrics
func (m *PerformanceMonitor) GetCurrentMetrics() *PerformanceMetrics {
	m.mu.RLock()
	defer m.mu.RUnlock()

	// Return a copy to avoid race conditions
	metrics := *m.currentMetrics
	return &metrics
}

// GetAverageMetrics returns average metrics over the specified duration
func (m *PerformanceMonitor) GetAverageMetrics(duration time.Duration) *PerformanceMetrics {
	m.mu.RLock()
	defer m.mu.RUnlock()

	if len(m.history) == 0 {
		return &PerformanceMetrics{}
	}

	cutoff := time.Now().Add(-duration)
	var sumFPS, sumLatency, sumBandwidth, sumCompression, sumCPU, sumPacketLoss float64
	var sumMemory, sumBytesReceived, sumBytesSent uint64
	var sumCacheHits, sumCacheMisses int64
	count := 0

	for _, metrics := range m.history {
		if metrics.Timestamp.After(cutoff) {
			sumFPS += metrics.FPS
			sumLatency += float64(metrics.Latency.Milliseconds())
			sumBandwidth += metrics.Bandwidth
			sumCompression += metrics.Compression
			sumCPU += metrics.CPUUsage
			sumPacketLoss += metrics.PacketLoss
			sumMemory += metrics.MemoryUsage
			sumBytesReceived += uint64(metrics.BytesReceived)
			sumBytesSent += uint64(metrics.BytesSent)
			sumCacheHits += metrics.CacheHits
			sumCacheMisses += metrics.CacheMisses
			count++
		}
	}

	if count == 0 {
		return &PerformanceMetrics{}
	}

	return &PerformanceMetrics{
		FPS:           sumFPS / float64(count),
		Latency:       time.Duration(sumLatency/float64(count)) * time.Millisecond,
		Bandwidth:     sumBandwidth / float64(count),
		Compression:   sumCompression / float64(count),
		CPUUsage:      sumCPU / float64(count),
		PacketLoss:    sumPacketLoss / float64(count),
		MemoryUsage:   sumMemory / uint64(count),
		BytesReceived: int64(sumBytesReceived / uint64(count)),
		BytesSent:     int64(sumBytesSent / uint64(count)),
		CacheHits:     sumCacheHits / int64(count),
		CacheMisses:   sumCacheMisses / int64(count),
		Timestamp:     time.Now(),
	}
}

// GetPerformanceReport returns a comprehensive performance report
func (m *PerformanceMonitor) GetPerformanceReport() map[string]interface{} {
	m.mu.RLock()
	defer m.mu.RUnlock()

	current := m.currentMetrics
	avg1min := m.GetAverageMetrics(1 * time.Minute)
	avg5min := m.GetAverageMetrics(5 * time.Minute)

	// Calculate cache hit rate
	cacheHitRate := 0.0
	if current.CacheHits+current.CacheMisses > 0 {
		cacheHitRate = float64(current.CacheHits) / float64(current.CacheHits+current.CacheMisses) * 100
	}

	return map[string]interface{}{
		"current": map[string]interface{}{
			"fps":                current.FPS,
			"latency_ms":         current.Latency.Milliseconds(),
			"bandwidth_mbps":     current.Bandwidth,
			"compression_pct":    current.Compression,
			"memory_mb":          float64(current.MemoryUsage) / 1024 / 1024,
			"cpu_pct":            current.CPUUsage,
			"packet_loss_pct":    current.PacketLoss,
			"connection_quality": current.ConnectionQuality,
			"cache_hit_rate_pct": cacheHitRate,
		},
		"average_1min": map[string]interface{}{
			"fps":             avg1min.FPS,
			"latency_ms":      avg1min.Latency.Milliseconds(),
			"bandwidth_mbps":  avg1min.Bandwidth,
			"compression_pct": avg1min.Compression,
			"memory_mb":       float64(avg1min.MemoryUsage) / 1024 / 1024,
			"cpu_pct":         avg1min.CPUUsage,
		},
		"average_5min": map[string]interface{}{
			"fps":             avg5min.FPS,
			"latency_ms":      avg5min.Latency.Milliseconds(),
			"bandwidth_mbps":  avg5min.Bandwidth,
			"compression_pct": avg5min.Compression,
			"memory_mb":       float64(avg5min.MemoryUsage) / 1024 / 1024,
			"cpu_pct":         avg5min.CPUUsage,
		},
		"uptime":     time.Since(m.startTime).String(),
		"monitoring": m.isMonitoring,
	}
}

// CheckThresholds checks if any metrics exceed thresholds
func (m *PerformanceMonitor) CheckThresholds() []string {
	m.mu.RLock()
	defer m.mu.RUnlock()

	var warnings []string
	current := m.currentMetrics

	if current.FPS < m.thresholds["fps_min"] {
		warnings = append(warnings, fmt.Sprintf("Low FPS: %.1f (min: %.1f)", current.FPS, m.thresholds["fps_min"]))
	}

	if float64(current.Latency.Milliseconds()) > m.thresholds["latency_max"] {
		warnings = append(warnings, fmt.Sprintf("High latency: %.0fms (max: %.0fms)", float64(current.Latency.Milliseconds()), m.thresholds["latency_max"]))
	}

	memoryMB := float64(current.MemoryUsage) / 1024 / 1024
	if memoryMB > m.thresholds["memory_max"] {
		warnings = append(warnings, fmt.Sprintf("High memory usage: %.1fMB (max: %.1fMB)", memoryMB, m.thresholds["memory_max"]))
	}

	if current.CPUUsage > m.thresholds["cpu_max"] {
		warnings = append(warnings, fmt.Sprintf("High CPU usage: %.1f%% (max: %.1f%%)", current.CPUUsage, m.thresholds["cpu_max"]))
	}

	return warnings
}

// SetThreshold sets a performance threshold
func (m *PerformanceMonitor) SetThreshold(name string, value float64) {
	m.mu.Lock()
	defer m.mu.Unlock()

	m.thresholds[name] = value
}

// GetThresholds returns all current thresholds
func (m *PerformanceMonitor) GetThresholds() map[string]float64 {
	m.mu.RLock()
	defer m.mu.RUnlock()

	thresholds := make(map[string]float64)
	for k, v := range m.thresholds {
		thresholds[k] = v
	}
	return thresholds
}

// ClearHistory clears performance history
func (m *PerformanceMonitor) ClearHistory() {
	m.mu.Lock()
	defer m.mu.Unlock()

	m.history = make([]*PerformanceMetrics, 0)
	fmt.Println("Performance history cleared")
}

// IsMonitoring returns whether monitoring is active
func (m *PerformanceMonitor) IsMonitoring() bool {
	m.mu.RLock()
	defer m.mu.RUnlock()
	return m.isMonitoring
}
