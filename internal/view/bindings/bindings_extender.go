package bindings

import (
	"tbunny/internal/ui"
	"tbunny/internal/view"

	"github.com/gdamore/tcell/v2"
)

type ResourceWithBindingDetails interface {
	view.Resource

	GetBindingDetails() BindingDetails
}

type Extender[R ResourceWithBindingDetails] struct {
	view.ClusterAwareResourceView[R]

	subjectType SubjectType
}

type BindingDetails struct {
	Subject string
	Vhost   string
}

type BindingDetailsProvider interface {
	GetBindingDetails() BindingDetails
}

func NewBindingsExtender[R ResourceWithBindingDetails](r view.ClusterAwareResourceView[R], subjectType SubjectType) view.ClusterAwareResourceView[R] {
	e := Extender[R]{
		ClusterAwareResourceView: r,
		subjectType:              subjectType,
	}

	e.AddBindingKeysFn(e.bindKeys)

	return &e
}

func (e *Extender[R]) bindKeys(keyMap ui.KeyMap) {
	if e.Cluster().IsAvailable() {
		keyMap.Add(ui.KeyB, ui.NewKeyAction("Bindings", e.showBindingsCmd))
	}
}

func (e *Extender[R]) showBindingsCmd(*tcell.EventKey) *tcell.EventKey {
	row, ok := e.GetSelectedResource()
	if !ok {
		return nil
	}

	details := row.GetBindingDetails()
	bindings := NewBindings(e.subjectType, details.Subject, details.Vhost)

	e.App().AddView(bindings)

	return nil
}
