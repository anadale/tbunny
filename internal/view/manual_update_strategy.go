package view

import (
	"log/slog"
	"tbunny/internal/sl"
)

// ManualUpdateStrategy - strategy for updating only on request (without a timer)
type ManualUpdateStrategy struct {
	name        string
	updateFn    func(kind UpdateKind)
	initialized bool
}

func NewManualUpdateStrategy() *ManualUpdateStrategy {
	return &ManualUpdateStrategy{}
}

func (s *ManualUpdateStrategy) SetName(name string) {
	s.name = name
}

func (s *ManualUpdateStrategy) Name() string {
	return s.name
}

func (s *ManualUpdateStrategy) SetUpdateFn(fn func(kind UpdateKind)) {
	s.updateFn = fn
}

func (s *ManualUpdateStrategy) Start() {
	if !s.initialized {
		s.RequestUpdate(FullUpdate)
		s.initialized = true
	}
}

func (s *ManualUpdateStrategy) Stop() {}

func (s *ManualUpdateStrategy) Pause() {}

func (s *ManualUpdateStrategy) Resume() {}

func (s *ManualUpdateStrategy) RequestUpdate(kind UpdateKind) {
	if s.updateFn == nil {
		panic("No update function set")
	}

	slog.Debug("Performing manual update", sl.Component, s.name, "kind", kind.String())
	s.updateFn(kind)
}
