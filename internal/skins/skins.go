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

// skinFile represents the structure of the Skin file.
type skinFile struct {
	Skin *Skin `yaml:"skin" json:"skin"`
}

var (
	//go:embed default-skin.yaml
	defaultSkinContent []byte
	// skin is the current skin.
	skin *Skin
	// listeners is a list of notification handlers.
	listeners []Listener
)

func init() {
	var file skinFile

	if err := yaml.Unmarshal(defaultSkinContent, &file); err == nil {
		skin = file.Skin
	} else {
		skin = newSkin()
	}

	updateStyles()
}

// Current returns the current skin.
func Current() *Skin {
	return skin
}

func AddListener(l Listener) {
	listeners = append(listeners, l)
}

func RemoveListener(l Listener) {
	for i, l2 := range listeners {
		if l2 == l {
			listeners = append(listeners[:i], listeners[i+1:]...)
			return
		}
	}
}

func updateStyles() {
	fgColor, bgColor := skin.FgColor(), skin.BgColor()

	tview.Styles.PrimitiveBackgroundColor = bgColor
	tview.Styles.ContrastBackgroundColor = bgColor
	tview.Styles.MoreContrastBackgroundColor = bgColor
	tview.Styles.PrimaryTextColor = fgColor
	tview.Styles.BorderColor = skin.Frame.Border.FgColor.Color()
	tview.Styles.TitleColor = skin.Frame.Title.FgColor.Color()
	tview.Styles.GraphicsColor = fgColor
	tview.Styles.SecondaryTextColor = fgColor
	tview.Styles.TertiaryTextColor = fgColor
	tview.Styles.ContrastSecondaryTextColor = fgColor

	notifySkinChanged()
}

func notifySkinChanged() {
	for _, l := range listeners {
		l.SkinChanged(skin)
	}
}
