package zli

import (
	"fmt"
	"strconv"
)

// ErasesLine erases part of the line; it does not change the cursor position.
//
// Values for n:
//
//    0  Clear from cursor to end of line.
//    1  Clear from cursor to start of line.
//    2  Clear entire line.
func EraseLine(n int) {
	fmt.Fprint(Stdout, "\x1b["+strconv.Itoa(n)+"K")
}
