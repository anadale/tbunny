package model

import (
	"tbunny/internal/cluster"
	"tbunny/internal/config"

	"github.com/rivo/tview"
)

// KeyMap represents a map of key bindings to actions.
type KeyMap interface {
	// MenuHints returns hints for the menu.
	MenuHints() Hints

	// HelpHints returns hints for the help view.
	HelpHints() Hints
}

// ViewActionsListener receives notifications when view actions change.
type ViewActionsListener interface {
	// ViewActionsChanged is called when the view's actions have changed.
	ViewActionsChanged(view View)
}

type View interface {
	// Primitive returns the view primitive.
	Primitive() tview.Primitive

	// Name returns the view name.
	Name() string

	// Init initializes the view lifetime.
	Init(app App) error

	// Start starts the view (makes it active).
	Start()

	// Stop stops the view (makes it inactive).
	Stop()

	// Actions returns the key actions for this view.
	Actions() KeyMap

	// AddActionsListener adds a listener for actions changes.
	AddActionsListener(listener ViewActionsListener)

	// RemoveActionsListener removes a listener for actions changes.
	RemoveActionsListener(listener ViewActionsListener)
}

type App interface {
	StatusLine() *StatusLine

	Cluster() *cluster.Cluster
	Config() *config.Config

	ClusterManager() *cluster.Manager
	ConfigManager() *config.Manager

	// Actions returns global application key actions.
	Actions() KeyMap

	DisableKeys()
	EnableKeys()
	QueueUpdateDraw(f func())
	AddView(v View) error
	ReplaceOpenViews(v View) error
	CloseLastView()
	OpenClusterDefaultView()
	ShowModal(modal tview.Primitive)
	DismissModal()
}
