package mstring

import (
	"encoding/json"
	"strings"

	"github.com/lib/pq"
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

func SliceContains(slice []string, str string) bool {
	for _, part := range slice {
		if part == str {
			return true
		}
	}
	return false
}

func Literal(literal string) string {
	return pq.QuoteLiteral(literal)
}

func FormatStringValues(value ...string) string {
	if l := len(value); l == 0 {
		return Literal("")
	} else if l == 1 {
		return Literal(value[0])
	}
	strs := []string{}
	for _, iter := range value {
		strs = append(strs, Literal(iter))
	}
	return strings.Join(strs, ", ")
}
