package input

import (
	"fmt"
	"sync"
	"time"
	"unicode"

	"github.com/kdsmith18542/gordp"
	"github.com/kdsmith18542/gordp/proto/t128"
)

// KeyboardHandler handles keyboard input events
type KeyboardHandler struct {
	client       *gordp.Client
	modifierKeys t128.ModifierKey

	// Keyboard state
	mu        sync.RWMutex
	keyStates map[uint8]bool

	// Enhanced keyboard management
	keyboardStats    map[string]interface{}
	lastKeyPressTime map[uint8]time.Time
	keyRepeatDelay   time.Duration
	keyRepeatRate    time.Duration
	isKeyRepeating   bool
	repeatKeyCode    uint8
	repeatTimer      *time.Timer

	// Keyboard layout and IME support
	currentLayout   string
	imeEnabled      bool
	imeMode         string
	capsLockState   bool
	numLockState    bool
	scrollLockState bool

	// Input buffering for performance
	inputBuffer   []t128.TsFpKeyboardEvent
	bufferMutex   sync.Mutex
	bufferSize    int
	flushTimer    *time.Timer
	flushInterval time.Duration
}

// NewKeyboardHandler creates a new keyboard handler
func NewKeyboardHandler(client *gordp.Client) *KeyboardHandler {
	handler := &KeyboardHandler{
		client:           client,
		keyStates:        make(map[uint8]bool),
		keyboardStats:    make(map[string]interface{}),
		lastKeyPressTime: make(map[uint8]time.Time),
		keyRepeatDelay:   500 * time.Millisecond,
		keyRepeatRate:    50 * time.Millisecond,
		currentLayout:    "US",
		imeEnabled:       false,
		imeMode:          "NONE",
		bufferSize:       100,
		flushInterval:    16 * time.Millisecond, // ~60 FPS
	}

	// Initialize statistics
	handler.initializeStats()

	// Start flush timer
	handler.flushTimer = time.AfterFunc(handler.flushInterval, handler.flushInputBuffer)

	return handler
}

// HandleKeyPress handles key press events
func (h *KeyboardHandler) HandleKeyPress(keyCode uint8, isExtended bool) {
	h.mu.Lock()
	defer h.mu.Unlock()

	if h.client == nil {
		return
	}

	// Update key state
	h.keyStates[keyCode] = true
	h.lastKeyPressTime[keyCode] = time.Now()

	// Update modifier keys
	h.updateModifierKeys(keyCode, true)

	// Handle special keys
	h.handleSpecialKeyPress(keyCode)

	// Create keyboard event
	event := t128.NewFastPathKeyboardEvent(keyCode, true)
	if isExtended {
		// Set extended key flag if needed
		event.EventHeader |= 0x80
	}

	// Buffer the event for performance
	h.bufferInputEvent(event)

	h.updateKeyboardStats("presses", 1)
	h.updateKeyboardStats("key_"+fmt.Sprintf("0x%02x", keyCode), 1)
}

// HandleKeyRelease handles key release events
func (h *KeyboardHandler) HandleKeyRelease(keyCode uint8, isExtended bool) {
	h.mu.Lock()
	defer h.mu.Unlock()

	if h.client == nil {
		return
	}

	// Update key state
	h.keyStates[keyCode] = false

	// Update modifier keys
	h.updateModifierKeys(keyCode, false)

	// Handle special key release
	h.handleSpecialKeyRelease(keyCode)

	// Stop key repeat if this is the repeating key
	if h.isKeyRepeating && h.repeatKeyCode == keyCode {
		h.stopKeyRepeat()
	}

	// Create keyboard event
	event := t128.NewFastPathKeyboardEvent(keyCode, false)
	if isExtended {
		// Set extended key flag if needed
		event.EventHeader |= 0x80
	}

	// Buffer the event for performance
	h.bufferInputEvent(event)

	h.updateKeyboardStats("releases", 1)
}

// HandleUnicodeKey handles unicode key events
func (h *KeyboardHandler) HandleUnicodeKey(r rune) {
	h.mu.Lock()
	defer h.mu.Unlock()

	if h.client == nil {
		return
	}

	// Handle IME input if enabled
	if h.imeEnabled && h.shouldUseIME(r) {
		h.handleIMEInput(r)
		return
	}

	// Convert unicode to virtual key code
	keyCode := h.unicodeToKeyCode(r)
	if keyCode == 0 {
		return
	}

	// Determine if shift is needed for uppercase
	modifiers := t128.ModifierKey{}
	if unicode.IsUpper(r) {
		modifiers.Shift = true
	}

	// Send the key press and release
	h.sendKeyPressWithModifiers(keyCode, modifiers)

	h.updateKeyboardStats("unicode_input", 1)
	h.updateKeyboardStats("unicode_"+string(r), 1)
}

// HandleKeyCombo handles key combinations (Ctrl+A, Alt+Tab, etc.)
func (h *KeyboardHandler) HandleKeyCombo(keyCode uint8, ctrl, alt, shift, meta bool) {
	h.mu.Lock()
	defer h.mu.Unlock()

	if h.client == nil {
		return
	}

	modifiers := t128.ModifierKey{
		Control: ctrl,
		Alt:     alt,
		Shift:   shift,
		Meta:    meta,
	}

	h.sendKeyPressWithModifiers(keyCode, modifiers)
	h.updateKeyboardStats("combos", 1)
}

// HandleFunctionKey handles function keys (F1-F24)
func (h *KeyboardHandler) HandleFunctionKey(functionNumber int, modifiers t128.ModifierKey) {
	if functionNumber < 1 || functionNumber > 24 {
		return
	}

	keyCode := uint8(0x6F + functionNumber) // F1 = 0x70, F2 = 0x71, etc.
	h.HandleKeyCombo(keyCode, modifiers.Control, modifiers.Alt, modifiers.Shift, modifiers.Meta)
	h.updateKeyboardStats("function_keys", 1)
}

// HandleArrowKey handles arrow key input
func (h *KeyboardHandler) HandleArrowKey(direction string, modifiers t128.ModifierKey) {
	var keyCode uint8

	switch direction {
	case "UP":
		keyCode = t128.VK_UP
	case "DOWN":
		keyCode = t128.VK_DOWN
	case "LEFT":
		keyCode = t128.VK_LEFT
	case "RIGHT":
		keyCode = t128.VK_RIGHT
	default:
		return
	}

	h.HandleKeyCombo(keyCode, modifiers.Control, modifiers.Alt, modifiers.Shift, modifiers.Meta)
	h.updateKeyboardStats("arrow_keys", 1)
}

// HandleNavigationKey handles navigation keys
func (h *KeyboardHandler) HandleNavigationKey(keyName string, modifiers t128.ModifierKey) {
	var keyCode uint8

	switch keyName {
	case "HOME":
		keyCode = t128.VK_HOME
	case "END":
		keyCode = t128.VK_END
	case "PAGEUP":
		keyCode = t128.VK_PRIOR
	case "PAGEDOWN":
		keyCode = t128.VK_NEXT
	case "INSERT":
		keyCode = t128.VK_INSERT
	case "DELETE":
		keyCode = t128.VK_DELETE
	default:
		return
	}

	h.HandleKeyCombo(keyCode, modifiers.Control, modifiers.Alt, modifiers.Shift, modifiers.Meta)
	h.updateKeyboardStats("navigation_keys", 1)
}

// updateModifierKeys updates the modifier key state
func (h *KeyboardHandler) updateModifierKeys(keyCode uint8, pressed bool) {
	switch keyCode {
	case t128.VK_SHIFT:
		h.modifierKeys.Shift = pressed
	case t128.VK_CONTROL:
		h.modifierKeys.Control = pressed
	case t128.VK_MENU: // Alt
		h.modifierKeys.Alt = pressed
	case t128.VK_LWIN, t128.VK_RWIN:
		h.modifierKeys.Meta = pressed
	case t128.VK_CAPITAL:
		if pressed {
			h.capsLockState = !h.capsLockState
		}
	case t128.VK_NUMLOCK:
		if pressed {
			h.numLockState = !h.numLockState
		}
	case t128.VK_SCROLL:
		if pressed {
			h.scrollLockState = !h.scrollLockState
		}
	}
}

// handleSpecialKeyPress handles special key press events
func (h *KeyboardHandler) handleSpecialKeyPress(keyCode uint8) {
	switch keyCode {
	case t128.VK_CAPITAL:
		h.updateKeyboardStats("caps_lock_toggles", 1)
	case t128.VK_NUMLOCK:
		h.updateKeyboardStats("num_lock_toggles", 1)
	case t128.VK_SCROLL:
		h.updateKeyboardStats("scroll_lock_toggles", 1)
	}
}

// handleSpecialKeyRelease handles special key release events
func (h *KeyboardHandler) handleSpecialKeyRelease(keyCode uint8) {
	// Handle any special key release logic here
}

// unicodeToKeyCode converts unicode character to virtual key code
func (h *KeyboardHandler) unicodeToKeyCode(unicode rune) uint8 {
	// Check the key map first
	if keyCode, exists := t128.KeyMap[unicode]; exists {
		return keyCode
	}

	// Handle uppercase letters
	if unicode >= 'A' && unicode <= 'Z' {
		return uint8(unicode)
	}

	// Handle lowercase letters
	if unicode >= 'a' && unicode <= 'z' {
		return uint8(unicode - 32) // Convert to uppercase
	}

	// Handle numbers
	if unicode >= '0' && unicode <= '9' {
		return uint8(unicode)
	}

	// Handle common symbols
	switch unicode {
	case '!':
		return '1'
	case '@':
		return '2'
	case '#':
		return '3'
	case '$':
		return '4'
	case '%':
		return '5'
	case '^':
		return '6'
	case '&':
		return '7'
	case '*':
		return '8'
	case '(':
		return '9'
	case ')':
		return '0'
	case '-':
		return t128.VK_OEM_MINUS
	case '=':
		return t128.VK_OEM_PLUS
	case '[':
		return t128.VK_OEM_4
	case ']':
		return t128.VK_OEM_6
	case '\\':
		return t128.VK_OEM_5
	case ';':
		return t128.VK_OEM_1
	case '\'':
		return t128.VK_OEM_7
	case ',':
		return t128.VK_OEM_COMMA
	case '.':
		return t128.VK_OEM_PERIOD
	case '/':
		return t128.VK_OEM_2
	case '`':
		return t128.VK_OEM_3
	case '~':
		return t128.VK_OEM_3
	case '{':
		return t128.VK_OEM_4
	case '}':
		return t128.VK_OEM_6
	case '|':
		return t128.VK_OEM_5
	case ':':
		return t128.VK_OEM_1
	case '"':
		return t128.VK_OEM_7
	case '<':
		return t128.VK_OEM_COMMA
	case '>':
		return t128.VK_OEM_PERIOD
	case '?':
		return t128.VK_OEM_2
	}

	return 0
}

// shouldUseIME determines if IME should be used for this unicode character
func (h *KeyboardHandler) shouldUseIME(unicode rune) bool {
	// Use IME for non-ASCII characters when IME is enabled
	return h.imeEnabled && unicode > 127
}

// handleIMEInput handles IME input for complex characters
func (h *KeyboardHandler) handleIMEInput(unicode rune) {
	// This is a simplified IME implementation
	// In a full implementation, you would integrate with the system IME

	// For now, we'll try to map to available keys or use Unicode input
	keyCode := h.unicodeToKeyCode(unicode)
	if keyCode != 0 {
		h.sendKeyPressWithModifiers(keyCode, t128.ModifierKey{})
	} else {
		// Fallback: send as Unicode input
		h.sendUnicodeInput(unicode)
	}
}

// sendUnicodeInput sends Unicode input directly
func (h *KeyboardHandler) sendUnicodeInput(unicode rune) {
	// This would need to be implemented based on RDP Unicode input support
	// For now, we'll log the attempt
	fmt.Printf("Unicode input not fully implemented: %c (0x%04x)\n", unicode, unicode)
}

// sendKeyPressWithModifiers sends a key press with modifiers
func (h *KeyboardHandler) sendKeyPressWithModifiers(keyCode uint8, modifiers t128.ModifierKey) {
	// Send modifier keys first
	if modifiers.Shift {
		h.bufferInputEvent(t128.NewFastPathKeyboardEvent(t128.VK_SHIFT, true))
	}
	if modifiers.Control {
		h.bufferInputEvent(t128.NewFastPathKeyboardEvent(t128.VK_CONTROL, true))
	}
	if modifiers.Alt {
		h.bufferInputEvent(t128.NewFastPathKeyboardEvent(t128.VK_MENU, true))
	}
	if modifiers.Meta {
		h.bufferInputEvent(t128.NewFastPathKeyboardEvent(t128.VK_LWIN, true))
	}

	// Send the actual key
	h.bufferInputEvent(t128.NewFastPathKeyboardEvent(keyCode, true))
	h.bufferInputEvent(t128.NewFastPathKeyboardEvent(keyCode, false))

	// Release modifier keys
	if modifiers.Shift {
		h.bufferInputEvent(t128.NewFastPathKeyboardEvent(t128.VK_SHIFT, false))
	}
	if modifiers.Control {
		h.bufferInputEvent(t128.NewFastPathKeyboardEvent(t128.VK_CONTROL, false))
	}
	if modifiers.Alt {
		h.bufferInputEvent(t128.NewFastPathKeyboardEvent(t128.VK_MENU, false))
	}
	if modifiers.Meta {
		h.bufferInputEvent(t128.NewFastPathKeyboardEvent(t128.VK_LWIN, false))
	}
}

// bufferInputEvent adds an input event to the buffer
func (h *KeyboardHandler) bufferInputEvent(event *t128.TsFpKeyboardEvent) {
	h.bufferMutex.Lock()
	defer h.bufferMutex.Unlock()

	h.inputBuffer = append(h.inputBuffer, *event)

	// Flush if buffer is full
	if len(h.inputBuffer) >= h.bufferSize {
		h.flushInputBuffer()
	}
}

// flushInputBuffer sends all buffered input events
func (h *KeyboardHandler) flushInputBuffer() {
	h.bufferMutex.Lock()
	defer h.bufferMutex.Unlock()

	if len(h.inputBuffer) == 0 {
		return
	}

	// Send all buffered events
	for _, event := range h.inputBuffer {
		h.sendKeyboardEvent(&event)
	}

	// Clear buffer
	h.inputBuffer = h.inputBuffer[:0]

	// Reset flush timer
	h.flushTimer.Reset(h.flushInterval)
}

// startKeyRepeat starts key repeat for the specified key
func (h *KeyboardHandler) startKeyRepeat(keyCode uint8) {
	if h.isKeyRepeating {
		return
	}

	h.isKeyRepeating = true
	h.repeatKeyCode = keyCode

	h.repeatTimer = time.AfterFunc(h.keyRepeatRate, func() {
		h.mu.Lock()
		defer h.mu.Unlock()

		if h.isKeyRepeating && h.keyStates[keyCode] {
			event := t128.NewFastPathKeyboardEvent(keyCode, true)
			h.bufferInputEvent(event)
			h.updateKeyboardStats("repeats", 1)

			// Schedule next repeat
			h.repeatTimer.Reset(h.keyRepeatRate)
		} else {
			h.stopKeyRepeat()
		}
	})
}

// stopKeyRepeat stops key repeat
func (h *KeyboardHandler) stopKeyRepeat() {
	h.isKeyRepeating = false
	h.repeatKeyCode = 0
	if h.repeatTimer != nil {
		h.repeatTimer.Stop()
	}
}

// initializeStats initializes keyboard statistics
func (h *KeyboardHandler) initializeStats() {
	h.keyboardStats["total_presses"] = 0
	h.keyboardStats["total_releases"] = 0
	h.keyboardStats["total_combos"] = 0
	h.keyboardStats["total_function_keys"] = 0
	h.keyboardStats["total_arrow_keys"] = 0
	h.keyboardStats["total_navigation_keys"] = 0
	h.keyboardStats["total_unicode_input"] = 0
	h.keyboardStats["total_repeats"] = 0
	h.keyboardStats["caps_lock_toggles"] = 0
	h.keyboardStats["num_lock_toggles"] = 0
	h.keyboardStats["scroll_lock_toggles"] = 0
	h.keyboardStats["start_time"] = time.Now()
}

// updateKeyboardStats updates keyboard statistics
func (h *KeyboardHandler) updateKeyboardStats(key string, value int) {
	if current, exists := h.keyboardStats[key]; exists {
		if intValue, ok := current.(int); ok {
			h.keyboardStats[key] = intValue + value
		}
	} else {
		h.keyboardStats[key] = value
	}
}

// GetKeyboardStats returns keyboard statistics
func (h *KeyboardHandler) GetKeyboardStats() map[string]interface{} {
	h.mu.RLock()
	defer h.mu.RUnlock()

	stats := make(map[string]interface{})
	for k, v := range h.keyboardStats {
		stats[k] = v
	}

	// Add current state
	stats["caps_lock"] = h.capsLockState
	stats["num_lock"] = h.numLockState
	stats["scroll_lock"] = h.scrollLockState
	stats["ime_enabled"] = h.imeEnabled
	stats["current_layout"] = h.currentLayout
	stats["pressed_keys_count"] = len(h.keyStates)

	return stats
}

// IsKeyPressed checks if a key is currently pressed
func (h *KeyboardHandler) IsKeyPressed(keyCode uint8) bool {
	h.mu.RLock()
	defer h.mu.RUnlock()
	return h.keyStates[keyCode]
}

// GetModifierKeys returns the current modifier key state
func (h *KeyboardHandler) GetModifierKeys() t128.ModifierKey {
	h.mu.RLock()
	defer h.mu.RUnlock()
	return h.modifierKeys
}

// GetKeyboardState returns the current keyboard state
func (h *KeyboardHandler) GetKeyboardState() map[uint8]bool {
	h.mu.RLock()
	defer h.mu.RUnlock()

	state := make(map[uint8]bool)
	for k, v := range h.keyStates {
		state[k] = v
	}
	return state
}

// SetKeyboardLayout sets the keyboard layout
func (h *KeyboardHandler) SetKeyboardLayout(layout string) {
	h.mu.Lock()
	defer h.mu.Unlock()
	h.currentLayout = layout
	h.updateKeyboardStats("layout_changes", 1)
}

// SetIMEEnabled enables or disables IME support
func (h *KeyboardHandler) SetIMEEnabled(enabled bool) {
	h.mu.Lock()
	defer h.mu.Unlock()
	h.imeEnabled = enabled
	h.updateKeyboardStats("ime_toggles", 1)
}

// SetKeyRepeatSettings sets key repeat delay and rate
func (h *KeyboardHandler) SetKeyRepeatSettings(delay, rate time.Duration) {
	h.mu.Lock()
	defer h.mu.Unlock()
	h.keyRepeatDelay = delay
	h.keyRepeatRate = rate
}

// sendKeyboardEvent sends a keyboard event to the RDP client
func (h *KeyboardHandler) sendKeyboardEvent(event *t128.TsFpKeyboardEvent) {
	if h.client == nil {
		fmt.Println("RDP client is not initialized; cannot send keyboard event")
		return
	}

	// Determine if this is a key press or release
	isKeyDown := (event.EventHeader & 0x01) != 0

	// Send the keyboard event using the client's SendKeyEvent method
	err := h.client.SendKeyEvent(event.KeyCode, isKeyDown, h.modifierKeys)

	if err != nil {
		fmt.Printf("Failed to send keyboard event: %v\n", err)
		h.updateKeyboardStats("errors", 1)
	} else {
		fmt.Printf("Keyboard event sent: flags=0x%02x, keyCode=0x%02x, down=%v\n",
			event.EventHeader, event.KeyCode, isKeyDown)
	}
}

// SetClient sets the RDP client reference
func (h *KeyboardHandler) SetClient(client *gordp.Client) {
	h.mu.Lock()
	defer h.mu.Unlock()
	h.client = client
}

// GetClient returns the RDP client reference
func (h *KeyboardHandler) GetClient() *gordp.Client {
	h.mu.RLock()
	defer h.mu.RUnlock()
	return h.client
}

// Close cleans up the keyboard handler
func (h *KeyboardHandler) Close() {
	h.mu.Lock()
	defer h.mu.Unlock()

	// Stop timers
	if h.repeatTimer != nil {
		h.repeatTimer.Stop()
	}
	if h.flushTimer != nil {
		h.flushTimer.Stop()
	}

	// Flush any remaining input
	h.flushInputBuffer()

	// Clear state
	h.keyStates = make(map[uint8]bool)
	h.inputBuffer = h.inputBuffer[:0]
}

// KeyCode represents virtual key codes
type KeyCode uint8

const (
	// Common key codes
	KeyEscape    KeyCode = t128.VK_ESCAPE
	KeyEnter     KeyCode = t128.VK_ENTER
	KeyTab       KeyCode = t128.VK_TAB
	KeySpace     KeyCode = t128.VK_SPACE
	KeyBackspace KeyCode = t128.VK_BACK
	KeyDelete    KeyCode = t128.VK_DELETE
	KeyInsert    KeyCode = t128.VK_INSERT
	KeyHome      KeyCode = t128.VK_HOME
	KeyEnd       KeyCode = t128.VK_END
	KeyPageUp    KeyCode = t128.VK_PRIOR
	KeyPageDown  KeyCode = t128.VK_NEXT

	// Arrow keys
	KeyLeft  KeyCode = t128.VK_LEFT
	KeyUp    KeyCode = t128.VK_UP
	KeyRight KeyCode = t128.VK_RIGHT
	KeyDown  KeyCode = t128.VK_DOWN

	// Modifier keys
	KeyShift   KeyCode = t128.VK_SHIFT
	KeyControl KeyCode = t128.VK_CONTROL
	KeyAlt     KeyCode = t128.VK_MENU
	KeyMeta    KeyCode = t128.VK_LWIN

	// Lock keys
	KeyCapsLock   KeyCode = t128.VK_CAPITAL
	KeyNumLock    KeyCode = t128.VK_NUMLOCK
	KeyScrollLock KeyCode = t128.VK_SCROLL
)
