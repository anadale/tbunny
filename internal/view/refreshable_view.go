package view

import (
	"tbunny/internal/ui"

	"github.com/gdamore/tcell/v2"
)

type RefreshableView[U UiComponent] struct {
	*View[U]

	strategy UpdateStrategy
}

func NewRefreshableView[U UiComponent](name string, ui U, strategy UpdateStrategy) *RefreshableView[U] {
	c := RefreshableView[U]{
		View:     NewView(name, ui),
		strategy: strategy,
	}

	c.strategy.SetName(name)
	c.AddBindingKeysFn(c.bindKeys)

	return &c
}

func (v *RefreshableView[U]) SetUpdateFn(fn func(kind UpdateKind)) {
	v.strategy.SetUpdateFn(fn)
}

func (v *RefreshableView[U]) Strategy() UpdateStrategy {
	return v.strategy
}

func (v *RefreshableView[U]) Start() {
	v.View.Start()
	v.strategy.Start()
}

func (v *RefreshableView[U]) Stop() {
	v.strategy.Stop()
	v.View.Stop()
}

func (v *RefreshableView[U]) RequestUpdate(kind UpdateKind) {
	v.strategy.RequestUpdate(kind)
}

func (v *RefreshableView[U]) bindKeys(km ui.KeyMap) {
	km.Add(tcell.KeyCtrlR, ui.NewKeyActionWithGroup("Refresh", v.refreshCmd, false, 100))
}

func (v *RefreshableView[U]) refreshCmd(*tcell.EventKey) *tcell.EventKey {
	v.App().StatusLine().Info("Refreshing " + v.Name() + "...")
	v.RequestUpdate(PartialUpdate)

	return nil
}
