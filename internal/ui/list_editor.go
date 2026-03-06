package ui

import (
	"slices"
	"strings"

	"github.com/gdamore/tcell/v2"
	"github.com/rivo/tview"
)

type listEditor struct {
	*tview.Box

	grid *tview.Grid

	rows []*listRow

	fieldTextColor tcell.Color
	fieldBgColor   tcell.Color
	fieldStyle     tcell.Style

	valuePlaceholder string

	disabled  bool
	focus     func(p tview.Primitive)
	finished  func(tcell.Key)
	lastFocus tview.Primitive
	navigate  func(current tview.Primitive, direction int) bool

	rowsChanged func(height int)
	changed     func()

	listStyleSet        bool
	listStyleUnselected tcell.Style
	listStyleSelected   tcell.Style

	invalidStyleSet   bool
	invalidFieldStyle tcell.Style
}

type listRow struct {
	value     *tview.InputField
	typeDrop  *tview.DropDown
	valueType valueType
	invalid   bool
}

type typedValue struct {
	valueType valueType
	text      string
}

func newListEditor() *listEditor {
	l := &listEditor{
		Box:              tview.NewBox(),
		grid:             tview.NewGrid(),
		valuePlaceholder: "Argument value",
	}

	l.grid.SetColumns(0, typeColumnWidth)

	return l
}

func (l *listEditor) SetFieldStyle(textColor, bgColor tcell.Color) {
	l.fieldTextColor = textColor
	l.fieldBgColor = bgColor
	l.fieldStyle = tcell.StyleDefault.Foreground(textColor).Background(bgColor)

	if !l.invalidStyleSet {
		l.invalidFieldStyle = l.fieldStyle
	}

	for _, row := range l.rows {
		l.applyRowStyles(row)
		row.typeDrop.SetFieldTextColor(textColor)
		row.typeDrop.SetFieldBackgroundColor(bgColor)
	}
}

func (l *listEditor) SetFieldPlaceholder(placeholder string) {
	l.valuePlaceholder = placeholder

	for _, row := range l.rows {
		row.value.SetPlaceholder(placeholder)
	}
}

func (l *listEditor) SetDisabled(disabled bool) {
	l.disabled = disabled

	for _, row := range l.rows {
		row.value.SetDisabled(disabled)
		row.typeDrop.SetDisabled(disabled)
	}
}

func (l *listEditor) SetFinishedFunc(handler func(key tcell.Key)) {
	l.finished = handler
}

func (l *listEditor) SetRowsChangedFunc(handler func(height int)) {
	l.rowsChanged = handler
}

func (l *listEditor) SetChangedFunc(handler func()) {
	l.changed = handler
}

func (l *listEditor) SetNavigateFunc(handler func(current tview.Primitive, direction int) bool) {
	l.navigate = handler
}

func (l *listEditor) SetListStyles(unselected, selected tcell.Style) {
	l.listStyleSet = true
	l.listStyleUnselected = unselected
	l.listStyleSelected = selected
	for _, row := range l.rows {
		row.typeDrop.SetListStyles(unselected, selected)
	}
}

func (l *listEditor) SetInvalidFieldStyle(style tcell.Style) {
	l.invalidStyleSet = true
	l.invalidFieldStyle = style

	for _, row := range l.rows {
		l.applyRowStyles(row)
	}
}

func (l *listEditor) SetValue(values []any) int {
	l.grid.Clear()
	l.rows = nil
	l.grid.SetColumns(0, typeColumnWidth)

	firstField := true
	for _, value := range values {
		row := l.newRow(firstField)
		l.setRowValue(row, value)
		firstField = false
	}

	l.ensureTrailingEmptyRow(firstField)
	l.rebuildGrid(false, nil)
	height := l.Height()
	l.emitRowsChanged(height)

	return height
}

func (l *listEditor) SetTypedValues(values []typedValue) int {
	l.grid.Clear()
	l.rows = nil
	l.grid.SetColumns(0, typeColumnWidth)

	firstField := true
	for _, value := range values {
		row := l.newRow(firstField)
		row.valueType = value.valueType
		row.typeDrop.SetCurrentOption(int(value.valueType))
		row.value.SetText(value.text)
		row.invalid = false
		l.applyRowStyles(row)
		firstField = false
	}

	l.ensureTrailingEmptyRow(firstField)
	l.rebuildGrid(false, nil)
	height := l.Height()
	l.emitRowsChanged(height)

	return height
}

func (l *listEditor) GetValue() []any {
	result := make([]any, 0, len(l.rows))

	for _, row := range l.rows {
		text := strings.TrimSpace(row.value.GetText())
		if text == "" {
			continue
		}

		value, ok := parseTypedValue(row.valueType, text)
		if !ok {
			continue
		}

		result = append(result, value)
	}

	return result
}

func (l *listEditor) Height() int {
	if len(l.rows) == 0 {
		return 1
	}

	return len(l.rows)
}

func (l *listEditor) IsEmpty() bool {
	for _, row := range l.rows {
		if strings.TrimSpace(row.value.GetText()) != "" {
			return false
		}
	}

	return true
}

func (l *listEditor) Draw(screen tcell.Screen) {
	l.DrawForSubclass(screen, l)
	x, y, width, height := l.GetInnerRect()
	l.grid.SetRect(x, y, width, height)
	l.grid.Draw(screen)
}

func (l *listEditor) Focus(delegate func(p tview.Primitive)) {
	if l.finished != nil && l.disabled {
		l.finished(-1)
		return
	}

	l.focus = delegate
	if l.lastFocus != nil && containsIn(l.focusables(), l.lastFocus) {
		delegate(l.lastFocus)
		return
	}

	if focusables := l.focusables(); len(focusables) > 0 {
		l.lastFocus = focusables[0]
		delegate(focusables[0])
		return
	}

	if l.grid != nil {
		delegate(l.grid)
		return
	}

	l.Box.Focus(delegate)
}

func (l *listEditor) HasFocus() bool {
	for _, item := range l.focusables() {
		if item.HasFocus() {
			return true
		}
	}

	if l.grid != nil {
		return l.grid.HasFocus()
	}

	return l.Box.HasFocus()
}

func (l *listEditor) MouseHandler() func(action tview.MouseAction, event *tcell.EventMouse, setFocus func(p tview.Primitive)) (consumed bool, capture tview.Primitive) {
	return l.WrapMouseHandler(func(action tview.MouseAction, event *tcell.EventMouse, setFocus func(p tview.Primitive)) (consumed bool, capture tview.Primitive) {
		if !l.InRect(event.Position()) {
			return false, nil
		}

		l.focus = setFocus

		if l.grid != nil {
			consumed, capture = l.grid.MouseHandler()(action, event, setFocus)
			if consumed {
				if capture != nil && containsIn(l.focusables(), capture) {
					l.lastFocus = capture
				} else if focused := focusedIn(l.focusables()); focused != nil {
					l.lastFocus = focused
				}
				return true, capture
			}
		}

		return
	})
}

func (l *listEditor) InputHandler() func(event *tcell.EventKey, setFocus func(p tview.Primitive)) {
	return l.WrapInputHandler(func(event *tcell.EventKey, setFocus func(p tview.Primitive)) {
		l.focus = setFocus

		if handler := l.focusedPrimitiveHandler(); handler != nil {
			handler(event, setFocus)
			return
		}

		if l.grid == nil {
			return
		}

		if handler := l.grid.InputHandler(); handler != nil {
			handler(event, setFocus)
			return
		}
	})
}

func (l *listEditor) PasteHandler() func(pastedText string, setFocus func(p tview.Primitive)) {
	return l.WrapPasteHandler(func(pastedText string, setFocus func(p tview.Primitive)) {
		if handler := l.focusedPrimitivePasteHandler(); handler != nil {
			handler(pastedText, setFocus)
			return
		}

		if l.grid == nil {
			return
		}

		if handler := l.grid.PasteHandler(); handler != nil {
			handler(pastedText, setFocus)
			return
		}
	})
}

func (l *listEditor) newRow(focus bool) *listRow {
	row := &listRow{
		value:     tview.NewInputField().SetPlaceholder(l.valuePlaceholder),
		typeDrop:  tview.NewDropDown().SetOptions(listValueTypeLabels, nil),
		valueType: valueTypeString,
	}

	row.typeDrop.SetCurrentOption(0)

	if l.listStyleSet {
		row.typeDrop.SetListStyles(l.listStyleUnselected, l.listStyleSelected)
	}

	row.value.SetDisabled(l.disabled)
	row.typeDrop.SetDisabled(l.disabled)
	l.applyRowStyles(row)
	row.typeDrop.SetFieldTextColor(l.fieldTextColor)
	row.typeDrop.SetFieldBackgroundColor(l.fieldBgColor)
	l.rows = append(l.rows, row)

	if focus && l.focus != nil {
		l.lastFocus = row.value
		l.focus(row.value)
	}

	return row
}

func (l *listEditor) ensureTrailingEmptyRow(focus bool) {
	if len(l.rows) == 0 {
		l.newRow(focus)
		return
	}

	last := l.rows[len(l.rows)-1]

	if strings.TrimSpace(last.value.GetText()) != "" {
		l.newRow(focus)
	}
}

func (l *listEditor) setRowValue(row *listRow, value any) {
	if value == nil {
		return
	}

	text, vt := formatValueText(value)
	row.value.SetText(text)
	row.valueType = vt
	row.typeDrop.SetCurrentOption(int(vt))
	row.invalid = false

	l.applyRowStyles(row)
}

func (l *listEditor) rebuildGrid(preserveFocus bool, preferred tview.Primitive) {
	var focused tview.Primitive

	if preserveFocus {
		if preferred != nil && containsIn(l.focusables(), preferred) {
			focused = preferred
		} else {
			focused = focusedIn(l.focusables())
		}
	}

	l.grid.Clear()
	l.grid.SetColumns(0, typeColumnWidth)

	rows := make([]int, 0, len(l.rows))

	for rowIndex, row := range l.rows {
		rows = append(rows, 1)
		l.grid.AddItem(row.value, rowIndex, 0, 1, 1, 0, 0, rowIndex == 0)
		l.grid.AddItem(row.typeDrop, rowIndex, 1, 1, 1, 0, 0, false)
	}

	if len(rows) == 0 {
		rows = append(rows, 1)
	}

	l.grid.SetRows(rows...)
	l.setRowHandlers()

	if focused != nil && l.focus != nil && containsIn(l.focusables(), focused) {
		l.lastFocus = focused
		l.focus(focused)
	}
}

func (l *listEditor) setRowHandlers() {
	for _, row := range l.rows {
		current := row

		row.value.SetDoneFunc(func(key tcell.Key) {
			l.handleDone(key)
		})
		row.value.SetChangedFunc(func(text string) {
			if current.invalid {
				current.invalid = false
				l.applyRowStyles(current)
			}
			index := l.indexOfRow(current)
			if index >= 0 {
				l.handleRowChange(index)
			}
		})

		row.typeDrop.SetDoneFunc(func(key tcell.Key) {
			l.handleDone(key)
		})
		row.typeDrop.SetSelectedFunc(func(text string, index int) {
			newType := valueType(index)
			oldText := current.value.GetText()
			valueKept := canConvertText(newType, oldText)
			current.valueType = newType
			if !valueKept {
				current.value.SetText("")
				current.invalid = strings.TrimSpace(oldText) != ""
				l.applyRowStyles(current)
			}
			if l.focus != nil {
				if valueKept {
					l.lastFocus = current.typeDrop
					l.focus(current.typeDrop)
				} else {
					l.lastFocus = current.value
					l.focus(current.value)
				}
			}
			if l.changed != nil {
				l.changed()
			}
		})
	}
}

func (l *listEditor) handleDone(key tcell.Key) {
	if l.disabled {
		if l.finished != nil {
			l.finished(key)
		}
		return
	}

	switch key {
	case tcell.KeyTab, tcell.KeyEnter:
		if l.navigate != nil && l.navigate(focusedIn(l.focusables()), 1) {
			return
		}
		if l.focusRelative(1) {
			return
		}
	case tcell.KeyBacktab:
		if l.navigate != nil && l.navigate(focusedIn(l.focusables()), -1) {
			return
		}
		if l.focusRelative(-1) {
			return
		}
	default:
	}

	if l.finished != nil {
		l.finished(key)
	}
}

func (l *listEditor) handleRowChange(rowIndex int) {
	if rowIndex < 0 || rowIndex >= len(l.rows) {
		return
	}

	row := l.rows[rowIndex]
	isEmpty := strings.TrimSpace(row.value.GetText()) == ""
	handleRowChangeAt(rowIndex, len(l.rows), isEmpty,
		func() { l.appendEmptyRow(row.value) },
		l.removeRow,
	)

	if l.changed != nil {
		l.changed()
	}
}

func (l *listEditor) appendEmptyRow(preferred tview.Primitive) {
	l.newRow(false)
	l.rebuildGrid(true, preferred)
	l.emitRowsChanged(l.Height())
}

func (l *listEditor) removeRow(rowIndex int) {
	if rowIndex < 0 || rowIndex >= len(l.rows)-1 {
		return
	}

	prevFocusedIndex := focusedIndexIn(l.focusables())
	l.rows = slices.Delete(l.rows, rowIndex, rowIndex+1)
	l.rebuildGrid(false, nil)
	l.emitRowsChanged(l.Height())
	restoreFocusAfterRemoval(prevFocusedIndex, l.focusables(), l.focus)
}

func (l *listEditor) focusRelative(delta int) bool {
	if l.focus == nil {
		return false
	}

	focusables := l.focusables()
	if len(focusables) == 0 {
		return false
	}

	index := focusedIndexIn(focusables)
	if index == -1 {
		return false
	}

	next := index + delta
	if next < 0 || next >= len(focusables) {
		return false
	}

	l.lastFocus = focusables[next]
	l.focus(focusables[next])

	return true
}

func (l *listEditor) focusedPrimitiveHandler() func(event *tcell.EventKey, setFocus func(p tview.Primitive)) {
	for _, item := range l.focusables() {
		if item.HasFocus() {
			l.lastFocus = item
			return item.InputHandler()
		}
	}

	return nil
}

func (l *listEditor) focusedPrimitivePasteHandler() func(pastedText string, setFocus func(p tview.Primitive)) {
	for _, item := range l.focusables() {
		if item.HasFocus() {
			l.lastFocus = item
			return item.PasteHandler()
		}
	}

	return nil
}

func (l *listEditor) focusables() []tview.Primitive {
	entries := make([]tview.Primitive, 0, len(l.rows)*2)

	for _, row := range l.rows {
		entries = append(entries, row.value, row.typeDrop)
	}

	return entries
}

// focusablesWithExternalDrop returns the focusable primitives of the list,
// with externalDrop inserted after the first row's fields. This produces the
// expected Tab-order when the list is embedded in a parent row that has its
// own type drop-down sitting between the list's first and remaining items.
func (l *listEditor) focusablesWithExternalDrop(externalDrop tview.Primitive) []tview.Primitive {
	if len(l.rows) == 0 {
		return []tview.Primitive{externalDrop}
	}

	entries := make([]tview.Primitive, 0, len(l.rows)*2+1)

	first := l.rows[0]
	if first.value != nil {
		entries = append(entries, first.value)
	}
	if first.typeDrop != nil {
		entries = append(entries, first.typeDrop)
	}

	entries = append(entries, externalDrop)

	for _, item := range l.rows[1:] {
		if item.value != nil {
			entries = append(entries, item.value)
		}
		if item.typeDrop != nil {
			entries = append(entries, item.typeDrop)
		}
	}

	return entries
}

func (l *listEditor) emitRowsChanged(height int) {
	if l.rowsChanged != nil {
		l.rowsChanged(height)
	}
}

func (l *listEditor) indexOfRow(row *listRow) int {
	for i, current := range l.rows {
		if current == row {
			return i
		}
	}

	return -1
}

func (l *listEditor) applyRowStyles(row *listRow) {
	if row.invalid {
		row.value.SetFieldStyle(l.invalidFieldStyle)
	} else {
		row.value.SetFieldStyle(l.fieldStyle)
	}
}

func (l *listEditor) lastValueEditor() tview.Primitive {
	for i := len(l.rows) - 1; i >= 0; i-- {
		if l.rows[i].value != nil {
			return l.rows[i].value
		}
	}

	return nil
}

func (l *listEditor) firstEmptyValueEditor() tview.Primitive {
	for _, row := range l.rows {
		if row.value == nil {
			continue
		}
		if strings.TrimSpace(row.value.GetText()) == "" {
			return row.value
		}
	}

	return nil
}

func (l *listEditor) singleTypedValue() (typedValue, bool) {
	var result typedValue

	found := false
	for _, row := range l.rows {
		text := strings.TrimSpace(row.value.GetText())

		if text == "" {
			continue
		}

		if found {
			return typedValue{}, false
		}

		result = typedValue{valueType: row.valueType, text: text}
		found = true
	}

	return result, found
}
