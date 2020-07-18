package zli

import (
	"bytes"
	"errors"
	"fmt"
	"net/mail"
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
