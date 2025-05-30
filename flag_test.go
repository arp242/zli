package zli_test

import (
	"errors"
	"fmt"
	"os"
	"regexp"
	"strings"
	"testing"

	"zgo.at/zli"
)

func ExampleFlags() {
	// Create new flags from os.Args.
	f := zli.NewFlags([]string{"example", "-vv", "-f=csv", "-a", "xx", "yy"})

	// Add a string, bool, and "counter" flag.
	var (
		verbose = f.IntCounter(0, "v", "verbose")
		all     = f.Bool(false, "a", "all")
		format  = f.String("", "f", "format")
	)

	// Shift the first argument (i.e. os.Args[1], if any, empty string if there
	// isn't). Useful to get the "subcommand" name. This works before and after
	// Parse().
	switch f.Shift() {
	case "help":
		// Run help
	case "install":
		// Run install
	case "":
		// Error: need a command (or just print the usage)
	default:
		// Error: Unknown command
	}

	// Parse the shebang!
	err := f.Parse()
	if err != nil {
		// Print error, usage.
	}

	// You can check if the flag was present on the CLI with Set(). This way you
	// can distinguish between "was an empty value passed" // (-format '') and
	// "this flag wasn't on the CLI".
	if format.Set() {
		fmt.Println("Format was set to", format.String())
	}

	// The IntCounter adds 1 for every time the -v flag is on the CLI.
	if verbose.Int() > 1 {
		// ...Print very verbose info.
	} else if verbose.Int() > 0 {
		// ...Print less verbose info.
	}

	// Just a bool!
	fmt.Println("All:", all.Bool())

	// f.Args is set to everything that's not a flag or argument.
	fmt.Println("Remaining:", f.Args)

	// Output:
	// Format was set to csv
	// All: true
	// Remaining: [xx yy]
}

func ExampleFlags_ShiftCommand() {
	f := zli.NewFlags([]string{"prog", "i"})

	// Known commands.
	commands := []string{"help", "version", "verbose", "install"}

	switch cmd, err := f.ShiftCommand(commands...); cmd {
	// On error the return value is "" and err is set to something useful; for
	// example:
	//
	//    % prog
	//    prog: no command given
	//
	//    % prog hello
	//    prog: unknown command: "hello"
	//
	//    % prog v
	//    prog: ambigious command: "v"; matches: "verbose", "version"
	case "":
		zli.F(err)

	// The full command is returned, e.g. "prog h" will return "help".
	case "help":
		fmt.Println("cmd: help")
	case "version":
		fmt.Println("cmd: version")
	case "install":
		fmt.Println("cmd: install")
	}

	// Output:
	// cmd: install
}

func TestFlags(t *testing.T) {
	tests := []struct {
		name    string
		args    []string
		flags   func(*zli.Flags) []any
		want    string
		wantErr string
	}{
		// No arguments, no problem.
		{"nil args", nil,
			func(f *zli.Flags) []any {
				return []any{f.Bool(false, "b")}
			}, `
			bool 1 → false
			args   → 0 []
			`, ""},
		{"empty args", []string{},
			func(f *zli.Flags) []any {
				return []any{f.Bool(false, "b")}
			}, `
			bool 1 → false
			args   → 0 []
			`, ""},
		{"prog name", []string{"progname"},
			func(f *zli.Flags) []any {
				return []any{f.Bool(false, "b")}
			}, `
			bool 1 → false
			args   → 0 []
			`, ""},

		// Get positional arguments
		{"1 arg", []string{"progname", "pos1"},
			func(f *zli.Flags) []any {
				return []any{f.String("", "s")}
			}, `
			string 1 → ""
			args     → 1 [pos1]
			`, ""},
		{"args with space", []string{"progname", "pos1", "pos 2", "pos\n3"},
			func(f *zli.Flags) []any {
				return []any{f.String("", "s")}
			}, `
			string 1 → ""
			args     → 3 [pos1 pos 2 pos
							3]
			`, ""},
		{"after flag", []string{"progname", "-s", "arg", "pos 1", "pos 2"},
			func(f *zli.Flags) []any {
				return []any{f.String("", "s")}
			}, `
			string 1 → "arg"
			args     → 2 [pos 1 pos 2]
			`, ""},
		{"before flag", []string{"progname", "pos 1", "pos 2", "-s", "arg"},
			func(f *zli.Flags) []any {
				return []any{f.String("", "s")}
			}, `
			string 1 → "arg"
			args     → 2 [pos 1 pos 2]
			`, ""},
		{"before and after flag", []string{"progname", "pos 1", "-s", "arg", "pos 2"},
			func(f *zli.Flags) []any {
				return []any{f.String("", "s")}
			}, `
			string 1 → "arg"
			args     → 2 [pos 1 pos 2]
			`, ""},

		{"single - is a valid argument", []string{"progname", "-s", "-", "-"},
			func(f *zli.Flags) []any {
				return []any{f.String("", "s")}
			}, `
			string 1 → "-"
			args     → 1 [-]
			`, ""},
		// Make sure parsing is stopped after --
		{
			"-- bool", []string{"prog", "-b", "--"},
			func(f *zli.Flags) []any {
				return []any{
					f.Bool(false, "b"),
				}
			}, `
				bool 1 → true
				args   → 0 []
			`, ""},
		//
		/*
			{[]string{"prog", "-b", "--", "-str"}, "", `
					str   | false | "default"
					bool  | true  | true
					args  | 1     | [-str]
				`},
			{[]string{"prog", "-b", "--", ""}, "", `
					str   | false | "default"
					bool  | true  | true
					args  | 1     | []
				`},
			// Various --
		*/

		// Basic test for all the different flag types.
		{"bool", []string{"prog", "-b"},
			func(f *zli.Flags) []any {
				return []any{f.Bool(false, "b")}
			}, `
			bool 1 → true
			args   → 0 []
			`, ""},
		{"string", []string{"prog", "-s", "val"},
			func(f *zli.Flags) []any {
				return []any{f.String("", "s")}
			}, `
				string 1 → "val"
				args     → 0 []
			`, ""},
		{"int", []string{"prog", "-i", "42"},
			func(f *zli.Flags) []any {
				return []any{f.Int(0, "i")}
			}, `
				int 1 → 42
				args  → 0 []
			`, ""},
		{"int64", []string{"prog", "-i", "42"},
			func(f *zli.Flags) []any {
				return []any{f.Int64(0, "i")}
			}, `
				int64 1 → 42
				args  → 0 []
			`, ""},
		{"int64", []string{"prog", "-i", "1_000_000"},
			func(f *zli.Flags) []any {
				return []any{f.Int64(0, "i")}
			}, `
				int64 1 → 1000000
				args  → 0 []
			`, ""},
		{"int64", []string{"prog", "-i", "0x10"},
			func(f *zli.Flags) []any {
				return []any{f.Int64(0, "i")}
			}, `
				int64 1 → 16
				args  → 0 []
			`, ""},
		{"float64", []string{"prog", "-i", "42.666"},
			func(f *zli.Flags) []any {
				return []any{f.Float64(0, "i")}
			}, `
				float64 1 → 42.666000
				args  → 0 []
			`, ""},
		{"intcounter", []string{"prog", "-i", "-i", "-i"},
			func(f *zli.Flags) []any {
				return []any{f.IntCounter(0, "i")}
			}, `
				int 1 → 3
				args  → 0 []
			`, ""},
		{"stringlist", []string{"prog", "-s", "a", "-s", "b", "-s", "c"},
			func(f *zli.Flags) []any {
				return []any{f.StringList(nil, "s")}
			}, `
				list 1 → [a b c]
				args   → 0 []
			`, ""},
		{"intlist", []string{"prog", "-s", "1", "-s", "3", "-s", "5"},
			func(f *zli.Flags) []any {
				return []any{f.IntList(nil, "s")}
			}, `
				list 1 → [1 3 5]
				args   → 0 []
			`, ""},

		// Various kinds of wrong input.
		{"unknown", []string{"prog", "-x"},
			func(f *zli.Flags) []any {
				return []any{f.String("", "s")}
			}, `
				string 1 → ""
				args     → 1 [-x]
			`, `unknown flag: "-x"`},
		{"no argument", []string{"prog", "-s"},
			func(f *zli.Flags) []any {
				return []any{f.String("", "s")}
			}, `
				string 1 → ""
				args     → 1 [-s]
			`, "-s: needs an argument"},
		{"multiple", []string{"prog", "-s=a", "-s=b"},
			func(f *zli.Flags) []any {
				return []any{f.String("", "s")}
			}, `
				string 1 → "a"
				args     → 2 [-s=a -s=b]
			`, `flag given more than once: "-s=b"`},
		{"not an int", []string{"prog", "-i=no"},
			func(f *zli.Flags) []any {
				return []any{f.Int(42, "i")}
			}, `
				int 1 → 42
				args  → 1 [-i=no]
		`, `-i=no: invalid syntax (must be a number)`},
		{"not an int64", []string{"prog", "-i=no"},
			func(f *zli.Flags) []any {
				return []any{f.Int64(42, "i")}
			}, `
				int64 1 → 42
				args    → 1 [-i=no]
		`, `-i=no: invalid syntax (must be a number)`},
		{"not a float", []string{"prog", "-i=no"},
			func(f *zli.Flags) []any {
				return []any{f.Float64(42, "i")}
			}, `
				float64 1 → 42.000000
				args      → 1 [-i=no]
		`, `-i=no: invalid syntax (must be a number)`},

		// Argument parsing
		{"-s=arg", []string{"prog", "-s=xx"},
			func(f *zli.Flags) []any {
				return []any{
					f.String("default", "s"),
				}
			}, `
				string 1 → "xx"
				args     → 0 []
			`, ""},
		{"--s=arg", []string{"prog", "--s=xx"},
			func(f *zli.Flags) []any {
				return []any{
					f.String("default", "s"),
				}
			}, `
				string 1 → "xx"
				args     → 0 []
			`, ""},
		{"--s=-arg", []string{"prog", "--s=-xx"},
			func(f *zli.Flags) []any {
				return []any{
					f.String("default", "s"),
				}
			}, `
				string 1 → "-xx"
				args     → 0 []
			`, ""},
		{"--s=-o", []string{"prog", "--s=-o"},
			func(f *zli.Flags) []any {
				return []any{
					f.String("default", "s"),
					f.String("default", "o"),
				}
			}, `
				string 1 → "-o"
				string 2 → "default"
				args     → 0 []
			`, ""},
		{"--s -o", []string{"prog", "-s", "-o"},
			func(f *zli.Flags) []any {
				return []any{
					f.String("", "s"),
					f.String("", "o"),
				}
			}, `
				string 1 → ""
				string 2 → ""
				args     → 2 [-s -o]
			`, "-o: needs an argument"},
		{"--s arg", []string{"prog", "--s", "xx"},
			func(f *zli.Flags) []any {
				return []any{
					f.String("default", "s"),
				}
			}, `
				string 1 → "xx"
				args     → 0 []
			`, ""},
		{"blank =", []string{"prog", "-s="},
			func(f *zli.Flags) []any {
				return []any{
					f.String("default", "s"),
				}
			}, `
				string 1 → ""
				args     → 0 []
			`, ""},
		{"blank space", []string{"prog", "-s", ""},
			func(f *zli.Flags) []any {
				return []any{
					f.String("default", "s"),
				}
			}, `
				string 1 → ""
				args     → 0 []
			`, ""},

		// Okay for booleans to have multiple flags, as it doesn't really
		// matter.
		{"multiple bool", []string{"prog", "-b", "-b"},
			func(f *zli.Flags) []any {
				return []any{f.Bool(false, "b")}
			}, `
				bool 1 → true
				args   → 0 []
			`, ""},

		// Group -ab as -a -b if they're booleans.
		{"group bool", []string{"prog", "-a", "-b"},
			func(f *zli.Flags) []any {
				return []any{
					f.Bool(false, "a"),
					f.Bool(false, "b"),
				}
			}, `
				bool 1 → true
				bool 2 → true
				args   → 0 []
			`, ""},
		{"group bool", []string{"prog", "-ab"},
			func(f *zli.Flags) []any {
				return []any{
					f.Bool(false, "a"),
					f.Bool(false, "b"),
				}
			}, `
				bool 1 → true
				bool 2 → true
				args   → 0 []
			`, ""},
		{"group bool only with single -", []string{"prog", "--ab"},
			func(f *zli.Flags) []any {
				return []any{
					f.Bool(false, "a"),
					f.Bool(false, "b"),
				}
			}, `
				bool 1 → false
				bool 2 → false
				args   → 1 [--ab]
			`, `unknown flag: "--ab"`},
		{"long flag overrides grouped bool", []string{"prog", "--ab", "x"},
			func(f *zli.Flags) []any {
				return []any{
					f.String("", "ab"),
					f.Bool(false, "a"),
					f.Bool(false, "b"),
				}
			}, `
				string 1 → "x"
				bool 2 → false
				bool 3 → false
				args   → 0 []
			`, ""},

		{"arguments starting with - work", []string{"prog", "-b=-arg", "--long=--long"},
			func(f *zli.Flags) []any {
				return []any{
					f.String("", "b"),
					f.String("", "l", "long"),
				}
			}, `
				string 1 → "-arg"
				string 2 → "--long"
				args     → 0 []
			`, ""},

		{"prefer_long", []string{"prog", "-long"},
			func(f *zli.Flags) []any {
				return []any{
					f.Bool(false, "long"),
					f.Bool(false, "l"),
					f.Bool(false, "o"),
					f.Bool(false, "n"),
					f.Bool(false, "g"),
				}
			}, `
				bool 1 → true
				bool 2 → false
				bool 3 → false
				bool 4 → false
				bool 5 → false
				args   → 0 []
			`, ""},

		{"prefer_long", []string{"prog", "-long"},
			func(f *zli.Flags) []any {
				return []any{
					f.Bool(false, "l"),
					f.Bool(false, "o"),
					f.Bool(false, "long"),
					f.Bool(false, "n"),
					f.Bool(false, "g"),
				}
			}, `
				bool 1 → false
				bool 2 → false
				bool 3 → true
				bool 4 → false
				bool 5 → false
				args   → 0 []
			`, ""},

		// Optional()
		{"optional", []string{"prog", "-s1", "-s2", "val"},
			func(f *zli.Flags) []any {
				return []any{
					f.Optional().String("def1", "s1"),
					f.String("def2", "s2"),
				}
			}, `
				string 1 → "def1"
				string 2 → "val"
				args     → 0 []
			`, ""},
		{"optional at end", []string{"prog", "-s2", "val", "-s1"},
			func(f *zli.Flags) []any {
				return []any{
					f.Optional().String("def1", "s1"),
					f.String("def2", "s2"),
				}
			}, `
				string 1 → "def1"
				string 2 → "val"
				args     → 0 []
			`, ""},

		{"optional works for one flag only", []string{"prog", "-s1", "-s2"},
			func(f *zli.Flags) []any {
				return []any{
					f.Optional().String("def1", "s1"),
					f.String("def2", "s2"),
				}
			}, `
				string 1 → "def1"
				string 2 → "def2"
				args     → 2 [-s1 -s2]
				`, "-s2: needs an argument"},
		{"optional int", []string{"prog", "-i1", "-s2", "val", "-i2", "2"},
			func(f *zli.Flags) []any {
				return []any{
					f.Optional().Int(11, "i1"),
					f.Optional().Int(12, "i2"),
					f.String("def2", "s2"),
				}
			}, `
				int 1    → 11
				int 2    → 2
				string 3 → "val"
				args     → 0 []
			`, ""},
		{"optional int at end", []string{"prog", "-i1", "-s2", "val", "-i2"},
			func(f *zli.Flags) []any {
				return []any{
					f.Optional().Int(11, "i1"),
					f.Optional().Int(12, "i2"),
					f.String("def2", "s2"),
				}
			}, `
				int 1    → 11
				int 2    → 12
				string 3 → "val"
				args     → 0 []
			`, ""},

		// Flag definitions can start with "-"
		{"leading -", []string{"", "-b"},
			func(f *zli.Flags) []any {
				return []any{f.Bool(false, "-b")}
			}, `
				bool 1 → true
				args   → 0 []
			`, ""},
		{"leading -", []string{"", "-b"},
			func(f *zli.Flags) []any {
				return []any{f.Bool(false, "--b", "--bool")}
			}, `
				bool 1 → true
				args   → 0 []
			`, ""},

		{"leading -", []string{"", "-bool"},
			func(f *zli.Flags) []any {
				return []any{f.Bool(false, "--a", "--bool")}
			}, `
				bool 1 → true
				args   → 0 []
			`, ""},

		// Short flags with values
		{"short without space", []string{"prog", "-w8"},
			func(f *zli.Flags) []any {
				return []any{
					f.Int(0, "w"),
				}
			}, `
				int 1    → 8
				args     → 0 []
			`, ""},
		{"short without space", []string{"prog", "-bw8", "X"},
			func(f *zli.Flags) []any {
				return []any{
					f.Bool(false, "b"),
					f.Int(0, "w"),
				}
			}, `
				bool 1   → true
				int 2    → 8
				args     → 1 [X]
			`, ""},
		{"short without space", []string{"prog", "-w", "81"},
			func(f *zli.Flags) []any {
				return []any{
					f.Bool(false, "b"),
					f.Int(0, "w"),
				}
			}, `
				bool 1   → false
				int 2    → 81
				args     → 0 []
			`, ``},
		{"short without space", []string{"prog", "-bw81", "X"},
			func(f *zli.Flags) []any {
				return []any{
					f.Bool(false, "b"),
					f.Int(0, "w"),
				}
			}, `
				bool 1   → true
				int 2    → 81
				args     → 1 [X]
			`, ``},
		{"short without space", []string{"prog", "-w81"},
			func(f *zli.Flags) []any {
				return []any{
					f.Bool(false, "1"),
					f.Int(0, "w"),
				}
			}, `
				bool 1   → false
				int 2    → 81
				args     → 0 []
			`, ``},
		{"short without space", []string{"prog", "-w18"},
			func(f *zli.Flags) []any {
				return []any{
					f.Bool(false, "1"),
					f.Int(0, "w"),
				}
			}, `
				bool 1   → false
				int 2    → 18
				args     → 0 []
			`, ``},
		// Not when it's a bool
		{"short without space", []string{"prog", "-w8"},
			func(f *zli.Flags) []any {
				return []any{
					f.Bool(false, "w"),
				}
			}, `
				bool 1   → false
				args     → 1 [-w8]
			`, `unknown flag: "-w8"`},

		// AllowMultiple
		{"multiple flags opt:multiple", []string{"prog", "-w10"},
			func(f *zli.Flags) []any {
				return []any{
					f.Int(0, "w"),
				}
			}, `
				int 1    → 10
				args     → 0 []
			`, ``},
		{"multiple flags opt:multiple", []string{"prog", "-w10", "-w20"},
			func(f *zli.Flags) []any {
				return []any{
					f.Int(0, "w"),
				}
			}, `
				int 1    → 20
				args     → 0 []
			`, ``},
		{"multiple flags opt:multiple", []string{"prog", "-w10", "-w20"},
			func(f *zli.Flags) []any {
				return []any{
					f.Optional().Int(0, "w"),
				}
			}, `
				int 1    → 20
				args     → 0 []
			`, ``},

		// TODO: maybe this should go back to 0? Not sure what makes the most
		// sense here.
		{"multiple flags opt:multiple", []string{"prog", "-w10", "-w"},
			func(f *zli.Flags) []any {
				return []any{
					f.Optional().Int(0, "w"),
				}
			}, `
				int 1    → 10
				args     → 0 []
			`, ``},

		// - and _ are identical
		{"- and _ are identical", []string{"prog", "-st-r", "a", "-st_r=b", "-cn-t", "-cn_t", "-boo_l"},
			func(f *zli.Flags) []any {
				return []any{f.StringList(nil, "st-r"), f.IntCounter(0, "cn-t"), f.Bool(false, "boo-l")}
			}, `
				list 1 → [a b]
				int 2  → 2
				bool 3 → true
				args   → 0 []
			`, ""},
		{"- and _ are identical", []string{"prog", "-st-r", "a", "-st_r=b", "-cn-t", "-cn_t", "-boo-l"},
			func(f *zli.Flags) []any {
				return []any{f.StringList(nil, "st_r"), f.IntCounter(0, "cn_t"), f.Bool(false, "boo_l")}
			}, `
				list 1 → [a b]
				int 2  → 2
				bool 3 → true
				args   → 0 []
			`, ""},
	}

	type (
		booler       interface{ Bool() bool }
		stringer     interface{ String() string }
		inter        interface{ Int() int }
		int64er      interface{ Int64() int64 }
		floater      interface{ Float64() float64 }
		stringlister interface{ Strings() []string }
		intlister    interface{ Ints() []int }
	)

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			flag := zli.NewFlags(tt.args)
			setFlags := tt.flags(&flag)

			var err error
			if strings.Contains(tt.name, "opt:multiple") { // Hackity hack!
				err = flag.Parse(zli.AllowMultiple())
			} else {
				err = flag.Parse()
			}
			if !errorContains(err, tt.wantErr) {
				t.Fatalf("wrong error\nout:  %v\nwant: %v", err, tt.wantErr)
			}

			var out string
			for i, f := range setFlags {
				switch ff := f.(type) {
				case booler:
					out += fmt.Sprintf("bool %d → %t\n", i+1, ff.Bool())
				case stringer:
					out += fmt.Sprintf("string %d → %q\n", i+1, ff.String())
				case inter:
					out += fmt.Sprintf("int %d → %d\n", i+1, ff.Int())
				case int64er:
					out += fmt.Sprintf("int64 %d → %d\n", i+1, ff.Int64())
				case floater:
					out += fmt.Sprintf("float64 %d → %f\n", i+1, ff.Float64())
				case stringlister:
					out += fmt.Sprintf("list %d → %v\n", i+1, ff.Strings())
				case intlister:
					out += fmt.Sprintf("list %d → %v\n", i+1, ff.Ints())
				default:
					t.Fatalf("unknown type: %T", f)
				}
			}
			out += fmt.Sprintf("args → %d %v", len(flag.Args), flag.Args)

			want := strings.TrimSpace(strings.ReplaceAll(tt.want, "\t", ""))
			want = regexp.MustCompile(`\s+→\s+`).ReplaceAllString(want, " → ")

			// Indent so it looks nicer.
			out = "        " + strings.ReplaceAll(out, "\n", "\n        ")
			want = "        " + strings.ReplaceAll(want, "\n", "\n        ")

			if out != want {
				t.Errorf("\nout:\n%s\nwant:\n%s\n", out, want)
			}
		})
	}
}

func TestShiftCommand(t *testing.T) {
	tests := []struct {
		in       []string
		commands []string
		want     string
		wantErr  string
	}{
		{[]string{""}, nil, "", "no command given"},
		{[]string{"-a"}, nil, "", "no command given"},

		{[]string{"help"}, []string{"asd"}, "", `unknown command: "help"`},

		{[]string{"help"}, []string{"help", "heee"}, "help", ""},
		{[]string{"hel"}, []string{"help", "heee"}, "help", ""},
		{[]string{"he"}, []string{"help", "heee"}, "", `ambigious command: "he"; matches: "help", "heee"`},

		{[]string{"usage"}, []string{"help", "usage=help"}, "help", ""},

		{[]string{"create", "-db=x"}, []string{"create"}, "create", ""},
		{[]string{"-flag", "create", "-db=x"}, []string{"create"}, "create", ""},

		{[]string{"-flag", "create", "-db=x"}, nil, "create", ""},
	}

	for _, tt := range tests {
		t.Run("", func(t *testing.T) {
			f := zli.NewFlags(append([]string{"prog"}, tt.in...))
			have, err := f.ShiftCommand(tt.commands...)
			f.Bool(false, "a")
			f.Parse()

			if !errorContains(err, tt.wantErr) {
				t.Fatalf("wrong error\nhave: %q\nwant: %q", err, tt.wantErr)
			}
			if have != tt.want {
				t.Errorf("wrong cmd\nhave: %q\nwant: %q", have, tt.want)
			}
		})
	}
}

func TestPositional(t *testing.T) {
	tests := []struct {
		args    []string
		pos     [2]int
		wantErr string
	}{
		{[]string{"a", "b"}, [2]int{}, ""},

		{[]string{"a"}, [2]int{0, 1}, ""},
		{[]string{"a", "b"}, [2]int{0, 1}, "at most 1 positional argument accepted, but 2 given"},
		{[]string{"a", "b"}, [2]int{1, 1}, "exactly 1 positional argument required, but 2 given"},

		{[]string{"a", "b"}, [2]int{1, 2}, ""},
		{[]string{"a", "b", "c"}, [2]int{1, 2}, "between 1 and 2 positional arguments accepted, but 3 given"},

		{[]string{"a", "b", "c"}, [2]int{3, 0}, ""},
		{[]string{"a", "b"}, [2]int{3, 0}, "at least 3 positional arguments required, but 2 given"},
	}

	for _, tt := range tests {
		t.Run("", func(t *testing.T) {
			f := zli.NewFlags(append([]string{"prog"}, tt.args...))
			err := f.Parse(zli.Positional(tt.pos[0], tt.pos[1]))
			if !errorContains(err, tt.wantErr) {
				t.Fatalf("wrong error\nhave: %q\nwant: %q", err, tt.wantErr)
			}
		})
	}
}

func TestDoubleParse(t *testing.T) {
	f := zli.NewFlags([]string{"prog", "-global", "cmd", "-other"})

	var global = f.Bool(false, "global")
	{
		err := f.Parse(zli.AllowUnknown())
		if err != nil {
			t.Fatal(err)
		}
		if !global.Set() {
			t.Fatal("global not set")
		}
	}

	var other = f.Bool(false, "other")
	err := f.Parse()
	if err != nil {
		t.Fatal(err)
	}

	if !other.Set() {
		t.Error("other not set", f.Args)
	}
	if len(f.Args) != 1 && f.Args[1] != "cmd" {
		t.Error(f.Args)
	}
}

// TODO: test single-letters are not overriden
func TestFromEnv(t *testing.T) {
	t.Run("basic", func(t *testing.T) {
		f := zli.NewFlags([]string{"prog", "-str1=cli-value", "-bool1"})
		var (
			str1  = f.String("", "str1")
			str2  = f.String("", "str2")
			bool1 = f.Bool(false, "bool1")
			bool2 = f.Bool(false, "bool2", "alt-bool2")
			bool3 = f.Bool(false, "bool3")
			b     = f.Bool(false, "b")
		)

		os.Setenv("XX_STR1", "str1 from env")
		os.Setenv("XX_STR2", "str2 from env")
		os.Setenv("XX_BOOL2", "true")
		os.Setenv("XX_BOOL3", "")
		os.Setenv("XX_B", "")
		defer func() {
			for _, k := range []string{"STR1", "STR2", "BOOL2", "BOOL3", "B"} {
				os.Unsetenv("XX_" + k)
			}
		}()
		if err := f.Parse(zli.FromEnv("XX")); err != nil {
			t.Fatal(err)
		}

		have := fmt.Sprintf("str1=%q str2=%q bool1=%t bool2=%t bool3=%t b=%t",
			str1, str2, bool1.Bool(), bool2.Bool(), bool3.Bool(), b.Bool())
		want := `str1="cli-value" str2="str2 from env" bool1=true bool2=true bool3=true b=false`
		if have != want {
			t.Errorf("\nhave: %s\nwant: %s", have, want)
		}
	})

	t.Run("append", func(t *testing.T) {
		f := zli.NewFlags([]string{"prog", "-list1=cli1,cli2", "-count1"})
		var (
			list1  = f.StringList(nil, "list1")
			list2  = f.StringList(nil, "list2")
			count1 = f.IntCounter(0, "count1")
			count2 = f.IntCounter(0, "count2")
			count3 = f.IntCounter(0, "count3")
		)

		os.Setenv("XX_LIST1", "env1,env2")
		os.Setenv("XX_LIST2", "env1,env2")
		os.Setenv("XX_COUNT2", "1")
		os.Setenv("XX_COUNT3", "2")
		defer func() {
			for _, k := range []string{"LIST1", "LIST2", "COUNT2", "COUNT3"} {
				os.Unsetenv("XX_" + k)
			}
		}()
		if err := f.Parse(zli.FromEnv("XX")); err != nil {
			t.Fatal(err)
		}

		have := fmt.Sprintf("list1=%q list2=%q count1=%d count2=%d count3=%d",
			list1.StringsSplit(","), list2.StringsSplit(","), count1.Int(), count2.Int(), count3.Int())
		want := `list1=["cli1" "cli2"] list2=["env1" "env2"] count1=1 count2=1 count3=2`
		if have != want {
			t.Errorf("\nhave: %s\nwant: %s", have, want)
		}
	})

	t.Run("error unknown", func(t *testing.T) {
		f := zli.NewFlags([]string{"prog"})

		os.Setenv("XX_FOO", "aa")
		os.Setenv("XX_TWO", "aa")
		os.Setenv("XXBAR", "bb")
		os.Setenv("X_FOO", "bb")
		defer func() {
			for _, k := range []string{"XX_FOO", "XX_TWO", "XXBAR", "X_FOO"} {
				os.Unsetenv(k)
			}
		}()
		err := f.Parse(zli.FromEnv("XX"))
		if err == nil {
			t.Fatal("err is nil")
		}
		var uErr zli.ErrUnknownEnv
		if !errors.As(err, &uErr) {
			t.Fatalf("wrong error type: %#v", err)
		}
		if uErr.Error() != `unknown environment variables starting with "XX_": "XX_FOO", "XX_TWO"` {
			t.Errorf("wrong error message: %v", uErr)
		}
	})
}

// Just to make sure it's not ridiculously slow or anything.
func BenchmarkFlag(b *testing.B) {
	b.ReportAllocs()
	var err error
	for n := 0; n < b.N; n++ {
		flag := zli.NewFlags([]string{"prog", "cmd", "-vv", "-V", "str foo"})
		flag.Shift()
		flag.String("", "s", "str")
		flag.Bool(false, "V", "version")
		flag.IntCounter(0, "v", "verbose")
		err = flag.Parse()
	}
	_ = err
}

func errorContains(out error, want string) bool {
	if out == nil {
		return want == ""
	}
	if want == "" {
		return false
	}
	return strings.Contains(out.Error(), want)
}
