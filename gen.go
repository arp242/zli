//go:build go_gen_only
// +build go_gen_only

package main

import (
	"fmt"
	"reflect"

	"zgo.at/zli"
)

func main() {
	var infos []*zli.Terminfo
	for _, term := range models {
		ti, err := zli.NT(term)
		if err != nil {
			//fmt.Println("Error reading", m, ":", err)
			continue
		}
		infos = append(infos, ti)
	}

	// TODO: use this.
	aliases := make(map[string]string)
	for i, ti := range infos {
		for _, ti2 := range infos[i+1:] {
			if reflect.DeepEqual(ti.Keys(), ti2.Keys()) && reflect.DeepEqual(ti.Caps(), ti2.Caps()) {
				aliases[ti2.Name] = ti.Name
			}
		}
	}

	fmt.Println("package zli")
	fmt.Println("\nvar builtinTerms = map[string]*Terminfo{")
	for _, ti := range infos {
		fmt.Printf("\t%q: &Terminfo{\n", ti.Name)
		fmt.Printf("\t\tName: %q,\n", ti.Name)
		fmt.Printf("\t\tDesc: %q,\n", ti.Desc+" (built-in)")
		fmt.Printf("\t\tAliases: %#v,\n", ti.Aliases)

		fmt.Println("\t\tkeys: map[string]Key{")
		for _, k := range ti.KeysSorted() {
			fmt.Print(k)
		}
		fmt.Println("\t\t},")

		fmt.Println("\t\tcaps: map[string]Cap{")
		for _, k := range ti.CapsSorted() {
			fmt.Print(k)
		}
		fmt.Println("\t\t},")

		fmt.Println("\t},")
	}
	fmt.Println("}")
}

var models = []string{
	"aixterm",
	"alacritty",
	"ansi",
	"beterm",
	"cygwin",
	"dtterm",
	//"eterm,eterm-color|emacs",
	"eterm", "eterm-color",

	//"gnome,gnome-256color",
	"gnome", "gnome-256color",

	"hpterm",
	//"konsole,konsole-256color",
	"konsole", "konsole-256color",

	"kterm",
	"linux",
	"pcansi",
	//"rxvt,rxvt-256color,rxvt-88color,rxvt-unicode,rxvt-unicode-256color",
	"rxvt", "rxvt-256color", "rxvt-88color", "rxvt-unicode", "rxvt-unicode-256color",

	//"screen,screen-256color",
	"screen", "screen-256color",

	//"st,st-256color|simpleterm",
	"st", "st-256color",

	"termite",
	"tmux",
	"vt100",
	"vt102",
	"vt220",
	"vt320",
	"vt400",
	"vt420",
	"xfce",

	//"xterm,xterm-88color,xterm-256color",
	"xterm", "xterm-88color", "xterm-256color",

	"xterm-kitty",
}
