package zli

import (
	"fmt"
	"os"
	"path/filepath"
	"runtime/debug"
	"strings"
	"time"
)

var version, progname string

// Verion gets a version string, as a tag name ("v2.1.0") or a hash and date
// ("53097a4/2025-12-26"). Both may be followed by "(modified)" if built from a
// modified source directory.
//
// The version global always takes precedence if set; this can be set at build
// time with:
//
//	go build -ldflags "-X zgo.at/zli.version=v1.2.3"
//
// This can be any string and doesn't need to be SemVer; it's not parsed.
func Version() string {
	if version != "" {
		return version
	}

	info, ok := debug.ReadBuildInfo()
	if !ok {
		return "failed reading build info"
	}

	var (
		tag         = info.Main.Version
		vcs, commit string
		mod         bool
		date        time.Time
	)
	for _, s := range info.Settings {
		switch s.Key {
		case "vcs":
			vcs = s.Value
		case "vcs.revision":
			commit = s.Value
		case "vcs.modified":
			mod = s.Value == "true"
		case "vcs.time":
			date, _ = time.Parse(time.RFC3339, s.Value)
		}
	}
	if vcs == "git" && len(commit) > 8 {
		commit = commit[:8]
	}

	// go install . w tag   → tag="v2.1.0";                               commit="fefa60ba"; vcs="git"
	// go install [..]@tag  → tag="v2.0.0";                               commit="";         vcs=""
	// go install . w/o tag → tag="v2.0.1-0.20251226200708-fefa60ba4a0a"; commit="fefa60ba"; vcs="git"
	// go install [..]@hash → tag="v2.0.1-0.20251218155453-229ce2e7bb56"; commit="";         vcs=""

	if strings.HasSuffix(tag, "+dirty") {
		tag, mod = tag[:len(tag)-6], true
	}
	split := strings.Split(tag, "-")
	if len(split) == 3 {
		_, split[1], _ = strings.Cut(split[1], ".")
		if commit == "" {
			commit = split[2]
		}
		if date.IsZero() {
			date, _ = time.Parse("20060102150405", split[1])
		}
		tag = fmt.Sprintf("%s/%s", commit, date.Format("2006-01-02"))
	}
	if mod {
		tag += " (modified)"
	}
	return tag
}

// PrintVersion prints this program's version.
//
// The format is one of:
//
//	prog v2.0.0; go1.25.5 linux/amd64; race=false; cgo=true
//	prog 229ce2e7bb56/2025-12-18; go1.25.5 linux/amd64; race=false; cgo=true
//
// The version can be overridden by setting the version global at build time.
//
// The program name is normally taken from os.Args[0], but can be overridden at
// build time by setting the progname global. The version string can similarly
// be overridden with the version global:
//
//	go build -ldflags '-X "zgo.at/zli.version=VERSION" -X "zgo.at/zli.progname=PROG"'
//
// If verbose is true it also prints detailed build information (similar to "go
// version -m bin")
func PrintVersion(verbose bool) {
	prog := progname
	if prog == "" && len(os.Args) > 0 {
		prog = filepath.Base(os.Args[0])
	}

	info, ok := debug.ReadBuildInfo()
	if !ok {
		fmt.Fprintln(Stdout, "failed reading detailed build info")
	}

	var (
		race, cgo    bool
		goos, goarch string
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
		}
	}

	fmt.Fprintf(Stdout, "%s %s; %s %s/%s; race=%t; cgo=%t\n",
		prog, Version(), info.GoVersion, goos, goarch, race, cgo)
	if verbose && info != nil {
		fmt.Fprint(Stdout, "\n", info)
	}
}
