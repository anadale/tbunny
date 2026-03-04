package application

import (
	"context"
	"fmt"
	"time"

	"github.com/gdamore/tcell/v2"
	"github.com/rivo/tview"
)

const (
	defaultStatusLineDelay                 = 3 * time.Second
	statusLineInfo         statusLineLevel = iota
	statusLineWarning
	statusLineError
)

type statusLineLevel int

type statusLineMessage struct {
	Level statusLineLevel
	Text  string
}

func newClearMessage() statusLineMessage { return statusLineMessage{} }

func (m statusLineMessage) IsClear() bool { return m.Text == "" }

type statusLine struct {
	*tview.TextView

	app     *App
	cancel  context.CancelFunc
	delay   time.Duration
	msgChan chan statusLineMessage
}

func newStatusLine(app *App, delay time.Duration) *statusLine {
	s := statusLine{
		TextView: tview.NewTextView(),
		app:      app,
		delay:    delay,
		msgChan:  make(chan statusLineMessage, 3),
	}

	s.SetTextColor(tcell.ColorAqua)
	s.SetDynamicColors(true)
	s.SetTextAlign(tview.AlignCenter)
	s.SetBorderPadding(0, 0, 1, 1)

	go s.watch()

	return &s
}

func (s *statusLine) Info(msg string) {
	s.setMessage(statusLineInfo, msg)
}

func (s *statusLine) Infof(format string, args ...any) {
	s.setMessage(statusLineInfo, fmt.Sprintf(format, args...))
}

func (s *statusLine) Warning(msg string) {
	s.setMessage(statusLineWarning, msg)
}

func (s *statusLine) Warningf(format string, args ...any) {
	s.setMessage(statusLineWarning, fmt.Sprintf(format, args...))
}

func (s *statusLine) Error(msg string) {
	s.setMessage(statusLineError, msg)
}

func (s *statusLine) Errorf(format string, args ...any) {
	s.setMessage(statusLineError, fmt.Sprintf(format, args...))
}

func (s *statusLine) Clear() {
	s.fireCleared()
}

func (s *statusLine) setMessage(level statusLineLevel, text string) {
	if s.cancel != nil {
		s.cancel()
		s.cancel = nil
	}

	s.msgChan <- statusLineMessage{Level: level, Text: text}

	ctx := context.Background()
	ctx, s.cancel = context.WithCancel(ctx)

	go s.fireClearedAfterDelay(ctx)
}

func (s *statusLine) watch() {
	for {
		msg := <-s.msgChan
		s.processMessage(msg)
	}
}

func (s *statusLine) processMessage(msg statusLineMessage) {
	s.app.QueueUpdateDraw(func() {
		if msg.IsClear() {
			s.TextView.Clear()
			return
		}
		s.SetTextColor(levelColor(msg.Level))
		s.SetText(msg.Text)
	})
}

func (s *statusLine) fireClearedAfterDelay(ctx context.Context) {
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

func (s *statusLine) fireCleared() {
	s.msgChan <- newClearMessage()
}

func levelColor(level statusLineLevel) tcell.Color {
	switch level {
	case statusLineError:
		return tcell.ColorOrangeRed
	case statusLineWarning:
		return tcell.ColorOrange
	default:
		return tcell.ColorNavajoWhite
	}
}
