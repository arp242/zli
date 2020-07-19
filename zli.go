package zli

import (
	"bytes"
	"fmt"
	"io"
	"io/ioutil"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	"zgo.at/zli/internal/isatty"
	"zgo.at/zli/internal/terminal"
)

var (
	Exit   func(int) = os.Exit
	Stdin  io.Reader = os.Stdin
	Stdout io.Writer = os.Stdout
	Stderr io.Writer = os.Stderr
)

// IsTerminal reports if this file descriptor is an interactive terminal.
func IsTerminal(fd uintptr) bool { return isatty.IsTerminal(fd) }

// TerminalSize gets the dimensions of the given terminal.
func TerminalSize(fd uintptr) (width, height int, err error) { return terminal.GetSize(int(fd)) }

// Program gets the program name from argv.
func Program() string {
	if len(os.Args) == 0 {
		return ""
	}
	return filepath.Base(os.Args[0])
}

// Error prints an error message to stderr.
//
//   Error("oh noes: %q", something)   // printf arguments
//   Error(err)                        // Print err.Error()
//   Error(123)                        // Print %v (makes little sense, but okay)
func Error(s interface{}, args ...interface{}) {
	prog := Program()
	if prog != "" {
		prog += ": "
	}

	switch ss := s.(type) {
	case error:
		if len(args) > 0 {
			fmt.Fprintf(Stderr, "%s%s %v\n", prog, ss.Error(), args)
		} else {
			fmt.Fprintf(Stderr, prog+ss.Error()+"\n")
		}
	case string:
		fmt.Fprintf(Stderr, prog+ss+"\n", args...)
	default:
		if len(args) > 0 {
			fmt.Fprintf(Stderr, prog+"%v %v\n", ss, args)
		} else {
			fmt.Fprintf(Stderr, prog+"%v\n", ss)
		}
	}
}

// Fatal prints the given message to stderr with Error() and exits.
func Fatal(s interface{}, args ...interface{}) {
	Error(s, args...)
	Exit(1)
}

// F is like Fatal(), but won't do anything for nil errors.
//
// This is mostly intended for quick tests/scripts and the like; you need to be
// careful with this as the following is hard to test:
//
//   err := f()
//   zli.F(err)
//
// Because while you can swap out the exit with Test(), the code will still
// continue. Use Fatal() followed by a return or Error() instead if you want to
// test the code.
func F(err error) {
	if err != nil {
		Fatal(err)
	}
}

// InputOrFile will return a reader connected to stdin if path is "" or "-", or
// open a path for any other value.
//
// It will print a message to stderr notifying the user it's reading from stdin
// if the terminal is interactive and quiet is false.
// See: https://www.arp242.net/read-stdin.html
func InputOrFile(path string, quiet bool) (io.ReadCloser, error) {
	if path != "" && path != "-" {
		fp, err := os.Open(path)
		if err != nil {
			err = fmt.Errorf("zli.InputOrFile: %w", err)
		}
		return fp, err
	}

	if !quiet && IsTerminal(os.Stdin.Fd()) {
		fmt.Fprintf(Stderr, "%s: reading from stdin...\r", Program())
		os.Stderr.Sync()
	}
	return ioutil.NopCloser(Stdin), nil
}

// InputOrArgs reads arguments separated by sep from stdin if args is empty, or
// returns args unmodified if it's not.
//
// The argument are split on newline; the following are all identical:
//
//   prog foo bar
//   printf "foo\nbar\n" | prog
//
//   prog 'foo bar' 'x y'
//   printf "foo bar\nx y\n" | prog
//
// It will print a message to stderr notifying the user it's reading from stdin
// if the terminal is interactive and quiet is false.
// See: https://www.arp242.net/read-stdin.html
func InputOrArgs(args []string, quiet bool) ([]string, error) {
	if len(args) > 0 {
		return args, nil
	}

	interactive := IsTerminal(os.Stdin.Fd())

	if !quiet && interactive {
		fmt.Fprintf(Stderr, "%s: reading from stdin...", Program())
		os.Stderr.Sync()
	}
	in, err := ioutil.ReadAll(os.Stdin)
	if err != nil {
		return nil, fmt.Errorf("zli.InputOrArgs: read stdin: %w", err)
	}
	if !quiet && interactive {
		fmt.Fprintf(Stderr, "\r")
	}

	//for _, l := range strings.Split(strings.TrimRight(string(stdin), "\n"), "\n") {
	//	args = append(args, strings.Split(l, " ")...)
	//}

	//return strings.Fields(string(bytes.TrimRight(in, "\n"))), nil
	in = bytes.TrimSuffix(in, []byte("\n"))
	return strings.Split(string(in), "\n"), nil
}

// Pager pipes the content of text to $PAGER, or prints it to stdout of this
// fails.
func Pager(text io.Reader) {
	if !IsTerminal(os.Stdout.Fd()) {
		io.Copy(Stdout, text)
		return
	}

	pager := os.Getenv("PAGER")
	if pager == "" {
		io.Copy(Stdout, text)
		return
	}

	var args []string
	if i := strings.IndexByte(pager, ' '); i > -1 {
		args = strings.Split(pager[i+1:], " ")
		pager = pager[:i]
	}

	pager, err := exec.LookPath(pager)
	if err != nil {
		fmt.Fprintf(Stderr, "running $PAGER: %s\n", err)
		io.Copy(Stdout, text)
		return
	}

	cmd := exec.Command(pager, args...)
	cmd.Stdin = text
	cmd.Stdout = Stdout
	cmd.Stderr = Stderr

	err = cmd.Start()
	if err != nil {
		fmt.Fprintf(os.Stderr, "running $PAGER: %s\n", err)
		io.Copy(Stdout, text)
		return
	}

	_ = cmd.Wait()
}
