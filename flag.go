package zli

import (
	"fmt"
	"path/filepath"
	"regexp"
	"strconv"
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
				split[i] = Colorf(split[i], Bold, Underline)
			}
		}
		text = strings.Join(split, "\n")
	}

	if opts&UsageFlags != 0 {
		text = reFlags.ReplaceAllString(text, Colorf(`$0`, Underline))
	}

	return text
}

// ErrUnknownFlag is used when the flag parsing encounters unknown flags.
type ErrUnknownFlag struct{ flag string }

func (e ErrUnknownFlag) Error() string { return fmt.Sprintf("unknown flag: %q", e.flag) }

type Flags struct {
	Program string   // Program name.
	Args    []string // List of arguments, after parsing this will be reduces to non-flags.

	flags []flagValue
}

type flagValue struct {
	names []string
	value interface{}
}

// NewFlags creates a new Flags from os.Args.
func NewFlags(args []string) Flags {
	f := Flags{}
	if len(args) > 0 {
		f.Program = filepath.Base(args[0])
	}
	if len(args) > 1 {
		f.Args = args[1:]
	}
	return f
}

// Shift a value from the argument list.
func (f *Flags) Shift() string {
	if len(f.Args) == 0 {
		return ""
	}
	a := f.Args[0]
	f.Args = f.Args[1:]
	return a
}

func (f *Flags) Parse() error {
	var (
		p    []string
		skip bool
	)
	for i, a := range f.Args {
		if skip {
			skip = false
			continue
		}

		if a == "" || a[0] != '-' {
			p = append(p, a)
			continue
		}

		if a == "--" {
			p = append(p, f.Args[i+1:]...)
			break
		}

		flag, ok := f.match(a)
		if !ok {
			return &ErrUnknownFlag{flag: a}
		}

		var err error
		next := func() (string, bool) {
			if j := strings.IndexByte(f.Args[i], '='); j > -1 {
				return f.Args[i][j+1:], true
			}
			if i >= len(f.Args)-1 {
				err = fmt.Errorf("needs an argument")
				return "", false
			}
			skip = true
			return f.Args[i+1], true
		}

		var val string
		switch v := flag.value.(type) {
		case flagBool:
			*v.s = true
			*v.v = true
		case flagString:
			*v.v, *v.s = next()
		case flagInt:
			val, *v.s = next()
			x, err := strconv.ParseInt(val, 10, 64)
			if err != nil {
				return err
			}
			*v.v = int(x)
		case flagInt64:
			val, *v.s = next()
			x, err := strconv.ParseInt(val, 10, 64)
			if err != nil {
				return err
			}
			*v.v = x
		case flagFloat64:
			val, *v.s = next()
			x, err := strconv.ParseFloat(val, 64)
			if err != nil {
				return err
			}
			*v.v = x
		case flagIntCounter:
			*v.s = true
			*v.v++
		}
		if err != nil {
			return fmt.Errorf("%s: %s", a, err)
		}
	}
	f.Args = p
	return nil
}

func (f *Flags) match(arg string) (flagValue, bool) {
	arg = strings.TrimLeft(arg, "-")
	for _, flag := range f.flags {
		for _, name := range flag.names {
			if name == arg || strings.HasPrefix(arg, name+"=") {
				return flag, true
			}
		}
	}
	return flagValue{}, false
}

type (
	flagBool struct {
		v *bool
		s *bool
	}
	flagString struct {
		v *string
		s *bool
	}
	flagInt struct {
		v *int
		s *bool
	}
	flagInt64 struct {
		v *int64
		s *bool
	}
	flagFloat64 struct {
		v *float64
		s *bool
	}
	flagIntCounter struct {
		v *int
		s *bool
	}
)

func (f flagBool) Bool() bool          { return *f.v }
func (f flagString) String() string    { return *f.v }
func (f flagInt) Int() int             { return *f.v }
func (f flagInt64) Int64() int64       { return *f.v }
func (f flagFloat64) Float64() float64 { return *f.v }
func (f flagIntCounter) Int() int      { return *f.v }

func (f flagBool) Set() bool       { return *f.s }
func (f flagString) Set() bool     { return *f.s }
func (f flagInt) Set() bool        { return *f.s }
func (f flagInt64) Set() bool      { return *f.s }
func (f flagFloat64) Set() bool    { return *f.s }
func (f flagIntCounter) Set() bool { return *f.s }

func (f *Flags) append(v interface{}, n string, a ...string) {
	f.flags = append(f.flags, flagValue{value: v, names: append([]string{n}, a...)})
}

func (f *Flags) Bool(def bool, name string, aliases ...string) flagBool {
	v := flagBool{v: &def, s: new(bool)}
	f.append(v, name, aliases...)
	return v
}
func (f *Flags) String(def, name string, aliases ...string) flagString {
	v := flagString{v: &def, s: new(bool)}
	f.append(v, name, aliases...)
	return v
}
func (f *Flags) Int(def int, name string, aliases ...string) flagInt {
	v := flagInt{v: &def, s: new(bool)}
	f.append(v, name, aliases...)
	return v
}
func (f *Flags) Int64(def int64, name string, aliases ...string) flagInt64 {
	v := flagInt64{v: &def, s: new(bool)}
	f.append(v, name, aliases...)
	return v
}
func (f *Flags) Float64(def float64, name string, aliases ...string) flagFloat64 {
	v := flagFloat64{v: &def, s: new(bool)}
	f.append(v, name, aliases...)
	return v
}
func (f *Flags) IntCounter(def int, name string, aliases ...string) flagIntCounter {
	v := flagIntCounter{v: &def, s: new(bool)}
	f.append(v, name, aliases...)
	return v
}
