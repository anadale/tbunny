package ui

import (
	"fmt"
	"reflect"
	"slices"
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
	formWidgetBase

	keyPlaceholder   string
	valuePlaceholder string

	rows []*argRow
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
		formWidgetBase: formWidgetBase{
			Box:  tview.NewBox(),
			grid: tview.NewGrid(),
		},
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
	drawFormWidget(screen, a, a.Box, a.grid, a.label, a.labelWidth, a.labelColor)
}

// SetFormAttributes sets a number of item attributes at once.
func (a *Arguments) SetFormAttributes(labelWidth int, labelColor, bgColor, fieldTextColor, fieldBgColor tcell.Color) tview.FormItem {
	a.applyFormAttributes(labelWidth, labelColor, bgColor, fieldTextColor, fieldBgColor)
	a.applyStyles()

	return a
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
	a.setInvalidStyle(style)
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
	a.focus = delegate

	focusFormWidget(delegate, a.finished, a.disabled, a.lastValueEditor, a.focusables, a.grid, a.Box)
}

// HasFocus returns whether this primitive has focus.
func (a *Arguments) HasFocus() bool {
	return hasFormWidgetFocus(a.focusables(), a.grid, a.Box)
}

// MouseHandler returns the mouse handler for this primitive.
func (a *Arguments) MouseHandler() func(action tview.MouseAction, event *tcell.EventMouse, setFocus func(p tview.Primitive)) (consumed bool, capture tview.Primitive) {
	return formWidgetMouseHandler(a.Box, a.grid, &a.focus)
}

// InputHandler returns the handler for this primitive.
func (a *Arguments) InputHandler() func(event *tcell.EventKey, setFocus func(p tview.Primitive)) {
	return formWidgetInputHandler(a.Box, a.grid, &a.focus, a.focusables)
}

// PasteHandler returns the handler for this primitive.
func (a *Arguments) PasteHandler() func(pastedText string, setFocus func(p tview.Primitive)) {
	return formWidgetPasteHandler(a.Box, a.grid, a.focusables)
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
		a.initRowType(row, valueTypeList)

		row.list.SetValue(items)

		return
	}

	text, vt := formatValueText(value)
	row.value.SetText(text)

	a.initRowType(row, vt)
}

// initRowType sets the type of the row programmatically, without carrying over the
// existing value. Used when loading data into the form.
func (a *Arguments) initRowType(row *argRow, newType valueType) {
	if row.valueType == newType {
		return
	}

	row.valueType = newType
	row.typeDrop.SetOptions(valueTypeLabels, nil)
	row.typeDrop.SetCurrentOption(int(newType))
	row.invalid = false

	var preferred tview.Primitive

	if newType == valueTypeList {
		if row.list == nil {
			a.setupRowList(row)
		}
		preferred = row.list
	} else {
		row.list = nil
		preferred = row.value
	}

	a.rebuildGrid(true, preferred)
}

// changeRowType changes the type of the row interactively, attempting to carry
// over the existing value to the new type when possible. Used when the user
// selects a different type from the drop-down.
func (a *Arguments) changeRowType(row *argRow, newType valueType) {
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

	preferred := a.preferredAfterTypeChange(row, newType, oldType, oldText)
	a.applyRowStyles(row)
	a.rebuildGrid(true, preferred)
}

// preferredAfterTypeChange applies the type change to the row's value field(s)
// and returns the primitive that should receive focus after the rebuild.
func (a *Arguments) preferredAfterTypeChange(row *argRow, newType, oldType valueType, oldText string) tview.Primitive {
	if newType == valueTypeList {
		if row.list == nil {
			a.setupRowList(row)
		}

		if strings.TrimSpace(oldText) != "" {
			row.list.SetTypedValues([]typedValue{{valueType: oldType, text: oldText}})
			return row.typeDrop
		}

		row.list.SetValue(nil)

		if first := row.list.firstEmptyValueEditor(); first != nil {
			return first
		}

		return row.list
	}

	// Switching to a scalar type: try to carry the value over.
	text, valueKept := "", false
	if oldType == valueTypeList && row.list != nil {
		if single, ok := row.list.singleTypedValue(); ok && canConvertText(newType, single.text) {
			text, valueKept = single.text, true
		}
	} else if canConvertText(newType, oldText) || (newType == valueTypeString && strings.TrimSpace(oldText) != "") {
		text, valueKept = oldText, true
	}

	row.list = nil
	row.value.SetText(text)

	if !valueKept {
		row.invalid = strings.TrimSpace(oldText) != ""
		return row.value
	}

	return row.typeDrop
}

func (a *Arguments) setupRowList(row *argRow) {
	row.list = newListEditor()

	if a.listStyleSet {
		row.list.SetListStyles(a.listStyleUnselected, a.listStyleSelected)
	}

	row.list.SetDisabled(a.disabled)
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
			a.changeRowType(currentRow, valueType(index))
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
	dispatchDoneKey(key, a.disabled, a.finished, a.focusRelative)
}

func (a *Arguments) handleRowChange(rowIndex int) {
	if rowIndex < 0 || rowIndex >= len(a.rows) {
		return
	}

	handleRowChangeAt(rowIndex, len(a.rows), a.isRowEmpty(a.rows[rowIndex]), a.appendEmptyRow, a.removeRow)
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

	prevFocusedIndex := focusedIndexIn(a.focusables())
	a.rows = slices.Delete(a.rows, rowIndex, rowIndex+1)
	a.rebuildGrid(false, nil)
	a.emitRowsChanged(a.totalHeight())

	restoreFocusAfterRemoval(prevFocusedIndex, a.focusables(), a.focus)
}

func (a *Arguments) focusRelative(delta int) bool {
	return a.focusRelativeFrom(a.focusedPrimitive(), delta)
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
	return focusedIn(a.focusables())
}

func (a *Arguments) listFocusablesWithRowType(row *argRow) []tview.Primitive {
	return row.list.focusablesWithExternalDrop(row.typeDrop)
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

func (a *Arguments) containsPrimitive(p tview.Primitive) bool {
	return containsIn(a.focusables(), p)
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

// formatValueText returns the display text and value type for an arbitrary value.
// Used by setRowValue in both Arguments and listEditor.
func formatValueText(value any) (string, valueType) {
	switch typed := value.(type) {
	case bool:
		return strconv.FormatBool(typed), valueTypeBoolean
	case string:
		return typed, valueTypeString
	case fmt.Stringer:
		return typed.String(), valueTypeString
	default:
		if numText, ok := formatNumber(value); ok {
			return numText, valueTypeNumber
		}
		return fmt.Sprint(value), valueTypeString
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
