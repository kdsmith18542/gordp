package glog

import (
	"encoding/json"
	"os"
	"strings"
	"testing"
	"time"
)

func TestNewStructuredLogger(t *testing.T) {
	logger := NewStructuredLogger(nil, INFO)
	if logger == nil {
		t.Error("expected logger to be created")
	}
	if logger.level != INFO {
		t.Errorf("expected level to be INFO, got %v", logger.level)
	}
}

func TestLogEntry(t *testing.T) {
	entry := LogEntry{
		Timestamp: time.Now(),
		Level:     "INFO",
		Message:   "test message",
		Fields: map[string]interface{}{
			"key1": "value1",
			"key2": 42,
		},
	}

	// Test JSON marshaling
	data, err := json.Marshal(entry)
	if err != nil {
		t.Errorf("failed to marshal log entry: %v", err)
	}

	// Verify JSON contains expected fields
	jsonStr := string(data)
	if !strings.Contains(jsonStr, "test message") {
		t.Error("expected JSON to contain message")
	}
	if !strings.Contains(jsonStr, "INFO") {
		t.Error("expected JSON to contain level")
	}
	if !strings.Contains(jsonStr, "key1") {
		t.Error("expected JSON to contain field key1")
	}
	if !strings.Contains(jsonStr, "value1") {
		t.Error("expected JSON to contain field value1")
	}
}

func TestStructuredLoggerLogging(t *testing.T) {
	// Create a temporary file for testing
	tmpFile, err := os.CreateTemp("", "test_log")
	if err != nil {
		t.Fatalf("failed to create temp file: %v", err)
	}
	defer os.Remove(tmpFile.Name())
	defer tmpFile.Close()

	logger := NewStructuredLogger(tmpFile, DEBUG)

	// Test debug logging
	logger.DebugStructured("debug message", map[string]interface{}{
		"debug_key": "debug_value",
	})

	// Test info logging
	logger.InfoStructured("info message", map[string]interface{}{
		"info_key": "info_value",
	})

	// Test warning logging
	logger.WarnStructured("warning message", map[string]interface{}{
		"warn_key": "warn_value",
	})

	// Test error logging
	testErr := &testError{message: "test error"}
	logger.ErrorStructured("error message", testErr, map[string]interface{}{
		"error_key": "error_value",
	})

	// Read the log file
	tmpFile.Seek(0, 0)
	content, err := os.ReadFile(tmpFile.Name())
	if err != nil {
		t.Fatalf("failed to read log file: %v", err)
	}

	lines := strings.Split(string(content), "\n")
	if len(lines) < 4 {
		t.Errorf("expected at least 4 log lines, got %d", len(lines))
	}

	// Verify each log entry
	for _, line := range lines {
		if line == "" {
			continue
		}
		var entry LogEntry
		if err := json.Unmarshal([]byte(line), &entry); err != nil {
			t.Errorf("failed to unmarshal log entry: %v", err)
		}
		if entry.Message == "" {
			t.Error("expected log entry to have a message")
		}
		if entry.Level == "" {
			t.Error("expected log entry to have a level")
		}
	}
}

func TestStructuredLoggerLevelFiltering(t *testing.T) {
	tmpFile, err := os.CreateTemp("", "test_log")
	if err != nil {
		t.Fatalf("failed to create temp file: %v", err)
	}
	defer os.Remove(tmpFile.Name())
	defer tmpFile.Close()

	// Create logger with INFO level (should filter out DEBUG)
	logger := NewStructuredLogger(tmpFile, INFO)

	// This should be filtered out
	logger.DebugStructured("debug message", nil)

	// This should be logged
	logger.InfoStructured("info message", nil)

	// Read the log file
	tmpFile.Seek(0, 0)
	content, err := os.ReadFile(tmpFile.Name())
	if err != nil {
		t.Fatalf("failed to read log file: %v", err)
	}

	lines := strings.Split(string(content), "\n")
	// Should only have one non-empty line (the info message)
	nonEmptyLines := 0
	for _, line := range lines {
		if strings.TrimSpace(line) != "" {
			nonEmptyLines++
		}
	}

	if nonEmptyLines != 1 {
		t.Errorf("expected 1 log entry, got %d", nonEmptyLines)
	}
}

func TestPerformanceLogging(t *testing.T) {
	tmpFile, err := os.CreateTemp("", "test_log")
	if err != nil {
		t.Fatalf("failed to create temp file: %v", err)
	}
	defer os.Remove(tmpFile.Name())
	defer tmpFile.Close()

	logger := NewStructuredLogger(tmpFile, INFO)

	// Test performance logging
	duration := 100 * time.Millisecond
	logger.LogPerformance("test_operation", duration, map[string]interface{}{
		"custom_field": "custom_value",
	})

	// Read and verify
	tmpFile.Seek(0, 0)
	content, err := os.ReadFile(tmpFile.Name())
	if err != nil {
		t.Fatalf("failed to read log file: %v", err)
	}

	var entry LogEntry
	if err := json.Unmarshal(content, &entry); err != nil {
		t.Fatalf("failed to unmarshal log entry: %v", err)
	}

	if entry.Message != "Performance measurement" {
		t.Errorf("expected message 'Performance measurement', got '%s'", entry.Message)
	}

	if entry.Fields["operation"] != "test_operation" {
		t.Errorf("expected operation 'test_operation', got '%v'", entry.Fields["operation"])
	}

	if entry.Fields["duration_ms"] != float64(100) {
		t.Errorf("expected duration_ms 100, got %v", entry.Fields["duration_ms"])
	}

	if entry.Fields["custom_field"] != "custom_value" {
		t.Errorf("expected custom_field 'custom_value', got '%v'", entry.Fields["custom_field"])
	}
}

func TestConnectionLogging(t *testing.T) {
	tmpFile, err := os.CreateTemp("", "test_log")
	if err != nil {
		t.Fatalf("failed to create temp file: %v", err)
	}
	defer os.Remove(tmpFile.Name())
	defer tmpFile.Close()

	logger := NewStructuredLogger(tmpFile, INFO)

	// Test successful connection
	logger.LogConnection("localhost:3389", true, 50*time.Millisecond, nil)

	// Test failed connection
	testErr := &testError{message: "connection failed"}
	logger.LogConnection("invalid:9999", false, 5*time.Second, testErr)

	// Read and verify
	tmpFile.Seek(0, 0)
	content, err := os.ReadFile(tmpFile.Name())
	if err != nil {
		t.Fatalf("failed to read log file: %v", err)
	}

	lines := strings.Split(string(content), "\n")
	if len(lines) < 2 {
		t.Errorf("expected at least 2 log lines, got %d", len(lines))
	}

	// Verify successful connection log
	var successEntry LogEntry
	if err := json.Unmarshal([]byte(lines[0]), &successEntry); err != nil {
		t.Fatalf("failed to unmarshal success log entry: %v", err)
	}

	if successEntry.Message != "Connection established" {
		t.Errorf("expected message 'Connection established', got '%s'", successEntry.Message)
	}

	if successEntry.Fields["address"] != "localhost:3389" {
		t.Errorf("expected address 'localhost:3389', got '%v'", successEntry.Fields["address"])
	}

	if successEntry.Fields["success"] != true {
		t.Errorf("expected success true, got %v", successEntry.Fields["success"])
	}

	// Verify failed connection log
	var failureEntry LogEntry
	if err := json.Unmarshal([]byte(lines[1]), &failureEntry); err != nil {
		t.Fatalf("failed to unmarshal failure log entry: %v", err)
	}

	if failureEntry.Message != "Connection failed" {
		t.Errorf("expected message 'Connection failed', got '%s'", failureEntry.Message)
	}

	if failureEntry.Fields["success"] != false {
		t.Errorf("expected success false, got %v", failureEntry.Fields["success"])
	}

	if failureEntry.Fields["error"] != "connection failed" {
		t.Errorf("expected error 'connection failed', got '%v'", failureEntry.Fields["error"])
	}
}

func TestInputLogging(t *testing.T) {
	tmpFile, err := os.CreateTemp("", "test_log")
	if err != nil {
		t.Fatalf("failed to create temp file: %v", err)
	}
	defer os.Remove(tmpFile.Name())
	defer tmpFile.Close()

	logger := NewStructuredLogger(tmpFile, DEBUG)

	// Test input logging
	logger.LogInput("keyboard", "a", map[string]interface{}{
		"key_code": 65,
	})

	// Read and verify
	tmpFile.Seek(0, 0)
	content, err := os.ReadFile(tmpFile.Name())
	if err != nil {
		t.Fatalf("failed to read log file: %v", err)
	}

	var entry LogEntry
	if err := json.Unmarshal(content, &entry); err != nil {
		t.Fatalf("failed to unmarshal log entry: %v", err)
	}

	if entry.Message != "Input event" {
		t.Errorf("expected message 'Input event', got '%s'", entry.Message)
	}

	if entry.Fields["input_type"] != "keyboard" {
		t.Errorf("expected input_type 'keyboard', got '%v'", entry.Fields["input_type"])
	}

	if entry.Fields["data"] != "a" {
		t.Errorf("expected data 'a', got '%v'", entry.Fields["data"])
	}

	if entry.Fields["key_code"] != float64(65) {
		t.Errorf("expected key_code 65, got %v", entry.Fields["key_code"])
	}
}

func TestBitmapLogging(t *testing.T) {
	tmpFile, err := os.CreateTemp("", "test_log")
	if err != nil {
		t.Fatalf("failed to create temp file: %v", err)
	}
	defer os.Remove(tmpFile.Name())
	defer tmpFile.Close()

	logger := NewStructuredLogger(tmpFile, DEBUG)

	// Test bitmap logging
	logger.LogBitmap(1920, 1080, "RDP6", true, map[string]interface{}{
		"cache_hit": true,
	})

	// Read and verify
	tmpFile.Seek(0, 0)
	content, err := os.ReadFile(tmpFile.Name())
	if err != nil {
		t.Fatalf("failed to read log file: %v", err)
	}

	var entry LogEntry
	if err := json.Unmarshal(content, &entry); err != nil {
		t.Fatalf("failed to unmarshal log entry: %v", err)
	}

	if entry.Message != "Bitmap received" {
		t.Errorf("expected message 'Bitmap received', got '%s'", entry.Message)
	}

	if entry.Fields["width"] != float64(1920) {
		t.Errorf("expected width 1920, got %v", entry.Fields["width"])
	}

	if entry.Fields["height"] != float64(1080) {
		t.Errorf("expected height 1080, got %v", entry.Fields["height"])
	}

	if entry.Fields["format"] != "RDP6" {
		t.Errorf("expected format 'RDP6', got '%v'", entry.Fields["format"])
	}

	if entry.Fields["compressed"] != true {
		t.Errorf("expected compressed true, got %v", entry.Fields["compressed"])
	}

	if entry.Fields["cache_hit"] != true {
		t.Errorf("expected cache_hit true, got %v", entry.Fields["cache_hit"])
	}
}

func TestVirtualChannelLogging(t *testing.T) {
	tmpFile, err := os.CreateTemp("", "test_log")
	if err != nil {
		t.Fatalf("failed to create temp file: %v", err)
	}
	defer os.Remove(tmpFile.Name())
	defer tmpFile.Close()

	logger := NewStructuredLogger(tmpFile, DEBUG)

	// Test virtual channel logging
	logger.LogVirtualChannel("cliprdr", 1024, map[string]interface{}{
		"format": "text",
	})

	// Read and verify
	tmpFile.Seek(0, 0)
	content, err := os.ReadFile(tmpFile.Name())
	if err != nil {
		t.Fatalf("failed to read log file: %v", err)
	}

	var entry LogEntry
	if err := json.Unmarshal(content, &entry); err != nil {
		t.Fatalf("failed to unmarshal log entry: %v", err)
	}

	if entry.Message != "Virtual channel data" {
		t.Errorf("expected message 'Virtual channel data', got '%s'", entry.Message)
	}

	if entry.Fields["channel_name"] != "cliprdr" {
		t.Errorf("expected channel_name 'cliprdr', got '%v'", entry.Fields["channel_name"])
	}

	if entry.Fields["data_size"] != float64(1024) {
		t.Errorf("expected data_size 1024, got %v", entry.Fields["data_size"])
	}

	if entry.Fields["format"] != "text" {
		t.Errorf("expected format 'text', got '%v'", entry.Fields["format"])
	}
}

// Helper type for testing
type testError struct {
	message string
}

func (e *testError) Error() string {
	return e.message
}
