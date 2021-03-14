package main

import (
	"fmt"
	"io/ioutil"
	"os"
	"testing"

	"zgo.at/zli"
)

func read(t *testing.T, f string) string {
	d, err := ioutil.ReadFile(f)
	if err != nil {
		t.Fatal(err)
	}
	return string(d)
}

func TestGrep(t *testing.T) {
	tests := []struct {
		args       []string
		stdin      string
		wantTerm   string
		wantNoTerm string
		wantExit   int
	}{
		{
			[]string{"grep", "^package", "main.go"},
			"",
			"\x1b[1;4mmain.go\x1b[0m\n\x1b[35m2\x1b[0m:\x1b[31mpackage\x1b[0m main\n",
			"main.go:2:package main\n",
			0,
		},
		{
			[]string{"grep", "^package"},
			read(t, "main.go"),
			"grep: reading from stdin...\r\x1b[35m2\x1b[0m:\x1b[31mpackage\x1b[0m main\n",
			"2:package main\n",
			0,
		},
		{
			[]string{"grep", "this string is not in the file", "main.go"},
			"", "", "", 1,
		},
		{
			[]string{"grep", "(invalid", "main.go"},
			"",
			"grep: error parsing regexp: missing closing ): `(invalid`\n",
			"grep: error parsing regexp: missing closing ): `(invalid`\n", 2,
		},

		// -o flag
		{
			[]string{"grep", "-o", "^package", "main.go"},
			"",
			"\x1b[1;4mmain.go\x1b[0m\n\x1b[35m2\x1b[0m:\x1b[31mpackage\x1b[0m\n",
			"main.go:2:package\n",
			0,
		},
		// -q flag
		{
			[]string{"grep", "-q", "^package", "main.go"},
			"", "", "", 0,
		},
	}

	for i, tt := range tests {
		t.Run(fmt.Sprintf("%d", i), func(t *testing.T) {
			run := func(isTerm bool, want string) {
				exit, in, out := zli.Test(t)

				s := zli.IsTerminal
				zli.WantColor = isTerm
				zli.IsTerminal = func(uintptr) bool { return isTerm }
				defer func() { zli.IsTerminal = s }()

				if tt.stdin != "" {
					in.WriteString(tt.stdin)
				}

				os.Args = tt.args
				func() {
					defer exit.Recover()
					main()
				}()
				exit.Want(t, tt.wantExit)

				if out.String() != want {
					t.Errorf("wrong output:\nout:\n%s\nwant:\n%s", out.String(), want)
				}
			}

			t.Run("no-term", func(t *testing.T) { run(false, tt.wantNoTerm) })
			t.Run("term", func(t *testing.T) { run(true, tt.wantTerm) })
		})
	}
}
