// Copyright 2019 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package term

import (
	"os"
	"syscall"
	"unsafe"
)

// Console related constants used for the mode parameter to SetConsoleMode. See
// https://docs.microsoft.com/en-us/windows/console/setconsolemode for details.

const (
	ENABLE_PROCESSED_INPUT        = 0x1
	ENABLE_LINE_INPUT             = 0x2
	ENABLE_ECHO_INPUT             = 0x4
	ENABLE_WINDOW_INPUT           = 0x8
	ENABLE_MOUSE_INPUT            = 0x10
	ENABLE_INSERT_MODE            = 0x20
	ENABLE_QUICK_EDIT_MODE        = 0x40
	ENABLE_EXTENDED_FLAGS         = 0x80
	ENABLE_AUTO_POSITION          = 0x100
	ENABLE_VIRTUAL_TERMINAL_INPUT = 0x200

	ENABLE_PROCESSED_OUTPUT            = 0x1
	ENABLE_WRAP_AT_EOL_OUTPUT          = 0x2
	ENABLE_VIRTUAL_TERMINAL_PROCESSING = 0x4
	DISABLE_NEWLINE_AUTO_RETURN        = 0x8
	ENABLE_LVB_GRID_WORLDWIDE          = 0x10
)

var (
	modkernel32                    = syscall.NewLazyDLL("kernel32.dll")
	procSetConsoleMode             = modkernel32.NewProc("SetConsoleMode")
	procGetConsoleScreenBufferInfo = modkernel32.NewProc("GetConsoleScreenBufferInfo")
)

func setConsoleMode(console syscall.Handle, mode uint32) (err error) {
	r1, _, e1 := syscall.Syscall(procSetConsoleMode.Addr(), 2, uintptr(console), uintptr(mode), 0)
	if r1 == 0 {
		err = errnoErr(e1)
	}
	return
}

// Used with GetConsoleScreenBuffer to retrieve information about a console
// screen buffer. See
// https://docs.microsoft.com/en-us/windows/console/console-screen-buffer-info-str
// for details.

type coord struct{ X, Y int16 }
type smallRect struct{ Left, Top, Right, Bottom int16 }
type consoleScreenBufferInfo struct {
	Size              coord
	CursorPosition    coord
	Attributes        uint16
	Window            smallRect
	MaximumWindowSize coord
}

func getConsoleScreenBufferInfo(console syscall.Handle, info *consoleScreenBufferInfo) (err error) {
	r1, _, e1 := syscall.Syscall(procGetConsoleScreenBufferInfo.Addr(), 2, uintptr(console), uintptr(unsafe.Pointer(info)), 0)
	if r1 == 0 {
		err = errnoErr(e1)
	}
	return
}

// Do the interface allocations only once for common
// Errno values.
const (
	errnoERROR_IO_PENDING = 997
)

var (
	errERROR_IO_PENDING error = syscall.Errno(errnoERROR_IO_PENDING)
	errERROR_EINVAL     error = syscall.EINVAL
)

// errnoErr returns common boxed Errno values, to prevent
// allocations at runtime.
func errnoErr(e syscall.Errno) error {
	switch e {
	case 0:
		return errERROR_EINVAL
	case errnoERROR_IO_PENDING:
		return errERROR_IO_PENDING
	}
	// TODO: add more here, after collecting data on the common
	// error values see on Windows. (perhaps when running
	// all.bat?)
	return e
}

// ---

type state struct {
	mode uint32
}

func isTerminal(fd int) bool {
	var st uint32
	err := syscall.GetConsoleMode(syscall.Handle(fd), &st)
	return err == nil
}

func makeRaw(fd int) (*State, error) {
	var st uint32
	if err := syscall.GetConsoleMode(syscall.Handle(fd), &st); err != nil {
		return nil, err
	}
	raw := st &^ (ENABLE_ECHO_INPUT | ENABLE_PROCESSED_INPUT | ENABLE_LINE_INPUT | ENABLE_PROCESSED_OUTPUT)
	if err := setConsoleMode(syscall.Handle(fd), raw); err != nil {
		return nil, err
	}
	return &State{state{st}}, nil
}

func getState(fd int) (*State, error) {
	var st uint32
	if err := syscall.GetConsoleMode(syscall.Handle(fd), &st); err != nil {
		return nil, err
	}
	return &State{state{st}}, nil
}

func restore(fd int, state *State) error {
	return setConsoleMode(syscall.Handle(fd), state.mode)
}

func getSize(fd int) (width, height int, err error) {
	var info consoleScreenBufferInfo
	if err := getConsoleScreenBufferInfo(syscall.Handle(fd), &info); err != nil {
		return 0, 0, err
	}
	return int(info.Window.Right - info.Window.Left + 1), int(info.Window.Bottom - info.Window.Top + 1), nil
}

func readPassword(fd int) ([]byte, error) {
	var st uint32
	if err := syscall.GetConsoleMode(syscall.Handle(fd), &st); err != nil {
		return nil, err
	}
	old := st

	st &^= (ENABLE_ECHO_INPUT | ENABLE_LINE_INPUT)
	st |= (ENABLE_PROCESSED_OUTPUT | ENABLE_PROCESSED_INPUT)
	if err := setConsoleMode(syscall.Handle(fd), st); err != nil {
		return nil, err
	}

	defer setConsoleMode(syscall.Handle(fd), old)

	var h syscall.Handle
	p, _ := syscall.GetCurrentProcess()
	if err := syscall.DuplicateHandle(p, syscall.Handle(fd), p, &h, 0, false, syscall.DUPLICATE_SAME_ACCESS); err != nil {
		return nil, err
	}

	f := os.NewFile(uintptr(h), "stdin")
	defer f.Close()
	return readPasswordLine(f)
}
