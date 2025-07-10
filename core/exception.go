package core

import (
	"fmt"
	"runtime/debug"
)

// ThrowError throws an error with stack trace
func ThrowError(err error) {
	if err != nil {
		panic(fmt.Errorf("error occurred: %w", err))
	}
}

// ThrowErrorString throws a string as an error
func ThrowErrorString(msg string) {
	panic(fmt.Errorf("error: %s", msg))
}

// Throw throws any value with stack trace
func Throw(e interface{}) {
	panic(fmt.Errorf("thrown: %v", e))
}

// ThrowIf throws an error if the condition is true
func ThrowIf(cond bool, e interface{}) {
	if cond {
		Throw(e)
	}
}

// ThrowNil throws an error for nil values
func ThrowNil() {
	panic(fmt.Errorf("unexpected nil value"))
}

// ThrowErrorf throws a formatted error with stack trace
func ThrowErrorf(format string, args ...interface{}) {
	panic(fmt.Errorf(format, args...))
}

// Throwf throws a formatted string (alias for ThrowErrorf for backward compatibility)
func Throwf(format string, args ...interface{}) {
	ThrowErrorf(format, args...)
}

// Try executes a function and recovers from panics, converting them to errors
func Try(fn func()) (err error) {
	defer func() {
		if r := recover(); r != nil {
			if e, ok := r.(error); ok {
				err = fmt.Errorf("recovered from panic: %w", e)
			} else {
				err = fmt.Errorf("recovered from panic: %v", r)
			}
		}
	}()
	fn()
	return nil
}

// TryWithContext executes a function with context and recovers from panics
func TryWithContext(ctx interface{ Done() <-chan struct{} }, fn func()) (err error) {
	defer func() {
		if r := recover(); r != nil {
			if e, ok := r.(error); ok {
				err = fmt.Errorf("recovered from panic: %w", e)
			} else {
				err = fmt.Errorf("recovered from panic: %v", r)
			}
		}
	}()

	// Check if context is done before executing
	select {
	case <-ctx.Done():
		return fmt.Errorf("context cancelled: %w", ctx.(interface{ Err() error }).Err())
	default:
	}

	fn()
	return nil
}

// WrapErrorWithContext wraps an error with additional context
func WrapErrorWithContext(err error, context string) error {
	if err == nil {
		return nil
	}
	return fmt.Errorf("%s: %w", context, err)
}

// WrapErrorWithContextf wraps an error with formatted context
func WrapErrorWithContextf(err error, format string, args ...interface{}) error {
	if err == nil {
		return nil
	}
	return fmt.Errorf("%s: %w", fmt.Sprintf(format, args...), err)
}

// IsContextError checks if an error is a context cancellation error
func IsContextError(err error) bool {
	if err == nil {
		return false
	}
	return err.Error() == "context cancelled" || err.Error() == "context deadline exceeded"
}

// GetErrorContext returns additional context about an error
func GetErrorContext(err error) string {
	if err == nil {
		return ""
	}
	return fmt.Sprintf("Error: %v\nStack trace:\n%s", err, debug.Stack())
}

// CreateRDPError creates a new RDP error with context
func CreateRDPError(errType ErrorType, message string, cause error) *RDPError {
	return &RDPError{
		Type:    errType,
		Message: message,
		Cause:   cause,
		Context: make(map[string]interface{}),
	}
}

// CreateRDPErrorWithContext creates a new RDP error with additional context
func CreateRDPErrorWithContext(errType ErrorType, message string, cause error, context map[string]interface{}) *RDPError {
	return &RDPError{
		Type:    errType,
		Message: message,
		Cause:   cause,
		Context: context,
	}
}
