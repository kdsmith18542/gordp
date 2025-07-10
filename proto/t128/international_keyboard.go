// International Keyboard Layout Support for GoRDP
// Provides comprehensive keyboard layout support for international languages and scripts
// including CJK (Chinese, Japanese, Korean), Greek, Cyrillic, Arabic, Hebrew, Thai, and European layouts

package t128

import (
	"fmt"
	"strings"
	"unicode"
)

// KeyboardLayout represents a specific keyboard layout
type KeyboardLayout struct {
	ID          string            // Layout identifier (e.g., "en_US", "ja_JP", "ru_RU")
	Name        string            // Human-readable name
	Language    string            // Language name
	Script      string            // Writing system
	KeyMap      map[rune]uint8    // Character to virtual key code mapping
	ShiftMap    map[rune]rune     // Shift key combinations
	AltGrMap    map[rune]rune     // AltGr key combinations
	DeadKeys    map[rune]rune     // Dead key combinations
	ComposeKeys map[string]rune   // Compose key sequences
	IME         bool              // Whether this layout requires IME
	RTL         bool              // Right-to-left script
}

// InternationalKeyboardManager manages multiple keyboard layouts
type InternationalKeyboardManager struct {
	layouts     map[string]*KeyboardLayout
	current     *KeyboardLayout
	defaultLang string
	imeEnabled  bool
	composeBuffer []rune
}

// NewInternationalKeyboardManager creates a new keyboard layout manager
func NewInternationalKeyboardManager() *InternationalKeyboardManager {
	manager := &InternationalKeyboardManager{
		layouts:     make(map[string]*KeyboardLayout),
		defaultLang: "en_US",
		imeEnabled:  false,
		composeBuffer: make([]rune, 0, 10),
	}
	
	// Initialize all supported layouts
	manager.initializeLayouts()
	
	// Set default layout
	manager.SetLayout("en_US")
	
	return manager
}

// initializeLayouts sets up all supported keyboard layouts
func (m *InternationalKeyboardManager) initializeLayouts() {
	// English layouts
	m.addLayout(createEnglishUSLayout())
	m.addLayout(createEnglishUKLayout())
	
	// European layouts
	m.addLayout(createGermanLayout())
	m.addLayout(createFrenchLayout())
	m.addLayout(createSpanishLayout())
	m.addLayout(createItalianLayout())
	m.addLayout(createPortugueseLayout())
	m.addLayout(createDutchLayout())
	m.addLayout(createSwedishLayout())
	m.addLayout(createNorwegianLayout())
	m.addLayout(createDanishLayout())
	m.addLayout(createFinnishLayout())
	
	// Greek layout
	m.addLayout(createGreekLayout())
	
	// Cyrillic layouts
	m.addLayout(createRussianLayout())
	m.addLayout(createUkrainianLayout())
	m.addLayout(createBulgarianLayout())
	m.addLayout(createSerbianLayout())
	
	// CJK layouts
	m.addLayout(createJapaneseLayout())
	m.addLayout(createChineseSimplifiedLayout())
	m.addLayout(createChineseTraditionalLayout())
	m.addLayout(createKoreanLayout())
	
	// Middle Eastern layouts
	m.addLayout(createArabicLayout())
	m.addLayout(createHebrewLayout())
	m.addLayout(createPersianLayout())
	m.addLayout(createTurkishLayout())
	
	// South Asian layouts
	m.addLayout(createHindiLayout())
	m.addLayout(createThaiLayout())
	m.addLayout(createVietnameseLayout())
}

// addLayout adds a keyboard layout to the manager
func (m *InternationalKeyboardManager) addLayout(layout *KeyboardLayout) {
	m.layouts[layout.ID] = layout
}

// SetLayout changes the current keyboard layout
func (m *InternationalKeyboardManager) SetLayout(layoutID string) error {
	layout, exists := m.layouts[layoutID]
	if !exists {
		return fmt.Errorf("keyboard layout '%s' not found", layoutID)
	}
	
	m.current = layout
	m.imeEnabled = layout.IME
	
	// Clear compose buffer when changing layouts
	m.composeBuffer = m.composeBuffer[:0]
	
	return nil
}

// GetCurrentLayout returns the current keyboard layout
func (m *InternationalKeyboardManager) GetCurrentLayout() *KeyboardLayout {
	return m.current
}

// GetAvailableLayouts returns all available keyboard layouts
func (m *InternationalKeyboardManager) GetAvailableLayouts() []*KeyboardLayout {
	layouts := make([]*KeyboardLayout, 0, len(m.layouts))
	for _, layout := range m.layouts {
		layouts = append(layouts, layout)
	}
	return layouts
}

// ConvertCharToKeyCode converts a character to virtual key code with current layout
func (m *InternationalKeyboardManager) ConvertCharToKeyCode(char rune, shift, altGr bool) (uint8, bool) {
	if m.current == nil {
		return 0, false
	}
	
	// Handle compose sequences
	if char == 0x1B { // Escape key
		m.composeBuffer = m.composeBuffer[:0]
		return VK_ESCAPE, true
	}
	
	// Check for compose sequences
	if len(m.composeBuffer) > 0 {
		m.composeBuffer = append(m.composeBuffer, char)
		composeKey := string(m.composeBuffer)
		
		if composed, exists := m.current.ComposeKeys[composeKey]; exists {
			m.composeBuffer = m.composeBuffer[:0]
			return m.ConvertCharToKeyCode(composed, shift, altGr)
		}
		
		// Check if this could be part of a longer compose sequence
		for key := range m.current.ComposeKeys {
			if strings.HasPrefix(key, composeKey) {
				return 0, false // Wait for more input
			}
		}
		
		// Not a valid compose sequence, clear buffer and process normally
		m.composeBuffer = m.composeBuffer[:0]
	}
	
	// Handle AltGr combinations
	if altGr {
		if altGrChar, exists := m.current.AltGrMap[char]; exists {
			char = altGrChar
		}
	}
	
	// Handle Shift combinations
	if shift {
		if shiftChar, exists := m.current.ShiftMap[char]; exists {
			char = shiftChar
		}
	}
	
	// Convert to virtual key code
	if keyCode, exists := m.current.KeyMap[char]; exists {
		return keyCode, true
	}
	
	// Fallback to Unicode event for unsupported characters
	return 0, false
}

// CreateUnicodeEvent creates a Unicode input event for characters not in the key map
func (m *InternationalKeyboardManager) CreateUnicodeEvent(char rune) *TsFpUnicodeEvent {
	return &TsFpUnicodeEvent{
		EventHeader: 1, // Key down
		UnicodeCode: uint16(char),
	}
}

// IsIMERequired returns whether the current layout requires IME
func (m *InternationalKeyboardManager) IsIMERequired() bool {
	return m.current != nil && m.current.IME
}

// IsRTL returns whether the current layout is right-to-left
func (m *InternationalKeyboardManager) IsRTL() bool {
	return m.current != nil && m.current.RTL
}

// ============================================================================
// Layout Creation Functions
// ============================================================================

func createEnglishUSLayout() *KeyboardLayout {
	return &KeyboardLayout{
		ID:       "en_US",
		Name:     "English (US)",
		Language: "English",
		Script:   "Latin",
		KeyMap: map[rune]uint8{
			'a': VK_A, 'b': VK_B, 'c': VK_C, 'd': VK_D, 'e': VK_E,
			'f': VK_F, 'g': VK_G, 'h': VK_H, 'i': VK_I, 'j': VK_J,
			'k': VK_K, 'l': VK_L, 'm': VK_M, 'n': VK_N, 'o': VK_O,
			'p': VK_P, 'q': VK_Q, 'r': VK_R, 's': VK_S, 't': VK_T,
			'u': VK_U, 'v': VK_V, 'w': VK_W, 'x': VK_X, 'y': VK_Y, 'z': VK_Z,
			'0': VK_0, '1': VK_1, '2': VK_2, '3': VK_3, '4': VK_4,
			'5': VK_5, '6': VK_6, '7': VK_7, '8': VK_8, '9': VK_9,
			' ': VK_SPACE, '\t': VK_TAB, '\n': VK_RETURN, '\r': VK_RETURN,
			'`': VK_OEM_3, '-': VK_OEM_MINUS, '=': VK_OEM_PLUS,
			'[': VK_OEM_4, ']': VK_OEM_6, '\\': VK_OEM_5,
			';': VK_OEM_1, '\'': VK_OEM_7, ',': VK_OEM_COMMA,
			'.': VK_OEM_PERIOD, '/': VK_OEM_2,
		},
		ShiftMap: map[rune]rune{
			'a': 'A', 'b': 'B', 'c': 'C', 'd': 'D', 'e': 'E',
			'f': 'F', 'g': 'G', 'h': 'H', 'i': 'I', 'j': 'J',
			'k': 'K', 'l': 'L', 'm': 'M', 'n': 'N', 'o': 'O',
			'p': 'P', 'q': 'Q', 'r': 'R', 's': 'S', 't': 'T',
			'u': 'U', 'v': 'V', 'w': 'W', 'x': 'X', 'y': 'Y', 'z': 'Z',
			'1': '!', '2': '@', '3': '#', '4': '$', '5': '%',
			'6': '^', '7': '&', '8': '*', '9': '(', '0': ')',
			'`': '~', '-': '_', '=': '+', '[': '{', ']': '}',
			'\\': '|', ';': ':', '\'': '"', ',': '<', '.': '>', '/': '?',
		},
		IME: false,
		RTL: false,
	}
}

func createGreekLayout() *KeyboardLayout {
	return &KeyboardLayout{
		ID:       "el_GR",
		Name:     "Greek",
		Language: "Greek",
		Script:   "Greek",
		KeyMap: map[rune]uint8{
			// Greek letters
			'α': VK_A, 'β': VK_B, 'γ': VK_C, 'δ': VK_D, 'ε': VK_E,
			'ζ': VK_F, 'η': VK_G, 'θ': VK_H, 'ι': VK_I, 'κ': VK_J,
			'λ': VK_K, 'μ': VK_L, 'ν': VK_M, 'ξ': VK_N, 'ο': VK_O,
			'π': VK_P, 'ρ': VK_Q, 'σ': VK_R, 'τ': VK_S, 'υ': VK_T,
			'φ': VK_U, 'χ': VK_V, 'ψ': VK_W, 'ω': VK_X, 'ς': VK_Y,
			// Numbers and symbols
			'0': VK_0, '1': VK_1, '2': VK_2, '3': VK_3, '4': VK_4,
			'5': VK_5, '6': VK_6, '7': VK_7, '8': VK_8, '9': VK_9,
			' ': VK_SPACE, '\t': VK_TAB, '\n': VK_RETURN,
		},
		ShiftMap: map[rune]rune{
			// Capital Greek letters
			'α': 'Α', 'β': 'Β', 'γ': 'Γ', 'δ': 'Δ', 'ε': 'Ε',
			'ζ': 'Ζ', 'η': 'Η', 'θ': 'Θ', 'ι': 'Ι', 'κ': 'Κ',
			'λ': 'Λ', 'μ': 'Μ', 'ν': 'Ν', 'ξ': 'Ξ', 'ο': 'Ο',
			'π': 'Π', 'ρ': 'Ρ', 'σ': 'Σ', 'τ': 'Τ', 'υ': 'Υ',
			'φ': 'Φ', 'χ': 'Χ', 'ψ': 'Ψ', 'ω': 'Ω',
		},
		IME: false,
		RTL: false,
	}
}

func createRussianLayout() *KeyboardLayout {
	return &KeyboardLayout{
		ID:       "ru_RU",
		Name:     "Russian",
		Language: "Russian",
		Script:   "Cyrillic",
		KeyMap: map[rune]uint8{
			// Cyrillic letters
			'а': VK_A, 'б': VK_B, 'в': VK_C, 'г': VK_D, 'д': VK_E,
			'е': VK_F, 'ё': VK_G, 'ж': VK_H, 'з': VK_I, 'и': VK_J,
			'й': VK_K, 'к': VK_L, 'л': VK_M, 'м': VK_N, 'н': VK_O,
			'о': VK_P, 'п': VK_Q, 'р': VK_R, 'с': VK_S, 'т': VK_T,
			'у': VK_U, 'ф': VK_V, 'х': VK_W, 'ц': VK_X, 'ч': VK_Y,
			'ш': VK_Z, 'щ': VK_OEM_1, 'ъ': VK_OEM_2, 'ы': VK_OEM_3,
			'ь': VK_OEM_4, 'э': VK_OEM_5, 'ю': VK_OEM_6, 'я': VK_OEM_7,
			// Numbers and symbols
			'0': VK_0, '1': VK_1, '2': VK_2, '3': VK_3, '4': VK_4,
			'5': VK_5, '6': VK_6, '7': VK_7, '8': VK_8, '9': VK_9,
			' ': VK_SPACE, '\t': VK_TAB, '\n': VK_RETURN,
		},
		ShiftMap: map[rune]rune{
			// Capital Cyrillic letters
			'а': 'А', 'б': 'Б', 'в': 'В', 'г': 'Г', 'д': 'Д',
			'е': 'Е', 'ё': 'Ё', 'ж': 'Ж', 'з': 'З', 'и': 'И',
			'й': 'Й', 'к': 'К', 'л': 'Л', 'м': 'М', 'н': 'Н',
			'о': 'О', 'п': 'П', 'р': 'Р', 'с': 'С', 'т': 'Т',
			'у': 'У', 'ф': 'Ф', 'х': 'Х', 'ц': 'Ц', 'ч': 'Ч',
			'ш': 'Ш', 'щ': 'Щ', 'ъ': 'Ъ', 'ы': 'Ы',
			'ь': 'Ь', 'э': 'Э', 'ю': 'Ю', 'я': 'Я',
		},
		IME: false,
		RTL: false,
	}
}

func createJapaneseLayout() *KeyboardLayout {
	return &KeyboardLayout{
		ID:       "ja_JP",
		Name:     "Japanese",
		Language: "Japanese",
		Script:   "Hiragana/Katakana/Kanji",
		KeyMap: map[rune]uint8{
			// Basic Latin for romaji input
			'a': VK_A, 'b': VK_B, 'c': VK_C, 'd': VK_D, 'e': VK_E,
			'f': VK_F, 'g': VK_G, 'h': VK_H, 'i': VK_I, 'j': VK_J,
			'k': VK_K, 'l': VK_L, 'm': VK_M, 'n': VK_N, 'o': VK_O,
			'p': VK_P, 'q': VK_Q, 'r': VK_R, 's': VK_S, 't': VK_T,
			'u': VK_U, 'v': VK_V, 'w': VK_W, 'x': VK_X, 'y': VK_Y, 'z': VK_Z,
			'0': VK_0, '1': VK_1, '2': VK_2, '3': VK_3, '4': VK_4,
			'5': VK_5, '6': VK_6, '7': VK_7, '8': VK_8, '9': VK_9,
			' ': VK_SPACE, '\t': VK_TAB, '\n': VK_RETURN,
		},
		ComposeKeys: map[string]rune{
			// Hiragana compose sequences (simplified)
			"ka": 'か', "ki": 'き', "ku": 'く', "ke": 'け', "ko": 'こ',
			"sa": 'さ', "shi": 'し', "su": 'す', "se": 'せ', "so": 'そ',
			"ta": 'た', "chi": 'ち', "tsu": 'つ', "te": 'て', "to": 'と',
			"na": 'な', "ni": 'に', "nu": 'ぬ', "ne": 'ね', "no": 'の',
			"ha": 'は', "hi": 'ひ', "fu": 'ふ', "he": 'へ', "ho": 'ほ',
			"ma": 'ま', "mi": 'み', "mu": 'む', "me": 'め', "mo": 'も',
			"ya": 'や', "yu": 'ゆ', "yo": 'よ',
			"ra": 'ら', "ri": 'り', "ru": 'る', "re": 'れ', "ro": 'ろ',
			"wa": 'わ', "wo": 'を', "n": 'ん',
		},
		IME: true,
		RTL: false,
	}
}

func createChineseSimplifiedLayout() *KeyboardLayout {
	return &KeyboardLayout{
		ID:       "zh_CN",
		Name:     "Chinese (Simplified)",
		Language: "Chinese",
		Script:   "Hanzi",
		KeyMap: map[rune]uint8{
			// Pinyin input uses Latin characters
			'a': VK_A, 'b': VK_B, 'c': VK_C, 'd': VK_D, 'e': VK_E,
			'f': VK_F, 'g': VK_G, 'h': VK_H, 'i': VK_I, 'j': VK_J,
			'k': VK_K, 'l': VK_L, 'm': VK_M, 'n': VK_N, 'o': VK_O,
			'p': VK_P, 'q': VK_Q, 'r': VK_R, 's': VK_S, 't': VK_T,
			'u': VK_U, 'v': VK_V, 'w': VK_W, 'x': VK_X, 'y': VK_Y, 'z': VK_Z,
			'0': VK_0, '1': VK_1, '2': VK_2, '3': VK_3, '4': VK_4,
			'5': VK_5, '6': VK_6, '7': VK_7, '8': VK_8, '9': VK_9,
			' ': VK_SPACE, '\t': VK_TAB, '\n': VK_RETURN,
		},
		IME: true,
		RTL: false,
	}
}

func createKoreanLayout() *KeyboardLayout {
	return &KeyboardLayout{
		ID:       "ko_KR",
		Name:     "Korean",
		Language: "Korean",
		Script:   "Hangul",
		KeyMap: map[rune]uint8{
			// Hangul input uses Latin characters for romanization
			'a': VK_A, 'b': VK_B, 'c': VK_C, 'd': VK_D, 'e': VK_E,
			'f': VK_F, 'g': VK_G, 'h': VK_H, 'i': VK_I, 'j': VK_J,
			'k': VK_K, 'l': VK_L, 'm': VK_M, 'n': VK_N, 'o': VK_O,
			'p': VK_P, 'q': VK_Q, 'r': VK_R, 's': VK_S, 't': VK_T,
			'u': VK_U, 'v': VK_V, 'w': VK_W, 'x': VK_X, 'y': VK_Y, 'z': VK_Z,
			'0': VK_0, '1': VK_1, '2': VK_2, '3': VK_3, '4': VK_4,
			'5': VK_5, '6': VK_6, '7': VK_7, '8': VK_8, '9': VK_9,
			' ': VK_SPACE, '\t': VK_TAB, '\n': VK_RETURN,
		},
		IME: true,
		RTL: false,
	}
}

func createArabicLayout() *KeyboardLayout {
	return &KeyboardLayout{
		ID:       "ar_SA",
		Name:     "Arabic",
		Language: "Arabic",
		Script:   "Arabic",
		KeyMap: map[rune]uint8{
			// Arabic letters
			'ا': VK_A, 'ب': VK_B, 'ت': VK_C, 'ث': VK_D, 'ج': VK_E,
			'ح': VK_F, 'خ': VK_G, 'د': VK_H, 'ذ': VK_I, 'ر': VK_J,
			'ز': VK_K, 'س': VK_L, 'ش': VK_M, 'ص': VK_N, 'ض': VK_O,
			'ط': VK_P, 'ظ': VK_Q, 'ع': VK_R, 'غ': VK_S, 'ف': VK_T,
			'ق': VK_U, 'ك': VK_V, 'ل': VK_W, 'م': VK_X, 'ن': VK_Y,
			'ه': VK_Z, 'و': VK_OEM_1, 'ي': VK_OEM_2, 'ة': VK_OEM_3,
			'ى': VK_OEM_4, 'ء': VK_OEM_5,
			// Numbers
			'0': VK_0, '1': VK_1, '2': VK_2, '3': VK_3, '4': VK_4,
			'5': VK_5, '6': VK_6, '7': VK_7, '8': VK_8, '9': VK_9,
			' ': VK_SPACE, '\t': VK_TAB, '\n': VK_RETURN,
		},
		IME: false,
		RTL: true,
	}
}

// Additional layout creation functions (simplified for brevity)
func createEnglishUKLayout() *KeyboardLayout {
	layout := createEnglishUSLayout()
	layout.ID = "en_GB"
	layout.Name = "English (UK)"
	// Add UK-specific key mappings
	return layout
}

func createGermanLayout() *KeyboardLayout {
	layout := createEnglishUSLayout()
	layout.ID = "de_DE"
	layout.Name = "German"
	layout.Language = "German"
	// Add German-specific key mappings (umlauts, etc.)
	return layout
}

func createFrenchLayout() *KeyboardLayout {
	layout := createEnglishUSLayout()
	layout.ID = "fr_FR"
	layout.Name = "French"
	layout.Language = "French"
	// Add French-specific key mappings (accents, etc.)
	return layout
}

func createSpanishLayout() *KeyboardLayout {
	layout := createEnglishUSLayout()
	layout.ID = "es_ES"
	layout.Name = "Spanish"
	layout.Language = "Spanish"
	// Add Spanish-specific key mappings
	return layout
}

func createItalianLayout() *KeyboardLayout {
	layout := createEnglishUSLayout()
	layout.ID = "it_IT"
	layout.Name = "Italian"
	layout.Language = "Italian"
	// Add Italian-specific key mappings
	return layout
}

func createPortugueseLayout() *KeyboardLayout {
	layout := createEnglishUSLayout()
	layout.ID = "pt_PT"
	layout.Name = "Portuguese"
	layout.Language = "Portuguese"
	// Add Portuguese-specific key mappings
	return layout
}

func createDutchLayout() *KeyboardLayout {
	layout := createEnglishUSLayout()
	layout.ID = "nl_NL"
	layout.Name = "Dutch"
	layout.Language = "Dutch"
	// Add Dutch-specific key mappings
	return layout
}

func createSwedishLayout() *KeyboardLayout {
	layout := createEnglishUSLayout()
	layout.ID = "sv_SE"
	layout.Name = "Swedish"
	layout.Language = "Swedish"
	// Add Swedish-specific key mappings
	return layout
}

func createNorwegianLayout() *KeyboardLayout {
	layout := createEnglishUSLayout()
	layout.ID = "no_NO"
	layout.Name = "Norwegian"
	layout.Language = "Norwegian"
	// Add Norwegian-specific key mappings
	return layout
}

func createDanishLayout() *KeyboardLayout {
	layout := createEnglishUSLayout()
	layout.ID = "da_DK"
	layout.Name = "Danish"
	layout.Language = "Danish"
	// Add Danish-specific key mappings
	return layout
}

func createFinnishLayout() *KeyboardLayout {
	layout := createEnglishUSLayout()
	layout.ID = "fi_FI"
	layout.Name = "Finnish"
	layout.Language = "Finnish"
	// Add Finnish-specific key mappings
	return layout
}

func createUkrainianLayout() *KeyboardLayout {
	layout := createRussianLayout()
	layout.ID = "uk_UA"
	layout.Name = "Ukrainian"
	layout.Language = "Ukrainian"
	// Add Ukrainian-specific Cyrillic mappings
	return layout
}

func createBulgarianLayout() *KeyboardLayout {
	layout := createRussianLayout()
	layout.ID = "bg_BG"
	layout.Name = "Bulgarian"
	layout.Language = "Bulgarian"
	// Add Bulgarian-specific Cyrillic mappings
	return layout
}

func createSerbianLayout() *KeyboardLayout {
	layout := createRussianLayout()
	layout.ID = "sr_RS"
	layout.Name = "Serbian"
	layout.Language = "Serbian"
	// Add Serbian-specific Cyrillic mappings
	return layout
}

func createChineseTraditionalLayout() *KeyboardLayout {
	layout := createChineseSimplifiedLayout()
	layout.ID = "zh_TW"
	layout.Name = "Chinese (Traditional)"
	layout.Language = "Chinese"
	// Traditional Chinese uses different character set
	return layout
}

func createHebrewLayout() *KeyboardLayout {
	return &KeyboardLayout{
		ID:       "he_IL",
		Name:     "Hebrew",
		Language: "Hebrew",
		Script:   "Hebrew",
		KeyMap: map[rune]uint8{
			// Hebrew letters
			'א': VK_A, 'ב': VK_B, 'ג': VK_C, 'ד': VK_D, 'ה': VK_E,
			'ו': VK_F, 'ז': VK_G, 'ח': VK_H, 'ט': VK_I, 'י': VK_J,
			'כ': VK_K, 'ל': VK_L, 'מ': VK_M, 'נ': VK_N, 'ס': VK_O,
			'ע': VK_P, 'פ': VK_Q, 'צ': VK_R, 'ק': VK_S, 'ר': VK_T,
			'ש': VK_U, 'ת': VK_V,
			// Numbers
			'0': VK_0, '1': VK_1, '2': VK_2, '3': VK_3, '4': VK_4,
			'5': VK_5, '6': VK_6, '7': VK_7, '8': VK_8, '9': VK_9,
			' ': VK_SPACE, '\t': VK_TAB, '\n': VK_RETURN,
		},
		IME: false,
		RTL: true,
	}
}

func createPersianLayout() *KeyboardLayout {
	layout := createArabicLayout()
	layout.ID = "fa_IR"
	layout.Name = "Persian"
	layout.Language = "Persian"
	// Add Persian-specific mappings
	return layout
}

func createTurkishLayout() *KeyboardLayout {
	layout := createEnglishUSLayout()
	layout.ID = "tr_TR"
	layout.Name = "Turkish"
	layout.Language = "Turkish"
	// Add Turkish-specific mappings (ç, ğ, ı, ö, ş, ü)
	return layout
}

func createHindiLayout() *KeyboardLayout {
	return &KeyboardLayout{
		ID:       "hi_IN",
		Name:     "Hindi",
		Language: "Hindi",
		Script:   "Devanagari",
		KeyMap: map[rune]uint8{
			// Devanagari letters (simplified mapping)
			'अ': VK_A, 'ब': VK_B, 'च': VK_C, 'द': VK_D, 'ए': VK_E,
			'फ': VK_F, 'ग': VK_G, 'ह': VK_H, 'इ': VK_I, 'ज': VK_J,
			'क': VK_K, 'ल': VK_L, 'म': VK_M, 'न': VK_N, 'ओ': VK_O,
			'प': VK_P, 'क़': VK_Q, 'र': VK_R, 'स': VK_S, 'ट': VK_T,
			'उ': VK_U, 'व': VK_V, 'ड': VK_W, 'ख': VK_X, 'य': VK_Y,
			'ज़': VK_Z,
			// Numbers
			'0': VK_0, '1': VK_1, '2': VK_2, '3': VK_3, '4': VK_4,
			'5': VK_5, '6': VK_6, '7': VK_7, '8': VK_8, '9': VK_9,
			' ': VK_SPACE, '\t': VK_TAB, '\n': VK_RETURN,
		},
		IME: true,
		RTL: false,
	}
}

func createThaiLayout() *KeyboardLayout {
	return &KeyboardLayout{
		ID:       "th_TH",
		Name:     "Thai",
		Language: "Thai",
		Script:   "Thai",
		KeyMap: map[rune]uint8{
			// Thai letters
			'ก': VK_A, 'ข': VK_B, 'ค': VK_C, 'ง': VK_D, 'จ': VK_E,
			'ฉ': VK_F, 'ช': VK_G, 'ซ': VK_H, 'ญ': VK_I, 'ฎ': VK_J,
			'ฏ': VK_K, 'ฐ': VK_L, 'ฑ': VK_M, 'ฒ': VK_N, 'ณ': VK_O,
			'ด': VK_P, 'ต': VK_Q, 'ถ': VK_R, 'ท': VK_S, 'ธ': VK_T,
			'น': VK_U, 'บ': VK_V, 'ป': VK_W, 'ผ': VK_X, 'ฝ': VK_Y,
			'พ': VK_Z, 'ฟ': VK_OEM_1, 'ภ': VK_OEM_2, 'ม': VK_OEM_3,
			'ย': VK_OEM_4, 'ร': VK_OEM_5, 'ล': VK_OEM_6, 'ว': VK_OEM_7,
			'ศ': VK_OEM_8, 'ษ': VK_OEM_9, 'ส': VK_OEM_10, 'ห': VK_OEM_11,
			'ฬ': VK_OEM_12, 'อ': VK_OEM_13, 'ฮ': VK_OEM_14,
			// Numbers
			'0': VK_0, '1': VK_1, '2': VK_2, '3': VK_3, '4': VK_4,
			'5': VK_5, '6': VK_6, '7': VK_7, '8': VK_8, '9': VK_9,
			' ': VK_SPACE, '\t': VK_TAB, '\n': VK_RETURN,
		},
		IME: false,
		RTL: false,
	}
}

func createVietnameseLayout() *KeyboardLayout {
	layout := createEnglishUSLayout()
	layout.ID = "vi_VN"
	layout.Name = "Vietnamese"
	layout.Language = "Vietnamese"
	// Add Vietnamese-specific mappings (diacritics)
	return layout
} 