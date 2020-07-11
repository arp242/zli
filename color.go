package zli

import (
	"fmt"
	"os"
	"strconv"
	"strings"

	"zgo.at/zli/internal/isatty"
)

const (
	// Offsets where foreground and background colors are stored.
	ColorOffsetFg = 16
	ColorOffsetBg = 40

	// Mask anything that's not a foreground or background colur.
	maskFg Color = (255 | (255 << 8) | (16777215 << 40))
	maskBg Color = (255 | (255 << 8) | (16777215 << 16))

	// Mask just the foreground color.
	maskFgOnly Color = (255 << 16) | (255 << 24) | (255 << 32)
)

/*
Color is a set of attributes to apply; the attributes are stored as follows:

                                         fg true, 256, 16 color mode ─┬──┐
                                      bg true, 256, 16 color mode ─┬─┐│  │
                                                                   │ ││  │┌── error parsing hex color
       ┌───── bg color ────────────┐ ┌───── fg color ────────────┐ │ ││  ││┌─ term attr
       v                           v v                           v v vv  vvv         v
    0b 0000_0000 0000_0000 0000_0000 0000_0000 0000_0000 0000_0000 0000_0000 0000_0000

The terminal attributes are bold, underline, etc. are stored as a bitmask. The
error flag signals there was an error parsing a hex color with ColorHex().

The colors are stored separately for the background and foreground, the color
mode bitmasks store which colors to apply and in which mode. No colors are
applied if none of the color mode flags for the bg or fg are set.

The biggest advantage of this entire things is that we can can use a single
variable/function parameter to represent all terminal attributes, for example:

    var colorMatch = zli.Bold | zli.Red | zli.ColorHex("#f71").Bg()
    fmt.Println(zli.Colorf("foo", colorMatch))

Which gives us a rather nicer API than using a slice or whatnot :-)

If you want to be really savvy about it then you can store it as a constant too:

    const colorMatch = zli.Bold | zli.Red | (zli.Color(0xff|0x77<<8|0x11<<16) << zli.ColorOffsetBg) | zli.ColorModeTrueBg

This creates 24bit color stored as an int (0xff, 0x77, 0x11 is the same as
"#ff7711", or "#f71" in short notation) shifts it to the correct location, and
sets the flag so the background is read as a 24bit color.
*/
type Color uint64

// Basic terminal attributes.
const (
	Reset Color = 0
	Bold  Color = 1 << (iota - 1)
	Faint
	Italic
	Underline
	BlinkSlow
	BlinkRapid
	ReverseVideo
	Concealed
	CrossedOut
)

// ColorError signals there was an error in parsing a color hex attribute.
const ColorError Color = CrossedOut << 1

// Color modes.
const (
	ColorMode16 Color = ColorError << (iota + 1)
	ColorMode256
	ColorModeTrue

	ColorMode16Bg
	ColorMode256Bg
	ColorModeTrueBg
)

// First 16 colors.
const (
	Black Color = (iota << ColorOffsetFg) | ColorMode16
	Red
	Green
	Yellow
	Blue
	Magenta
	Cyan
	White
	BrightBlack
	BrightRed
	BrightGreen
	BrightYellow
	BrightBlue
	BrightMagenta
	BrightCyan
	BrightWhite
)

// Bg returns the background variant of this color. If doesn't do anything if
// this is already a background color.
func (c Color) Bg() Color {
	if c&ColorMode16 != 0 {
		c ^= ColorMode16 | ColorMode16Bg
	} else if c&ColorMode256 != 0 {
		c ^= ColorMode256 | ColorMode256Bg
	} else if c&ColorModeTrue != 0 {
		c ^= ColorModeTrue | ColorModeTrueBg
	}
	return (c &^ maskFgOnly) | ((c & maskFgOnly) << 24)
}

// String gets the starting escape sequence for this color code.
func (c Color) String() string {
	if !WantColor || c&ColorError != 0 {
		return ""
	}
	if c == Reset {
		return "\x1b[0m"
	}

	// Allocate space for 4 attributes; most people will rarely go over that,
	// and it'll avoid re-allocing on append() (e.g. for 3 attributes it's alloc
	// 3 times to grow: 1 → 2 → 4).
	attrs := make([]string, 0, 4)

	// Apply basic attributes.
	if c&Bold != 0 {
		attrs = append(attrs, "1")
	}
	if c&Faint != 0 {
		attrs = append(attrs, "2")
	}
	if c&Italic != 0 {
		attrs = append(attrs, "3")
	}
	if c&Underline != 0 {
		attrs = append(attrs, "4")
	}
	if c&BlinkSlow != 0 {
		attrs = append(attrs, "5")
	}
	if c&BlinkRapid != 0 {
		attrs = append(attrs, "6")
	}
	if c&ReverseVideo != 0 {
		attrs = append(attrs, "7")
	}
	if c&Concealed != 0 {
		attrs = append(attrs, "8")
	}
	if c&CrossedOut != 0 {
		attrs = append(attrs, "9")
	}

	switch {
	case c&ColorMode16 != 0:
		x := (c&^maskFg)>>ColorOffsetFg + 30
		if x > 37 {
			x += 52
		}
		attrs = append(attrs, strconv.FormatUint(uint64(x), 10))
	case c&ColorMode256 != 0:
		x := (c &^ maskFg) >> ColorOffsetFg
		attrs = append(attrs, "38;5;"+strconv.FormatUint(uint64(x), 10))
	case c&ColorModeTrue != 0:
		x := (c &^ maskFg) >> ColorOffsetFg
		attrs = append(attrs, "38;2;"+
			strconv.FormatUint(uint64(x%256), 10)+";"+
			strconv.FormatUint(uint64(x>>8%256), 10)+";"+
			strconv.FormatUint(uint64(x>>16%256), 10))
	}

	switch {
	case c&ColorMode16Bg != 0:
		x := (c&^maskBg)>>ColorOffsetBg + 40
		if x > 47 {
			x += 52
		}
		attrs = append(attrs, strconv.FormatUint(uint64(x), 10))
	case c&ColorMode256Bg != 0:
		x := (c &^ maskBg) >> ColorOffsetBg
		attrs = append(attrs, "48;5;"+strconv.FormatUint(uint64(x), 10))
	case c&ColorModeTrueBg != 0:
		x := (c &^ maskBg) >> ColorOffsetBg
		attrs = append(attrs, "48;2;"+
			strconv.FormatUint(uint64(x%256), 10)+";"+
			strconv.FormatUint(uint64(x>>8%256), 10)+";"+
			strconv.FormatUint(uint64(x>>16%256), 10))
	}

	// This is a bit faster than "\x1b[" + strings.Join() + "m" ... gotta
	// optimize that stuff for no reason in particular 🙃
	var b strings.Builder
	b.Grow(20)             // 1 alloc
	b.WriteString("\x1b[") // 1 alloc
	for i, a := range attrs {
		b.WriteString(a)
		if len(attrs)-1 != i {
			b.WriteRune(';')
		}
	}
	b.WriteRune('m')
	return b.String()
}

// Color256 created a new 256-mode color.
//
// The first 16 (starting at 0) are the same as the color names (Black, Red,
// etc.) 16 to 231 are various colors. 232 to 255 are greyscale tones.
func Color256(n uint8) Color {
	return Color(uint64(n)<<ColorOffsetFg) | ColorMode256
}

// ColorHex gets a 24-bit "true color" from a hex string such as "#f44" or
// "#ff4444". The leading "#" is optional.
//
// Parsing errors are signaled with -0 (signed zero), which Colorf() shows as
// "zli.Color!(ERROR n=1)", where 1 is the argument index.
func ColorHex(hex string) Color {
	hex = strings.TrimPrefix(hex, "#")
	if len(hex) == 3 {
		hex = strings.Repeat(string(hex[0]), 2) +
			strings.Repeat(string(hex[1]), 2) +
			strings.Repeat(string(hex[2]), 2)
	}

	var rgb []byte
	n, err := fmt.Sscanf(strings.ToLower(hex), "%x", &rgb)
	if err != nil || n != 1 || len(rgb) != 3 {
		return 0 | ColorError
	}

	nc := uint64(rgb[0]) | uint64(rgb[1])<<8 | uint64(rgb[2])<<16
	return Color(nc<<ColorOffsetFg) | ColorModeTrue
}

// WantColor indicates if the program should output any colors. This is
// automatically set from from the output terminal and NO_COLOR environment
// variable.
//
// You can override this if the user sets "--color=force" or the like.
//
// TODO: maybe expand this a bit with WantMonochrome or some such, so you can
// still output bold/underline/reverse text for people who don't want colors.
var WantColor = func() bool {
	_, ok := os.LookupEnv("NO_COLOR")
	return os.Getenv("TERM") != "dumb" && isatty.IsTerminal(os.Stdout.Fd()) && !ok
}()

// Colorln prints colorized output.
func Colorln(text string, c Color) {
	fmt.Println(Colorf(text, c))
}

// Colorf applies terminal escape codes on the text.
func Colorf(text string, c Color) string {
	if c == Reset {
		return text
	}
	if WantColor && c&ColorError != 0 {
		return "(zli.Color ERROR invalid hex color)" + text
	}

	attrs := c.String()
	if attrs == "" {
		return text
	}

	return attrs + text + Reset.String()
}

// DeColor removes ANSI color escape sequences from a string.
func DeColor(text string) string {
	for {
		i := strings.Index(text, "\x1b")
		if i == -1 {
			break
		}
		e := strings.IndexByte(text[i:], 'm')
		if e == -1 {
			break
		}
		text = text[:i] + text[i+e+1:]
	}
	return text
}
