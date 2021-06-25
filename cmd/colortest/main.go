// Command colortest prints an overview of colors for testing.
package main

import (
	"fmt"
	"os"

	"zgo.at/zli"
)

var std = []zli.Color{zli.Black, zli.Red, zli.Green, zli.Yellow, zli.Blue,
	zli.Magenta, zli.Cyan, zli.White, zli.BrightBlack, zli.BrightRed,
	zli.BrightGreen, zli.BrightYellow, zli.BrightBlue, zli.BrightMagenta,
	zli.BrightCyan, zli.BrightWhite}

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
	bg := len(os.Args) > 1
	toBg := func(c zli.Color) zli.Color {
		// TODO: also add something to get a good-looking contrast color:
		// c2 := c.Contrast()
		if bg {
			return c.Bg()
		}
		return c
	}

	fmt.Print("Attributes:  ")
	fmt.Print("Bold       ", zli.Colorf("XX", zli.Bold), " ")
	fmt.Print("Faint   ", zli.Colorf("XX", zli.Faint), " ")
	fmt.Print("Italic    ", zli.Colorf("XX", zli.Italic), " ")
	fmt.Print("Underline  ", zli.Colorf("XX", zli.Underline), " ")
	fmt.Print("BlinkSlow ", zli.Colorf("XX", zli.BlinkSlow), " ")
	fmt.Print("\n             ")
	fmt.Print("BlinkRapid ", zli.Colorf("XX", zli.BlinkRapid), " ")
	fmt.Print("Reverse ", zli.Colorf("XX", zli.ReverseVideo), " ")
	fmt.Print("Concealed ", zli.Colorf("XX", zli.Concealed), " ")
	fmt.Print("CrossedOut ", zli.Colorf("XX", zli.CrossedOut), " ")

	fmt.Print("\n\nStandard colors:       ")
	for i, c := range std {
		fmt.Print(zli.Colorf(fmt.Sprintf("%-3d", i), toBg(c)))
	}

	fmt.Print("\nStandard colors (256): ")
	for i := uint8(0); i <= 16; i++ {
		fmt.Print(zli.Colorf(fmt.Sprintf("%-3d", i), toBg(zli.Color256(i))))
	}

	fmt.Print("\n\n")
	//fmt.Println("\n\n216 colors:")
	for _, i := range ranges(16, 33, 52, 69, 88, 105, 124, 141, 160, 177, 196, 213) {
		if i > 16 && (i-16)%18 == 0 {
			fmt.Println("")
		}
		fmt.Print(zli.Colorf(fmt.Sprintf("%-4d", i), toBg(zli.Color256(i))))
	}
	for _, i := range ranges(34, 51, 70, 87, 106, 123, 142, 159, 178, 195, 214, 231) {
		if i > 16 && (i-16)%18 == 0 {
			fmt.Println("")
		}
		fmt.Print(zli.Colorf(fmt.Sprintf("%-4d", i), toBg(zli.Color256(i))))
	}

	fmt.Print("\nGrey-tones: ")
	for i := 232; i <= 255; i++ {
		if i == 244 {
			fmt.Print("\n            ")
		}
		fmt.Print(zli.Colorf(fmt.Sprintf("%-4d", i), toBg(zli.Color256(uint8(i)))))
	}
	fmt.Printf("\n\nRun as '%s bg' to set background instead of foreground.\n", zli.Program())
}
