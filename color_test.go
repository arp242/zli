package zli_test

import (
	"strings"
	"testing"

	"zgo.at/zli"
)

func ExampleColor() {
	zli.Colorln("You're looking rather red", zli.Red) // Apply a colour.
	zli.Colorln("A bold move", zli.Bold)              // Or an attribute.
	zli.Colorln("Tomato", zli.Red.Bg())               // Transform to background colour.

	zli.Colorln("Wow, such beautiful text", // Can be combined.
		zli.Bold, zli.Underline, zli.Red, zli.Green.Bg())

	zli.Colorln("Contrast ratios is for suckers", // 256 colour
		zli.Palette(56), zli.Palette(99).Bg())

	zli.Colorln("REAL men use TRUE color!", // True colour
		zli.TrueColor("#fff"), zli.TrueColor("#00f").Bg())

	// Output:
	// [38;5;1mYou're looking rather red[0m
	// [1mA bold move[0m
	// [48;5;1mTomato[0m
	// [1;4;38;5;1;48;5;2mWow, such beautiful text[0m
	// [38;5;56;48;5;99mContrast ratios is for suckers[0m
	// [38;2;255;255;255;48;2;0;0;255mREAL men use TRUE color![0m
}

func TestColor(t *testing.T) {
	tests := []struct {
		in   []zli.Attribute
		want string
	}{
		{[]zli.Attribute{}, "Hello"},
		{[]zli.Attribute{zli.TrueColor("chucknorris")}, "zli.Color!(ERROR n=0)"},

		// Test boundaries (first and last).
		{[]zli.Attribute{zli.Black, zli.Black.Bg()}, "\x1b[38;5;0;48;5;0mHello\x1b[0m"},
		{[]zli.Attribute{zli.BrightWhite, zli.BrightWhite.Bg()}, "\x1b[38;5;15;48;5;15mHello\x1b[0m"},
		{[]zli.Attribute{zli.BrightWhite + 1, (zli.BrightWhite + 1).Bg()}, "\x1b[38;5;16;48;5;16mHello\x1b[0m"},
		{[]zli.Attribute{355, zli.Attribute(355).Bg()}, "\x1b[38;5;255;48;5;255mHello\x1b[0m"},
		{
			[]zli.Attribute{zli.TrueColor("#000"), zli.TrueColor("#000").Bg()},
			"\x1b[38;2;0;0;0;48;2;0;0;0mHello\x1b[0m",
		},
		{
			[]zli.Attribute{zli.TrueColor("#fff"), zli.TrueColor("#fff").Bg()},
			"\x1b[38;2;255;255;255;48;2;255;255;255mHello\x1b[0m",
		},

		{[]zli.Attribute{zli.Bold}, "\x1b[1mHello\x1b[0m"},
		{[]zli.Attribute{zli.Bold.Bg()}, "\x1b[1mHello\x1b[0m"}, // Doesn't make much sense, but should work nonetheless.

		{[]zli.Attribute{zli.Red}, "\x1b[38;5;1mHello\x1b[0m"},
		{[]zli.Attribute{zli.Bold, zli.Red}, "\x1b[1;38;5;1mHello\x1b[0m"},
		{[]zli.Attribute{zli.Red, zli.Underline}, "\x1b[38;5;1;4mHello\x1b[0m"},
		{[]zli.Attribute{zli.Green.Bg()}, "\x1b[48;5;2mHello\x1b[0m"},
		{[]zli.Attribute{zli.Green.Bg(), zli.Bold}, "\x1b[48;5;2;1mHello\x1b[0m"},
		{[]zli.Attribute{zli.Green.Bg(), zli.Red}, "\x1b[48;5;2;38;5;1mHello\x1b[0m"},
		{[]zli.Attribute{zli.Green, zli.Red.Bg(), zli.Bold, zli.Underline}, "\x1b[38;5;2;48;5;1;1;4mHello\x1b[0m"},
	}

	for _, tt := range tests {
		zli.NoColor = true
		got := zli.Color("Hello", tt.in...)
		if got != "Hello" {
			t.Errorf("NoColor not respected? got: %q", got)
		}

		zli.NoColor = false
		got = zli.Color("Hello", tt.in...)
		if got != tt.want {
			t.Errorf("\ngot:  %q → %[1]s\nwant: %q → %[2]s", got, tt.want)
		}

		if !strings.Contains(got, "ERROR") {
			de := zli.DeColor(got)
			if de != "Hello" {
				t.Errorf("DeColor: %q", de)
			}
		}
	}
}

func BenchmarkColor(b *testing.B) {
	attr := []zli.Attribute{zli.Green, zli.Red.Bg(), zli.Bold, zli.Underline}
	var s string

	b.ReportAllocs()
	for n := 0; n < b.N; n++ {
		s = zli.Color("Hello", attr...)
	}
	_ = s
}
