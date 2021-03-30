package zli

import (
	"bytes"
	"errors"
	"fmt"
	"io"
	"io/ioutil"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	"golang.org/x/term"
)

var (
	Exit   func(int) = os.Exit
	Stdin  io.Reader = os.Stdin
	Stdout io.Writer = os.Stdout
	Stderr io.Writer = os.Stderr
)

// IsTerminal reports if this file descriptor is an interactive terminal.
//
// TODO: this is a bit tricky now, as we can replace zli.Stdout with something
// else; checking os.Stdout may not be correct in those cases.
var IsTerminal = func(fd uintptr) bool { return term.IsTerminal(int(fd)) }

// TerminalSize gets the dimensions of the given terminal.
var TerminalSize = func(fd uintptr) (width, height int, err error) { return term.GetSize(int(fd)) }

// Program gets the program name from argv.
func Program() string {
	if len(os.Args) == 0 {
		return ""
	}
	return filepath.Base(os.Args[0])
}

// Error prints an error message to stderr prepended with the program name and
// with a newline appended.
func Errorf(s interface{}, args ...interface{}) {
	prog := Program()
	if prog != "" {
		prog += ": "
	}

	switch ss := s.(type) {
	case string:
		fmt.Fprintf(Stderr, prog+ss+"\n", args...)
	case []byte:
		fmt.Fprintf(Stderr, prog+string(ss)+"\n", args...)
	case error:
		if len(args) > 0 {
			fmt.Fprintf(Stderr, "%s%s %v\n", prog, ss.Error(), args)
		} else {
			fmt.Fprintln(Stderr, prog+ss.Error())
		}
	default:
		if len(args) > 0 {
			fmt.Fprintf(Stderr, prog+"%v %v\n", ss, args)
		} else {
			fmt.Fprintf(Stderr, prog+"%v\n", ss)
		}
	}
}

// ExitCode is the exit code to use for Fatalf() and F()
var ExitCode = 1

// Fatalf is like Errorf(), but will exit with a code of 1.
func Fatalf(s interface{}, args ...interface{}) {
	Errorf(s, args...)
	Exit(ExitCode)
}

// F prints the err.Error() to stderr with Errorf() and exits, but it won't do
// anything if the error is nil.
func F(err error) {
	if err != nil {
		Fatalf(err)
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
func InputOrArgs(args []string, sep string, quiet bool) ([]string, error) {
	if len(args) > 0 {
		return args, nil
	}

	interactive := IsTerminal(os.Stdin.Fd())

	if !quiet && interactive {
		fmt.Fprintf(Stderr, "%s: reading from stdin...", Program())
		os.Stderr.Sync()
	}
	in, err := ioutil.ReadAll(Stdin)
	if err != nil {
		return nil, fmt.Errorf("zli.InputOrArgs: read stdin: %w", err)
	}
	if !quiet && interactive {
		fmt.Fprintf(Stderr, "\r")
	}

	in = bytes.Trim(bytes.TrimSuffix(in, []byte("\n")), sep)
	return strings.FieldsFunc(string(in), func(c rune) bool {
		return strings.ContainsRune(sep, c)
	}), nil
}

// PagerStdout replaces Stdout with a buffer and pipes the content of it to
// $PAGER.
//
// The typical way to use this is at the start of a function like so:
//
//    defer zli.PageStdout()()
//
// You need to be a bit careful when calling Exit() explicitly, since that will
// exit immediately without running any defered functions. You have to either
// use a wrapper or call the returned function explicitly.
func PagerStdout() func() {
	buf := new(bytes.Buffer)
	save := Stdout
	Stdout = buf
	return func() {
		Stdout = save
		Pager(buf)
	}
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
		Errorf("zli.Pager: running $PAGER: %s", err)
		io.Copy(Stdout, text)
		return
	}

	cmd := exec.Command(pager, args...)
	cmd.Stdin = text
	cmd.Stdout = Stdout
	cmd.Stderr = Stderr

	err = cmd.Start()
	if err != nil {
		Errorf("zli.Pager: running $PAGER: %s", err)
		io.Copy(Stdout, text)
		return
	}

	err = cmd.Wait()
	if err != nil {
		var exitErr *exec.ExitError
		if errors.As(err, &exitErr) && !exitErr.Success() {
			// We're not sure if the program actually did something, so don't
			// copy the text here.
			Errorf("zli.Pager: running $PAGER: %s", err)
		}
	}
}
