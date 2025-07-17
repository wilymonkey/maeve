package utils

import (
	"github.com/charmbracelet/lipgloss"
)

func TruncateStr(s string, width int) string {
	sWidth := lipgloss.Width(s)
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
