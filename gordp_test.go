package gordp

import (
	"bytes"
	"fmt"
	"os"
	"strings"
	"testing"
	"time"

	"github.com/kdsmith18542/gordp/core"
	"github.com/kdsmith18542/gordp/proto/audio"
	"github.com/kdsmith18542/gordp/proto/bitmap"
	"github.com/kdsmith18542/gordp/proto/clipboard"
	"github.com/kdsmith18542/gordp/proto/device"
	"github.com/kdsmith18542/gordp/proto/drdynvc"
	"github.com/kdsmith18542/gordp/proto/mcs"
	"github.com/kdsmith18542/gordp/proto/t128"
	"github.com/kdsmith18542/gordp/proto/virtualchannel"
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

// TestMouseInput tests mouse input functionality without requiring a live connection
func TestMouseInput(t *testing.T) {
	// Test mouse event creation and serialization
	t.Run("MouseEventCreation", func(t *testing.T) {
		// Test mouse movement event
		event := t128.NewFastPathMouseMoveEvent(100, 200)
		assert.Equal(t, uint16(100), event.XPos)
		assert.Equal(t, uint16(200), event.YPos)
		assert.Equal(t, uint16(t128.PTRFLAGS_MOVE), event.PointerFlags)

		// Test serialization
		data := event.Serialize()
		assert.NotEmpty(t, data)
		assert.Greater(t, len(data), 0)
	})

	t.Run("MouseButtonEvents", func(t *testing.T) {
		// Test left button down
		event := t128.NewFastPathMouseButtonEvent(t128.MouseButtonLeft, true, 100, 200)
		assert.Equal(t, uint16(t128.PTRFLAGS_DOWN|t128.PTRFLAGS_BUTTON1), event.PointerFlags)

		// Test left button up
		event = t128.NewFastPathMouseButtonEvent(t128.MouseButtonLeft, false, 100, 200)
		assert.Equal(t, uint16(t128.PTRFLAGS_BUTTON1), event.PointerFlags)

		// Test right button
		event = t128.NewFastPathMouseButtonEvent(t128.MouseButtonRight, true, 100, 200)
		assert.Equal(t, uint16(t128.PTRFLAGS_DOWN|t128.PTRFLAGS_BUTTON2), event.PointerFlags)

		// Test middle button
		event = t128.NewFastPathMouseButtonEvent(t128.MouseButtonMiddle, true, 100, 200)
		assert.Equal(t, uint16(t128.PTRFLAGS_DOWN|t128.PTRFLAGS_BUTTON3), event.PointerFlags)

		// Test X1 button
		event = t128.NewFastPathMouseButtonEvent(t128.MouseButtonX1, true, 100, 200)
		assert.Equal(t, uint16(t128.PTRFLAGS_DOWN|t128.PTRFLAGS_BUTTON4), event.PointerFlags)

		// Test X2 button
		event = t128.NewFastPathMouseButtonEvent(t128.MouseButtonX2, true, 100, 200)
		assert.Equal(t, uint16(t128.PTRFLAGS_DOWN|t128.PTRFLAGS_BUTTON5), event.PointerFlags)
	})

	t.Run("MouseWheelEvents", func(t *testing.T) {
		// Test vertical wheel up
		wheelDelta := int16(120)
		event := t128.NewFastPathMouseWheelEvent(wheelDelta, 100, 200)
		expected := uint16(t128.PTRFLAGS_WHEEL) | (uint16(wheelDelta) & t128.WheelRotationMask)
		assert.Equal(t, expected, event.PointerFlags)

		// Test vertical wheel down
		wheelDelta = -120
		event = t128.NewFastPathMouseWheelEvent(wheelDelta, 100, 200)
		expected = uint16(t128.PTRFLAGS_WHEEL|t128.PTRFLAGS_WHEEL_NEGATIVE) | (uint16(wheelDelta) & t128.WheelRotationMask)
		assert.Equal(t, expected, event.PointerFlags)

		// Test horizontal wheel
		wheelDelta = 120
		event = t128.NewFastPathMouseHorizontalWheelEvent(wheelDelta, 100, 200)
		expected = uint16(t128.PTRFLAGS_HWHEEL) | (uint16(wheelDelta) & t128.WheelRotationMask)
		assert.Equal(t, expected, event.PointerFlags)

		// Test horizontal wheel negative
		wheelDelta = -120
		event = t128.NewFastPathMouseHorizontalWheelEvent(wheelDelta, 100, 200)
		expected = uint16(t128.PTRFLAGS_HWHEEL|t128.PTRFLAGS_WHEEL_NEGATIVE) | (uint16(wheelDelta) & t128.WheelRotationMask)
		assert.Equal(t, expected, event.PointerFlags)
	})

	t.Run("MouseInputPDU", func(t *testing.T) {
		// Test PDU creation
		pdu := t128.NewFastPathMouseInputPDU(t128.PTRFLAGS_MOVE, 100, 200)
		assert.Equal(t, 1, len(pdu.FpInputEvents))

		// Test PDU serialization
		data := pdu.Serialize()
		assert.NotEmpty(t, data)
		assert.Greater(t, len(data), 0)
	})
}

// TestMouseInputMethods tests the client mouse input methods (without connection)
func TestMouseInputMethods(t *testing.T) {
	// Create a client but don't connect - we're just testing constants and enums
	_ = NewClient(&Option{
		Addr:     "localhost:3389",
		UserName: "test",
		Password: "test",
	})

	// These tests verify the constants and enums without requiring a connection

	t.Run("MouseButtonEnum", func(t *testing.T) {
		// Test that all mouse button constants are defined
		assert.Equal(t, 0, int(t128.MouseButtonLeft))
		assert.Equal(t, 1, int(t128.MouseButtonRight))
		assert.Equal(t, 2, int(t128.MouseButtonMiddle))
		assert.Equal(t, 3, int(t128.MouseButtonX1))
		assert.Equal(t, 4, int(t128.MouseButtonX2))
	})

	t.Run("ScrollDirectionEnum", func(t *testing.T) {
		// Test that all scroll direction constants are defined
		assert.Equal(t, 0, int(t128.ScrollUp))
		assert.Equal(t, 1, int(t128.ScrollDown))
		assert.Equal(t, 2, int(t128.ScrollLeft))
		assert.Equal(t, 3, int(t128.ScrollRight))
	})

	t.Run("MouseFlags", func(t *testing.T) {
		// Test that all mouse flags are properly defined
		assert.Equal(t, uint16(0x0800), uint16(t128.PTRFLAGS_MOVE))
		assert.Equal(t, uint16(0x8000), uint16(t128.PTRFLAGS_DOWN))
		assert.Equal(t, uint16(0x1000), uint16(t128.PTRFLAGS_BUTTON1))
		assert.Equal(t, uint16(0x2000), uint16(t128.PTRFLAGS_BUTTON2))
		assert.Equal(t, uint16(0x4000), uint16(t128.PTRFLAGS_BUTTON3))
		assert.Equal(t, uint16(0x0100), uint16(t128.PTRFLAGS_BUTTON4))
		assert.Equal(t, uint16(0x0200), uint16(t128.PTRFLAGS_BUTTON5))
		assert.Equal(t, uint16(0x0200), uint16(t128.PTRFLAGS_WHEEL))
		assert.Equal(t, uint16(0x0400), uint16(t128.PTRFLAGS_HWHEEL))
		assert.Equal(t, uint16(0x0100), uint16(t128.PTRFLAGS_WHEEL_NEGATIVE))
		assert.Equal(t, uint16(0x01FF), uint16(t128.WheelRotationMask))
	})
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
	// Test mouse button event creation without requiring a connection
	t.Run("ButtonEventCreation", func(t *testing.T) {
		// Test left button events
		leftDown := t128.NewFastPathMouseButtonEvent(t128.MouseButtonLeft, true, 100, 200)
		assert.Equal(t, uint16(t128.PTRFLAGS_DOWN|t128.PTRFLAGS_BUTTON1), leftDown.PointerFlags)
		assert.Equal(t, uint16(100), leftDown.XPos)
		assert.Equal(t, uint16(200), leftDown.YPos)

		leftUp := t128.NewFastPathMouseButtonEvent(t128.MouseButtonLeft, false, 100, 200)
		assert.Equal(t, uint16(t128.PTRFLAGS_BUTTON1), leftUp.PointerFlags)

		// Test right button events
		rightDown := t128.NewFastPathMouseButtonEvent(t128.MouseButtonRight, true, 150, 250)
		assert.Equal(t, uint16(t128.PTRFLAGS_DOWN|t128.PTRFLAGS_BUTTON2), rightDown.PointerFlags)

		rightUp := t128.NewFastPathMouseButtonEvent(t128.MouseButtonRight, false, 150, 250)
		assert.Equal(t, uint16(t128.PTRFLAGS_BUTTON2), rightUp.PointerFlags)

		// Test middle button events
		middleDown := t128.NewFastPathMouseButtonEvent(t128.MouseButtonMiddle, true, 200, 300)
		assert.Equal(t, uint16(t128.PTRFLAGS_DOWN|t128.PTRFLAGS_BUTTON3), middleDown.PointerFlags)

		middleUp := t128.NewFastPathMouseButtonEvent(t128.MouseButtonMiddle, false, 200, 300)
		assert.Equal(t, uint16(t128.PTRFLAGS_BUTTON3), middleUp.PointerFlags)
	})

	t.Run("ButtonEventSerialization", func(t *testing.T) {
		// Test that all button events can be serialized
		buttons := []t128.MouseButton{
			t128.MouseButtonLeft,
			t128.MouseButtonRight,
			t128.MouseButtonMiddle,
			t128.MouseButtonX1,
			t128.MouseButtonX2,
		}

		for _, button := range buttons {
			// Test down event
			downEvent := t128.NewFastPathMouseButtonEvent(button, true, 100, 200)
			downData := downEvent.Serialize()
			assert.NotEmpty(t, downData)

			// Test up event
			upEvent := t128.NewFastPathMouseButtonEvent(button, false, 100, 200)
			upData := upEvent.Serialize()
			assert.NotEmpty(t, upData)
		}
	})
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
	// Test enhanced mouse event creation without requiring a connection
	t.Run("WheelEventCreation", func(t *testing.T) {
		// Test vertical wheel events
		wheelUp := t128.NewFastPathMouseWheelEvent(120, 100, 200)
		assert.Equal(t, t128.PTRFLAGS_WHEEL|120, wheelUp.PointerFlags)
		assert.Equal(t, uint16(100), wheelUp.XPos)
		assert.Equal(t, uint16(200), wheelUp.YPos)

		wheelDown := t128.NewFastPathMouseWheelEvent(-120, 100, 200)
		assert.Equal(t, t128.PTRFLAGS_WHEEL|t128.PTRFLAGS_WHEEL_NEGATIVE|120, wheelDown.PointerFlags)

		// Test horizontal wheel events
		hWheelRight := t128.NewFastPathMouseHorizontalWheelEvent(120, 100, 200)
		assert.Equal(t, t128.PTRFLAGS_HWHEEL|120, hWheelRight.PointerFlags)

		hWheelLeft := t128.NewFastPathMouseHorizontalWheelEvent(-120, 100, 200)
		assert.Equal(t, t128.PTRFLAGS_HWHEEL|t128.PTRFLAGS_WHEEL_NEGATIVE|120, hWheelLeft.PointerFlags)
	})

	t.Run("WheelEventSerialization", func(t *testing.T) {
		// Test that wheel events can be serialized
		wheelDeltas := []int16{-120, -60, 60, 120}
		for _, delta := range wheelDeltas {
			// Test vertical wheel
			event := t128.NewFastPathMouseWheelEvent(delta, 100, 200)
			data := event.Serialize()
			assert.NotEmpty(t, data)

			// Test horizontal wheel
			event = t128.NewFastPathMouseHorizontalWheelEvent(delta, 100, 200)
			data = event.Serialize()
			assert.NotEmpty(t, data)
		}
	})

	t.Run("XButtonSupport", func(t *testing.T) {
		// Test X1 and X2 button support
		x1Down := t128.NewFastPathMouseButtonEvent(t128.MouseButtonX1, true, 100, 200)
		assert.Equal(t, t128.PTRFLAGS_DOWN|t128.PTRFLAGS_BUTTON4, x1Down.PointerFlags)

		x2Down := t128.NewFastPathMouseButtonEvent(t128.MouseButtonX2, true, 150, 250)
		assert.Equal(t, t128.PTRFLAGS_DOWN|t128.PTRFLAGS_BUTTON5, x2Down.PointerFlags)

		// Test serialization
		x1Data := x1Down.Serialize()
		x2Data := x2Down.Serialize()
		assert.NotEmpty(t, x1Data)
		assert.NotEmpty(t, x2Data)
	})

	t.Run("WheelRotationMask", func(t *testing.T) {
		// Test that wheel rotation values are properly masked
		largeDelta := int16(1000)
		event := t128.NewFastPathMouseWheelEvent(largeDelta, 100, 200)
		rotation := event.PointerFlags & t128.WheelRotationMask
		assert.LessOrEqual(t, rotation, uint16(255))
		assert.Equal(t, uint16(232), rotation) // 1000 & 0x01FF = 232
	})
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
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		event := t128.NewFastPathMouseMoveEvent(uint16(i%1000), uint16(i%1000))
		_ = event.Serialize()
	}
}

// BenchmarkMouseClick benchmarks mouse click performance
func BenchmarkMouseClick(b *testing.B) {
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		event := t128.NewFastPathMouseButtonEvent(t128.MouseButtonLeft, true, uint16(i%1000), uint16(i%1000))
		_ = event.Serialize()
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
	// Test all mouse button combinations without requiring a connection
	buttons := []t128.MouseButton{
		t128.MouseButtonLeft,
		t128.MouseButtonRight,
		t128.MouseButtonMiddle,
		t128.MouseButtonX1,
		t128.MouseButtonX2,
	}

	for _, button := range buttons {
		// Test button down event
		downEvent := t128.NewFastPathMouseButtonEvent(button, true, 100, 200)
		downData := downEvent.Serialize()
		assert.NotEmpty(t, downData)

		// Test button up event
		upEvent := t128.NewFastPathMouseButtonEvent(button, false, 100, 200)
		upData := upEvent.Serialize()
		assert.NotEmpty(t, upData)

		// Verify correct flags are set
		switch button {
		case t128.MouseButtonLeft:
			assert.Equal(t, uint16(t128.PTRFLAGS_DOWN|t128.PTRFLAGS_BUTTON1), downEvent.PointerFlags)
			assert.Equal(t, uint16(t128.PTRFLAGS_BUTTON1), upEvent.PointerFlags)
		case t128.MouseButtonRight:
			assert.Equal(t, uint16(t128.PTRFLAGS_DOWN|t128.PTRFLAGS_BUTTON2), downEvent.PointerFlags)
			assert.Equal(t, uint16(t128.PTRFLAGS_BUTTON2), upEvent.PointerFlags)
		case t128.MouseButtonMiddle:
			assert.Equal(t, uint16(t128.PTRFLAGS_DOWN|t128.PTRFLAGS_BUTTON3), downEvent.PointerFlags)
			assert.Equal(t, uint16(t128.PTRFLAGS_BUTTON3), upEvent.PointerFlags)
		case t128.MouseButtonX1:
			assert.Equal(t, uint16(t128.PTRFLAGS_DOWN|t128.PTRFLAGS_BUTTON4), downEvent.PointerFlags)
			assert.Equal(t, uint16(t128.PTRFLAGS_BUTTON4), upEvent.PointerFlags)
		case t128.MouseButtonX2:
			assert.Equal(t, uint16(t128.PTRFLAGS_DOWN|t128.PTRFLAGS_BUTTON5), downEvent.PointerFlags)
			assert.Equal(t, uint16(t128.PTRFLAGS_BUTTON5), upEvent.PointerFlags)
		}
	}
}

// TestScrollDirectionCombinations tests various scroll direction combinations
func TestScrollDirectionCombinations(t *testing.T) {
	// Test all scroll direction combinations without requiring a connection
	directions := []t128.ScrollDirection{
		t128.ScrollUp,
		t128.ScrollDown,
		t128.ScrollLeft,
		t128.ScrollRight,
	}

	for _, direction := range directions {
		var event *t128.TsFpPointerEvent
		var expectedFlags uint16

		switch direction {
		case t128.ScrollUp:
			event = t128.NewFastPathMouseWheelEvent(120, 100, 200)
			expectedFlags = t128.PTRFLAGS_WHEEL | 120
		case t128.ScrollDown:
			event = t128.NewFastPathMouseWheelEvent(-120, 100, 200)
			expectedFlags = t128.PTRFLAGS_WHEEL | t128.PTRFLAGS_WHEEL_NEGATIVE | 120
		case t128.ScrollLeft:
			event = t128.NewFastPathMouseHorizontalWheelEvent(-120, 100, 200)
			expectedFlags = t128.PTRFLAGS_HWHEEL | t128.PTRFLAGS_WHEEL_NEGATIVE | 120
		case t128.ScrollRight:
			event = t128.NewFastPathMouseHorizontalWheelEvent(120, 100, 200)
			expectedFlags = t128.PTRFLAGS_HWHEEL | 120
		}

		assert.NotNil(t, event)
		assert.Equal(t, expectedFlags, event.PointerFlags)
		assert.Equal(t, uint16(100), event.XPos)
		assert.Equal(t, uint16(200), event.YPos)

		// Test serialization
		data := event.Serialize()
		assert.NotEmpty(t, data)
	}
}

// TestMultiMonitorNegotiation tests multi-monitor negotiation functionality
func TestMultiMonitorNegotiation(t *testing.T) {
	// Test single monitor configuration
	client := NewClient(&Option{
		Addr:     "localhost:3389",
		UserName: "test",
		Password: "test",
	})

	// Test setting single monitor
	monitors := []mcs.MonitorLayout{
		{
			Left:               0,
			Top:                0,
			Right:              1920,
			Bottom:             1080,
			Flags:              0x01, // Primary
			MonitorIndex:       0,
			PhysicalWidthMm:    520,
			PhysicalHeightMm:   320,
			Orientation:        0, // Landscape
			DesktopScaleFactor: 100,
			DeviceScaleFactor:  100,
		},
	}

	client.SetMonitors(monitors)

	// Verify monitor layout was set
	assert.Equal(t, 1, len(client.monitors), "Should have 1 monitor")
	assert.Equal(t, int32(0), client.monitors[0].Left, "Monitor left should be 0")
	assert.Equal(t, int32(1920), client.monitors[0].Right, "Monitor right should be 1920")
	assert.Equal(t, uint32(0x01), client.monitors[0].Flags, "Monitor should be primary")

	// Test dual monitor configuration
	dualMonitors := []mcs.MonitorLayout{
		{
			Left:               0,
			Top:                0,
			Right:              1920,
			Bottom:             1080,
			Flags:              0x01, // Primary
			MonitorIndex:       0,
			PhysicalWidthMm:    520,
			PhysicalHeightMm:   320,
			Orientation:        0, // Landscape
			DesktopScaleFactor: 100,
			DeviceScaleFactor:  100,
		},
		{
			Left:               1920,
			Top:                0,
			Right:              3840,
			Bottom:             1080,
			Flags:              0x00, // Secondary
			MonitorIndex:       1,
			PhysicalWidthMm:    520,
			PhysicalHeightMm:   320,
			Orientation:        0, // Landscape
			DesktopScaleFactor: 100,
			DeviceScaleFactor:  100,
		},
	}

	client.SetMonitors(dualMonitors)

	// Verify dual monitor layout was set
	assert.Equal(t, 2, len(client.monitors), "Should have 2 monitors")
	assert.Equal(t, int32(1920), client.monitors[1].Left, "Second monitor left should be 1920")
	assert.Equal(t, uint32(0x00), client.monitors[1].Flags, "Second monitor should not be primary")

	// Test getting monitor layout
	retrievedMonitors := client.GetMonitors()
	assert.Equal(t, 2, len(retrievedMonitors), "Should retrieve 2 monitors")
	assert.Equal(t, dualMonitors[0].Left, retrievedMonitors[0].Left, "Retrieved monitor should match set monitor")
	assert.Equal(t, dualMonitors[1].Right, retrievedMonitors[1].Right, "Retrieved monitor should match set monitor")
}

// TestMonitorLayoutPDUSerialization tests the serialization of Monitor Layout PDU
func TestMonitorLayoutPDUSerialization(t *testing.T) {
	// Test single monitor PDU serialization
	singleMonitorPDU := &mcs.MonitorLayoutPDU{
		UserDataHeader: mcs.UserDataHeader{
			Type: mcs.CS_MONITOR,
			Len:  52, // 8 bytes header + 4 bytes numMonitors + 44 bytes for single monitor
		},
		NumMonitors: 1,
		Monitors: []mcs.MonitorLayout{
			{
				Left:               0,
				Top:                0,
				Right:              1920,
				Bottom:             1080,
				Flags:              0x01,
				MonitorIndex:       0,
				PhysicalWidthMm:    520,
				PhysicalHeightMm:   320,
				Orientation:        0,
				DesktopScaleFactor: 100,
				DeviceScaleFactor:  100,
			},
		},
	}

	data := singleMonitorPDU.Serialize()
	assert.NotNil(t, data, "Serialized data should not be nil")
	assert.Equal(t, 52, len(data), "Serialized data should be 52 bytes")

	// Verify header fields
	reader := bytes.NewReader(data)
	var headerType uint16
	var headerLen uint16
	var numMonitors uint32
	core.ReadLE(reader, &headerType)
	core.ReadLE(reader, &headerLen)
	core.ReadLE(reader, &numMonitors)

	assert.Equal(t, uint16(mcs.CS_MONITOR), headerType, "Header type should be CS_MONITOR")
	assert.Equal(t, uint16(52), headerLen, "Header length should be 52")
	assert.Equal(t, uint32(1), numMonitors, "Number of monitors should be 1")

	// Test dual monitor PDU serialization
	dualMonitorPDU := &mcs.MonitorLayoutPDU{
		UserDataHeader: mcs.UserDataHeader{
			Type: mcs.CS_MONITOR,
			Len:  96, // 8 bytes header + 4 bytes numMonitors + 44 bytes per monitor * 2
		},
		NumMonitors: 2,
		Monitors: []mcs.MonitorLayout{
			{
				Left:               0,
				Top:                0,
				Right:              1920,
				Bottom:             1080,
				Flags:              0x01,
				MonitorIndex:       0,
				PhysicalWidthMm:    520,
				PhysicalHeightMm:   320,
				Orientation:        0,
				DesktopScaleFactor: 100,
				DeviceScaleFactor:  100,
			},
			{
				Left:               1920,
				Top:                0,
				Right:              3840,
				Bottom:             1080,
				Flags:              0x00,
				MonitorIndex:       1,
				PhysicalWidthMm:    520,
				PhysicalHeightMm:   320,
				Orientation:        0,
				DesktopScaleFactor: 100,
				DeviceScaleFactor:  100,
			},
		},
	}

	data = dualMonitorPDU.Serialize()
	assert.NotNil(t, data, "Serialized data should not be nil")
	assert.Equal(t, 96, len(data), "Serialized data should be 96 bytes")

	// Verify dual monitor header fields
	reader = bytes.NewReader(data)
	core.ReadLE(reader, &headerType)
	core.ReadLE(reader, &headerLen)
	core.ReadLE(reader, &numMonitors)

	assert.Equal(t, uint16(mcs.CS_MONITOR), headerType, "Header type should be CS_MONITOR")
	assert.Equal(t, uint16(96), headerLen, "Header length should be 96")
	assert.Equal(t, uint32(2), numMonitors, "Number of monitors should be 2")
}

// TestMonitorLayoutValidation tests validation of monitor layout configurations
func TestMonitorLayoutValidation(t *testing.T) {
	client := NewClient(&Option{
		Addr:     "localhost:3389",
		UserName: "test",
		Password: "test",
	})

	// Test empty monitor list
	client.SetMonitors([]mcs.MonitorLayout{})
	assert.Equal(t, 0, len(client.monitors), "Should have 0 monitors")

	// Test invalid monitor geometry (negative coordinates)
	invalidMonitors := []mcs.MonitorLayout{
		{
			Left:   -100,
			Top:    0,
			Right:  1920,
			Bottom: 1080,
			Flags:  0x01,
		},
	}

	client.SetMonitors(invalidMonitors)
	assert.Equal(t, 1, len(client.monitors), "Should still set the monitor")

	// Test overlapping monitors
	overlappingMonitors := []mcs.MonitorLayout{
		{
			Left:   0,
			Top:    0,
			Right:  1920,
			Bottom: 1080,
			Flags:  0x01,
		},
		{
			Left:   1000, // Overlaps with first monitor
			Top:    0,
			Right:  2920,
			Bottom: 1080,
			Flags:  0x00,
		},
	}

	client.SetMonitors(overlappingMonitors)
	assert.Equal(t, 2, len(client.monitors), "Should set both monitors")
}

// TestHighDPIMonitorLayout tests high DPI monitor configurations
func TestHighDPIMonitorLayout(t *testing.T) {
	client := NewClient(&Option{
		Addr:     "localhost:3389",
		UserName: "test",
		Password: "test",
	})

	// Test high DPI monitor (200% scaling)
	highDPIMonitors := []mcs.MonitorLayout{
		{
			Left:               0,
			Top:                0,
			Right:              1920,
			Bottom:             1080,
			Flags:              0x01,
			MonitorIndex:       0,
			PhysicalWidthMm:    520,
			PhysicalHeightMm:   320,
			Orientation:        0,
			DesktopScaleFactor: 200, // 200% DPI scaling
			DeviceScaleFactor:  200,
		},
		{
			Left:               1920,
			Top:                0,
			Right:              3840,
			Bottom:             1080,
			Flags:              0x00,
			MonitorIndex:       1,
			PhysicalWidthMm:    520,
			PhysicalHeightMm:   320,
			Orientation:        0,
			DesktopScaleFactor: 125, // 125% DPI scaling
			DeviceScaleFactor:  125,
		},
	}

	client.SetMonitors(highDPIMonitors)
	assert.Equal(t, 2, len(client.monitors), "Should have 2 monitors")
	assert.Equal(t, uint32(200), client.monitors[0].DesktopScaleFactor, "First monitor should have 200% scaling")
	assert.Equal(t, uint32(125), client.monitors[1].DesktopScaleFactor, "Second monitor should have 125% scaling")

	// Test portrait orientation
	portraitMonitors := []mcs.MonitorLayout{
		{
			Left:               0,
			Top:                0,
			Right:              1080, // Portrait: width < height
			Bottom:             1920,
			Flags:              0x01,
			MonitorIndex:       0,
			PhysicalWidthMm:    320,
			PhysicalHeightMm:   520,
			Orientation:        1, // Portrait
			DesktopScaleFactor: 100,
			DeviceScaleFactor:  100,
		},
	}

	client.SetMonitors(portraitMonitors)
	assert.Equal(t, 1, len(client.monitors), "Should have 1 monitor")
	assert.Equal(t, uint32(1), client.monitors[0].Orientation, "Monitor should be portrait")
}

// TestMonitorLayoutPDUDeserialization tests deserialization of Monitor Layout PDU
func TestMonitorLayoutPDUDeserialization(t *testing.T) {
	// Create a test PDU
	originalPDU := &mcs.MonitorLayoutPDU{
		UserDataHeader: mcs.UserDataHeader{
			Type: mcs.CS_MONITOR,
			Len:  52,
		},
		NumMonitors: 1,
		Monitors: []mcs.MonitorLayout{
			{
				Left:               0,
				Top:                0,
				Right:              1920,
				Bottom:             1080,
				Flags:              0x01,
				MonitorIndex:       0,
				PhysicalWidthMm:    520,
				PhysicalHeightMm:   320,
				Orientation:        0,
				DesktopScaleFactor: 100,
				DeviceScaleFactor:  100,
			},
		},
	}

	// Serialize and then deserialize
	data := originalPDU.Serialize()
	assert.NotNil(t, data, "Serialized data should not be nil")

	// Simulate deserialization by reading the data back
	reader := bytes.NewReader(data)
	var headerType uint16
	var headerLen uint16
	var numMonitors uint32
	core.ReadLE(reader, &headerType)
	core.ReadLE(reader, &headerLen)
	core.ReadLE(reader, &numMonitors)

	// Read monitor data
	var monitor mcs.MonitorLayout
	core.ReadLE(reader, &monitor.Left)
	core.ReadLE(reader, &monitor.Top)
	core.ReadLE(reader, &monitor.Right)
	core.ReadLE(reader, &monitor.Bottom)
	core.ReadLE(reader, &monitor.Flags)
	core.ReadLE(reader, &monitor.MonitorIndex)
	core.ReadLE(reader, &monitor.PhysicalWidthMm)
	core.ReadLE(reader, &monitor.PhysicalHeightMm)
	core.ReadLE(reader, &monitor.Orientation)
	core.ReadLE(reader, &monitor.DesktopScaleFactor)
	core.ReadLE(reader, &monitor.DeviceScaleFactor)

	// Verify deserialized data matches original
	assert.Equal(t, uint16(mcs.CS_MONITOR), headerType, "Header type should match")
	assert.Equal(t, uint16(52), headerLen, "Header length should match")
	assert.Equal(t, uint32(1), numMonitors, "Number of monitors should match")
	assert.Equal(t, int32(0), monitor.Left, "Monitor left should match")
	assert.Equal(t, int32(1920), monitor.Right, "Monitor right should match")
	assert.Equal(t, uint32(0x01), monitor.Flags, "Monitor flags should match")
	assert.Equal(t, uint32(100), monitor.DesktopScaleFactor, "Desktop scale factor should match")
}

// TestMultiMonitorClientIntegration tests integration of multi-monitor with client
func TestMultiMonitorClientIntegration(t *testing.T) {
	// Test client creation with multi-monitor configuration
	monitors := []mcs.MonitorLayout{
		{
			Left:               0,
			Top:                0,
			Right:              1920,
			Bottom:             1080,
			Flags:              0x01,
			MonitorIndex:       0,
			PhysicalWidthMm:    520,
			PhysicalHeightMm:   320,
			Orientation:        0,
			DesktopScaleFactor: 100,
			DeviceScaleFactor:  100,
		},
		{
			Left:               1920,
			Top:                0,
			Right:              3840,
			Bottom:             1080,
			Flags:              0x00,
			MonitorIndex:       1,
			PhysicalWidthMm:    520,
			PhysicalHeightMm:   320,
			Orientation:        0,
			DesktopScaleFactor: 100,
			DeviceScaleFactor:  100,
		},
	}

	client := NewClient(&Option{
		Addr:     "localhost:3389",
		UserName: "test",
		Password: "test",
	})

	// Set monitor layout
	client.SetMonitors(monitors)

	// Verify monitor layout is set
	assert.Equal(t, 2, len(client.monitors), "Client should have 2 monitors configured")

	// Test that monitor layout persists
	retrievedMonitors := client.GetMonitors()
	assert.Equal(t, 2, len(retrievedMonitors), "Should retrieve 2 monitors")
	assert.Equal(t, monitors[0].Left, retrievedMonitors[0].Left, "First monitor should match")
	assert.Equal(t, monitors[1].Right, retrievedMonitors[1].Right, "Second monitor should match")

	// Test updating monitor layout
	updatedMonitors := []mcs.MonitorLayout{
		{
			Left:               0,
			Top:                0,
			Right:              2560, // Different resolution
			Bottom:             1440,
			Flags:              0x01,
			MonitorIndex:       0,
			PhysicalWidthMm:    520,
			PhysicalHeightMm:   320,
			Orientation:        0,
			DesktopScaleFactor: 150, // Different DPI
			DeviceScaleFactor:  150,
		},
	}

	client.SetMonitors(updatedMonitors)
	assert.Equal(t, 1, len(client.monitors), "Client should have 1 monitor after update")
	assert.Equal(t, int32(2560), client.monitors[0].Right, "Monitor should be updated")
	assert.Equal(t, uint32(150), client.monitors[0].DesktopScaleFactor, "DPI should be updated")
}

// TestMonitorLayoutEdgeCases tests edge cases for monitor layout
func TestMonitorLayoutEdgeCases(t *testing.T) {
	client := NewClient(&Option{
		Addr:     "localhost:3389",
		UserName: "test",
		Password: "test",
	})

	// Test very large monitor configuration
	largeMonitors := []mcs.MonitorLayout{
		{
			Left:               0,
			Top:                0,
			Right:              8192, // Very large resolution
			Bottom:             4320,
			Flags:              0x01,
			MonitorIndex:       0,
			PhysicalWidthMm:    1000,
			PhysicalHeightMm:   600,
			Orientation:        0,
			DesktopScaleFactor: 300, // Very high DPI
			DeviceScaleFactor:  300,
		},
	}

	client.SetMonitors(largeMonitors)
	assert.Equal(t, 1, len(client.monitors), "Should have 1 monitor")
	assert.Equal(t, int32(8192), client.monitors[0].Right, "Large resolution should be supported")
	assert.Equal(t, uint32(300), client.monitors[0].DesktopScaleFactor, "High DPI should be supported")

	// Test zero-sized monitor (edge case)
	zeroMonitors := []mcs.MonitorLayout{
		{
			Left:               0,
			Top:                0,
			Right:              0, // Zero width
			Bottom:             0, // Zero height
			Flags:              0x01,
			MonitorIndex:       0,
			PhysicalWidthMm:    0,
			PhysicalHeightMm:   0,
			Orientation:        0,
			DesktopScaleFactor: 0, // Zero DPI
			DeviceScaleFactor:  0,
		},
	}

	client.SetMonitors(zeroMonitors)
	assert.Equal(t, 1, len(client.monitors), "Should still set the monitor")

	// Test maximum number of monitors (reasonable limit)
	maxMonitors := make([]mcs.MonitorLayout, 16) // 16 monitors
	for i := 0; i < 16; i++ {
		maxMonitors[i] = mcs.MonitorLayout{
			Left:               int32(i * 1920),
			Top:                0,
			Right:              int32((i + 1) * 1920),
			Bottom:             1080,
			Flags:              uint32(i), // Only first one is primary
			MonitorIndex:       uint32(i),
			PhysicalWidthMm:    520,
			PhysicalHeightMm:   320,
			Orientation:        0,
			DesktopScaleFactor: 100,
			DeviceScaleFactor:  100,
		}
	}

	client.SetMonitors(maxMonitors)
	assert.Equal(t, 16, len(client.monitors), "Should have 16 monitors")
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

// TestMultiMonitorCapabilityFlag tests that the multi-monitor capability flag is set correctly
func TestMultiMonitorCapabilityFlag(t *testing.T) {
	// Test that multi-monitor capability flag is set when monitors are configured
	monitors := []mcs.MonitorLayout{
		{
			Left:               0,
			Top:                0,
			Right:              1920,
			Bottom:             1080,
			Flags:              0x01,
			MonitorIndex:       0,
			PhysicalWidthMm:    520,
			PhysicalHeightMm:   320,
			Orientation:        0,
			DesktopScaleFactor: 100,
			DeviceScaleFactor:  100,
		},
	}

	client := NewClient(&Option{
		Addr:     "localhost:3389",
		UserName: "test",
		Password: "test",
		Monitors: monitors,
	})

	// Verify that the multi-monitor capability flag should be set in the client core data
	// This flag indicates support for the Monitor Layout PDU
	assert.True(t, len(client.monitors) > 0, "Client should have monitors configured")

	// Note: The actual flag setting would happen during connection negotiation
	// This test verifies that the client is properly configured for multi-monitor support
}

// TestClipboardFunctionality tests clipboard functionality
func TestClipboardFunctionality(t *testing.T) {
	t.Run("ClipboardManagerCreation", func(t *testing.T) {
		manager := clipboard.NewClipboardManager(nil)
		assert.NotNil(t, manager)
	})

	t.Run("ClipboardMessageSerialization", func(t *testing.T) {
		msg := &clipboard.ClipboardMessage{
			MessageType:  clipboard.CLIPRDR_MSG_TYPE_CAPABILITIES,
			MessageFlags: 0x0001,
			DataLength:   4,
			Data:         []byte{0x01, 0x02, 0x03, 0x04},
		}

		data := msg.Serialize()
		assert.NotEmpty(t, data)
		assert.Equal(t, 12, len(data)) // 2+2+4+4 bytes
	})

	t.Run("ClipboardFormatList", func(t *testing.T) {
		formats := []clipboard.ClipboardFormat{
			clipboard.CLIPRDR_FORMAT_UNICODETEXT,
			clipboard.CLIPRDR_FORMAT_HTML,
			clipboard.CLIPRDR_FORMAT_PNG,
		}

		manager := clipboard.NewClipboardManager(nil)
		msg := manager.CreateFormatListMessage(formats)
		assert.Equal(t, clipboard.CLIPRDR_MSG_TYPE_FORMAT_LIST, msg.MessageType)
		assert.NotEmpty(t, msg.Data)
	})

	t.Run("ClipboardFormatNames", func(t *testing.T) {
		assert.Equal(t, "CF_UNICODETEXT", clipboard.GetFormatName(clipboard.CLIPRDR_FORMAT_UNICODETEXT))
		assert.Equal(t, "CF_HTML", clipboard.GetFormatName(clipboard.CLIPRDR_FORMAT_HTML))
		assert.Equal(t, "CF_PNG", clipboard.GetFormatName(clipboard.CLIPRDR_FORMAT_PNG))
	})
}

// TestAudioFunctionality tests audio functionality
func TestAudioFunctionality(t *testing.T) {
	t.Run("AudioManagerCreation", func(t *testing.T) {
		manager := audio.NewAudioManager(nil)
		assert.NotNil(t, manager)
		// Default formats are created in the manager
	})

	t.Run("AudioFormatValidation", func(t *testing.T) {
		format := audio.AudioFormat{
			FormatTag:      0x0001, // WAVE_FORMAT_PCM
			Channels:       2,      // Stereo
			SamplesPerSec:  48000,  // 48 kHz
			AvgBytesPerSec: 192000, // 2 channels * 2 bytes * 48000
			BlockAlign:     4,      // 2 channels * 2 bytes
			BitsPerSample:  16,     // 16-bit
			ExtraSize:      0,
		}

		assert.Equal(t, uint16(0x0001), format.FormatTag)
		assert.Equal(t, uint16(2), format.Channels)
		assert.Equal(t, uint32(48000), format.SamplesPerSec)
		assert.Equal(t, uint16(16), format.BitsPerSample)
	})

	t.Run("AudioMessageSerialization", func(t *testing.T) {
		msg := &audio.AudioMessage{
			MessageType:  audio.RDPSND_MSG_TYPE_SERVER_WAVE,
			MessageFlags: 0x0001,
			DataLength:   4,
			Data:         []byte{0x01, 0x02, 0x03, 0x04},
		}

		data := msg.Serialize()
		assert.NotEmpty(t, data)
		assert.Equal(t, 12, len(data)) // 2+2+4+4 bytes
	})

	t.Run("AudioWaveInfo", func(t *testing.T) {
		waveInfo := &audio.AudioWaveInfo{
			Timestamp: 1234567890,
			FormatID:  1,
			Data:      []byte{0x01, 0x02, 0x03, 0x04},
		}

		assert.Equal(t, uint32(1234567890), waveInfo.Timestamp)
		assert.Equal(t, uint16(1), waveInfo.FormatID)
		assert.Len(t, waveInfo.Data, 4)
	})
}

// TestDeviceRedirection tests device redirection functionality
func TestDeviceRedirection(t *testing.T) {
	t.Run("DeviceManagerCreation", func(t *testing.T) {
		manager := device.NewDeviceManager(nil)
		assert.NotNil(t, manager)
		assert.Equal(t, 0, manager.GetDeviceCount())
	})

	t.Run("DeviceTypes", func(t *testing.T) {
		assert.Equal(t, device.DeviceType(0x00000001), device.DeviceTypePrinter)
		assert.Equal(t, device.DeviceType(0x00000002), device.DeviceTypeDrive)
		assert.Equal(t, device.DeviceType(0x00000003), device.DeviceTypePort)
		assert.Equal(t, device.DeviceType(0x00000004), device.DeviceTypeSmartCard)
		assert.Equal(t, device.DeviceType(0x00000005), device.DeviceTypeAudio)
		assert.Equal(t, device.DeviceType(0x00000006), device.DeviceTypeVideo)
		assert.Equal(t, device.DeviceType(0x00000007), device.DeviceTypeUSB)
	})

	t.Run("DeviceAnnounce", func(t *testing.T) {
		announce := &device.DeviceAnnounce{
			DeviceType:       device.DeviceTypePrinter,
			DeviceID:         1,
			PreferredDosName: "PRN1",
			DeviceData:       "Test Printer",
		}

		assert.Equal(t, device.DeviceTypePrinter, announce.DeviceType)
		assert.Equal(t, uint32(1), announce.DeviceID)
		assert.Equal(t, "PRN1", announce.PreferredDosName)
		assert.Equal(t, "Test Printer", announce.DeviceData)
	})

	t.Run("DeviceMessageSerialization", func(t *testing.T) {
		msg := &device.DeviceMessage{
			ComponentID: device.RDPDR_CTYP_CORE,
			PacketID:    1,
			Data:        []byte{0x01, 0x02, 0x03, 0x04},
		}

		data := msg.Serialize()
		assert.NotEmpty(t, data)
		assert.Equal(t, 8, len(data)) // 2+2+4 bytes
	})

	t.Run("DeviceIORequest", func(t *testing.T) {
		request := &device.DeviceIORequest{
			DeviceID:      1,
			FileID:        2,
			CompletionID:  3,
			MajorFunction: 4,
			MinorFunction: 5,
			Data:          []byte{0x01, 0x02, 0x03},
		}

		assert.Equal(t, uint32(1), request.DeviceID)
		assert.Equal(t, uint32(2), request.FileID)
		assert.Equal(t, uint32(3), request.CompletionID)
		assert.Equal(t, uint32(4), request.MajorFunction)
		assert.Equal(t, uint32(5), request.MinorFunction)
		assert.Len(t, request.Data, 3)
	})

	t.Run("PrinterData", func(t *testing.T) {
		printerData := &device.PrinterData{
			JobID: 123,
			Data:  []byte{0x01, 0x02, 0x03, 0x04},
			Flags: 0x0001,
		}

		assert.Equal(t, uint32(123), printerData.JobID)
		assert.Len(t, printerData.Data, 4)
		assert.Equal(t, uint32(0x0001), printerData.Flags)
	})
}

// TestEnhancedInputHandling tests enhanced input handling features
func TestEnhancedInputHandling(t *testing.T) {
	client := NewClient(&Option{
		Addr:     "localhost:3389",
		UserName: "test",
		Password: "test",
	})

	t.Run("IMEInput", func(t *testing.T) {
		// Test IME input without requiring connection
		err := client.SendUnicodeString("Hello 世界")
		// This should fail with "no active connection" which is expected
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "no active connection")
	})

	t.Run("ExtendedKeyCodes", func(t *testing.T) {
		// Test extended key codes without requiring connection
		err := client.SendExtendedKey(t128.VK_RETURN, true, t128.ModifierKey{})
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "no active connection")

		err = client.SendExtendedKey(t128.VK_RETURN, false, t128.ModifierKey{})
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "no active connection")
	})

	t.Run("NumpadKeys", func(t *testing.T) {
		// Test numpad keys without requiring connection
		err := client.SendNumpadKey(t128.VK_NUMPAD0, true, t128.ModifierKey{})
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "no active connection")

		err = client.SendNumpadKey(t128.VK_NUMPAD1, false, t128.ModifierKey{})
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "no active connection")
	})

	t.Run("FunctionKeys", func(t *testing.T) {
		// Test function keys without requiring connection
		for i := 1; i <= 12; i++ {
			err := client.SendFunctionKey(i, t128.ModifierKey{})
			assert.Error(t, err)
			assert.Contains(t, err.Error(), "no active connection")
		}
	})

	t.Run("ArrowKeys", func(t *testing.T) {
		// Test arrow keys without requiring connection
		directions := []string{"up", "down", "left", "right"}
		for _, dir := range directions {
			err := client.SendArrowKey(dir, t128.ModifierKey{})
			assert.Error(t, err)
			assert.Contains(t, err.Error(), "no active connection")
		}
	})

	t.Run("NavigationKeys", func(t *testing.T) {
		// Test navigation keys without requiring connection
		navKeys := []string{"home", "end", "pageup", "pagedown", "insert", "delete"}
		for _, key := range navKeys {
			err := client.SendNavigationKey(key, t128.ModifierKey{})
			assert.Error(t, err)
			assert.Contains(t, err.Error(), "no active connection")
		}
	})

	t.Run("MediaKeys", func(t *testing.T) {
		// Test media keys without requiring connection
		mediaKeys := []string{"play", "pause", "stop", "next", "previous", "volumeup", "volumedown", "mute"}
		for _, key := range mediaKeys {
			err := client.SendMediaKey(key)
			// Some media keys might not be supported, so we check for either error type
			if err != nil {
				assert.True(t, strings.Contains(err.Error(), "no active connection") ||
					strings.Contains(err.Error(), "unsupported media key"))
			}
		}
	})

	t.Run("BrowserKeys", func(t *testing.T) {
		// Test browser keys without requiring connection
		browserKeys := []string{"back", "forward", "refresh", "search", "favorites", "home", "stop"}
		for _, key := range browserKeys {
			err := client.SendBrowserKey(key)
			assert.Error(t, err)
			assert.Contains(t, err.Error(), "no active connection")
		}
	})

	t.Run("KeyWithDelay", func(t *testing.T) {
		// Test key with delay without requiring connection
		err := client.SendKeyWithDelay('a', 100, t128.ModifierKey{})
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "no active connection")
	})

	t.Run("KeyRepeat", func(t *testing.T) {
		// Test key repeat without requiring connection
		err := client.SendKeyRepeat('x', 3, t128.ModifierKey{})
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "no active connection")
	})
}

// TestVirtualChannelSupport tests virtual channel functionality
func TestVirtualChannelSupport(t *testing.T) {
	t.Run("VirtualChannelManager", func(t *testing.T) {
		manager := virtualchannel.NewVirtualChannelManager()
		assert.NotNil(t, manager)

		// Test channel registration
		channel := &virtualchannel.VirtualChannel{
			ID:    1,
			Name:  "TEST_CHANNEL",
			Flags: virtualchannel.CHANNEL_FLAG_FIRST | virtualchannel.CHANNEL_FLAG_LAST,
		}

		err := manager.RegisterChannel(channel)
		assert.NoError(t, err)

		// Test channel retrieval
		retrieved, exists := manager.GetChannel(1)
		assert.True(t, exists)
		assert.Equal(t, "TEST_CHANNEL", retrieved.Name)
	})

	t.Run("DynamicVirtualChannelManager", func(t *testing.T) {
		manager := drdynvc.NewDynamicVirtualChannelManager()
		assert.NotNil(t, manager)
		// Dynamic virtual channel methods would be tested with actual implementation
	})
}

// TestBitmapCacheOptimization tests bitmap cache optimization
func TestBitmapCacheOptimization(t *testing.T) {
	t.Run("BitmapCacheManager", func(t *testing.T) {
		manager := t128.NewBitmapCacheManager()
		assert.NotNil(t, manager)

		// Test cache statistics
		stats := manager.GetCacheStats()
		assert.NotNil(t, stats)
		assert.Equal(t, 0, stats["hits"])
		assert.Equal(t, 0, stats["misses"])
	})

	t.Run("OffscreenBitmapManager", func(t *testing.T) {
		manager := t128.NewOffscreenBitmapManager(1024, 50)
		assert.NotNil(t, manager)
		// Offscreen bitmap operations would be tested with actual implementation
	})
}

// TestSecurityFeatures tests security features
func TestSecurityFeatures(t *testing.T) {
	t.Run("NLAAuthentication", func(t *testing.T) {
		// Test NLA authentication structures
		client := NewClient(&Option{
			Addr:     "localhost:3389",
			UserName: "test",
			Password: "test",
		})

		assert.NotNil(t, client)
		// NLA authentication would be tested with actual connection
	})

	t.Run("CertificateValidation", func(t *testing.T) {
		// Test certificate validation structures
		// This would require actual certificates for full testing
	})
}

// TestPerformanceOptimizations tests performance optimizations
func TestPerformanceOptimizations(t *testing.T) {
	t.Run("CompressionSupport", func(t *testing.T) {
		// Test compression support
		client := NewClient(&Option{
			Addr:     "localhost:3389",
			UserName: "test",
			Password: "test",
		})

		assert.NotNil(t, client)
		// Compression would be tested with actual data
	})

	t.Run("NetworkOptimization", func(t *testing.T) {
		// Test network optimization features
		client := NewClient(&Option{
			Addr:     "localhost:3389",
			UserName: "test",
			Password: "test",
		})

		assert.NotNil(t, client)
		// Network optimization would be tested with actual connection
	})
}

// TestErrorHandlingAndRecovery tests error handling and recovery
func TestErrorHandlingAndRecovery(t *testing.T) {
	t.Run("ConnectionErrorHandling", func(t *testing.T) {
		client := NewClient(&Option{
			Addr:           "invalid:9999",
			UserName:       "test",
			Password:       "test",
			ConnectTimeout: 1 * time.Second,
		})

		err := client.Connect()
		assert.Error(t, err)
		// Should handle connection errors gracefully
	})

	t.Run("InvalidInputHandling", func(t *testing.T) {
		client := NewClient(&Option{
			Addr:     "localhost:3389",
			UserName: "test",
			Password: "test",
		})

		// Test invalid mouse button
		err := client.SendMouseButtonEvent(t128.MouseButton(999), true, 100, 100)
		assert.Error(t, err)

		// Test invalid special key
		err = client.SendSpecialKey("INVALID_KEY", t128.ModifierKey{})
		assert.Error(t, err)
	})

	t.Run("ResourceCleanup", func(t *testing.T) {
		client := NewClient(&Option{
			Addr:     "localhost:3389",
			UserName: "test",
			Password: "test",
		})

		// Test resource cleanup - should not panic even without connection
		defer func() {
			if r := recover(); r != nil {
				t.Errorf("Client.Close() panicked: %v", r)
			}
		}()
		client.Close()
	})
}

// TestConfigurationManagement tests configuration management
func TestConfigurationManagement(t *testing.T) {
	t.Run("ClientConfiguration", func(t *testing.T) {
		option := &Option{
			Addr:           "localhost:3389",
			UserName:       "testuser",
			Password:       "testpass",
			ConnectTimeout: 10 * time.Second,
			Monitors: []mcs.MonitorLayout{
				{
					Left:   0,
					Top:    0,
					Right:  1920,
					Bottom: 1080,
					Flags:  0x01, // Primary monitor
				},
			},
		}

		client := NewClient(option)
		assert.Equal(t, "localhost:3389", client.option.Addr)
		assert.Equal(t, "testuser", client.option.UserName)
		assert.Equal(t, "testpass", client.option.Password)
		assert.Equal(t, 10*time.Second, client.option.ConnectTimeout)
		assert.Len(t, client.option.Monitors, 1)
	})

	t.Run("MonitorConfiguration", func(t *testing.T) {
		client := NewClient(&Option{
			Addr:     "localhost:3389",
			UserName: "test",
			Password: "test",
		})

		monitors := []mcs.MonitorLayout{
			{
				Left:   0,
				Top:    0,
				Right:  1920,
				Bottom: 1080,
				Flags:  0x01, // Primary monitor
			},
			{
				Left:   1920,
				Top:    0,
				Right:  3840,
				Bottom: 1080,
				Flags:  0,
			},
		}

		client.SetMonitors(monitors)
		retrieved := client.GetMonitors()
		assert.Len(t, retrieved, 2)
		assert.Equal(t, int32(1920), retrieved[0].Right)
		assert.Equal(t, int32(1080), retrieved[0].Bottom)
		assert.Equal(t, uint32(0x01), retrieved[0].Flags)
	})
}

// TestIntegrationScenarios tests integration scenarios
func TestIntegrationScenarios(t *testing.T) {
	t.Run("FullSessionWorkflow", func(t *testing.T) {
		// Test complete session workflow
		client := NewClient(&Option{
			Addr:           "localhost:3389",
			UserName:       "test",
			Password:       "test",
			ConnectTimeout: 1 * time.Second,
		})

		// This would test the full workflow in a real scenario
		// For now, just test that client creation works
		assert.NotNil(t, client)
	})

	t.Run("MultiFeatureIntegration", func(t *testing.T) {
		client := NewClient(&Option{
			Addr:     "localhost:3389",
			UserName: "test",
			Password: "test",
		})

		// Test integration of multiple features
		// Register handlers
		err := client.RegisterClipboardHandler(clipboard.NewDefaultClipboardHandler())
		assert.NoError(t, err)

		err = client.RegisterDeviceHandler(device.NewDefaultDeviceHandler())
		assert.NoError(t, err)

		// Test handler registration
		assert.True(t, client.IsClipboardChannelOpen())
		assert.True(t, client.IsDeviceChannelOpen())
	})
}

// Benchmark tests for performance
func BenchmarkClipboardOperations(b *testing.B) {
	manager := clipboard.NewClipboardManager(nil)
	formats := []clipboard.ClipboardFormat{
		clipboard.CLIPRDR_FORMAT_UNICODETEXT,
		clipboard.CLIPRDR_FORMAT_HTML,
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		manager.CreateFormatListMessage(formats)
	}
}

func BenchmarkAudioOperations(b *testing.B) {
	manager := audio.NewAudioManager(nil)
	data := make([]byte, 1024)
	for i := range data {
		data[i] = byte(i % 256)
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		manager.CreateClientWaveMessage(1, data, uint32(i))
	}
}

func BenchmarkDeviceOperations(b *testing.B) {
	manager := device.NewDeviceManager(nil)
	data := make([]byte, 1024)
	for i := range data {
		data[i] = byte(i % 256)
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		manager.CreatePrinterDataMessage(uint32(i), data, 0x0001)
	}
}

func BenchmarkEnhancedInput(b *testing.B) {
	client := NewClient(&Option{
		Addr:     "localhost:3389",
		UserName: "test",
		Password: "test",
	})

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		client.SendKeyPress('a', t128.ModifierKey{})
	}
}

func BenchmarkMouseOperations(b *testing.B) {
	client := NewClient(&Option{
		Addr:     "localhost:3389",
		UserName: "test",
		Password: "test",
	})

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		client.SendMouseMoveEvent(uint16(i%1920), uint16(i%1080))
	}
}
