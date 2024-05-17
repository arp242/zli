package zli

import (
	"bytes"
	"errors"
	"fmt"
	"io"
	"io/ioutil"
	"net/mail"
	"os"
	"reflect"
	"runtime"
	"strings"
	"testing"
)

func TestFatal(t *testing.T) {
	tests := []struct {
		in   any
		args []any
		want string
	}{
		{"", nil, "zli.test: \n"},
		{nil, nil, "zli.test: <nil>\n"},
		{nil, nil, "zli.test: <nil>\n"},
		{42, nil, "zli.test: 42\n"},

		{"oh noes", nil, "zli.test: oh noes\n"},
		{"oh noes: %d", []any{42}, "zli.test: oh noes: 42\n"},
		{"oh noes: %d %d", []any{42, 666}, "zli.test: oh noes: 42 666\n"},
		{[]byte("oh noes: %d %d"), []any{42, 666}, "zli.test: oh noes: 42 666\n"},

		{errors.New("oh noes"), nil, "zli.test: oh noes\n"},
		{errors.New("oh noes"), []any{"data", 666}, "zli.test: oh noes [data 666]\n"},

		{mail.Address{Name: "asd", Address: "qwe"}, nil, "zli.test: {asd qwe}\n"},
		{mail.Address{Name: "asd", Address: "qwe"}, []any{"data", 666}, "zli.test: {asd qwe} [data 666]\n"},
	}

	for i, tt := range tests {
		t.Run(fmt.Sprintf("%d", i), func(t *testing.T) {
			exit, _, out := Test(t)

			func() {
				defer exit.Recover()
				Fatalf(tt.in, tt.args...)
			}()

			if *exit != 1 {
				t.Errorf("wrong exit: %d", *exit)
			}
			got := out.String()
			if got != tt.want {
				t.Errorf("\ngot:  %q\nwant: %q", got, tt.want)
			}
		})
	}
}

func TestF(t *testing.T) {
	t.Run("nil", func(t *testing.T) {
		_, _, out := Test(t)

		var err error
		F(err) // Will panic if exit is set.

		if out.String() != "" {
			t.Errorf("out has data: %q", out.String())
		}
	})

	t.Run("nil", func(t *testing.T) {
		exit, _, out := Test(t)

		func() {
			defer exit.Recover()
			F(errors.New("oh noes"))
		}()

		if *exit != 1 {
			t.Errorf("wrong exit: %d", *exit)
		}
		if out.String() != "zli.test: oh noes\n" {
			t.Errorf("wrong out: %q", out.String())
		}
	})
}

func TestInputOrFile(t *testing.T) {
	tests := []struct {
		in            string
		stdin         io.Reader
		want, wantErr string
	}{
		{"/nonexistent", nil, "", "no such file or directory"},
		{"zli_test.go", strings.NewReader("xx"), "package zli", ""},
		{"-", strings.NewReader("xx yy\nzz"), "xx yy\nzz", ""},
		{"", strings.NewReader("xx yy\nzz"), "xx yy\nzz", ""},
	}

	for _, tt := range tests {
		t.Run(tt.in, func(t *testing.T) {
			Stdin = tt.stdin
			defer func() { Stdin = os.Stdin }()

			fp, err := InputOrFile(tt.in, true)
			if !errorContains(err, tt.wantErr) {
				t.Errorf("wrong error\ngot:  %s\nwant: %s", err, tt.wantErr)
			}

			// No need to test the test if there's an error.
			if err != nil {
				return
			}

			if fp == nil {
				t.Fatal("fp is nil")
			}

			got, err := ioutil.ReadAll(fp)
			if err != nil {
				t.Errorf("error reading fp: %s", err)
			}

			err = fp.Close()
			if err != nil {
				t.Errorf("error closing fp: %s", err)
			}

			g := string(got)
			if len(g) > len(tt.want) {
				g = g[:len(tt.want)]
			}
			if !strings.HasPrefix(g, tt.want) {
				t.Errorf("wrong output\ngot:  %q\nwant: %q", g, tt.want)
			}
		})
	}
}

func TestInputOrArgs(t *testing.T) {
	tests := []struct {
		in    []string
		sep   string
		stdin io.Reader
		want  []string
	}{
		{[]string{"arg"}, "", nil, []string{"arg"}},
		{nil, "", strings.NewReader(""), []string{}},
		{nil, "\n ", strings.NewReader(""), []string{}},

		{nil, "", strings.NewReader("a"), []string{"a"}},
		{[]string{}, "", strings.NewReader("a"), []string{"a"}},
		{[]string{}, "", strings.NewReader("a\nb c"), []string{"a\nb c"}},

		{[]string{}, " ", strings.NewReader(" a b  c "), []string{"a", "b", "c"}},
		{[]string{}, "\x00", strings.NewReader("aa\x00bb"), []string{"aa", "bb"}},
	}

	for _, tt := range tests {
		t.Run(fmt.Sprintf("%s", tt.in), func(t *testing.T) {
			Stdin = tt.stdin
			defer func() { Stdin = os.Stdin }()

			got, err := InputOrArgs(tt.in, tt.sep, true)
			if err != nil {
				t.Fatal(err)
			}

			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("\ngot:  %#v\nwant: %#v", got, tt.want)
			}
		})
	}
}

func TestPager(t *testing.T) {
	set := func(term bool) (*bytes.Buffer, func()) {
		buf := new(bytes.Buffer)
		Stdout = buf
		Stderr = buf

		save := IsTerminal
		if term {
			IsTerminal = func(uintptr) bool { return true }
		}

		return buf, func() {
			Stdout = os.Stdout
			Stderr = os.Stderr
			IsTerminal = save
		}
	}

	t.Run("not a terminal", func(t *testing.T) {
		buf, c := set(false)
		defer c()

		Pager(strings.NewReader("buffy"))
		if buf.String() != "buffy" {
			t.Error(buf.String())
		}
	})

	t.Run("no PAGER", func(t *testing.T) {
		buf, c := set(true)
		defer c()

		os.Unsetenv("PAGER")
		Pager(strings.NewReader("buffy"))
		if buf.String() != "buffy" {
			t.Errorf("out: %q", buf.String())
		}
	})

	t.Run("PAGER doesn't exist", func(t *testing.T) {
		buf, c := set(true)
		defer c()

		os.Setenv("PAGER", "doesntexistasdad")
		Pager(strings.NewReader("buffy"))

		want := "zli.test: zli.Pager: running $PAGER: exec: \"doesntexistasdad\": executable file not found in $PATH\nbuffy"
		if buf.String() != want {
			t.Errorf("out: %q", buf.String())
		}
	})

	t.Run("PAGER doesn't exist w/ args", func(t *testing.T) {
		buf, c := set(true)
		defer c()

		os.Setenv("PAGER", "doesntexistasdad -r -f")
		Pager(strings.NewReader("buffy"))

		want := "zli.test: zli.Pager: running $PAGER: exec: \"doesntexistasdad\": executable file not found in $PATH\nbuffy"
		if buf.String() != want {
			t.Errorf("out: %q", buf.String())
		}
	})

	t.Run("error", func(t *testing.T) {
		buf, c := set(true)
		defer c()

		// zli.Pager: running $PAGER: exit status 1
		os.Setenv("PAGER", "false")
		Pager(strings.NewReader("buffy"))

		want := "zli.test: zli.Pager: running $PAGER: exit status 1\n"
		if buf.String() != want {
			t.Errorf("out: %q", buf.String())
		}
	})

	t.Run("run it", func(t *testing.T) {
		if runtime.GOOS == "windows" || runtime.GOOS == "js" {
			t.Skip("requires cat shell tool")
		}

		buf, c := set(true)
		defer c()

		os.Setenv("PAGER", "cat")
		Pager(strings.NewReader("buffy"))

		if buf.String() != "buffy" {
			t.Errorf("out: %q", buf.String())
		}
	})

	t.Run("pagestdout", func(t *testing.T) {
		buf, c := set(true)
		defer c()

		func() {
			defer PagerStdout()()
			fmt.Fprintf(Stdout, "buffy")
		}()

		if buf.String() != "buffy" {
			t.Errorf("out: %q", buf.String())
		}

	})
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
