package utils

import (
	"strings"
)

func TruncateStr(s string, targetLen int) string {
	runes := []rune(s)
	paddingNeeded := targetLen - len(runes)

	if len(runes) > targetLen {
		shortenBy := max(len(runes)-targetLen, 3)
		runes = append([]rune("..."), runes[shortenBy+1:]...)
		if len(runes) == targetLen {
			return string(runes)
		}
		paddingNeeded = len(runes) - targetLen
	}

	// runes can change in if, so can't assign len to variable.
	return string(runes) + strings.Repeat(" ", paddingNeeded)
}
