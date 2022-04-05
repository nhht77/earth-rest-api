package mstring

import (
	"encoding/json"
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

func ToJSON(obj interface{}) string {
	if obj == nil {
		return ""
	}
	b, err := json.MarshalIndent(obj, "", "  ")
	if err == nil && len(b) > 0 {
		return string(b)
	}
	return ""
}

func FormatFields(str ...string) string {
	return strings.Join(str, ", ")
}
