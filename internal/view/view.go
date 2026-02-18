package view

import (
	"tbunny/internal/model"
	"tbunny/internal/ui"

	"github.com/gdamore/tcell/v2"
	"github.com/rivo/tview"
)

type UiComponent interface {
	tview.Primitive

	SetInputCapture(fn func(event *tcell.EventKey) *tcell.EventKey) *tview.Box
}

type View[U UiComponent] struct {
	name             string
	app              model.App
	ui               U
	actions          ui.KeyMap
	bindingKeysFn    []ui.BindingKeysFn
	actionsListeners []model.ViewActionsListener
}

func NewView[U UiComponent](name string, ui U) *View[U] {
	c := View[U]{
		ui:   ui,
		name: name,
	}

	c.ui.SetInputCapture(c.keyboard)

	return &c
}

func (v *View[U]) Ui() U {
	return v.ui
}

func (v *View[U]) Primitive() tview.Primitive {
	return v.ui
}

func (v *View[U]) Name() string {
	return v.name
}

func (v *View[U]) Actions() model.KeyMap {
	return v.actions
}

func (v *View[U]) AddActionsListener(listener model.ViewActionsListener) {
	v.actionsListeners = append(v.actionsListeners, listener)
}

func (v *View[U]) RemoveActionsListener(listener model.ViewActionsListener) {
	for i, l := range v.actionsListeners {
		if l == listener {
			v.actionsListeners = append(v.actionsListeners[:i], v.actionsListeners[i+1:]...)
			break
		}
	}
}

func (v *View[U]) App() model.App {
	return v.app
}

func (v *View[U]) AddBindingKeysFn(fn ui.BindingKeysFn) {
	v.bindingKeysFn = append(v.bindingKeysFn, fn)
}

func (v *View[U]) Init(app model.App) (err error) {
	v.app = app

	return nil
}

func (v *View[U]) Start() {
	v.refreshActionsInUpdateDraw()
}

func (v *View[U]) Stop() {
	// Clear listeners to prevent memory leaks
	v.actionsListeners = nil
}

func (v *View[U]) RefreshActions() {
	v.app.QueueUpdateDraw(func() {
		v.refreshActionsInUpdateDraw()
		v.notifyActionsChanged()
	})
}

func (v *View[U]) refreshActionsInUpdateDraw() {
	a := ui.NewKeyMap()

	for _, fn := range v.bindingKeysFn {
		fn(a)
	}

	v.actions = a
}

func (v *View[U]) keyboard(event *tcell.EventKey) *tcell.EventKey {
	if a, ok := v.actions[ui.AsKey(event)]; ok {
		return a.Action(event)
	}

	return event
}

func (v *View[U]) notifyActionsChanged() {
	for _, listener := range v.actionsListeners {
		listener.ViewActionsChanged(v)
	}
}
