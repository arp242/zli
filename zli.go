package zli

import (
	"fmt"
	"io"
	"io/ioutil"
	"os"
	"os/exec"
	"path/filepath"

	"zgo.at/zli/internal/isatty"
	"zgo.at/zli/internal/terminal"
)

type in interface {
	io.Reader
	Fd() uintptr
}

var (
	exit   func(int) = os.Exit
	stdin  in        = os.Stdin
	stdout io.Writer = os.Stdout
	stderr io.Writer = os.Stderr
)

// IsTerminal reports if this file descriptor is an interactive terminal.
func IsTerminal(fd uintptr) bool {
	return isatty.IsTerminal(fd)
}

// TerminalSize gets the dimensions of the given terminal.
func TerminalSize(fd uintptr) (width, height int, err error) {
	return terminal.GetSize(int(fd))
}

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
			fmt.Fprintf(stderr, "%s%s %v\n", prog, ss.Error(), args)
		} else {
			fmt.Fprintf(stderr, prog+ss.Error()+"\n")
		}
	case string:
		fmt.Fprintf(stderr, prog+ss+"\n", args...)
	default:
		if len(args) > 0 {
			fmt.Fprintf(stderr, prog+"%v %v\n", ss, args)
		} else {
			fmt.Fprintf(stderr, prog+"%v\n", ss)
		}
	}
}

// Fatal prints the given message to stderr with Error() and exits.
func Fatal(s interface{}, args ...interface{}) {
	Error(s, args...)
	exit(1)
}

// F is like Fatal(), but won't do anything for nil errors.
func F(err error) {
	if err != nil {
		Fatal(err)
	}
}

// FileOrInput will read from stdin if path is "" or "-", or the path otherwise.
//
// It will print a message to stderr notifying the user it's reading from stdin.
// See: https://www.arp242.net/read-stdin.html
func FileOrInput(path string) (io.ReadCloser, error) {
	if path == "" || path == "-" {
		if IsTerminal(os.Stdin.Fd()) {
			fmt.Fprintf(stderr, "  %s: reading from stdin...\r", Program())
			os.Stderr.Sync()
		}
		return ioutil.NopCloser(stdin), nil
	}

	fp, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	return fp, nil
}

// Pager pipes the content of out to $PAGER, or prints it to stdout of this
// fails.
func Pager(out io.Reader) {
	if !IsTerminal(os.Stdout.Fd()) {
		io.Copy(stdout, out)
		return
	}

	pager := os.Getenv("PAGER")
	if pager == "" {
		io.Copy(stdout, out)
		return
	}

	pager, err := exec.LookPath(pager)
	if err != nil {
		fmt.Fprintf(stderr, "running $PAGER: %s\n", err)
		io.Copy(stdout, out)
		return
	}

	cmd := exec.Command(pager)
	cmd.Stdin = out
	cmd.Stdout = stdout
	cmd.Stderr = stderr

	err = cmd.Start()
	if err != nil {
		fmt.Fprintf(os.Stderr, "running $PAGER: %s\n", err)
		io.Copy(stdout, out)
		return
	}

	_ = cmd.Wait()
}
