package ui

import (
	"github.com/gdamore/tcell/v2"
	"github.com/rivo/tview"
)

// formWidgetBase holds the fields and simple methods shared between Arguments and Properties.
// Embed it in any struct that implements the form widget pattern (grid + label + row management).
type formWidgetBase struct {
	*tview.Box

	grid *tview.Grid

	label      string
	labelWidth int
	labelColor tcell.Color
	bgColor    tcell.Color

	fieldTextColor tcell.Color
	fieldBgColor   tcell.Color
	fieldStyle     tcell.Style

	finished func(tcell.Key)
	disabled bool
	focus    func(p tview.Primitive)

	rowsChanged func(height int)

	listStyleSet        bool
	listStyleUnselected tcell.Style
	listStyleSelected   tcell.Style

	invalidStyleSet   bool
	invalidFieldStyle tcell.Style
}

// GetLabel returns the item's label text.
func (b *formWidgetBase) GetLabel() string {
	return b.label
}

// GetFieldWidth returns the width of the form item's field area.
func (b *formWidgetBase) GetFieldWidth() int {
	return 0
}

// emitRowsChanged calls the rowsChanged handler, if set.
func (b *formWidgetBase) emitRowsChanged(height int) {
	if b.rowsChanged != nil {
		b.rowsChanged(height)
	}
}

// applyFormAttributes sets the common form styling fields and updates the background color.
// The caller is responsible for calling applyStyles() afterwards.
func (b *formWidgetBase) applyFormAttributes(labelWidth int, labelColor, bgColor, fieldTextColor, fieldBgColor tcell.Color) {
	b.labelWidth = labelWidth
	b.labelColor = labelColor
	b.bgColor = bgColor
	b.fieldTextColor = fieldTextColor
	b.fieldBgColor = fieldBgColor
	b.fieldStyle = tcell.StyleDefault.Foreground(fieldTextColor).Background(fieldBgColor)

	if !b.invalidStyleSet {
		b.invalidFieldStyle = b.fieldStyle
	}

	b.SetBackgroundColor(bgColor)
}

// setInvalidStyle stores the invalid-field style.
// The caller is responsible for calling applyStyles() afterwards.
func (b *formWidgetBase) setInvalidStyle(style tcell.Style) {
	b.invalidStyleSet = true
	b.invalidFieldStyle = style
}

// drawFormWidget implements the common Draw logic for grid-based form widgets.
// It renders the label on the left and delegates the rest of the area to the grid.
func drawFormWidget(screen tcell.Screen, subclass tview.Primitive, box *tview.Box, grid *tview.Grid, label string, labelWidth int, labelColor tcell.Color) {
	box.DrawForSubclass(screen, subclass)

	x, y, width, height := box.GetInnerRect()
	if labelWidth > width {
		labelWidth = width
	}
	if height > 0 && labelWidth > 0 {
		tview.Print(screen, label, x, y, labelWidth, tview.AlignLeft, labelColor)
	}

	fieldX := x + labelWidth
	fieldWidth := width - labelWidth
	if fieldWidth < 0 {
		fieldWidth = 0
	}

	grid.SetRect(fieldX, y, fieldWidth, height)
	grid.Draw(screen)
}

// hasFormWidgetFocus implements the common HasFocus logic for grid-based form widgets.
func hasFormWidgetFocus(focusables []tview.Primitive, grid *tview.Grid, box *tview.Box) bool {
	for _, item := range focusables {
		if item.HasFocus() {
			return true
		}
	}

	if grid != nil {
		return grid.HasFocus()
	}

	return box.HasFocus()
}

// dispatchDoneKey implements the common handleDone logic for form widgets:
// moves focus forward/backward on Tab/Enter/Backtab, or calls finished otherwise.
func dispatchDoneKey(key tcell.Key, disabled bool, finished func(tcell.Key), focusRelative func(int) bool) {
	if disabled {
		if finished != nil {
			finished(key)
		}

		return
	}

	switch key {
	case tcell.KeyTab, tcell.KeyEnter:
		if focusRelative(1) {
			return
		}
	case tcell.KeyBacktab:
		if focusRelative(-1) {
			return
		}
	default:
	}

	if finished != nil {
		finished(key)
	}
}

// formWidgetMouseHandler implements the common MouseHandler for grid-based form widgets.
func formWidgetMouseHandler(
	box *tview.Box,
	grid *tview.Grid,
	focus *func(tview.Primitive),
) func(tview.MouseAction, *tcell.EventMouse, func(tview.Primitive)) (bool, tview.Primitive) {
	return box.WrapMouseHandler(func(action tview.MouseAction, event *tcell.EventMouse, setFocus func(p tview.Primitive)) (consumed bool, capture tview.Primitive) {
		if !box.InRect(event.Position()) {
			return false, nil
		}

		*focus = setFocus

		if grid != nil {
			consumed, capture = grid.MouseHandler()(action, event, setFocus)
			if consumed {
				return true, capture
			}
		}

		return
	})
}

// formWidgetInputHandler implements the common InputHandler for grid-based form widgets.
func formWidgetInputHandler(
	box *tview.Box,
	grid *tview.Grid,
	focus *func(tview.Primitive),
	focusables func() []tview.Primitive,
) func(*tcell.EventKey, func(tview.Primitive)) {
	return box.WrapInputHandler(func(event *tcell.EventKey, setFocus func(p tview.Primitive)) {
		*focus = setFocus

		if handler := inputHandlerOf(focusables()); handler != nil {
			handler(event, setFocus)
			return
		}

		if grid == nil {
			return
		}

		if handler := grid.InputHandler(); handler != nil {
			handler(event, setFocus)
		}
	})
}

// formWidgetPasteHandler implements the common PasteHandler for grid-based form widgets.
func formWidgetPasteHandler(
	box *tview.Box,
	grid *tview.Grid,
	focusables func() []tview.Primitive,
) func(string, func(tview.Primitive)) {
	return box.WrapPasteHandler(func(pastedText string, setFocus func(p tview.Primitive)) {
		if handler := pasteHandlerOf(focusables()); handler != nil {
			handler(pastedText, setFocus)
			return
		}

		if grid == nil {
			return
		}

		if handler := grid.PasteHandler(); handler != nil {
			handler(pastedText, setFocus)
		}
	})
}

// focusedIn returns the first focused primitive in items, or nil.
func focusedIn(items []tview.Primitive) tview.Primitive {
	for _, item := range items {
		if item.HasFocus() {
			return item
		}
	}

	return nil
}

// focusedIndexIn returns the index of the first focused primitive in items, or -1.
func focusedIndexIn(items []tview.Primitive) int {
	for i, item := range items {
		if item.HasFocus() {
			return i
		}
	}

	return -1
}

// containsIn reports whether p is present in items.
func containsIn(items []tview.Primitive, p tview.Primitive) bool {
	for _, item := range items {
		if item == p {
			return true
		}
	}

	return false
}

// inputHandlerOf returns the InputHandler of the focused primitive in items, or nil.
func inputHandlerOf(items []tview.Primitive) func(*tcell.EventKey, func(tview.Primitive)) {
	for _, item := range items {
		if item.HasFocus() {
			return item.InputHandler()
		}
	}

	return nil
}

// pasteHandlerOf returns the PasteHandler of the focused primitive in items, or nil.
func pasteHandlerOf(items []tview.Primitive) func(string, func(tview.Primitive)) {
	for _, item := range items {
		if item.HasFocus() {
			return item.PasteHandler()
		}
	}

	return nil
}

// focusRelativeIn moves focus by delta steps within items.
// Returns true if focus was moved, false if it hits a boundary or focus is nil.
func focusRelativeIn(items []tview.Primitive, delta int, setFocus func(tview.Primitive)) bool {
	if setFocus == nil || len(items) == 0 {
		return false
	}

	index := focusedIndexIn(items)
	if index == -1 {
		return false
	}

	next := index + delta
	if next < 0 || next >= len(items) {
		return false
	}

	setFocus(items[next])

	return true
}

// restoreFocusAfterRemoval sets focus to the element at prevIndex in newFocusables,
// clamping to the last item if the index is out of range.
func restoreFocusAfterRemoval(prevIndex int, newFocusables []tview.Primitive, setFocus func(tview.Primitive)) {
	if prevIndex < 0 || setFocus == nil || len(newFocusables) == 0 {
		return
	}

	if prevIndex >= len(newFocusables) {
		prevIndex = len(newFocusables) - 1
	}

	setFocus(newFocusables[prevIndex])
}

// focusFormWidget implements the common Focus logic for row-based form widgets.
// The caller must set its own focus field to delegate before calling this.
func focusFormWidget(
	delegate func(tview.Primitive),
	finished func(tcell.Key),
	disabled bool,
	lastValueEditor func() tview.Primitive,
	focusables func() []tview.Primitive,
	grid *tview.Grid,
	box *tview.Box,
) {
	if finished != nil && disabled {
		finished(-1)
		return
	}

	if ConsumeBacktab() {
		if last := lastValueEditor(); last != nil {
			delegate(last)
			return
		}
	}

	if items := focusables(); len(items) > 0 {
		delegate(items[0])
		return
	}

	if grid != nil {
		delegate(grid)
		return
	}

	box.Focus(delegate)
}

// handleRowChangeAt implements the trailing-empty-row management logic shared
// across row-based form widgets. Call it when a row's content changes.
// isEmpty indicates whether the row at rowIndex is currently empty.
// If it is the last row and not empty, appendEmpty is called.
// If it is not the last row and is empty, remove is called.
func handleRowChangeAt(rowIndex, totalRows int, isEmpty bool, appendEmpty func(), remove func(int)) {
	lastIndex := totalRows - 1

	if rowIndex == lastIndex {
		if !isEmpty {
			appendEmpty()
		}

		return
	}

	if isEmpty {
		remove(rowIndex)
	}
}
