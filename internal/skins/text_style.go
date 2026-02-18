package skins

type TextStyle string

const (
	// TextStyleNormal is the default text style.
	TextStyleNormal TextStyle = "normal"

	// TextStyleBold is the bold text style.
	TextStyleBold TextStyle = "bold"

	// TextStyleDim is the dim text style.
	TextStyleDim TextStyle = "dim"
)

// ToShortString returns a short string representation of the text style.
func (ts TextStyle) ToShortString() string {
	switch ts {
	case TextStyleNormal:
		return "-"
	case TextStyleBold:
		return "b"
	case TextStyleDim:
		return "d"
	default:
		return "d"
	}
}
