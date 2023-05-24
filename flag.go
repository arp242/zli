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

	// ErrPositional is used when there are too few or too many positional
	// arguments.
	ErrPositional struct {
		min, max, n int
	}
)

func (e ErrFlagInvalid) Unwrap() error { return e.err }
func (e ErrFlagInvalid) Error() string {
	return fmt.Sprintf("%s: %s (must be a %s)", e.flag, e.err, e.kind)
}
func (e ErrFlagUnknown) Error() string { return fmt.Sprintf("unknown flag: %q", e.flag) }
func (e ErrFlagDouble) Error() string  { return fmt.Sprintf("flag given more than once: %q", e.flag) }
func (e ErrPositional) Error() string {
	pl := func(n int) string {
		if n == 1 {
			return "argument"
		}
		return "arguments"
	}
	switch {
	case e.min == e.max:
		return fmt.Sprintf("exactly %d positional %s required, but %d given", e.min, pl(e.min), e.n)
	case e.max == 0 && e.min > 0:
		return fmt.Sprintf("at least %d positional %s required, but %d given", e.min, pl(e.min), e.n)
	case e.min == 0 && e.max > 0:
		return fmt.Sprintf("at most %d positional %s accepted, but %d given", e.max, pl(e.max), e.n)
	default:
		return fmt.Sprintf("between %d and %d positional arguments accepted, but %d given", e.min, e.max, e.n)
	}
}

// Flags are a set of parsed flags.
//
// The rules for parsing are as follows:
//
//   - Flags start with one or more '-'s; '-a' and '--a' are identical, as are
//     '-long' and '--long'.
//
//   - Flags are separated with arguments by one space or '='. This is required:
//     '-vVALUE' is invalid; you must use '-v VALUE' or '-v=VALUE'.
//
//   - Single-letter flags can be grouped; '-ab' is identical to '-a -b', and
//     '-ab VAL' is identical to '-a -b VAL'. "Long" flags cannot be grouped.
//
//   - Long flag names take precedence over single-letter ones, e.g. if you
//     define the flags '-long', '-l', '-o', '-n', and '-g' then '-long' will be
//     parsed as '-long'.
//
//   - Anything that doesn't start with a '-' or follows '--' is treated as a
//     positional argument. This can be freely interspersed with flags.
type Flags struct {
	Program string   // Program name.
	Args    []string // List of arguments, after parsing this will be reduces to non-flags.

	flags    []flagValue
	optional bool
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
type (
	ErrCommandNoneGiven struct{}
	ErrCommandUnknown   string
	ErrCommandAmbiguous struct {
		Cmd  string
		Opts []string
	}
)

func (e ErrCommandNoneGiven) Error() string { return "no command given" }
func (e ErrCommandUnknown) Error() string   { return fmt.Sprintf("unknown command: %q", string(e)) }
func (e ErrCommandAmbiguous) Error() string {
	return fmt.Sprintf(`ambigious command: %q; matches: "%s"`, e.Cmd, strings.Join(e.Opts, `", "`))
}

// ShiftCommand shifts the first non-flag value from the argument list.
//
// This can work both before or after f.Parse(); this is useful if you want to
// have different flags for different arguments, and both of these will work:
//
//	$ prog -flag cmd
//	$ prog cmd -flag
//
// If cmds is given then it matches commands with this list; commands can be
// matched as an abbreviation as long as it's unambiguous; if you have "search"
// and "identify" then "i", "id", etc. will all return "identify". If you have
// the commands "search" and "see", then "s" or "se" are ambiguous, and it will
// return an ErrCommandAmbiguous error.
//
// Commands can also contain aliases as "alias=cmd"; for example "ci=commit".
//
// It will return ErrCommandNoneGiven if there is no command, and
// ErrCommandUnknown if the command is not found.
func (f *Flags) ShiftCommand(cmds ...string) (string, error) {
	var (
		pushback []string
		cmd      string
	)
	for {
		cmd = f.Shift()
		if cmd == "" {
			return "", ErrCommandNoneGiven{}
		}
		if cmd[0] == '-' || strings.ContainsRune(cmd, '=') {
			pushback = append(pushback, cmd)
			continue
		}

		break
	}
	f.Args = append(pushback, f.Args...)
	cmd = strings.ToLower(cmd)

	if len(cmds) == 0 {
		return cmd, nil
	}

	var found []string
	for _, c := range cmds {
		if c == cmd {
			return cmd, nil
		}

		if strings.HasPrefix(c, cmd) {
			if i := strings.IndexRune(c, '='); i > -1 { // Alias
				c = c[i+1:]
			}
			found = append(found, c)
		}
	}

	switch len(found) {
	case 0:
		return "", ErrCommandUnknown(cmd)
	case 1:
		return found[0], nil
	default:
		return "", ErrCommandAmbiguous{Cmd: cmd, Opts: found}
	}
}

var (
	// AllowUnknown indicates that unknown flags are not an error; unknown flags
	// are added to the Args list.
	//
	// This is useful if you have subcommands with different flags, for example:
	//
	//     f := zli.NewFlags(os.Args)
	//     globalFlag := f.String(..)
	//     f.Parse(zli.AllowUnknown())
	//
	//     switch cmd := f.ShiftCommand(..) {
	//     case "serve":
	//         serveFlag := f.String(..)
	//         f.Parse()   // *Will* error out on unknown flags.
	//     }
	AllowUnknown = func() parseOpt { return func(o *parseOpts) { o.allowUnknown = true } }

	// Positional sets the lower and upper bounds for the number of positional
	// arguments.
	//
	//   Positional(0, 0)     No limit and accept everything (the default).
	//   Positional(1, 0)     Must have at least one positional argument.
	//   Positional(1, 1)     Must have exactly one positional argument.
	//   Positional(0, 3)     May optionally have up to three positional arguments.
	//   Positional(-1, 0)    Don't accept any conditionals (the max is ignored).only the min is
	Positional = func(min, max int) parseOpt { return func(o *parseOpts) { o.pos = [2]int{min, max} } }

	// NoPositional is a shortcut for Positional(-1, 0)
	NoPositional = func() parseOpt { return func(o *parseOpts) { o.pos = [2]int{-1, -1} } }
)

type (
	parseOpts struct {
		allowUnknown bool
		pos          [2]int
	}
	parseOpt func(*parseOpts)
)

// Parse the set of flags in f.Args.
func (f *Flags) Parse(opts ...parseOpt) error {
	var opt parseOpts
	for _, o := range opts {
		o(&opt)
	}

	// Modify f.Args to split out grouped boolean values: "prog -ab" becomes
	// "prog -a -b"
	args := make([]string, 0, len(f.Args))
	for _, arg := range f.Args {
		if !strings.HasPrefix(arg, "-") || arg == "-" {
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
		// "-arg -42"; we reject unknown flags later.
		if !found {
			args = append(args, arg)
			continue
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
			if opt.allowUnknown {
				p = append(p, a)
				continue
			}
			return &ErrFlagUnknown{a}
		}

		var err error
		next := func(opt bool) (string, bool, bool) {
			if j := strings.IndexByte(f.Args[i], '='); j > -1 {
				return f.Args[i][j+1:], true, true
			}
			if i >= len(f.Args)-1 {
				if !opt {
					err = fmt.Errorf("needs an argument")
					return "", false, false
				}
				return "", true, false
			}

			v := f.Args[i+1]
			if len(v) > 1 && v[0] == '-' {
				return "", true, false
			}

			skip = true
			return v, true, true
		}

		// TODO: it might make more sense to have two interfaces: singleSetter
		// and multiSetter.
		if set := flag.value.(setter); set.Set() {
			switch flag.value.(type) {
			case flagIntCounter, flagStringList, flagIntList, flagBool: // Not an error.
			default:
				return &ErrFlagDouble{a}
			}
		}

		var (
			val      string
			hasValue bool
		)
		switch v := flag.value.(type) {
		case flagBool:
			*v.s = true
			*v.v = true
		case flagString:
			val, *v.s, hasValue = next(v.o)
			if hasValue {
				*v.v = val
			}
		case flagInt:
			val, *v.s, hasValue = next(v.o)
			if hasValue {
				x, err := strconv.ParseInt(val, 0, 64)
				if err != nil {
					if nErr := errors.Unwrap(err); nErr != nil {
						err = nErr
					}
					return ErrFlagInvalid{a, err, "number"}
				}
				*v.v = int(x)
			}
		case flagInt64:
			val, *v.s, hasValue = next(v.o)
			if hasValue {
				x, err := strconv.ParseInt(val, 0, 64)
				if err != nil {
					if nErr := errors.Unwrap(err); nErr != nil {
						err = nErr
					}
					return ErrFlagInvalid{a, err, "number"}
				}
				*v.v = x
			}
		case flagFloat64:
			val, *v.s, hasValue = next(v.o)
			if hasValue {
				x, err := strconv.ParseFloat(val, 64)
				if err != nil {
					if nErr := errors.Unwrap(err); nErr != nil {
						err = nErr
					}
					return ErrFlagInvalid{a, err, "number"}
				}
				*v.v = x
			}
		case flagIntCounter:
			*v.s = true
			*v.v++
		case flagStringList:
			if !*v.s {
				*v.v = nil
			}
			n, s, hasValue := next(v.o)
			if hasValue {
				*v.s = s
				*v.v = append(*v.v, n)
			}
		case flagIntList:
			if !*v.s {
				*v.v = nil
			}

			n, s, hasValue := next(v.o)
			if hasValue {
				x, err := strconv.ParseInt(n, 0, 64)
				if err != nil {
					if nErr := errors.Unwrap(err); nErr != nil {
						err = nErr
					}
					return ErrFlagInvalid{a, err, "number"}
				}

				*v.s = s
				*v.v = append(*v.v, int(x))
			}
		}
		if err != nil {
			return fmt.Errorf("%s: %s", a, err)
		}
	}

	if (opt.pos[0] > 0 && len(p) < opt.pos[0]) ||
		(opt.pos[1] > 0 && len(p) > opt.pos[1]) ||
		opt.pos[0] == -1 && len(p) > 0 {
		return ErrPositional{min: opt.pos[0], max: opt.pos[1], n: len(p)}
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
		o bool // Doesn't make much sense here, but just for consistency.
	}
	flagString struct {
		v *string
		s *bool
		o bool
	}
	flagInt struct {
		v *int
		s *bool
		o bool
	}
	flagInt64 struct {
		v *int64
		s *bool
		o bool
	}
	flagFloat64 struct {
		v *float64
		s *bool
		o bool
	}
	flagIntCounter struct {
		v *int
		s *bool
		o bool
	}
	flagStringList struct {
		v *[]string
		s *bool
		o bool
	}
	flagIntList struct {
		v *[]int
		s *bool
		o bool
	}
)

func (f flagBool) Pointer() *bool           { return f.v }
func (f flagString) Pointer() *string       { return f.v }
func (f flagInt) Pointer() *int             { return f.v }
func (f flagInt64) Pointer() *int64         { return f.v }
func (f flagFloat64) Pointer() *float64     { return f.v }
func (f flagIntCounter) Pointer() *int      { return f.v }
func (f flagStringList) Pointer() *[]string { return f.v }
func (f flagIntList) Pointer() *[]int       { return f.v }

func (f flagBool) Bool() bool              { return *f.v }
func (f flagString) String() string        { return *f.v }
func (f flagInt) Int() int                 { return *f.v }
func (f flagInt64) Int64() int64           { return *f.v }
func (f flagFloat64) Float64() float64     { return *f.v }
func (f flagIntCounter) Int() int          { return *f.v }
func (f flagStringList) Strings() []string { return *f.v }
func (f flagIntList) Ints() []int          { return *f.v }

// StringsExpanded returns a list of strings, and every string is split on sep.
//
// This means that these two are identical:
//
//	-skip=foo,bar
//	-skip=foo -skip=bar
func (f flagStringList) StringsSplit(sep string) []string {
	l := make([]string, 0, len(*f.v))
	for _, ll := range *f.v {
		split := strings.Split(ll, sep)
		for i := range split {
			split[i] = strings.TrimSpace(split[i])
		}
		l = append(l, split...)
	}
	return l
}

func (f flagBool) Set() bool       { return *f.s }
func (f flagString) Set() bool     { return *f.s }
func (f flagInt) Set() bool        { return *f.s }
func (f flagInt64) Set() bool      { return *f.s }
func (f flagFloat64) Set() bool    { return *f.s }
func (f flagIntCounter) Set() bool { return *f.s }
func (f flagStringList) Set() bool { return *f.s }
func (f flagIntList) Set() bool    { return *f.s }

func (f *Flags) append(v interface{}, n string, a ...string) {
	for i := range a {
		a[i] = strings.TrimLeft(a[i], "-")
	}
	f.flags = append(f.flags, flagValue{
		value: v,
		names: append([]string{strings.TrimLeft(n, "-")}, a...),
	})
}

// Optional indicates the next flag may optionally have value.
//
// By default String(), Int(), etc. require a value, but with Optional() set
// both "-str" and "-str foo" will work. The default value will be used if
// "-str" was used.
func (f *Flags) Optional() *Flags {
	f.optional = true
	return f
}

// TODO: consider adding a method to automatically generate errors on conflicts;
// for example:
//
//   f.Conflicts(
//      "-json", "-toml",    // These two conflict
//      "cmd1", "-json",     // cmd1 doesn't support -json
//   )
//
// func (f *Flags) Conflicts(args ...string) {
// }

func (f *Flags) Bool(def bool, name string, aliases ...string) flagBool {
	v := flagBool{v: &def, s: new(bool), o: f.optional}
	if f.optional {
		f.optional = false
	}
	f.append(v, name, aliases...)
	return v
}
func (f *Flags) String(def, name string, aliases ...string) flagString {
	v := flagString{v: &def, s: new(bool), o: f.optional}
	if f.optional {
		f.optional = false
	}
	f.append(v, name, aliases...)
	return v
}
func (f *Flags) Int(def int, name string, aliases ...string) flagInt {
	v := flagInt{v: &def, s: new(bool), o: f.optional}
	if f.optional {
		f.optional = false
	}
	f.append(v, name, aliases...)
	return v
}
func (f *Flags) Int64(def int64, name string, aliases ...string) flagInt64 {
	v := flagInt64{v: &def, s: new(bool), o: f.optional}
	if f.optional {
		f.optional = false
	}
	f.append(v, name, aliases...)
	return v
}
func (f *Flags) Float64(def float64, name string, aliases ...string) flagFloat64 {
	v := flagFloat64{v: &def, s: new(bool), o: f.optional}
	if f.optional {
		f.optional = false
	}
	f.append(v, name, aliases...)
	return v
}
func (f *Flags) IntCounter(def int, name string, aliases ...string) flagIntCounter {
	v := flagIntCounter{v: &def, s: new(bool), o: f.optional}
	if f.optional {
		f.optional = false
	}
	f.append(v, name, aliases...)
	return v
}
func (f *Flags) StringList(def []string, name string, aliases ...string) flagStringList {
	v := flagStringList{v: &def, s: new(bool), o: f.optional}
	if f.optional {
		f.optional = false
	}
	f.append(v, name, aliases...)
	return v
}
func (f *Flags) IntList(def []int, name string, aliases ...string) flagIntList {
	v := flagIntList{v: &def, s: new(bool), o: f.optional}
	if f.optional {
		f.optional = false
	}
	f.append(v, name, aliases...)
	return v
}
