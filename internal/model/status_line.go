package model

import (
	"context"
	"time"
)

const (
	DefaultStatusLineDelay = 3 * time.Second

	StatusLineInfo StatusLineLevel = iota
	StatusLineWarning
	StatusLineError
)

type StatusLineLevel int

type StatusLineMessage struct {
	Level StatusLineLevel
	Text  string
}

func newClearMessage() StatusLineMessage { return StatusLineMessage{} }

func (m StatusLineMessage) IsClear() bool { return m.Text == "" }

type StatusLineChan chan StatusLineMessage

type StatusLineListener interface {
	StatusLineChanged(StatusLineLevel, string)
	StatusLineCleared()
}

type StatusLine struct {
	msg     StatusLineMessage
	cancel  context.CancelFunc
	delay   time.Duration
	msgChan StatusLineChan
}

func NewStatusLine(delay time.Duration) *StatusLine {
	return &StatusLine{
		delay:   delay,
		msgChan: make(StatusLineChan, 3),
	}
}

func (s *StatusLine) Channel() StatusLineChan {
	return s.msgChan
}

func (s *StatusLine) Info(msg string) {
	s.SetMessage(StatusLineInfo, msg)
}

func (s *StatusLine) Warning(msg string) {
	s.SetMessage(StatusLineWarning, msg)
}

func (s *StatusLine) Error(msg string) {
	s.SetMessage(StatusLineError, msg)
}

func (s *StatusLine) Clear() {
	s.fireCleared()
}

func (s *StatusLine) SetMessage(level StatusLineLevel, msg string) {
	if s.cancel != nil {
		s.cancel()
		s.cancel = nil
	}

	s.msg = StatusLineMessage{Level: level, Text: msg}
	s.fireChanged()

	ctx := context.Background()
	ctx, s.cancel = context.WithCancel(ctx)

	go s.refresh(ctx)
}

func (s *StatusLine) refresh(ctx context.Context) {
	for {
		select {
		case <-ctx.Done():
			return
		case <-time.After(s.delay):
			s.fireCleared()
			return
		}
	}
}

func (s *StatusLine) fireChanged() {
	s.msgChan <- s.msg
}

func (s *StatusLine) fireCleared() {
	s.msgChan <- newClearMessage()
}
