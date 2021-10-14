package zli

// We could get this from terminfo, but a look at this database and it seems
// that pretty much all modern terminals support these, so don't bother.

import (
	"fmt"
)

// Direction.
type Direction int

// Directions.
const (
	_ Direction = iota
	Up
	Down
	Left
	Right
)

func sendCSI(s string, a ...interface{}) { Stdout.Write([]byte("\x1b[" + fmt.Sprintf(s, a...))) }

// ErasesLine erases the entire line and puts the cursor at the start of the
// line.
func EraseLine() { sendCSI("2K\r") }

// type Display struct {}
//
// EraseDisplayBelow 0J
// EraseDisplayAbove 1J
// EraseDisplayAll 2J
//
// InsertLines nL
// DeleteLines nM
// DeleteCharacters nP
//
// EraseToRight() 0K
// EraseToLeft() 1K
// EraseLine() 2K
//
// CursorStyle
//   0, 1 blink blocka
//   2 steady block
//   3 blink underl
//   4 steady underl
//   5 blink bar
//   6 steady bar
//
// s -> save cursor
// u -> restore cursor
//
//

// func WriteLine(s string) {
// fmt.Print()
// senfCSI("0K")
// }

// ReplaceLine replaces the current line.
func ReplaceLine(a ...interface{}) {
	EraseLine()
	fmt.Fprint(Stdout, a...)
}

// ReplaceLinef replaces the current line.
func ReplaceLinef(s string, a ...interface{}) {
	EraseLine()
	fmt.Fprintf(Stdout, s, a...)
}

// Clear the screen and put the cursor at 1×1.
func ClearScreen() {
	sendCSI("2J")
	CursorSet(1, 1)
}

// SetCursor sets the cursor to a specific position.
//
// Rows and columns are numbered from 1.
func CursorSet(row, col int) { sendCSI("%d;%dH", row, col) }

// ShowCursor sets the cursor visibility.
func CursorShow(show bool) { sendCSI(map[bool]string{true: "?25h", false: "?25l"}[show]) }

// CursorMove moves the cursor n cells in a direction.
func CursorMove(n int, dir Direction) {
	switch dir {
	case Up:
		sendCSI("%dA", n)
	case Down:
		sendCSI("%dB", n)
	case Right:
		sendCSI("%dC", n)
	case Left:
		sendCSI("%dD", n)
	}
}
