package utils

import (
	"fmt"
	"math"
)

func BytesToHuman(bytes int64) string {
	if bytes < 1000 {
		return fmt.Sprintf("%d B", bytes)
	}
	units := []string{"KB", "MB", "GB", "TB", "PB", "EB", "ZB", "YB"}
	exponent := min(int(math.Floor(math.Log10(float64(bytes))/math.Log10(1000))), len(units))
	decimal := float64(bytes) / math.Pow(1000, float64(exponent))
	return fmt.Sprintf("%.2f %s", decimal, units[exponent-1])
}
