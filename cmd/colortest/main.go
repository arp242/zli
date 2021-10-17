// Command colortest prints an overview of colors for testing.
package main

import (
	"fmt"
	"os"
	"strconv"
	"strings"

	"zgo.at/zli"
)

var std = []zli.Color{zli.Black, zli.Red, zli.Green, zli.Yellow, zli.Blue, zli.Magenta, zli.Cyan, zli.White,
	zli.Black.Brighten(1), zli.Red.Brighten(1), zli.Green.Brighten(1), zli.Yellow.Brighten(1),
	zli.Blue.Brighten(1), zli.Magenta.Brighten(1), zli.Cyan.Brighten(1), zli.White.Brighten(1)}

func ranges(n ...int) []uint8 {
	if len(n)%2 != 0 {
		panic("X")
	}

	var rng []uint8
	for j := 0; j < len(n); j += 2 {
		for i := n[j]; i <= n[j+1]; i++ {
			rng = append(rng, uint8(i))
		}
	}
	return rng
}

func main() {
	zli.WantColor = true
	bg := false
	if len(os.Args) > 1 {
		switch os.Args[1] {
		case "bg":
			bg = true
		case "brighten":
			if len(os.Args) != 3 {
				zli.Fatalf("specify a color:\n  colortest brighten 26\n  colortest brighten #123123")
			}
			brightTest(os.Args[2])
			return
		default:
			zli.Fatalf("unknown command; supported commands: 'bg', 'brighten'")
		}
	}
	toBg := func(c zli.Color) zli.Color {
		// TODO: also add something to get a good-looking contrast color:
		// c2 := c.Contrast()
		if bg {
			return c.Bg()
		}
		return c
	}

	fmt.Print("Attrs:  ")
	fmt.Print("Bold      ", zli.Colorize("11", zli.Bold), " ")
	fmt.Print("Faint      ", zli.Colorize("22", zli.Faint), " ")
	fmt.Print("Italic  ", zli.Colorize("33", zli.Italic), " ")
	fmt.Print("Underline ", zli.Colorize("44", zli.Underline), " ")
	fmt.Print("\n        ")
	fmt.Print("BlinkSlow ", zli.Colorize("55", zli.BlinkSlow), " ")
	fmt.Print("BlinkRapid ", zli.Colorize("66", zli.BlinkRapid), " ")
	fmt.Print("Reverse ", zli.Colorize("77", zli.ReverseVideo), " ")
	fmt.Print("Concealed ", zli.Colorize("88", zli.Concealed), " ")
	fmt.Print("CrossedOut ", zli.Colorize("99", zli.CrossedOut), " ")

	fmt.Print("\n")
	fmt.Println("                       ┌ Regular ──────────────┐  ┌ Bright ─────────────┐")
	fmt.Print("Standard colors:       ")
	for i, c := range std {
		zli.Colorf("%-3d", toBg(c), i)
	}

	fmt.Print("\nStandard colors (256): ")
	for i := uint8(0); i <= 16; i++ {
		zli.Colorf("%-3d", toBg(zli.Color256(i)), i)
	}

	fmt.Print("\n\n")
	//fmt.Println("\n\n216 colors:")
	for _, i := range ranges(16, 33, 52, 69, 88, 105, 124, 141, 160, 177, 196, 213) {
		if i > 16 && (i-16)%18 == 0 {
			fmt.Println("")
		}
		zli.Colorf("%-4d", toBg(zli.Color256(i)), i)
	}
	for _, i := range ranges(34, 51, 70, 87, 106, 123, 142, 159, 178, 195, 214, 231) {
		if i > 16 && (i-16)%18 == 0 {
			fmt.Println("")
		}
		zli.Colorf("%-4d", toBg(zli.Color256(i)), i)
	}

	fmt.Print("\nGrey-tones: ")
	for i := 232; i <= 255; i++ {
		if i == 244 {
			fmt.Print("\n            ")
		}
		zli.Colorf("%-4d", toBg(zli.Color256(uint8(i))), i)
	}
	fmt.Printf("\nRun '%s bg' to set background instead of foreground.\n", zli.Program())
	fmt.Printf("Run '%s brighten [color]' to test the Brighten() method.\n", zli.Program())
}

func brightTest(name string) {
	var c zli.Color
	if name[0] == '#' {
		c = zli.ColorHex(name)
		if c == zli.ColorError {
			zli.Fatalf("errror parsing RGB")
		}
	} else {
		n, err := strconv.ParseUint(name, 10, 8)
		zli.F(err)
		c = zli.Color256(uint8(n))
	}
	c = c.Bg()

	br := make([]zli.Color, 0, 32)
	for i := 0; ; i++ {
		b := c.Brighten(i)
		if i > 1 && b == br[len(br)-1] {
			break
		}
		br = append(br, b)
	}

	dr := make([]zli.Color, 0, 32)
	for i := 0; ; i-- {
		b := c.Brighten(i)
		if i < -1 && b == dr[len(dr)-1] {
			break
		}
		dr = append(dr, b)
	}

	w, _, _ := zli.TerminalSize(os.Stdout.Fd())
	if w <= 0 {
		w = 76
	}
	w -= 12

	fmt.Printf("Brighten: %s%s\n", pr(br, w), zli.Reset)
	fmt.Printf("Darken:   %s%s\n", pr(dr, w), zli.Reset)
}

func pr(t []zli.Color, w int) string {
	pad := strings.Repeat(" ", 10)
	out := ""
	for i, c := range t {
		out += c.String() + " "
		if i > 0 && (i+1)%w == 0 {
			out += zli.Reset.String() + "\n" + pad
		}
	}

	return out + zli.Reset.String() +
		fmt.Sprintf("\n%s%s → %s in %d steps", pad, cname(t[0]), cname(t[len(t)-1]), len(t)-1)
}

func cname(c zli.Color) string {
	if c&zli.ColorMode256Bg != 0 {
		return fmt.Sprintf("%d", int(c>>zli.ColorOffsetBg))
	}
	c = c >> zli.ColorOffsetBg
	return fmt.Sprintf("#%02x%02x%02x", int(c%256), int(c>>8%256), int(c>>16%256))
}
