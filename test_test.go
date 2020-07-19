package zli

import (
	"fmt"
	"io/ioutil"
	"os"
	"testing"
)

func TestTestExit(t *testing.T) {
	exit := TestExit(-1)
	Exit = exit.Exit
	defer func() { Exit = os.Exit }()

	func() {
		defer exit.Recover()
	}()
	if exit != -1 {
		t.Errorf("unexpected code: %d", exit)
	}

	func() {
		defer exit.Recover()
		Fatal("oh noes!")
	}()
	if exit != 1 {
		t.Errorf("unexpected code: %d", exit)
	}
	fmt.Println("Exit", exit)
}

func TestTest(t *testing.T) {
	exit, in, out, reset := Test()
	defer reset()

	Error("oh noes!")
	if out.String() != "zli.test: oh noes!\n" {
		t.Errorf("wrong stderr: %q", out.String())
	}

	in.WriteString("Hello")
	fp, _ := InputOrFile("-", true)
	got, _ := ioutil.ReadAll(fp)
	if string(got) != "Hello" {
		t.Errorf("wrong stdin: %q", string(got))
	}

	out.Reset()

	et := func() {
		fmt.Fprintln(Stdout, "ET START")
		Exit(1)
		fmt.Fprintln(Stdout, "ET END")
	}

	func() {
		defer exit.Recover()
		et()
	}()
	if *exit != 1 {
		t.Error("wrong exit")
	}
	if out.String() != "ET START\n" {
		t.Errorf("wrong stderr: %q", out.String())
	}
}
