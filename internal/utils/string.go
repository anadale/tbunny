package utils

import "strings"

func PadRight(s string, n int) string {
	if n <= len(s) {
		return s
	}

	return s + strings.Repeat(" ", n-len(s))
}

func PadLeft(s string, n int) string {
	if n <= len(s) {
		return s
	}

	return strings.Repeat(" ", n-len(s)) + s
}
