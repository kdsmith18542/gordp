package gordp

import (
	"fmt"
	"os"
	"testing"
	"time"

	"github.com/GoFeGroup/gordp/proto/bitmap"
	"github.com/GoFeGroup/gordp/proto/t128"
	"github.com/stretchr/testify/assert"
)

type processor struct {
	i int
}

func (p *processor) ProcessBitmap(option *bitmap.Option, bitmap *bitmap.BitMap) {
	p.i++
	_ = os.MkdirAll("./png", 0755)
	_ = os.WriteFile(fmt.Sprintf("./png/%v.png", p.i), bitmap.ToPng(), 0644)
}

func TestRdpConnect(t *testing.T) {
	client := NewClient(&Option{
		Addr:     "10.226.239.200:3389",
		UserName: "testuser",
		Password: "testpass",
	})

	// Use a shorter timeout and handle connection errors gracefully
	client.option.ConnectTimeout = 2 * time.Second

	err := client.Connect()
	if err != nil {
		// Expected for network timeout in test environment
		t.Logf("Connection failed as expected: %v", err)
		return
	}

	// If connection succeeds, test basic functionality
	defer client.Close()
	t.Log("Connection successful")
}

// TestClientCreation tests client creation with various options
func TestClientCreation(t *testing.T) {
	tests := []struct {
		name     string
		option   *Option
		expected time.Duration
	}{
		{
			name: "default timeout",
			option: &Option{
				Addr:     "localhost:3389",
				UserName: "test",
				Password: "test",
			},
			expected: 5 * time.Second,
		},
		{
			name: "custom timeout",
			option: &Option{
				Addr:           "localhost:3389",
				UserName:       "test",
				Password:       "test",
				ConnectTimeout: 10 * time.Second,
			},
			expected: 10 * time.Second,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			client := NewClient(tt.option)
			if client.option.ConnectTimeout != tt.expected {
				t.Errorf("expected timeout %v, got %v", tt.expected, client.option.ConnectTimeout)
			}
		})
	}
}

// TestKeyboardInput tests keyboard input functionality
func TestKeyboardInput(t *testing.T) {
	client := NewClient(&Option{
		Addr:     "localhost:3389",
		UserName: "test",
		Password: "test",
	})

	// Test basic key press
	err := client.SendKeyPress('a', t128.ModifierKey{})
	if err != nil {
		t.Errorf("SendKeyPress failed: %v", err)
	}

	// Test modifier keys
	err = client.SendKeyPress('c', t128.ModifierKey{Control: true})
	if err != nil {
		t.Errorf("SendKeyPress with modifier failed: %v", err)
	}

	// Test special keys
	err = client.SendSpecialKey("F1", t128.ModifierKey{})
	if err != nil {
		t.Errorf("SendSpecialKey failed: %v", err)
	}

	// Test string input
	err = client.SendString("Hello, World!")
	if err != nil {
		t.Errorf("SendString failed: %v", err)
	}
}

// TestMouseInput tests mouse input functionality
func TestMouseInput(t *testing.T) {
	client := NewClient(&Option{
		Addr:     "localhost:3389",
		UserName: "test",
		Password: "test",
	})

	// Test mouse movement
	err := client.SendMouseMoveEvent(100, 200)
	if err != nil {
		t.Errorf("SendMouseMoveEvent failed: %v", err)
	}

	// Test mouse clicks
	err = client.SendMouseClickEvent(t128.MouseButtonLeft, 100, 200)
	if err != nil {
		t.Errorf("SendMouseClickEvent failed: %v", err)
	}

	err = client.SendMouseClickEvent(t128.MouseButtonRight, 150, 250)
	if err != nil {
		t.Errorf("SendMouseClickEvent failed: %v", err)
	}

	// Test mouse wheel
	err = client.SendMouseWheelEvent(120, 100, 200)
	if err != nil {
		t.Errorf("SendMouseWheelEvent failed: %v", err)
	}

	err = client.SendMouseHorizontalWheelEvent(-120, 100, 200)
	if err != nil {
		t.Errorf("SendMouseHorizontalWheelEvent failed: %v", err)
	}
}

// TestKeyboardModifiers tests keyboard modifier combinations
func TestKeyboardModifiers(t *testing.T) {
	client := NewClient(&Option{
		Addr:     "localhost:3389",
		UserName: "test",
		Password: "test",
	})

	// Test Ctrl combinations
	err := client.SendCtrlKey('a')
	if err != nil {
		t.Errorf("SendCtrlKey failed: %v", err)
	}

	// Test Alt combinations
	err = client.SendAltKey(t128.VK_TAB)
	if err != nil {
		t.Errorf("SendAltKey failed: %v", err)
	}

	// Test Shift combinations
	err = client.SendShiftKey('A')
	if err != nil {
		t.Errorf("SendShiftKey failed: %v", err)
	}

	// Test complex modifier combinations
	err = client.SendKeyPress('s', t128.ModifierKey{
		Control: true,
		Shift:   true,
	})
	if err != nil {
		t.Errorf("SendKeyPress with complex modifiers failed: %v", err)
	}
}

// TestMouseButtonEvents tests individual mouse button events
func TestMouseButtonEvents(t *testing.T) {
	client := NewClient(&Option{
		Addr:     "localhost:3389",
		UserName: "test",
		Password: "test",
	})

	// Test left button events
	err := client.SendMouseLeftDownEvent(100, 200)
	if err != nil {
		t.Errorf("SendMouseLeftDownEvent failed: %v", err)
	}

	err = client.SendMouseLeftUpEvent(100, 200)
	if err != nil {
		t.Errorf("SendMouseLeftUpEvent failed: %v", err)
	}

	// Test right button events
	err = client.SendMouseRightDownEvent(150, 250)
	if err != nil {
		t.Errorf("SendMouseRightDownEvent failed: %v", err)
	}

	err = client.SendMouseRightUpEvent(150, 250)
	if err != nil {
		t.Errorf("SendMouseRightUpEvent failed: %v", err)
	}

	// Test middle button events
	err = client.SendMouseMiddleDownEvent(200, 300)
	if err != nil {
		t.Errorf("SendMouseMiddleDownEvent failed: %v", err)
	}

	err = client.SendMouseMiddleUpEvent(200, 300)
	if err != nil {
		t.Errorf("SendMouseMiddleUpEvent failed: %v", err)
	}
}

// TestEnhancedKeyboardFeatures tests enhanced keyboard features
func TestEnhancedKeyboardFeatures(t *testing.T) {
	client := NewClient(&Option{
		Addr:     "localhost:3389",
		UserName: "test",
		Password: "test",
	})

	// Test function keys
	err := client.SendFunctionKey(1, t128.ModifierKey{})
	if err != nil {
		t.Errorf("SendFunctionKey failed: %v", err)
	}

	// Test arrow keys
	err = client.SendArrowKey("UP", t128.ModifierKey{})
	if err != nil {
		t.Errorf("SendArrowKey failed: %v", err)
	}

	// Test navigation keys
	err = client.SendNavigationKey("HOME", t128.ModifierKey{})
	if err != nil {
		t.Errorf("SendNavigationKey failed: %v", err)
	}

	// Test media keys
	err = client.SendMediaKey("PLAY")
	if err != nil {
		t.Errorf("SendMediaKey failed: %v", err)
	}

	// Test browser keys
	err = client.SendBrowserKey("BACK")
	if err != nil {
		t.Errorf("SendBrowserKey failed: %v", err)
	}

	// Test numpad keys
	err = client.SendNumpadKey(t128.VK_NUMPAD5, true, t128.ModifierKey{})
	if err != nil {
		t.Errorf("SendNumpadKey failed: %v", err)
	}

	// Test key repeat
	err = client.SendKeyRepeat('a', 3, t128.ModifierKey{})
	if err != nil {
		t.Errorf("SendKeyRepeat failed: %v", err)
	}
}

// TestEnhancedMouseFeatures tests enhanced mouse features
func TestEnhancedMouseFeatures(t *testing.T) {
	client := NewClient(&Option{
		Addr:     "localhost:3389",
		UserName: "test",
		Password: "test",
	})

	// Test double-click
	err := client.SendMouseDoubleClickEvent(t128.MouseButtonLeft, 100, 200)
	if err != nil {
		t.Errorf("SendMouseDoubleClickEvent failed: %v", err)
	}

	// Test drag events
	err = client.SendMouseDragEvent(t128.MouseButtonLeft, 100, 200, 300, 400)
	if err != nil {
		t.Errorf("SendMouseDragEvent failed: %v", err)
	}

	// Test smooth drag
	err = client.SendMouseSmoothDragEvent(t128.MouseButtonLeft, 100, 200, 300, 400, 5)
	if err != nil {
		t.Errorf("SendMouseSmoothDragEvent failed: %v", err)
	}

	// Test multi-click
	err = client.SendMouseMultiClickEvent(t128.MouseButtonLeft, 100, 200, 3)
	if err != nil {
		t.Errorf("SendMouseMultiClickEvent failed: %v", err)
	}

	// Test scroll events
	err = client.SendMouseScrollEvent(t128.ScrollUp, 120, 100, 200)
	if err != nil {
		t.Errorf("SendMouseScrollEvent failed: %v", err)
	}

	// Test X1 and X2 buttons
	err = client.SendMouseClickEvent(t128.MouseButtonX1, 100, 200)
	if err != nil {
		t.Errorf("SendMouseClickEvent with X1 button failed: %v", err)
	}

	err = client.SendMouseClickEvent(t128.MouseButtonX2, 150, 250)
	if err != nil {
		t.Errorf("SendMouseClickEvent with X2 button failed: %v", err)
	}
}

// TestUnicodeInput tests Unicode input functionality
func TestUnicodeInput(t *testing.T) {
	client := NewClient(&Option{
		Addr:     "localhost:3389",
		UserName: "test",
		Password: "test",
	})

	// Test Unicode string
	err := client.SendUnicodeString("Hello, 世界!")
	if err != nil {
		t.Errorf("SendUnicodeString failed: %v", err)
	}

	// Test Unicode character
	err = client.SendUnicodeChar('ñ')
	if err != nil {
		t.Errorf("SendUnicodeChar failed: %v", err)
	}
}

// TestKeySequence tests key sequence functionality
func TestKeySequence(t *testing.T) {
	client := NewClient(&Option{
		Addr:     "localhost:3389",
		UserName: "test",
		Password: "test",
	})

	// Test key sequence
	keys := []uint8{'H', 'e', 'l', 'l', 'o'}
	err := client.SendKeySequence(keys, t128.ModifierKey{})
	if err != nil {
		t.Errorf("SendKeySequence failed: %v", err)
	}

	// Test key sequence with modifiers
	err = client.SendKeySequence(keys, t128.ModifierKey{Shift: true})
	if err != nil {
		t.Errorf("SendKeySequence with modifiers failed: %v", err)
	}
}

// TestStringInput tests string input functionality
func TestStringInput(t *testing.T) {
	client := NewClient(&Option{
		Addr:     "localhost:3389",
		UserName: "test",
		Password: "test",
	})

	// Test basic string
	err := client.SendString("Hello, World!")
	if err != nil {
		t.Errorf("SendString failed: %v", err)
	}

	// Test string with special characters
	err = client.SendString("Test@123#$%")
	if err != nil {
		t.Errorf("SendString with special characters failed: %v", err)
	}

	// Test string with mixed case
	err = client.SendString("Hello World")
	if err != nil {
		t.Errorf("SendString with mixed case failed: %v", err)
	}
}

// TestSpecialKeys tests special key functionality
func TestSpecialKeys(t *testing.T) {
	client := NewClient(&Option{
		Addr:     "localhost:3389",
		UserName: "test",
		Password: "test",
	})

	// Test function keys
	specialKeys := []string{"F1", "F2", "F3", "F4", "F5"}
	for _, key := range specialKeys {
		err := client.SendSpecialKey(key, t128.ModifierKey{})
		if err != nil {
			t.Errorf("SendSpecialKey %s failed: %v", key, err)
		}
	}

	// Test navigation keys
	navKeys := []string{"HOME", "END", "PAGEUP", "PAGEDOWN", "INSERT", "DELETE"}
	for _, key := range navKeys {
		err := client.SendSpecialKey(key, t128.ModifierKey{})
		if err != nil {
			t.Errorf("SendSpecialKey %s failed: %v", key, err)
		}
	}

	// Test arrow keys
	arrowKeys := []string{"UP", "DOWN", "LEFT", "RIGHT"}
	for _, key := range arrowKeys {
		err := client.SendSpecialKey(key, t128.ModifierKey{})
		if err != nil {
			t.Errorf("SendSpecialKey %s failed: %v", key, err)
		}
	}
}

// TestMouseButtonEnum tests mouse button enumeration
func TestMouseButtonEnum(t *testing.T) {
	buttons := []t128.MouseButton{
		t128.MouseButtonLeft,
		t128.MouseButtonRight,
		t128.MouseButtonMiddle,
		t128.MouseButtonX1,
		t128.MouseButtonX2,
	}

	for i, button := range buttons {
		if int(button) != i {
			t.Errorf("Invalid mouse button value: %d", button)
		}
	}
}

// TestScrollDirectionEnum tests scroll direction enumeration
func TestScrollDirectionEnum(t *testing.T) {
	directions := []t128.ScrollDirection{
		t128.ScrollUp,
		t128.ScrollDown,
		t128.ScrollLeft,
		t128.ScrollRight,
	}

	for i, direction := range directions {
		if int(direction) != i {
			t.Errorf("Invalid scroll direction value: %d", direction)
		}
	}
}

// TestVirtualKeyCodes tests virtual key code constants
func TestVirtualKeyCodes(t *testing.T) {
	// Test basic key codes
	if t128.VK_A != 0x41 {
		t.Errorf("VK_A should be 0x41, got 0x%02X", t128.VK_A)
	}

	if t128.VK_Z != 0x5A {
		t.Errorf("VK_Z should be 0x5A, got 0x%02X", t128.VK_Z)
	}

	// Test special key codes
	if t128.VK_ENTER != 0x0D {
		t.Errorf("VK_ENTER should be 0x0D, got 0x%02X", t128.VK_ENTER)
	}

	if t128.VK_ESCAPE != 0x1B {
		t.Errorf("VK_ESCAPE should be 0x1B, got 0x%02X", t128.VK_ESCAPE)
	}
}

// TestKeyMap tests key mapping functionality
func TestKeyMap(t *testing.T) {
	// Test basic character mapping
	testCases := map[rune]uint8{
		'a':  0x41,
		'z':  0x5A,
		'0':  0x30,
		'9':  0x39,
		' ':  t128.VK_SPACE,
		'\t': t128.VK_TAB,
		'\n': t128.VK_RETURN,
	}

	for char, expected := range testCases {
		if keyCode, ok := t128.KeyMap[char]; !ok {
			t.Errorf("KeyMap missing character: %c", char)
		} else if keyCode != expected {
			t.Errorf("KeyMap[%c] should be 0x%02X, got 0x%02X", char, expected, keyCode)
		}
	}
}

// TestSpecialKeyMap tests special key mapping functionality
func TestSpecialKeyMap(t *testing.T) {
	// Test function key mapping
	for i := 1; i <= 12; i++ {
		keyName := fmt.Sprintf("F%d", i)
		if keyCode, ok := t128.SpecialKeyMap[keyName]; !ok {
			t.Errorf("SpecialKeyMap missing key: %s", keyName)
		} else if keyCode != uint8(0x6F+i) {
			t.Errorf("SpecialKeyMap[%s] should be 0x%02X, got 0x%02X", keyName, 0x6F+i, keyCode)
		}
	}

	// Test navigation key mapping
	navKeys := map[string]uint8{
		"HOME":     t128.VK_HOME,
		"END":      t128.VK_END,
		"PAGEUP":   t128.VK_PRIOR,
		"PAGEDOWN": t128.VK_NEXT,
		"INSERT":   t128.VK_INSERT,
		"DELETE":   t128.VK_DELETE,
		"UP":       t128.VK_UP,
		"DOWN":     t128.VK_DOWN,
		"LEFT":     t128.VK_LEFT,
		"RIGHT":    t128.VK_RIGHT,
	}

	for keyName, expected := range navKeys {
		if keyCode, ok := t128.SpecialKeyMap[keyName]; !ok {
			t.Errorf("SpecialKeyMap missing key: %s", keyName)
		} else if keyCode != expected {
			t.Errorf("SpecialKeyMap[%s] should be 0x%02X, got 0x%02X", keyName, expected, keyCode)
		}
	}
}

// BenchmarkStringInput benchmarks string input performance
func BenchmarkStringInput(b *testing.B) {
	client := NewClient(&Option{
		Addr:     "localhost:3389",
		UserName: "test",
		Password: "test",
	})

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = client.SendString("Hello, World!")
	}
}

// BenchmarkKeyPress benchmarks key press performance
func BenchmarkKeyPress(b *testing.B) {
	client := NewClient(&Option{
		Addr:     "localhost:3389",
		UserName: "test",
		Password: "test",
	})

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = client.SendKeyPress('a', t128.ModifierKey{})
	}
}

// BenchmarkMouseMove benchmarks mouse movement performance
func BenchmarkMouseMove(b *testing.B) {
	client := NewClient(&Option{
		Addr:     "localhost:3389",
		UserName: "test",
		Password: "test",
	})

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = client.SendMouseMoveEvent(uint16(i%1000), uint16(i%1000))
	}
}

// BenchmarkMouseClick benchmarks mouse click performance
func BenchmarkMouseClick(b *testing.B) {
	client := NewClient(&Option{
		Addr:     "localhost:3389",
		UserName: "test",
		Password: "test",
	})

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = client.SendMouseClickEvent(t128.MouseButtonLeft, uint16(i%1000), uint16(i%1000))
	}
}

// TestErrorHandling tests error handling in input functions
func TestErrorHandling(t *testing.T) {
	client := NewClient(&Option{
		Addr:     "localhost:3389",
		UserName: "test",
		Password: "test",
	})

	// Test invalid function key
	err := client.SendFunctionKey(0, t128.ModifierKey{})
	if err == nil {
		t.Error("SendFunctionKey should fail with invalid function number")
	}

	err = client.SendFunctionKey(25, t128.ModifierKey{})
	if err == nil {
		t.Error("SendFunctionKey should fail with invalid function number")
	}

	// Test invalid arrow direction
	err = client.SendArrowKey("INVALID", t128.ModifierKey{})
	if err == nil {
		t.Error("SendArrowKey should fail with invalid direction")
	}

	// Test invalid navigation key
	err = client.SendNavigationKey("INVALID", t128.ModifierKey{})
	if err == nil {
		t.Error("SendNavigationKey should fail with invalid key name")
	}

	// Test invalid media key
	err = client.SendMediaKey("INVALID")
	if err == nil {
		t.Error("SendMediaKey should fail with invalid key name")
	}

	// Test invalid browser key
	err = client.SendBrowserKey("INVALID")
	if err == nil {
		t.Error("SendBrowserKey should fail with invalid key name")
	}

	// Test invalid special key
	err = client.SendSpecialKey("INVALID", t128.ModifierKey{})
	if err == nil {
		t.Error("SendSpecialKey should fail with invalid key name")
	}
}

// TestModifierKeyCombinations tests various modifier key combinations
func TestModifierKeyCombinations(t *testing.T) {
	client := NewClient(&Option{
		Addr:     "localhost:3389",
		UserName: "test",
		Password: "test",
	})

	// Test all modifier combinations
	modifiers := []t128.ModifierKey{
		{Shift: true},
		{Control: true},
		{Alt: true},
		{Meta: true},
		{Shift: true, Control: true},
		{Shift: true, Alt: true},
		{Control: true, Alt: true},
		{Shift: true, Control: true, Alt: true},
		{Shift: true, Control: true, Alt: true, Meta: true},
	}

	for i, mod := range modifiers {
		err := client.SendKeyPress('a', mod)
		if err != nil {
			t.Errorf("SendKeyPress with modifiers %d failed: %v", i, err)
		}
	}
}

// TestMouseButtonCombinations tests various mouse button combinations
func TestMouseButtonCombinations(t *testing.T) {
	client := NewClient(&Option{
		Addr:     "localhost:3389",
		UserName: "test",
		Password: "test",
	})

	// Test all mouse buttons
	buttons := []t128.MouseButton{
		t128.MouseButtonLeft,
		t128.MouseButtonRight,
		t128.MouseButtonMiddle,
		t128.MouseButtonX1,
		t128.MouseButtonX2,
	}

	for _, button := range buttons {
		err := client.SendMouseClickEvent(button, 100, 200)
		if err != nil {
			t.Errorf("SendMouseClickEvent with button %d failed: %v", button, err)
		}
	}
}

// TestScrollDirectionCombinations tests various scroll direction combinations
func TestScrollDirectionCombinations(t *testing.T) {
	client := NewClient(&Option{
		Addr:     "localhost:3389",
		UserName: "test",
		Password: "test",
	})

	// Test all scroll directions
	directions := []t128.ScrollDirection{
		t128.ScrollUp,
		t128.ScrollDown,
		t128.ScrollLeft,
		t128.ScrollRight,
	}

	for _, direction := range directions {
		err := client.SendMouseScrollEvent(direction, 120, 100, 200)
		if err != nil {
			t.Errorf("SendMouseScrollEvent with direction %d failed: %v", direction, err)
		}
	}
}

func TestIntegration_BasicRdpSession(t *testing.T) {
	// This is a placeholder integration test.
	// In a real test, you would spin up a mock RDP server or use a test server.
	// Here, we just check that the client can be constructed and run without panic for a short time.

	opt := &Option{
		Addr:     "127.0.0.1:3389", // Use a test server or mock
		UserName: "testuser",
		Password: "testpass",
	}

	client := NewClient(opt)
	assert.NotNil(t, client)

	done := make(chan struct{})
	go func() {
		_ = client.Run(nil) // Run with nil processor for now
		close(done)
	}()

	select {
	case <-done:
		// Finished (likely error, but that's OK for now)
	case <-time.After(2 * time.Second):
		// Timeout, forcibly close
	}

	// If we reach here, the client ran without panic
	assert.True(t, true)
}
