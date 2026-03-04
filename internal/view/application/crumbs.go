package application

import (
	"fmt"
	"strings"
	"tbunny/internal/model"
	"tbunny/internal/skins"

	"github.com/rivo/tview"
)

type crumbs struct {
	*tview.TextView

	skin      *skins.Skin
	viewNames []string
}

func newCrumbs(app *App) *crumbs {
	c := &crumbs{
		TextView: tview.NewTextView(),
	}

	skins.AddListener(c)
	c.SkinChanged(skins.Current())

	c.SetTextAlign(tview.AlignLeft)
	c.SetBorderPadding(0, 0, 1, 1)
	c.SetDynamicColors(true)

	app.content.AddListener(c)

	return c
}

func (c *crumbs) SkinChanged(skin *skins.Skin) {
	c.skin = skin

	c.SetBackgroundColor(skin.BgColor())
	c.refresh()
}

func (c *crumbs) StackPushed(v model.View) {
	c.viewNames = append(c.viewNames, strings.ToLower(v.Name()))
	c.refresh()
}

func (c *crumbs) StackPopped(_, _ model.View) {
	c.viewNames = c.viewNames[:len(c.viewNames)-1]
	c.refresh()
}

func (c *crumbs) StackTop(model.View) {}

func (c *crumbs) refresh() {
	last, fgColor, bgColor, bbgColor := len(c.viewNames)-1, c.skin.Frame.Crumb.FgColor.Color(), c.skin.Frame.Crumb.BgColor.Color(), c.skin.Body.BgColor.Color()

	c.Clear()

	for i, crumb := range c.viewNames {
		if i == last {
			bgColor = c.skin.Frame.Crumb.ActiveColor.Color()
		}

		_, _ = fmt.Fprintf(
			c,
			"[%s:%s:b] <%s> [-:%s:-] ",
			fgColor,
			bgColor,
			strings.ToLower(crumb), //strings.ReplaceAll(strings.ToLower(crumb), " ", ""),
			bbgColor)
	}
}
