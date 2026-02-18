package skins

import (
	_ "embed"

	"github.com/rivo/tview"
	"gopkg.in/yaml.v3"
)

type Listener interface {
	// SkinChanged notifies the listener that the skin has changed.
	SkinChanged(*Skin)
}

var (
	//go:embed default-skin.yaml
	defaultSkinContent []byte
)

type Manager struct {
	Skin      *Skin `yaml:"skin" json:"skin"`
	listeners []Listener
}

func NewSkinManager() *Manager {
	var s Manager

	if err := yaml.Unmarshal(defaultSkinContent, &s); err != nil {
		s = Manager{Skin: newSkin()}
	}

	s.UpdateStyles()

	return &s
}

func (s *Manager) AddListener(l Listener) {
	s.listeners = append(s.listeners, l)
}

func (s *Manager) RemoveListener(l Listener) {
	for i, l2 := range s.listeners {
		if l2 == l {
			s.listeners = append(s.listeners[:i], s.listeners[i+1:]...)
			return
		}
	}
}

func (s *Manager) UpdateStyles() {
	fgColor, bgColor := s.Skin.FgColor(), s.Skin.BgColor()

	tview.Styles.PrimitiveBackgroundColor = bgColor
	tview.Styles.ContrastBackgroundColor = bgColor
	tview.Styles.MoreContrastBackgroundColor = bgColor
	tview.Styles.PrimaryTextColor = fgColor
	tview.Styles.BorderColor = s.Skin.Frame.Border.FgColor.Color()
	tview.Styles.TitleColor = s.Skin.Frame.Title.FgColor.Color()
	tview.Styles.GraphicsColor = fgColor
	tview.Styles.SecondaryTextColor = fgColor
	tview.Styles.TertiaryTextColor = fgColor
	tview.Styles.ContrastSecondaryTextColor = fgColor

	s.notifySkinChanged()
}

func (s *Manager) notifySkinChanged() {
	for _, l := range s.listeners {
		l.SkinChanged(s.Skin)
	}
}
