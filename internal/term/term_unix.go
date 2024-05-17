// Copyright 2019 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

//go:build aix || darwin || dragonfly || freebsd || linux || netbsd || openbsd || solaris || zos

package term

import (
	"fmt"
	"syscall"
	"unsafe"
)

type state struct {
	termios termios
}

type termios struct {
	Iflag, Oflag, Cflag, Lflag uint32
	Line                       uint8
	Cc                         [19]uint8
	Ispeed, Ospeed             uint32
}

func getTermios(fd int) (termios, error) {
	var t termios
	_, _, errno := syscall.Syscall(syscall.SYS_IOCTL, uintptr(fd), ioctlReadTermios, uintptr(unsafe.Pointer(&t)))
	if errno > 0 {
		return t, fmt.Errorf("getTermios: %s", errno)
	}
	return t, nil
}

func setTermios(fd int, t termios) error {
	_, _, errno := syscall.Syscall(syscall.SYS_IOCTL, uintptr(fd), ioctlWriteTermios, uintptr(unsafe.Pointer(&t)))
	if errno > 0 {
		return fmt.Errorf("setTermios: %s", errno)
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

func getSize(fd int) (width, height int, err error) {
	var size struct {
		Height, Width  uint16
		Xpixel, Ypixel uint16
	}
	_, _, errno := syscall.Syscall(syscall.SYS_IOCTL, uintptr(fd), syscall.TIOCGWINSZ, uintptr(unsafe.Pointer(&size)))
	if errno > 0 || size.Width <= 0 || size.Height <= 0 {
		return 0, 0, fmt.Errorf("%v", errno)
	}
	return int(size.Width), int(size.Height), nil
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
