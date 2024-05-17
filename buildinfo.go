package zli

import (
	"fmt"
	"os"
	"runtime/debug"
	"time"
)

var (
	progname = ""
	version  = "dev"
)

// GetVersion gets this program's version.
func GetVersion() (tag string, commit string, date time.Time) {
	b, ok := debug.ReadBuildInfo()
	if !ok {
		return version, "failed reading detailed build info", time.Time{}
	}

	var vcs string
	for _, s := range b.Settings {
		switch s.Key {
		case "vcs.revision":
			commit = s.Value
		case "vcs.time":
			date, _ = time.Parse(time.RFC3339, s.Value)
		case "vcs":
			vcs = s.Value
		}
	}
	if vcs == "git" && len(commit) > 8 {
		commit = commit[:8]
	}
	return version, commit, date
}

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
		if b, ok := debug.ReadBuildInfo(); !ok {
			fmt.Fprintln(Stdout, "failed reading detailed build info")
		} else {
			fmt.Fprint(Stdout, "\n", b)
		}
	}
}
