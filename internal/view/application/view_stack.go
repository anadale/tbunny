package application

import (
	"fmt"
	"log/slog"
	"tbunny/internal/model"
	"tbunny/internal/sl"
	"tbunny/internal/ui"

	"github.com/rivo/tview"
)

const modalDialogKey = "dialog"

type ViewStack struct {
	*model.ViewStack

	app   *App
	pages *tview.Pages
	top   model.View
}

func NewViewStack(app *App) *ViewStack {
	p := ViewStack{
		ViewStack: model.NewViewStack(),
		app:       app,
		pages:     tview.NewPages(),
	}

	p.AddListener(&p)

	return &p
}

func (vs *ViewStack) Primitive() tview.Primitive {
	return vs.pages
}

func (vs *ViewStack) IsTopDialog() bool {
	_, item := vs.pages.GetFrontPage()

	switch item.(type) {
	case *tview.Modal, *ui.ModalDialog:
		return true
	default:
		return false
	}
}

func (vs *ViewStack) ShowModal(modal tview.Primitive) {
	resize := true
	visible := true

	if _, ok := modal.(*tview.Modal); ok {
		resize = false
		visible = false
	}

	vs.pages.AddPage(modalDialogKey, modal, resize, visible)
	vs.pages.ShowPage(modalDialogKey)
}

func (vs *ViewStack) DismissModal() {
	vs.pages.RemovePage(modalDialogKey)
	vs.app.SetFocus(vs.top.Primitive())
}

func (vs *ViewStack) addAndShow(c model.View) {
	vs.add(c)
	vs.show(c)
}

func (vs *ViewStack) add(v model.View) {
	vs.pages.AddPage(viewID(v), v.Primitive(), true, true)
}

func (vs *ViewStack) show(v model.View) {
	vs.pages.SwitchToPage(viewID(v))
}

func (vs *ViewStack) delete(v model.View) {
	vs.pages.RemovePage(viewID(v))
}

func (vs *ViewStack) StackPushed(top model.View) {
	if vs.top != nil {
		vs.top.Stop()
	}

	vs.top = top

	vs.addAndShow(top)
	top.Start()
	vs.app.SetFocus(top.Primitive())
}

func (vs *ViewStack) StackPopped(_, top model.View) {
	if vs.top != nil {
		vs.top.Stop()
		vs.delete(vs.top)
	}

	vs.StackTop(top)
}

func (vs *ViewStack) StackTop(top model.View) {
	if top == nil {
		return
	}

	vs.top = top

	vs.show(top)
	top.Start()
	vs.app.SetFocus(top.Primitive())
}

func viewID(v model.View) string {
	if v.Name() == "" {
		slog.Error("View has no name", sl.Component, fmt.Sprintf("%T", v))
	}

	return fmt.Sprintf("%s-%p", v.Name(), v)
}
