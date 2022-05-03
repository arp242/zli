Rationale for writing new flag parsing stuff.

The Go standard package has some annoying properties:

- It stops parsing flags at the first non-flag argument:

      prog -v test     Works as expected
      prog test -v     Stops parsing at "test" and "-v" is added to flag.Args()

  I don't really like that. If I write `foo bar` then I want to be able to
  quickly add `-v` as `foo bar -v` without having to edit the commandline to put
  it as `foo -v bar`. Tools like "go test" contain some special hackery  to work
  around this, so "go test -v" works.

  You can still use `--` if you want to stop flag parsing: `foo -- bar -v`.

- Awkward to see if a flag wasn't given at all vs. whether someone passed an
  empty string. This is useful sometimes. You also can't make a flag optional,
  (i.e. have `-opt` and `-opt value` both work).

- I don't like the automatic usage-generation (it looks ugly) and usage often
  prints to stderr even when it really shouldn't (this is fixable, but many
  people don't).

- Awkward to add flag aliases (e.g. have both `-v` and `-verbose`).

- `prog -ab` is interpreted as a single flag even when they're boolean flags,
  instead of `prog -a -b`.

- You can't add flags more than once (and it won't give an error if you do!),
  the last value is always used. So something like the "-v for verbose, -vv for
  more verbose"-pattern or "-exclude foo -exclude bar" is hard.

- `flag.StringVar(&s, ..)` requires declaring the variable first, and `s :=
  flag.String(..)` means having pointers everywhere. Neither is great IMO.

I looked at some existing libraries and there were always some things to my
dislike.

Note that the below isn't a full review, just a very short "why not?" list. Most
of these packages are really not bad at all and I've used several of them with
great success, and all of them also cover more use cases. Please don't take it
as a "this sucks"-list.

- https://github.com/jessevdk/go-flags<br>
  Struct tag misuse, seems pretty complex.

- https://github.com/spf13/cobra<br>
  Has some specific assumptions (i.e. that you'll use subcommands), can be kinda
  tricky to use, has weird caveats.

- https://github.com/urfave/cli<br>
  Need to make a struct.

- https://github.com/peterbourgon/ff<br>
  Front-end for flag which inherits all the problems.

- https://github.com/alecthomas/kingpin<br>
  Has quite a few dependencies; not entirely sure if I like the API.

- https://github.com/alecthomas/kong<br>
  Struct tag abuse.
