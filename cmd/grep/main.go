// Command grep is a simple grep implementation for demo purposes.
package main

import (
	"bufio"
	"bytes"
	"fmt"
	"io"
	"os"
	"regexp"
	"strconv"

	"zgo.at/zli"
)

// There is no automatic generation of a usage message; I find that much of the
// time you get a much higher quality by writing one manually.
//
// That being said, the zli.Usage() function can apply some generic
// substitutions; here it applies some formatting to the headers ("Usage:",
// "Description:", etc.) and the flags ("-p", "-only-matching", etc.) which
// makes it look a bit nicer and somewhat reminiscent of manpages.
var usage = zli.Usage(zli.UsageTrim|zli.UsageHeaders|zli.UsageFlags, `
Usage:
    grep [options..] pattern [file..]

Description:
    grep searches for a pattern in each file.

Options:
	pattern
		A regular expression.

	file [file..]
		Zero or more files; if none are given read from stdin.

    -o, -only-matching
        Print only the matching part, instead the entire line.

    -q, -quiet, -silent
        Don't show any output, exit with 0 on the first match found.

    -color=when, --color=when
        When to display colors: auto (default), never, or always.

	-p, -no-pager
		Don't pipe the output to $PAGER.

Exit code:
    0 if a pattern is found, 1 if nothing is found, 2 if there was an error.
`)

// Define some colors we'll use later on.
const (
	colorMatch  = zli.Red
	colorLineNr = zli.Magenta
	colorPath   = zli.Bold | zli.Underline
)

func main() {
	// Parse the flags.
	f := zli.NewFlags(os.Args)
	var (
		only    = f.Bool(false, "o", "only-matching")
		silent  = f.Bool(false, "q", "quiet", "silent")
		noPager = f.Bool(false, "p", "no-pager")
		color   = f.String("auto", "color", "colour")
	)
	err := f.Parse()
	if err != nil {
		fmt.Println(usage)
		zli.Fatal(err)
	}

	// The value needs to be retrieved through a getting function; this avoids
	// having to deal with pointers and the like. You can use color.Set() to see
	// if the flag was present on the commandline at all.
	switch color.String() {
	case "always":
		zli.WantColor = true
	case "never":
		zli.WantColor = false
	}

	// Shift() removes the first positional argument, or returns an empty string
	// if there isn't any. In this case, the first positional argument is the
	// regexp we want to match with.
	patt := f.Shift()
	if patt == "" {
		fmt.Println(usage)
		zli.Fatal("need a pattern")
	}
	re, err := regexp.Compile(patt)
	zli.F(err)

	// No File arguments? Read from stdin, InputOrFile() will take care of that.
	if len(f.Args) == 0 {
		f.Args = []string{"-"}
	}

	// Collect output in a memory buffer so we can send it to the pager.
	var out io.ReadWriter
	if noPager.Set() {
		out = os.Stdout
	} else {
		out = new(bytes.Buffer)
	}

	exit := 1 // Nothing selected is exit 1
	for _, path := range f.Args {
		// Read either the file or stdin (if "" or "-").
		fp, err := zli.InputOrFile(path, false)
		zli.F(err)
		defer fp.Close()

		var shownPath = false
		var (
			scan   = bufio.NewScanner(fp)
			lineNr = int64(0)
		)
		for scan.Scan() {
			l := scan.Text()
			lineNr++

			match := re.FindAllStringIndex(l, -1)
			if len(match) == 0 {
				continue
			}

			// We can exit in -quiet mode on the first match.
			//
			// Can also use Bool(), but it doesn't really matter, and Set()
			// reads a bit nicer IMHO :-)
			if silent.Set() {
				os.Exit(0)
			}
			exit = 0

			// Apply the color highlighting for the matches, loop over the
			// matches in reverse order so the inserted color codes for the
			// first match won't affect the string indexing for the second
			// match.
			for i := len(match) - 1; i >= 0; i-- {
				m := match[i]
				if only.Set() {
					l = zli.Colorf(l[m[0]:m[1]], colorMatch)
				} else {
					l = l[:m[0]] + zli.Colorf(l[m[0]:m[1]], colorMatch) + l[m[1]:]
				}
			}

			if !zli.IsTerminal(os.Stdout.Fd()) {
				// Not a terminal: print file path for every line.
				fmt.Fprint(out, path, ":")
			} else if !shownPath {
				// Print file path as a header once on interactive terminals.
				fmt.Fprintln(out, zli.Colorf(path, colorPath))
				shownPath = true
			}
			fmt.Fprintln(out, zli.Colorf(strconv.FormatInt(lineNr, 10), colorLineNr)+":"+l)
		}
	}

	// Show in pager.
	if !noPager.Set() {
		zli.Pager(out)
	}

	os.Exit(exit)
}
