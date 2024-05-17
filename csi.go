package zli

import "fmt"

// ErasesLine erases the entire line and puts the cursor at the start of the
// line.
func EraseLine() { Stdout.Write([]byte("\x1b[2K\r")) }

// ReplaceLine replaces the current line.
func ReplaceLine(a ...any) {
	EraseLine()
	fmt.Fprint(Stdout, a...)
}

// ReplaceLinef replaces the current line.
func ReplaceLinef(s string, a ...any) {
	EraseLine()
	fmt.Fprintf(Stdout, s, a...)
}
