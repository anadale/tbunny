package ui

import (
	"tbunny/internal/skins"

	"github.com/gdamore/tcell/v2"
	"github.com/rivo/tview"
)

type TableColumn struct {
	Name      string
	Title     string
	Expansion int
	MaxWidth  int
	Align     int
}

type TableRow interface {
	GetTableRowID() string
	GetTableColumnValue(columnName string) string
}

type Table[R TableRow] struct {
	*tview.Table

	columns []TableColumn
	rows    []R
	skin    *skins.Skin
}

func NewTable[R TableRow]() *Table[R] {
	t := Table[R]{
		Table: tview.NewTable(),
	}

	t.SetSelectable(true, false)
	t.SetFixed(1, 1)
	t.SetBorder(true)
	//t.SetBorderAttributes(tcell.AttrBold)
	t.SetBorderPadding(0, 0, 1, 1)

	return &t
}

func (t *Table[R]) Reset() {
	t.rows = nil
}

func (t *Table[R]) SetColumns(columns []TableColumn) {
	t.columns = columns

	if t.rows == nil {
		return
	}

	t.rebuildTable()
}

func (t *Table[R]) Rows() []R {
	return t.rows
}

func (t *Table[R]) SetRows(rows []R) {
	oldRows := t.rows

	t.rows = rows

	if oldRows == nil || rows == nil {
		t.rebuildTable()
	} else {
		t.updateDataRows(oldRows)
	}
}

func (t *Table[R]) GetSelectedRow() (row R, ok bool) {
	if len(t.rows) == 0 {
		return
	}

	rowIdx, _ := t.Table.GetSelection()
	if rowIdx < 1 || rowIdx > len(t.rows) {
		return
	}

	return t.rows[rowIdx-1], true
}

func (t *Table[R]) ApplySkin(skin *skins.Skin) {
	t.skin = skin

	s := skin.Views.Table

	t.SetBackgroundColor(s.BgColor.Color())
	t.SetBorderColor(skin.Frame.Border.FgColor.Color())
	t.SetSelectedStyle(tcell.StyleDefault.Foreground(s.CursorFgColor.Color()).Background(s.CursorBgColor.Color()).Attributes(tcell.AttrBold))

	t.updateStyles()
}

func (t *Table[R]) rebuildTable() {
	t.Clear()

	t.createHeaderRow()
	t.updateDataRows(nil)

	t.ScrollToBeginning()

	if len(t.rows) > 0 {
		t.SetSelectable(true, false)
		t.Select(1, 0)
	} else {
		t.SetSelectable(false, false)
	}
}

func (t *Table[R]) updateDataRows(oldRows []R) {
	// Save horizontal scroll position to preserve it across updates
	_, colOffset := t.GetOffset()

	// Save current selection
	selectedRow, _ := t.GetSelection()
	selectedRowID := t.getRowID(selectedRow)

	// Create a map of new rows for a quick existence check
	newRowsMap := make(map[string]bool)
	for _, row := range t.rows {
		newRowsMap[row.GetTableRowID()] = true
	}

	oldIdx := 0
	newIdx := 0
	tableRowIdx := 1 // current position in table (1 = first row after header)

	for newIdx < len(t.rows) {
		newRow := t.rows[newIdx]
		newID := newRow.GetTableRowID()

		// Skip and delete old rows that don't exist in new ones
		for oldIdx < len(oldRows) && !newRowsMap[oldRows[oldIdx].GetTableRowID()] {
			t.RemoveRow(tableRowIdx)
			oldIdx++
			// We don't increment tableRowIdx here, 'cause we've deleted the row and the next one has "slided" up
		}

		// Check if the current row matches
		if oldIdx < len(oldRows) && oldRows[oldIdx].GetTableRowID() == newID {
			// ID matches - update only cell contents
			for j, column := range t.columns {
				content := newRow.GetTableColumnValue(column.Name)
				cell := t.GetCell(tableRowIdx, j)
				if cell != nil {
					cell.SetText(content)
					cell.SetReference(newRow)
				}
			}
			oldIdx++
			newIdx++
			tableRowIdx++
		} else {
			// ID doesn't match - insert the new row
			t.InsertRow(tableRowIdx)

			for j, column := range t.columns {
				content := newRow.GetTableColumnValue(column.Name)
				cell := t.createDataRowCell(newRow, column, content)

				t.SetCell(tableRowIdx, j, cell)
			}

			newIdx++
			tableRowIdx++
		}
	}

	// Delete all remaining old rows
	for tableRowIdx <= t.GetRowCount()-1 {
		t.RemoveRow(tableRowIdx)
	}

	// Restoring selection
	if selectedRowID != "" && len(t.rows) > 0 {
		// Try to find the row with the same ID
		found := false
		for i, row := range t.rows {
			if row.GetTableRowID() == selectedRowID {
				t.Select(i+1, 0)
				found = true
				break
			}
		}

		// If the row was deleted, select an alternative
		if !found {
			// Index of the row in the old list
			oldSelectedIdx := selectedRow - 1

			// Try to select the row at the same position (next one)
			if oldSelectedIdx < len(t.rows) {
				t.Select(oldSelectedIdx+1, 0)
			} else if len(t.rows) > 0 {
				// If there's no row at the same position, select the last one
				t.Select(len(t.rows), 0)
			}
		}
		t.SetSelectable(true, false)
	} else if len(t.rows) == 0 {
		// If there are no rows, deselect
		t.SetSelectable(false, false)
		t.Select(0, 0)
	} else {
		// Select first row
		t.SetSelectable(true, false)
		t.Select(1, 0)
	}

	// Restore horizontal scroll position
	// (row offset is managed by clampToSelection from Select() calls above)
	rowOff, _ := t.GetOffset()
	t.SetOffset(rowOff, colOffset)
}

func (t *Table[R]) getRowID(rowIndex int) string {
	if rowIndex > 0 {
		cell := t.GetCell(rowIndex, 0)
		if cell != nil {
			if ref := cell.GetReference(); ref != nil {
				if row, ok := ref.(TableRow); ok {
					return row.GetTableRowID()
				}
			}
		}
	}

	return ""
}

func (t *Table[R]) createHeaderRow() {
	for i, column := range t.columns {
		cell := t.createHeaderRowCell(column)

		t.SetCell(0, i, cell)
	}
}

func (t *Table[R]) createHeaderRowCell(column TableColumn) *tview.TableCell {
	return t.createCell(column, column.Title, t.getHeaderRowCellStyle(), false, &column)
}

func (t *Table[R]) createDataRowCell(row TableRow, column TableColumn, content string) *tview.TableCell {
	return t.createCell(column, content, t.getDataRowCellStyle(), true, row)
}

func (t *Table[R]) createCell(column TableColumn, content string, style tcell.Style, selectable bool, ref any) *tview.TableCell {
	cell := tview.NewTableCell(content)

	cell.SetStyle(style)
	cell.SetAlign(column.Align)
	cell.SetExpansion(column.Expansion)
	cell.SetSelectable(selectable)
	cell.SetReference(ref)

	if column.MaxWidth > 0 {
		cell.SetMaxWidth(column.MaxWidth)
	}

	return cell
}

func (t *Table[R]) updateStyles() {
	rowCount := t.GetRowCount()

	style := t.getHeaderRowCellStyle()

	if rowCount > 0 {
		for i := range t.GetColumnCount() {
			t.GetCell(0, i).SetStyle(style)
		}
	}

	style = t.getDataRowCellStyle()

	for i := 1; i < rowCount; i++ {
		for j := range t.GetColumnCount() {
			t.GetCell(i, j).SetStyle(style)
		}
	}
}

func (t *Table[R]) getHeaderRowCellStyle() tcell.Style {
	if t.skin == nil {
		return tcell.StyleDefault
	}

	return tcell.StyleDefault.
		Foreground(t.skin.Views.Table.Header.FgColor.Color()).
		Background(t.skin.Views.Table.Header.BgColor.Color())
}

func (t *Table[R]) getDataRowCellStyle() tcell.Style {
	if t.skin == nil {
		return tcell.StyleDefault
	}

	return tcell.StyleDefault.
		Foreground(t.skin.Views.Table.FgColor.Color()).
		Background(t.skin.Views.Table.BgColor.Color())
}
