package utils

import (
	"fmt"
	"math"
	"strconv"
	"strings"
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
	txtWidth := stringWidth(txt)
	if txtWidth <= maxwidth {
		return txt
	}

	ellipsis := "…"
	ellipsisWidth := widthWide
	availWidth := maxwidth - ellipsisWidth
	if availWidth <= 0 {
		return ellipsis
	}

	avgCharWidth := txtWidth / len(txt)
	estChars := availWidth / avgCharWidth
	if estChars <= 0 {
		return ellipsis
	}
	if estChars > len(txt) {
		return txt
	}
	start := len(txt) - estChars
	return ellipsis + txt[start:]
}

const (
	widthThin   = 1
	widthNormal = 2
	widthWide   = 3
	widthCJK    = 4
)

var runeWidth = map[rune]int{
	// THIN
	'i': widthThin, 'l': widthThin, '!': widthThin, '.': widthThin,
	',': widthThin, ':': widthThin, ';': widthThin, '|': widthThin,
	'\'': widthThin, '"': widthThin,
	// THICC
	'W': widthWide, 'M': widthWide, 'O': widthWide, 'Q': widthWide,
}

func stringWidth(s string) int {
	var width int
	for _, r := range s {
		switch {
		// Fast path for ASCII first
		case r < 0x80:
			if val, ok := runeWidth[r]; ok {
				width += val
			} else {
				width += widthNormal
			}
		// CJK ranges (common + extensions + Hangul + Kana)
		case (r >= 0x4E00 && r <= 0x9FFF) || // CJK Unified
			(r >= 0x3400 && r <= 0x4DBF) || // CJK Extension A
			(r >= 0xAC00 && r <= 0xD7AF) || // Hangul
			(r >= 0x3040 && r <= 0x309F) || // Hiragana
			(r >= 0x30A0 && r <= 0x30FF): // Katakana
			width += widthCJK

		default:
			width += widthNormal
		}
	}
	return width
}
