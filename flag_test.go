package zli_test

import (
	"fmt"
	"regexp"
	"strings"
	"testing"

	"zgo.at/zli"
)

func ExampleFlag() {
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

	// You can check if the flag was present on the CLI with Set(). This way you can
	// distinish between "was an empty value passed" // (-format '') and "this flag
	// wasn't on the CLI".
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

func TestFlag(t *testing.T) {
	tests := []struct {
		args    []string
		wantErr string
		want    string
	}{
		{[]string{}, "", `
				str   | false | "default"
				bool  | false | false
				args  | 0     | []
		`},
		{[]string{"prog"}, "", `
				str   | false | "default"
				bool  | false | false
				args  | 0     | []
		`},
		{[]string{"prog", "arg"}, "", `
				str   | false | "default"
				bool  | false | false
				args  | 1     | [arg]
		`},

		// -s
		{[]string{"prog", "-s", "fmt"}, "", `
				str   | true  | "fmt"
				bool  | false | false
				args  | 0     | []
			`},
		{[]string{"prog", "-s=fmt"}, "", `
				str   | true  | "fmt"
				bool  | false | false
				args  | 0     | []
			`},
		{[]string{"prog", "-str=fmt"}, "", `
				str   | true  | "fmt"
				bool  | false | false
				args  | 0     | []
			`},
		{[]string{"prog", "--str=fmt"}, "", `
				str   | true  | "fmt"
				bool  | false | false
				args  | 0     | []
			`},
		{[]string{"prog", "--s=fmt"}, "", `
				str   | true  | "fmt"
				bool  | false | false
				args  | 0     | []
			`},
		{[]string{"prog", "--str", "fmt"}, "", `
				str   | true  | "fmt"
				bool  | false | false
				args  | 0     | []
			`},
		{[]string{"prog", "--s=fmt"}, "", `
				str   | true  | "fmt"
				bool  | false | false
				args  | 0     | []
			`},
		{[]string{"prog", "-s", ""}, "", `
				str   | true  | ""
				bool  | false | false
				args  | 0     | []
			`},

		{[]string{"prog", "-s"}, "-s: needs an argument", `
				str   | false | ""
				bool  | false | false
				args  | 1     | [-s]
			`},
		{[]string{"prog", "-str"}, "-str: needs an argument", `
				str   | false | ""
				bool  | false | false
				args  | 1     | [-str]
			`},
		{[]string{"prog", "-s="}, "", `
				str   | true  | ""
				bool  | false | false
				args  | 0     | []
			`},
		{[]string{"prog", "-str="}, "", `
				str   | true  | ""
				bool  | false | false
				args  | 0     | []
			`},

		{[]string{"prog", "-str", "-str"}, "", `
				str   | true  | "-str"
				bool  | false | false
				args  | 0     | []
			`},
		{[]string{"prog", "-str=-str"}, "", `
				str   | true  | "-str"
				bool  | false | false
				args  | 0     | []
			`},

		// --
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

		{[]string{"prog", "-b", "--"}, "", `
				str   | false | "default"
				bool  | true  | true
				args  | 0     | []
			`},
	}

	for _, tt := range tests {
		t.Run(fmt.Sprintf("%s", tt.args), func(t *testing.T) {
			flag := zli.NewFlags(tt.args)

			str := flag.String("default", "s", "str")
			b := flag.Bool(false, "b", "bool")
			err := flag.Parse()
			if !errorContains(err, tt.wantErr) {
				t.Fatalf("wrong error\nout:  %v\nwant: %v", err, tt.wantErr)
			}

			out := fmt.Sprintf("str\t%t\t%q\nbool\t%t\t%v\nargs\t%d\t%v",
				str.Set(), str.String(), b.Set(), b.Bool(),
				len(flag.Args), flag.Args)
			want := strings.TrimSpace(strings.ReplaceAll(tt.want, "\t", ""))
			want = regexp.MustCompile(`\s+\|\s+`).ReplaceAllString(want, "\t")

			if out != want {
				t.Errorf("\nout:\n%s\nwant:\n%s\n", out, want)
			}
		})
	}
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
