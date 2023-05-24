//go:build !go1.18
// +build !go1.18

package zli

import (
	"fmt"
	"os"
)

var (
	progname = ""
	version  = "dev"
)

// PrintVersion prints this program's version.
//
// If verbose is true it also prints detailed build information. This only works
// for Go 1.18 or newer.
//
// This assumes that zgo.at/zli.version and zgo.at/zli.progname were set at
// build time:
//
//	go build -ldflags '-X "zgo.at/zli.version=VERSION" -X "zgo.at/zli.progname=PROG"'
func PrintVersion(verbose bool) {
	if progname == "" && len(os.Args) > 0 {
		progname = os.Args[0]
	}
	fmt.Fprintln(Stdout, progname, version)

	if verbose {
		fmt.Fprintln(Stdout, "expanded build info only available if built with Go 1.18 or newer")
	}
}
