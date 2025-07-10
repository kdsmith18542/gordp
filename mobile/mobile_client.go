package mobile

import (
	"context"
	"fmt"
	"sync"
	"time"

	"github.com/kdsmith18542/gordp"
	"github.com/kdsmith18542/gordp/config"
	"github.com/kdsmith18542/gordp/proto/bitmap"
)

// MobileClient represents a mobile RDP client
type MobileClient struct {
	client    *gordp.Client
	config    *config.Config
	ctx       context.Context
	cancel    context.CancelFunc
	callbacks *MobileCallbacks
	status    ConnectionStatus
	lastError string

	// Mobile input handling
	inputMutex        sync.RWMutex
	touchState        *TouchState
	gestureRecognizer *GestureRecognizer
	keyboardLayout    *MobileKeyboardLayout
	inputStats        *InputStatistics
	hapticFeedback    *HapticFeedback
	mobileConfig      *MobileConfig
}

// TouchState tracks touch input state
type TouchState struct {
	ActiveTouches map[int]*TouchPoint
	LastTapTime   time.Time
	LastTapPos    TouchPoint
	DoubleTapTime time.Duration
	LongPressTime time.Duration
}

// TouchPoint represents a single touch point
type TouchPoint struct {
	ID        int
	X, Y      int
	Pressure  float64
	Timestamp time.Time
	StartTime time.Time
}

// GestureRecognizer handles mobile gesture recognition
type GestureRecognizer struct {
	PinchStartDistance float64
	PinchStartAngle    float64
	RotationStartAngle float64
	SwipeStartPos      TouchPoint
	SwipeThreshold     int
	PinchThreshold     float64
	RotationThreshold  float64
}

// MobileKeyboardLayout handles mobile keyboard input
type MobileKeyboardLayout struct {
	CurrentLayout  string
	ModifierKeys   map[int]bool
	KeyMap         map[int]int
	AutoComplete   bool
	PredictiveText bool
}

// InputStatistics tracks input performance metrics
type InputStatistics struct {
	TouchEvents    int64
	KeyEvents      int64
	GestureEvents  int64
	AverageLatency time.Duration
	LastEventTime  time.Time
	ErrorCount     int64
	SuccessCount   int64
}

// HapticFeedback handles mobile haptic feedback
type HapticFeedback struct {
	Enabled      bool
	Intensity    float64
	Pattern      []time.Duration
	LastFeedback time.Time
	MinInterval  time.Duration
}

// ConnectionStatus represents the connection status
type ConnectionStatus int

const (
	StatusDisconnected ConnectionStatus = iota
	StatusConnecting
	StatusConnected
	StatusError
)

// Touch event types
const (
	TouchDown = iota
	TouchMove
	TouchUp
)

// Gesture types
const (
	GestureTap = iota
	GestureDoubleTap
	GestureLongPress
	GesturePinch
	GestureRotate
	GestureSwipe
	GesturePan
)

// MobileCallbacks contains callback functions for mobile events
type MobileCallbacks struct {
	OnStatusChanged  func(status ConnectionStatus)
	OnBitmapReceived func(x, y, width, height int, data []byte)
	OnError          func(error string)
	OnConnected      func(width, height int)
	OnDisconnected   func()
	OnHapticFeedback func(intensity float64, pattern []time.Duration)
	OnGesture        func(gestureType int, data map[string]interface{})
}

// NewMobileClient creates a new mobile RDP client
func NewMobileClient() *MobileClient {
	ctx, cancel := context.WithCancel(context.Background())

	client := &MobileClient{
		ctx:       ctx,
		cancel:    cancel,
		status:    StatusDisconnected,
		callbacks: &MobileCallbacks{},
		touchState: &TouchState{
			ActiveTouches: make(map[int]*TouchPoint),
			DoubleTapTime: 300 * time.Millisecond,
			LongPressTime: 500 * time.Millisecond,
		},
		gestureRecognizer: &GestureRecognizer{
			SwipeThreshold:    50,
			PinchThreshold:    20.0,
			RotationThreshold: 15.0,
		},
		keyboardLayout: &MobileKeyboardLayout{
			CurrentLayout:  "en_US",
			ModifierKeys:   make(map[int]bool),
			KeyMap:         make(map[int]int),
			AutoComplete:   true,
			PredictiveText: true,
		},
		inputStats: &InputStatistics{
			AverageLatency: 0,
		},
		hapticFeedback: &HapticFeedback{
			Enabled:     true,
			Intensity:   0.5,
			MinInterval: 100 * time.Millisecond,
		},
		mobileConfig: DefaultMobileConfig(),
	}

	// Initialize keyboard layout
	client.initializeKeyboardLayout()

	return client
}

// SetCallbacks sets the callback functions
func (mc *MobileClient) SetCallbacks(callbacks *MobileCallbacks) {
	mc.callbacks = callbacks
}

// Connect connects to an RDP server
func (mc *MobileClient) Connect(host, username, password string, port, width, height int) error {
	mc.updateStatus(StatusConnecting)

	// Create configuration
	mc.config = &config.Config{
		Connection: config.ConnectionConfig{
			Address: host,
			Port:    port,
		},
		Authentication: config.AuthConfig{
			Username: username,
			Password: password,
		},
		Display: config.DisplayConfig{
			Width:      width,
			Height:     height,
			ColorDepth: 24,
		},
		Performance: config.PerformanceConfig{
			BitmapCache: true,
			Compression: true,
		},
	}

	// Create RDP client
	mc.client = gordp.NewClient(&gordp.Option{
		Addr:           fmt.Sprintf("%s:%d", host, port),
		UserName:       username,
		Password:       password,
		ConnectTimeout: 10 * time.Second,
	})

	// Connect to server
	if err := mc.client.ConnectWithContext(mc.ctx); err != nil {
		mc.lastError = err.Error()
		mc.updateStatus(StatusError)
		if mc.callbacks.OnError != nil {
			mc.callbacks.OnError(err.Error())
		}
		return err
	}

	mc.updateStatus(StatusConnected)
	if mc.callbacks.OnConnected != nil {
		mc.callbacks.OnConnected(width, height)
	}

	// Start RDP session
	processor := &MobileBitmapProcessor{client: mc}
	if err := mc.client.RunWithContext(mc.ctx, processor); err != nil {
		mc.lastError = err.Error()
		mc.updateStatus(StatusError)
		if mc.callbacks.OnError != nil {
			mc.callbacks.OnError(err.Error())
		}
		return err
	}

	return nil
}

// Disconnect disconnects from the RDP server
func (mc *MobileClient) Disconnect() {
	if mc.client != nil {
		mc.client.Close()
	}
	mc.cancel()
	mc.updateStatus(StatusDisconnected)
	if mc.callbacks.OnDisconnected != nil {
		mc.callbacks.OnDisconnected()
	}
}

// SendKeyPress sends a key press event with mobile keyboard support
func (mc *MobileClient) SendKeyPress(keyCode int, down bool) error {
	mc.inputMutex.Lock()
	defer mc.inputMutex.Unlock()

	if mc.client == nil || mc.status != StatusConnected {
		return fmt.Errorf("not connected")
	}

	startTime := time.Now()
	defer func() {
		mc.updateInputStats(startTime, true)
	}()

	// Convert mobile key code to RDP virtual key code
	rdpKeyCode := mc.convertMobileKeyToRDP(keyCode)

	// Handle modifier keys
	if mc.isModifierKey(rdpKeyCode) {
		mc.keyboardLayout.ModifierKeys[rdpKeyCode] = down
	}

	// Apply modifier key combinations
	finalKeyCode := mc.applyModifierKeys(rdpKeyCode)

	// Send key event to RDP server
	if err := mc.sendRDPKeyEvent(finalKeyCode, down); err != nil {
		mc.updateInputStats(startTime, false)
		return fmt.Errorf("failed to send key event: %w", err)
	}

	// Provide haptic feedback for key presses
	if down && mc.hapticFeedback.Enabled {
		mc.triggerHapticFeedback(0.3, []time.Duration{50 * time.Millisecond})
	}

	return nil
}

// SendMouseMove sends a mouse move event with touch conversion
func (mc *MobileClient) SendMouseMove(x, y int) error {
	mc.inputMutex.Lock()
	defer mc.inputMutex.Unlock()

	if mc.client == nil || mc.status != StatusConnected {
		return fmt.Errorf("not connected")
	}

	startTime := time.Now()
	defer func() {
		mc.updateInputStats(startTime, true)
	}()

	// Convert coordinates to RDP format
	rdpX, rdpY := mc.convertCoordinates(x, y)

	// Send mouse move event to RDP server
	if err := mc.sendRDPMouseEvent(rdpX, rdpY, 0, false, false); err != nil {
		mc.updateInputStats(startTime, false)
		return fmt.Errorf("failed to send mouse move: %w", err)
	}

	return nil
}

// SendMouseClick sends a mouse click event with touch conversion
func (mc *MobileClient) SendMouseClick(button int, down bool, x, y int) error {
	mc.inputMutex.Lock()
	defer mc.inputMutex.Unlock()

	if mc.client == nil || mc.status != StatusConnected {
		return fmt.Errorf("not connected")
	}

	startTime := time.Now()
	defer func() {
		mc.updateInputStats(startTime, true)
	}()

	// Convert coordinates to RDP format
	rdpX, rdpY := mc.convertCoordinates(x, y)

	// Convert button to RDP format
	rdpButton := mc.convertButtonToRDP(button)

	// Send mouse click event to RDP server
	if err := mc.sendRDPMouseEvent(rdpX, rdpY, rdpButton, down, false); err != nil {
		mc.updateInputStats(startTime, false)
		return fmt.Errorf("failed to send mouse click: %w", err)
	}

	// Provide haptic feedback for clicks
	if down && mc.hapticFeedback.Enabled {
		mc.triggerHapticFeedback(0.5, []time.Duration{100 * time.Millisecond})
	}

	return nil
}

// SendTouch sends a touch event with gesture recognition
func (mc *MobileClient) SendTouch(x, y int, touchType int) error {
	mc.inputMutex.Lock()
	defer mc.inputMutex.Unlock()

	if mc.client == nil || mc.status != StatusConnected {
		return fmt.Errorf("not connected")
	}

	startTime := time.Now()
	defer func() {
		mc.updateInputStats(startTime, true)
	}()

	// Handle touch event based on type
	switch touchType {
	case TouchDown:
		return mc.handleTouchDown(x, y)
	case TouchMove:
		return mc.handleTouchMove(x, y)
	case TouchUp:
		return mc.handleTouchUp(x, y)
	default:
		return fmt.Errorf("unknown touch type: %d", touchType)
	}
}

// handleTouchDown processes touch down events
func (mc *MobileClient) handleTouchDown(x, y int) error {
	touchID := len(mc.touchState.ActiveTouches) + 1

	touchPoint := &TouchPoint{
		ID:        touchID,
		X:         x,
		Y:         y,
		Pressure:  1.0,
		Timestamp: time.Now(),
		StartTime: time.Now(),
	}

	mc.touchState.ActiveTouches[touchID] = touchPoint

	// Convert to mouse click for single touch
	if len(mc.touchState.ActiveTouches) == 1 {
		return mc.SendMouseClick(0, true, x, y)
	}

	// Handle multi-touch gestures
	if len(mc.touchState.ActiveTouches) == 2 {
		return mc.handleMultiTouchGesture()
	}

	return nil
}

// handleTouchMove processes touch move events
func (mc *MobileClient) handleTouchMove(x, y int) error {
	// Update touch point position
	for _, touch := range mc.touchState.ActiveTouches {
		touch.X = x
		touch.Y = y
		touch.Timestamp = time.Now()
		break // For single touch, update the first touch point
	}

	// Convert to mouse move for single touch
	if len(mc.touchState.ActiveTouches) == 1 {
		return mc.SendMouseMove(x, y)
	}

	// Handle multi-touch gestures
	if len(mc.touchState.ActiveTouches) == 2 {
		return mc.handleMultiTouchGesture()
	}

	return nil
}

// handleTouchUp processes touch up events
func (mc *MobileClient) handleTouchUp(x, y int) error {
	// Remove touch point
	for id, touch := range mc.touchState.ActiveTouches {
		if touch.X == x && touch.Y == y {
			delete(mc.touchState.ActiveTouches, id)
			break
		}
	}

	// Convert to mouse click for single touch
	if len(mc.touchState.ActiveTouches) == 0 {
		return mc.SendMouseClick(0, false, x, y)
	}

	// Handle multi-touch gestures
	if len(mc.touchState.ActiveTouches) == 1 {
		return mc.handleMultiTouchGesture()
	}

	return nil
}

// handleMultiTouchGesture processes multi-touch gestures
func (mc *MobileClient) handleMultiTouchGesture() error {
	if len(mc.touchState.ActiveTouches) != 2 {
		return nil
	}

	var touches []*TouchPoint
	for _, touch := range mc.touchState.ActiveTouches {
		touches = append(touches, touch)
	}

	// Calculate gesture parameters
	distance := mc.calculateDistance(touches[0], touches[1])
	angle := mc.calculateAngle(touches[0], touches[1])

	// Detect pinch gesture
	if mc.mobileConfig.EnablePinchGesture {
		if distance != mc.gestureRecognizer.PinchStartDistance {
			scale := distance / mc.gestureRecognizer.PinchStartDistance
			if scale > 1.1 || scale < 0.9 {
				mc.triggerGesture(GesturePinch, map[string]interface{}{
					"scale":    scale,
					"center_x": (touches[0].X + touches[1].X) / 2,
					"center_y": (touches[0].Y + touches[1].Y) / 2,
				})
			}
		}
	}

	// Detect rotation gesture
	if mc.mobileConfig.EnableRotateGesture {
		if angle != mc.gestureRecognizer.RotationStartAngle {
			rotation := angle - mc.gestureRecognizer.RotationStartAngle
			if rotation > mc.gestureRecognizer.RotationThreshold || rotation < -mc.gestureRecognizer.RotationThreshold {
				mc.triggerGesture(GestureRotate, map[string]interface{}{
					"rotation": rotation,
					"center_x": (touches[0].X + touches[1].X) / 2,
					"center_y": (touches[0].Y + touches[1].Y) / 2,
				})
			}
		}
	}

	return nil
}

// SendGesture sends a gesture event
func (mc *MobileClient) SendGesture(gestureType int, data map[string]interface{}) error {
	mc.inputMutex.Lock()
	defer mc.inputMutex.Unlock()

	if mc.client == nil || mc.status != StatusConnected {
		return fmt.Errorf("not connected")
	}

	startTime := time.Now()
	defer func() {
		mc.updateInputStats(startTime, true)
	}()

	// Handle different gesture types
	switch gestureType {
	case GestureTap:
		return mc.handleTapGesture(data)
	case GestureDoubleTap:
		return mc.handleDoubleTapGesture(data)
	case GestureLongPress:
		return mc.handleLongPressGesture(data)
	case GesturePinch:
		return mc.handlePinchGesture(data)
	case GestureRotate:
		return mc.handleRotateGesture(data)
	case GestureSwipe:
		return mc.handleSwipeGesture(data)
	default:
		return fmt.Errorf("unknown gesture type: %d", gestureType)
	}
}

// handleTapGesture processes tap gestures
func (mc *MobileClient) handleTapGesture(data map[string]interface{}) error {
	x, _ := data["x"].(int)
	y, _ := data["y"].(int)

	// Convert to mouse click
	if err := mc.SendMouseClick(0, true, x, y); err != nil {
		return err
	}

	// Small delay for tap effect
	time.Sleep(50 * time.Millisecond)

	return mc.SendMouseClick(0, false, x, y)
}

// handleDoubleTapGesture processes double tap gestures
func (mc *MobileClient) handleDoubleTapGesture(data map[string]interface{}) error {
	x, _ := data["x"].(int)
	y, _ := data["y"].(int)

	// Convert to double click
	if err := mc.SendMouseClick(0, true, x, y); err != nil {
		return err
	}
	time.Sleep(50 * time.Millisecond)
	if err := mc.SendMouseClick(0, false, x, y); err != nil {
		return err
	}
	time.Sleep(50 * time.Millisecond)
	if err := mc.SendMouseClick(0, true, x, y); err != nil {
		return err
	}
	time.Sleep(50 * time.Millisecond)
	return mc.SendMouseClick(0, false, x, y)
}

// handleLongPressGesture processes long press gestures
func (mc *MobileClient) handleLongPressGesture(data map[string]interface{}) error {
	x, _ := data["x"].(int)
	y, _ := data["y"].(int)

	// Convert to right click (context menu)
	return mc.SendMouseClick(1, true, x, y)
}

// handlePinchGesture processes pinch gestures
func (mc *MobileClient) handlePinchGesture(data map[string]interface{}) error {
	scale, _ := data["scale"].(float64)

	// Convert pinch to zoom commands
	if scale > 1.0 {
		// Zoom in - send Ctrl+Plus
		if err := mc.SendKeyPress(0x11, true); err != nil { // Ctrl
			return err
		}
		if err := mc.SendKeyPress(0xBB, true); err != nil { // Plus
			return err
		}
		time.Sleep(50 * time.Millisecond)
		if err := mc.SendKeyPress(0xBB, false); err != nil { // Plus
			return err
		}
		return mc.SendKeyPress(0x11, false) // Ctrl
	} else {
		// Zoom out - send Ctrl+Minus
		if err := mc.SendKeyPress(0x11, true); err != nil { // Ctrl
			return err
		}
		if err := mc.SendKeyPress(0xBD, true); err != nil { // Minus
			return err
		}
		time.Sleep(50 * time.Millisecond)
		if err := mc.SendKeyPress(0xBD, false); err != nil { // Minus
			return err
		}
		return mc.SendKeyPress(0x11, false) // Ctrl
	}
}

// handleRotateGesture processes rotation gestures
func (mc *MobileClient) handleRotateGesture(data map[string]interface{}) error {
	rotation, _ := data["rotation"].(float64)

	// Convert rotation to scroll commands
	if rotation > 0 {
		// Rotate clockwise - scroll down
		return mc.sendScrollEvent(0, -120) // Negative for down
	} else {
		// Rotate counter-clockwise - scroll up
		return mc.sendScrollEvent(0, 120) // Positive for up
	}
}

// handleSwipeGesture processes swipe gestures
func (mc *MobileClient) handleSwipeGesture(data map[string]interface{}) error {
	direction, _ := data["direction"].(string)

	// Convert swipe to arrow keys
	switch direction {
	case "up":
		return mc.SendKeyPress(0x26, true) // Up arrow
	case "down":
		return mc.SendKeyPress(0x28, true) // Down arrow
	case "left":
		return mc.SendKeyPress(0x25, true) // Left arrow
	case "right":
		return mc.SendKeyPress(0x27, true) // Right arrow
	}

	return nil
}

// GetStatus returns the current connection status
func (mc *MobileClient) GetStatus() ConnectionStatus {
	return mc.status
}

// GetLastError returns the last error message
func (mc *MobileClient) GetLastError() string {
	return mc.lastError
}

// GetBitmapCacheStats returns bitmap cache statistics
func (mc *MobileClient) GetBitmapCacheStats() map[string]interface{} {
	if mc.client == nil {
		return nil
	}
	return mc.client.GetBitmapCacheStats()
}

// ClearBitmapCache clears the bitmap cache
func (mc *MobileClient) ClearBitmapCache() {
	if mc.client != nil {
		mc.client.ClearBitmapCache()
	}
}

// GetInputStatistics returns input performance statistics
func (mc *MobileClient) GetInputStatistics() map[string]interface{} {
	mc.inputMutex.RLock()
	defer mc.inputMutex.RUnlock()

	return map[string]interface{}{
		"touch_events":    mc.inputStats.TouchEvents,
		"key_events":      mc.inputStats.KeyEvents,
		"gesture_events":  mc.inputStats.GestureEvents,
		"average_latency": mc.inputStats.AverageLatency,
		"error_count":     mc.inputStats.ErrorCount,
		"success_count":   mc.inputStats.SuccessCount,
		"success_rate":    float64(mc.inputStats.SuccessCount) / float64(mc.inputStats.SuccessCount+mc.inputStats.ErrorCount),
	}
}

// SetMobileConfig sets mobile-specific configuration
func (mc *MobileClient) SetMobileConfig(config *MobileConfig) {
	mc.inputMutex.Lock()
	defer mc.inputMutex.Unlock()
	mc.mobileConfig = config
}

// GetMobileConfig returns current mobile configuration
func (mc *MobileClient) GetMobileConfig() *MobileConfig {
	mc.inputMutex.RLock()
	defer mc.inputMutex.RUnlock()
	return mc.mobileConfig
}

// updateStatus updates the connection status
func (mc *MobileClient) updateStatus(status ConnectionStatus) {
	mc.status = status
	if mc.callbacks.OnStatusChanged != nil {
		mc.callbacks.OnStatusChanged(status)
	}
}

// updateInputStats updates input performance statistics
func (mc *MobileClient) updateInputStats(startTime time.Time, success bool) {
	latency := time.Since(startTime)

	if success {
		mc.inputStats.SuccessCount++
	} else {
		mc.inputStats.ErrorCount++
	}

	// Update average latency
	if mc.inputStats.SuccessCount > 0 {
		mc.inputStats.AverageLatency = (mc.inputStats.AverageLatency*time.Duration(mc.inputStats.SuccessCount-1) + latency) / time.Duration(mc.inputStats.SuccessCount)
	}

	mc.inputStats.LastEventTime = time.Now()
}

// triggerHapticFeedback triggers haptic feedback
func (mc *MobileClient) triggerHapticFeedback(intensity float64, pattern []time.Duration) {
	if !mc.hapticFeedback.Enabled {
		return
	}

	// Check minimum interval
	if time.Since(mc.hapticFeedback.LastFeedback) < mc.hapticFeedback.MinInterval {
		return
	}

	mc.hapticFeedback.LastFeedback = time.Now()

	if mc.callbacks.OnHapticFeedback != nil {
		mc.callbacks.OnHapticFeedback(intensity, pattern)
	}
}

// triggerGesture triggers a gesture event
func (mc *MobileClient) triggerGesture(gestureType int, data map[string]interface{}) {
	if mc.callbacks.OnGesture != nil {
		mc.callbacks.OnGesture(gestureType, data)
	}
}

// initializeKeyboardLayout initializes mobile keyboard layout
func (mc *MobileClient) initializeKeyboardLayout() {
	// Map common mobile keys to RDP virtual key codes
	mc.keyboardLayout.KeyMap = map[int]int{
		// Navigation keys
		0x26: 0x26, // Up arrow
		0x28: 0x28, // Down arrow
		0x25: 0x25, // Left arrow
		0x27: 0x27, // Right arrow

		// Function keys
		0x70: 0x70, // F1
		0x71: 0x71, // F2
		0x72: 0x72, // F3
		0x73: 0x73, // F4
		0x74: 0x74, // F5
		0x75: 0x75, // F6
		0x76: 0x76, // F7
		0x77: 0x77, // F8
		0x78: 0x78, // F9
		0x79: 0x79, // F10
		0x7A: 0x7A, // F11
		0x7B: 0x7B, // F12

		// Modifier keys
		0x11: 0x11, // Ctrl
		0x12: 0x12, // Alt
		0x10: 0x10, // Shift
		0x5B: 0x5B, // Windows key

		// Special keys
		0x08: 0x08, // Backspace
		0x09: 0x09, // Tab
		0x0D: 0x0D, // Enter
		0x1B: 0x1B, // Escape
		0x20: 0x20, // Space
		0x2E: 0x2E, // Delete
		0x2D: 0x2D, // Insert
		0x21: 0x21, // Page Up
		0x22: 0x22, // Page Down
		0x24: 0x24, // Home
		0x23: 0x23, // End
	}
}

// convertMobileKeyToRDP converts mobile key code to RDP virtual key code
func (mc *MobileClient) convertMobileKeyToRDP(keyCode int) int {
	if rdpKey, exists := mc.keyboardLayout.KeyMap[keyCode]; exists {
		return rdpKey
	}
	return keyCode
}

// isModifierKey checks if a key is a modifier key
func (mc *MobileClient) isModifierKey(keyCode int) bool {
	modifierKeys := []int{0x11, 0x12, 0x10, 0x5B} // Ctrl, Alt, Shift, Windows
	for _, modKey := range modifierKeys {
		if keyCode == modKey {
			return true
		}
	}
	return false
}

// applyModifierKeys applies modifier key combinations
func (mc *MobileClient) applyModifierKeys(keyCode int) int {
	// For now, just return the key code as-is
	// In a full implementation, this would handle key combinations
	return keyCode
}

// convertCoordinates converts screen coordinates to RDP coordinates
func (mc *MobileClient) convertCoordinates(x, y int) (int, int) {
	// For now, assume 1:1 mapping
	// In a full implementation, this would handle scaling and offset
	return x, y
}

// convertButtonToRDP converts button to RDP format
func (mc *MobileClient) convertButtonToRDP(button int) int {
	switch button {
	case 0:
		return 0x01 // Left button
	case 1:
		return 0x02 // Right button
	case 2:
		return 0x04 // Middle button
	default:
		return 0x01 // Default to left button
	}
}

// sendRDPKeyEvent sends key event to RDP server
func (mc *MobileClient) sendRDPKeyEvent(keyCode int, down bool) error {
	// This would send the actual RDP key event
	// For now, we'll use the existing RDP client methods
	if mc.client != nil {
		// Use the RDP client's input handling
		// This is a placeholder - actual implementation would use RDP protocol
		return nil
	}
	return fmt.Errorf("RDP client not available")
}

// sendRDPMouseEvent sends mouse event to RDP server
func (mc *MobileClient) sendRDPMouseEvent(x, y, button int, down, wheel bool) error {
	// This would send the actual RDP mouse event
	// For now, we'll use the existing RDP client methods
	if mc.client != nil {
		// Use the RDP client's input handling
		// This is a placeholder - actual implementation would use RDP protocol
		return nil
	}
	return fmt.Errorf("RDP client not available")
}

// sendScrollEvent sends scroll event to RDP server
func (mc *MobileClient) sendScrollEvent(x, y int) error {
	// This would send the actual RDP scroll event
	// For now, we'll use the existing RDP client methods
	if mc.client != nil {
		// Use the RDP client's input handling
		// This is a placeholder - actual implementation would use RDP protocol
		return nil
	}
	return fmt.Errorf("RDP client not available")
}

// calculateDistance calculates distance between two touch points
func (mc *MobileClient) calculateDistance(p1, p2 *TouchPoint) float64 {
	dx := float64(p2.X - p1.X)
	dy := float64(p2.Y - p1.Y)
	return (dx*dx + dy*dy)
}

// calculateAngle calculates angle between two touch points
func (mc *MobileClient) calculateAngle(p1, p2 *TouchPoint) float64 {
	dx := float64(p2.X - p1.X)
	dy := float64(p2.Y - p1.Y)
	return atan2(dy, dx) * 180 / 3.14159
}

// atan2 calculates arctangent of y/x
func atan2(y, x float64) float64 {
	if x > 0 {
		return atan(y / x)
	} else if x < 0 && y >= 0 {
		return atan(y/x) + 3.14159
	} else if x < 0 && y < 0 {
		return atan(y/x) - 3.14159
	} else if x == 0 && y > 0 {
		return 3.14159 / 2
	} else if x == 0 && y < 0 {
		return -3.14159 / 2
	} else if x == 0 && y == 0 {
		return 0
	}
	return 0
}

// atan calculates arctangent
func atan(x float64) float64 {
	// Simple approximation of arctangent
	if x < -1 {
		return -3.14159/2 - atan(1/x)
	} else if x > 1 {
		return 3.14159/2 - atan(1/x)
	} else {
		return x - x*x*x/3 + x*x*x*x*x/5 - x*x*x*x*x*x*x/7
	}
}

// MobileBitmapProcessor processes bitmap data for mobile clients
type MobileBitmapProcessor struct {
	client *MobileClient
}

// ProcessBitmap processes bitmap data and sends it to the mobile client
func (p *MobileBitmapProcessor) ProcessBitmap(option *bitmap.Option, bitmap *bitmap.BitMap) {
	// Convert bitmap to PNG for mobile transmission
	pngData := bitmap.ToPng()

	if p.client.callbacks.OnBitmapReceived != nil {
		p.client.callbacks.OnBitmapReceived(
			option.Left,
			option.Top,
			option.Width,
			option.Height,
			pngData,
		)
	}
}

// MobileConfig contains mobile-specific configuration
type MobileConfig struct {
	EnableTouchInput          bool `json:"enable_touch_input"`
	EnableGestureInput        bool `json:"enable_gesture_input"`
	EnableAccelerometer       bool `json:"enable_accelerometer"`
	EnableGyroscope           bool `json:"enable_gyroscope"`
	EnableVibration           bool `json:"enable_vibration"`
	EnableAudio               bool `json:"enable_audio"`
	EnableClipboard           bool `json:"enable_clipboard"`
	EnableFileTransfer        bool `json:"enable_file_transfer"`
	EnablePrinting            bool `json:"enable_printing"`
	EnableSmartCard           bool `json:"enable_smart_card"`
	EnableUSB                 bool `json:"enable_usb"`
	EnableSerial              bool `json:"enable_serial"`
	EnableParallel            bool `json:"enable_parallel"`
	EnableModem               bool `json:"enable_modem"`
	EnableAudioInput          bool `json:"enable_audio_input"`
	EnableVideoInput          bool `json:"enable_video_input"`
	EnableCamera              bool `json:"enable_camera"`
	EnableMicrophone          bool `json:"enable_microphone"`
	EnableSpeaker             bool `json:"enable_speaker"`
	EnableHeadset             bool `json:"enable_headset"`
	EnableBluetooth           bool `json:"enable_bluetooth"`
	EnableWiFi                bool `json:"enable_wifi"`
	EnableCellular            bool `json:"enable_cellular"`
	EnableGPS                 bool `json:"enable_gps"`
	EnableNFC                 bool `json:"enable_nfc"`
	EnableFingerprint         bool `json:"enable_fingerprint"`
	EnableFaceID              bool `json:"enable_face_id"`
	EnableTouchID             bool `json:"enable_touch_id"`
	EnableVoiceControl        bool `json:"enable_voice_control"`
	EnableSiri                bool `json:"enable_siri"`
	EnableGoogleAssistant     bool `json:"enable_google_assistant"`
	EnableAlexa               bool `json:"enable_alexa"`
	EnableCortana             bool `json:"enable_cortana"`
	EnableBixby               bool `json:"enable_bixby"`
	EnableHapticFeedback      bool `json:"enable_haptic_feedback"`
	EnableForceTouch          bool `json:"enable_force_touch"`
	Enable3DTouch             bool `json:"enable_3d_touch"`
	EnableLongPress           bool `json:"enable_long_press"`
	EnableDoubleTap           bool `json:"enable_double_tap"`
	EnablePinchToZoom         bool `json:"enable_pinch_to_zoom"`
	EnableSwipe               bool `json:"enable_swipe"`
	EnableRotate              bool `json:"enable_rotate"`
	EnableShake               bool `json:"enable_shake"`
	EnableTilt                bool `json:"enable_tilt"`
	EnablePan                 bool `json:"enable_pan"`
	EnableScroll              bool `json:"enable_scroll"`
	EnableZoom                bool `json:"enable_zoom"`
	EnableRotateGesture       bool `json:"enable_rotate_gesture"`
	EnablePinchGesture        bool `json:"enable_pinch_gesture"`
	EnableSwipeGesture        bool `json:"enable_swipe_gesture"`
	EnableTapGesture          bool `json:"enable_tap_gesture"`
	EnableLongPressGesture    bool `json:"enable_long_press_gesture"`
	EnableDoubleTapGesture    bool `json:"enable_double_tap_gesture"`
	EnableTripleTapGesture    bool `json:"enable_triple_tap_gesture"`
	EnableQuadrupleTapGesture bool `json:"enable_quadruple_tap_gesture"`
	EnableQuintupleTapGesture bool `json:"enable_quintuple_tap_gesture"`
}

// DefaultMobileConfig returns default mobile configuration
func DefaultMobileConfig() *MobileConfig {
	return &MobileConfig{
		EnableTouchInput:          true,
		EnableGestureInput:        true,
		EnableAccelerometer:       false,
		EnableGyroscope:           false,
		EnableVibration:           true,
		EnableAudio:               true,
		EnableClipboard:           true,
		EnableFileTransfer:        false,
		EnablePrinting:            false,
		EnableSmartCard:           false,
		EnableUSB:                 false,
		EnableSerial:              false,
		EnableParallel:            false,
		EnableModem:               false,
		EnableAudioInput:          false,
		EnableVideoInput:          false,
		EnableCamera:              false,
		EnableMicrophone:          false,
		EnableSpeaker:             true,
		EnableHeadset:             true,
		EnableBluetooth:           false,
		EnableWiFi:                false,
		EnableCellular:            false,
		EnableGPS:                 false,
		EnableNFC:                 false,
		EnableFingerprint:         false,
		EnableFaceID:              false,
		EnableTouchID:             false,
		EnableVoiceControl:        false,
		EnableSiri:                false,
		EnableGoogleAssistant:     false,
		EnableAlexa:               false,
		EnableCortana:             false,
		EnableBixby:               false,
		EnableHapticFeedback:      true,
		EnableForceTouch:          false,
		Enable3DTouch:             false,
		EnableLongPress:           true,
		EnableDoubleTap:           true,
		EnablePinchToZoom:         true,
		EnableSwipe:               true,
		EnableRotate:              false,
		EnableShake:               false,
		EnableTilt:                false,
		EnablePan:                 true,
		EnableScroll:              true,
		EnableZoom:                true,
		EnableRotateGesture:       false,
		EnablePinchGesture:        true,
		EnableSwipeGesture:        true,
		EnableTapGesture:          true,
		EnableLongPressGesture:    true,
		EnableDoubleTapGesture:    true,
		EnableTripleTapGesture:    false,
		EnableQuadrupleTapGesture: false,
		EnableQuintupleTapGesture: false,
	}
}
