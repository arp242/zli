package zli

// This file contains a simple and incomplete implementation of the terminfo
// database. Information was taken from the ncurses manpages term(5) and
// terminfo(5). Currently, only the string capabilities for special keys and for
// functions without parameters are actually used. Colors are still done with
// ANSI escape sequences. Other special features that are not (yet?) supported
// are reading from ~/.terminfo, the TERMINFO_DIRS variable, Berkeley database
// format and extended capabilities.

import (
	"bytes"
	"encoding/binary"
	"encoding/hex"
	"errors"
	"fmt"
	"io/ioutil"
	"os"
	"strings"
)

const (
	headerSize = 12
)

// "Maps" the function constants from termbox.go to the number of the respective
// string capability in the terminfo file. Taken from (ncurses) term.h.
//
// TODO: should make constants for the functions; but it's not actually used
// now.
var capsMap = map[int16]Cap{
	28:  CapEnterCA,
	40:  CapExitCA,
	16:  CapShowCursor,
	13:  CapHideCursor,
	5:   CapClearScreen,
	39:  CapSGR0,
	36:  CapUnderline,
	27:  CapBold,
	32:  CapHidden,
	26:  CapBlink,
	30:  CapDim,
	311: CapCursive,
	34:  CapReverse,
	89:  CapEnterKeypad, // keypad_xmit
	88:  CapExitKeypad,  // keypad_local
}

var keysMap = map[int16]Key{
	66:  KeyF1,
	68:  KeyF2, // apparently not a typo; 67 is F10 for whatever reason
	69:  KeyF3,
	70:  KeyF4,
	71:  KeyF5,
	72:  KeyF6,
	73:  KeyF7,
	74:  KeyF8,
	75:  KeyF9,
	67:  KeyF10,
	216: KeyF11,
	217: KeyF12,
	77:  KeyInsert,
	59:  KeyDelete,
	76:  KeyHome,
	164: KeyEnd,
	82:  KeyPgUp,
	81:  KeyPgDn,
	87:  KeyUp,
	61:  KeyDown,
	79:  KeyLeft,
	83:  KeyRight,
}

func newTerminfo(term string, noBuiltin bool) (*Terminfo, error) {
	if term == "" {
		return nil, errors.New("TERM not set")
	}

	// Prefer the built-in ones.
	if !noBuiltin {
		t, err := getBuiltin(term)
		if err == nil {
			return t, nil
		}
	}

	// Load from filesystem.
	data, err := findTerminfo(term)
	if err != nil {
		return nil, err
	}
	rd := bytes.NewReader(data)

	// 0: magic number
	// 1: size of names section
	// 2: size of boolean section
	// 3: size of numbers section (in integers)
	// 4: size of the strings section (in integers)
	// 5: size of the string table
	var header [6]int16
	err = binary.Read(rd, binary.LittleEndian, header[:])
	if err != nil {
		return nil, fmt.Errorf("terminfo: reading header: %w", err)
	}

	names := make([]byte, header[1]-1)
	err = binary.Read(rd, binary.LittleEndian, &names)
	snames := strings.Split(string(names), "|")
	ti := &Terminfo{
		Name:    snames[0],
		Aliases: snames[1 : len(snames)-1],
		Desc:    snames[len(snames)-1],
		keys:    make(map[string]Key, len(keysMap)),
		caps:    make(map[string]Cap, len(capsMap)),
	}

	number_sec_len := int16(2)
	if header[0] == 542 {
		// Doc says it should be octal 0542, but what I see it terminfo files is
		// 542, learn to program please... thank you..
		number_sec_len = 4
	}

	if (header[1]+header[2])%2 != 0 {
		// Old quirk to align everything on word boundaries
		header[2] += 1
	}
	strOffset := headerSize + header[1] + header[2] + number_sec_len*header[3]
	tableOffset := strOffset + 2*header[4]

	for o, k := range keysMap {
		seq, err := readString(rd, strOffset+2*o, tableOffset)
		if err != nil {
			return nil, fmt.Errorf("terminfo: reading key %q at 0x%x: %w", k, strOffset+2*o, err)
		}
		ti.keys[seq] = k
		addModifierKeys(ti, seq, k)
	}

	for o, c := range capsMap {
		seq, err := readString(rd, strOffset+2*o, tableOffset)
		if err != nil {
			return nil, fmt.Errorf("terminfo: reading cap %q at 0x%x: %w", c, strOffset+2*o, err)
		}
		ti.caps[seq] = c
	}

	return ti, nil
}

// Modifiers for special keys work with suffixes:
//
//      Regular   Ctrl     Shift    Alt
// F1   OP        [1;5P    [1;2P    [1;3P
// F5   [15~      [15;5~   [15;2~   [15;3~
// Up   OA        [1;5A    [1;2A    [1;3A
//
//   2 = Shift
//   3 = Alt
//   5 = Ctrl
//
// There are some others (Meta) and combinations (Shift+Ctrl), but we don't
// support this.
//
// https://invisible-island.net/xterm/ctlseqs/ctlseqs.pdf
func addModifierKeys(ti *Terminfo, seq string, k Key) {
	if strings.HasPrefix(seq, "\x1b[") && seq[len(seq)-1] == '~' {
		noTilde := seq[:len(seq)-1]
		ti.keys[noTilde+";2~"] = k | Shift
		ti.keys[noTilde+";3~"] = k | Alt
		ti.keys[noTilde+";5~"] = k | Ctrl
	} else if strings.HasPrefix(seq, "\x1bO") && len(seq) == 3 {
		noCSI := seq[2:]
		ti.keys["\x1b[1;2"+noCSI] = k | Shift
		ti.keys["\x1b[1;3"+noCSI] = k | Alt
		ti.keys["\x1b[1;5"+noCSI] = k | Ctrl
	}
}

// This behaviour follows the one described in terminfo(5) as distributed by
// ncurses.
func findTerminfo(term string) ([]byte, error) {
	ti := os.Getenv("TERMINFO")
	if ti != "" { // No other directory should be searched.
		return fromPath(term, ti)
	}

	if h := os.Getenv("HOME"); h != "" {
		data, err := fromPath(term, h+"/.terminfo")
		if err == nil {
			return data, nil
		}
	}

	if dirs := os.Getenv("TERMINFO_DIRS"); dirs != "" {
		for _, dir := range strings.Split(dirs, ":") {
			if dir == "" {
				dir = "/usr/share/terminfo"
			}
			data, err := fromPath(term, dir)
			if err == nil {
				return data, nil
			}
		}
	}

	data, err := fromPath(term, "/lib/terminfo")
	if err == nil {
		return data, nil
	}

	return fromPath(term, "/usr/share/terminfo")
}

func fromPath(term, path string) ([]byte, error) {
	// The typical *nix path ("x/xterm")
	terminfo := path + "/" + term[0:1] + "/" + term
	data, err := ioutil.ReadFile(terminfo)
	if err == nil {
		return data, nil
	}

	// Darwin specific dirs structure.
	terminfo = path + "/" + hex.EncodeToString([]byte(term[:1])) + "/" + term
	return ioutil.ReadFile(terminfo)
}

var builtinTermsCompat = map[string]*Terminfo{
	// "xterm":  builtinTerms["xterm"],
	// "rxvt":   builtinTerms["rxvt-unicode"],
	// "linux":  builtinTerms["linux"],
	// "Eterm":  builtinTerms["Eterm"],
	// "screen": builtinTerms["screen"],

	// // let's assume that 'cygwin' is xterm compatible
	// "cygwin": builtinTerms["xterm"],
	// "st":     builtinTerms["xterm"],
}

func getBuiltin(term string) (*Terminfo, error) {
	if t, ok := builtinTerms[term]; ok {
		return t, nil
	}

	// Try compatibility variants.
	for m, t := range builtinTermsCompat {
		if strings.Contains(term, m) {
			return t, nil
		}
	}

	return nil, fmt.Errorf("unsupported terminal %q", term)
}

func readString(rd *bytes.Reader, strOff, table int16) (string, error) {
	_, err := rd.Seek(int64(strOff), 0)
	if err != nil {
		return "", fmt.Errorf("seek strOff 0x%x: %w", strOff, err)
	}

	var off int16
	err = binary.Read(rd, binary.LittleEndian, &off)
	if err != nil {
		return "", fmt.Errorf("read table: %w", err)
	}
	_, err = rd.Seek(int64(table+off), 0)
	if err != nil {
		return "", fmt.Errorf("seek table 0x%x: %w", table+off, err)
	}

	var bs []byte
	for {
		b, err := rd.ReadByte()
		if err != nil {
			return "", fmt.Errorf("read data: %w", err)
		}
		if b == 0x00 {
			break
		}
		bs = append(bs, b)
	}

	return string(bs), nil
}
