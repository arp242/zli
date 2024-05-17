This is a fork of golang.org/x/term, but without a dependency on
golang.org/x/sys, using the stdlib syscall package instead.

This is useful in some cases because x/sys is rather large (~11M) and often you
just want a quick check to get the terminal width, or something like that. If
you're already using x/sys that's fine, but often you're not, and it's the only
dependency.

This can also be easily copy/pasted in internal, for dependency-free term
operations.

Main downside is that syscall isn't really maintained and new platforms may be
slower to get support. None of this is cutting edge stuff, so should be fine.
