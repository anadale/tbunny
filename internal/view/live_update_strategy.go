package view

import (
	"fmt"
	"log/slog"
	"tbunny/internal/sl"
	"time"
)

// LiveUpdateStrategy - strategy for updating by timer and on request
type LiveUpdateStrategy struct {
	name           string
	updateFn       func(kind UpdateKind)
	updateChan     chan UpdateKind
	updateInterval time.Duration
	initialized    bool
	paused         bool
}

func NewLiveUpdateStrategy() *LiveUpdateStrategy {
	return &LiveUpdateStrategy{
		updateInterval: 5 * time.Second,
	}
}

func (s *LiveUpdateStrategy) SetName(name string) {
	s.name = name
}

func (s *LiveUpdateStrategy) Name() string {
	return s.name
}

func (s *LiveUpdateStrategy) SetUpdateFn(fn func(kind UpdateKind)) {
	s.updateFn = fn
}

func (s *LiveUpdateStrategy) SetUpdateInterval(interval time.Duration) {
	s.updateInterval = interval
}

func (s *LiveUpdateStrategy) Start() {
	if !s.paused {
		s.startUpdateLoop()
	}

	if !s.initialized {
		s.RequestUpdate(FullUpdate)
		s.initialized = true
	} else if !s.paused {
		s.RequestUpdate(PartialUpdate)
	}
}

func (s *LiveUpdateStrategy) Pause() {
	s.paused = true
	s.stopUpdateLoop()
}

func (s *LiveUpdateStrategy) Resume() {
	s.paused = false
	s.startUpdateLoop()
	s.RequestUpdate(PartialUpdate)
}

func (s *LiveUpdateStrategy) IsPaused() bool {
	return s.paused
}

func (s *LiveUpdateStrategy) Stop() {
	s.stopUpdateLoop()
}

func (s *LiveUpdateStrategy) RequestUpdate(kind UpdateKind) {
	if s.updateChan == nil {
		if s.updateFn == nil {
			panic("No update function set")
		}

		slog.Debug("Performing update without update loop", sl.Component, s.name, "kind", kind.String())
		s.updateFn(kind)

		return
	}

	select {
	case s.updateChan <- kind:
		slog.Debug("Update requested", sl.Component, s.name, "kind", kind.String())
	default:
		// Update is already requested
	}
}

func (s *LiveUpdateStrategy) startUpdateLoop() {
	if s.updateChan != nil {
		return
	}

	slog.Debug("Starting update loop...", sl.Component, s.name)

	s.updateChan = make(chan UpdateKind, 1)

	go s.updateLoop(s.updateChan)
}

func (s *LiveUpdateStrategy) stopUpdateLoop() {
	if s.updateChan != nil {
		close(s.updateChan)
		s.updateChan = nil

		slog.Debug("Stopping update loop...", sl.Component, s.name)
	}
}

func (s *LiveUpdateStrategy) updateLoop(ch <-chan UpdateKind) {
	round := 0

	slog.Debug("Update loop started", sl.Component, s.name)

	for {
		kind := PartialUpdate

		select {
		case k, ok := <-ch:
			if !ok {
				slog.Debug("Update loop stopped", sl.Component, s.name)
				return
			}

			slog.Debug("Update requested received", sl.Component, s.name, "kind", k.String())

			kind = k
		case <-time.After(s.updateInterval):
		}

		round++
		slog.Debug(fmt.Sprintf("Updating, round %d", round), sl.Component, s.name, "kind", kind.String())

		if s.updateFn == nil {
			panic("No update function set")
		}

		s.updateFn(kind)
	}
}
