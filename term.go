package zli

import (
	"bytes"
	"fmt"
	"os"
	"syscall"

	"zgo.at/zli/internal/term"
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
