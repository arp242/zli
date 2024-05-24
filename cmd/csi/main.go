// Command csi tests CSI escapes.
package main

import (
	"fmt"
	"os"

	"zgo.at/zli"
)

// Simple example using CSI escape sequences to build a basic TUI.
//
// All of this should be safe to use on all modern terminals; see:
// http://www.arp242.net/safeterm.html
//
// For more advanced stuff, use: https://github.com/arp242/termfo
func main() {
	// Put terminal in raw mode, hide cursor, and get size.
	defer zli.MakeRaw(true)()
	width, height, err := zli.TerminalSize(os.Stdout.Fd())
	zli.F(err)

	// Keep track of currently selected row.
	sel := 2

	// Do a full redraw.
	redraw := func() {
		zli.EraseScreen()
		fmt.Print(zli.Bold, "q to exit; j/k to move, space/enter to select", zli.Reset)
		for i := 2; i < 24; i++ {
			zli.To(i, 1, "  line number %d", i)
		}
		zli.To(sel, 1, "")
		zli.Colorf("→", zli.Bold)
	}
	redraw()

	// Redraw on terminal size changes.
	go func() {
		for ch := zli.TerminalSizeChange(); ; <-ch {
			width, height, err = zli.TerminalSize(os.Stdout.Fd())
			zli.F(err)
			redraw()
		}
	}()

	// Main eventloop to read keys.
	b := make([]byte, 3)
	for {
		n, _ := os.Stdin.Read(b)
		switch k := string(b[:n]); k {
		case "q", "\x03": // ^C
			return

		case "\f": // ^L
			redraw()

		case "j", "\x1b[B", "\x1bOB", // Down
			"k", "\x1b[A", "\x1bOA": // Up
			zli.Move(0, -1, " ") // Erase previous arrow.

			switch k {
			case "j", "\x1b[B", "\x1bOB":
				sel = min(sel+1, 23)
			default:
				sel = max(sel-1, 2)
			}
			zli.To(sel, 1, "")
			zli.Colorf("→", zli.Bold)

		case " ", "\r": // Space, Enter
			x, y := width/2-11, height/2-2
			zli.To(y+0, x, "┌────────────────────┐")
			zli.To(y+1, x, "│                    │")
			zli.To(y+2, x, "│  %sSelected line %-2d%s  │", zli.Bold, sel, zli.Reset)
			zli.To(y+3, x, "│                    │")
			zli.To(y+4, x, "└────────────────────┘")
			zli.To(sel, 1, "")

			// Wait for any key and redraw.
			os.Stdin.Read(b)
			redraw()
		}
	}
}

func max(x int, y ...int) int {
	m := x
	for _, yy := range y {
		if yy > m {
			m = yy
		}
	}
	return m
}
func min(x int, y ...int) int {
	m := x
	for _, yy := range y {
		if yy < m {
			m = yy
		}
	}
	return m
}
