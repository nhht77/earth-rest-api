package mstring

import (
	"strings"
)

func Between(value string, begin string, end string) string {
	// Get substring between two strings.
	posFirst := strings.Index(value, begin)
	if posFirst == -1 {
		return ""
	}
	posLast := strings.Index(value, end)
	if posLast == -1 {
		return ""
	}
	posFirstAdjusted := posFirst + len(begin)
	if posFirstAdjusted >= posLast {
		return ""
	}
	return value[posFirstAdjusted:posLast]
}
