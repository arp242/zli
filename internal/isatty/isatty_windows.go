// +build windows
// +build !appengine

package isatty

import (
	"syscall"
	"unsafe"
)

// IsTerminal return true if the file descriptor is terminal.
func IsTerminal(fd uintptr) bool {
	var (
		kernel32           = syscall.NewLazyDLL("kernel32.dll")
		procGetConsoleMode = kernel32.NewProc("GetConsoleMode")
	)
	var st uint32
	r, _, e := syscall.Syscall(procGetConsoleMode.Addr(), 2, fd, uintptr(unsafe.Pointer(&st)), 0)
	return r != 0 && e == 0
}
