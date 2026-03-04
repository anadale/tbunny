package application

import (
	"fmt"
	"log/slog"
	"sync"
	"tbunny/internal/model"
	"tbunny/internal/sl"
	"tbunny/internal/ui"

	"github.com/rivo/tview"
)

const modalDialogKey = "dialog"

const (
	stackPush stackAction = 1 << iota
	stackPop
)

type stackAction int

type ViewStackListener interface {
	StackPushed(model.View)
	StackPopped(old, new model.View)
	StackTop(model.View)
}

type ViewStack struct {
	views     []model.View
	listeners []ViewStackListener
	mx        sync.RWMutex

	app   *App
	pages *tview.Pages
}

func NewViewStack(app *App) *ViewStack {
	p := ViewStack{
		app:   app,
		pages: tview.NewPages(),
	}

	return &p
}

func (vs *ViewStack) AddListener(l ViewStackListener) {
	vs.mx.Lock()
	vs.listeners = append(vs.listeners, l)
	vs.mx.Unlock()

	if !vs.Empty() {
		l.StackTop(vs.Top())
	}
}

func (vs *ViewStack) RemoveListener(l ViewStackListener) {
	vs.mx.Lock()
	defer vs.mx.Unlock()

	for i, l2 := range vs.listeners {
		if l2 == l {
			vs.listeners = append(vs.listeners[:i], vs.listeners[i+1:]...)
			return
		}
	}
}

func (vs *ViewStack) Empty() bool {
	vs.mx.RLock()
	defer vs.mx.RUnlock()

	return len(vs.views) == 0
}

func (vs *ViewStack) Push(v model.View) {
	var top model.View

	vs.mx.Lock()
	if len(vs.views) > 0 {
		top = vs.views[len(vs.views)-1]
	}

	vs.views = append(vs.views, v)
	vs.mx.Unlock()

	if top != nil {
		top.Stop()
	}

	vs.addAndShow(v)
	v.Start()
	vs.app.SetFocus(v.Primitive())

	vs.notify(stackPush, v)
}

func (vs *ViewStack) Pop() (model.View, bool) {
	if vs.Empty() {
		return nil, false
	}

	var top model.View

	vs.mx.Lock()
	v := vs.views[len(vs.views)-1]
	vs.views = vs.views[:len(vs.views)-1]

	if len(vs.views) > 0 {
		top = vs.views[len(vs.views)-1]
	}
	vs.mx.Unlock()

	v.Stop()
	vs.delete(v)

	if top != nil {
		vs.show(top)
		top.Start()
		vs.app.SetFocus(top.Primitive())
	}

	vs.notify(stackPop, v)

	return v, true
}

func (vs *ViewStack) Clear() {
	for range vs.views {
		vs.Pop()
	}
}

func (vs *ViewStack) Last() bool {
	vs.mx.RLock()
	defer vs.mx.RUnlock()

	return len(vs.views) == 1
}

func (vs *ViewStack) Top() model.View {
	if vs.Empty() {
		return nil
	}

	vs.mx.RLock()
	defer vs.mx.RUnlock()

	return vs.views[len(vs.views)-1]
}

func (vs *ViewStack) CollectNames() []string {
	vs.mx.RLock()
	defer vs.mx.RUnlock()

	names := make([]string, len(vs.views))
	for i, c := range vs.views {
		names[i] = c.Name()
	}

	return names
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

	top := vs.Top()

	if top != nil {
		vs.app.SetFocus(top.Primitive())
	}
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

func (vs *ViewStack) notify(action stackAction, v model.View) {
	vs.mx.RLock()
	listeners := make([]ViewStackListener, len(vs.listeners))
	copy(listeners, vs.listeners)
	vs.mx.RUnlock()

	for _, l := range listeners {
		switch action {
		case stackPush:
			l.StackPushed(v)
		case stackPop:
			l.StackPopped(v, vs.Top())
		}
	}
}

func viewID(v model.View) string {
	if v.Name() == "" {
		slog.Error("View has no name", sl.Component, fmt.Sprintf("%T", v))
	}

	return fmt.Sprintf("%s-%p", v.Name(), v)
}
