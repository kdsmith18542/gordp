package glog

import (
	"encoding/json"
	"log"
	"os"
	"time"
)

// LogEntry represents a structured log entry
type LogEntry struct {
	Timestamp time.Time              `json:"timestamp"`
	Level     string                 `json:"level"`
	Message   string                 `json:"message"`
	Fields    map[string]interface{} `json:"fields,omitempty"`
	Error     string                 `json:"error,omitempty"`
	Function  string                 `json:"function,omitempty"`
	File      string                 `json:"file,omitempty"`
	Line      int                    `json:"line,omitempty"`
}

// StructuredLogger provides structured logging capabilities
type StructuredLogger struct {
	logger *log.Logger
	level  LEVEL
	output *os.File
}

// NewStructuredLogger creates a new structured logger
func NewStructuredLogger(output *os.File, level LEVEL) *StructuredLogger {
	if output == nil {
		output = os.Stdout
	}

	return &StructuredLogger{
		logger: log.New(output, "", 0),
		level:  level,
		output: output,
	}
}

// logStructured logs a structured message
func (sl *StructuredLogger) logStructured(level LEVEL, message string, fields map[string]interface{}) {
	if level < sl.level {
		return
	}

	entry := LogEntry{
		Timestamp: time.Now(),
		Level:     levelToString(level),
		Message:   message,
		Fields:    fields,
	}

	// Convert to JSON
	jsonData, err := json.Marshal(entry)
	if err != nil {
		// Fallback to simple logging if JSON marshaling fails
		sl.logger.Printf("[%s] %s", levelToString(level), message)
		return
	}

	sl.logger.Println(string(jsonData))
}

// levelToString converts LEVEL to string
func levelToString(level LEVEL) string {
	switch level {
	case DEBUG:
		return "DEBUG"
	case INFO:
		return "INFO"
	case WARN:
		return "WARN"
	case ERROR:
		return "ERROR"
	default:
		return "UNKNOWN"
	}
}

// DebugStructured logs a debug message with structured fields
func (sl *StructuredLogger) DebugStructured(message string, fields map[string]interface{}) {
	sl.logStructured(DEBUG, message, fields)
}

// InfoStructured logs an info message with structured fields
func (sl *StructuredLogger) InfoStructured(message string, fields map[string]interface{}) {
	sl.logStructured(INFO, message, fields)
}

// WarnStructured logs a warning message with structured fields
func (sl *StructuredLogger) WarnStructured(message string, fields map[string]interface{}) {
	sl.logStructured(WARN, message, fields)
}

// ErrorStructured logs an error message with structured fields
func (sl *StructuredLogger) ErrorStructured(message string, err error, fields map[string]interface{}) {
	if fields == nil {
		fields = make(map[string]interface{})
	}
	if err != nil {
		fields["error"] = err.Error()
	}
	sl.logStructured(ERROR, message, fields)
}

// WithFields creates a new logger with additional fields
func (sl *StructuredLogger) WithFields(fields map[string]interface{}) *StructuredLogger {
	// Create a new logger that includes the additional fields
	newLogger := *sl
	return &newLogger
}

// Performance logging functions
func (sl *StructuredLogger) LogPerformance(operation string, duration time.Duration, fields map[string]interface{}) {
	if fields == nil {
		fields = make(map[string]interface{})
	}
	fields["operation"] = operation
	fields["duration_ms"] = duration.Milliseconds()
	fields["duration_ns"] = duration.Nanoseconds()

	sl.InfoStructured("Performance measurement", fields)
}

// Connection logging functions
func (sl *StructuredLogger) LogConnection(addr string, success bool, duration time.Duration, err error) {
	fields := map[string]interface{}{
		"address":     addr,
		"success":     success,
		"duration_ms": duration.Milliseconds(),
	}

	if err != nil {
		fields["error"] = err.Error()
		sl.ErrorStructured("Connection failed", err, fields)
	} else {
		sl.InfoStructured("Connection established", fields)
	}
}

// Input logging functions
func (sl *StructuredLogger) LogInput(inputType string, data interface{}, fields map[string]interface{}) {
	if fields == nil {
		fields = make(map[string]interface{})
	}
	fields["input_type"] = inputType
	fields["data"] = data

	sl.DebugStructured("Input event", fields)
}

// Bitmap logging functions
func (sl *StructuredLogger) LogBitmap(width, height int, format string, compressed bool, fields map[string]interface{}) {
	if fields == nil {
		fields = make(map[string]interface{})
	}
	fields["width"] = width
	fields["height"] = height
	fields["format"] = format
	fields["compressed"] = compressed

	sl.DebugStructured("Bitmap received", fields)
}

// Virtual channel logging functions
func (sl *StructuredLogger) LogVirtualChannel(channelName string, dataSize int, fields map[string]interface{}) {
	if fields == nil {
		fields = make(map[string]interface{})
	}
	fields["channel_name"] = channelName
	fields["data_size"] = dataSize

	sl.DebugStructured("Virtual channel data", fields)
}

// Global structured logger instance
var structuredLogger *StructuredLogger

func init() {
	structuredLogger = NewStructuredLogger(nil, DEBUG)
}

// SetStructuredLogger sets the global structured logger
func SetStructuredLogger(logger *StructuredLogger) {
	structuredLogger = logger
}

// GetStructuredLogger returns the global structured logger
func GetStructuredLogger() *StructuredLogger {
	return structuredLogger
}
