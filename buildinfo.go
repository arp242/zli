package zli

import (
	"fmt"
	"os"
	"path/filepath"
	"runtime/debug"
	"strings"
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
// The format is:
//
//	prog 336b4c73 2024-06-07; go1.22.4 linux/amd64; race=false; cgo=false
//
// Where prog is os.Args[0], followed by the commit and date of the commit. You
// can print a tagged version by setting zgo.at/zli.version at build time:
//
//	go build -ldflags "-X zgo.at/zli.version=v1.2.3"
//
// After which it will print:
//
//	elles v1.2.3 336b4c73 2024-06-07; go1.22.4 linux/amd64; race=false; cgo=true
//
// In addition, zgo.at/zli.progname can be set to override os.Args[0]:
//
//	go build -ldflags '-X "zgo.at/zli.version=VERSION" -X "zgo.at/zli.progname=PROG"'
//
// If verbose is true it also prints detailed build information (similar to "go
// version -m bin")
func PrintVersion(verbose bool) {
	if progname == "" && len(os.Args) > 0 {
		progname = filepath.Base(os.Args[0])
	}

	info, ok := debug.ReadBuildInfo()
	if !ok {
		fmt.Fprintln(Stdout, "failed reading detailed build info")
	}

	var (
		race, cgo, mod            bool
		goos, goarch, commit, vcs string
		date                      time.Time
	)
	for _, s := range info.Settings {
		switch s.Key {
		case "-race":
			race = s.Value == "true"
		case "CGO_ENABLED":
			cgo = s.Value == "1"
		case "GOARCH":
			goarch = s.Value
		case "GOOS":
			goos = s.Value
		case "vcs.revision":
			commit = s.Value
		case "vcs.time":
			date, _ = time.Parse(time.RFC3339, s.Value)
		case "vcs.modified":
			mod = s.Value == "true"
		case "vcs":
			vcs = s.Value
		}
	}
	if vcs == "git" && len(commit) > 8 {
		commit = commit[:8]
	}

	v := make([]string, 0, 4)
	if version != "" && version != "dev" {
		v = append(v, version)
	}
	if commit != "" {
		v = append(v, commit)
	}
	if !date.IsZero() {
		v = append(v, date.Format("2006-01-02"))
	}
	if mod {
		v = append(v, "(modified)")
	}

	fmt.Fprintf(Stdout, "%s %s; %s %s/%s; race=%t; cgo=%t\n",
		progname, strings.Join(v, " "), info.GoVersion, goos, goarch, race, cgo)

	if verbose && info != nil {
		fmt.Fprint(Stdout, "\n", info)
	}
}
