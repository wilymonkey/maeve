package utils

import (
	"strings"
)

func TruncateStr(s string, targetLen int) string {
	runes := []rune(s)

	if len(runes) > targetLen {
		shortenBy := max(len(runes)-targetLen, 3)
		runes = append([]rune("..."), runes[shortenBy:]...)
		if len(runes) == targetLen {
			return string(runes)
		}
	}
	// runes can change in if, so can't assign len to variable.
	paddingNeeded := len(runes) - targetLen
	return s + strings.Repeat(" ", paddingNeeded)
}
