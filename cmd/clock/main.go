package main

import (
	"os"
	"strings"
	"time"

	"zgo.at/zli"
)

var width, height = func() (int, int) {
	w, h, err := zli.TerminalSize(os.Stdout.Fd())
	zli.F(err)
	return w, h
}()

var (
	graphs = []string{zero, one, two, three, four, five, six, seven, eight, nine}
	cell   = zli.Red.Bg().String() + " " + zli.Reset.String()
)

func main() {
	defer zli.MakeRaw(true)()
	zli.TerminalSizeChange(func() {
		var err error
		width, height, err = zli.TerminalSize(os.Stdout.Fd())
		zli.F(err)
	})

	var (
		in = make(chan string)
		t  = time.NewTicker(time.Second)
	)
	go func() {
		b := make([]byte, 6)
		for {
			n, _ := os.Stdin.Read(b)
			in <- string(b[:n])
		}
	}()

	for {
		select {
		case k := <-in:
			switch k {
			case "q", "\x03", "\x1b":
				return
			}
		case <-t.C:
			t := time.Now()
			n := []string{graphs[t.Hour()/10], graphs[t.Hour()%10], colon, graphs[t.Minute()/10], graphs[t.Minute()%10]}

			r, c := height/2-1, width/2-(25/2)
			zli.EraseScreen()
			zli.To(r, c, "")
			for i, nn := range n {
				for j, line := range strings.Split(nn, "\n") {
					line = strings.ReplaceAll(line, "X", cell)
					zli.To(r+j, c+(6*i), line)
				}
			}
		}
	}
}

// Simple 4x5 font

var zero = `
XXXX
X  X
X  X
X  X
XXXX`[1:]
var one = `
 XX
X X
  X
  X
 XXX`[1:]
var two = `
XXXX
   X
 XX
X
XXXX`[1:]
var three = `
XXXX
   X
XXXX
   X
XXXX`[1:]
var four = `
  XX
 X X
X  X
XXXX
   X`[1:]
var five = `
XXXX
X
XXXX
   X
XXXX`[1:]
var six = `
XXXX
X
XXXX
X  X
XXXX`[1:]
var seven = `
XXXX
   X
  X
 X
X`[1:]
var eight = `
XXXX
X  X
XXXX
X  X
XXXX`[1:]
var nine = `
XXXX
X  X
XXXX
   X
XXXX`[1:]
var colon = `

 XX

 XX`[1:]
