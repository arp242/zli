zli is a Go library for writing CLI programs. It includes flag parsing, color
escape codes, various helpful utility functions, and makes testing fairly easy.
There's a little example at [cmd/grep](cmd/grep), which should give a decent
overview of how actual programs look like.

Import as `zgo.at/zli`; API docs: https://pkg.go.dev/zgo.at/zli

**Readme index**:
[Utility functions](#utility-functions) ·
[Flag parsing](#flag-parsing) ·
[Colours](#colours) ·
[Testing](#testing)


### Utility functions

`zli.Errorf()` and `zli.Fatalf()` work like `fmt.Printf()`, except that they
print to stderr, prepend the program name, and always append a newline:

```go
zli.Errorf("oh noes: %s", "u brok it")   // "progname: oh noes: u brok it"
zli.Fatalf("I swear it was %s", "Dave")  // "progname: I swear it was Dave" and exit 1
```

`zli.F()` is a small wrapper/shortcut around `zli.Fatalf()` which accepts an
error and checks if it's `nil` first:

```go
err := f()
zli.F(err)
```

For many programs it's useful to be able to read from stdin or from a file,
depending on what arguments the user gave. With `zli.InputOrFile()` this is
pretty easy:

```go
fp, err := zli.InputOrFile("/a-file", false)  // Open a file.

fp, err := zli.InputOrFile("-", false)        // Read from stdin; can also use "" for stdin.
defer fp.Close()                              // No-op close on stdin.
```

The second argument controls if a `reading from stdin...` message should be
printed to stderr, which is a bit better UX IMHO (how often have you typed `grep
foo` and waited, only to realize it's waiting for stdin?) See [Better UX when
reading from stdin][stdin].

With `zli.InputOrArgs()` you can read arguments from stdin if it's an empty
list:

```go
args := zli.InputOrArgs(os.Args[1:], "\n", false)     // Split arguments on newline.
args := zli.InputOrArgs(os.Args[1:], "\n\t ", false)  // Or on spaces and tabs too.
```

[stdin]: https://www.arp242.net/read-stdin.html

`zli.Pager()` pipes the contents of a reader `$PAGER`. It will copy the contents
to stdout if `$PAGER` isn't set or on other errors:

```go
fp, _ := os.Open("/file")        // Display file in $PAGER.
zli.Pager(fp)
```

If you want to page output your program generates you can use
`zli.PagerStdout()` to swap `zli.Stdout` to a buffer:

```go
defer zli.PagerStdout()()               // Double ()()!
fmt.Fprintln(zli.Stdout, "page me!")    // Displayed in the $PAGER.
```

This does require that your program writes to `zli.Stdout` instead of
`os.Stdout`, which is probably a good idea for testing anyway. See the
[Testing](#testing) section.

You need to be a bit careful when calling `Exit()` explicitly, since that will
exit immediately without running any defered functions. You have to either use a
wrapper or call the returned function explicitly:

```go
func main() { zli.Exit(run()) }

func run() int {
    defer zli.PagerStdout()()
    fmt.Fprintln(zli.Stdout, "XXX")
    return 1
}
```

```go
func main() {
    runPager := zli.PagerStdout()
    fmt.Fprintln(zli.Stdout, "XXX")

    runPager()
    zli.Exit(1)
}
```

zli helpfully includes the [go-isatty][isatty] and `GetSize()` from
[x/crypto/ssh/terminal][ssh] as they're so commonly used:

```go
interactive := zli.IsTerminal(os.Stdout.Fd())  // Check if stdout is a terminal.
w, h, err := zli.TerminalSize(os.Stdout.Fd())  // Get terminal size.
```

[isatty]: https://github.com/mattn/go-isatty/
[ssh]: https://godoc.org/golang.org/x/crypto/ssh/terminal#GetSize


### Flag parsing

zli comes with a flag parser which, IMHO, gives a better experience than Go's
`flag` package. See [flag.markdown](/flag.markdown) for some rationale on "why
this and not stdlib flags?"

```go
// Create new flags; normally you'd pass in os.Args here.
f := zli.NewFlags([]string{"example", "-vv", "-f=csv", "-a", "xx", "yy"})

// The first argument is the default and everything after that is the flag name
// with aliases.
var (
    verbose = f.IntCounter(0, "v", "verbose")   // Count the number of -v flags.
    exclude = f.StringList(nil, "e", "exclude") // Can appear more than once.
    all     = f.Bool(false, "a", "all")         // Regular bool.
    format  = f.String("", "f", "format")       // Regular string.
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

// Allow a flag to appear more than once.
fmt.Println("%s exclude patterns: %v", len(all.Strings()), all.Strings())

// f.Args is set to everything that's not a flag or argument.
fmt.Println("Remaining:", f.Args)
```

The flag format is as follows:

- Flags can have a single `-` or two `--`, they're treated identical.

- Arguments are after a space or `=`: `-f v` or `-f=v`.

- Booleans can be grouped; `-ab` is the same as `-a -b`; this only works with a
  single `-` (`--ab` would be an error).

- Positional arguments may appear anywhere; these are all identical:
  `-a -b arg`, `arg -a -b`, `-a arg -b`.

---

There is no automatic generation of a usage message; I find that much of the
time you get a much higher quality by writing one manually. It does provide
`zli.Usage()` you can apply some generic substitutions giving a format somewhat
reminiscent of manpages:

    UsageTrim      Trim leading/trailing whitespace, and ensure it ends with \n
    UsageHeaders   Format headers in the form "^Name:" as bold and underline.
    UsageFlags     Format flags (-v, --flag, --flag=foo) as underlined.

See the grep example.

### Colors

You can add colors and some other text attributes to a string with
`zli.Colorf()`, which returns a modified string with the terminal escape codes,
ending with reset.

It won't do anything if `zli.WantColor` is `false`; this is disabled by default
if the output isn't a terminal or `NO_COLOR` is set, but you can override it if
the user sets `--color=force` or something.

`zli.Colorln()` is a convenience wrapper for `fmt.Println(zli.Colorf(..))`.

There are constants for the basic terminal attributes and 16-color palette which
may be combined freely by adding them together:

```go
zli.Colorln("You're looking rather red", zli.Red)     // Apply a color.
zli.Colorln("A bold move", zli.Bold)                  // Or an attribute.
zli.Colorln("A bold move", zli.Red | zli.Bold)        // Or both.
```

To set a background color transform the color with the `Bg()` method:

```go
zli.Colorln("Tomato", zli.Red.Bg())                   // Transform to background color.
zli.Colorln("Wow, such beautiful text",               // Can be combined.
    zli.Bold | zli.Red | zli.Green.Bg())
```

There are no pre-defined constants for the 256-color palette or true colors, you
need to use `Color256()` and `ColorHex()` to create them; you can use the `Bg()`
to transform them to a background color as well:

```go
zli.Colorln("Contrast ratios is for suckers",         // 256 color.
    zli.Color256(56) | zli.Color256(99).Bg())

zli.Colorln("REAL men use TRUE color!",               // True color.
    zli.ColorHex("#fff") | zli.ColorHex("#00f").Bg())
```

See [cmd/colortest/main.go](cmd/colortest/main.go) for a little program to
display and test colors.

---

For some more advanced cases you can use `Color.String()` directly, but this
won't look at `zli.WantColor` and you'll need to manually apply the reset code:

```go
fmt.Println(zli.Red|zli.Bold, "red!")                 // Print escape codes.
fmt.Println("and bold!", zli.Reset)

fmt.Printf("%sc%so%sl%so%sr%s\n", zli.Red, zli.Magenta, zli.Cyan, zli.Blue, zli.Yellow, zli.Reset)
```

Because the color is stored in an `uint64` you can assign them to a constant:

```go
const colorMatch = zli.Bold | zli.Red
```

This won't work if you use `Color256()` or `ColorHex()`; although you can get
around this by constructing it all yourself:

```go
// zli.Color256(99)
const color = zli.Bold | (zli.Color(99) << zli.ColorOffsetFg) | zli.ColorMode256

// zli.ColorHex("#ff7711").Bg(); can also use 1144831 directly instead of the
// bit shifts.
const color2 = zli.Bold | zli.Red | zli.ColorModeTrueBg |
               (zli.Color(0xff|0x77<<8|0x11<<16) << zli.ColorOffsetBg)
```

This creates a color stored as an int, shifts it to the correct location, and
sets the flag to signal how to interpret it.

Do you really want to do this just to create a `const` instead of a `var`?
Probably not 😅


### Testing

zli uses to `zli.Stdin`, `zli.Stdout`, `zli.Stderr`, and `zli.Exit` instead of
the `os.*` variants for everything. You can swap this out with test variants
with the `zli.Test()` function.

You can use these in your own program as well, if you want to test the output of
a program.

```go
func TestX(t *testing.T) {
    exit, in, out, reset := Test()
    defer reset() // Reset everything back to the os.* functions.

    // Write something to stderr (a bytes.Buffer) and read the output.
    Error("oh noes!")
    fmt.Println(out.String()) // zli.test: oh noes!

    // Read from stdin.
    in.WriteString("Hello")
    fp, _ := InputOrFile("-", true)
    got, _ := ioutil.ReadAll(fp)
    fmt.Println(string(got)) // Hello

    out.Reset()

    et := func() {
        fmt.Fprintln(Stdout, "one")
        Exit(1)
        fmt.Fprintln(Stdout, "two")
    }

    // exit panics to ensure the regular control flow of the program is aborted;
    // to capture this run the function to be tested in a closure with
    // exit.Recover(), which will recover() from the panic and set the exit
    // code.
    func() {
        defer exit.Recover()
        et()
    }()
    // Helper to check the statis code, so you don't have to derefrence and cast
    // the value to int.
    exit.Want(t, 1)

    fmt.Println("Exit %d: %s\n", *exit, out.String()) // Exit 1: one
```

You don't need to use the `zli.Test()` function if you won't want to, you can
just swap out stuff yourself as well:

```go
buf := new(bytes.Buffer)
zli.Stderr = buf
defer func() { Stderr = os.Stderr }()

Error("oh noes!")
out := buf.String()
fmt.Printf("buffer has: %q\n", out) // buffer has: "zli.test: oh noes!\n"
```

`zli.IsTerminal()` and `zli.TerminalSize()` are variables, and can be swapped
out as well:

```go
save := zli.IsTerminal
zli.IsTerminal = func(uintptr) bool { return true }
defer func() { IsTerminal = save }()
```


#### Exit

A few notes on replacing `zli.Exit()` in tests: the difficulty with this is that
`os.Exit()` will terminate the entire program, including the test, which is
rarely what you want and difficult to test. You can replace `zli.Exit` with
something like (`zli.TestExit()` takes care of all of this):

```go
var code int
zli.Exit = func(c int) { code = c }
mayExit()
fmt.Println("exit code", code)
```

This works well enough for simple cases, but there's a big caveat with this; for
example consider:

```go
func mayExit() {
    err := f()
    if err != nil {
        zli.Error(err)
        zli.Exit(4)
    }

    fmt.Println("X")
}
```

With the above the program will continue after `zli.Exit()`; which is a
different program flow from normal execution. A simple way to fix it so to
modify the function to explicitly call `return`:

```go
func mayExit() {
    err := f()
    if err != nil {
        zli.Error(err)
        zli.Exit(4)
        return
    }

    fmt.Println("X")
}
```

This still isn't *quite* the same, as callers of `mayExit()` in your program
will still continue happily. It's also rather ugly and clunky.

To solve this you can replace `zli.Exit` with a function that panics and then
recover that:

```go
func TestFoo(t *testing.T) {
    var code int
    zli.Exit = func(c int) {
        code = c
        panic("zli.Exit")
    }

    func() {
        defer func() {
            r := recover()
            if r == nil {
                return
            }
        }()

        mayExit()
    }()

    fmt.Println("Exited with", code)
}
```

This will abort the program flow similar to `os.Exit()`, and the call to
`mayExit` is wrapped in a function the test function itself will continue after
the recover.
