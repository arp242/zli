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
	"strings"
	"testing"
)

func TestFatal(t *testing.T) {
	tests := []struct {
		in   interface{}
		args []interface{}
		want string
	}{
		{"", nil, "zli.test: \n"},
		{nil, nil, "zli.test: <nil>\n"},
		{nil, nil, "zli.test: <nil>\n"},
		{42, nil, "zli.test: 42\n"},

		{"oh noes", nil, "zli.test: oh noes\n"},
		{"oh noes: %d", []interface{}{int(42)}, "zli.test: oh noes: 42\n"},

		{errors.New("oh noes"), nil, "zli.test: oh noes\n"},
		{errors.New("oh noes"), []interface{}{"data", 666}, "zli.test: oh noes [data 666]\n"},

		{mail.Address{Name: "asd", Address: "qwe"}, nil, "zli.test: {asd qwe}\n"},
		{mail.Address{Name: "asd", Address: "qwe"}, []interface{}{"data", 666}, "zli.test: {asd qwe} [data 666]\n"},
	}

	Exit = func(int) {}

	for i, tt := range tests {
		t.Run(fmt.Sprintf("%d", i), func(t *testing.T) {
			buf := new(bytes.Buffer)
			Stderr = buf
			Fatal(tt.in, tt.args...)

			got := buf.String()
			if got != tt.want {
				t.Errorf("\ngot:  %q\nwant: %q", got, tt.want)
			}
		})
	}
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

func errorContains(out error, want string) bool {
	if out == nil {
		return want == ""
	}
	if want == "" {
		return false
	}
	return strings.Contains(out.Error(), want)
}
