//go:build unix

package zli

import (
	"os"
	"syscall"
)

var signals = []os.Signal{syscall.SIGHUP, syscall.SIGTERM, os.Interrupt}
