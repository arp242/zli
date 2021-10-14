package zli

import (
	"fmt"
	"testing"
)

func TestTerminfo(t *testing.T) {
	ti, _ := NewTerminfo()
	fmt.Println(ti)

	// fmt.Printf("%s\n", ti.FindKey("\x1bOP"))    // F1
	// fmt.Printf("%s\n", ti.FindKey("\x1bOA"))    // Up
	// fmt.Printf("%s\n", ti.FindKey("\x1b[1;5A")) // C-Up

	// fmt.Printf("%s\n", ti.FindKey("\x1b[15~"))   // F5
	// fmt.Printf("%s\n", ti.FindKey("\x1b[15;2~")) // F5
}
