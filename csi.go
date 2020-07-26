package zli

import "fmt"

// ErasesLine erases the entire line and puts the cursor at the start of the
// line.
func EraseLine() { Stdout.Write([]byte("\x1b[2K\r")) }

// ReplaceLine replaces the current line.
func ReplaceLine(a ...interface{}) {
	EraseLine()
	fmt.Print(a...)
}

// ReplaceLinef replaces the current line.
func ReplaceLinef(s string, a ...interface{}) {
	EraseLine()
	fmt.Printf(s, a...)
}
