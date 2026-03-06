package view

import (
	"fmt"
	"strings"
	"tbunny/internal/utils"
)

// TextRow is a label/value pair for a text section.
type TextRow struct {
	Label string
	Value string
}

// WriteTextSection writes a caption, separator, and rows into b,
// aligning labels to the maximum label length using utils.PadRight.
func WriteTextSection(b *strings.Builder, caption string, rows []TextRow) {
	maxLen := 0
	for _, r := range rows {
		if len(r.Label) > maxLen {
			maxLen = len(r.Label)
		}
	}

	_, _ = fmt.Fprintf(b, "[caption]%s[-]\n", caption)
	b.WriteString(strings.Repeat("─", 30) + "\n")

	for _, r := range rows {
		_, _ = fmt.Fprintf(b, "[label]%s[-] [value]%s[-]\n", utils.PadRight(r.Label, maxLen), r.Value)
	}

	b.WriteString("\n")
}
