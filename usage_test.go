package zli_test

import (
	"fmt"
	"strings"
	"testing"

	"zgo.at/zli"
)

func TestUsage(t *testing.T) {
	tests := []struct {
		flags    int
		in, want string
	}{
		{0, "", ""},

		{zli.UsageHeaders,
			`
				Hello:
				  text:

				s p:

				s-p:
				No blank line:

			`,
			"\n\x1b[1mHello:\x1b[0m\n  text:\n\n\x1b[1ms p:\x1b[0m\n\n\x1b[1ms-p:\x1b[0m\nNo blank line:\n\n",
		},

		{
			zli.UsageFlags,
			`
				Hello, -flag
				-flag
				-flag-name, --flag
				-flag=foo

				hyphen-word.
			`,
			"\nHello, \x1b[4m-flag\x1b[0m\n\x1b[4m-flag\x1b[0m\n\x1b[4m-flag-name\x1b[0m, \x1b[4m--flag\x1b[0m\n\x1b[4m-flag=foo\x1b[0m\n\nhyphen-word.\n",
		},
	}

	for i, tt := range tests {
		zli.WantColor = true
		tt.in = strings.ReplaceAll(tt.in, "\t", "")
		tt.want = strings.ReplaceAll(tt.want, "\t", "")
		t.Run(fmt.Sprintf("%d", i), func(t *testing.T) {
			got := zli.Usage(tt.flags, tt.in)
			if got != tt.want {
				t.Errorf("\ngot:  %q\nwant: %q\n\n%s", got, tt.want, got)
				fmt.Println(got)
			}
		})
	}
}
