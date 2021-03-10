package zli_test

import (
	"fmt"
	"os"
	"testing"

	"zgo.at/zli"
)

func ExampleColor() {
	zli.Stdout = os.Stdout
	zli.Colorln("You're looking rather red", zli.Red) // Apply a color.
	zli.Colorln("A bold move", zli.Bold)              // Or an attribute.
	zli.Colorln("Tomato", zli.Red.Bg())               // Transform to background color.

	zli.Colorln("Wow, such beautiful text", // Can be combined.
		zli.Bold|zli.Underline|zli.Red|zli.Green.Bg())

	zli.Colorln("Contrast ratios is for suckers", // 256 color
		zli.Color256(56)|zli.Color256(99).Bg())

	zli.Colorln("REAL men use TRUE color!", // True color
		zli.ColorHex("#678")|zli.ColorHex("#abc").Bg())

	fmt.Println(zli.Red|zli.Bold, "red!") // Set colors "directly"
	fmt.Println("and bold!", zli.Reset)
	fmt.Printf("%sc%so%sl%so%sr%s\n", zli.Red, zli.Magenta, zli.Cyan, zli.Blue, zli.Yellow, zli.Reset)

	// Output:
	// [31mYou're looking rather red[0m
	// [1mA bold move[0m
	// [41mTomato[0m
	// [1;4;31;42mWow, such beautiful text[0m
	// [38;5;56;48;5;99mContrast ratios is for suckers[0m
	// [38;2;102;119;136;48;2;170;187;204mREAL men use TRUE color![0m
	// [1;31m red!
	// and bold! [0m
	// [31mc[35mo[36ml[34mo[33mr[0m
}

func TestColor(t *testing.T) {
	tests := []struct {
		in   zli.Color
		want string
	}{
		// Basic terminal attributes
		{zli.Bold, "\x1b[1m"},
		{zli.Underline, "\x1b[4m"},
		{zli.Bold | zli.Underline, "\x1b[1;4m"},

		// Color boundaries (first and last).
		{zli.Black | zli.Black.Bg(), "\x1b[30;40m"},
		{zli.BrightWhite | zli.BrightWhite.Bg(), "\x1b[97;107m"},

		{zli.Color256(0) | zli.Color256(0).Bg(), "\x1b[38;5;0;48;5;0m"},
		{zli.Color256(255) | zli.Color256(255).Bg(), "\x1b[38;5;255;48;5;255m"},
		{zli.ColorHex("#678") | zli.ColorHex("#abc").Bg(), "\x1b[38;2;102;119;136;48;2;170;187;204m"},
		{zli.ColorHex("#678") | zli.ColorHex("#abc").Bg(), "\x1b[38;2;102;119;136;48;2;170;187;204m"},

		// Various combinations.
		{zli.Red, "\x1b[31m"},
		{zli.Bold | zli.Red, "\x1b[1;31m"},
		{zli.Red | zli.Underline, "\x1b[4;31m"},
		{zli.Green.Bg(), "\x1b[42m"},
		{zli.Green.Bg() | zli.Bold, "\x1b[1;42m"},
		{zli.BrightGreen.Bg() | zli.Red, "\x1b[31;102m"},
		{zli.Color256(99) | zli.Red.Bg() | zli.Bold | zli.Underline, "\x1b[1;4;38;5;99;41m"},

		{zli.Bold | zli.Faint | zli.Italic | zli.Underline | zli.BlinkSlow | zli.BlinkRapid | zli.ReverseVideo | zli.Concealed | zli.CrossedOut,
			"\x1b[1;2;3;4;5;6;7;8;9m"},

		{zli.Bold.Bg(), "\x1b[1m"},                 // Doesn't make much sense, but should work nonetheless.
		{zli.Color(zli.Red.Bg().Bg()), "\x1b[41m"}, // Double .Bg() does nothing
	}

	for _, tt := range tests {
		t.Run("", func(t *testing.T) {
			zli.WantColor = false
			t.Run("WantColor=false", func(t *testing.T) {
				got := tt.in.String()
				if got != "" {
					t.Errorf("Colorf WantColor not respected? got: %q", got)
				}
				got = zli.Colorf("Hello", tt.in)
				if got != "Hello" {
					t.Errorf("Colorf WantColor not respected? got: %q", got)
				}
			})

			zli.WantColor = true
			t.Run("String", func(t *testing.T) {
				got := tt.in.String()
				if got != tt.want {
					t.Errorf("Color.String()\ngot:  %q → %[1]s\nwant: %q → %[2]s", got, tt.want)
				}
			})

			t.Run("Colorf", func(t *testing.T) {
				got := zli.Colorf("Hello", tt.in)
				if got != tt.want+"Hello\x1b[0m" {
					t.Errorf("Colorf()\ngot:  %q → %[1]s\nwant: %q → %[2]s", got, tt.want)
				}
			})

			t.Run("DeColor", func(t *testing.T) {
				got := zli.Colorf("Hello", tt.in)
				de := zli.DeColor(got)
				if de != "Hello" {
					t.Errorf("DeColor: %q", de)
				}
			})
		})
	}

	t.Run("Reset", func(t *testing.T) {
		c := zli.Reset

		zli.WantColor = false
		got := c.String()
		if got != "" {
			t.Errorf("Color.String()\ngot:  %q\nwant: %q", got, "")
		}

		zli.WantColor = true
		got = c.String()
		if got != "\x1b[0m" {
			t.Errorf("Color.String()\ngot:  %q\nwant: %q", got, "\x1b[0m")
		}

		got = zli.Colorf("Hello", c)
		if got != "Hello" {
			t.Errorf("Color.String()\ngot:  %q\nwant: %q", got, "Hello")
		}
	})

	t.Run("errors", func(t *testing.T) {
		tests := []zli.Color{
			//zli.Color256(-1),
			//zli.Color256(256),
			zli.ColorHex("chucknorris"),
			zli.ColorHex("#12"),
			zli.ColorHex("#1234"),
			zli.ColorHex("#12345"),
			zli.ColorHex("#1234567"),
			zli.ColorHex("#12345678"),
			zli.ColorHex("#123456789"),
			zli.ColorHex("#1234567890"),
		}

		zli.WantColor = true
		for _, tt := range tests {
			t.Run("String()", func(t *testing.T) {
				got := tt.String()
				if got != "" {
					t.Errorf("%q", got)
				}
			})
			t.Run("Colorf()", func(t *testing.T) {
				got := zli.Colorf("Hello", tt)
				want := "(zli.Color ERROR invalid hex color)Hello"
				if got != want {
					t.Errorf("\ngot:  %q\nwant: %q", got, want)
				}
			})
		}
	})
}

func BenchmarkColor(b *testing.B) {
	c := zli.Green | zli.Red.Bg() | zli.Bold | zli.Underline
	var s string

	b.ReportAllocs()
	for n := 0; n < b.N; n++ {
		s = zli.Colorf("Hello", c)
	}
	_ = s
}
