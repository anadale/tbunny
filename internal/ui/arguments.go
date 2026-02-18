package ui

import (
	"fmt"
	"reflect"
	"strconv"
	"strings"

	"github.com/gdamore/tcell/v2"
	"github.com/rivo/tview"
)

const typeColumnWidth = 12
const valueExtraColumnWidth = 12

type valueType int

const (
	valueTypeString valueType = iota
	valueTypeNumber
	valueTypeBoolean
	valueTypeList
)

var valueTypeLabels = []string{"String", "Number", "Boolean", "List"}
var listValueTypeLabels = []string{"String", "Number", "Boolean"}

type Arguments struct {
	*tview.Box

	grid *tview.Grid

	label      string
	labelWidth int
	labelColor tcell.Color
	bgColor    tcell.Color

	fieldTextColor tcell.Color
	fieldBgColor   tcell.Color
	fieldStyle     tcell.Style

	keyPlaceholder   string
	valuePlaceholder string

	rows     []*argRow
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

type argRow struct {
	key       *tview.InputField
	value     *tview.InputField
	list      *listEditor
	typeDrop  *tview.DropDown
	valueType valueType
	invalid   bool
}

func NewArguments() *Arguments {
	a := Arguments{
		Box:              tview.NewBox(),
		grid:             tview.NewGrid(),
		keyPlaceholder:   "Argument name",
		valuePlaceholder: "Argument value",
	}

	a.SetValue(nil)

	return &a
}

func (a *Arguments) SetLabel(label string) *Arguments {
	a.label = label
	return a
}

func (a *Arguments) SetValue(args map[string]any) int {
	a.grid.Clear()
	a.rows = nil
	a.grid.SetColumns(-11, -9, valueExtraColumnWidth, typeColumnWidth)

	firstField := true
	for k, v := range args {
		row := a.newRow(firstField)
		row.key.SetText(k)
		a.setRowValue(row, v)
		firstField = false
	}

	a.ensureTrailingEmptyRow(firstField)
	a.rebuildGrid(false, nil)

	height := a.totalHeight()
	a.emitRowsChanged(height)

	return height
}

func (a *Arguments) GetValue() map[string]any {
	result := make(map[string]any)

	for _, row := range a.rows {
		key := strings.TrimSpace(row.key.GetText())
		if key == "" {
			continue
		}

		if row.valueType == valueTypeList {
			if row.list == nil {
				continue
			}

			values := row.list.GetValue()
			if len(values) == 0 {
				continue
			}

			result[key] = values
			continue
		}

		if row.value == nil {
			continue
		}

		text := strings.TrimSpace(row.value.GetText())
		if text == "" {
			continue
		}

		if value, ok := parseTypedValue(row.valueType, text); ok {
			result[key] = value
		}
	}

	return result
}

func (a *Arguments) SetKeyPlaceholder(placeholder string) *Arguments {
	a.keyPlaceholder = placeholder

	for _, row := range a.rows {
		row.key.SetPlaceholder(placeholder)
	}

	return a
}
func (a *Arguments) SetValuePlaceholder(placeholder string) *Arguments {
	a.valuePlaceholder = placeholder

	for _, row := range a.rows {
		if row.valueType == valueTypeList {
			row.list.SetFieldPlaceholder(placeholder)
		} else {
			row.value.SetPlaceholder(placeholder)
		}
	}

	return a
}

func (a *Arguments) Draw(screen tcell.Screen) {
	a.Box.DrawForSubclass(screen, a)

	x, y, width, height := a.GetInnerRect()
	labelWidth := a.labelWidth
	if labelWidth > width {
		labelWidth = width
	}
	if height > 0 && labelWidth > 0 {
		tview.Print(screen, a.label, x, y, labelWidth, tview.AlignLeft, a.labelColor)
	}

	fieldX := x + labelWidth
	fieldWidth := width - labelWidth
	if fieldWidth < 0 {
		fieldWidth = 0
	}

	a.grid.SetRect(fieldX, y, fieldWidth, height)
	a.grid.Draw(screen)
}

// GetLabel returns the item's label text.
func (a *Arguments) GetLabel() string {
	return a.label
}

// SetFormAttributes sets a number of item attributes at once.
func (a *Arguments) SetFormAttributes(labelWidth int, labelColor, bgColor, fieldTextColor, fieldBgColor tcell.Color) tview.FormItem {
	a.labelWidth = labelWidth
	a.labelColor = labelColor
	a.bgColor = bgColor
	a.fieldTextColor = fieldTextColor
	a.fieldBgColor = fieldBgColor
	a.fieldStyle = tcell.StyleDefault.Foreground(fieldTextColor).Background(fieldBgColor)
	if !a.invalidStyleSet {
		a.invalidFieldStyle = a.fieldStyle
	}

	a.SetBackgroundColor(bgColor)
	a.applyStyles()

	return a
}

// GetFieldWidth returns the width of the form item's field area.
func (a *Arguments) GetFieldWidth() int {
	return 0
}

// GetFieldHeight returns the height of the form item's field area.
func (a *Arguments) GetFieldHeight() int {
	height := a.totalHeight()
	if height > 0 {
		return height
	}

	return 1
}

// SetFinishedFunc sets the handler function for when the users finished entering data.
func (a *Arguments) SetFinishedFunc(handler func(key tcell.Key)) tview.FormItem {
	a.finished = handler

	return a
}

// SetDisabled sets whether the item is disabled / read-only.
func (a *Arguments) SetDisabled(disabled bool) tview.FormItem {
	a.disabled = disabled

	for _, row := range a.rows {
		row.key.SetDisabled(disabled)
		row.typeDrop.SetDisabled(disabled)

		if row.valueType == valueTypeList && row.list != nil {
			row.list.SetDisabled(disabled)
		} else if row.value != nil {
			row.value.SetDisabled(disabled)
		}
	}

	if a.finished != nil {
		a.finished(-1)
	}

	return a
}

// SetRowsChangedFunc sets a handler which is called when the number of rows changes.
func (a *Arguments) SetRowsChangedFunc(handler func(height int)) *Arguments {
	a.rowsChanged = handler

	return a
}

// SetInvalidFieldStyle sets the style used to indicate an invalid value.
func (a *Arguments) SetInvalidFieldStyle(style tcell.Style) *Arguments {
	a.invalidStyleSet = true
	a.invalidFieldStyle = style

	a.applyStyles()

	return a
}

// SetListStyles sets the styles for all drop-down lists, including nested list editors.
func (a *Arguments) SetListStyles(unselected, selected tcell.Style) *Arguments {
	a.listStyleSet = true
	a.listStyleUnselected = unselected
	a.listStyleSelected = selected

	for _, row := range a.rows {
		row.typeDrop.SetListStyles(unselected, selected)
		if row.list != nil {
			row.list.SetListStyles(unselected, selected)
		}
	}

	return a
}

// Focus is called when this primitive receives focus.
func (a *Arguments) Focus(delegate func(p tview.Primitive)) {
	if a.finished != nil && a.disabled {
		a.finished(-1)
		return
	}

	a.focus = delegate

	if ConsumeBacktab() {
		if lastValue := a.lastValueEditor(); lastValue != nil {
			delegate(lastValue)
			return
		}
	}

	if focusables := a.focusables(); len(focusables) > 0 {
		delegate(focusables[0])
		return
	}

	if a.grid != nil {
		delegate(a.grid)
		return
	}

	a.Box.Focus(delegate)
}

// HasFocus returns whether or not this primitive has focus.
func (a *Arguments) HasFocus() bool {
	for _, item := range a.focusables() {
		if item.HasFocus() {
			return true
		}
	}

	if a.grid != nil {
		return a.grid.HasFocus()
	}

	return a.Box.HasFocus()
}

// MouseHandler returns the mouse handler for this primitive.
func (a *Arguments) MouseHandler() func(action tview.MouseAction, event *tcell.EventMouse, setFocus func(p tview.Primitive)) (consumed bool, capture tview.Primitive) {
	return a.WrapMouseHandler(func(action tview.MouseAction, event *tcell.EventMouse, setFocus func(p tview.Primitive)) (consumed bool, capture tview.Primitive) {
		if !a.InRect(event.Position()) {
			return false, nil
		}

		a.focus = setFocus

		// Pass mouse events on to contained primitive.
		if a.grid != nil {
			consumed, capture = a.grid.MouseHandler()(action, event, setFocus)
			if consumed {
				return true, capture
			}
		}

		return
	})
}

// InputHandler returns the handler for this primitive.
func (a *Arguments) InputHandler() func(event *tcell.EventKey, setFocus func(p tview.Primitive)) {
	return a.WrapInputHandler(func(event *tcell.EventKey, setFocus func(p tview.Primitive)) {
		a.focus = setFocus

		if handler := a.focusedPrimitiveHandler(); handler != nil {
			handler(event, setFocus)
			return
		}

		if a.grid == nil {
			return
		}

		if handler := a.grid.InputHandler(); handler != nil {
			handler(event, setFocus)
			return
		}
	})
}

// PasteHandler returns the handler for this primitive.
func (a *Arguments) PasteHandler() func(pastedText string, setFocus func(p tview.Primitive)) {
	return a.WrapPasteHandler(func(pastedText string, setFocus func(p tview.Primitive)) {
		if handler := a.focusedPrimitivePasteHandler(); handler != nil {
			handler(pastedText, setFocus)
			return
		}

		if a.grid == nil {
			return
		}

		if handler := a.grid.PasteHandler(); handler != nil {
			handler(pastedText, setFocus)
			return
		}
	})
}

func (a *Arguments) newRow(focus bool) *argRow {
	row := &argRow{
		key:       tview.NewInputField().SetPlaceholder(a.keyPlaceholder),
		value:     tview.NewInputField().SetPlaceholder(a.valuePlaceholder),
		typeDrop:  tview.NewDropDown().SetOptions(valueTypeLabels, nil),
		valueType: valueTypeString,
	}

	row.typeDrop.SetCurrentOption(0)

	if a.listStyleSet {
		row.typeDrop.SetListStyles(a.listStyleUnselected, a.listStyleSelected)
	}

	row.key.SetDisabled(a.disabled)
	row.value.SetDisabled(a.disabled)
	row.typeDrop.SetDisabled(a.disabled)

	a.applyRowStyles(row)
	a.rows = append(a.rows, row)

	if focus && a.focus != nil {
		a.focus(row.key)
	}

	return row
}

func (a *Arguments) ensureTrailingEmptyRow(focus bool) {
	if len(a.rows) == 0 {
		a.newRow(focus)
		return
	}

	last := a.rows[len(a.rows)-1]
	if !a.isRowEmpty(last) {
		a.newRow(focus)
	}
}

func (a *Arguments) setRowValue(row *argRow, value any) {
	if value == nil {
		return
	}

	if items, ok := normalizeSlice(value); ok {
		a.setRowType(row, valueTypeList, true, false)

		if row.list == nil {
			row.list = newListEditor()
			a.applyListStyles(row.list)
		}

		row.list.SetValue(items)
		return
	}

	switch typed := value.(type) {
	case bool:
		row.value.SetText(strconv.FormatBool(typed))
		a.setRowType(row, valueTypeBoolean, true, false)
	case string:
		row.value.SetText(typed)
		a.setRowType(row, valueTypeString, true, false)
	case fmt.Stringer:
		row.value.SetText(typed.String())
		a.setRowType(row, valueTypeString, true, false)
	default:
		if numText, ok := formatNumber(value); ok {
			row.value.SetText(numText)
			a.setRowType(row, valueTypeNumber, true, false)
			return
		}
		row.value.SetText(fmt.Sprint(value))
		a.setRowType(row, valueTypeString, true, false)
	}
}

func (a *Arguments) setRowType(row *argRow, newType valueType, keepDropdown bool, carryValue bool) {
	if row.valueType == newType {
		return
	}

	oldType := row.valueType
	oldText := ""
	if row.value != nil {
		oldText = row.value.GetText()
	}

	row.valueType = newType
	row.typeDrop.SetOptions(valueTypeLabels, nil)
	row.typeDrop.SetCurrentOption(int(newType))
	row.invalid = false

	var preferred tview.Primitive = row.typeDrop
	if !carryValue {
		if newType == valueTypeList {
			if row.list == nil {
				a.setupRowList(row)
			}

			row.invalid = false
			preferred = row.list
		} else {
			row.list = nil
			row.invalid = false
			preferred = row.value
		}

		a.rebuildGrid(true, preferred)
		return
	}

	valueKept := false
	if newType == valueTypeList {
		if row.list == nil {
			a.setupRowList(row)
		}

		if strings.TrimSpace(oldText) != "" {
			row.list.SetTypedValues([]typedValue{{valueType: oldType, text: oldText}})
			valueKept = true
		} else {
			row.list.SetValue(nil)
		}

		preferred = row.list.firstEmptyValueEditor()
		if preferred == nil {
			preferred = row.list
		}

		row.invalid = false
	} else {
		text := ""
		row.invalid = false

		if oldType == valueTypeList && row.list != nil {
			if single, ok := row.list.singleTypedValue(); ok && canConvertText(newType, single.text) {
				text = single.text
				valueKept = true
			}
		} else if canConvertText(newType, oldText) || (newType == valueTypeString && strings.TrimSpace(oldText) != "") {
			text = oldText
			valueKept = true
		}

		row.list = nil
		row.value.SetText(text)

		if !valueKept {
			row.invalid = strings.TrimSpace(oldText) != ""
			preferred = row.value
		}
	}

	if keepDropdown && valueKept {
		preferred = row.typeDrop
	}

	a.applyRowStyles(row)
	a.rebuildGrid(true, preferred)
}

func (a *Arguments) setupRowList(row *argRow) {
	row.list = newListEditor()

	if a.listStyleSet {
		row.list.SetListStyles(a.listStyleUnselected, a.listStyleSelected)
	}

	row.list.SetDisabled(a.disabled)
	row.list.SetFinishedFunc(func(key tcell.Key) {
		a.handleDone(key)
	})
	row.list.SetRowsChangedFunc(func(height int) {
		a.rebuildGrid(true, nil)
		a.emitRowsChanged(a.totalHeight())
	})
	row.list.SetChangedFunc(func() {
		index := a.indexOfRow(row)
		if index >= 0 {
			a.handleRowChange(index)
		}
	})

	a.applyListStyles(row.list)
}

func (a *Arguments) applyStyles() {
	for _, row := range a.rows {
		a.applyRowStyles(row)

		if row.list != nil {
			a.applyListStyles(row.list)
		}

		if a.listStyleSet {
			row.typeDrop.SetListStyles(a.listStyleUnselected, a.listStyleSelected)
		}
	}
}

func (a *Arguments) applyRowStyles(row *argRow) {
	row.key.SetFieldStyle(a.fieldStyle)

	if row.invalid {
		row.value.SetFieldStyle(a.invalidFieldStyle)
	} else {
		row.value.SetFieldStyle(a.fieldStyle)
	}

	row.typeDrop.SetFieldTextColor(a.fieldTextColor)
	row.typeDrop.SetFieldBackgroundColor(a.fieldBgColor)
}

func (a *Arguments) applyListStyles(list *listEditor) {
	list.SetFieldStyle(a.fieldTextColor, a.fieldBgColor)
	list.SetInvalidFieldStyle(a.invalidFieldStyle)
	if a.listStyleSet {
		list.SetListStyles(a.listStyleUnselected, a.listStyleSelected)
	}
}

func (a *Arguments) rebuildGrid(preserveFocus bool, preferred tview.Primitive) {
	var focused tview.Primitive

	if preserveFocus {
		if preferred != nil && a.containsPrimitive(preferred) {
			focused = preferred
		} else {
			focused = a.focusedPrimitive()
		}
	}

	a.grid.Clear()
	a.grid.SetColumns(-11, -9, valueExtraColumnWidth, typeColumnWidth)

	rows, starts := a.rowLayout()
	for rowIndex, row := range a.rows {
		rowStart := starts[rowIndex]
		height := a.rowHeight(row)

		a.grid.AddItem(row.key, rowStart, 0, height, 1, 0, 0, rowIndex == 0)

		if row.valueType == valueTypeList && row.list != nil {
			a.grid.AddItem(row.list, rowStart, 1, height, 2, 0, 0, false)
		} else {
			a.grid.AddItem(row.value, rowStart, 1, height, 2, 0, 0, false)
		}

		a.grid.AddItem(row.typeDrop, rowStart, 3, 1, 1, 0, 0, false)
	}

	a.grid.SetRows(rows...)
	a.setRowHandlers()

	if focused != nil && a.focus != nil {
		if a.containsPrimitive(focused) {
			a.focus(focused)
			return
		}

		focusables := a.focusables()
		if len(focusables) > 0 {
			a.focus(focusables[0])
		}
	}
}

func (a *Arguments) setRowHandlers() {
	for _, row := range a.rows {
		currentRow := row

		row.key.SetDoneFunc(func(key tcell.Key) {
			a.handleDone(key)
		})
		row.key.SetChangedFunc(func(text string) {
			index := a.indexOfRow(currentRow)
			if index >= 0 {
				a.handleRowChange(index)
			}
		})

		if currentRow.valueType != valueTypeList {
			row.value.SetDoneFunc(func(key tcell.Key) {
				a.handleDone(key)
			})
			row.value.SetChangedFunc(func(text string) {
				if currentRow.invalid {
					currentRow.invalid = false
					a.applyRowStyles(currentRow)
				}
				index := a.indexOfRow(currentRow)
				if index >= 0 {
					a.handleRowChange(index)
				}
			})
		}

		row.typeDrop.SetDoneFunc(func(key tcell.Key) {
			a.handleDone(key)
		})
		row.typeDrop.SetSelectedFunc(func(text string, index int) {
			a.setRowType(currentRow, valueType(index), true, true)
		})

		if row.list != nil {
			row.list.SetFinishedFunc(func(key tcell.Key) {
				a.handleDone(key)
			})
			row.list.SetRowsChangedFunc(func(height int) {
				a.rebuildGrid(true, currentRow.list)
				a.emitRowsChanged(a.totalHeight())
			})
			row.list.SetChangedFunc(func() {
				index := a.indexOfRow(currentRow)
				if index >= 0 {
					a.handleRowChange(index)
				}
			})
			row.list.SetNavigateFunc(func(current tview.Primitive, direction int) bool {
				return a.focusRelativeFrom(current, direction)
			})
		}
	}
}

func (a *Arguments) handleDone(key tcell.Key) {
	if a.disabled {
		if a.finished != nil {
			a.finished(key)
		}
		return
	}

	switch key {
	case tcell.KeyTab, tcell.KeyEnter:
		if a.focusRelative(1) {
			return
		}
	case tcell.KeyBacktab:
		if a.focusRelative(-1) {
			return
		}
	default:
	}

	if a.finished != nil {
		a.finished(key)
	}
}

func (a *Arguments) handleRowChange(rowIndex int) {
	if rowIndex < 0 || rowIndex >= len(a.rows) {
		return
	}

	row := a.rows[rowIndex]
	rowEmpty := a.isRowEmpty(row)
	lastIndex := len(a.rows) - 1

	if rowIndex == lastIndex {
		if !rowEmpty {
			a.appendEmptyRow()
		}
		return
	}

	if rowEmpty {
		a.removeRow(rowIndex)
	}
}

func (a *Arguments) appendEmptyRow() {
	a.newRow(false)
	a.rebuildGrid(true, nil)
	a.emitRowsChanged(a.totalHeight())
}

func (a *Arguments) removeRow(rowIndex int) {
	if rowIndex < 0 || rowIndex >= len(a.rows)-1 {
		return
	}

	focusables := a.focusables()
	focusedIndex := a.focusedIndex(focusables)

	a.rows = append(a.rows[:rowIndex], a.rows[rowIndex+1:]...)
	a.rebuildGrid(false, nil)
	a.emitRowsChanged(a.totalHeight())

	if focusedIndex >= 0 && a.focus != nil {
		nextFocusables := a.focusables()
		if len(nextFocusables) == 0 {
			return
		}
		if focusedIndex >= len(nextFocusables) {
			focusedIndex = len(nextFocusables) - 1
		}
		a.focus(nextFocusables[focusedIndex])
	}
}

func (a *Arguments) focusRelative(delta int) bool {
	if a.focus == nil {
		return false
	}

	focusables := a.focusables()
	if len(focusables) == 0 {
		return false
	}

	index := a.focusedIndex(focusables)
	if index == -1 {
		return false
	}

	next := index + delta
	if next < 0 || next >= len(focusables) {
		return false
	}

	a.focus(focusables[next])
	return true
}

func (a *Arguments) focusRelativeFrom(current tview.Primitive, delta int) bool {
	if a.focus == nil {
		return false
	}

	focusables := a.focusables()
	if len(focusables) == 0 {
		return false
	}

	index := -1
	for i, item := range focusables {
		if item == current {
			index = i
			break
		}
	}

	if index == -1 {
		return false
	}

	next := index + delta
	if next < 0 || next >= len(focusables) {
		return false
	}

	a.focus(focusables[next])
	return true
}

func (a *Arguments) focusedPrimitiveHandler() func(event *tcell.EventKey, setFocus func(p tview.Primitive)) {
	for _, item := range a.focusables() {
		if item.HasFocus() {
			return item.InputHandler()
		}
	}

	return nil
}

func (a *Arguments) focusedPrimitivePasteHandler() func(pastedText string, setFocus func(p tview.Primitive)) {
	for _, item := range a.focusables() {
		if item.HasFocus() {
			return item.PasteHandler()
		}
	}

	return nil
}

func (a *Arguments) focusables() []tview.Primitive {
	entries := make([]tview.Primitive, 0, len(a.rows)*3)

	for _, row := range a.rows {
		entries = append(entries, row.key)

		if row.valueType == valueTypeList && row.list != nil {
			listEntries := a.listFocusablesWithRowType(row)
			entries = append(entries, listEntries...)
		} else {
			entries = append(entries, row.value, row.typeDrop)
		}
	}

	return entries
}

func (a *Arguments) focusedPrimitive() tview.Primitive {
	for _, item := range a.focusables() {
		if item.HasFocus() {
			return item
		}
	}

	return nil
}

func (a *Arguments) focusedIndex(items []tview.Primitive) int {
	for index, item := range items {
		if item.HasFocus() {
			return index
		}
	}

	return -1
}

func (a *Arguments) listFocusablesWithRowType(row *argRow) []tview.Primitive {
	if row.list == nil || len(row.list.rows) == 0 {
		return []tview.Primitive{row.typeDrop}
	}

	entries := make([]tview.Primitive, 0, len(row.list.rows)*2+1)

	first := row.list.rows[0]
	if first.value != nil {
		entries = append(entries, first.value)
	}

	if first.typeDrop != nil {
		entries = append(entries, first.typeDrop)
	}

	entries = append(entries, row.typeDrop)

	for i := 1; i < len(row.list.rows); i++ {
		item := row.list.rows[i]

		if item.value != nil {
			entries = append(entries, item.value)
		}

		if item.typeDrop != nil {
			entries = append(entries, item.typeDrop)
		}
	}

	return entries
}

func (a *Arguments) lastValueEditor() tview.Primitive {
	for i := len(a.rows) - 1; i >= 0; i-- {
		row := a.rows[i]

		if row.valueType == valueTypeList && row.list != nil {
			if last := row.list.lastValueEditor(); last != nil {
				return last
			}

			return row.list
		}

		if row.value != nil {
			return row.value
		}
	}

	return nil
}

func (a *Arguments) isRowEmpty(row *argRow) bool {
	if strings.TrimSpace(row.key.GetText()) != "" {
		return false
	}

	if row.valueType == valueTypeList {
		if row.list == nil {
			return true
		}

		return row.list.IsEmpty()
	}

	return strings.TrimSpace(row.value.GetText()) == ""
}

func (a *Arguments) rowHeight(row *argRow) int {
	if row.valueType == valueTypeList && row.list != nil {
		return row.list.Height()
	}

	return 1
}

func (a *Arguments) totalHeight() int {
	height := 0

	for _, row := range a.rows {
		height += a.rowHeight(row)
	}

	if height == 0 {
		return 1
	}

	return height
}

func (a *Arguments) rowLayout() ([]int, []int) {
	rows := make([]int, 0, a.totalHeight())
	starts := make([]int, 0, len(a.rows))

	for _, row := range a.rows {
		starts = append(starts, len(rows))
		height := a.rowHeight(row)

		for i := 0; i < height; i++ {
			rows = append(rows, 1)
		}
	}

	if len(rows) == 0 {
		rows = append(rows, 1)
	}

	return rows, starts
}

func (a *Arguments) emitRowsChanged(height int) {
	if a.rowsChanged != nil {
		a.rowsChanged(height)
	}
}

func (a *Arguments) containsPrimitive(p tview.Primitive) bool {
	for _, item := range a.focusables() {
		if item == p {
			return true
		}
	}

	return false
}

func (a *Arguments) indexOfRow(row *argRow) int {
	for i, current := range a.rows {
		if current == row {
			return i
		}
	}

	return -1
}

func parseTypedValue(vType valueType, text string) (any, bool) {
	switch vType {
	case valueTypeNumber:
		return parseNumber(text)
	case valueTypeBoolean:
		return parseBoolean(text)
	default:
		return text, text != ""
	}
}

func canConvertText(vType valueType, text string) bool {
	text = strings.TrimSpace(text)

	if vType == valueTypeString {
		return true
	}

	switch vType {
	case valueTypeNumber:
		_, ok := parseNumber(text)
		return ok
	case valueTypeBoolean:
		_, ok := parseBoolean(text)
		return ok
	default:
		return false
	}
}

func parseBoolean(text string) (bool, bool) {
	value, err := strconv.ParseBool(strings.TrimSpace(text))

	if err != nil {
		return false, false
	}

	return value, true
}

func parseNumber(text string) (any, bool) {
	text = strings.TrimSpace(text)

	if text == "" {
		return nil, false
	}

	if intValue, err := strconv.Atoi(text); err == nil {
		return intValue, true
	}

	if floatValue, err := strconv.ParseFloat(text, 64); err == nil {
		return floatValue, true
	}

	return nil, false
}

func formatNumber(value any) (string, bool) {
	switch typed := value.(type) {
	case int, int8, int16, int32, int64:
		return fmt.Sprintf("%d", typed), true
	case uint, uint8, uint16, uint32, uint64:
		return fmt.Sprintf("%d", typed), true
	case float32:
		return strconv.FormatFloat(float64(typed), 'f', -1, 32), true
	case float64:
		return strconv.FormatFloat(typed, 'f', -1, 64), true
	default:
		return "", false
	}
}

func normalizeSlice(value any) ([]any, bool) {
	rv := reflect.ValueOf(value)
	if !rv.IsValid() || rv.Kind() != reflect.Slice {
		return nil, false
	}

	length := rv.Len()
	items := make([]any, 0, length)

	for i := 0; i < length; i++ {
		items = append(items, rv.Index(i).Interface())
	}

	return items, true
}

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
	result := make([]any, 0)

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
	l.Box.DrawForSubclass(screen, l)
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
	if l.lastFocus != nil && l.containsPrimitive(l.lastFocus) {
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
				if capture != nil && l.containsPrimitive(capture) {
					l.lastFocus = capture
				} else if focused := l.focusedPrimitive(); focused != nil {
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

	switch typed := value.(type) {
	case bool:
		row.value.SetText(strconv.FormatBool(typed))
		row.valueType = valueTypeBoolean
		row.typeDrop.SetCurrentOption(int(valueTypeBoolean))
	case string:
		row.value.SetText(typed)
		row.valueType = valueTypeString
		row.typeDrop.SetCurrentOption(int(valueTypeString))
	default:
		if numText, ok := formatNumber(value); ok {
			row.value.SetText(numText)
			row.valueType = valueTypeNumber
			row.typeDrop.SetCurrentOption(int(valueTypeNumber))
			return
		}
		row.value.SetText(fmt.Sprint(value))
		row.valueType = valueTypeString
		row.typeDrop.SetCurrentOption(int(valueTypeString))
	}

	row.invalid = false
	l.applyRowStyles(row)
}

func (l *listEditor) rebuildGrid(preserveFocus bool, preferred tview.Primitive) {
	var focused tview.Primitive

	if preserveFocus {
		if preferred != nil && l.containsPrimitive(preferred) {
			focused = preferred
		} else {
			focused = l.focusedPrimitive()
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

	if focused != nil && l.focus != nil && l.containsPrimitive(focused) {
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
		if l.navigate != nil {
			if l.navigate(l.focusedPrimitive(), 1) {
				return
			}
		}
		if l.focusRelative(1) {
			return
		}
	case tcell.KeyBacktab:
		if l.navigate != nil {
			if l.navigate(l.focusedPrimitive(), -1) {
				return
			}
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

	lastIndex := len(l.rows) - 1
	valueEmpty := strings.TrimSpace(l.rows[rowIndex].value.GetText()) == ""

	if rowIndex == lastIndex {
		if !valueEmpty {
			l.appendEmptyRow(l.rows[rowIndex].value)
		}
	} else if valueEmpty {
		l.removeRow(rowIndex)
	}

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

	focusables := l.focusables()
	focusedIndex := l.focusedIndex(focusables)

	l.rows = append(l.rows[:rowIndex], l.rows[rowIndex+1:]...)
	l.rebuildGrid(false, nil)
	l.emitRowsChanged(l.Height())

	if focusedIndex >= 0 && l.focus != nil {
		nextFocusables := l.focusables()

		if len(nextFocusables) == 0 {
			return
		}

		if focusedIndex >= len(nextFocusables) {
			focusedIndex = len(nextFocusables) - 1
		}

		l.focus(nextFocusables[focusedIndex])
	}
}

func (l *listEditor) focusRelative(delta int) bool {
	if l.focus == nil {
		return false
	}

	focusables := l.focusables()
	if len(focusables) == 0 {
		return false
	}

	index := l.focusedIndex(focusables)
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

func (l *listEditor) focusedPrimitive() tview.Primitive {
	for _, item := range l.focusables() {
		if item.HasFocus() {
			return item
		}
	}

	return nil
}

func (l *listEditor) focusedIndex(items []tview.Primitive) int {
	for index, item := range items {
		if item.HasFocus() {
			return index
		}
	}

	return -1
}

func (l *listEditor) emitRowsChanged(height int) {
	if l.rowsChanged != nil {
		l.rowsChanged(height)
	}
}

func (l *listEditor) containsPrimitive(p tview.Primitive) bool {
	for _, item := range l.focusables() {
		if item == p {
			return true
		}
	}

	return false
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
