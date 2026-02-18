package view

// UpdateStrategy defines the view update strategy
type UpdateStrategy interface {
	// Start starts the strategy, performs initial update
	Start()
	// Stop stops the strategy
	Stop()
	// RequestUpdate requests an update
	RequestUpdate(kind UpdateKind)
	// SetUpdateFn sets the update function
	SetUpdateFn(fn func(kind UpdateKind))
	// Pause pauses automatic updates (for timer-based strategies)
	Pause()
	// Resume resumes automatic updates (for timer-based strategies)
	Resume()
	// Name returns the component name for logging
	Name() string
	// SetName sets the component name
	SetName(name string)
}
