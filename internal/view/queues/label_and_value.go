package queues

import (
	"fmt"
	"tbunny/internal/view"

	"github.com/gdamore/tcell/v2"
	"github.com/rivo/tview"
)

type labelAndValue struct {
	label *tview.TextView
	value *tview.TextView
}

func createLabelAndValue(labelText string) *labelAndValue {
	label := tview.NewTextView().
		SetText(labelText + " ").
		SetTextColor(tcell.ColorWhite).
		SetTextAlign(tview.AlignRight).
		SetMaxLines(1)

	value := tview.NewTextView().
		SetTextColor(tcell.ColorWhite).
		SetMaxLines(1)

	lnv := labelAndValue{
		label: label,
		value: value,
	}

	return &lnv
}

func (lnv *labelAndValue) ApplyStyles(labelStyle, valueStyle tcell.Style) {
	lnv.label.SetTextStyle(labelStyle)
	lnv.value.SetTextStyle(valueStyle)
}

func (lnv *labelAndValue) SetCount(count int) {
	lnv.value.SetText(fmt.Sprintf("%d", count))
}

func (lnv *labelAndValue) SetRate(rate float32) {
	lnv.value.SetText(fmt.Sprintf("%.2f", rate))
}

func (lnv *labelAndValue) SetText(text string) {
	lnv.value.SetText(text)
}

func (lnv *labelAndValue) SetBool(b bool) {
	if b {
		lnv.value.SetText("true")
	} else {
		lnv.value.SetText("false")
	}
}

func (lnv *labelAndValue) SetBytes(bytes int64) {
	lnv.value.SetText(view.FormatBytes(bytes))
}

func (lnv *labelAndValue) SetNotApplicable() {
	lnv.value.SetText("N/A")
}

func (lnv *labelAndValue) AddToGrid(grid *tview.Grid, row, col int) {
	// minGridHeight = 1, minGridWidth = 0.
	grid.AddItem(lnv.label, row, col, 1, 1, 1, 0, false)
	grid.AddItem(lnv.value, row, col+1, 1, 1, 1, 0, false)
}
