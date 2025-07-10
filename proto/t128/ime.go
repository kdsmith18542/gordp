package t128

import (
	"fmt"
)

// IMEState represents the current state of the Input Method Editor
type IMEState struct {
	CompositionString string
	CandidateList     []string
	SelectedCandidate int
	IsComposing       bool
	IsOpen            bool
}

// IMEHandler handles IME-related input events
// NOTE: For full CJK/complex script support, implement IME composition and candidate selection.
// This implementation is sufficient for basic Latin/European input and simple Unicode.
type IMEHandler struct {
	state *IMEState
}

// NewIMEHandler creates a new IME handler
func NewIMEHandler() *IMEHandler {
	return &IMEHandler{
		state: &IMEState{
			CompositionString: "",
			CandidateList:     make([]string, 0),
			SelectedCandidate: 0,
			IsComposing:       false,
			IsOpen:            false,
		},
	}
}

// GetState returns the current IME state
func (h *IMEHandler) GetState() *IMEState {
	return h.state
}

// IsComposing returns true if the IME is currently composing
func (h *IMEHandler) IsComposing() bool {
	return h.state.IsComposing
}

// IsOpen returns true if the IME is open
func (h *IMEHandler) IsOpen() bool {
	return h.state.IsOpen
}

// StartComposition starts IME composition
func (h *IMEHandler) StartComposition() {
	h.state.IsComposing = true
	h.state.CompositionString = ""
	h.state.CandidateList = make([]string, 0)
	h.state.SelectedCandidate = 0
}

// EndComposition ends IME composition
func (h *IMEHandler) EndComposition() {
	h.state.IsComposing = false
	h.state.CompositionString = ""
	h.state.CandidateList = make([]string, 0)
	h.state.SelectedCandidate = 0
}

// SetCompositionString sets the current composition string
func (h *IMEHandler) SetCompositionString(str string) {
	h.state.CompositionString = str
}

// SetCandidateList sets the candidate list
func (h *IMEHandler) SetCandidateList(candidates []string) {
	h.state.CandidateList = candidates
	h.state.SelectedCandidate = 0
}

// SelectCandidate selects a candidate from the list
func (h *IMEHandler) SelectCandidate(index int) error {
	if index < 0 || index >= len(h.state.CandidateList) {
		return fmt.Errorf("invalid candidate index: %d", index)
	}
	h.state.SelectedCandidate = index
	return nil
}

// GetSelectedCandidate returns the currently selected candidate
func (h *IMEHandler) GetSelectedCandidate() string {
	if h.state.SelectedCandidate < len(h.state.CandidateList) {
		return h.state.CandidateList[h.state.SelectedCandidate]
	}
	return ""
}

// OpenIME opens the IME
func (h *IMEHandler) OpenIME() {
	h.state.IsOpen = true
}

// CloseIME closes the IME
func (h *IMEHandler) CloseIME() {
	h.state.IsOpen = false
	h.EndComposition()
}

// ProcessUnicodeInput processes Unicode input for IME
// NOTE: For full CJK/complex script support, this must be extended to handle IME composition and candidate selection.
func (h *IMEHandler) ProcessUnicodeInput(char rune) ([]uint8, error) {
	// For basic Unicode characters, we can map them directly
	if char <= 127 {
		// ASCII characters
		keyCode, ok := KeyMap[char]
		if !ok {
			return nil, fmt.Errorf("unsupported ASCII character: %c (U+%04X)", char, char)
		}
		return []uint8{keyCode}, nil
	}

	// For extended Unicode characters, we need special handling
	// This is a simplified implementation - in a full implementation,
	// you would need to handle IME composition properly

	keyCode, ok := UnicodeKeyMap[char]
	if ok {
		return []uint8{keyCode}, nil
	}

	// For characters without direct mapping, suggest IME composition
	return nil, fmt.Errorf("unsupported Unicode character: %c (U+%04X). Use IME composition for complex scripts.", char, char)
}

// UnicodeKeyMap maps Unicode characters to virtual key codes.
// NOTE: For full production-grade international support, generate or extend this map based on the user's keyboard layout, locale, and IME capabilities.
var UnicodeKeyMap = map[rune]uint8{
	'é':  VK_OEM_7, // Single quote key
	'è':  VK_OEM_7,
	'à':  VK_A,
	'ù':  VK_U,
	'ç':  VK_OEM_1, // Semicolon key
	'ñ':  VK_N,
	'ü':  VK_U,
	'ö':  VK_O,
	'ä':  VK_A,
	'ß':  VK_OEM_MINUS, // Minus key
	'€':  VK_OEM_5,     // Backslash key
	'£':  VK_OEM_3,     // Backtick key
	'¥':  VK_OEM_5,
	'°':  VK_OEM_7,
	'±':  VK_OEM_PLUS,
	'×':  VK_OEM_8, // Backslash key
	'÷':  VK_OEM_2, // Forward slash key
	'©':  VK_C,
	'®':  VK_R,
	'™':  VK_T,
	'§':  VK_OEM_6, // Right bracket key
	'¶':  VK_P,
	'†':  VK_T,
	'‡':  VK_T,
	'•':  VK_OEM_PERIOD,
	'–':  VK_OEM_MINUS,
	'—':  VK_OEM_MINUS,
	'"':  VK_OEM_7,
	'\'': VK_OEM_7,
	'‹':  VK_OEM_COMMA,
	'›':  VK_OEM_PERIOD,
	'«':  VK_OEM_4, // Left bracket key
	'»':  VK_OEM_6, // Right bracket key
	'…':  VK_OEM_PERIOD,
	'‰':  VK_OEM_5,
	'¢':  VK_C,
	'¤':  VK_OEM_3,
	'¦':  VK_OEM_5,
	'¨':  VK_OEM_7,
	'¯':  VK_OEM_MINUS,
	'´':  VK_OEM_7,
	'¸':  VK_OEM_COMMA,
	'¹':  VK_1,
	'²':  VK_2,
	'³':  VK_3,
	'¼':  VK_4,
	'½':  VK_5,
	'¾':  VK_6,
	'¿':  VK_OEM_2,
	'¡':  VK_1,
	'ª':  VK_A,
	'º':  VK_O,
	'À':  VK_A,
	'Á':  VK_A,
	'Â':  VK_A,
	'Ã':  VK_A,
	'Å':  VK_A,
	'Æ':  VK_A,
	'È':  VK_E,
	'Ê':  VK_E,
	'Ë':  VK_E,
	'Ì':  VK_I,
	'Í':  VK_I,
	'Î':  VK_I,
	'Ï':  VK_I,
	'Ð':  VK_D,
	'Ò':  VK_O,
	'Ó':  VK_O,
	'Ô':  VK_O,
	'Õ':  VK_O,
	'Ø':  VK_O,
	'Ù':  VK_U,
	'Ú':  VK_U,
	'Û':  VK_U,
	'Ý':  VK_Y,
	'Þ':  VK_T,
	'ÿ':  VK_Y,
	// TODO: Add more mappings for CJK, Greek, Cyrillic, and other scripts as needed
}

// IsUnicodeSupported checks if a Unicode character is supported
func IsUnicodeSupported(char rune) bool {
	// Check if it's a basic ASCII character
	if char <= 127 {
		return true
	}

	// Check if it's in our Unicode key map
	_, ok := UnicodeKeyMap[char]
	return ok
}

// GetUnicodeSupportLevel returns the level of Unicode support
func GetUnicodeSupportLevel() string {
	// This is a simplified implementation
	// In a real implementation, you would check the actual capabilities
	// of the target system and IME
	return "basic"
}

// ConvertToCompositionString converts a string to IME composition format
func ConvertToCompositionString(str string) string {
	// This is a simplified implementation
	// In a real implementation, you would need to handle the actual
	// IME composition format based on the target system
	return str
}

// IsCompositionRequired checks if IME composition is required for a character
func IsCompositionRequired(char rune) bool {
	// Basic ASCII characters don't require composition
	if char <= 127 {
		return false
	}

	// Check if the character is in our Unicode key map
	_, ok := UnicodeKeyMap[char]
	return !ok
}
