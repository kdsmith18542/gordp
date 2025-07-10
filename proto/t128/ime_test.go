package t128

import (
	"testing"
)

func TestNewIMEHandler(t *testing.T) {
	handler := NewIMEHandler()
	if handler == nil {
		t.Error("expected IME handler to be created")
	}

	state := handler.GetState()
	if state == nil {
		t.Error("expected IME state to be created")
	}

	if state.IsComposing {
		t.Error("expected IME to not be composing initially")
	}

	if state.IsOpen {
		t.Error("expected IME to not be open initially")
	}

	if len(state.CandidateList) != 0 {
		t.Error("expected candidate list to be empty initially")
	}

	if state.SelectedCandidate != 0 {
		t.Error("expected selected candidate to be 0 initially")
	}
}

func TestIMEStateManagement(t *testing.T) {
	handler := NewIMEHandler()

	// Test opening IME
	handler.OpenIME()
	if !handler.IsOpen() {
		t.Error("expected IME to be open")
	}

	// Test starting composition
	handler.StartComposition()
	if !handler.IsComposing() {
		t.Error("expected IME to be composing")
	}

	// Test setting composition string
	handler.SetCompositionString("test")
	state := handler.GetState()
	if state.CompositionString != "test" {
		t.Errorf("expected composition string 'test', got '%s'", state.CompositionString)
	}

	// Test setting candidate list
	candidates := []string{"test", "testing", "tested"}
	handler.SetCandidateList(candidates)
	state = handler.GetState()
	if len(state.CandidateList) != 3 {
		t.Errorf("expected 3 candidates, got %d", len(state.CandidateList))
	}

	// Test selecting candidate
	err := handler.SelectCandidate(1)
	if err != nil {
		t.Errorf("failed to select candidate: %v", err)
	}
	state = handler.GetState()
	if state.SelectedCandidate != 1 {
		t.Errorf("expected selected candidate 1, got %d", state.SelectedCandidate)
	}

	// Test getting selected candidate
	selected := handler.GetSelectedCandidate()
	if selected != "testing" {
		t.Errorf("expected selected candidate 'testing', got '%s'", selected)
	}

	// Test ending composition
	handler.EndComposition()
	if handler.IsComposing() {
		t.Error("expected IME to not be composing after end")
	}
	state = handler.GetState()
	if state.CompositionString != "" {
		t.Error("expected composition string to be empty after end")
	}
	if len(state.CandidateList) != 0 {
		t.Error("expected candidate list to be empty after end")
	}

	// Test closing IME
	handler.CloseIME()
	if handler.IsOpen() {
		t.Error("expected IME to be closed")
	}
	if handler.IsComposing() {
		t.Error("expected IME to not be composing after close")
	}
}

func TestSelectCandidateErrors(t *testing.T) {
	handler := NewIMEHandler()

	// Test selecting candidate with empty list
	err := handler.SelectCandidate(0)
	if err == nil {
		t.Error("expected error when selecting candidate from empty list")
	}

	// Test selecting candidate with invalid index
	handler.SetCandidateList([]string{"test"})
	err = handler.SelectCandidate(1)
	if err == nil {
		t.Error("expected error when selecting invalid candidate index")
	}

	err = handler.SelectCandidate(-1)
	if err == nil {
		t.Error("expected error when selecting negative candidate index")
	}
}

func TestProcessUnicodeInput(t *testing.T) {
	handler := NewIMEHandler()

	// Test ASCII character
	keyCodes, err := handler.ProcessUnicodeInput('a')
	if err != nil {
		t.Errorf("failed to process ASCII character: %v", err)
	}
	if len(keyCodes) != 1 {
		t.Errorf("expected 1 key code, got %d", len(keyCodes))
	}
	if keyCodes[0] != VK_A {
		t.Errorf("expected key code VK_A, got %d", keyCodes[0])
	}

	// Test supported Unicode character
	keyCodes, err = handler.ProcessUnicodeInput('é')
	if err != nil {
		t.Errorf("failed to process supported Unicode character: %v", err)
	}
	if len(keyCodes) != 1 {
		t.Errorf("expected 1 key code, got %d", len(keyCodes))
	}
	if keyCodes[0] != VK_OEM_7 {
		t.Errorf("expected key code VK_OEM_7, got %d", keyCodes[0])
	}

	// Test unsupported Unicode character
	keyCodes, err = handler.ProcessUnicodeInput('🚀')
	if err == nil {
		t.Error("expected error for unsupported Unicode character")
	}
}

func TestIsUnicodeSupported(t *testing.T) {
	// Test ASCII characters
	if !IsUnicodeSupported('a') {
		t.Error("expected ASCII character 'a' to be supported")
	}
	if !IsUnicodeSupported('Z') {
		t.Error("expected ASCII character 'Z' to be supported")
	}
	if !IsUnicodeSupported('1') {
		t.Error("expected ASCII character '1' to be supported")
	}

	// Test supported Unicode characters
	if !IsUnicodeSupported('é') {
		t.Error("expected Unicode character 'é' to be supported")
	}
	if !IsUnicodeSupported('ñ') {
		t.Error("expected Unicode character 'ñ' to be supported")
	}
	if !IsUnicodeSupported('€') {
		t.Error("expected Unicode character '€' to be supported")
	}

	// Test unsupported Unicode characters
	if IsUnicodeSupported('🚀') {
		t.Error("expected Unicode character '🚀' to not be supported")
	}
	if IsUnicodeSupported('🎉') {
		t.Error("expected Unicode character '🎉' to not be supported")
	}
}

func TestGetUnicodeSupportLevel(t *testing.T) {
	level := GetUnicodeSupportLevel()
	if level != "basic" {
		t.Errorf("expected support level 'basic', got '%s'", level)
	}
}

func TestConvertToCompositionString(t *testing.T) {
	// Test basic string conversion
	result := ConvertToCompositionString("test")
	if result != "test" {
		t.Errorf("expected 'test', got '%s'", result)
	}

	// Test string with Unicode characters
	result = ConvertToCompositionString("café")
	if result != "café" {
		t.Errorf("expected 'café', got '%s'", result)
	}
}

func TestIsCompositionRequired(t *testing.T) {
	// Test ASCII characters (should not require composition)
	if IsCompositionRequired('a') {
		t.Error("expected ASCII character 'a' to not require composition")
	}
	if IsCompositionRequired('Z') {
		t.Error("expected ASCII character 'Z' to not require composition")
	}
	if IsCompositionRequired('1') {
		t.Error("expected ASCII character '1' to not require composition")
	}

	// Test supported Unicode characters (should not require composition)
	if IsCompositionRequired('é') {
		t.Error("expected supported Unicode character 'é' to not require composition")
	}
	if IsCompositionRequired('ñ') {
		t.Error("expected supported Unicode character 'ñ' to not require composition")
	}

	// Test unsupported Unicode characters (should require composition)
	if !IsCompositionRequired('🚀') {
		t.Error("expected unsupported Unicode character '🚀' to require composition")
	}
	if !IsCompositionRequired('🎉') {
		t.Error("expected unsupported Unicode character '🎉' to require composition")
	}
}

func TestUnicodeKeyMap(t *testing.T) {
	// Test some key mappings
	testCases := []struct {
		char     rune
		expected uint8
	}{
		{'é', VK_OEM_7},
		{'ñ', VK_N},
		{'€', VK_OEM_5},
		{'£', VK_OEM_3},
		{'©', VK_C},
		{'®', VK_R},
		{'™', VK_T},
		{'°', VK_OEM_7},
		{'±', VK_OEM_PLUS},
		{'×', VK_OEM_8},
		{'÷', VK_OEM_2},
	}

	for _, tc := range testCases {
		if keyCode, exists := UnicodeKeyMap[tc.char]; exists {
			if keyCode != tc.expected {
				t.Errorf("expected key code %d for character '%c', got %d", tc.expected, tc.char, keyCode)
			}
		} else {
			t.Errorf("character '%c' not found in UnicodeKeyMap", tc.char)
		}
	}
}

func TestIMEHandlerWithContext(t *testing.T) {
	handler := NewIMEHandler()

	// Simulate a typical IME workflow
	handler.OpenIME()
	handler.StartComposition()
	handler.SetCompositionString("caf")

	// Set candidate list
	candidates := []string{"café", "cafe", "cafeteria"}
	handler.SetCandidateList(candidates)

	// Select first candidate
	err := handler.SelectCandidate(0)
	if err != nil {
		t.Errorf("failed to select candidate: %v", err)
	}

	// Verify state
	state := handler.GetState()
	if !state.IsOpen {
		t.Error("expected IME to be open")
	}
	if !state.IsComposing {
		t.Error("expected IME to be composing")
	}
	if state.CompositionString != "caf" {
		t.Errorf("expected composition string 'caf', got '%s'", state.CompositionString)
	}
	if len(state.CandidateList) != 3 {
		t.Errorf("expected 3 candidates, got %d", len(state.CandidateList))
	}
	if state.SelectedCandidate != 0 {
		t.Errorf("expected selected candidate 0, got %d", state.SelectedCandidate)
	}

	// End composition
	handler.EndComposition()
	if handler.IsComposing() {
		t.Error("expected IME to not be composing after end")
	}

	// Close IME
	handler.CloseIME()
	if handler.IsOpen() {
		t.Error("expected IME to be closed")
	}
}
