package zli

// These exists for gen.go

/*
func (t Terminfo) Keys() map[string]Key { return t.keys }
func (t Terminfo) Caps() map[string]Cap { return t.caps }
func NT(term string) (*Terminfo, error) { return newTerminfo(term, true) }
func (t Terminfo) KeysSorted() []string {
	sorted := make([]string, 0, len(t.keys))
	for seq, k := range t.keys {
		if k.Shift() || k.Ctrl() || k.Alt() {
			continue
		}
		sorted = append(sorted, fmt.Sprintf("\t\t\t%q: Key%s,\n", seq, k.Name()))
	}
	sort.Strings(sorted)
	return sorted
}
func (t Terminfo) CapsSorted() []string {
	sorted := make([]string, 0, len(t.caps))
	for seq, c := range t.caps {
		sorted = append(sorted, fmt.Sprintf("\t\t\t%q: Cap%s,\n", seq, c.String()))
	}
	sort.Strings(sorted)
	return sorted
}
*/
