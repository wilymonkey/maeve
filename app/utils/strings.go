package utils

import (
	"fmt"
	"strconv"
	"strings"

	"github.com/charmbracelet/x/ansi"
)

func TruncateStr(s string, width int) string {
	sWidth := ansi.StringWidth(s)
	if sWidth > width {
		// We need space for the ellipsis.
		availWidth := width - 2

		// Convert to runes for proper Unicode character handling.
		runes := []rune(s)
		if len(runes) > availWidth {
			return "… " + string(runes[len(runes)-availWidth:])
		}
	}
	return s
}

func ParseHumanBytes(s string) (int64, error) {
	units := map[string]int64{
		"b":  1,
		"kb": 1 << 10,
		"mb": 1 << 20,
		"gb": 1 << 30,
		"tb": 1 << 40,
		"pb": 1 << 50,
	}

	s = strings.TrimSpace(strings.ToLower(s))
	var numPart string
	var unitPart string

	for i, r := range s {
		if (r < '0' || r > '9') && r != '.' {
			numPart = strings.TrimSpace(s[:i])
			unitPart = strings.TrimSpace(s[i:])
			break
		}
	}
	if numPart == "" {
		numPart = s
		unitPart = "b"
	}

	multiplier, ok := units[unitPart]
	if !ok {
		return 0, fmt.Errorf("unknown unit: %q", unitPart)
	}

	val, err := strconv.ParseFloat(numPart, 64)
	if err != nil {
		return 0, fmt.Errorf("invalid number: %q", numPart)
	}

	return int64(val * float64(multiplier)), nil
}
