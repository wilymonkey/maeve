package utils

import (
	"fmt"
)

// Takes an int64 filesize in bytes and returns a human-readable string.
func BytesToHuman(bytes int64) string {
	if bytes < 0 {
		return "Invalid size" // Or handle as an error
	}
	if bytes == 0 {
		return "0 B"
	}

	const (
		kb = 1024
		mb = kb * 1024
		gb = mb * 1024
		tb = gb * 1024
	)

	units := []string{"B", "KB", "MB", "GB", "TB"}
	thresholds := []int64{1, kb, mb, gb, tb} // The 1 is for "Bytes"

	// Iterate backwards
	var i int
	for i = len(thresholds) - 1; i >= 0; i-- {
		if bytes >= thresholds[i] {
			break
		}
	}

	// Calculate the value and format
	val := float64(bytes) / float64(thresholds[i])
	return fmt.Sprintf("%.2f %s", val, units[i])
}
