Some common functions for writing CLI programs.

[GoDoc](https://pkg.go.dev/zgo.at/zli)

```go
// Quick exit with nice message to stderr.
zli.Fatal("oh noes: %q", "data")       // progname: oh noes: "data"
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
