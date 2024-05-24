//go:build !unix

package zli

import "os"

var exitSignals = []os.Signal{os.Interrupt}

// TerminalSizeChange is run if the terminal window size is changed.
func TerminalSizeChange() <-chan struct{} {
	return make(chan struct{})
	// Do nothing.
	// TODO: it looks like this may be possible on Windows:
	// https://stackoverflow.com/questions/10856926/sigwinch-equivalent-on-windows
}
