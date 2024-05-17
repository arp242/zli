package zli

import (
	"fmt"
	"math"
	"strconv"
	"strings"
)

/*
Color is a set of attributes to apply; the attributes are stored as follows:

	                                     fg true, 256, 16 color mode ─┬──┐
	                                  bg true, 256, 16 color mode ─┬─┐│  │
	                                                               │ ││  │┌── error parsing hex color
	   ┌───── bg color ────────────┐ ┌───── fg color ────────────┐ │ ││  ││┌─ term attr
	   v                           v v                           v v vv  vvv         v
	0b 0000_0000 0000_0000 0000_0000 0000_0000 0000_0000 0000_0000 0000_0000 0000_0000

The terminal attributes are bold, underline, etc. are stored as flags. The error
flag signals there was an error parsing a hex color with ColorHex().

The colors are stored for the background and foreground and are applied
depending on the values of the color mode bitmasks.

The biggest advantage of storing it like this is that we can can use a single
variable/function parameter to represent all terminal attributes, which IMHO
gives a rather nicer API than using a slice or composing the colors with
functions or some such.
*/
type Color uint64

// Offsets where foreground and background colors are stored.
const (
	ColorOffsetFg = 16
	ColorOffsetBg = 40
)

// Mask anything that's not a foreground or background color.
const (
	maskFg Color = (256*256*256 - 1) << ColorOffsetFg
	maskBg Color = maskFg << (ColorOffsetBg - ColorOffsetFg)
)

// Basic terminal attributes.
const (
	Reset Color = 0
	Bold  Color = 1 << (iota - 1)
	Dim
	Italic
	Underline
	Undercurl
	Overline
	Reverse
	Concealed
	StrikeOut
)

var allAttrs = []Color{Bold, Dim, Italic, Underline, Undercurl, Overline, Reverse, Concealed, StrikeOut}

// ColorError signals there was an error in parsing a color hex attribute.
const ColorError Color = StrikeOut << 1

// Color modes.
const (
	ColorMode16Fg Color = ColorError << (iota + 1)
	ColorMode256Fg
	ColorModeTrueFg

	ColorMode16Bg
	ColorMode256Bg
	ColorModeTrueBg
)

// First 8 colors; use Brighten(1) to get the bright variants.
const (
	Black Color = (iota << ColorOffsetFg) | ColorMode16Fg
	Red
	Green
	Yellow
	Blue
	Magenta
	Cyan
	White
)

// Bg returns the background variant of this color. If doesn't do anything if
// this is already a background color.
func (c Color) Bg() Color {
	if c&ColorMode16Fg != 0 {
		c ^= ColorMode16Fg | ColorMode16Bg
	} else if c&ColorMode256Fg != 0 {
		c ^= ColorMode256Fg | ColorMode256Bg
	} else if c&ColorModeTrueFg != 0 {
		c ^= ColorModeTrueFg | ColorModeTrueBg
	}
	return (c &^ maskFg) | (c & maskFg << 24)
}

// Brighten or darken (for negative values) a color.
//
// Operations will never go out-of-bounds; brighting the brightest variant will
// do nothing and will simply return the same color.
//
// For 16 colors it will convert a normal color to a "bright" variant, or vice versa.
//
// For 256 colors it will shift to the same column position in the next
// "square"; see the chart printed by ./cmd/colortest. The scale of n is 6.
//
// For true colors it will brighten the color; the scale of n is 255.
func (c Color) Brighten(n int) Color {
	if n == 0 {
		return c
	}

	mask, off := maskFg, ColorOffsetFg
	if c&ColorMode16Bg != 0 || c&ColorMode256Bg != 0 || c&ColorModeTrueBg != 0 {
		mask, off = maskBg, ColorOffsetBg
	}
	keep := c &^ mask
	cc := c & mask >> off

	switch {
	case c&ColorMode16Fg != 0 || c&ColorMode16Bg != 0:
		if n > 0 {
			cc |= 8
		} else {
			cc &^= 8
		}

	case c&ColorMode256Fg != 0 || c&ColorMode256Bg != 0:
		switch {
		// First 16 (starting at 0) are the same as 16 color.
		case cc <= 15:
			if n > 0 {
				cc |= 8
			} else {
				cc &^= 8
			}
		// 232-255 are grayscale.
		case cc >= 232:
			cc = clamp(int(cc)+n, 232, 255)
		// Move to same square in next cube; see ./cmd/colortest
		default:
			col := int(15+cc) % 6
			if col == 0 {
				col = 6
			}
			row := int(math.Ceil(float64(cc-15) / 36))
			if row == 0 {
				row = 1
			}
			// TODO: this seems ... too complicated...
			max := 15 + (row*30 + ((row - 1) * 6)) + col
			min := max - 30
			//fmt.Printf("%d×%d max=%d; min=%d\n", row, col, max, min)

			cc = clamp(int(cc)+n*6, min, max)
		}
	case c&ColorModeTrueFg != 0 || c&ColorModeTrueBg != 0:
		//mod := math.Round(float64(n*3) * 1 / 3)
		or, og, ob := int(cc%256), int(cc>>8%256), int(cc>>16%256)
		r, g, b := or+n, og+n, ob+n

		if r > 255 {
			s := 255 - or
			r, g, b = 255, clampC(og+s, 0, 255), clampC(ob+s, 0, 255)
		} else if g > 255 {
			s := 255 - or
			r, g, b = clampC(or+s, 0, 255), 255, clampC(ob+s, 0, 255)
		} else if b > 255 {
			s := 255 - or
			r, g, b = clampC(or+s, 0, 255), clampC(og+s, 0, 255), 255
		} else if r < 0 {
			s := -or
			r, g, b = 0, clampC(og+s, 0, 255), clampC(ob+s, 0, 255)
		} else if g < 0 {
			s := -og
			r, g, b = clampC(or+s, 0, 255), 0, clampC(ob+s, 0, 255)
		} else if b < 0 {
			s := -ob
			r, g, b = clampC(or+s, 0, 255), clampC(og+s, 0, 255), 0
		}

		cc = Color(r) + Color(g)<<8 + Color(b)<<16
	}

	return keep | (cc << off)
}

func clampC(c, min, max int) int { return int(clamp(c, min, max)) }
func clamp(c, min, max int) Color {
	if c < min {
		return Color(min)
	}
	if c > max {
		return Color(max)
	}
	return Color(c)
}

// String gets the escape sequence for this color code.
//
// This will always return an empty string if WantColor is false or if the error
// flag is set.
//
// You can use this to set colors directly with fmt.Print:
//
//	fmt.Println(zli.Red|zli.Bold, "red!") // Set colors "directly"; Println() will call String()
//	fmt.Println("and bold!", zli.Reset)   // Don't forget to reset it!
//
//	fmt.Printf("%sc%so%sl%so%sr%s\n", zli.Red, zli.Magenta, zli.Cyan, zli.Blue, zli.Yellow, zli.Reset)
func (c Color) String() string {
	if !WantColor || c&ColorError != 0 {
		return ""
	}
	if c == Reset {
		return "\x1b[0m"
	}

	attrs := make([]string, 0, 4)
	for i := range allAttrs {
		if c&allAttrs[i] != 0 {
			switch allAttrs[i] {
			case Overline:
				attrs = append(attrs, "53")
			case Undercurl:
				attrs = append(attrs, "4:3")
			default:
				attrs = append(attrs, strconv.Itoa(i+1))
			}
		}
	}

	switch {
	case c&ColorMode16Fg != 0:
		cc := c&maskFg>>ColorOffsetFg + 30
		if cc > 37 { // Bright colors
			cc += 52
		}
		attrs = append(attrs, strconv.FormatUint(uint64(cc), 10))
	case c&ColorMode256Fg != 0:
		attrs = append(attrs, "38;5;"+strconv.FormatUint(uint64(c&maskFg>>ColorOffsetFg), 10))
	case c&ColorModeTrueFg != 0:
		cc := c & maskFg >> ColorOffsetFg
		attrs = append(attrs, "38;2;"+
			strconv.FormatUint(uint64(cc%256), 10)+";"+
			strconv.FormatUint(uint64(cc>>8%256), 10)+";"+
			strconv.FormatUint(uint64(cc>>16%256), 10))
	}

	switch {
	case c&ColorMode16Bg != 0:
		cc := c>>ColorOffsetBg + 40
		if cc > 47 { // Bright colors
			cc += 52
		}
		attrs = append(attrs, strconv.FormatUint(uint64(cc), 10))
	case c&ColorMode256Bg != 0:
		attrs = append(attrs, "48;5;"+strconv.FormatUint(uint64(c&maskBg>>ColorOffsetBg), 10))
	case c&ColorModeTrueBg != 0:
		cc := c & maskBg >> ColorOffsetBg
		attrs = append(attrs, "48;2;"+
			strconv.FormatUint(uint64(cc%256), 10)+";"+
			strconv.FormatUint(uint64(cc>>8%256), 10)+";"+
			strconv.FormatUint(uint64(cc>>16%256), 10))
	}

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

// Color256 creates a new 256-mode color.
//
// The first 16 (starting at 0) are the same as the color names (Black, Red,
// etc.) 16 to 231 are various colors. 232 to 255 are greyscale tones.
//
// The 16-231 colors should always be identical on every display (unlike the
// basic colors, whose exact color codes are undefined and differ per
// implementation).
//
// See ./cmd/colortest for a little CLI to display the colors.
func Color256(n uint8) Color { return Color(uint64(n)<<ColorOffsetFg) | ColorMode256Fg }

// ColorHex gets a 24-bit "true color" from a hex string such as "#f44" or
// "#ff4444". The leading "#" is optional.
//
// Parsing errors are signaled with by setting the ColorError flag, which
// String() shows as "(zli.Color ERROR invalid hex color)".
func ColorHex(h string) Color {
	h = strings.TrimPrefix(h, "#")
	if len(h) == 3 {
		h = string(h[0]) + string(h[0]) + string(h[1]) + string(h[1]) + string(h[2]) + string(h[2])
	}

	var rgb []byte
	n, err := fmt.Sscanf(strings.ToLower(h), "%x", &rgb)
	if err != nil || n != 1 || len(rgb) != 3 {
		return ColorError
	}
	return ColorModeTrueFg | Color((uint64(rgb[0])|uint64(rgb[1])<<8|uint64(rgb[2])<<16)<<ColorOffsetFg)
}

// Colorize the text with a color if WantColor is true.
//
// The text will end with the reset code.
func Colorize(text string, c Color) string {
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

// Colorf prints colorized output if WantColor is true.
//
// The text will end with the reset code. Note that this is always added at the
// end, after any newlines in the string.
func Colorf(format string, c Color, a ...interface{}) { fmt.Fprintf(Stdout, Colorize(format, c), a...) }

// Colorln prints colorized output if WantColor is true.
//
// The text will end with the reset code.
func Colorln(text string, c Color) { fmt.Fprintln(Stdout, Colorize(text, c)) }

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
