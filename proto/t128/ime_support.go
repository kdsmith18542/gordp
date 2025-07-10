// IME (Input Method Editor) Support for GoRDP
// Provides comprehensive input method support for CJK, Greek, Cyrillic, and other complex scripts
// Handles composition, candidate selection, and proper Unicode input

package t128

import (
	"fmt"
	"unicode"
	"unicode/utf8"
)

// IMEType represents the type of input method
type IMEType int

const (
	IMETypeNone     IMEType = iota
	IMETypePinyin           // Chinese Pinyin
	IMETypeZhuyin           // Chinese Zhuyin (Bopomofo)
	IMETypeHiragana         // Japanese Hiragana
	IMETypeKatakana         // Japanese Katakana
	IMETypeHangul           // Korean Hangul
	IMETypeGreek            // Greek
	IMETypeCyrillic         // Cyrillic
	IMETypeArabic           // Arabic
	IMETypeHebrew           // Hebrew
	IMETypeThai             // Thai
	IMETypeHindi            // Hindi (Devanagari)
)

// IMEState represents the current state of the IME
type IMEState int

const (
	IMEStateInactive IMEState = iota
	IMEStateComposing
	IMEStateCandidate
	IMEStateConverting
)

// IMECandidate represents a candidate character/word
type IMECandidate struct {
	Text        string
	Description string
	Index       int
}

// IMEComposition represents the current composition
type IMEComposition struct {
	Text        string
	CursorPos   int
	StartPos    int
	EndPos      int
	Candidates  []IMECandidate
	SelectedIdx int
}

// IMEManager manages input method editing
type IMEManager struct {
	imeType     IMEType
	state       IMEState
	composition IMEComposition
	history     []string
	maxHistory  int
	enabled     bool

	// Language-specific handlers
	pinyinHandler   *PinyinHandler
	japaneseHandler *JapaneseHandler
	koreanHandler   *KoreanHandler
	greekHandler    *GreekHandler
	cyrillicHandler *CyrillicHandler
	arabicHandler   *ArabicHandler
	hebrewHandler   *HebrewHandler
	thaiHandler     *ThaiHandler
	hindiHandler    *HindiHandler
}

// NewIMEManager creates a new IME manager
func NewIMEManager() *IMEManager {
	manager := &IMEManager{
		imeType:    IMETypeNone,
		state:      IMEStateInactive,
		enabled:    false,
		maxHistory: 100,
		history:    make([]string, 0, 100),
	}

	// Initialize language-specific handlers
	manager.initializeHandlers()

	return manager
}

// initializeHandlers sets up language-specific IME handlers
func (manager *IMEManager) initializeHandlers() {
	manager.pinyinHandler = NewPinyinHandler()
	manager.japaneseHandler = NewJapaneseHandler()
	manager.koreanHandler = NewKoreanHandler()
	manager.greekHandler = NewGreekHandler()
	manager.cyrillicHandler = NewCyrillicHandler()
	manager.arabicHandler = NewArabicHandler()
	manager.hebrewHandler = NewHebrewHandler()
	manager.thaiHandler = NewThaiHandler()
	manager.hindiHandler = NewHindiHandler()
}

// SetIMEType sets the IME type and enables IME processing
func (manager *IMEManager) SetIMEType(imeType IMEType) {
	manager.imeType = imeType
	manager.enabled = (imeType != IMETypeNone)
	manager.ClearComposition()
}

// IsEnabled returns whether IME is enabled
func (manager *IMEManager) IsEnabled() bool {
	return manager.enabled
}

// GetState returns the current IME state
func (manager *IMEManager) GetState() IMEState {
	return manager.state
}

// GetComposition returns the current composition
func (manager *IMEManager) GetComposition() IMEComposition {
	return manager.composition
}

// ProcessInput processes input for IME composition
func (manager *IMEManager) ProcessInput(char rune, modifiers ModifierKey) ([]TsFpInputEvent, bool) {
	if !manager.enabled {
		return nil, false
	}

	switch manager.imeType {
	case IMETypePinyin:
		return manager.processPinyinInput(char, modifiers)
	case IMETypeHiragana, IMETypeKatakana:
		return manager.processJapaneseInput(char, modifiers)
	case IMETypeHangul:
		return manager.processKoreanInput(char, modifiers)
	case IMETypeGreek:
		return manager.processGreekInput(char, modifiers)
	case IMETypeCyrillic:
		return manager.processCyrillicInput(char, modifiers)
	case IMETypeArabic:
		return manager.processArabicInput(char, modifiers)
	case IMETypeHebrew:
		return manager.processHebrewInput(char, modifiers)
	case IMETypeThai:
		return manager.processThaiInput(char, modifiers)
	case IMETypeHindi:
		return manager.processHindiInput(char, modifiers)
	default:
		return nil, false
	}
}

// ClearComposition clears the current composition
func (manager *IMEManager) ClearComposition() {
	manager.composition = IMEComposition{}
	manager.state = IMEStateInactive
}

// SelectCandidate selects a candidate from the composition
func (manager *IMEManager) SelectCandidate(index int) ([]TsFpInputEvent, bool) {
	if manager.state != IMEStateCandidate || index < 0 || index >= len(manager.composition.Candidates) {
		return nil, false
	}

	candidate := manager.composition.Candidates[index]
	manager.addToHistory(candidate.Text)

	// Convert to Unicode events
	events := make([]TsFpInputEvent, 0, utf8.RuneCountInString(candidate.Text))
	for _, char := range candidate.Text {
		events = append(events, &TsFpUnicodeEvent{
			EventHeader: 1, // Key down
			UnicodeCode: uint16(char),
		})
	}

	manager.ClearComposition()
	return events, true
}

// ConfirmComposition confirms the current composition
func (manager *IMEManager) ConfirmComposition() ([]TsFpInputEvent, bool) {
	if manager.state == IMEStateInactive {
		return nil, false
	}

	text := manager.composition.Text
	if text == "" {
		return nil, false
	}

	manager.addToHistory(text)

	// Convert to Unicode events
	events := make([]TsFpInputEvent, 0, utf8.RuneCountInString(text))
	for _, char := range text {
		events = append(events, &TsFpUnicodeEvent{
			EventHeader: 1, // Key down
			UnicodeCode: uint16(char),
		})
	}

	manager.ClearComposition()
	return events, true
}

// CancelComposition cancels the current composition
func (manager *IMEManager) CancelComposition() bool {
	if manager.state == IMEStateInactive {
		return false
	}

	manager.ClearComposition()
	return true
}

// addToHistory adds text to the input history
func (manager *IMEManager) addToHistory(text string) {
	if text == "" {
		return
	}

	// Add to history (avoid duplicates)
	for i, hist := range manager.history {
		if hist == text {
			// Move to front
			manager.history = append(manager.history[:i], manager.history[i+1:]...)
			break
		}
	}

	manager.history = append([]string{text}, manager.history...)

	// Keep history size manageable
	if len(manager.history) > manager.maxHistory {
		manager.history = manager.history[:manager.maxHistory]
	}
}

// GetHistory returns the input history
func (manager *IMEManager) GetHistory() []string {
	return manager.history
}

// ============================================================================
// Language-specific IME handlers
// ============================================================================

// PinyinHandler handles Chinese Pinyin input
type PinyinHandler struct {
	pinyinMap map[string][]string
}

func NewPinyinHandler() *PinyinHandler {
	handler := &PinyinHandler{
		pinyinMap: make(map[string][]string),
	}
	handler.initializePinyinMap()
	return handler
}

func (h *PinyinHandler) initializePinyinMap() {
	// Common Pinyin mappings (simplified)
	h.pinyinMap["ni"] = []string{"你", "尼", "泥", "逆", "拟"}
	h.pinyinMap["hao"] = []string{"好", "号", "毫", "豪", "耗"}
	h.pinyinMap["wo"] = []string{"我", "握", "卧", "沃", "涡"}
	h.pinyinMap["shi"] = []string{"是", "时", "事", "世", "市"}
	h.pinyinMap["de"] = []string{"的", "得", "德", "地", "底"}
	h.pinyinMap["le"] = []string{"了", "乐", "勒", "雷", "累"}
	h.pinyinMap["zai"] = []string{"在", "再", "载", "灾", "栽"}
	h.pinyinMap["you"] = []string{"有", "又", "右", "由", "游"}
	h.pinyinMap["he"] = []string{"和", "河", "合", "何", "核"}
	h.pinyinMap["zhe"] = []string{"这", "者", "着", "折", "哲"}
}

func (manager *IMEManager) processPinyinInput(char rune, modifiers ModifierKey) ([]TsFpInputEvent, bool) {
	handler := manager.pinyinHandler

	// Handle space to confirm
	if char == ' ' {
		return manager.ConfirmComposition()
	}

	// Handle backspace
	if char == '\b' {
		if len(manager.composition.Text) > 0 {
			manager.composition.Text = manager.composition.Text[:len(manager.composition.Text)-1]
			if len(manager.composition.Text) == 0 {
				manager.ClearComposition()
				return nil, true
			}
		}
		return nil, true
	}

	// Handle number keys for candidate selection
	if char >= '1' && char <= '9' {
		index := int(char - '1')
		return manager.SelectCandidate(index)
	}

	// Add character to composition
	if unicode.IsLetter(char) {
		manager.composition.Text += string(char)
		manager.state = IMEStateComposing

		// Look up candidates
		if candidates, exists := handler.pinyinMap[manager.composition.Text]; exists {
			manager.composition.Candidates = make([]IMECandidate, len(candidates))
			for i, candidate := range candidates {
				manager.composition.Candidates[i] = IMECandidate{
					Text:        candidate,
					Description: fmt.Sprintf("%d. %s", i+1, candidate),
					Index:       i,
				}
			}
			manager.state = IMEStateCandidate
		}

		return nil, true
	}

	return nil, false
}

// JapaneseHandler handles Japanese input (Hiragana/Katakana)
type JapaneseHandler struct {
	romajiMap map[string]string
}

func NewJapaneseHandler() *JapaneseHandler {
	handler := &JapaneseHandler{
		romajiMap: make(map[string]string),
	}
	handler.initializeRomajiMap()
	return handler
}

func (h *JapaneseHandler) initializeRomajiMap() {
	// Common Romaji to Hiragana mappings
	h.romajiMap["a"] = "あ"
	h.romajiMap["i"] = "い"
	h.romajiMap["u"] = "う"
	h.romajiMap["e"] = "え"
	h.romajiMap["o"] = "お"
	h.romajiMap["ka"] = "か"
	h.romajiMap["ki"] = "き"
	h.romajiMap["ku"] = "く"
	h.romajiMap["ke"] = "け"
	h.romajiMap["ko"] = "こ"
	h.romajiMap["sa"] = "さ"
	h.romajiMap["shi"] = "し"
	h.romajiMap["su"] = "す"
	h.romajiMap["se"] = "せ"
	h.romajiMap["so"] = "そ"
	h.romajiMap["ta"] = "た"
	h.romajiMap["chi"] = "ち"
	h.romajiMap["tsu"] = "つ"
	h.romajiMap["te"] = "て"
	h.romajiMap["to"] = "と"
	h.romajiMap["na"] = "な"
	h.romajiMap["ni"] = "に"
	h.romajiMap["nu"] = "ぬ"
	h.romajiMap["ne"] = "ね"
	h.romajiMap["no"] = "の"
	h.romajiMap["ha"] = "は"
	h.romajiMap["hi"] = "ひ"
	h.romajiMap["fu"] = "ふ"
	h.romajiMap["he"] = "へ"
	h.romajiMap["ho"] = "ほ"
	h.romajiMap["ma"] = "ま"
	h.romajiMap["mi"] = "み"
	h.romajiMap["mu"] = "む"
	h.romajiMap["me"] = "め"
	h.romajiMap["mo"] = "も"
	h.romajiMap["ya"] = "や"
	h.romajiMap["yu"] = "ゆ"
	h.romajiMap["yo"] = "よ"
	h.romajiMap["ra"] = "ら"
	h.romajiMap["ri"] = "り"
	h.romajiMap["ru"] = "る"
	h.romajiMap["re"] = "れ"
	h.romajiMap["ro"] = "ろ"
	h.romajiMap["wa"] = "わ"
	h.romajiMap["wo"] = "を"
	h.romajiMap["n"] = "ん"
}

func (manager *IMEManager) processJapaneseInput(char rune, modifiers ModifierKey) ([]TsFpInputEvent, bool) {
	handler := manager.japaneseHandler

	// Handle space to confirm
	if char == ' ' {
		return manager.ConfirmComposition()
	}

	// Handle backspace
	if char == '\b' {
		if len(manager.composition.Text) > 0 {
			manager.composition.Text = manager.composition.Text[:len(manager.composition.Text)-1]
			if len(manager.composition.Text) == 0 {
				manager.ClearComposition()
				return nil, true
			}
		}
		return nil, true
	}

	// Add character to composition
	if unicode.IsLetter(char) {
		manager.composition.Text += string(char)
		manager.state = IMEStateComposing

		// Look up Hiragana
		if hiragana, exists := handler.romajiMap[manager.composition.Text]; exists {
			manager.composition.Candidates = []IMECandidate{
				{
					Text:        hiragana,
					Description: fmt.Sprintf("1. %s", hiragana),
					Index:       0,
				},
			}
			manager.state = IMEStateCandidate
		}

		return nil, true
	}

	return nil, false
}

// KoreanHandler handles Korean Hangul input
type KoreanHandler struct {
	jamoMap map[string]string
}

func NewKoreanHandler() *KoreanHandler {
	handler := &KoreanHandler{
		jamoMap: make(map[string]string),
	}
	handler.initializeJamoMap()
	return handler
}

func (h *KoreanHandler) initializeJamoMap() {
	// Common Jamo to Hangul mappings
	h.jamoMap["ga"] = "가"
	h.jamoMap["na"] = "나"
	h.jamoMap["da"] = "다"
	h.jamoMap["ra"] = "라"
	h.jamoMap["ma"] = "마"
	h.jamoMap["ba"] = "바"
	h.jamoMap["sa"] = "사"
	h.jamoMap["a"] = "아"
	h.jamoMap["ja"] = "자"
	h.jamoMap["cha"] = "차"
	h.jamoMap["ka"] = "카"
	h.jamoMap["ta"] = "타"
	h.jamoMap["pa"] = "파"
	h.jamoMap["ha"] = "하"
}

func (manager *IMEManager) processKoreanInput(char rune, modifiers ModifierKey) ([]TsFpInputEvent, bool) {
	handler := manager.koreanHandler

	// Handle space to confirm
	if char == ' ' {
		return manager.ConfirmComposition()
	}

	// Handle backspace
	if char == '\b' {
		if len(manager.composition.Text) > 0 {
			manager.composition.Text = manager.composition.Text[:len(manager.composition.Text)-1]
			if len(manager.composition.Text) == 0 {
				manager.ClearComposition()
				return nil, true
			}
		}
		return nil, true
	}

	// Add character to composition
	if unicode.IsLetter(char) {
		manager.composition.Text += string(char)
		manager.state = IMEStateComposing

		// Look up Hangul
		if hangul, exists := handler.jamoMap[manager.composition.Text]; exists {
			manager.composition.Candidates = []IMECandidate{
				{
					Text:        hangul,
					Description: fmt.Sprintf("1. %s", hangul),
					Index:       0,
				},
			}
			manager.state = IMEStateCandidate
		}

		return nil, true
	}

	return nil, false
}

// GreekHandler handles Greek input
type GreekHandler struct {
	transliterationMap map[string]string
}

func NewGreekHandler() *GreekHandler {
	handler := &GreekHandler{
		transliterationMap: make(map[string]string),
	}
	handler.initializeTransliterationMap()
	return handler
}

func (h *GreekHandler) initializeTransliterationMap() {
	// Latin to Greek transliteration
	h.transliterationMap["a"] = "α"
	h.transliterationMap["b"] = "β"
	h.transliterationMap["g"] = "γ"
	h.transliterationMap["d"] = "δ"
	h.transliterationMap["e"] = "ε"
	h.transliterationMap["z"] = "ζ"
	h.transliterationMap["h"] = "η"
	h.transliterationMap["th"] = "θ"
	h.transliterationMap["i"] = "ι"
	h.transliterationMap["k"] = "κ"
	h.transliterationMap["l"] = "λ"
	h.transliterationMap["m"] = "μ"
	h.transliterationMap["n"] = "ν"
	h.transliterationMap["x"] = "ξ"
	h.transliterationMap["o"] = "ο"
	h.transliterationMap["p"] = "π"
	h.transliterationMap["r"] = "ρ"
	h.transliterationMap["s"] = "σ"
	h.transliterationMap["t"] = "τ"
	h.transliterationMap["u"] = "υ"
	h.transliterationMap["ph"] = "φ"
	h.transliterationMap["ch"] = "χ"
	h.transliterationMap["ps"] = "ψ"
	h.transliterationMap["w"] = "ω"
}

func (manager *IMEManager) processGreekInput(char rune, modifiers ModifierKey) ([]TsFpInputEvent, bool) {
	handler := manager.greekHandler

	// Handle space to confirm
	if char == ' ' {
		return manager.ConfirmComposition()
	}

	// Handle backspace
	if char == '\b' {
		if len(manager.composition.Text) > 0 {
			manager.composition.Text = manager.composition.Text[:len(manager.composition.Text)-1]
			if len(manager.composition.Text) == 0 {
				manager.ClearComposition()
				return nil, true
			}
		}
		return nil, true
	}

	// Add character to composition
	if unicode.IsLetter(char) {
		manager.composition.Text += string(char)
		manager.state = IMEStateComposing

		// Look up Greek character
		if greek, exists := handler.transliterationMap[manager.composition.Text]; exists {
			manager.composition.Candidates = []IMECandidate{
				{
					Text:        greek,
					Description: fmt.Sprintf("1. %s", greek),
					Index:       0,
				},
			}
			manager.state = IMEStateCandidate
		}

		return nil, true
	}

	return nil, false
}

// CyrillicHandler handles Cyrillic input
type CyrillicHandler struct {
	transliterationMap map[string]string
}

func NewCyrillicHandler() *CyrillicHandler {
	handler := &CyrillicHandler{
		transliterationMap: make(map[string]string),
	}
	handler.initializeTransliterationMap()
	return handler
}

func (h *CyrillicHandler) initializeTransliterationMap() {
	// Latin to Cyrillic transliteration
	h.transliterationMap["a"] = "а"
	h.transliterationMap["b"] = "б"
	h.transliterationMap["v"] = "в"
	h.transliterationMap["g"] = "г"
	h.transliterationMap["d"] = "д"
	h.transliterationMap["e"] = "е"
	h.transliterationMap["zh"] = "ж"
	h.transliterationMap["z"] = "з"
	h.transliterationMap["i"] = "и"
	h.transliterationMap["y"] = "й"
	h.transliterationMap["k"] = "к"
	h.transliterationMap["l"] = "л"
	h.transliterationMap["m"] = "м"
	h.transliterationMap["n"] = "н"
	h.transliterationMap["o"] = "о"
	h.transliterationMap["p"] = "п"
	h.transliterationMap["r"] = "р"
	h.transliterationMap["s"] = "с"
	h.transliterationMap["t"] = "т"
	h.transliterationMap["u"] = "у"
	h.transliterationMap["f"] = "ф"
	h.transliterationMap["kh"] = "х"
	h.transliterationMap["ts"] = "ц"
	h.transliterationMap["ch"] = "ч"
	h.transliterationMap["sh"] = "ш"
	h.transliterationMap["shch"] = "щ"
	h.transliterationMap["'"] = "ъ"
	h.transliterationMap["y"] = "ы"
	h.transliterationMap["'"] = "ь"
	h.transliterationMap["e"] = "э"
	h.transliterationMap["yu"] = "ю"
	h.transliterationMap["ya"] = "я"
}

func (manager *IMEManager) processCyrillicInput(char rune, modifiers ModifierKey) ([]TsFpInputEvent, bool) {
	handler := manager.cyrillicHandler

	// Handle space to confirm
	if char == ' ' {
		return manager.ConfirmComposition()
	}

	// Handle backspace
	if char == '\b' {
		if len(manager.composition.Text) > 0 {
			manager.composition.Text = manager.composition.Text[:len(manager.composition.Text)-1]
			if len(manager.composition.Text) == 0 {
				manager.ClearComposition()
				return nil, true
			}
		}
		return nil, true
	}

	// Add character to composition
	if unicode.IsLetter(char) || char == '\'' {
		manager.composition.Text += string(char)
		manager.state = IMEStateComposing

		// Look up Cyrillic character
		if cyrillic, exists := handler.transliterationMap[manager.composition.Text]; exists {
			manager.composition.Candidates = []IMECandidate{
				{
					Text:        cyrillic,
					Description: fmt.Sprintf("1. %s", cyrillic),
					Index:       0,
				},
			}
			manager.state = IMEStateCandidate
		}

		return nil, true
	}

	return nil, false
}

// ArabicHandler handles Arabic input
type ArabicHandler struct {
	transliterationMap map[string]string
}

func NewArabicHandler() *ArabicHandler {
	handler := &ArabicHandler{
		transliterationMap: make(map[string]string),
	}
	handler.initializeTransliterationMap()
	return handler
}

func (h *ArabicHandler) initializeTransliterationMap() {
	// Latin to Arabic transliteration
	h.transliterationMap["a"] = "ا"
	h.transliterationMap["b"] = "ب"
	h.transliterationMap["t"] = "ت"
	h.transliterationMap["th"] = "ث"
	h.transliterationMap["j"] = "ج"
	h.transliterationMap["h"] = "ح"
	h.transliterationMap["kh"] = "خ"
	h.transliterationMap["d"] = "د"
	h.transliterationMap["dh"] = "ذ"
	h.transliterationMap["r"] = "ر"
	h.transliterationMap["z"] = "ز"
	h.transliterationMap["s"] = "س"
	h.transliterationMap["sh"] = "ش"
	h.transliterationMap["s"] = "ص"
	h.transliterationMap["d"] = "ض"
	h.transliterationMap["t"] = "ط"
	h.transliterationMap["z"] = "ظ"
	h.transliterationMap["'"] = "ع"
	h.transliterationMap["gh"] = "غ"
	h.transliterationMap["f"] = "ف"
	h.transliterationMap["q"] = "ق"
	h.transliterationMap["k"] = "ك"
	h.transliterationMap["l"] = "ل"
	h.transliterationMap["m"] = "م"
	h.transliterationMap["n"] = "ن"
	h.transliterationMap["h"] = "ه"
	h.transliterationMap["w"] = "و"
	h.transliterationMap["y"] = "ي"
}

func (manager *IMEManager) processArabicInput(char rune, modifiers ModifierKey) ([]TsFpInputEvent, bool) {
	handler := manager.arabicHandler

	// Handle space to confirm
	if char == ' ' {
		return manager.ConfirmComposition()
	}

	// Handle backspace
	if char == '\b' {
		if len(manager.composition.Text) > 0 {
			manager.composition.Text = manager.composition.Text[:len(manager.composition.Text)-1]
			if len(manager.composition.Text) == 0 {
				manager.ClearComposition()
				return nil, true
			}
		}
		return nil, true
	}

	// Add character to composition
	if unicode.IsLetter(char) || char == '\'' {
		manager.composition.Text += string(char)
		manager.state = IMEStateComposing

		// Look up Arabic character
		if arabic, exists := handler.transliterationMap[manager.composition.Text]; exists {
			manager.composition.Candidates = []IMECandidate{
				{
					Text:        arabic,
					Description: fmt.Sprintf("1. %s", arabic),
					Index:       0,
				},
			}
			manager.state = IMEStateCandidate
		}

		return nil, true
	}

	return nil, false
}

// HebrewHandler handles Hebrew input
type HebrewHandler struct {
	transliterationMap map[string]string
}

func NewHebrewHandler() *HebrewHandler {
	handler := &HebrewHandler{
		transliterationMap: make(map[string]string),
	}
	handler.initializeTransliterationMap()
	return handler
}

func (h *HebrewHandler) initializeTransliterationMap() {
	// Latin to Hebrew transliteration
	h.transliterationMap["a"] = "א"
	h.transliterationMap["b"] = "ב"
	h.transliterationMap["g"] = "ג"
	h.transliterationMap["d"] = "ד"
	h.transliterationMap["h"] = "ה"
	h.transliterationMap["v"] = "ו"
	h.transliterationMap["z"] = "ז"
	h.transliterationMap["ch"] = "ח"
	h.transliterationMap["t"] = "ט"
	h.transliterationMap["y"] = "י"
	h.transliterationMap["k"] = "כ"
	h.transliterationMap["l"] = "ל"
	h.transliterationMap["m"] = "מ"
	h.transliterationMap["n"] = "נ"
	h.transliterationMap["s"] = "ס"
	h.transliterationMap["'"] = "ע"
	h.transliterationMap["p"] = "פ"
	h.transliterationMap["ts"] = "צ"
	h.transliterationMap["q"] = "ק"
	h.transliterationMap["r"] = "ר"
	h.transliterationMap["sh"] = "ש"
	h.transliterationMap["t"] = "ת"
}

func (manager *IMEManager) processHebrewInput(char rune, modifiers ModifierKey) ([]TsFpInputEvent, bool) {
	handler := manager.hebrewHandler

	// Handle space to confirm
	if char == ' ' {
		return manager.ConfirmComposition()
	}

	// Handle backspace
	if char == '\b' {
		if len(manager.composition.Text) > 0 {
			manager.composition.Text = manager.composition.Text[:len(manager.composition.Text)-1]
			if len(manager.composition.Text) == 0 {
				manager.ClearComposition()
				return nil, true
			}
		}
		return nil, true
	}

	// Add character to composition
	if unicode.IsLetter(char) || char == '\'' {
		manager.composition.Text += string(char)
		manager.state = IMEStateComposing

		// Look up Hebrew character
		if hebrew, exists := handler.transliterationMap[manager.composition.Text]; exists {
			manager.composition.Candidates = []IMECandidate{
				{
					Text:        hebrew,
					Description: fmt.Sprintf("1. %s", hebrew),
					Index:       0,
				},
			}
			manager.state = IMEStateCandidate
		}

		return nil, true
	}

	return nil, false
}

// ThaiHandler handles Thai input
type ThaiHandler struct {
	transliterationMap map[string]string
}

func NewThaiHandler() *ThaiHandler {
	handler := &ThaiHandler{
		transliterationMap: make(map[string]string),
	}
	handler.initializeTransliterationMap()
	return handler
}

func (h *ThaiHandler) initializeTransliterationMap() {
	// Latin to Thai transliteration
	h.transliterationMap["k"] = "ก"
	h.transliterationMap["kh"] = "ข"
	h.transliterationMap["kh"] = "ฃ"
	h.transliterationMap["kh"] = "ค"
	h.transliterationMap["kh"] = "ฅ"
	h.transliterationMap["ng"] = "ง"
	h.transliterationMap["j"] = "จ"
	h.transliterationMap["ch"] = "ฉ"
	h.transliterationMap["ch"] = "ช"
	h.transliterationMap["s"] = "ซ"
	h.transliterationMap["ch"] = "ฌ"
	h.transliterationMap["y"] = "ญ"
	h.transliterationMap["d"] = "ฎ"
	h.transliterationMap["t"] = "ฏ"
	h.transliterationMap["th"] = "ฐ"
	h.transliterationMap["th"] = "ฑ"
	h.transliterationMap["th"] = "ฒ"
	h.transliterationMap["n"] = "ณ"
	h.transliterationMap["d"] = "ด"
	h.transliterationMap["t"] = "ต"
	h.transliterationMap["th"] = "ถ"
	h.transliterationMap["th"] = "ท"
	h.transliterationMap["th"] = "ธ"
	h.transliterationMap["n"] = "น"
	h.transliterationMap["b"] = "บ"
	h.transliterationMap["p"] = "ป"
	h.transliterationMap["ph"] = "ผ"
	h.transliterationMap["f"] = "ฝ"
	h.transliterationMap["ph"] = "พ"
	h.transliterationMap["f"] = "ฟ"
	h.transliterationMap["ph"] = "ภ"
	h.transliterationMap["m"] = "ม"
	h.transliterationMap["y"] = "ย"
	h.transliterationMap["r"] = "ร"
	h.transliterationMap["l"] = "ล"
	h.transliterationMap["w"] = "ว"
	h.transliterationMap["s"] = "ศ"
	h.transliterationMap["s"] = "ษ"
	h.transliterationMap["s"] = "ส"
	h.transliterationMap["h"] = "ห"
	h.transliterationMap["l"] = "ฬ"
	h.transliterationMap["'"] = "อ"
	h.transliterationMap["h"] = "ฮ"
}

func (manager *IMEManager) processThaiInput(char rune, modifiers ModifierKey) ([]TsFpInputEvent, bool) {
	handler := manager.thaiHandler

	// Handle space to confirm
	if char == ' ' {
		return manager.ConfirmComposition()
	}

	// Handle backspace
	if char == '\b' {
		if len(manager.composition.Text) > 0 {
			manager.composition.Text = manager.composition.Text[:len(manager.composition.Text)-1]
			if len(manager.composition.Text) == 0 {
				manager.ClearComposition()
				return nil, true
			}
		}
		return nil, true
	}

	// Add character to composition
	if unicode.IsLetter(char) || char == '\'' {
		manager.composition.Text += string(char)
		manager.state = IMEStateComposing

		// Look up Thai character
		if thai, exists := handler.transliterationMap[manager.composition.Text]; exists {
			manager.composition.Candidates = []IMECandidate{
				{
					Text:        thai,
					Description: fmt.Sprintf("1. %s", thai),
					Index:       0,
				},
			}
			manager.state = IMEStateCandidate
		}

		return nil, true
	}

	return nil, false
}

// HindiHandler handles Hindi (Devanagari) input
type HindiHandler struct {
	transliterationMap map[string]string
}

func NewHindiHandler() *HindiHandler {
	handler := &HindiHandler{
		transliterationMap: make(map[string]string),
	}
	handler.initializeTransliterationMap()
	return handler
}

func (h *HindiHandler) initializeTransliterationMap() {
	// Latin to Devanagari transliteration
	h.transliterationMap["a"] = "अ"
	h.transliterationMap["aa"] = "आ"
	h.transliterationMap["i"] = "इ"
	h.transliterationMap["ii"] = "ई"
	h.transliterationMap["u"] = "उ"
	h.transliterationMap["uu"] = "ऊ"
	h.transliterationMap["e"] = "ए"
	h.transliterationMap["ai"] = "ऐ"
	h.transliterationMap["o"] = "ओ"
	h.transliterationMap["au"] = "औ"
	h.transliterationMap["k"] = "क"
	h.transliterationMap["kh"] = "ख"
	h.transliterationMap["g"] = "ग"
	h.transliterationMap["gh"] = "घ"
	h.transliterationMap["ng"] = "ङ"
	h.transliterationMap["ch"] = "च"
	h.transliterationMap["chh"] = "छ"
	h.transliterationMap["j"] = "ज"
	h.transliterationMap["jh"] = "झ"
	h.transliterationMap["ny"] = "ञ"
	h.transliterationMap["t"] = "ट"
	h.transliterationMap["th"] = "ठ"
	h.transliterationMap["d"] = "ड"
	h.transliterationMap["dh"] = "ढ"
	h.transliterationMap["n"] = "ण"
	h.transliterationMap["p"] = "प"
	h.transliterationMap["ph"] = "फ"
	h.transliterationMap["b"] = "ब"
	h.transliterationMap["bh"] = "भ"
	h.transliterationMap["m"] = "म"
	h.transliterationMap["y"] = "य"
	h.transliterationMap["r"] = "र"
	h.transliterationMap["l"] = "ल"
	h.transliterationMap["v"] = "व"
	h.transliterationMap["sh"] = "श"
	h.transliterationMap["s"] = "ष"
	h.transliterationMap["s"] = "स"
	h.transliterationMap["h"] = "ह"
	h.transliterationMap["l"] = "ळ"
	h.transliterationMap["ksh"] = "क्ष"
	h.transliterationMap["tr"] = "त्र"
	h.transliterationMap["gy"] = "ज्ञ"
}

func (manager *IMEManager) processHindiInput(char rune, modifiers ModifierKey) ([]TsFpInputEvent, bool) {
	handler := manager.hindiHandler

	// Handle space to confirm
	if char == ' ' {
		return manager.ConfirmComposition()
	}

	// Handle backspace
	if char == '\b' {
		if len(manager.composition.Text) > 0 {
			manager.composition.Text = manager.composition.Text[:len(manager.composition.Text)-1]
			if len(manager.composition.Text) == 0 {
				manager.ClearComposition()
				return nil, true
			}
		}
		return nil, true
	}

	// Add character to composition
	if unicode.IsLetter(char) {
		manager.composition.Text += string(char)
		manager.state = IMEStateComposing

		// Look up Devanagari character
		if devanagari, exists := handler.transliterationMap[manager.composition.Text]; exists {
			manager.composition.Candidates = []IMECandidate{
				{
					Text:        devanagari,
					Description: fmt.Sprintf("1. %s", devanagari),
					Index:       0,
				},
			}
			manager.state = IMEStateCandidate
		}

		return nil, true
	}

	return nil, false
}
