package model

import (
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

// View represents a view in the application.
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

// Filterer provides methods for filtering views.
type Filterer interface {
	// Filter applies a filter to the view.
	Filter(filter string)
	// Clear clears the filter and returns true if the filter was cleared.
	Clear() bool
}

// Navigator provides methods for navigating between views.
type Navigator interface {
	// AddView adds a view to the stack of open views.
	AddView(v View)
	// ReplaceOpenViews replaces the current open views with a new top-level view.
	ReplaceOpenViews(v View)
	// CloseLastView closes the last view in the stack (the one that is visible).
	CloseLastView()
	// OpenClusterDefaultView opens the default view for the current cluster.
	OpenClusterDefaultView()
}

// ModalManager manages modal windows (dialogs).
type ModalManager interface {
	// ShowModal shows a modal window.
	ShowModal(modal tview.Primitive)
	// DismissModal dismisses the modal window.
	DismissModal()
}

// App represents the application.
type App interface {
	Navigator
	ModalManager

	// StatusLine returns the status line.
	StatusLine() StatusLine

	// Actions returns global application key actions.
	Actions() KeyMap

	// DisableKeys disables keys processing.
	DisableKeys()
	// EnableKeys enables keys processing.
	EnableKeys()
	// QueueUpdateDraw queues a redrawing with the provided function.
	QueueUpdateDraw(f func())
	// OpenFilter opens a filter input and sets focus to it.
	OpenFilter(filterer Filterer)
}
