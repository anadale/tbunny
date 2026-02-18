package skins

import (
	"fmt"

	"github.com/gdamore/tcell/v2"
)

const (
	// DefaultColor represents the default color.
	DefaultColor Color = "default"

	// TransparentColor represents the terminal background color.
	TransparentColor Color = "-"
)

type Color string

func NewColor(c string) Color { return Color(c) }

func (c Color) String() string {
	if c.isHex() {
		return string(c)
	}

	if c == DefaultColor {
		return "-"
	}

	col := c.Color().TrueColor().Hex()
	if col < 0 {
		return "-"
	}

	return fmt.Sprintf("#%06x", col)
}

func (c Color) Color() tcell.Color {
	if c == DefaultColor {
		return tcell.ColorDefault
	}

	return tcell.GetColor(string(c)).TrueColor()
}

func (c Color) isHex() bool {
	return len(c) == 7 && c[0] == '#'
}
