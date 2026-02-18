package ui

import (
	"context"
	"log/slog"
	"tbunny/internal/model"

	"github.com/gdamore/tcell/v2"
	"github.com/rivo/tview"
)

type application interface {
	QueueUpdateDraw(fn func())
}

type StatusLine struct {
	*tview.TextView

	app application
}

func NewStatusLine(app application) *StatusLine {
	s := StatusLine{
		TextView: tview.NewTextView(),
		app:      app,
	}

	s.SetTextColor(tcell.ColorAqua)
	s.SetDynamicColors(true)
	s.SetTextAlign(tview.AlignCenter)
	s.SetBorderPadding(0, 0, 1, 1)

	return &s
}

func (s *StatusLine) Watch(ctx context.Context, ch model.StatusLineChan) {
	defer slog.Debug("StatusLine stopped watching")

	for {
		select {
		case <-ctx.Done():
			return
		case msg := <-ch:
			s.SetMessage(msg)
		}
	}
}

func (s *StatusLine) SetMessage(msg model.StatusLineMessage) {
	fn := func() {
		if msg.Text == "" {
			s.Clear()
			return
		}
		s.SetTextColor(levelColor(msg.Level))
		s.SetText(msg.Text)
	}

	s.app.QueueUpdateDraw(fn)
}

func levelColor(level model.StatusLineLevel) tcell.Color {
	switch level {
	case model.StatusLineError:
		return tcell.ColorOrangeRed
	case model.StatusLineWarning:
		return tcell.ColorOrange
	default:
		return tcell.ColorNavajoWhite
	}
}
