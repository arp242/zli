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
		args     []string
		stdin    string
		want     string
		wantExit int
	}{
		{
			[]string{"grep", "^package", "main.go"},
			"",
			"main.go:2:package main\n",
			0,
		},
		{
			[]string{"grep", "^package"},
			read(t, "main.go"),
			"2:package main\n",
			0,
		},
	}

	for i, tt := range tests {
		t.Run(fmt.Sprintf("%d", i), func(t *testing.T) {
			exit, in, out, reset := zli.Test()
			defer reset()

			if tt.stdin != "" {
				in.WriteString(tt.stdin)
			}

			os.Args = tt.args
			func() {
				defer exit.Recover()
				main()
			}()
			exit.Want(t, tt.wantExit)

			if out.String() != tt.want {
				t.Errorf("wrong output:\nout:  %s\nwant: %s", out.String(), tt.want)
			}
		})
	}
}
