package zli

// ErasesLine erases the entire line and puts the cursor at the start of the
// line.
func EraseLine() { Stdout.Write([]byte("\x1b[2K\r")) }
