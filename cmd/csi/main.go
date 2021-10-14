// Command csi tests CSI escape sequences for demo purposes.
package main

import (
	"fmt"
	"os"
	"strconv"
	"time"

	"zgo.at/zli"
)

func main() {
	defer zli.CursorShow(true)
	s := 200
	if len(os.Args) > 1 {
		// Just report current position.
		if os.Args[1] == "pos" {
			r, c, _ := zli.CursorPosition()
			fmt.Printf("%d×%d\n", r, c)
			return
		}
		s, _ = strconv.Atoi(os.Args[1])
	}

	steps := []func(){
		func() { zli.ClearScreen() },
		func() { fmt.Println("Hello!") },
		func() { fmt.Println("Hella!") },
		func() { zli.CursorSet(2, 5) },
		func() { zli.CursorShow(false) },
		func() { fmt.Print("o") },
		func() { zli.CursorShow(true) },
		func() { zli.CursorMove(2, zli.Right) },
		func() { zli.CursorMove(2, zli.Down) },
		func() { zli.CursorMove(2, zli.Left) },
		func() { zli.CursorMove(2, zli.Up) },
		func() { zli.CursorMove(1, zli.Down) },
		func() { zli.EraseLine() },
		func() {
			r, c, _ := zli.CursorPosition()
			zli.CursorMove(1, zli.Right)
			r2, c2, _ := zli.CursorPosition()
			zli.CursorMove(1, zli.Left)
			fmt.Printf("pos: %d×%d; %d×%d\n", r, c, r2, c2)
		},
		func() { fmt.Println("Done") },
	}
	for _, f := range steps {
		f()
		time.Sleep(time.Duration(s) * time.Millisecond)
	}
}
