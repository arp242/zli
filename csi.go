package zli

import (
	"cmp"
	"fmt"
)

// Erase line from the cursor to the end, leaving the cursor in the current
// position.
func Erase() { fmt.Fprint(Stdout, "\x1b[K") }

// Replacef replaces the current line.
func Replacef(text string, a ...any) { fmt.Fprint(Stdout, "\x1b[K\r"); fmt.Fprintf(Stdout, text, a...) }

// EraseScreen erases the entire screen and puts the cursor at position 1, 1.
func EraseScreen() { fmt.Fprint(Stdout, "\x1b[0;0H\x1b[J") }

// HideCursor hides the cursor, returning a function to display it again.
func HideCursor() func() {
	fmt.Fprint(Stdout, "\x1b[?25l")
	return func() { fmt.Fprint(Stdout, "\x1b[?25h") }
}

func max[T cmp.Ordered](x T, y ...T) T {
	m := x
	for _, yy := range y {
		if yy > m {
			m = yy
		}
	}
	return m
}

// To sets the cursor at the given position and prints the text.
//
// The top-left corner is 1, 1.
func To(row, col int, text string, a ...any) {
	fmt.Fprintf(Stdout, "\x1b[%d;%dH", max(row, 1), max(col, 1))
	if text != "" {
		fmt.Fprintf(Stdout, text, a...)
	}
}

// Move the cursor relative to current position and print the text.
//
// Positive values move down or right, negative values move up or left, and 0
// doesn't move anything.
func Move(row, col int, text string, a ...any) {
	if row < 0 {
		fmt.Fprintf(Stdout, "\x1b[%dA", -row)
	} else if row > 0 {
		fmt.Fprintf(Stdout, "\x1b[%dB", row)
	}
	if col > 0 {
		fmt.Fprintf(Stdout, "\x1b[%dC", col)
	} else if col < 0 {
		fmt.Fprintf(Stdout, "\x1b[%dD", -col)
	}
	if text != "" {
		fmt.Fprintf(Stdout, text, a...)
	}
}

// Modify text, inserting or deleting lines, and print the text.
//
// On positive values it will insert text, moving existing text below (for
// lines) or to the right (for characters). On negative values it will delete
// text, moving existing text upwards (for lines) or to the left (for
// characters). On 0 nothing is modified.
func Modify(line, char int, text string, a ...any) {
	if line > 0 {
		fmt.Fprintf(Stdout, "\x1b[%dL", line)
	} else if line < 0 {
		fmt.Fprintf(Stdout, "\x1b[%dM", -line)
	}
	if char > 0 {
		fmt.Fprintf(Stdout, "\x1b[%d@", char)
	} else if char < 0 {
		fmt.Fprintf(Stdout, "\x1b[%dP", -char)
	}
	if text != "" {
		fmt.Fprintf(Stdout, text, a...)
	}
}
