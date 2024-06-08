// Copyright 2021 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

//go:build solaris

package term

import (
	"fmt"
	"syscall"
	"unsafe"
)

const ioctlReadTermios = 0x540d  // syscall.TCGETS
const ioctlWriteTermios = 0x540e // syscall.TCSETS

func errnoErr(e syscall.Errno) error { return e }

type syscallFunc uintptr

//go:linkname procioctl libc_ioctl

var (
	procioctl syscallFunc
)

func ioctlPtrRet(fd int, req int, arg unsafe.Pointer) (ret int, err error) {
	r0, _, e1 := sysvicall6(uintptr(unsafe.Pointer(&procioctl)), 3, uintptr(fd), uintptr(req), uintptr(arg), 0, 0, 0)
	ret = int(r0)
	if e1 != 0 {
		err = errnoErr(e1)
	}
	return
}
func ioctlPtr(fd int, req int, arg unsafe.Pointer) (err error) {
	_, err = ioctlPtrRet(fd, req, arg)
	return err
}

type termios struct {
	Iflag uint32
	Oflag uint32
	Cflag uint32
	Lflag uint32
	Cc    [16]uint8
}

// func sysvicall6(trap, nargs, a1, a2, a3, a4, a5, a6 uintptr) (r1, r2 uintptr, err syscall.Errno)

func sysvicall6(trap, nargs, a1, a2, a3, a4, a5, a6 uintptr) (r1, r2 uintptr, err syscall.Errno)

type state struct {
	termios termios
}

// func IoctlGetTermios(fd int, req int) (*termios, error) {
// }

func getTermios(fd int) (termios, error) {
	var t termios
	err := ioctlPtr(fd, ioctlReadTermios, unsafe.Pointer(&t))
	if err != nil {
		return t, fmt.Errorf("getTermios: %s", err)
	}
	return t, nil

	// var t termios
	// _, _, errno := syscall.Syscall(syscall.SYS_IOCTL, uintptr(fd), ioctlReadTermios, uintptr(unsafe.Pointer(&t)))
	// if errno > 0 {
	// 	return t, fmt.Errorf("getTermios: %s", errno)
	// }
	// return t, nil
}

// func IoctlSetTermios(fd int, req int, value *termios) error {
// }
func setTermios(fd int, t termios) error {
	err := ioctlPtr(fd, ioctlWriteTermios, unsafe.Pointer(&t))
	//_, _, errno := syscall.Syscall(syscall.SYS_IOCTL, uintptr(fd), ioctlWriteTermios, uintptr(unsafe.Pointer(&t)))
	if err != nil {
		return fmt.Errorf("setTermios: %s", err)
	}
	return nil
}

func isTerminal(fd int) bool {
	_, err := getTermios(fd)
	return err == nil
}

func makeRaw(fd int) (*State, error) {
	termios, err := getTermios(fd)
	if err != nil {
		return nil, err
	}

	oldState := State{state{termios: termios}}

	// This attempts to replicate the behaviour documented for cfmakeraw in
	// the termios(3) manpage.
	termios.Iflag &^= syscall.IGNBRK | syscall.BRKINT | syscall.PARMRK | syscall.ISTRIP |
		syscall.INLCR | syscall.IGNCR | syscall.ICRNL | syscall.IXON
	termios.Lflag &^= syscall.ECHO | syscall.ECHONL | syscall.ICANON | syscall.ISIG | syscall.IEXTEN
	termios.Oflag &^= syscall.OPOST
	termios.Cflag &^= syscall.CSIZE | syscall.PARENB
	termios.Cflag |= syscall.CS8
	termios.Cc[syscall.VMIN] = 1
	termios.Cc[syscall.VTIME] = 0
	if err := setTermios(fd, termios); err != nil {
		return nil, fmt.Errorf("set: %s", err)
	}

	return &oldState, nil
}

func getState(fd int) (*State, error) {
	termios, err := getTermios(fd)
	if err != nil {
		return nil, err
	}

	return &State{state{termios: termios}}, nil
}

func restore(fd int, state *State) error {
	return setTermios(fd, state.termios)
}

// ws, err := unix.IoctlGetWinsize(fd, unix.TIOCGWINSZ)
// func IoctlGetWinsize(fd int, req int) (*Winsize, error) {

func getSize(fd int) (width, height int, err error) {
	var size struct {
		Height, Width  uint16
		Xpixel, Ypixel uint16
	}
	err = ioctlPtr(fd, syscall.TIOCGWINSZ, unsafe.Pointer(&size))
	return int(size.Width), int(size.Height), err
}

// passwordReader is an io.Reader that reads from a specific file descriptor.
type passwordReader int

func (r passwordReader) Read(buf []byte) (int, error) {
	return syscall.Read(int(r), buf)
}

func readPassword(fd int) ([]byte, error) {
	termios, err := getTermios(fd)
	if err != nil {
		return nil, err
	}

	newState := termios
	newState.Lflag &^= syscall.ECHO
	newState.Lflag |= syscall.ICANON | syscall.ISIG
	newState.Iflag |= syscall.ICRNL
	err = setTermios(fd, newState)
	if err != nil {
		return nil, err
	}

	defer setTermios(fd, termios)

	return readPasswordLine(passwordReader(fd))
}
