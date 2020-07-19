package zli

import (
	"bytes"
	"os"
	"unsafe"
)

// TestExit records the exit code and aborts the normal program execution. It's
// intended to test exit codes in a program.
//
// The Exit() method call is a replacement for zli.Exit:
//
//     exit := TestExit(-1)
//     Exit = exit.Exit
//     defer func() { Exit = os.Exit }()
//
// This can be recovered like so:
//
//     func() {
//         defer exit.Recover()
//         Fatal("oh noes!")
//     }()
//     fmt.Println("Exit", exit)
//
// The function wrapper is needed so that the test function itself doesn't get
// aborted.
type TestExit int

// Exit sets TestExit to the given status code and panics with itself.
func (t *TestExit) Exit(c int) {
	*t = TestExit(c)
	panic(t)
}

// Recover any panics where the argument is this TestExit instance. it will
// re-panic on any other errors (including other TestExit instances).
func (t *TestExit) Recover() {
	r := recover()
	if r == nil {
		return
	}
	exit, ok := r.(*TestExit)
	if !ok || unsafe.Pointer(t) != unsafe.Pointer(exit) {
		panic(r)
	}
}

// Test replaces Stdin, Stdout, Stderr, and Exit for testing.
//
// The code points to the latest zli.Exit() return code;
func Test() (exit *TestExit, in, out *bytes.Buffer, reset func()) {
	in = new(bytes.Buffer)
	Stdin = in

	out = new(bytes.Buffer)
	Stdout = out
	Stderr = out

	exit = new(TestExit)
	*exit = -1
	Exit = exit.Exit

	return exit, in, out, func() {
		Exit = os.Exit
		Stdin = os.Stdin
		Stdout = os.Stdout
		Stderr = os.Stderr
	}
}
