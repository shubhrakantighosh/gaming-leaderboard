package util

import "strings"

func TrimSpace(str string) string {
	return strings.TrimSpace(str)
}

func TrimStrings(strings []string) []string {
	result := make([]string, 0, len(strings))
	for _, s := range strings {
		result = append(result, TrimSpace(s))
	}

	return result
}
