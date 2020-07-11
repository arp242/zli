zli is a Go library for writing CLI programs. It includes flag parsing, color
escape codes, and various helpful utility functions.

Import a `zgo.at/zli`; API docs: https://pkg.go.dev/zgo.at/zli

There's a little example at [cmd/grep/main.go](cmd/grep/main.go), which should
give a decent overview of how actual programs look like.

Readme index:
[Utility functions](#utility-functions) ·
[Flag parsing](#flag-parsing) ·
[Colors](#colors)


### Utility functions

```go
zli.Error("oh noes: %s", "u brok it")  // Prints to stderr: "progname: oh noes: u brok it"
zli.Fatal(errors.New("yikes"))         // Like Error() but exits: "progname: yikes"
zli.F(errors.New("yikes"))             // Shorter version which checks if err is nil first.

fp, err := zli.FileOrInput("/a-file")  // Read data from a file...
fp, err := zli.FileOrInput("-")        // ...or read from stdin; can also use "" for stdin
defer fp.Close()                       // No-op close on stdin.

fp, _ := os.Open("/file")              // Display contents of a reader in $PAGER.
zli.Pager(fp)

w, h, err := zli.TerminalSize(os.Stdout.Fd())  // Get terminal size.
interactive := zli.IsTerminal(os.Stdout.Fd())  // Check if stdout is a terminal.
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

### Colors

You can add colors and some other text attributes to a string with
`zli.Colorf()`, which returns a modified string with the terminal escape codes.

It won't do anything if `zli.WantColor` is `false`; this is set automatically
disabled if the output isn't a terminal or `NO_COLOR` is set, but you can
override it if the user set `--color=force` or something.

`zli.Colorln()` is a convenience wrapper for `fmt.Println(zli.Colorf(..))`.

```go
zli.Colorln("You're looking rather red", zli.Red)     // Apply a color.
zli.Colorln("A bold move", zli.Bold)                  // Or an attribute.
zli.Colorln("Tomato", zli.Red.Bg())                   // Transform to background color.

zli.Colorln("Wow, such beautiful text",               // Can be combined.
    zli.Bold|zli.Underline|zli.Red|zli.Green.Bg())

zli.Colorln("Contrast ratios is for suckers",         // 256 color.
    zli.Color256(56)|zli.Color256(99).Bg())

zli.Colorln("REAL men use TRUE color!",               // True color.
    zli.ColorHex("#fff"), zli.ColorHex("#00f").Bg())

// Any combination can be used; the way this works is that the various
// attributes are keps in different bit flags in Color (uint64), so it's easy to
// use a single constant to represent it.
zli.Colorln("Combine as you want",
    zli.Bold|zli.Underline|zli.Red|zli.ColorHex("#0f0").Bg())
```

See [cmd/colortest/main.go](cmd/colortest/main.go) for a little program to
display and test colors.
