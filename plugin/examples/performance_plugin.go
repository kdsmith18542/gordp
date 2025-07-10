package examples

import (
	"context"
	"fmt"
	"runtime"
	"sync"
	"time"

	"github.com/kdsmith18542/gordp/plugin"
)

// PerformanceMetrics contains various performance metrics
type PerformanceMetrics struct {
	Timestamp         time.Time        `json:"timestamp"`
	ConnectionLatency time.Duration    `json:"connection_latency"`
	FrameRate         float64          `json:"frame_rate"`
	BandwidthUsage    BandwidthMetrics `json:"bandwidth_usage"`
	CachePerformance  CacheMetrics     `json:"cache_performance"`
	MemoryUsage       MemoryMetrics    `json:"memory_usage"`
	CPUUsage          float64          `json:"cpu_usage"`
	InputLatency      time.Duration    `json:"input_latency"`
	DisplayLatency    time.Duration    `json:"display_latency"`
	AudioLatency      time.Duration    `json:"audio_latency"`
	VirtualChannels   []ChannelMetrics `json:"virtual_channels"`
	Errors            []ErrorMetric    `json:"errors"`
	Warnings          []WarningMetric  `json:"warnings"`
}

// BandwidthMetrics contains bandwidth-related metrics
type BandwidthMetrics struct {
	BytesSent       int64   `json:"bytes_sent"`
	BytesReceived   int64   `json:"bytes_received"`
	PacketsSent     int64   `json:"packets_sent"`
	PacketsReceived int64   `json:"packets_received"`
	BandwidthIn     float64 `json:"bandwidth_in_mbps"`
	BandwidthOut    float64 `json:"bandwidth_out_mbps"`
	PacketLoss      float64 `json:"packet_loss_percent"`
	Retransmissions int64   `json:"retransmissions"`
}

// CacheMetrics contains cache-related metrics
type CacheMetrics struct {
	BitmapCacheHits    int64   `json:"bitmap_cache_hits"`
	BitmapCacheMisses  int64   `json:"bitmap_cache_misses"`
	BitmapCacheHitRate float64 `json:"bitmap_cache_hit_rate"`
	GlyphCacheHits     int64   `json:"glyph_cache_hits"`
	GlyphCacheMisses   int64   `json:"glyph_cache_misses"`
	GlyphCacheHitRate  float64 `json:"glyph_cache_hit_rate"`
	CacheSize          int64   `json:"cache_size_bytes"`
	CacheEvictions     int64   `json:"cache_evictions"`
}

// MemoryMetrics contains memory-related metrics
type MemoryMetrics struct {
	HeapAlloc    uint64 `json:"heap_alloc_bytes"`
	HeapSys      uint64 `json:"heap_sys_bytes"`
	HeapIdle     uint64 `json:"heap_idle_bytes"`
	HeapInuse    uint64 `json:"heap_inuse_bytes"`
	HeapReleased uint64 `json:"heap_released_bytes"`
	HeapObjects  uint64 `json:"heap_objects"`
	StackInuse   uint64 `json:"stack_inuse_bytes"`
	StackSys     uint64 `json:"stack_sys_bytes"`
	TotalAlloc   uint64 `json:"total_alloc_bytes"`
	NumGC        uint32 `json:"num_gc"`
	PauseTotalNs uint64 `json:"pause_total_ns"`
}

// ChannelMetrics contains virtual channel metrics
type ChannelMetrics struct {
	ChannelID        uint32        `json:"channel_id"`
	ChannelName      string        `json:"channel_name"`
	BytesSent        int64         `json:"bytes_sent"`
	BytesReceived    int64         `json:"bytes_received"`
	MessagesSent     int64         `json:"messages_sent"`
	MessagesReceived int64         `json:"messages_received"`
	Latency          time.Duration `json:"latency"`
	Errors           int64         `json:"errors"`
}

// ErrorMetric contains error information
type ErrorMetric struct {
	Timestamp    time.Time `json:"timestamp"`
	ErrorType    string    `json:"error_type"`
	ErrorMessage string    `json:"error_message"`
	Severity     string    `json:"severity"`
	Component    string    `json:"component"`
}

// WarningMetric contains warning information
type WarningMetric struct {
	Timestamp      time.Time `json:"timestamp"`
	WarningType    string    `json:"warning_type"`
	WarningMessage string    `json:"warning_message"`
	Component      string    `json:"component"`
}

// PerformancePlugin is a plugin that monitors performance metrics
type PerformancePlugin struct {
	info           *plugin.PluginInfo
	status         plugin.PluginStatus
	metrics        *PerformanceMetrics
	history        []*PerformanceMetrics
	maxHistorySize int
	mutex          sync.RWMutex
	ctx            context.Context
	cancel         context.CancelFunc
	startTime      time.Time
	lastUpdate     time.Time
	updateInterval time.Duration
	handlers       []plugin.PluginEventHandler
}

// NewPerformancePlugin creates a new performance monitoring plugin
func NewPerformancePlugin() *PerformancePlugin {
	return &PerformancePlugin{
		info: &plugin.PluginInfo{
			Name:        "performance",
			Version:     "1.0.0",
			Type:        plugin.PluginTypePerformance,
			Description: "Monitors RDP connection performance metrics",
			Author:      "GoRDP Team",
			License:     "MIT",
			Config: map[string]interface{}{
				"update_interval": 1000, // milliseconds
				"max_history":     100,
				"enable_alerts":   true,
			},
		},
		status:         plugin.PluginStatusUnloaded,
		metrics:        &PerformanceMetrics{},
		history:        make([]*PerformanceMetrics, 0),
		maxHistorySize: 100,
		updateInterval: time.Second,
		handlers:       make([]plugin.PluginEventHandler, 0),
	}
}

// Info returns plugin information
func (pp *PerformancePlugin) Info() *plugin.PluginInfo {
	return pp.info
}

// Initialize initializes the plugin
func (pp *PerformancePlugin) Initialize(config map[string]interface{}) error {
	pp.mutex.Lock()
	defer pp.mutex.Unlock()

	pp.status = plugin.PluginStatusLoading

	// Get configuration values
	if interval, ok := config["update_interval"].(int); ok {
		pp.updateInterval = time.Duration(interval) * time.Millisecond
	}

	if maxHistory, ok := config["max_history"].(int); ok {
		pp.maxHistorySize = maxHistory
	}

	pp.metrics = &PerformanceMetrics{
		Timestamp:       time.Now(),
		VirtualChannels: make([]ChannelMetrics, 0),
		Errors:          make([]ErrorMetric, 0),
		Warnings:        make([]WarningMetric, 0),
	}

	pp.status = plugin.PluginStatusLoaded
	return nil
}

// Start starts the plugin
func (pp *PerformancePlugin) Start(ctx context.Context) error {
	pp.mutex.Lock()
	defer pp.mutex.Unlock()

	if pp.status != plugin.PluginStatusLoaded {
		return fmt.Errorf("plugin not in loaded state: %s", pp.status)
	}

	pp.ctx, pp.cancel = context.WithCancel(ctx)
	pp.startTime = time.Now()
	pp.lastUpdate = time.Now()
	pp.status = plugin.PluginStatusRunning

	// Start monitoring goroutine
	go pp.monitor()

	return nil
}

// Stop stops the plugin
func (pp *PerformancePlugin) Stop() error {
	pp.mutex.Lock()
	defer pp.mutex.Unlock()

	if pp.status != plugin.PluginStatusRunning {
		return nil
	}

	if pp.cancel != nil {
		pp.cancel()
	}

	pp.status = plugin.PluginStatusStopped
	return nil
}

// Status returns the current status
func (pp *PerformancePlugin) Status() plugin.PluginStatus {
	pp.mutex.RLock()
	defer pp.mutex.RUnlock()
	return pp.status
}

// monitor runs the monitoring loop
func (pp *PerformancePlugin) monitor() {
	ticker := time.NewTicker(pp.updateInterval)
	defer ticker.Stop()

	for {
		select {
		case <-pp.ctx.Done():
			return
		case <-ticker.C:
			pp.updateMetrics()
		}
	}
}

// updateMetrics updates the current metrics
func (pp *PerformancePlugin) updateMetrics() {
	pp.mutex.Lock()
	defer pp.mutex.Unlock()

	now := time.Now()

	// Update timestamp
	pp.metrics.Timestamp = now

	// Calculate frame rate (simplified)
	if !pp.lastUpdate.IsZero() {
		elapsed := now.Sub(pp.lastUpdate).Seconds()
		if elapsed > 0 {
			pp.metrics.FrameRate = 1.0 / elapsed
		}
	}

	// Update memory metrics
	pp.updateMemoryMetrics()

	// Add to history
	pp.addToHistory(pp.metrics)

	// Emit performance event
	pp.emitPerformanceEvent()

	pp.lastUpdate = now
}

// updateMemoryMetrics updates memory-related metrics
func (pp *PerformancePlugin) updateMemoryMetrics() {
	var memStats runtime.MemStats
	runtime.ReadMemStats(&memStats)

	pp.metrics.MemoryUsage = MemoryMetrics{
		HeapAlloc:    memStats.HeapAlloc,
		HeapSys:      memStats.HeapSys,
		HeapIdle:     memStats.HeapIdle,
		HeapInuse:    memStats.HeapInuse,
		HeapReleased: memStats.HeapReleased,
		HeapObjects:  memStats.HeapObjects,
		StackInuse:   memStats.StackInuse,
		StackSys:     memStats.StackSys,
		TotalAlloc:   memStats.TotalAlloc,
		NumGC:        memStats.NumGC,
		PauseTotalNs: memStats.PauseTotalNs,
	}
}

// addToHistory adds metrics to history
func (pp *PerformancePlugin) addToHistory(metrics *PerformanceMetrics) {
	// Create a copy of the metrics
	metricsCopy := *metrics
	pp.history = append(pp.history, &metricsCopy)

	// Trim history if it exceeds max size
	if len(pp.history) > pp.maxHistorySize {
		pp.history = pp.history[1:]
	}
}

// emitPerformanceEvent emits a performance event
func (pp *PerformancePlugin) emitPerformanceEvent() {
	event := &plugin.PluginEvent{
		PluginName: pp.info.Name,
		EventType:  "performance_update",
		Timestamp:  time.Now(),
		Data:       pp.metrics,
	}

	for _, handler := range pp.handlers {
		handler(event)
	}
}

// RegisterEventHandler registers an event handler
func (pp *PerformancePlugin) RegisterEventHandler(handler plugin.PluginEventHandler) {
	pp.mutex.Lock()
	defer pp.mutex.Unlock()
	pp.handlers = append(pp.handlers, handler)
}

// UpdateConnectionLatency updates connection latency
func (pp *PerformancePlugin) UpdateConnectionLatency(latency time.Duration) {
	pp.mutex.Lock()
	defer pp.mutex.Unlock()
	pp.metrics.ConnectionLatency = latency
}

// UpdateBandwidthUsage updates bandwidth metrics
func (pp *PerformancePlugin) UpdateBandwidthUsage(bytesSent, bytesReceived int64) {
	pp.mutex.Lock()
	defer pp.mutex.Unlock()

	pp.metrics.BandwidthUsage.BytesSent = bytesSent
	pp.metrics.BandwidthUsage.BytesReceived = bytesReceived

	// Calculate bandwidth in Mbps
	elapsed := time.Since(pp.startTime).Seconds()
	if elapsed > 0 {
		pp.metrics.BandwidthUsage.BandwidthIn = float64(bytesReceived) * 8 / (1024 * 1024 * elapsed)
		pp.metrics.BandwidthUsage.BandwidthOut = float64(bytesSent) * 8 / (1024 * 1024 * elapsed)
	}
}

// UpdateCachePerformance updates cache metrics
func (pp *PerformancePlugin) UpdateCachePerformance(cacheType string, hits, misses int64) {
	pp.mutex.Lock()
	defer pp.mutex.Unlock()

	switch cacheType {
	case "bitmap":
		pp.metrics.CachePerformance.BitmapCacheHits = hits
		pp.metrics.CachePerformance.BitmapCacheMisses = misses
		total := hits + misses
		if total > 0 {
			pp.metrics.CachePerformance.BitmapCacheHitRate = float64(hits) / float64(total)
		}
	case "glyph":
		pp.metrics.CachePerformance.GlyphCacheHits = hits
		pp.metrics.CachePerformance.GlyphCacheMisses = misses
		total := hits + misses
		if total > 0 {
			pp.metrics.CachePerformance.GlyphCacheHitRate = float64(hits) / float64(total)
		}
	}
}

// UpdateVirtualChannelMetrics updates virtual channel metrics
func (pp *PerformancePlugin) UpdateVirtualChannelMetrics(channelID uint32, channelName string, bytesSent, bytesReceived int64) {
	pp.mutex.Lock()
	defer pp.mutex.Unlock()

	// Find existing channel or create new one
	for i, channel := range pp.metrics.VirtualChannels {
		if channel.ChannelID == channelID {
			pp.metrics.VirtualChannels[i].BytesSent = bytesSent
			pp.metrics.VirtualChannels[i].BytesReceived = bytesReceived
			return
		}
	}

	// Add new channel
	pp.metrics.VirtualChannels = append(pp.metrics.VirtualChannels, ChannelMetrics{
		ChannelID:     channelID,
		ChannelName:   channelName,
		BytesSent:     bytesSent,
		BytesReceived: bytesReceived,
	})
}

// AddError adds an error metric
func (pp *PerformancePlugin) AddError(errorType, message, severity, component string) {
	pp.mutex.Lock()
	defer pp.mutex.Unlock()

	errorMetric := ErrorMetric{
		Timestamp:    time.Now(),
		ErrorType:    errorType,
		ErrorMessage: message,
		Severity:     severity,
		Component:    component,
	}

	pp.metrics.Errors = append(pp.metrics.Errors, errorMetric)
}

// AddWarning adds a warning metric
func (pp *PerformancePlugin) AddWarning(warningType, message, component string) {
	pp.mutex.Lock()
	defer pp.mutex.Unlock()

	warningMetric := WarningMetric{
		Timestamp:      time.Now(),
		WarningType:    warningType,
		WarningMessage: message,
		Component:      component,
	}

	pp.metrics.Warnings = append(pp.metrics.Warnings, warningMetric)
}

// UpdateInputLatency updates input latency
func (pp *PerformancePlugin) UpdateInputLatency(latency time.Duration) {
	pp.mutex.Lock()
	defer pp.mutex.Unlock()
	pp.metrics.InputLatency = latency
}

// UpdateDisplayLatency updates display latency
func (pp *PerformancePlugin) UpdateDisplayLatency(latency time.Duration) {
	pp.mutex.Lock()
	defer pp.mutex.Unlock()
	pp.metrics.DisplayLatency = latency
}

// UpdateAudioLatency updates audio latency
func (pp *PerformancePlugin) UpdateAudioLatency(latency time.Duration) {
	pp.mutex.Lock()
	defer pp.mutex.Unlock()
	pp.metrics.AudioLatency = latency
}

// UpdateCPUUsage updates CPU usage
func (pp *PerformancePlugin) UpdateCPUUsage(usage float64) {
	pp.mutex.Lock()
	defer pp.mutex.Unlock()
	pp.metrics.CPUUsage = usage
}

// GetCurrentMetrics returns the current performance metrics
func (pp *PerformancePlugin) GetCurrentMetrics() *PerformanceMetrics {
	pp.mutex.RLock()
	defer pp.mutex.RUnlock()

	// Return a copy to avoid race conditions
	metricsCopy := *pp.metrics
	return &metricsCopy
}

// GetMetricsHistory returns the metrics history
func (pp *PerformancePlugin) GetMetricsHistory() []*PerformanceMetrics {
	pp.mutex.RLock()
	defer pp.mutex.RUnlock()

	// Return a copy of the history
	historyCopy := make([]*PerformanceMetrics, len(pp.history))
	copy(historyCopy, pp.history)
	return historyCopy
}

// GetAverageMetrics returns average metrics over the specified duration
func (pp *PerformancePlugin) GetAverageMetrics(duration time.Duration) *PerformanceMetrics {
	pp.mutex.RLock()
	defer pp.mutex.RUnlock()

	if len(pp.history) == 0 {
		return &PerformanceMetrics{}
	}

	cutoff := time.Now().Add(-duration)
	var validMetrics []*PerformanceMetrics

	for _, metric := range pp.history {
		if metric.Timestamp.After(cutoff) {
			validMetrics = append(validMetrics, metric)
		}
	}

	if len(validMetrics) == 0 {
		return &PerformanceMetrics{}
	}

	// Calculate averages
	avgMetrics := &PerformanceMetrics{
		Timestamp: time.Now(),
	}

	var totalLatency, totalInputLatency, totalDisplayLatency, totalAudioLatency time.Duration
	var totalFrameRate, totalCPUUsage float64
	var totalBytesSent, totalBytesReceived int64

	for _, metric := range validMetrics {
		totalLatency += metric.ConnectionLatency
		totalInputLatency += metric.InputLatency
		totalDisplayLatency += metric.DisplayLatency
		totalAudioLatency += metric.AudioLatency
		totalFrameRate += metric.FrameRate
		totalCPUUsage += metric.CPUUsage
		totalBytesSent += metric.BandwidthUsage.BytesSent
		totalBytesReceived += metric.BandwidthUsage.BytesReceived
	}

	count := float64(len(validMetrics))
	avgMetrics.ConnectionLatency = totalLatency / time.Duration(len(validMetrics))
	avgMetrics.InputLatency = totalInputLatency / time.Duration(len(validMetrics))
	avgMetrics.DisplayLatency = totalDisplayLatency / time.Duration(len(validMetrics))
	avgMetrics.AudioLatency = totalAudioLatency / time.Duration(len(validMetrics))
	avgMetrics.FrameRate = totalFrameRate / count
	avgMetrics.CPUUsage = totalCPUUsage / count
	avgMetrics.BandwidthUsage.BytesSent = totalBytesSent
	avgMetrics.BandwidthUsage.BytesReceived = totalBytesReceived

	return avgMetrics
}

// GetPerformanceSummary returns a summary of performance metrics
func (pp *PerformancePlugin) GetPerformanceSummary() map[string]interface{} {
	pp.mutex.RLock()
	defer pp.mutex.RUnlock()

	summary := map[string]interface{}{
		"uptime":                time.Since(pp.startTime).String(),
		"current_frame_rate":    pp.metrics.FrameRate,
		"current_cpu_usage":     pp.metrics.CPUUsage,
		"connection_latency":    pp.metrics.ConnectionLatency.String(),
		"input_latency":         pp.metrics.InputLatency.String(),
		"display_latency":       pp.metrics.DisplayLatency.String(),
		"audio_latency":         pp.metrics.AudioLatency.String(),
		"bandwidth_in_mbps":     pp.metrics.BandwidthUsage.BandwidthIn,
		"bandwidth_out_mbps":    pp.metrics.BandwidthUsage.BandwidthOut,
		"bitmap_cache_hit_rate": pp.metrics.CachePerformance.BitmapCacheHitRate,
		"glyph_cache_hit_rate":  pp.metrics.CachePerformance.GlyphCacheHitRate,
		"total_errors":          len(pp.metrics.Errors),
		"total_warnings":        len(pp.metrics.Warnings),
		"virtual_channels":      len(pp.metrics.VirtualChannels),
	}

	return summary
}
