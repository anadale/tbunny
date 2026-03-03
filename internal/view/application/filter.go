package application

import (
	"tbunny/internal/model"
	"unicode/utf8"

	"github.com/gdamore/tcell/v2"
	"github.com/rivo/tview"
)

type Filter struct {
	*tview.TextView

	app      *App
	text     string
	filterer model.Filterer
}

func NewFilter(app *App) *Filter {
	f := &Filter{
		TextView: tview.NewTextView(),
		app:      app,
	}

	f.SetBorder(true)
	f.SetBorderPadding(0, 0, 1, 1)
	f.SetInputCapture(f.keyboard)

	return f
}

func (f *Filter) Open(filterer model.Filterer) {
	f.filterer = filterer
	f.SetText("")
}

func (f *Filter) SetText(text string) {
	f.text = text
	f.TextView.SetText(text)

	f.filterer.Filter(text)
}

func (f *Filter) keyboard(event *tcell.EventKey) *tcell.EventKey {
	switch event.Key() {
	case tcell.KeyBackspace2, tcell.KeyBackspace:
		f.delete()
	case tcell.KeyRune:
		f.insert(event.Rune())
	case tcell.KeyEnter:
		f.filterer.Filter(f.text)
		f.app.closeFilter()
	case tcell.KeyEscape:
		f.filterer.Filter("")
		f.app.closeFilter()
	default:
	}

	return nil
}

func (f *Filter) delete() {
	if len(f.text) == 0 {
		return
	}

	_, size := utf8.DecodeLastRuneInString(f.text)

	f.SetText(f.text[:len(f.text)-size])
}

func (f *Filter) insert(c rune) {
	f.SetText(f.text + string(c))
}
