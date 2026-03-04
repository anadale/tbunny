package view

import (
	"tbunny/internal/cluster"
	"tbunny/internal/model"
	"tbunny/internal/ui"
)

// Resource represents a resource that can be displayed in a view.
type Resource interface {
	ui.TableRow

	// GetName returns the name of the resource.
	GetName() string

	// GetDisplayName returns the display name of the resource.
	GetDisplayName() string
}

// ResourceProvider provides resources for a view.
type ResourceProvider[R Resource] interface {
	// GetResources returns the resources for the view.
	GetResources() ([]R, error)
	// GetColumns returns the columns for the view.
	GetColumns() []ui.TableColumn
	// CanDeleteResources returns true if the view can delete resources.
	CanDeleteResources() bool
	// DeleteResource deletes the resource.
	DeleteResource(resource R) error
}

// ResourceView represents a view that displays resources.
type ResourceView[R Resource] interface {
	model.View
	model.Filterer

	// App returns the application.
	App() model.App

	// SetPath sets the path for the view.
	SetPath(path string)

	// SetResourceProvider sets the resource provider for the view.
	SetResourceProvider(rp ResourceProvider[R])
	// AddBindingKeysFn adds a function to bind keys to actions.
	AddBindingKeysFn(fn ui.BindingKeysFn)
	// SetEnterAction sets the action to be performed when the `Enter` key is pressed.
	SetEnterAction(title string, fn func(R))
	// GetSelectedResource returns the selected resource.
	GetSelectedResource() (row R, ok bool)

	// RequestUpdate requests an update for the view.
	RequestUpdate(kind UpdateKind)
	// RefreshActions refreshes the actions for the view.
	RefreshActions()
}

// ClusterAwareResourceView represents a view that displays resources and is aware of the RabbitMQ cluster.
type ClusterAwareResourceView[R Resource] interface {
	ResourceView[R]

	// Cluster returns the cluster.
	Cluster() *cluster.Cluster
}
