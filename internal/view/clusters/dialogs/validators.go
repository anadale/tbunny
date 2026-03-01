package dialogs

import (
	"net/url"
	"strings"
	"unicode"
	"unicode/utf8"
)

func validateUri(text string) bool {
	uri, err := url.ParseRequestURI(text)
	if err != nil {
		return false
	}
	if uri.Scheme != "http" && uri.Scheme != "https" {
		return false
	}
	if uri.Host == "" {
		return false
	}

	return true
}

// checkClusterNames checks that provided string is a valid cross-platform file name:
// - it's not a path (no / and \)
// - not "." and not ".."
// - doesn't start with '.' (hidden files like .env are forbidden)
// - Windows-safe: no <>:"/\|?*, NUL, control symbols
// - doesn't end with a space or dot (important for Windows)
// - not a reserved Windows word (CON, NUL, COM1...) — basename is checked until the first dot.
func validateClusterName(s string) bool {
	if s == "" {
		return false
	}
	if !utf8.ValidString(s) {
		return false
	}

	// Special names
	if s == "." || s == ".." {
		return false
	}

	// Doesn't start ith a dot
	if strings.HasPrefix(s, ".") {
		return false
	}

	// Not a path
	if strings.ContainsRune(s, '/') || strings.ContainsRune(s, '\\') {
		return false
	}

	// NUL is forbidden
	if strings.ContainsRune(s, '\x00') {
		return false
	}

	// Not a spaces
	if strings.TrimSpace(s) == "" {
		return false
	}

	// Windows: doesn't end with a space or dot
	if strings.HasSuffix(s, " ") || strings.HasSuffix(s, ".") {
		return false
	}

	// Windows-forbidden characters
	const bad = `<>:"/\|?*`
	if strings.ContainsAny(s, bad) {
		return false
	}

	// Control characters
	for _, r := range s {
		if r < 32 || r == 127 || unicode.IsControl(r) {
			return false
		}
	}

	// Reserved Windows names:
	base := s
	if i := strings.IndexByte(s, '.'); i >= 0 {
		base = s[:i]
	}
	up := strings.ToUpper(strings.TrimSpace(base))

	switch up {
	case "CON", "PRN", "AUX", "NUL":
		return false
	}
	if len(up) == 4 {
		prefix := up[:3]
		last := up[3]
		if (prefix == "COM" || prefix == "LPT") && last >= '1' && last <= '9' {
			return false
		}
	}

	return true
}
