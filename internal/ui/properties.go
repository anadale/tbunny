package ui

import (
	"slices"
	"strings"

	"github.com/gdamore/tcell/v2"
	"github.com/rivo/tview"
)

// propertyDef defines a message property
type propertyDef struct {
	jsonKey     string
	label       string
	valueType   valueType
	placeholder string
}

// Available message properties (without DeliveryMode and Headers)
var messageProperties = []propertyDef{
	{jsonKey: "app_id", label: "App ID", valueType: valueTypeString, placeholder: "Application ID"},
	{jsonKey: "content_encoding", label: "Content Encoding", valueType: valueTypeString, placeholder: "Content encoding"},
	{jsonKey: "content_type", label: "Content Type", valueType: valueTypeString, placeholder: "Content type"},
	{jsonKey: "correlation_id", label: "Correlation ID", valueType: valueTypeString, placeholder: "Correlation ID"},
	{jsonKey: "expiration", label: "Expiration", valueType: valueTypeString, placeholder: "Expiration time (ms)"},
	{jsonKey: "message_id", label: "Message ID", valueType: valueTypeString, placeholder: "Message ID"},
	{jsonKey: "priority", label: "Priority", valueType: valueTypeNumber, placeholder: "Priority (0-9)"},
	{jsonKey: "reply_to", label: "Reply To", valueType: valueTypeString, placeholder: "Reply-to address"},
	{jsonKey: "timestamp", label: "Timestamp", valueType: valueTypeNumber, placeholder: "Unix timestamp"},
	{jsonKey: "type", label: "Type", valueType: valueTypeString, placeholder: "Message type"},
	{jsonKey: "user_id", label: "User ID", valueType: valueTypeString, placeholder: "User ID"},
}

// allPropertyLabels is the list of all message property labels, computed once.
var allPropertyLabels = func() []string {
	labels := make([]string, len(messageProperties))
	for i, prop := range messageProperties {
		labels[i] = prop.label
	}
	return labels
}()

type Properties struct {
	formWidgetBase

	rows []*propertyRow
}

type propertyRow struct {
	keyDrop   *tview.DropDown
	value     *tview.InputField
	propDef   *propertyDef
	invalid   bool
	usedIndex int // Property index in the messageProperties or -1 if not selected
}

func NewProperties() *Properties {
	p := Properties{
		formWidgetBase: formWidgetBase{
			Box:  tview.NewBox(),
			grid: tview.NewGrid(),
		},
	}

	p.SetValue(nil)

	return &p
}

func (p *Properties) SetLabel(label string) *Properties {
	p.label = label
	return p
}

func (p *Properties) SetValue(props map[string]any) int {
	p.grid.Clear()
	p.rows = nil
	p.grid.SetColumns(-11, -9, valueExtraColumnWidth, typeColumnWidth)

	firstField := true
	for key, value := range props {
		propDef := findPropertyByJsonKey(key)
		if propDef == nil {
			continue
		}

		row := p.newRow(firstField)
		p.setRowProperty(row, propDef)
		p.setRowValue(row, value)
		firstField = false
	}

	p.ensureTrailingEmptyRow(firstField)
	p.rebuildGrid(false, nil)

	height := p.totalHeight()
	p.emitRowsChanged(height)

	return height
}

func (p *Properties) GetValue() map[string]any {
	result := make(map[string]any)

	for _, row := range p.rows {
		if row.propDef == nil {
			continue
		}

		text := strings.TrimSpace(row.value.GetText())
		if text == "" {
			continue
		}

		value, ok := parseTypedValue(row.propDef.valueType, text)
		if !ok {
			continue
		}

		result[row.propDef.jsonKey] = value
	}

	return result
}

func (p *Properties) Draw(screen tcell.Screen) {
	drawFormWidget(screen, p, p.Box, p.grid, p.label, p.labelWidth, p.labelColor)
}

// SetFormAttributes sets form attributes.
func (p *Properties) SetFormAttributes(labelWidth int, labelColor, bgColor, fieldTextColor, fieldBgColor tcell.Color) tview.FormItem {
	p.applyFormAttributes(labelWidth, labelColor, bgColor, fieldTextColor, fieldBgColor)
	p.applyStyles()

	return p
}

// GetFieldHeight returns the height of the form item's field area.
func (p *Properties) GetFieldHeight() int {
	height := p.totalHeight()
	if height > 0 {
		return height
	}

	return 1
}

// SetFinishedFunc sets the handler for when the users finished entering data.
func (p *Properties) SetFinishedFunc(handler func(key tcell.Key)) tview.FormItem {
	p.finished = handler

	return p
}

// SetDisabled sets whether the item is disabled / read-only.
func (p *Properties) SetDisabled(disabled bool) tview.FormItem {
	p.disabled = disabled

	for _, row := range p.rows {
		row.keyDrop.SetDisabled(disabled)
		if row.value != nil {
			row.value.SetDisabled(disabled)
		}
	}

	if p.finished != nil {
		p.finished(-1)
	}

	return p
}

// SetRowsChangedFunc sets a handler which is called when the number of rows changes.
func (p *Properties) SetRowsChangedFunc(handler func(height int)) *Properties {
	p.rowsChanged = handler

	return p
}

// SetInvalidFieldStyle sets the style used to indicate an invalid value.
func (p *Properties) SetInvalidFieldStyle(style tcell.Style) *Properties {
	p.setInvalidStyle(style)
	p.applyStyles()

	return p
}

// SetListStyles sets the styles for all dropdown lists.
func (p *Properties) SetListStyles(unselected, selected tcell.Style) *Properties {
	p.listStyleSet = true
	p.listStyleUnselected = unselected
	p.listStyleSelected = selected

	for _, row := range p.rows {
		row.keyDrop.SetListStyles(unselected, selected)
	}

	return p
}

// Focus is called when this primitive receives focus.
func (p *Properties) Focus(delegate func(p tview.Primitive)) {
	p.focus = delegate

	focusFormWidget(delegate, p.finished, p.disabled, p.lastValueEditor, p.focusables, p.grid, p.Box)
}

// HasFocus returns whether this primitive has focus.
func (p *Properties) HasFocus() bool {
	return hasFormWidgetFocus(p.focusables(), p.grid, p.Box)
}

// MouseHandler returns the mouse handler for this primitive.
func (p *Properties) MouseHandler() func(action tview.MouseAction, event *tcell.EventMouse, setFocus func(p tview.Primitive)) (consumed bool, capture tview.Primitive) {
	return formWidgetMouseHandler(p.Box, p.grid, &p.focus)
}

// InputHandler returns the handler for this primitive.
func (p *Properties) InputHandler() func(event *tcell.EventKey, setFocus func(p tview.Primitive)) {
	return formWidgetInputHandler(p.Box, p.grid, &p.focus, p.focusables)
}

// PasteHandler returns the paste handler for this primitive.
func (p *Properties) PasteHandler() func(pastedText string, setFocus func(p tview.Primitive)) {
	return formWidgetPasteHandler(p.Box, p.grid, p.focusables)
}

func (p *Properties) newRow(focus bool) *propertyRow {
	row := &propertyRow{
		keyDrop:   tview.NewDropDown(),
		value:     tview.NewInputField().SetPlaceholder("Property value"),
		usedIndex: -1,
	}

	// Set up the dropdown with available properties (no placeholder)
	row.keyDrop.SetOptions(allPropertyLabels, nil)
	// Don't set the current option - leave the dropdown empty

	if p.listStyleSet {
		row.keyDrop.SetListStyles(p.listStyleUnselected, p.listStyleSelected)
	}

	row.keyDrop.SetDisabled(p.disabled)
	row.value.SetDisabled(p.disabled)

	p.applyRowStyles(row)
	p.rows = append(p.rows, row)

	if focus && p.focus != nil {
		p.focus(row.keyDrop)
	}

	return row
}

func (p *Properties) ensureTrailingEmptyRow(focus bool) {
	if len(p.rows) == 0 {
		p.newRow(focus)
		return
	}

	last := p.rows[len(p.rows)-1]
	if !p.isRowEmpty(last) {
		p.newRow(focus)
	}
}

func (p *Properties) setRowProperty(row *propertyRow, propDef *propertyDef) {
	row.propDef = propDef
	row.usedIndex = findPropertyIndex(propDef.jsonKey)
	row.value.SetPlaceholder(propDef.placeholder)
	p.updateRowDropdown(row)
}

func (p *Properties) setRowValue(row *propertyRow, value any) {
	if value == nil || row.propDef == nil {
		return
	}

	text, _ := formatValueText(value)
	row.value.SetText(text)
	row.invalid = false
	p.applyRowStyles(row)
}

func (p *Properties) applyStyles() {
	for _, row := range p.rows {
		p.applyRowStyles(row)

		if p.listStyleSet {
			row.keyDrop.SetListStyles(p.listStyleUnselected, p.listStyleSelected)
		}
	}
}

func (p *Properties) applyRowStyles(row *propertyRow) {
	row.keyDrop.SetFieldTextColor(p.fieldTextColor)
	row.keyDrop.SetFieldBackgroundColor(p.fieldBgColor)

	if row.invalid {
		row.value.SetFieldStyle(p.invalidFieldStyle)
	} else {
		row.value.SetFieldStyle(p.fieldStyle)
	}
}

func (p *Properties) rebuildGrid(preserveFocus bool, preferred tview.Primitive) {
	var focused tview.Primitive

	if preserveFocus {
		if preferred != nil && p.containsPrimitive(preferred) {
			focused = preferred
		} else {
			focused = p.focusedPrimitive()
		}
	}

	p.grid.Clear()
	p.grid.SetColumns(-11, -9, valueExtraColumnWidth, typeColumnWidth)

	rows := make([]int, 0, len(p.rows))
	for rowIndex, row := range p.rows {
		rows = append(rows, 1)
		p.grid.AddItem(row.keyDrop, rowIndex, 0, 1, 1, 0, 0, rowIndex == 0)
		p.grid.AddItem(row.value, rowIndex, 1, 1, 3, 0, 0, false)
	}

	if len(rows) == 0 {
		rows = append(rows, 1)
	}

	p.grid.SetRows(rows...)
	p.setRowHandlers()

	if focused != nil && p.focus != nil {
		if p.containsPrimitive(focused) {
			p.focus(focused)
			return
		}

		focusables := p.focusables()
		if len(focusables) > 0 {
			p.focus(focusables[0])
		}
	}
}

func (p *Properties) setRowHandlers() {
	for _, row := range p.rows {
		currentRow := row

		row.keyDrop.SetDoneFunc(func(key tcell.Key) {
			p.handleDone(key)
		})
		row.keyDrop.SetSelectedFunc(func(text string, index int) {
			availableProps := p.availablePropertiesForRow(currentRow)

			if index < 0 || index >= len(availableProps) {
				return
			}

			propDef := availableProps[index]
			oldPropDef := currentRow.propDef
			currentRow.propDef = &propDef
			currentRow.usedIndex = findPropertyIndex(propDef.jsonKey)

			// Update placeholder and check value compatibility
			currentRow.value.SetPlaceholder(propDef.placeholder)

			oldText := currentRow.value.GetText()
			if oldPropDef == nil || oldPropDef.valueType != propDef.valueType {
				// Check if the current value can be converted to the new type
				if _, ok := parseTypedValue(propDef.valueType, oldText); !ok {
					if strings.TrimSpace(oldText) != "" {
						currentRow.invalid = true
						p.applyRowStyles(currentRow)
					}
				}
			}

			rowIndex := p.indexOfRow(currentRow)
			if rowIndex >= 0 {
				p.handleRowChange(rowIndex)
			}

			// Update the lists of other rows
			p.updateAllRowDropdowns()
		})

		row.value.SetDoneFunc(func(key tcell.Key) {
			p.handleDone(key)
		})
		row.value.SetChangedFunc(func(text string) {
			if currentRow.invalid {
				currentRow.invalid = false
				p.applyRowStyles(currentRow)
			}
			index := p.indexOfRow(currentRow)
			if index >= 0 {
				p.handleRowChange(index)
			}
		})
	}
}

func (p *Properties) handleDone(key tcell.Key) {
	dispatchDoneKey(key, p.disabled, p.finished, p.focusRelative)
}

func (p *Properties) handleRowChange(rowIndex int) {
	if rowIndex < 0 || rowIndex >= len(p.rows) {
		return
	}

	handleRowChangeAt(rowIndex, len(p.rows), p.isRowEmpty(p.rows[rowIndex]), p.appendEmptyRow, p.removeRow)
}

func (p *Properties) appendEmptyRow() {
	p.newRow(false)
	p.rebuildGrid(true, nil)
	p.emitRowsChanged(p.totalHeight())
}

func (p *Properties) removeRow(rowIndex int) {
	if rowIndex < 0 || rowIndex >= len(p.rows)-1 {
		return
	}

	prevFocusedIndex := focusedIndexIn(p.focusables())
	p.rows = slices.Delete(p.rows, rowIndex, rowIndex+1)
	p.updateAllRowDropdowns()
	p.rebuildGrid(false, nil)
	p.emitRowsChanged(p.totalHeight())

	restoreFocusAfterRemoval(prevFocusedIndex, p.focusables(), p.focus)
}

func (p *Properties) focusRelative(delta int) bool {
	return focusRelativeIn(p.focusables(), delta, p.focus)
}

func (p *Properties) focusables() []tview.Primitive {
	entries := make([]tview.Primitive, 0, len(p.rows)*2)

	for _, row := range p.rows {
		entries = append(entries, row.keyDrop, row.value)
	}

	return entries
}

func (p *Properties) focusedPrimitive() tview.Primitive {
	return focusedIn(p.focusables())
}

func (p *Properties) lastValueEditor() tview.Primitive {
	if len(p.rows) == 0 {
		return nil
	}

	return p.rows[len(p.rows)-1].value
}

func (p *Properties) isRowEmpty(row *propertyRow) bool {
	if row.propDef == nil {
		return true
	}

	return strings.TrimSpace(row.value.GetText()) == ""
}

func (p *Properties) totalHeight() int {
	height := len(p.rows)

	if height == 0 {
		return 1
	}

	return height
}

func (p *Properties) containsPrimitive(prim tview.Primitive) bool {
	return containsIn(p.focusables(), prim)
}

func (p *Properties) indexOfRow(row *propertyRow) int {
	for i, current := range p.rows {
		if current == row {
			return i
		}
	}

	return -1
}

// updateAllRowDropdowns updates the available-properties dropdown for every row.
func (p *Properties) updateAllRowDropdowns() {
	for _, row := range p.rows {
		p.updateRowDropdown(row)
	}
}

// updateRowDropdown rebuilds the options for a single row's key dropdown,
// restricting choices to properties not already used by other rows.
func (p *Properties) updateRowDropdown(row *propertyRow) {
	availableProps := p.availablePropertiesForRow(row)

	labels := make([]string, len(availableProps))
	for i, prop := range availableProps {
		labels[i] = prop.label
	}
	row.keyDrop.SetOptions(labels, nil)

	if row.propDef != nil {
		for i, prop := range availableProps {
			if prop.jsonKey == row.propDef.jsonKey {
				row.keyDrop.SetCurrentOption(i)
				break
			}
		}
	}
}

// availablePropertiesForRow returns the list of available properties for the given row
func (p *Properties) availablePropertiesForRow(row *propertyRow) []propertyDef {
	usedIndices := make(map[int]bool)
	for _, r := range p.rows {
		if r != row && r.usedIndex >= 0 {
			usedIndices[r.usedIndex] = true
		}
	}

	available := make([]propertyDef, 0, len(messageProperties))
	for i, prop := range messageProperties {
		if !usedIndices[i] || i == row.usedIndex {
			available = append(available, prop)
		}
	}

	return available
}

// findPropertyByJsonKey finds property definition by JSON key
func findPropertyByJsonKey(jsonKey string) *propertyDef {
	for i := range messageProperties {
		if messageProperties[i].jsonKey == jsonKey {
			return &messageProperties[i]
		}
	}

	return nil
}

// findPropertyIndex finds property index by JSON key
func findPropertyIndex(jsonKey string) int {
	for i, prop := range messageProperties {
		if prop.jsonKey == jsonKey {
			return i
		}
	}

	return -1
}
