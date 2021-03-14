package zli

import (
	"bytes"
	"os"
	"testing"
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
//
// TODO: this isn't thread-safe in cases where os.Exit() gets called from a
// goroutine in the running program.
//
// type TextExit struct {
//   mu    *sync.Mutex
//   exits []int
// }
type TestExit int

// Exit sets TestExit to the given status code and panics with itself.
func (t *TestExit) Exit(c int) {
	*t = TestExit(c)
	panic(t)
}

// Want checks that the recorded exit code matches the given code and issues a
// t.Error() if it doesn't.
func (t *TestExit) Want(tt *testing.T, c int) {
	tt.Helper()
	if int(*t) != c {
		tt.Errorf("wrong exit: %d; want: %d", *t, c)
	}
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
// The state will be reset when the test finishes.
//
// The code points to the latest zli.Exit() return code.
func Test(t *testing.T) (exit *TestExit, in, out *bytes.Buffer) {
	in = new(bytes.Buffer)
	Stdin = in

	out = new(bytes.Buffer)
	Stdout = out
	Stderr = out

	exit = new(TestExit)
	*exit = -1
	Exit = exit.Exit

	t.Cleanup(func() {
		Exit = os.Exit
		Stdin = os.Stdin
		Stdout = os.Stdout
		Stderr = os.Stderr
	})

	return exit, in, out
}
