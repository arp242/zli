//go:build !no_term
// +build !no_term

package zli

import (
	"bytes"
	"errors"
	"fmt"
	"os"
	"runtime"
	"syscall"

	"golang.org/x/sys/unix"
	"golang.org/x/term"
)

// IsTerminal reports if this file descriptor is an interactive terminal.
//
// TODO: this is a bit tricky now, as we can replace zli.Stdout with something
// else; checking os.Stdout may not be correct in those cases.
var IsTerminal = func(fd uintptr) bool { return term.IsTerminal(int(fd)) }

// TerminalSize gets the dimensions of the given terminal.
var TerminalSize = func(fd uintptr) (width, height int, err error) { return term.GetSize(int(fd)) }

// WantColor indicates if the program should output any colors. This is
// automatically set from from the output terminal and NO_COLOR environment
// variable.
//
// You can override this if the user sets "--color=force" or the like.
//
// TODO: maybe expand this a bit with WantMonochrome or some such, so you can
// still output bold/underline/reverse text for people who don't want colors.
var WantColor = func() bool {
	_, ok := os.LookupEnv("NO_COLOR")
	return os.Getenv("TERM") != "dumb" && term.IsTerminal(int(os.Stdout.Fd())) && !ok
}()

// AskPassword interactively asks the user for a password and confirmation.
//
// Just a convenient wrapper for term.ReadPassword() to call it how you want to
// use it much of the time to ask for a new password.
func AskPassword(minlen int) (string, error) {
start:
	fmt.Fprintf(Stdout, "Enter password for new user (will not echo): ")
	pwd1, err := term.ReadPassword(int(syscall.Stdin))
	if err != nil {
		return "", err
	}
	if len(pwd1) < minlen {
		fmt.Fprintf(Stdout, "\nNeed at least %d characters\n", minlen)
		goto start
	}

	fmt.Fprintf(Stdout, "\nConfirm: ")
	pwd2, err := term.ReadPassword(int(syscall.Stdin))
	if err != nil {
		return "", err
	}
	fmt.Fprintln(Stdout, "")

	if !bytes.Equal(pwd1, pwd2) {
		fmt.Fprintln(Stdout, "Passwords did not match; try again.")
		goto start
	}

	return string(pwd1), nil
}

// RawTerminal sets the terminal to "raw" mode.
//
// The returned function restores the terminal to the previous state.
func RawTerminal() (func() error, error) {
	fd := int(os.Stdout.Fd())
	old, err := term.MakeRaw(fd)
	return func() error { return term.Restore(fd, old) }, err
}

const ioctlReadTermios = unix.TCGETS

func IsRawTerminal() bool {
	fd := int(os.Stdout.Fd())
	termios, _ := unix.IoctlGetTermios(fd, ioctlReadTermios)
	return termios.Lflag&unix.ICANON == 0
}

type KeyEvent struct {
	Key Key
	Err error
}

// ReadKeys reads keys from stdin.
func ReadKeys() (chan KeyEvent, error) {
	if !IsRawTerminal() {
		return nil, errors.New("zli.ReadKeys: need to operate on raw terminal")
	}

	//tty, err := syscall.Open("/dev/tty", unix.O_RDWR, 0)
	tty, err := syscall.Open("/dev/tty", unix.O_RDONLY, 0)
	if err != nil {
		return nil, fmt.Errorf("zli.ReadKeys: open /dev/tty: %w", err)
	}

	_, err = unix.FcntlInt(uintptr(tty), unix.F_SETOWN, unix.Getpid())
	if err != nil && runtime.GOOS != "darwin" {
		// termbox has this Darwin check; dunno why, but I bet it's for a
		// reason, so add it here too.
		return nil, fmt.Errorf("zli.ReadKeys: set owner: %w", err)
	}

	keys := make(chan KeyEvent)
	go func() {
		for {
			buf := make([]byte, 32)
			n, err := syscall.Read(tty, buf)
			if err != nil {
				keys <- KeyEvent{Err: err}
			}

			skip := 0
			for _, c := range buf[:n] {
				if skip > 0 {
					skip--
					continue
				}
				// if c == 0x1b {
				// 	skip = 2
				// 	keys <- KeyEvent{String: string(buf[i : i+3])}
				// 	continue
				// }
				keys <- KeyEvent{Key: Key(c)}
			}
		}
	}()

	return keys, nil
}

// CursorPosition gets the current cursor position.
func CursorPosition() (int, int, error) {
	if IsRawTerminal() {
		// TODO: do something with this.
		//sendCSI("6n")
		return 0, 0, nil
	}

	restore, err := RawTerminal()
	if err != nil {
		return 0, 0, err
	}

	sendCSI("6n")

	buf := make([]byte, 128)
	n, err := os.Stdout.Read(buf)
	if err != nil {
		return 0, 0, err
	}
	buf = buf[:n]

	err = restore()
	if err != nil {
		return 0, 0, err
	}

	var pushback []byte
	if i := bytes.Index(buf, []byte{0x1b, '['}); i > 0 {
		pushback = append(pushback, buf[:i]...)
		buf = buf[i:]
	}
	if i := bytes.IndexByte(buf, 'R'); i != len(buf)-1 {
		pushback = append(pushback, buf[i+1:]...)
		buf = buf[:i+1]
	}

	var r, c int
	n, _ = fmt.Sscanf(string(buf), "\x1b[%d;%dR", &r, &c)

	if len(pushback) > 0 {
		os.Stdout.Write(pushback)
	}
	return r, c, nil
}
