// Command input tests keyboard input
package main

import (
	"fmt"
	"os"

	"zgo.at/zli"
)

func main() {
	restore, err := zli.RawTerminal()
	zli.F(err)

	// TODO: really want this on ^C too.
	defer restore()

	// go func() {
	// 	r, err := io.ReadAll(os.Stdin)
	// 	zli.F(err)
	// 	fmt.Print("READ => ", len(r), r, "\r\n")
	// }()

	keys, err := zli.ReadKeys()
	zli.F(err)

	fmt.Println("Press q to quit\r")
	for {
		key := <-keys
		zli.F(key.Err)

		switch key.String {
		case "q":
			return
		}
		fmt.Printf("Pressed %q \t ", key.String)
		fmt.Print(zli.TerminalSize(os.Stdout.Fd()))
		fmt.Print(" \t ")
		fmt.Print(zli.CursorPosition())
		fmt.Print("\r\n")
	}
}
