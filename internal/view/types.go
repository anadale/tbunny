package view

import (
	"tbunny/internal/cluster"
	"tbunny/internal/model"
	"tbunny/internal/ui"
)

type BindKeysFunc func(ui.KeyMap)

type Resource interface {
	ui.TableRow

	GetName() string
	GetDisplayName() string
}

type ResourceProvider[R Resource] interface {
	GetResources() ([]R, error)
	GetColumns() []ui.TableColumn
	CanDeleteResources() bool
	DeleteResource(resource R) error
}

type ResourceView[R Resource] interface {
	model.View

	App() model.App

	SetPath(path string)

	SetResourceProvider(rp ResourceProvider[R])
	AddBindingKeysFn(fn ui.BindingKeysFn)
	SetEnterAction(title string, fn func(R))
	GetSelectedResource() (row R, ok bool)

	RequestUpdate(kind UpdateKind)
	RefreshActions()
}

type ClusterAwareResourceView[R Resource] interface {
	ResourceView[R]

	Cluster() *cluster.Cluster
}
