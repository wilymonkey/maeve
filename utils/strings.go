package utils

import (
	"fmt"
	"math"
	"strconv"
	"strings"
	"sync"
	"unicode"
)

func BytesToHuman(bytes int64) string {
	var units = []string{"KB", "MB", "GB", "TB", "PB"}

	if bytes < 1000 {
		return fmt.Sprintf("%d B", bytes)
	}

	exponent := min(int(math.Floor(math.Log10(float64(bytes))/math.Log10(1000))), len(units))
	decimal := float64(bytes) / math.Pow(1000, float64(exponent))
	return fmt.Sprintf("%.2f %s", decimal, units[exponent-1])
}

func ParseHumanBytes(s string) (int64, error) {
	units := map[string]int64{
		"B":  1,
		"KB": 1000,
		"MB": 1000 * 1000,
		"GB": 1000 * 1000 * 1000,
		"TB": 1000 * 1000 * 1000 * 1000,
		"PB": 1000 * 1000 * 1000 * 1000 * 1000,
	}

	s = strings.TrimSpace(strings.ToUpper(s))
	var numStr, unit string
	for i, r := range s {
		if !unicode.IsDigit(r) && r != '.' {
			numStr = strings.TrimSpace(s[:i])
			unit = strings.TrimSpace(s[i:])
			break
		}
	}

	if numStr == "" {
		numStr = s
		unit = "b"
	}

	multiplier, ok := units[unit]
	if !ok {
		return 0, fmt.Errorf("unknown unit: %q", unit)
	}

	val, err := strconv.ParseFloat(numStr, 64)
	if err != nil {
		return 0, fmt.Errorf("parsing number: %w", err)
	}

	return int64(val * float64(multiplier)), nil
}

func TruncateString(txt string, maxwidth int) string {
	var width int

	iter := NewRuneIterBackward(txt)
	for {
		r, ok := iter.Next()
		if !ok {
			return txt
		}
		width += runeWidth(r)
		if width >= maxwidth-widthWide {
			return "…" + txt[iter.i:]
		}
	}
}

type RuneIterBackward struct {
	s string
	i int
}

func NewRuneIterBackward(s string) *RuneIterBackward {
	return &RuneIterBackward{s: s, i: len(s) - 1}
}

// Next returns the next rune going backwards, or ok=false when done.
func (it *RuneIterBackward) Next() (r rune, ok bool) {
	if it.i < 0 {
		return 0, false
	}

	// Move to the start byte of the rune (skip continuation bytes)
	for (it.s[it.i] & 0xC0) == 0x80 {
		it.i--
	}

	b := it.s[it.i]
	switch {
	case b&0x80 == 0: // 1-byte rune
		r = rune(b)
	case b&0xE0 == 0xC0: // 2-byte rune
		r = rune(b&0x1F)<<6 |
			rune(it.s[it.i+1]&0x3F)
	case b&0xF0 == 0xE0: // 3-byte rune
		r = rune(b&0x0F)<<12 |
			rune(it.s[it.i+1]&0x3F)<<6 |
			rune(it.s[it.i+2]&0x3F)
	case b&0xF8 == 0xF0: // 4-byte rune
		r = rune(b&0x07)<<18 |
			rune(it.s[it.i+1]&0x3F)<<12 |
			rune(it.s[it.i+2]&0x3F)<<6 |
			rune(it.s[it.i+3]&0x3F)
	default:
		// invalid UTF-8 leading byte, treat as raw byte
		r = rune(b)
	}

	it.i-- // step left
	return r, true
}

const (
	widthThin   = 1
	widthNormal = 2
	widthWide   = 3
	widthCJK    = 4
)

var (
	asciiWidth [128]int
	initOnce   sync.Once
)

func initAsciiWidth() {
	for i := range asciiWidth {
		asciiWidth[i] = widthNormal
	}
	for _, r := range []byte{'i', 'l', '!', '.', ',', ':', ';', '|', '\'', '"'} {
		asciiWidth[r] = widthThin
	}
	for _, r := range []byte{'W', 'M', 'O', 'Q'} {
		asciiWidth[r] = widthWide
	}
}

func runeWidth(r rune) int {
	initOnce.Do(initAsciiWidth)

	if r < 0x80 {
		return asciiWidth[r]

	}

	switch {
	case r < 0x3000:
		// Nothing interesting below 0x3000 except ASCII/Latin ext already handled
		return widthNormal

	case r < 0x3100:
		// Hiragana 3040–309F
		if r >= 0x3040 {
			return widthCJK
		}
		return widthNormal

	case r < 0x3400:
		// Katakana 30A0–30FF
		if r >= 0x30A0 && r <= 0x30FF {
			return widthCJK
		}
		return widthNormal

	case r < 0xA000:
		// CJK Unified 4E00–9FFF
		if r >= 0x4E00 {
			return widthCJK
		}
		return widthNormal

	case r < 0xD800:
		// Hangul AC00–D7AF
		if r >= 0xAC00 {
			return widthCJK
		}
		return widthNormal

	default:
		return widthNormal
	}
}
