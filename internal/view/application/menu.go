package application

import (
	"fmt"
	"regexp"
	"runtime"
	"sort"
	"strconv"
	"strings"
	"tbunny/internal/model"
	"tbunny/internal/skins"

	"github.com/rivo/tview"
)

const (
	menuIndexFmt = " [key:-:b]<%d> [fg:-:fgstyle]%s "
	maxRows      = 6
)

var menuRX = regexp.MustCompile(`\d`)

type Menu struct {
	*tview.Table

	app         *App
	skin        *skins.Skin
	currentView model.View
}

func NewMenu(app *App) *Menu {
	m := Menu{
		Table: tview.NewTable(),
		app:   app,
		skin:  skins.Current(),
	}

	app.content.AddListener(&m)

	skins.AddListener(&m)
	m.SetBackgroundColor(m.skin.BgColor())

	return &m
}

func (m *Menu) SkinChanged(skin *skins.Skin) {
	bgColor := skin.BgColor()

	m.SetBackgroundColor(bgColor)

	for row := range m.GetRowCount() {
		for col := range m.GetColumnCount() {
			if c := m.GetCell(row, col); c != nil {
				c.SetBackgroundColor(bgColor)
			}
		}
	}
}

// ViewActionsChanged implements ViewActionsListener interface
func (m *Menu) ViewActionsChanged(model.View) {
	m.build()
}

func (m *Menu) StackPushed(view model.View) {
	// Unsubscribe from the previous view
	if m.currentView != nil {
		m.currentView.RemoveActionsListener(m)
	}

	// Subscribe to a new view
	m.currentView = view
	view.AddActionsListener(m)

	// Show menu
	m.build()
}

func (m *Menu) StackPopped(_, newView model.View) {
	// Unsubscribe from the old view
	if m.currentView != nil {
		m.currentView.RemoveActionsListener(m)
	}

	if newView != nil {
		// Subscribe to a new view
		m.currentView = newView
		newView.AddActionsListener(m)
	} else {
		m.currentView = nil
	}

	m.build()
}

func (m *Menu) StackTop(view model.View) {
	// StackTop is called when the order changes, but no views are added/removed
	// Re-subscribe if needed
	if m.currentView != view {
		if m.currentView != nil {
			m.currentView.RemoveActionsListener(m)
		}

		m.currentView = view
		view.AddActionsListener(m)
	}

	m.build()
}

// HydrateMenu populate menu ui from hints.
func (m *Menu) build() {
	m.Clear()

	if m.currentView == nil {
		return
	}

	hh := m.currentView.Actions().MenuHints()
	sort.Sort(hh)

	table := make([]model.Hints, maxRows+1)
	colCount := (len(hh) / maxRows) + 1

	if m.hasDigits(hh) {
		colCount++
	}

	for row := range maxRows {
		table[row] = make(model.Hints, colCount)
	}

	t := m.buildMenuTable(hh, table, colCount)

	for row := range t {
		for col := range len(t[row]) {
			c := tview.NewTableCell(t[row][col])

			if t[row][col] == "" {
				c = tview.NewTableCell("")
			}

			c.SetBackgroundColor(m.skin.BgColor())
			m.SetCell(row, col, c)
		}
	}
}

func (*Menu) hasDigits(hh model.Hints) bool {
	for _, h := range hh {
		if menuRX.MatchString(h.Mnemonic) {
			return true
		}
	}

	return false
}

func (m *Menu) buildMenuTable(hh model.Hints, table []model.Hints, colCount int) [][]string {
	var row, col int
	firstCmd := true
	maxKeyWidths := make([]int, colCount)

	for _, h := range hh {
		if !menuRX.MatchString(h.Mnemonic) && firstCmd {
			row, col, firstCmd = 0, col+1, false
			if table[0][0].IsBlank() {
				col = 0
			}
		}

		if maxKeyWidths[col] < len(h.Mnemonic) {
			maxKeyWidths[col] = len(h.Mnemonic)
		}

		table[row][col] = h
		row++

		if row >= maxRows {
			row, col = 0, col+1
		}
	}

	out := make([][]string, len(table))
	for r := range out {
		out[r] = make([]string, len(table[r]))
	}

	m.layout(table, maxKeyWidths, out)

	return out
}

func (m *Menu) layout(table []model.Hints, mm []int, out [][]string) {
	for r := range table {
		for c := range table[r] {
			out[r][c] = m.formatMenu(table[r][c], mm[c])
		}
	}
}

func (m *Menu) formatMenu(h model.Hint, size int) string {
	if h.Mnemonic == "" || h.Description == "" {
		return ""
	}
	styles := m.skin.Frame
	i, err := strconv.Atoi(h.Mnemonic)
	if err == nil {
		return formatNSMenu(i, h.Description, &styles)
	}

	return formatPlainMenu(h, size, &styles)
}

func keyConv(s string) string {
	if s == "" || !strings.Contains(s, "alt") {
		return s
	}
	if runtime.GOOS != "darwin" {
		return s
	}

	return strings.Replace(s, "alt", "opt", 1)
}

func ToMnemonic(s string) string {
	if s == "" {
		return s
	}

	return "<" + keyConv(strings.ToLower(s)) + ">"
}

func formatNSMenu(i int, name string, styles *skins.Frame) string {
	format := strings.Replace(menuIndexFmt, "[key", "["+styles.Menu.NumKeyColor.String(), 1)
	format = strings.ReplaceAll(format, ":bg:", ":"+styles.Title.BgColor.String()+":")
	format = strings.Replace(format, "[fg", "["+styles.Menu.FgColor.String(), 1)
	format = strings.Replace(format, "fgstyle]", styles.Menu.FgStyle.ToShortString()+"]", 1)

	return fmt.Sprintf(format, i, name)
}

func formatPlainMenu(h model.Hint, size int, styles *skins.Frame) string {
	menuFmt := " [key:-:b]%-" + strconv.Itoa(size+2) + "s [fg:-:fgstyle]%s "
	format := strings.Replace(menuFmt, "[key", "["+styles.Menu.KeyColor.String(), 1)
	format = strings.Replace(format, "[fg", "["+styles.Menu.FgColor.String(), 1)
	format = strings.ReplaceAll(format, ":bg:", ":"+styles.Title.BgColor.String()+":")
	format = strings.Replace(format, "fgstyle]", styles.Menu.FgStyle.ToShortString()+"]", 1)

	return fmt.Sprintf(format, ToMnemonic(h.Mnemonic), h.Description)
}
