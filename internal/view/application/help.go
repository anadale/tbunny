package application

import (
	"sort"
	"strconv"
	"strings"
	"tbunny/internal/model"
	"tbunny/internal/skins"
	"tbunny/internal/ui"
	"tbunny/internal/view"

	"github.com/gdamore/tcell/v2"
	"github.com/mattn/go-runewidth"
	"github.com/rivo/tview"
)

const helpViewName = "Help"
const helpTitle = " [aqua::b]Help "

type Help struct {
	*view.View[*tview.Table]

	app *App
}

func NewHelp(app *App) *Help {
	table := tview.NewTable()

	table.SetSelectable(false, false)
	table.SetBorder(true)
	table.SetBorderPadding(0, 0, 1, 1)
	table.SetTitle(helpTitle)

	return &Help{
		View: view.NewView[*tview.Table](helpViewName, table),
		app:  app,
	}
}

func (h *Help) Init(app model.App) error {
	err := h.View.Init(app)
	if err != nil {
		return err
	}

	h.build()

	h.AddBindingKeysFn(h.bindKeys)
	skins.AddListener(h)
	h.SkinChanged(skins.Current())

	return nil
}

func (h *Help) Start() {
	h.View.Start()

	skins.AddListener(h)
	h.SkinChanged(skins.Current())
}

func (h *Help) Stop() {
	skins.RemoveListener(h)

	h.View.Stop()
}

func (h *Help) SkinChanged(skin *skins.Skin) {
	h.updateStyles(skin)
}

func (h *Help) bindKeys(km ui.KeyMap) {
	km.Add(tcell.KeyEscape, ui.NewKeyAction("Back", h.closeHelpCmd))
	km.Add(ui.KeyQ, ui.NewHiddenKeyAction("Back", h.closeHelpCmd))
	km.Add(ui.KeyHelp, ui.NewHiddenKeyAction("Back", h.closeHelpCmd))
	km.Add(tcell.KeyEnter, ui.NewHiddenKeyAction("Back", h.closeHelpCmd))
}

func (h *Help) closeHelpCmd(*tcell.EventKey) *tcell.EventKey {
	h.App().CloseLastView()

	return nil
}

func (h *Help) build() {
	type providerFn func() (string, model.Hints)
	hintProviders := []providerFn{
		h.viewHints,
		h.generalHints,
		h.navigationHints,
	}

	h.Ui().Clear()

	for i, provider := range hintProviders {
		section, hints := provider()

		h.addSection(section, hints, i*2)
	}
}

func (h *Help) addSection(section string, hints model.Hints, col int) {
	t := h.Ui()

	maxKeyWidth, maxDescWidth := computeMaxWidths(hints)

	t.SetCell(0, col, h.titleCell(section))
	t.SetCell(0, col+1, h.spacerCell(maxKeyWidth))

	row := 1

	for _, hint := range hints {
		t.SetCell(row, col, padCellWithRef(ToMnemonic(hint.Mnemonic), maxKeyWidth, hint.Mnemonic))
		t.SetCell(row, col+1, padCell(hint.Description, maxDescWidth))

		row++
	}
}

func (h *Help) viewHints() (string, model.Hints) {
	top := h.app.content.Top()

	hints := top.Actions().HelpHints()
	sort.Sort(hints)

	return strings.ToUpper(top.Name()), hints
}

func (h *Help) generalHints() (string, model.Hints) {
	return "GENERAL", h.app.Actions().HelpHints()
}

func (h *Help) navigationHints() (string, model.Hints) {
	return "NAVIGATION", model.Hints{
		{
			Mnemonic:    "tab",
			Description: "Field Next",
		},
		{
			Mnemonic:    "backtab",
			Description: "Field Previous",
		},
		{
			Mnemonic:    "g",
			Description: "Goto Top",
		},
		{
			Mnemonic:    "Shift-g",
			Description: "Goto Bottom",
		},
		{
			Mnemonic:    "Ctrl-b",
			Description: "Page Up",
		},
		{
			Mnemonic:    "Ctrl-f",
			Description: "Page Down",
		},
		{
			Mnemonic:    "h",
			Description: "Left",
		},
		{
			Mnemonic:    "l",
			Description: "Right",
		},
		{
			Mnemonic:    "k",
			Description: "Up",
		},
		{
			Mnemonic:    "j",
			Description: "Down",
		},
		/*
			{
				Mnemonic:    "[",
				Description: "History Back",
			},
			{
				Mnemonic:    "]",
				Description: "History Forward",
			},
			{
				Mnemonic:    "-",
				Description: "Last Used Command",
			},
		*/
	}
}

func (h *Help) updateStyles(skin *skins.Skin) {
	h.Ui().SetBackgroundColor(skin.BgColor())

	var (
		t       = h.Ui()
		style   = tcell.StyleDefault.Background(skin.Help.BgColor.Color())
		key     = style.Foreground(skin.Help.KeyColor.Color()).Bold(true)
		numKey  = style.Foreground(skin.Help.NumKeyColor.Color()).Bold(true)
		info    = style.Foreground(skin.Help.FgColor.Color())
		heading = style.Foreground(skin.Help.SectionColor.Color()).Bold(true)
	)

	for col := range t.GetColumnCount() {
		for row := range t.GetRowCount() {
			c := t.GetCell(row, col)
			if c == nil {
				continue
			}

			switch {
			case row == 0:
				c.SetStyle(heading)
			case col%2 != 0:
				c.SetStyle(info)
			default:
				if _, err := strconv.Atoi(extractRef(c)); err == nil {
					c.SetStyle(numKey)
					continue
				}
				c.SetStyle(key)
			}
		}
	}
}

func computeMaxWidths(hints model.Hints) (int, int) {
	maxKeyWidth := 0
	maxDescWidth := 0

	for _, h := range hints {
		if len(h.Mnemonic) > maxKeyWidth {
			maxKeyWidth = len(h.Mnemonic)
		}

		if len(h.Description) > maxDescWidth {
			maxDescWidth = len(h.Description)
		}
	}

	return maxKeyWidth + 2, maxDescWidth
}

func extractRef(c *tview.TableCell) string {
	if ref, ok := c.GetReference().(string); ok {
		return ref
	}

	return c.Text
}

func padCellWithRef(s string, width int, ref any) *tview.TableCell {
	return padCell(s, width).SetReference(ref)
}

func padCell(s string, width int) *tview.TableCell {
	return tview.NewTableCell(pad(s, width))
}

func (h *Help) spacerCell(maxKeyWidth int) *tview.TableCell {
	cell := padCell("", maxKeyWidth)
	cell.SetExpansion(1)

	return cell
}

func (h *Help) titleCell(title string) *tview.TableCell {
	c := tview.NewTableCell(title)
	c.SetExpansion(1)
	c.SetAlign(tview.AlignLeft)

	return c
}

func pad(s string, width int) string {
	if len(s) == width {
		return s
	}

	if len(s) >= width {
		return runewidth.Truncate(s, width, string(tview.SemigraphicsHorizontalEllipsis))
	}

	return s + strings.Repeat(" ", width-len(s))
}
