zli is a Go library for writing CLI programs. It includes flag parsing, colour
escape codes, and various helpful utility functions.

Import a `zgo.at/zli`; API docs: https://pkg.go.dev/zgo.at/zli


### Utility functions

```go
// Show error on stderr.
zli.Error("oh noes: %q", "data")       // progname: oh noes: "data"

// Quick exit with nice message to stderr (Error() followed by exit).
zli.Fatal(errors.New("yikes"))         // progname: yikes

// Shorter version which won't do anything of the error is nil.
zli.F(errors.New("yikes"))             // progname: yikes

// Read from stdin or a file
fp, err := zli.FileOrInput("/a-file")  // Read from file.

fp, err := zli.FileOrInput("-")        // from stdin; can also use ""
defer fp.Close()                       // No-op close on stdin.

// Display contents of a reader in $PAGER.
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
// Create new flags; normally you'd pass in os.Args here.
f := zli.NewFlags([]string{"example", "-vv", "-f=csv", "-a", "xx", "yy"})

// Add a string, bool, and "counter" flag.
var (
    verbose = f.IntCounter(0, "v", "verbose")
    all     = f.Bool(false, "a", "all")
    format  = f.String("", "f", "format")
)

// Shift the first argument (i.e. os.Args[1]). Useful to get the "subcommand"
// name. This works before and after Parse().
switch f.Shift() {
case "help":
    // Run help
case "install":
    // Run install
case "": // os.Args wasn't long enough.
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
// distinguish between "was an empty value passed" (-format '') and "this flag
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

### Colours


You can add colours and some other text attributes to a string with
`zli.Color()`, which returns a modified string with the terminal escape codes.

It won't do anything if `zli.WantColor` is `false`; this is set automatically
disabled if the output isn't a terminal or `NO_COLOR` is set, but you can
override it if the user set `--color=force` or something.

`zli.Colorln()` is a convenience wrapped for `fmt.Println(zli.Color(...))`.

```go
zli.Colorln("You're looking rather red", zli.Red)     // Apply a colour.
zli.Colorln("A bold move", zli.Bold)                  // Or an attribute.
zli.Colorln("Tomato", zli.Red.Bg())                   // Transform to background colour.

zli.Colorln("Wow, such beautiful text",               // Can be combined.
    zli.Bold, zli.Underline, zli.Red, zli.Green.Bg())

zli.Colorln("Contrast ratios is for suckers",         // 256 colour
    zli.NewColor().From256(56), zli.NewColor().From256(99).Bg())

zli.Colorln("REAL men use TRUE color!",               // True colour
    zli.NewColor().FromHex("#fff"), zli.NewColor().FromHex("#00f").Bg())
```
