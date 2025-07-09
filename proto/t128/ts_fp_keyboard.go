package t128

// TsFpKeyboardEvent represents a keyboard input event in the Fast-Path Input Event format.
type TsFpKeyboardEvent struct {
	EventHeader uint8 // 2 bits: eventFlags (0 = key up, 1 = key down)
	KeyCode     uint8 // 8 bits: virtual key code
}

func (e *TsFpKeyboardEvent) iInputEvent() {}

// Serialize converts the keyboard event to bytes according to the RDP Fast-Path Input Event format.
func (e *TsFpKeyboardEvent) Serialize() []byte {
	b := make([]byte, 2)
	b[0] = (FASTPATH_INPUT_EVENT_SCANCODE << 5) | (e.EventHeader & 0x1F)
	b[1] = e.KeyCode
	return b
}

// NewFastPathKeyboardEvent creates a new keyboard input event.
func NewFastPathKeyboardEvent(keyCode uint8, down bool) *TsFpKeyboardEvent {
	eventHeader := uint8(0)
	if down {
		eventHeader = 1 // Key down event
	}
	return &TsFpKeyboardEvent{
		EventHeader: eventHeader,
		KeyCode:     keyCode,
	}
}

// Virtual key codes for special keys
const (
	VK_BACK     = 0x08
	VK_TAB      = 0x09
	VK_RETURN   = 0x0D
	VK_ENTER    = 0x0D // Alias for VK_RETURN
	VK_SHIFT    = 0x10
	VK_CONTROL  = 0x11
	VK_MENU     = 0x12 // Alt key
	VK_PAUSE    = 0x13
	VK_CAPITAL  = 0x14
	VK_ESCAPE   = 0x1B
	VK_SPACE    = 0x20
	VK_PRIOR    = 0x21 // Page Up
	VK_NEXT     = 0x22 // Page Down
	VK_END      = 0x23
	VK_HOME     = 0x24
	VK_LEFT     = 0x25
	VK_UP       = 0x26
	VK_RIGHT    = 0x27
	VK_DOWN     = 0x28
	VK_SELECT   = 0x29
	VK_PRINT    = 0x2A
	VK_EXECUTE  = 0x2B
	VK_SNAPSHOT = 0x2C // Print Screen
	VK_INSERT   = 0x2D
	VK_DELETE   = 0x2E
	VK_HELP     = 0x2F
	VK_CLEAR    = 0x0C // Clear key (numpad 5 when numlock is off)
	VK_LWIN     = 0x5B
	VK_RWIN     = 0x5C
	VK_APPS     = 0x5D
	VK_NUMLOCK  = 0x90
	VK_SCROLL   = 0x91
)

// Letter key codes
const (
	VK_A = 0x41
	VK_B = 0x42
	VK_C = 0x43
	VK_D = 0x44
	VK_E = 0x45
	VK_F = 0x46
	VK_G = 0x47
	VK_H = 0x48
	VK_I = 0x49
	VK_J = 0x4A
	VK_K = 0x4B
	VK_L = 0x4C
	VK_M = 0x4D
	VK_N = 0x4E
	VK_O = 0x4F
	VK_P = 0x50
	VK_Q = 0x51
	VK_R = 0x52
	VK_S = 0x53
	VK_T = 0x54
	VK_U = 0x55
	VK_V = 0x56
	VK_W = 0x57
	VK_X = 0x58
	VK_Y = 0x59
	VK_Z = 0x5A
)

// Extended key codes
const (
	VK_LSHIFT              = 0xA0
	VK_RSHIFT              = 0xA1
	VK_LCONTROL            = 0xA2
	VK_RCONTROL            = 0xA3
	VK_LMENU               = 0xA4
	VK_RMENU               = 0xA5
	VK_BROWSER_BACK        = 0xA6
	VK_BROWSER_FORWARD     = 0xA7
	VK_BROWSER_REFRESH     = 0xA8
	VK_BROWSER_STOP        = 0xA9
	VK_BROWSER_SEARCH      = 0xAA
	VK_BROWSER_FAVORITES   = 0xAB
	VK_BROWSER_HOME        = 0xAC
	VK_VOLUME_MUTE         = 0xAD
	VK_VOLUME_DOWN         = 0xAE
	VK_VOLUME_UP           = 0xAF
	VK_MEDIA_NEXT_TRACK    = 0xB0
	VK_MEDIA_PREV_TRACK    = 0xB1
	VK_MEDIA_STOP          = 0xB2
	VK_MEDIA_PLAY_PAUSE    = 0xB3
	VK_LAUNCH_MAIL         = 0xB4
	VK_LAUNCH_MEDIA_SELECT = 0xB5
	VK_LAUNCH_APP1         = 0xB6
	VK_LAUNCH_APP2         = 0xB7
)

// Numpad keys
const (
	VK_NUMPAD0   = 0x60
	VK_NUMPAD1   = 0x61
	VK_NUMPAD2   = 0x62
	VK_NUMPAD3   = 0x63
	VK_NUMPAD4   = 0x64
	VK_NUMPAD5   = 0x65
	VK_NUMPAD6   = 0x66
	VK_NUMPAD7   = 0x67
	VK_NUMPAD8   = 0x68
	VK_NUMPAD9   = 0x69
	VK_MULTIPLY  = 0x6A
	VK_ADD       = 0x6B
	VK_SEPARATOR = 0x6C
	VK_SUBTRACT  = 0x6D
	VK_DECIMAL   = 0x6E
	VK_DIVIDE    = 0x6F
)

// KeyMap maps common ASCII characters to their corresponding virtual key codes.
var KeyMap = map[rune]uint8{
	'0':    0x30,
	'1':    0x31,
	'2':    0x32,
	'3':    0x33,
	'4':    0x34,
	'5':    0x35,
	'6':    0x36,
	'7':    0x37,
	'8':    0x38,
	'9':    0x39,
	'a':    0x41,
	'b':    0x42,
	'c':    0x43,
	'd':    0x44,
	'e':    0x45,
	'f':    0x46,
	'g':    0x47,
	'h':    0x48,
	'i':    0x49,
	'j':    0x4A,
	'k':    0x4B,
	'l':    0x4C,
	'm':    0x4D,
	'n':    0x4E,
	'o':    0x4F,
	'p':    0x50,
	'q':    0x51,
	'r':    0x52,
	's':    0x53,
	't':    0x54,
	'u':    0x55,
	'v':    0x56,
	'w':    0x57,
	'x':    0x58,
	'y':    0x59,
	'z':    0x5A,
	' ':    VK_SPACE,
	'\t':   VK_TAB,
	'\n':   VK_RETURN,
	'\r':   VK_RETURN,
	'\b':   VK_BACK,
	'\x1b': VK_ESCAPE,
	'!':    0x31, // Shift + 1
	'@':    0x32, // Shift + 2
	'#':    0x33, // Shift + 3
	'$':    0x34, // Shift + 4
	'%':    0x35, // Shift + 5
	'^':    0x36, // Shift + 6
	'&':    0x37, // Shift + 7
	'*':    0x38, // Shift + 8
	'(':    0x39, // Shift + 9
	')':    0x30, // Shift + 0
	'-':    0xBD,
	'_':    0xBD, // Shift + -
	'=':    0xBB,
	'+':    0xBB, // Shift + =
	'[':    0xDB,
	'{':    0xDB, // Shift + [
	']':    0xDD,
	'}':    0xDD, // Shift + ]
	'\\':   0xDC,
	'|':    0xDC, // Shift + \
	';':    0xBA,
	':':    0xBA, // Shift + ;
	'\'':   0xDE,
	'"':    0xDE, // Shift + '
	',':    0xBC,
	'<':    0xBC, // Shift + ,
	'.':    0xBE,
	'>':    0xBE, // Shift + .
	'/':    0xBF,
	'?':    0xBF, // Shift + /
	'`':    0xC0,
	'~':    0xC0, // Shift + `
}

// SpecialKeyMap maps special key names to their virtual key codes.
var SpecialKeyMap = map[string]uint8{
	"F1":                0x70,
	"F2":                0x71,
	"F3":                0x72,
	"F4":                0x73,
	"F5":                0x74,
	"F6":                0x75,
	"F7":                0x76,
	"F8":                0x77,
	"F9":                0x78,
	"F10":               0x79,
	"F11":               0x7A,
	"F12":               0x7B,
	"F13":               0x7C,
	"F14":               0x7D,
	"F15":               0x7E,
	"F16":               0x7F,
	"F17":               0x80,
	"F18":               0x81,
	"F19":               0x82,
	"F20":               0x83,
	"F21":               0x84,
	"F22":               0x85,
	"F23":               0x86,
	"F24":               0x87,
	"NUMLOCK":           VK_NUMLOCK,
	"SCROLL":            VK_SCROLL,
	"PAUSE":             VK_PAUSE,
	"BREAK":             VK_PAUSE,
	"INSERT":            VK_INSERT,
	"DELETE":            VK_DELETE,
	"HOME":              VK_HOME,
	"END":               VK_END,
	"PAGEUP":            VK_PRIOR,
	"PAGEDOWN":          VK_NEXT,
	"UP":                VK_UP,
	"DOWN":              VK_DOWN,
	"LEFT":              VK_LEFT,
	"RIGHT":             VK_RIGHT,
	"PRINTSCREEN":       VK_SNAPSHOT,
	"PRINT":             VK_PRINT,
	"EXECUTE":           VK_EXECUTE,
	"SNAPSHOT":          VK_SNAPSHOT,
	"HELP":              VK_HELP,
	"SELECT":            VK_SELECT,
	"NUMPAD0":           VK_NUMPAD0,
	"NUMPAD1":           VK_NUMPAD1,
	"NUMPAD2":           VK_NUMPAD2,
	"NUMPAD3":           VK_NUMPAD3,
	"NUMPAD4":           VK_NUMPAD4,
	"NUMPAD5":           VK_NUMPAD5,
	"NUMPAD6":           VK_NUMPAD6,
	"NUMPAD7":           VK_NUMPAD7,
	"NUMPAD8":           VK_NUMPAD8,
	"NUMPAD9":           VK_NUMPAD9,
	"MULTIPLY":          VK_MULTIPLY,
	"ADD":               VK_ADD,
	"SEPARATOR":         VK_SEPARATOR,
	"SUBTRACT":          VK_SUBTRACT,
	"DECIMAL":           VK_DECIMAL,
	"DIVIDE":            VK_DIVIDE,
	"LSHIFT":            VK_LSHIFT,
	"RSHIFT":            VK_RSHIFT,
	"LCONTROL":          VK_LCONTROL,
	"RCONTROL":          VK_RCONTROL,
	"LMENU":             VK_LMENU,
	"RMENU":             VK_RMENU,
	"BROWSER_BACK":      VK_BROWSER_BACK,
	"BROWSER_FORWARD":   VK_BROWSER_FORWARD,
	"BROWSER_REFRESH":   VK_BROWSER_REFRESH,
	"BROWSER_STOP":      VK_BROWSER_STOP,
	"BROWSER_SEARCH":    VK_BROWSER_SEARCH,
	"BROWSER_FAVORITES": VK_BROWSER_FAVORITES,
	"BROWSER_HOME":      VK_BROWSER_HOME,
	"VOLUME_MUTE":       VK_VOLUME_MUTE,
	"VOLUME_DOWN":       VK_VOLUME_DOWN,
	"VOLUME_UP":         VK_VOLUME_UP,
	"MEDIA_NEXT":        VK_MEDIA_NEXT_TRACK,
	"MEDIA_PREV":        VK_MEDIA_PREV_TRACK,
	"MEDIA_STOP":        VK_MEDIA_STOP,
	"MEDIA_PLAY_PAUSE":  VK_MEDIA_PLAY_PAUSE,
	"LAUNCH_MAIL":       VK_LAUNCH_MAIL,
	"LAUNCH_MEDIA":      VK_LAUNCH_MEDIA_SELECT,
	"LAUNCH_APP1":       VK_LAUNCH_APP1,
	"LAUNCH_APP2":       VK_LAUNCH_APP2,
}

// ModifierKey represents a modifier key state.
type ModifierKey struct {
	Shift   bool
	Control bool
	Alt     bool
	Meta    bool // Windows/Command key
}
