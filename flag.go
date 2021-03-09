package zli

import (
	"errors"
	"fmt"
	"path/filepath"
	"strconv"
	"strings"
)

type (
	// ErrFlagUnknown is used when the flag parsing encounters unknown flags.
	ErrFlagUnknown struct{ flag string }

	// ErrFlagDouble is used when a flag is given more than once.
	ErrFlagDouble struct{ flag string }

	// ErrFlagInvalid is used when a flag has an invalid syntax (e.g. "no" for
	// an int flag).
	ErrFlagInvalid struct {
		flag string
		err  error
		kind string
	}
)

func (e ErrFlagUnknown) Error() string { return fmt.Sprintf("unknown flag: %q", e.flag) }
func (e ErrFlagDouble) Error() string  { return fmt.Sprintf("flag given more than once: %q", e.flag) }
func (e ErrFlagInvalid) Error() string {
	return fmt.Sprintf("%s: %s (must be a %s)", e.flag, e.err, e.kind)
}

func (e ErrFlagInvalid) Unwrap() error { return e.err }

type Flags struct {
	Program string   // Program name.
	Args    []string // List of arguments, after parsing this will be reduces to non-flags.

	flags []flagValue
}

type flagValue struct {
	names []string
	value interface{}
}

type setter interface{ Set() bool }

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

// Sentinel return values for ShiftCommand()
const (
	CommandNoneGiven = "\x00"
	CommandAmbiguous = "\x01"
	CommandUnknown   = "\x02"
)

// ShiftCommand shifts a value from the argument list, and matches it with the
// list of commands.
//
// Commands can be matched as an abbreviation as long as it's unambiguous; if
// you have "search" and "identify" then "i", "id", etc. will all return
// "identify".
//
// If you have the commands "search" and "see", then "s" or "se" are ambiguous,
// and it will return the special CommandAmbiguous sentinel value.
//
// Commands can also contain aliases as "alias=cmd"; for example "ci=commit".
//
// It will return CommandNoneGiven if there is no command, and CommandUnknown if
// the command is not found.
func (f *Flags) ShiftCommand(cmds ...string) string {
	cmd := f.Shift()
	if cmd == "" {
		return CommandNoneGiven
	}
	cmd = strings.ToLower(cmd)

	var found string
	for _, c := range cmds {
		if c == cmd {
			return cmd
		}

		if strings.HasPrefix(c, cmd) {
			if found != "" {
				return CommandAmbiguous
			}
			if i := strings.IndexRune(c, '='); i > -1 {
				c = c[i+1:]
			}
			found = c
		}
	}

	if found == "" {
		return CommandUnknown
	}
	return found
}

func (f *Flags) Parse() error {
	// Modify f.Args to split out grouped boolean values: "prog -ab" becomes
	// "prog -a -b"
	args := make([]string, 0, len(f.Args))
	for _, arg := range f.Args {
		if !strings.HasPrefix(arg, "-") || arg == "-" {
			args = append(args, arg)
			continue
		}

		if len(strings.TrimLeft(arg, "-")) <= 1 {
			args = append(args, arg)
			continue
		}

		_, ok := f.match(arg)
		if ok {
			args = append(args, arg)
			continue
		}

		split := strings.Split(arg[1:], "")
		found := true
		for _, s := range split {
			_, ok := f.match(s)
			if !ok {
				found = false
				break
			}
		}
		if !found {
			return &ErrFlagUnknown{arg}
		}
		for _, s := range split {
			args = append(args, "-"+s)
		}
	}
	f.Args = args

	var (
		p    []string
		skip bool
	)
	for i, a := range f.Args {
		if skip {
			skip = false
			continue
		}

		if a == "" || a == "-" || a[0] != '-' {
			p = append(p, a)
			continue
		}

		if a == "--" {
			p = append(p, f.Args[i+1:]...)
			break
		}

		flag, ok := f.match(a)
		if !ok {
			return &ErrFlagUnknown{a}
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

		// TODO: it might make more sense to have two interfaces: singleSetter
		// and multiSetter.
		if set := flag.value.(setter); set.Set() {
			switch flag.value.(type) {
			case flagIntCounter, flagStringList, flagBool: // Not an error.
			default:
				return &ErrFlagDouble{a}
			}
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
			x, err := strconv.ParseInt(val, 0, 64)
			if err != nil {
				if nErr := errors.Unwrap(err); nErr != nil {
					err = nErr
				}
				return ErrFlagInvalid{a, err, "number"}
			}
			*v.v = int(x)
		case flagInt64:
			val, *v.s = next()
			x, err := strconv.ParseInt(val, 0, 64)
			if err != nil {
				if nErr := errors.Unwrap(err); nErr != nil {
					err = nErr
				}
				return ErrFlagInvalid{a, err, "number"}
			}
			*v.v = x
		case flagFloat64:
			val, *v.s = next()
			x, err := strconv.ParseFloat(val, 64)
			if err != nil {
				if nErr := errors.Unwrap(err); nErr != nil {
					err = nErr
				}
				return ErrFlagInvalid{a, err, "number"}
			}
			*v.v = x
		case flagIntCounter:
			*v.s = true
			*v.v++
		case flagStringList:
			n, s := next()
			*v.s = s
			*v.v = append(*v.v, n)
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
	flagStringList struct {
		v *[]string
		s *bool
	}
)

func (f flagBool) Pointer() *bool           { return f.v }
func (f flagString) Pointer() *string       { return f.v }
func (f flagInt) Pointer() *int             { return f.v }
func (f flagInt64) Pointer() *int64         { return f.v }
func (f flagFloat64) Pointer() *float64     { return f.v }
func (f flagIntCounter) Pointer() *int      { return f.v }
func (f flagStringList) Pointer() *[]string { return f.v }

func (f flagBool) Bool() bool              { return *f.v }
func (f flagString) String() string        { return *f.v }
func (f flagInt) Int() int                 { return *f.v }
func (f flagInt64) Int64() int64           { return *f.v }
func (f flagFloat64) Float64() float64     { return *f.v }
func (f flagIntCounter) Int() int          { return *f.v }
func (f flagStringList) Strings() []string { return *f.v }

func (f flagBool) Set() bool       { return *f.s }
func (f flagString) Set() bool     { return *f.s }
func (f flagInt) Set() bool        { return *f.s }
func (f flagInt64) Set() bool      { return *f.s }
func (f flagFloat64) Set() bool    { return *f.s }
func (f flagIntCounter) Set() bool { return *f.s }
func (f flagStringList) Set() bool { return *f.s }

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

func (f *Flags) StringList(def []string, name string, aliases ...string) flagStringList {
	v := flagStringList{v: &def, s: new(bool)}
	f.append(v, name, aliases...)
	return v
}
