package examples

import (
	"context"
	"fmt"
	"log"
	"os"
	"sync"
	"time"

	"github.com/kdsmith18542/gordp/plugin"
)

// LoggerPlugin is an example plugin that logs all events
type LoggerPlugin struct {
	info    *plugin.PluginInfo
	status  plugin.PluginStatus
	logger  *log.Logger
	logFile *os.File
	mutex   sync.RWMutex
	ctx     context.Context
	cancel  context.CancelFunc
}

// NewLoggerPlugin creates a new logger plugin
func NewLoggerPlugin() *LoggerPlugin {
	return &LoggerPlugin{
		info: &plugin.PluginInfo{
			Name:        "logger",
			Version:     "1.0.0",
			Type:        plugin.PluginTypeCustom,
			Description: "Logs all RDP events to a file",
			Author:      "GoRDP Team",
			License:     "MIT",
			Config: map[string]interface{}{
				"log_file":  "rdp_events.log",
				"log_level": "info",
			},
		},
		status: plugin.PluginStatusUnloaded,
	}
}

// Info returns plugin information
func (lp *LoggerPlugin) Info() *plugin.PluginInfo {
	return lp.info
}

// Initialize initializes the plugin
func (lp *LoggerPlugin) Initialize(config map[string]interface{}) error {
	lp.mutex.Lock()
	defer lp.mutex.Unlock()

	lp.status = plugin.PluginStatusLoading

	// Get log file path from config
	logFile, ok := config["log_file"].(string)
	if !ok {
		logFile = "rdp_events.log"
	}

	// Open log file
	file, err := os.OpenFile(logFile, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0644)
	if err != nil {
		lp.status = plugin.PluginStatusError
		return fmt.Errorf("failed to open log file: %w", err)
	}

	lp.logFile = file
	lp.logger = log.New(file, "[RDP-LOGGER] ", log.LstdFlags|log.Lshortfile)

	lp.status = plugin.PluginStatusLoaded
	lp.logger.Printf("Logger plugin initialized with log file: %s", logFile)
	return nil
}

// Start starts the plugin
func (lp *LoggerPlugin) Start(ctx context.Context) error {
	lp.mutex.Lock()
	defer lp.mutex.Unlock()

	if lp.status != plugin.PluginStatusLoaded {
		return fmt.Errorf("plugin not in loaded state: %s", lp.status)
	}

	lp.ctx, lp.cancel = context.WithCancel(ctx)
	lp.status = plugin.PluginStatusRunning

	lp.logger.Printf("Logger plugin started")
	return nil
}

// Stop stops the plugin
func (lp *LoggerPlugin) Stop() error {
	lp.mutex.Lock()
	defer lp.mutex.Unlock()

	if lp.status != plugin.PluginStatusRunning {
		return nil
	}

	if lp.cancel != nil {
		lp.cancel()
	}

	if lp.logFile != nil {
		lp.logger.Printf("Logger plugin stopped")
		lp.logFile.Close()
	}

	lp.status = plugin.PluginStatusStopped
	return nil
}

// Status returns the current status
func (lp *LoggerPlugin) Status() plugin.PluginStatus {
	lp.mutex.RLock()
	defer lp.mutex.RUnlock()
	return lp.status
}

// LogEvent logs an event
func (lp *LoggerPlugin) LogEvent(eventType, message string, data interface{}) {
	lp.mutex.RLock()
	defer lp.mutex.RUnlock()

	if lp.logger != nil {
		lp.logger.Printf("[%s] %s - %v", eventType, message, data)
	}
}

// LogError logs an error
func (lp *LoggerPlugin) LogError(message string, err error) {
	lp.mutex.RLock()
	defer lp.mutex.RUnlock()

	if lp.logger != nil {
		lp.logger.Printf("[ERROR] %s: %v", message, err)
	}
}

// LogPerformance logs performance metrics
func (lp *LoggerPlugin) LogPerformance(metrics map[string]interface{}) {
	lp.mutex.RLock()
	defer lp.mutex.RUnlock()

	if lp.logger != nil {
		lp.logger.Printf("[PERFORMANCE] %v", metrics)
	}
}

// LogConnection logs connection events
func (lp *LoggerPlugin) LogConnection(host string, port int, success bool) {
	lp.mutex.RLock()
	defer lp.mutex.RUnlock()

	if lp.logger != nil {
		status := "FAILED"
		if success {
			status = "SUCCESS"
		}
		lp.logger.Printf("[CONNECTION] %s:%d - %s", host, port, status)
	}
}

// LogInput logs input events
func (lp *LoggerPlugin) LogInput(inputType string, details interface{}) {
	lp.mutex.RLock()
	defer lp.mutex.RUnlock()

	if lp.logger != nil {
		lp.logger.Printf("[INPUT] %s - %v", inputType, details)
	}
}

// LogDisplay logs display events
func (lp *LoggerPlugin) LogDisplay(width, height int, format string) {
	lp.mutex.RLock()
	defer lp.mutex.RUnlock()

	if lp.logger != nil {
		lp.logger.Printf("[DISPLAY] %dx%d - %s", width, height, format)
	}
}

// LogSecurity logs security events
func (lp *LoggerPlugin) LogSecurity(eventType string, details interface{}) {
	lp.mutex.RLock()
	defer lp.mutex.RUnlock()

	if lp.logger != nil {
		lp.logger.Printf("[SECURITY] %s - %v", eventType, details)
	}
}

// LogVirtualChannel logs virtual channel events
func (lp *LoggerPlugin) LogVirtualChannel(channelID uint32, channelName string, eventType string) {
	lp.mutex.RLock()
	defer lp.mutex.RUnlock()

	if lp.logger != nil {
		lp.logger.Printf("[VIRTUAL-CHANNEL] ID:%d Name:%s Event:%s", channelID, channelName, eventType)
	}
}

// LogDevice logs device events
func (lp *LoggerPlugin) LogDevice(deviceType string, deviceName string, eventType string) {
	lp.mutex.RLock()
	defer lp.mutex.RUnlock()

	if lp.logger != nil {
		lp.logger.Printf("[DEVICE] Type:%s Name:%s Event:%s", deviceType, deviceName, eventType)
	}
}

// LogClipboard logs clipboard events
func (lp *LoggerPlugin) LogClipboard(format string, dataSize int, direction string) {
	lp.mutex.RLock()
	defer lp.mutex.RUnlock()

	if lp.logger != nil {
		lp.logger.Printf("[CLIPBOARD] Format:%s Size:%d Direction:%s", format, dataSize, direction)
	}
}

// LogAudio logs audio events
func (lp *LoggerPlugin) LogAudio(formatID uint16, dataSize int, timestamp uint32) {
	lp.mutex.RLock()
	defer lp.mutex.RUnlock()

	if lp.logger != nil {
		lp.logger.Printf("[AUDIO] Format:%d Size:%d Timestamp:%d", formatID, dataSize, timestamp)
	}
}

// LogCache logs cache events
func (lp *LoggerPlugin) LogCache(cacheType string, hitRate float64, size int) {
	lp.mutex.RLock()
	defer lp.mutex.RUnlock()

	if lp.logger != nil {
		lp.logger.Printf("[CACHE] Type:%s HitRate:%.2f%% Size:%d", cacheType, hitRate*100, size)
	}
}

// LogNetwork logs network events
func (lp *LoggerPlugin) LogNetwork(bytesSent, bytesReceived int, latency time.Duration) {
	lp.mutex.RLock()
	defer lp.mutex.RUnlock()

	if lp.logger != nil {
		lp.logger.Printf("[NETWORK] Sent:%d Received:%d Latency:%v", bytesSent, bytesReceived, latency)
	}
}

// LogSession logs session events
func (lp *LoggerPlugin) LogSession(sessionID string, eventType string, details interface{}) {
	lp.mutex.RLock()
	defer lp.mutex.RUnlock()

	if lp.logger != nil {
		lp.logger.Printf("[SESSION] ID:%s Event:%s Details:%v", sessionID, eventType, details)
	}
}

// LogUser logs user events
func (lp *LoggerPlugin) LogUser(username string, eventType string, details interface{}) {
	lp.mutex.RLock()
	defer lp.mutex.RUnlock()

	if lp.logger != nil {
		lp.logger.Printf("[USER] %s Event:%s Details:%v", username, eventType, details)
	}
}

// LogSystem logs system events
func (lp *LoggerPlugin) LogSystem(eventType string, details interface{}) {
	lp.mutex.RLock()
	defer lp.mutex.RUnlock()

	if lp.logger != nil {
		lp.logger.Printf("[SYSTEM] %s - %v", eventType, details)
	}
}

// LogDebug logs debug information
func (lp *LoggerPlugin) LogDebug(message string, data interface{}) {
	lp.mutex.RLock()
	defer lp.mutex.RUnlock()

	if lp.logger != nil {
		lp.logger.Printf("[DEBUG] %s - %v", message, data)
	}
}

// LogInfo logs info messages
func (lp *LoggerPlugin) LogInfo(message string, data interface{}) {
	lp.mutex.RLock()
	defer lp.mutex.RUnlock()

	if lp.logger != nil {
		lp.logger.Printf("[INFO] %s - %v", message, data)
	}
}

// LogWarning logs warning messages
func (lp *LoggerPlugin) LogWarning(message string, data interface{}) {
	lp.mutex.RLock()
	defer lp.mutex.RUnlock()

	if lp.logger != nil {
		lp.logger.Printf("[WARNING] %s - %v", message, data)
	}
}

// LogCritical logs critical messages
func (lp *LoggerPlugin) LogCritical(message string, data interface{}) {
	lp.mutex.RLock()
	defer lp.mutex.RUnlock()

	if lp.logger != nil {
		lp.logger.Printf("[CRITICAL] %s - %v", message, data)
	}
}

// GetLogStats returns log statistics
func (lp *LoggerPlugin) GetLogStats() map[string]interface{} {
	lp.mutex.RLock()
	defer lp.mutex.RUnlock()

	stats := map[string]interface{}{
		"status":    lp.status,
		"log_file":  lp.info.Config["log_file"],
		"log_level": lp.info.Config["log_level"],
	}

	if lp.logFile != nil {
		if fileInfo, err := lp.logFile.Stat(); err == nil {
			stats["file_size"] = fileInfo.Size()
			stats["last_modified"] = fileInfo.ModTime()
		}
	}

	return stats
}
