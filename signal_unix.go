//go:build unix

package zli

import (
	"os"
	"os/signal"
	"syscall"
)

var exitSignals = []os.Signal{syscall.SIGHUP, syscall.SIGTERM, os.Interrupt}

// TerminalSizeChange sends on the channel if the terminal window is resized.
func TerminalSizeChange() <-chan struct{} {
	winch := make(chan os.Signal, 1)
	ch := make(chan struct{})
	signal.Notify(winch, syscall.SIGWINCH)
	go func() {
		for {
			<-winch
			ch <- struct{}{}
		}
	}()
	return ch
}
