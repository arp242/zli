Some common functions for writing CLI programs.

[GoDoc](https://pkg.go.dev/zgo.at/zli)

### Utility functions

```go
// Show error on stderr.
zli.Error("oh noes: %q", "data")       // progname: oh noes: "data"

// Quick exit with nice message to stderr, just Error() followed by exit().
zli.Fatal(errors.New("yikes"))         // progname: yikes

// Shorter version which won't do anything of the error is nil.
zli.F(errors.New("yikes"))             // progname: yikes

// Read from stdin or a file
fp, err := zli.FileOrInput("")         // from stdin; can also use "-"
defer fp.Close()                       // No-op close on stdin.

fp, err := zli.FileOrInput("/a-file")  // Read from file.

// Display in $PAGER.
fp, _ := os.Open("/file")
zli.Pager(fp)

// Get terminal size, check if stdout is a terminal.
width, height, err := zli.TerminalSize(os.Stdout.Fd())
interactive := zli.IsTerminal(os.Stdout.Fd())
```

### Flag parsing

zli comes with a simple no-nonsense flag parser which, IMHO, gives a better
experience than Go's `flag` package. See [flag.markdown](/flag.markdown) for
some rationale on "why this and not stdlib flags?"

```go
// Create new flags from os.Args.
f := zli.NewFlags([]string{"example", "-vv", "-f=csv", "-a", "xx", "yy"})

// Add a string, bool, and "counter" flag.
var (
    verbose = f.IntCounter(0, "v", "verbose")
    all     = f.Bool(false, "a", "all")
    format  = f.String("", "f", "format")
)

// Shift the first argument (i.e. os.Args[1], if any, empty string if there
// isn't). Useful to get the "subcommand" name. This works before and after
// Parse().
switch f.Shift() {
case "help":
    // Run help
case "install":
    // Run install
case "":
    // Error: need a command (or just print the usage)
default:
    // Error: Unknown command
}

// Parse the shebang!
err := f.Parse()
if err != nil {
    // Print error, usage.
}

// You can check if the flag was present on the CLI with Set(). This way you can
// distinish between "was an empty value passed" // (-format '') and "this flag
// wasn't on the CLI".
if format.Set() {
    fmt.Println("Format was set to", format.String())
}

// The IntCounter adds 1 for every time the -v flag is on the CLI.
if verbose.Int() > 1 {
    // ...Print very verbose info.
} else if verbose.Int() > 0 {
    // ...Print less verbose info.
}

// Just a bool!
fmt.Println("All:", all.Bool())

// f.Args is set to everything that's not a flag or argument.
fmt.Println("Remaining:", f.Args)
```
