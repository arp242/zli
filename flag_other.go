//go:build !unix

package zli

import "os"

var signals = []os.Signal{os.Interrupt}
