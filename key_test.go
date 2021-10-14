package zli

import (
	"testing"
)

func TestKey(t *testing.T) {
	tests := []struct {
		k    Key
		want string
	}{
		{'a', "<a>"},
		{'a' | Shift, "<S-a>"},
		{'a' | Ctrl | Shift, "<S-C-a>"},
		{'a' | Ctrl | Shift | Alt, "<S-C-A-a>"},
		{KeyTab, "<Tab>"},
		{KeyTab | Ctrl, "<C-Tab>"},
		{KeyUp, "<Up>"},
		{KeyUp | Ctrl, "<C-Up>"},
		{KeyF24 + 1, "<Unknown key: 0x10000002f>"},
		{KeyF24 + 1 | Ctrl, "<C-Unknown key: 0x10000002f>"},
	}

	for _, tt := range tests {
		t.Run("", func(t *testing.T) {
			h := tt.k.String()
			if h != tt.want {
				t.Errorf("\nwant: %s\nhave: %s", tt.want, h)
			}
		})
	}
}
