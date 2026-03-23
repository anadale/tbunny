package view

import (
	"fmt"

	"github.com/gdamore/tcell/v2"
	"github.com/rivo/tview"
)

// LabelAndValue is a reusable widget that displays a label and a value side by side.
type LabelAndValue struct {
	Label *tview.TextView
	Value *tview.TextView
}

// NewLabelAndValue creates a new LabelAndValue widget with the given label text.
func NewLabelAndValue(labelText string) *LabelAndValue {
	label := tview.NewTextView().
		SetText(labelText + " ").
		SetTextColor(tcell.ColorWhite).
		SetTextAlign(tview.AlignRight).
		SetMaxLines(1)

	value := tview.NewTextView().
		SetTextColor(tcell.ColorWhite).
		SetMaxLines(1)

	return &LabelAndValue{
		Label: label,
		Value: value,
	}
}

// ApplyStyles sets the text style for both the label and value.
func (lnv *LabelAndValue) ApplyStyles(labelStyle, valueStyle tcell.Style) {
	lnv.Label.SetTextStyle(labelStyle)
	lnv.Value.SetTextStyle(valueStyle)
}

// SetCount sets the value to a formatted integer.
func (lnv *LabelAndValue) SetCount(count int) {
	lnv.Value.SetText(fmt.Sprintf("%d", count))
}

// SetRate sets the value to a formatted float with 2 decimal places.
func (lnv *LabelAndValue) SetRate(rate float32) {
	lnv.Value.SetText(fmt.Sprintf("%.2f", rate))
}

// SetText sets the value to the given text.
func (lnv *LabelAndValue) SetText(text string) {
	lnv.Value.SetText(text)
}

// SetBool sets the value to "true" or "false".
func (lnv *LabelAndValue) SetBool(b bool) {
	if b {
		lnv.Value.SetText("true")
	} else {
		lnv.Value.SetText("false")
	}
}

// SetBytes sets the value to a human-readable byte format.
func (lnv *LabelAndValue) SetBytes(bytes int64) {
	lnv.Value.SetText(FormatBytes(bytes))
}

// SetNotApplicable sets the value to "N/A".
func (lnv *LabelAndValue) SetNotApplicable() {
	lnv.Value.SetText("N/A")
}

// AddToGrid adds the label and value to a tview.Grid at the specified row and column.
func (lnv *LabelAndValue) AddToGrid(grid *tview.Grid, row, col int) {
	// minGridHeight = 1, minGridWidth = 0.
	grid.AddItem(lnv.Label, row, col, 1, 1, 1, 0, false)
	grid.AddItem(lnv.Value, row, col+1, 1, 1, 1, 0, false)
}
