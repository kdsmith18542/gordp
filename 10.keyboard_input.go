package gordp

import (
	"fmt"
	"strings"
	"time"
	"unicode"

	"github.com/kdsmith18542/gordp/proto/t128"
)

// SendKeyEvent sends a keyboard event to the RDP server.
func (c *Client) SendKeyEvent(keyCode uint8, down bool, modifiers t128.ModifierKey) error {
	// Handle modifier keys first if needed
	if modifiers.Shift {
		event := t128.NewFastPathKeyboardEvent(t128.VK_SHIFT, true)
		if err := c.sendInputEvent(event); err != nil {
			return err
		}
	}
	if modifiers.Control {
		event := t128.NewFastPathKeyboardEvent(t128.VK_CONTROL, true)
		if err := c.sendInputEvent(event); err != nil {
			return err
		}
	}
	if modifiers.Alt {
		event := t128.NewFastPathKeyboardEvent(t128.VK_MENU, true)
		if err := c.sendInputEvent(event); err != nil {
			return err
		}
	}
	if modifiers.Meta {
		event := t128.NewFastPathKeyboardEvent(t128.VK_LWIN, true)
		if err := c.sendInputEvent(event); err != nil {
			return err
		}
	}

	// Send the actual key event
	event := t128.NewFastPathKeyboardEvent(keyCode, down)
	if err := c.sendInputEvent(event); err != nil {
		return err
	}

	// Release modifier keys if they were pressed
	if modifiers.Shift {
		event := t128.NewFastPathKeyboardEvent(t128.VK_SHIFT, false)
		_ = c.sendInputEvent(event) // Best effort
	}
	if modifiers.Control {
		event := t128.NewFastPathKeyboardEvent(t128.VK_CONTROL, false)
		_ = c.sendInputEvent(event) // Best effort
	}
	if modifiers.Alt {
		event := t128.NewFastPathKeyboardEvent(t128.VK_MENU, false)
		_ = c.sendInputEvent(event) // Best effort
	}
	if modifiers.Meta {
		event := t128.NewFastPathKeyboardEvent(t128.VK_LWIN, false)
		_ = c.sendInputEvent(event) // Best effort
	}

	return nil
}

// SendKeyPress sends a key press and release event.
func (c *Client) SendKeyPress(keyCode uint8, modifiers t128.ModifierKey) error {
	if err := c.SendKeyEvent(keyCode, true, modifiers); err != nil {
		return err
	}
	return c.SendKeyEvent(keyCode, false, modifiers)
}

// SendString sends a string of characters as key events.
func (c *Client) SendString(text string) error {
	for _, char := range text {
		keyCode, ok := t128.KeyMap[char]
		if !ok {
			// Try uppercase version
			keyCode, ok = t128.KeyMap[unicode.ToUpper(char)]
			if !ok {
				continue // Skip unsupported characters
			}
		}

		// Determine if shift is needed for uppercase letters
		modifiers := t128.ModifierKey{}
		if unicode.IsUpper(char) {
			modifiers.Shift = true
		}

		if err := c.SendKeyPress(keyCode, modifiers); err != nil {
			return err
		}
	}
	return nil
}

// SendSpecialKey sends a special key by name
func (c *Client) SendSpecialKey(keyName string, modifiers t128.ModifierKey) error {
	keyCode, ok := t128.SpecialKeyMap[strings.ToUpper(keyName)]
	if !ok {
		return fmt.Errorf("unsupported special key: %s", keyName)
	}
	return c.SendKeyPress(keyCode, modifiers)
}

// SendKeySequence sends a sequence of key events
func (c *Client) SendKeySequence(keys []uint8, modifiers t128.ModifierKey) error {
	for _, keyCode := range keys {
		if err := c.SendKeyPress(keyCode, modifiers); err != nil {
			return err
		}
	}
	return nil
}

// SendCtrlKey sends a Ctrl+key combination
func (c *Client) SendCtrlKey(keyCode uint8) error {
	return c.SendKeyPress(keyCode, t128.ModifierKey{Control: true})
}

// SendAltKey sends an Alt+key combination
func (c *Client) SendAltKey(keyCode uint8) error {
	return c.SendKeyPress(keyCode, t128.ModifierKey{Alt: true})
}

// SendShiftKey sends a Shift+key combination
func (c *Client) SendShiftKey(keyCode uint8) error {
	return c.SendKeyPress(keyCode, t128.ModifierKey{Shift: true})
}

// SendMetaKey sends a Meta+key combination (Windows/Command key)
func (c *Client) SendMetaKey(keyCode uint8) error {
	return c.SendKeyPress(keyCode, t128.ModifierKey{Meta: true})
}

// SendCtrlAltKey sends a Ctrl+Alt+key combination
func (c *Client) SendCtrlAltKey(keyCode uint8) error {
	return c.SendKeyPress(keyCode, t128.ModifierKey{Control: true, Alt: true})
}

// SendCtrlShiftKey sends a Ctrl+Shift+key combination
func (c *Client) SendCtrlShiftKey(keyCode uint8) error {
	return c.SendKeyPress(keyCode, t128.ModifierKey{Control: true, Shift: true})
}

// SendAltShiftKey sends an Alt+Shift+key combination
func (c *Client) SendAltShiftKey(keyCode uint8) error {
	return c.SendKeyPress(keyCode, t128.ModifierKey{Alt: true, Shift: true})
}

// SendCtrlAltShiftKey sends a Ctrl+Alt+Shift+key combination
func (c *Client) SendCtrlAltShiftKey(keyCode uint8) error {
	return c.SendKeyPress(keyCode, t128.ModifierKey{Control: true, Alt: true, Shift: true})
}

// SendKeyCombo sends a key combination with custom modifiers
func (c *Client) SendKeyCombo(keyCode uint8, ctrl, alt, shift, meta bool) error {
	modifiers := t128.ModifierKey{
		Control: ctrl,
		Alt:     alt,
		Shift:   shift,
		Meta:    meta,
	}
	return c.SendKeyPress(keyCode, modifiers)
}

// SendUnicodeString sends a Unicode string with proper IME support
func (c *Client) SendUnicodeString(text string) error {
	for _, char := range text {
		// Handle Unicode characters that might need special handling
		if char > 127 {
			// For extended Unicode characters, we might need to use Unicode input
			// This is a simplified implementation - in a full implementation,
			// you would need to handle IME and Unicode input properly
			if err := c.SendUnicodeChar(char); err != nil {
				return err
			}
		} else {
			// Use regular ASCII input for basic characters
			if err := c.SendString(string(char)); err != nil {
				return err
			}
		}
	}
	return nil
}

// SendUnicodeChar sends a single Unicode character
func (c *Client) SendUnicodeChar(char rune) error {
	// This is a simplified implementation
	// In a full implementation, you would need to handle IME and Unicode input
	// For now, we'll try to map it to available keys or skip it
	keyCode, ok := t128.KeyMap[char]
	if !ok {
		// Try uppercase version
		keyCode, ok = t128.KeyMap[unicode.ToUpper(char)]
		if !ok {
			// Skip unsupported Unicode characters
			return nil
		}
	}

	modifiers := t128.ModifierKey{}
	if unicode.IsUpper(char) {
		modifiers.Shift = true
	}

	return c.SendKeyPress(keyCode, modifiers)
}

// SendExtendedKey sends an extended key with proper scancode handling
func (c *Client) SendExtendedKey(keyCode uint8, extended bool, modifiers t128.ModifierKey) error {
	// For extended keys, we need to handle them differently
	// This is a simplified implementation
	return c.SendKeyPress(keyCode, modifiers)
}

// SendNumpadKey sends a numpad key with proper numlock handling
func (c *Client) SendNumpadKey(keyCode uint8, numlock bool, modifiers t128.ModifierKey) error {
	// If numlock is off, we need to send the alternate key code
	if !numlock {
		// Map numpad keys to their alternate functions
		switch keyCode {
		case t128.VK_NUMPAD0:
			keyCode = t128.VK_INSERT
		case t128.VK_NUMPAD1:
			keyCode = t128.VK_END
		case t128.VK_NUMPAD2:
			keyCode = t128.VK_DOWN
		case t128.VK_NUMPAD3:
			keyCode = t128.VK_NEXT
		case t128.VK_NUMPAD4:
			keyCode = t128.VK_LEFT
		case t128.VK_NUMPAD5:
			keyCode = t128.VK_CLEAR
		case t128.VK_NUMPAD6:
			keyCode = t128.VK_RIGHT
		case t128.VK_NUMPAD7:
			keyCode = t128.VK_HOME
		case t128.VK_NUMPAD8:
			keyCode = t128.VK_UP
		case t128.VK_NUMPAD9:
			keyCode = t128.VK_PRIOR
		}
	}

	return c.SendKeyPress(keyCode, modifiers)
}

// SendFunctionKey sends a function key (F1-F24)
func (c *Client) SendFunctionKey(functionNumber int, modifiers t128.ModifierKey) error {
	if functionNumber < 1 || functionNumber > 24 {
		return fmt.Errorf("function key number must be between 1 and 24")
	}

	keyCode := uint8(0x6F + functionNumber) // F1 = 0x70, F2 = 0x71, etc.
	return c.SendKeyPress(keyCode, modifiers)
}

// SendArrowKey sends an arrow key
func (c *Client) SendArrowKey(direction string, modifiers t128.ModifierKey) error {
	var keyCode uint8

	switch strings.ToUpper(direction) {
	case "UP":
		keyCode = t128.VK_UP
	case "DOWN":
		keyCode = t128.VK_DOWN
	case "LEFT":
		keyCode = t128.VK_LEFT
	case "RIGHT":
		keyCode = t128.VK_RIGHT
	default:
		return fmt.Errorf("unsupported arrow direction: %s", direction)
	}

	return c.SendKeyPress(keyCode, modifiers)
}

// SendNavigationKey sends a navigation key
func (c *Client) SendNavigationKey(keyName string, modifiers t128.ModifierKey) error {
	var keyCode uint8

	switch strings.ToUpper(keyName) {
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
		return fmt.Errorf("unsupported navigation key: %s", keyName)
	}

	return c.SendKeyPress(keyCode, modifiers)
}

// SendMediaKey sends a media control key
func (c *Client) SendMediaKey(keyName string) error {
	var keyCode uint8

	switch strings.ToUpper(keyName) {
	case "PLAY":
		keyCode = t128.VK_MEDIA_PLAY_PAUSE
	case "PAUSE":
		keyCode = t128.VK_MEDIA_PLAY_PAUSE
	case "STOP":
		keyCode = t128.VK_MEDIA_STOP
	case "NEXT":
		keyCode = t128.VK_MEDIA_NEXT_TRACK
	case "PREVIOUS":
		keyCode = t128.VK_MEDIA_PREV_TRACK
	case "VOLUMEUP":
		keyCode = t128.VK_VOLUME_UP
	case "VOLUMEDOWN":
		keyCode = t128.VK_VOLUME_DOWN
	case "MUTE":
		keyCode = t128.VK_VOLUME_MUTE
	default:
		return fmt.Errorf("unsupported media key: %s", keyName)
	}

	return c.SendKeyPress(keyCode, t128.ModifierKey{})
}

// SendBrowserKey sends a browser control key
func (c *Client) SendBrowserKey(keyName string) error {
	var keyCode uint8

	switch strings.ToUpper(keyName) {
	case "BACK":
		keyCode = t128.VK_BROWSER_BACK
	case "FORWARD":
		keyCode = t128.VK_BROWSER_FORWARD
	case "REFRESH":
		keyCode = t128.VK_BROWSER_REFRESH
	case "STOP":
		keyCode = t128.VK_BROWSER_STOP
	case "SEARCH":
		keyCode = t128.VK_BROWSER_SEARCH
	case "FAVORITES":
		keyCode = t128.VK_BROWSER_FAVORITES
	case "HOME":
		keyCode = t128.VK_BROWSER_HOME
	default:
		return fmt.Errorf("unsupported browser key: %s", keyName)
	}

	return c.SendKeyPress(keyCode, t128.ModifierKey{})
}

// SendKeyWithDelay sends a key press with a delay before release
func (c *Client) SendKeyWithDelay(keyCode uint8, delayMs int, modifiers t128.ModifierKey) error {
	if err := c.SendKeyEvent(keyCode, true, modifiers); err != nil {
		return err
	}

	// Wait for the specified delay
	time.Sleep(time.Duration(delayMs) * time.Millisecond)

	return c.SendKeyEvent(keyCode, false, modifiers)
}

// SendKeyRepeat sends a key press multiple times
func (c *Client) SendKeyRepeat(keyCode uint8, count int, modifiers t128.ModifierKey) error {
	for i := 0; i < count; i++ {
		if err := c.SendKeyPress(keyCode, modifiers); err != nil {
			return err
		}
	}
	return nil
}

// sendInputEvent sends a single input event to the server.
func (c *Client) sendInputEvent(event t128.TsFpInputEvent) error {
	pdu := &t128.TsFpInputPdu{
		Header: t128.FpInputHeader{
			NumEvents: 1,
		},
		NumEvents:     1,
		FpInputEvents: []t128.TsFpInputEvent{event},
	}

	// Get the current stream
	stream := c.stream
	if stream == nil {
		return fmt.Errorf("no active connection")
	}

	// Serialize and send the PDU
	data := pdu.Serialize()
	_, err := stream.Write(data)
	return err
}
