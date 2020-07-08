package zli

import (
	"fmt"
	"os"
	"strconv"
	"strings"

	"zgo.at/zli/internal/isatty"
)

type Attribute int64

// Basic terminal attributes.
const (
	Reset Attribute = iota
	Bold
	Faint
	Italic
	Underline
	BlinkSlow
	BlinkRapid
	ReverseVideo
	Concealed
	CrossedOut
)

// First 16 colours.
const (
	Black Attribute = iota + 100
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

// NoColor indicates that the program shouldn't output any colours, based on
// thet output terminal and NO_COLOR environment variable.
//
// You can override this if the user sets "--color=force" or the like.
var NoColor = func() bool {
	_, ok := os.LookupEnv("NO_COLOR")
	return os.Getenv("TERM") == "dumb" || !isatty.IsTerminal(os.Stdout.Fd()) || ok
}()

// Background transforms a colour to a background colour.
func Background(a Attribute) Attribute { return -a }

// Palette gets colour from the fixed 256-colour palette. The first 16 (starting
// at 0) are the same as the colour names (Black, Red, etc.)
//
// 16 to 231 are various colours. 232 to 255 are greyscale tones.
func Palette(n int) Attribute { return Attribute(n + 100) }

// TrueColor gets a 24-bit "true colour" from a hex string such as "#f44" or
// "#ff4444". The leading "#" is optional.
//
// Parsing errors are signaled with -0 (signed zero), which Color() shows as
// "zli.Color!(ERROR n=1)", where 1 is the argument index.
func TrueColor(hex string) Attribute {
	hex = strings.TrimPrefix(hex, "#")
	if len(hex) == 3 {
		hex = strings.Repeat(string(hex[0]), 2) +
			strings.Repeat(string(hex[1]), 2) +
			strings.Repeat(string(hex[2]), 2)
	}

	var rgb []byte
	n, err := fmt.Sscanf(strings.ToLower(hex), "%x", &rgb)
	if err != nil {
		return -0
	}
	if n != 1 || len(rgb) != 3 { // I don't think this can ever happen.
		return -0
	}

	return Attribute(1000 +
		int64(rgb[0]) +
		int64(rgb[1])<<8 +
		int64(rgb[2])<<16)
}

// Colorln prints colourized output.
func Colorln(text string, attrs ...Attribute) {
	fmt.Println(Color(text, attrs...))
}

// Color applies terminal escape codes on the text.
//
// This will do nothing of NoColor is true.
func Color(text string, attrs ...Attribute) string {
	if len(attrs) == 0 || NoColor {
		return text
	}

	buf := new(strings.Builder)
	buf.WriteString("\x1b[")
	for i, a := range attrs {
		if a == -0 {
			return fmt.Sprintf("zli.Color!(ERROR n=%d)", i)
		}
		bg := a < 0
		if bg {
			a = -a
		}

		switch {
		case a <= 10:
			buf.WriteString(strconv.FormatInt(int64(a), 10))

		// 256bit
		case a <= 355:
			if bg {
				buf.WriteString("48;5;")
			} else {
				buf.WriteString("38;5;")
			}
			buf.WriteString(strconv.FormatInt(int64(a-100), 10))

		// True colour
		case a <= 2<<23+1000:
			a -= 1000
			if bg {
				fmt.Fprintf(buf, "48;2;%d;%d;%d", a%256, a>>8%256, a>>16%256)
			} else {
				fmt.Fprintf(buf, "38;2;%d;%d;%d", a%256, a>>8%256, a>>16%256)
			}
		}
		if len(attrs)-1 != i {
			buf.WriteRune(';')
		}
	}
	buf.WriteRune('m')
	buf.WriteString(text)

	buf.WriteString("\x1b[0m")
	return buf.String()
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
