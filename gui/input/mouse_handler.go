package input

import (
	"fmt"
	"sync"
	"time"

	"github.com/kdsmith18542/gordp"
	"github.com/kdsmith18542/gordp/proto/t128"
)

// MouseHandler handles mouse input events
type MouseHandler struct {
	client        *gordp.Client
	displayWidget interface{} // Will be RDPDisplayWidget when Qt is integrated

	// Mouse state
	mu           sync.RWMutex
	lastX, lastY int
	isPressed    bool

	// Enhanced mouse management
	mouseStats       map[string]interface{}
	buttonStates     map[MouseButton]bool
	lastClickTime    map[MouseButton]time.Time
	doubleClickDelay time.Duration
	clickThreshold   int // pixels
	dragThreshold    int // pixels
	isDragging       bool
	dragStartX       int
	dragStartY       int
	dragStartTime    time.Time
}

// NewMouseHandler creates a new mouse handler
func NewMouseHandler(client *gordp.Client) *MouseHandler {
	handler := &MouseHandler{
		client:           client,
		mouseStats:       make(map[string]interface{}),
		buttonStates:     make(map[MouseButton]bool),
		lastClickTime:    make(map[MouseButton]time.Time),
		doubleClickDelay: 500 * time.Millisecond,
		clickThreshold:   5,  // pixels
		dragThreshold:    10, // pixels
	}

	// Initialize button states
	for button := MouseButtonLeft; button <= MouseButtonX2; button++ {
		handler.buttonStates[button] = false
		handler.lastClickTime[button] = time.Time{}
	}

	return handler
}

// HandleMouseMove handles mouse movement events
func (h *MouseHandler) HandleMouseMove(x, y int) {
	h.mu.Lock()
	defer h.mu.Unlock()

	if h.client == nil {
		return
	}

	// Validate coordinates
	if !h.validateCoordinates(x, y) {
		return
	}

	// Check if this is a drag operation
	if h.isDragging && h.isPressed {
		h.handleDragMove(x, y)
		return
	}

	h.lastX = x
	h.lastY = y

	// Create mouse move event
	event := t128.NewFastPathMouseMoveEvent(uint16(x), uint16(y))

	// Send to RDP client
	h.sendPointerEvent(event)

	h.updateMouseStats("moves", 1)
}

// HandleMousePress handles mouse press events
func (h *MouseHandler) HandleMousePress(x, y int, button MouseButton) {
	h.mu.Lock()
	defer h.mu.Unlock()

	if h.client == nil {
		return
	}

	// Validate coordinates
	if !h.validateCoordinates(x, y) {
		return
	}

	h.lastX = x
	h.lastY = y
	h.isPressed = true
	h.buttonStates[button] = true

	// Start drag detection
	h.dragStartX = x
	h.dragStartY = y
	h.dragStartTime = time.Now()
	h.isDragging = false

	// Create mouse press event
	event := t128.NewFastPathMouseButtonEvent(h.convertMouseButton(button), true, uint16(x), uint16(y))

	// Send to RDP client
	h.sendPointerEvent(event)

	h.updateMouseStats("presses", 1)
	h.updateMouseStats("button_"+button.String(), 1)
}

// HandleMouseRelease handles mouse release events
func (h *MouseHandler) HandleMouseRelease(x, y int, button MouseButton) {
	h.mu.Lock()
	defer h.mu.Unlock()

	if h.client == nil {
		return
	}

	// Validate coordinates
	if !h.validateCoordinates(x, y) {
		return
	}

	h.lastX = x
	h.lastY = y
	h.isPressed = false
	h.buttonStates[button] = false

	// Check for double click
	if h.isDoubleClick(button, x, y) {
		h.handleDoubleClick(button, x, y)
		return
	}

	// Check for drag completion
	if h.isDragging {
		h.handleDragEnd(button, x, y)
		return
	}

	// Regular click
	h.handleClick(button, x, y)

	// Create mouse release event
	event := t128.NewFastPathMouseButtonEvent(h.convertMouseButton(button), false, uint16(x), uint16(y))

	// Send to RDP client
	h.sendPointerEvent(event)

	h.updateMouseStats("releases", 1)
	h.updateMouseStats("clicks", 1)
}

// HandleMouseWheel handles mouse wheel events
func (h *MouseHandler) HandleMouseWheel(x, y int, delta int) {
	h.mu.Lock()
	defer h.mu.Unlock()

	if h.client == nil {
		return
	}

	// Validate coordinates
	if !h.validateCoordinates(x, y) {
		return
	}

	// Create mouse wheel event
	event := t128.NewFastPathMouseWheelEvent(int16(delta), uint16(x), uint16(y))

	// Send to RDP client
	h.sendPointerEvent(event)

	h.updateMouseStats("wheel_events", 1)
	h.updateMouseStats("wheel_delta", delta)
}

// HandleMouseHorizontalWheel handles horizontal mouse wheel events
func (h *MouseHandler) HandleMouseHorizontalWheel(x, y int, delta int) {
	h.mu.Lock()
	defer h.mu.Unlock()

	if h.client == nil {
		return
	}

	// Validate coordinates
	if !h.validateCoordinates(x, y) {
		return
	}

	// Create horizontal mouse wheel event
	event := t128.NewFastPathMouseHorizontalWheelEvent(int16(delta), uint16(x), uint16(y))

	// Send to RDP client
	h.sendPointerEvent(event)

	h.updateMouseStats("hwheel_events", 1)
	h.updateMouseStats("hwheel_delta", delta)
}

// convertMouseButton converts our MouseButton to t128.MouseButton
func (h *MouseHandler) convertMouseButton(button MouseButton) t128.MouseButton {
	switch button {
	case MouseButtonLeft:
		return t128.MouseButtonLeft
	case MouseButtonRight:
		return t128.MouseButtonRight
	case MouseButtonMiddle:
		return t128.MouseButtonMiddle
	case MouseButtonX1:
		return t128.MouseButtonX1
	case MouseButtonX2:
		return t128.MouseButtonX2
	default:
		return t128.MouseButtonLeft
	}
}

// validateCoordinates validates mouse coordinates
func (h *MouseHandler) validateCoordinates(x, y int) bool {
	// Basic coordinate validation
	if x < 0 || y < 0 {
		return false
	}

	// In a real implementation, you would check against display bounds
	// For now, we'll use reasonable limits
	if x > 65535 || y > 65535 {
		return false
	}

	return true
}

// isDoubleClick checks if this is a double click
func (h *MouseHandler) isDoubleClick(button MouseButton, x, y int) bool {
	lastTime := h.lastClickTime[button]
	now := time.Now()

	// Check time threshold
	if now.Sub(lastTime) > h.doubleClickDelay {
		return false
	}

	// Check distance threshold
	dx := x - h.lastX
	dy := y - h.lastY
	distance := dx*dx + dy*dy

	return distance <= h.clickThreshold*h.clickThreshold
}

// handleDoubleClick handles double click events
func (h *MouseHandler) handleDoubleClick(button MouseButton, x, y int) {
	// Send double click event
	if err := h.client.SendMouseDoubleClickEvent(h.convertMouseButton(button), uint16(x), uint16(y)); err != nil {
		fmt.Printf("Failed to send double click event: %v\n", err)
	} else {
		fmt.Printf("Double click: %s at (%d, %d)\n", button.String(), x, y)
	}

	h.updateMouseStats("double_clicks", 1)
	h.lastClickTime[button] = time.Now()
}

// handleClick handles single click events
func (h *MouseHandler) handleClick(button MouseButton, x, y int) {
	fmt.Printf("Click: %s at (%d, %d)\n", button.String(), x, y)
	h.lastClickTime[button] = time.Now()
}

// handleDragMove handles mouse movement during drag
func (h *MouseHandler) handleDragMove(x, y int) {
	// Check if we've exceeded drag threshold
	dx := x - h.dragStartX
	dy := y - h.dragStartY
	distance := dx*dx + dy*dy

	if !h.isDragging && distance > h.dragThreshold*h.dragThreshold {
		h.isDragging = true
		fmt.Printf("Drag started from (%d, %d)\n", h.dragStartX, h.dragStartY)
		h.updateMouseStats("drags", 1)
	}

	// Update position
	h.lastX = x
	h.lastY = y
}

// handleDragEnd handles drag end events
func (h *MouseHandler) handleDragEnd(button MouseButton, x, y int) {
	dragDuration := time.Since(h.dragStartTime)
	fmt.Printf("Drag ended: %s from (%d, %d) to (%d, %d), duration: %v\n",
		button.String(), h.dragStartX, h.dragStartY, x, y, dragDuration)

	h.isDragging = false
	h.updateMouseStats("drag_duration", dragDuration)
}

// sendPointerEvent sends a pointer event to the RDP client
func (h *MouseHandler) sendPointerEvent(event *t128.TsFpPointerEvent) {
	if h.client == nil {
		fmt.Println("RDP client is not initialized; cannot send mouse event")
		return
	}

	var err error

	// Determine event type and send using appropriate client method
	switch {
	case event.PointerFlags&t128.PTRFLAGS_MOVE != 0:
		// Mouse movement event
		err = h.client.SendMouseMoveEvent(event.XPos, event.YPos)

	case event.PointerFlags&t128.PTRFLAGS_WHEEL != 0:
		// Mouse wheel event
		wheelDelta := int16(event.PointerFlags & t128.WheelRotationMask)
		if event.PointerFlags&t128.PTRFLAGS_WHEEL_NEGATIVE != 0 {
			wheelDelta = -wheelDelta
		}
		err = h.client.SendMouseWheelEvent(wheelDelta, event.XPos, event.YPos)

	case event.PointerFlags&t128.PTRFLAGS_HWHEEL != 0:
		// Horizontal mouse wheel event
		wheelDelta := int16(event.PointerFlags & t128.WheelRotationMask)
		if event.PointerFlags&t128.PTRFLAGS_WHEEL_NEGATIVE != 0 {
			wheelDelta = -wheelDelta
		}
		err = h.client.SendMouseHorizontalWheelEvent(wheelDelta, event.XPos, event.YPos)

	case event.PointerFlags&t128.PTRFLAGS_DOWN != 0:
		// Mouse button down event
		var button t128.MouseButton
		switch {
		case event.PointerFlags&t128.PTRFLAGS_BUTTON1 != 0:
			button = t128.MouseButtonLeft
		case event.PointerFlags&t128.PTRFLAGS_BUTTON2 != 0:
			button = t128.MouseButtonRight
		case event.PointerFlags&t128.PTRFLAGS_BUTTON3 != 0:
			button = t128.MouseButtonMiddle
		case event.PointerFlags&t128.PTRFLAGS_BUTTON4 != 0:
			button = t128.MouseButtonX1
		case event.PointerFlags&t128.PTRFLAGS_BUTTON5 != 0:
			button = t128.MouseButtonX2
		default:
			err = fmt.Errorf("unknown mouse button in event")
		}
		if err == nil {
			err = h.client.SendMouseButtonEvent(button, true, event.XPos, event.YPos)
		}

	default:
		// Mouse button up event (no PTRFLAGS_DOWN)
		var button t128.MouseButton
		switch {
		case event.PointerFlags&t128.PTRFLAGS_BUTTON1 != 0:
			button = t128.MouseButtonLeft
		case event.PointerFlags&t128.PTRFLAGS_BUTTON2 != 0:
			button = t128.MouseButtonRight
		case event.PointerFlags&t128.PTRFLAGS_BUTTON3 != 0:
			button = t128.MouseButtonMiddle
		case event.PointerFlags&t128.PTRFLAGS_BUTTON4 != 0:
			button = t128.MouseButtonX1
		case event.PointerFlags&t128.PTRFLAGS_BUTTON5 != 0:
			button = t128.MouseButtonX2
		default:
			err = fmt.Errorf("unknown mouse button in event")
		}
		if err == nil {
			err = h.client.SendMouseButtonEvent(button, false, event.XPos, event.YPos)
		}
	}

	if err != nil {
		fmt.Printf("Failed to send mouse event: %v\n", err)
		h.updateMouseStats("send_errors", 1)
	} else {
		fmt.Printf("Mouse event sent: flags=0x%04x, x=%d, y=%d\n", event.PointerFlags, event.XPos, event.YPos)
		h.updateMouseStats("events_sent", 1)
	}
}

// SetClient sets the RDP client reference
func (h *MouseHandler) SetClient(client *gordp.Client) {
	h.mu.Lock()
	defer h.mu.Unlock()
	h.client = client
}

// GetClient returns the RDP client reference
func (h *MouseHandler) GetClient() *gordp.Client {
	h.mu.RLock()
	defer h.mu.RUnlock()
	return h.client
}

// GetMousePosition returns the current mouse position
func (h *MouseHandler) GetMousePosition() (int, int) {
	h.mu.RLock()
	defer h.mu.RUnlock()
	return h.lastX, h.lastY
}

// IsButtonPressed returns whether a specific button is pressed
func (h *MouseHandler) IsButtonPressed(button MouseButton) bool {
	h.mu.RLock()
	defer h.mu.RUnlock()
	return h.buttonStates[button]
}

// IsDragging returns whether the mouse is currently dragging
func (h *MouseHandler) IsDragging() bool {
	h.mu.RLock()
	defer h.mu.RUnlock()
	return h.isDragging
}

// GetMouseStats returns mouse statistics
func (h *MouseHandler) GetMouseStats() map[string]interface{} {
	h.mu.RLock()
	defer h.mu.RUnlock()

	stats := map[string]interface{}{
		"position_x":    h.lastX,
		"position_y":    h.lastY,
		"is_pressed":    h.isPressed,
		"is_dragging":   h.isDragging,
		"button_states": h.buttonStates,
	}

	// Merge with detailed stats
	for key, value := range h.mouseStats {
		stats[key] = value
	}

	return stats
}

// SetDoubleClickDelay sets the double click delay
func (h *MouseHandler) SetDoubleClickDelay(delay time.Duration) {
	h.mu.Lock()
	defer h.mu.Unlock()
	h.doubleClickDelay = delay
}

// SetClickThreshold sets the click threshold in pixels
func (h *MouseHandler) SetClickThreshold(threshold int) {
	h.mu.Lock()
	defer h.mu.Unlock()
	h.clickThreshold = threshold
}

// SetDragThreshold sets the drag threshold in pixels
func (h *MouseHandler) SetDragThreshold(threshold int) {
	h.mu.Lock()
	defer h.mu.Unlock()
	h.dragThreshold = threshold
}

// updateMouseStats updates mouse statistics
func (h *MouseHandler) updateMouseStats(key string, value interface{}) {
	h.mouseStats[key] = value
	h.mouseStats["last_update"] = time.Now()
}

// MouseButton represents mouse button types
type MouseButton int

const (
	MouseButtonLeft MouseButton = iota
	MouseButtonRight
	MouseButtonMiddle
	MouseButtonX1
	MouseButtonX2
)

// String returns the string representation of the mouse button
func (b MouseButton) String() string {
	switch b {
	case MouseButtonLeft:
		return "left"
	case MouseButtonRight:
		return "right"
	case MouseButtonMiddle:
		return "middle"
	case MouseButtonX1:
		return "x1"
	case MouseButtonX2:
		return "x2"
	default:
		return "unknown"
	}
}
