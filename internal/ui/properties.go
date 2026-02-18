package ui

import (
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

type Properties struct {
	*tview.Box

	grid *tview.Grid

	label      string
	labelWidth int
	labelColor tcell.Color
	bgColor    tcell.Color

	fieldTextColor tcell.Color
	fieldBgColor   tcell.Color
	fieldStyle     tcell.Style

	rows     []*propertyRow
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

type propertyRow struct {
	keyDrop   *tview.DropDown
	value     *tview.InputField
	propDef   *propertyDef
	invalid   bool
	usedIndex int // Property index in the messageProperties or -1 if not selected
}

func NewProperties() *Properties {
	p := Properties{
		Box:  tview.NewBox(),
		grid: tview.NewGrid(),
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
	p.Box.DrawForSubclass(screen, p)

	x, y, width, height := p.GetInnerRect()
	labelWidth := p.labelWidth
	if labelWidth > width {
		labelWidth = width
	}
	if height > 0 && labelWidth > 0 {
		tview.Print(screen, p.label, x, y, labelWidth, tview.AlignLeft, p.labelColor)
	}

	fieldX := x + labelWidth
	fieldWidth := width - labelWidth
	if fieldWidth < 0 {
		fieldWidth = 0
	}

	p.grid.SetRect(fieldX, y, fieldWidth, height)
	p.grid.Draw(screen)
}

// GetLabel returns the item's label text.
func (p *Properties) GetLabel() string {
	return p.label
}

// SetFormAttributes sets form attributes.
func (p *Properties) SetFormAttributes(labelWidth int, labelColor, bgColor, fieldTextColor, fieldBgColor tcell.Color) tview.FormItem {
	p.labelWidth = labelWidth
	p.labelColor = labelColor
	p.bgColor = bgColor
	p.fieldTextColor = fieldTextColor
	p.fieldBgColor = fieldBgColor
	p.fieldStyle = tcell.StyleDefault.Foreground(fieldTextColor).Background(fieldBgColor)
	if !p.invalidStyleSet {
		p.invalidFieldStyle = p.fieldStyle
	}

	p.SetBackgroundColor(bgColor)
	p.applyStyles()

	return p
}

// GetFieldWidth returns the width of the form item's field area.
func (p *Properties) GetFieldWidth() int {
	return 0
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
	p.invalidStyleSet = true
	p.invalidFieldStyle = style

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
	if p.finished != nil && p.disabled {
		p.finished(-1)
		return
	}

	p.focus = delegate

	if ConsumeBacktab() {
		if lastValue := p.lastValueEditor(); lastValue != nil {
			delegate(lastValue)
			return
		}
	}

	if focusables := p.focusables(); len(focusables) > 0 {
		delegate(focusables[0])
		return
	}

	if p.grid != nil {
		delegate(p.grid)
		return
	}

	p.Box.Focus(delegate)
}

// HasFocus returns whether this primitive has focus.
func (p *Properties) HasFocus() bool {
	for _, item := range p.focusables() {
		if item.HasFocus() {
			return true
		}
	}

	if p.grid != nil {
		return p.grid.HasFocus()
	}

	return p.Box.HasFocus()
}

// MouseHandler returns the mouse handler for this primitive.
func (p *Properties) MouseHandler() func(action tview.MouseAction, event *tcell.EventMouse, setFocus func(p tview.Primitive)) (consumed bool, capture tview.Primitive) {
	return p.WrapMouseHandler(func(action tview.MouseAction, event *tcell.EventMouse, setFocus func(p tview.Primitive)) (consumed bool, capture tview.Primitive) {
		if !p.InRect(event.Position()) {
			return false, nil
		}

		p.focus = setFocus

		// Pass mouse events to the contained primitive.
		if p.grid != nil {
			consumed, capture = p.grid.MouseHandler()(action, event, setFocus)
			if consumed {
				return true, capture
			}
		}

		return
	})
}

// InputHandler returns the handler for this primitive.
func (p *Properties) InputHandler() func(event *tcell.EventKey, setFocus func(p tview.Primitive)) {
	return p.WrapInputHandler(func(event *tcell.EventKey, setFocus func(p tview.Primitive)) {
		p.focus = setFocus

		if handler := p.focusedPrimitiveHandler(); handler != nil {
			handler(event, setFocus)
			return
		}

		if p.grid == nil {
			return
		}

		if handler := p.grid.InputHandler(); handler != nil {
			handler(event, setFocus)
			return
		}
	})
}

// PasteHandler returns the paste handler for this primitive.
func (p *Properties) PasteHandler() func(pastedText string, setFocus func(p tview.Primitive)) {
	return p.WrapPasteHandler(func(pastedText string, setFocus func(p tview.Primitive)) {
		if handler := p.focusedPrimitivePasteHandler(); handler != nil {
			handler(pastedText, setFocus)
			return
		}

		if p.grid == nil {
			return
		}

		if handler := p.grid.PasteHandler(); handler != nil {
			handler(pastedText, setFocus)
			return
		}
	})
}

func (p *Properties) newRow(focus bool) *propertyRow {
	row := &propertyRow{
		keyDrop:   tview.NewDropDown(),
		value:     tview.NewInputField().SetPlaceholder("Property value"),
		usedIndex: -1,
	}

	// Set up the dropdown with available properties (no placeholder)
	row.keyDrop.SetOptions(availablePropertyLabels(), nil)
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

	// Update the list of available properties
	row.keyDrop.SetOptions(p.availablePropertyLabelsForRow(row), nil)

	// Set the currently selected property
	availableProps := p.availablePropertiesForRow(row)
	for i, prop := range availableProps {
		if prop.jsonKey == propDef.jsonKey {
			row.keyDrop.SetCurrentOption(i)
			break
		}
	}

	row.value.SetPlaceholder(propDef.placeholder)
}

func (p *Properties) setRowValue(row *propertyRow, value any) {
	if value == nil || row.propDef == nil {
		return
	}

	switch typed := value.(type) {
	case string:
		row.value.SetText(typed)
	default:
		if numText, ok := formatNumber(value); ok {
			row.value.SetText(numText)
		} else {
			row.value.SetText("")
		}
	}

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
	if p.disabled {
		if p.finished != nil {
			p.finished(key)
		}
		return
	}

	switch key {
	case tcell.KeyTab, tcell.KeyEnter:
		if p.focusRelative(1) {
			return
		}
	case tcell.KeyBacktab:
		if p.focusRelative(-1) {
			return
		}
	default:
	}

	if p.finished != nil {
		p.finished(key)
	}
}

func (p *Properties) handleRowChange(rowIndex int) {
	if rowIndex < 0 || rowIndex >= len(p.rows) {
		return
	}

	row := p.rows[rowIndex]
	rowEmpty := p.isRowEmpty(row)
	lastIndex := len(p.rows) - 1

	if rowIndex == lastIndex {
		if !rowEmpty {
			p.appendEmptyRow()
		}
		return
	}

	if rowEmpty {
		p.removeRow(rowIndex)
	}
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

	focusables := p.focusables()
	focusedIndex := p.focusedIndex(focusables)

	p.rows = append(p.rows[:rowIndex], p.rows[rowIndex+1:]...)
	p.updateAllRowDropdowns()
	p.rebuildGrid(false, nil)
	p.emitRowsChanged(p.totalHeight())

	if focusedIndex >= 0 && p.focus != nil {
		nextFocusables := p.focusables()
		if len(nextFocusables) == 0 {
			return
		}
		if focusedIndex >= len(nextFocusables) {
			focusedIndex = len(nextFocusables) - 1
		}
		p.focus(nextFocusables[focusedIndex])
	}
}

func (p *Properties) focusRelative(delta int) bool {
	if p.focus == nil {
		return false
	}

	focusables := p.focusables()
	if len(focusables) == 0 {
		return false
	}

	index := p.focusedIndex(focusables)
	if index == -1 {
		return false
	}

	next := index + delta
	if next < 0 || next >= len(focusables) {
		return false
	}

	p.focus(focusables[next])
	return true
}

func (p *Properties) focusedPrimitiveHandler() func(event *tcell.EventKey, setFocus func(p tview.Primitive)) {
	for _, item := range p.focusables() {
		if item.HasFocus() {
			return item.InputHandler()
		}
	}

	return nil
}

func (p *Properties) focusedPrimitivePasteHandler() func(pastedText string, setFocus func(p tview.Primitive)) {
	for _, item := range p.focusables() {
		if item.HasFocus() {
			return item.PasteHandler()
		}
	}

	return nil
}

func (p *Properties) focusables() []tview.Primitive {
	entries := make([]tview.Primitive, 0, len(p.rows)*2)

	for _, row := range p.rows {
		entries = append(entries, row.keyDrop, row.value)
	}

	return entries
}

func (p *Properties) focusedPrimitive() tview.Primitive {
	for _, item := range p.focusables() {
		if item.HasFocus() {
			return item
		}
	}

	return nil
}

func (p *Properties) focusedIndex(items []tview.Primitive) int {
	for index, item := range items {
		if item.HasFocus() {
			return index
		}
	}

	return -1
}

func (p *Properties) lastValueEditor() tview.Primitive {
	for i := len(p.rows) - 1; i >= 0; i-- {
		row := p.rows[i]
		if row.value != nil {
			return row.value
		}
	}

	return nil
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

func (p *Properties) emitRowsChanged(height int) {
	if p.rowsChanged != nil {
		p.rowsChanged(height)
	}
}

func (p *Properties) containsPrimitive(prim tview.Primitive) bool {
	for _, item := range p.focusables() {
		if item == prim {
			return true
		}
	}

	return false
}

func (p *Properties) indexOfRow(row *propertyRow) int {
	for i, current := range p.rows {
		if current == row {
			return i
		}
	}

	return -1
}

// updateAllRowDropdowns updates the lists of available properties in all rows
func (p *Properties) updateAllRowDropdowns() {
	for _, row := range p.rows {
		labels := p.availablePropertyLabelsForRow(row)
		row.keyDrop.SetOptions(labels, nil)

		// Restore selection if a property was already selected
		if row.propDef != nil {
			availableProps := p.availablePropertiesForRow(row)
			for i, prop := range availableProps {
				if prop.jsonKey == row.propDef.jsonKey {
					row.keyDrop.SetCurrentOption(i)
					break
				}
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

// availablePropertyLabelsForRow returns the list of labels of available properties for the given row
func (p *Properties) availablePropertyLabelsForRow(row *propertyRow) []string {
	props := p.availablePropertiesForRow(row)
	labels := make([]string, len(props))
	for i, prop := range props {
		labels[i] = prop.label
	}
	return labels
}

// availablePropertyLabels returns the list of all property labels
func availablePropertyLabels() []string {
	labels := make([]string, len(messageProperties))
	for i, prop := range messageProperties {
		labels[i] = prop.label
	}
	return labels
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
