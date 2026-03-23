package view

import (
	"fmt"
	"strings"
	"tbunny/internal/skins"
)

// ProgressBarStyle defines the visual appearance of a progress bar
type ProgressBarStyle struct {
	EmptyRune  rune
	FilledRune rune
	Prefix     string
	Suffix     string
}

// DefaultProgressBarStyle provides a default style for progress bars
var DefaultProgressBarStyle = ProgressBarStyle{
	EmptyRune:  '░',
	FilledRune: '█',
	Prefix:     "",
	Suffix:     "",
}

// RenderProgressBar creates a visual progress bar representation
func RenderProgressBar(percentage float64, width int, style ProgressBarStyle) string {
	if percentage < 0 {
		percentage = 0
	}
	if percentage > 100 {
		percentage = 100
	}

	// Calculate filled width
	filledWidth := int((percentage / 100) * float64(width))

	var builder strings.Builder

	// Write prefix
	if style.Prefix != "" {
		builder.WriteString(style.Prefix)
	}

	// Write filled portion
	for i := 0; i < filledWidth; i++ {
		builder.WriteRune(style.FilledRune)
	}

	// Write empty portion
	for i := filledWidth; i < width; i++ {
		builder.WriteRune(style.EmptyRune)
	}

	// Write suffix
	if style.Suffix != "" {
		builder.WriteString(style.Suffix)
	}

	// Add percentage text
	builder.WriteString(" ")
	fmt.Fprintf(&builder, "%.0f%%", percentage)

	return builder.String()
}

// GetStateColor returns the appropriate color tag based on percentage and thresholds
func GetStateColor(percentage float64, skin *skins.Skin) string {
	s := skin.Views.Stats
	if percentage >= 95 {
		return fmt.Sprintf("[%s]", s.CriticalStateColor)
	} else if percentage >= 80 {
		return fmt.Sprintf("[%s]", s.WarningStateColor)
	} else if percentage >= 0 {
		return fmt.Sprintf("[%s]", s.NormalStateColor)
	}
	return fmt.Sprintf("[%s]", s.InfoStateColor)
}
