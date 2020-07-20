package zli

import (
	"regexp"
	"strings"
)

// Formatting flags for Usage.
const (
	// UsageTrim removes leading and trailing whitespace and appends a newline.
	//
	// This makes it easier to write usage strings without worrying too much
	// about leading/trailing whitespace, and with the trailing newline it's
	// easy to add a blank line between the usage and any error message
	// (fmt.Println if you wnat a blank line, fmt.Print if you don't).
	UsageTrim = 1

	// UsageHeaders formats headers in the form of:
	//
	//   Header:
	//
	// They must be at the start of the line and end with a :
	UsageHeaders = 2

	// UsageFlags formats flags in the form of:
	//
	//   -f
	//   -flag
	//   -flag=foo
	//   -flag=[foo]
	UsageFlags = 4
)

var (
	reHeader = regexp.MustCompile(`^\w[\w -]+:$`)
	reFlags  = regexp.MustCompile(`\B-{1,2}[a-z0-9=-]+\b`)
)

// Usage applies some formatting to a usage message. See the Usage* constants.
func Usage(opts int, text string) string {
	if opts&UsageTrim != 0 {
		text = strings.TrimSpace(text) + "\n"
	}

	if opts&UsageHeaders != 0 {
		split := strings.Split(text, "\n")
		for i := range split {
			if reHeader.MatchString(split[i]) {
				split[i] = Colorf(split[i], Bold|Underline)
			}
		}
		text = strings.Join(split, "\n")
	}

	if opts&UsageFlags != 0 {
		text = reFlags.ReplaceAllString(text, Colorf(`$0`, Underline))
	}

	return text
}
