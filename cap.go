package zli

// Cap represents a terminal capability.
type Cap uint16

func (c Cap) String() string { return capNames[c] }

// After adding something here, you have to add the corresponding index in a
// terminfo file to ti_funcs. The values can be taken from (ncurses) term.h.
//
// terminfo_builtin.go also needs adjusting with the new values.
const (
	_ Cap = iota
	CapEnterCA
	CapExitCA
	CapShowCursor
	CapHideCursor
	CapClearScreen
	CapSGR0
	CapUnderline
	CapBold
	CapHidden
	CapBlink
	CapDim
	CapCursive
	CapReverse
	CapEnterKeypad
	CapExitKeypad
)

// Not actually used.
var capNames = map[Cap]string{
	CapEnterCA:     "EnterCA",
	CapExitCA:      "ExitCA",
	CapShowCursor:  "ShowCursor",
	CapHideCursor:  "HideCursor",
	CapClearScreen: "ClearScreen",
	CapSGR0:        "SGR0",
	CapUnderline:   "Underline",
	CapBold:        "Bold",
	CapHidden:      "Hidden",
	CapBlink:       "Blink",
	CapDim:         "Dim",
	CapCursive:     "Cursive",
	CapReverse:     "Reverse",
	CapEnterKeypad: "EnterKeypad",
	CapExitKeypad:  "ExitKeypad",
}
