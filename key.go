package zli

import (
	"fmt"
	"strings"
)

// Key represents a keypress. This is formatted as follows:
//
//   - First 32 bits   → rune (int32)
//   - Next 16 bits    → Named key constant.
//   - Bits 49-61      → Currently unused.
//
// And the last three bits are flags for modifier keys:
//
//   - bit 62          → Alt
//   - bit 63          → Ctrl
//   - bit 64          → Shift
//
// The upshot of this is that you can now use a single value to test for all
// combinations:
//
//    switch Key(0x61) {
//    case 'a':                         // 'a' w/o modifiers
//    case 'a' | key.Ctrl:              // 'a' with control
//    case 'a' | key.Ctrl | key.Shift:  // 'a' with shift and control
//
//    case key.KeyUp:                   // Arrow up
//    case key.KeyUp | key.Ctrl:        // Arrow up with control
//    }
//
// Which is nicer than using two or three different variables to signal various
// things.
type Key uint64

// Shift reports if the Shift modifier is set.
func (k Key) Shift() bool { return k&Shift != 0 }

// Ctrl reports if the Ctrl modifier is set.
func (k Key) Ctrl() bool { return k&Ctrl != 0 }

// Alt reports if the Alt modifier is set.
func (k Key) Alt() bool { return k&Alt != 0 }

// Named reports if this is a named key.
func (k Key) Named() bool {
	_, ok := names[k&^Modmask]
	return ok
}

// Valid reports if this key is valid.
func (k Key) Valid() bool { return k&^Modmask <= (1<<31) || k.Named() }

// Name gets the key name. This doesn't show if any modifiers are set; use
// String() for that.
func (k Key) Name() string {
	k &^= Modmask

	n, ok := names[k]
	if ok {
		return n
	}
	if !k.Valid() {
		return fmt.Sprintf("Unknown key: 0x%x", uint64(k))
	}
	return fmt.Sprintf("%c", rune(k))
}

func (k Key) String() string {
	var b strings.Builder
	b.Grow(8)
	b.WriteRune('<')
	if k.Shift() {
		b.WriteString("S-")
	}
	if k.Ctrl() {
		b.WriteString("C-")
	}
	if k.Alt() {
		b.WriteString("A-")
	}
	b.WriteString(k.Name())
	b.WriteRune('>')
	return b.String()
}

// Modifiers
const (
	Shift   = 1 << 63
	Ctrl    = 1 << 62
	Alt     = 1 << 61
	Modmask = Shift | Ctrl | Alt
)

// Useful control characters.
const (
	KeyNull       = Key(0x00) // NUL
	KeyBackspace  = Key(0x08) // BS
	KeyTab        = Key(0x09) // HT
	KeyLinefeed   = Key(0x0a) // LF
	KeyEnter      = Key(0x0d) // CR
	KeyEsc        = Key(0x1b) // ESC
	KeyBackspace2 = Key(0x7f) // DEL
)

var names = map[Key]string{
	UnknownSequence: "Unknown escape sequence",

	KeyNull: "Null", KeyBackspace: "Backspace", KeyTab: "Tab", KeyLinefeed: "LF",
	KeyEnter: "Enter", KeyEsc: "Esc", KeyBackspace2: "Backspace2",

	KeyBacktab: "Backtab", KeyDelete: "Delete", KeyInsert: "Insert", KeyClear: "Clear",
	KeyExit: "Exit", KeyCancel: "Cancel", KeyPause: "Pause", KeyPrint: "Print",
	KeyHome: "Home", KeyEnd: "End", KeyPgDn: "PgDn", KeyPgUp: "PgUp",
	KeyUp: "Up", KeyDown: "Down", KeyLeft: "Left", KeyRight: "Right",
	KeyUpLeft: "UpLeft", KeyUpRight: "UpRight", KeyDownLeft: "DownLeft",
	KeyDownRight: "DownRight", KeyCenter: "Center",

	KeyF1: "F1", KeyF2: "F2", KeyF3: "F3", KeyF4: "F4", KeyF5: "F5", KeyF6: "F6", KeyF7: "F7", KeyF8: "F8",
	KeyF9: "F9", KeyF10: "F10", KeyF11: "F11", KeyF12: "F12", KeyF13: "F13", KeyF14: "F14", KeyF15: "F15", KeyF16: "F16",
	KeyF17: "F17", KeyF18: "F18", KeyF19: "F19", KeyF20: "F20", KeyF21: "F21", KeyF22: "F22", KeyF23: "F23", KeyF24: "F24",
}

// Named key constants.
const (
	UnknownSequence Key = iota + (1 << 32)
	KeyUp
	KeyDown
	KeyRight
	KeyLeft
	KeyUpLeft
	KeyUpRight
	KeyDownLeft
	KeyDownRight
	KeyCenter
	KeyPgUp
	KeyPgDn
	KeyHome
	KeyEnd
	KeyInsert
	KeyDelete
	KeyHelp
	KeyExit
	KeyClear
	KeyCancel
	KeyPrint
	KeyPause
	KeyBacktab
	KeyF1
	KeyF2
	KeyF3
	KeyF4
	KeyF5
	KeyF6
	KeyF7
	KeyF8
	KeyF9
	KeyF10
	KeyF11
	KeyF12
	KeyF13
	KeyF14
	KeyF15
	KeyF16
	KeyF17
	KeyF18
	KeyF19
	KeyF20
	KeyF21
	KeyF22
	KeyF23
	KeyF24
)
