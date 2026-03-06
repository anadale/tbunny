package utils

import (
	"fmt"
	"strings"
)

// Sbprintf writes a formatted string to the provided strings.Builder using the specified format and arguments.
func Sbprintf(b *strings.Builder, format string, args ...any) {
	_, _ = fmt.Fprintf(b, format, args...)
}
