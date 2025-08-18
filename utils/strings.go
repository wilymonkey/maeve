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
