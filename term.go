package zli

import (
	"bytes"
	"fmt"
	"syscall"

	"golang.org/x/term"
)

// AskPassword interactively asks the user for a password and confirmation.
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
