//go:build unix

package zli

import (
	"os"
	"os/signal"
	"syscall"
)

var exitSignals = []os.Signal{syscall.SIGHUP, syscall.SIGTERM, os.Interrupt}

// TerminalSizeChange is run if the terminal window size is changed.
func TerminalSizeChange(f func()) {
	winch := make(chan os.Signal, 1)
	signal.Notify(winch, syscall.SIGWINCH)
	go func() {
		for {
			<-winch
			f()
		}
	}()
}
