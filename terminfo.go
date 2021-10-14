package zli

import (
	"fmt"
	"os"
	"sort"
	"strings"
)

type Terminfo struct {
	Name, Desc string
	Aliases    []string
	keys       map[string]Key
	caps       map[string]Cap
}

// NewTerminfo reads the terminfo for the current terminal.
func NewTerminfo() (*Terminfo, error) {
	return newTerminfo(os.Getenv("TERM"), false)
}

func (t Terminfo) String() string {
	b := new(strings.Builder)
	b.WriteString(t.Name + " – " + t.Desc + "\n")
	b.WriteString(strings.Join(t.Aliases, ", ") + "\n")

	sorted := make([]string, 0, len(t.keys))
	for seq, k := range t.keys {
		if k.Shift() || k.Ctrl() || k.Alt() {
			continue
		}
		kk := k.String()
		sorted = append(sorted, fmt.Sprintf("  %s%s %#v\n",
			kk, strings.Repeat(" ", 20-len(kk)), seq))
	}
	sort.Strings(sorted)
	b.WriteString("\nKeys:\n")
	for _, s := range sorted {
		b.WriteString(s)
	}

	sorted = make([]string, 0, len(t.caps))
	for seq, c := range t.caps {
		cc := c.String()
		sorted = append(sorted, fmt.Sprintf("  %s%s %#v\n",
			cc, strings.Repeat(" ", 20-len(cc)), seq))
	}
	sort.Strings(sorted)
	b.WriteString("\nCaps:\n")
	for _, s := range sorted {
		b.WriteString(s)
	}

	return b.String()
}

// Find a key from an escape sequence.
func (t Terminfo) FindKey(s string) Key {
	k, ok := t.keys[s]
	if !ok {
		return UnknownSequence
	}
	return k
}
