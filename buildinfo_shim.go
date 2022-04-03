//go:build !go1.18
// +build !go1.18

package zli

import "fmt"

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
//   go build -ldflags '-X "zgo.at/zli.version=VERSION" -X "zgo.at/zli.progname=PROG"'
func PrintVersion(program, version string, verbose bool) {
	fmt.Fprintln(Stdout, program, version)
	if verbose {
		fmt.fprintln(Stdout, "expanded build info only available if built with Go 1.18 or newer")
	}
}
