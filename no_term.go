//go:build no_term
// +build no_term

// This file contains shims to prevent depending on x/term (and by extension,
// x/sys). Especially x/sys is fairly large, and makes vendoring zli tricky.

package zli

var (
	IsTerminal   = func(fd uintptr) bool { return true }
	TerminalSize = func(fd uintptr) (width, height int, err error) { return 0, 0, nil }
	WantColor    = true
)

func AskPassword(minlen int) (string, error)   { panic("not implemented; compiled with no_term flag") }
func RawTerminal(fd int) (func() error, error) { panic("not implemented; compiled with no_term flag") }
func CursorPosition() (int, int, error)        { panic("not implemented; compiled with no_term flag") }
