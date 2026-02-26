package application

import (
	"tbunny/internal/skins"

	"github.com/rivo/tview"
)

const (
	clusterInfoWidth = 50
	//clusterInfoPad   = 15
)

type Header struct {
	*tview.Flex

	ClusterInfo *ClusterInfo
	Menu        *Menu

	app *App
}

func NewHeader(app *App) *Header {
	h := Header{
		Flex:        tview.NewFlex().SetDirection(tview.FlexColumn),
		ClusterInfo: NewClusterInfo(app),
		Menu:        NewMenu(app),
		app:         app,
	}

	h.AddItem(h.ClusterInfo, clusterInfoWidth, 1, false)
	h.AddItem(h.Menu, 0, 1, false)

	skins.AddListener(&h)
	h.SkinChanged(skins.Current())

	return &h
}

func (h *Header) SkinChanged(skin *skins.Skin) {
	bgColor := skin.BgColor()

	h.SetBackgroundColor(bgColor)
}
