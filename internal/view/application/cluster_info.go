package application

import (
	"tbunny/internal/cluster"
	"tbunny/internal/skins"

	"github.com/gdamore/tcell/v2"
	"github.com/rivo/tview"
)

const NAValue = "N/A"

type ClusterInfo struct {
	*tview.Table

	cluster *cluster.Cluster
	app     *App
}

func NewClusterInfo(app *App) *ClusterInfo {
	c := ClusterInfo{
		Table: tview.NewTable(),
		app:   app,
	}

	c.SetBorderPadding(0, 0, 1, 0)
	c.layout()
	c.reset()

	c.ClusterChanged(c.app.ClusterManager().Cluster())
	c.SkinChanged(c.app.SkinManager().Skin)

	app.ClusterManager().AddListener(&c)
	app.SkinManager().AddListener(&c)

	return &c
}

func (c *ClusterInfo) ClusterChanged(cluster *cluster.Cluster) {
	if c.cluster != nil {
		c.cluster.RemoveListener(c)
	}

	c.cluster = cluster

	if c.cluster != nil {
		c.cluster.AddListener(c)
	}

	c.app.QueueUpdateDraw(func() {
		if cluster != nil {
			c.update()
		} else {
			c.reset()
		}
	})
}

func (c *ClusterInfo) ClusterInformationChanged(*cluster.Cluster) {
	c.app.QueueUpdateDraw(c.update)
}

func (c *ClusterInfo) SkinChanged(skin *skins.Skin) {
	c.SetBackgroundColor(skin.BgColor())
	c.updateStyles()
}

func (c *ClusterInfo) update() {
	info := c.cluster.Information()

	row := c.setCell(0, info.Name)
	row = c.setCell(row, info.ClusterName)
	row = c.setCell(row, c.cluster.Username())
	row = c.setCell(row, info.RabbitMQVersion)
	row = c.setCell(row, info.ManagementVersion)
	row = c.setCell(row, info.ErlangVersion)
	row = c.setCell(row, c.app.Version)
}

func (c *ClusterInfo) reset() {
	row := c.setCell(0, NAValue)
	row = c.setCell(row, NAValue)
	row = c.setCell(row, NAValue)
	row = c.setCell(row, NAValue)
	row = c.setCell(row, NAValue)
	row = c.setCell(row, NAValue)
	row = c.setCell(row, c.app.Version)
}

func (c *ClusterInfo) layout() {
	for row, section := range []string{"Cluster", "Name", "User", "Rabbit Rev", "Mgmt Rev", "Erlang Rev", "TBunny Rev"} {
		c.SetCell(row, 0, c.sectionCell(section))
		c.SetCell(row, 1, c.infoCell(NAValue))
	}
}

func (c *ClusterInfo) sectionCell(title string) *tview.TableCell {
	cell := tview.NewTableCell(title + ":")

	cell.SetAlign(tview.AlignLeft)

	return cell
}

func (c *ClusterInfo) infoCell(title string) *tview.TableCell {
	cell := tview.NewTableCell(title)

	cell.SetExpansion(2)
	cell.SetAlign(tview.AlignLeft)

	return cell
}

func (c *ClusterInfo) setCell(row int, s string) int {
	if s == "" {
		s = NAValue
	}

	c.GetCell(row, 1).SetText(s)

	return row + 1
}

func (c *ClusterInfo) updateStyles() {
	skin := c.app.SkinManager().Skin

	sectionColor := skin.Info.SectionColor.Color()
	fgColor := skin.Info.FgColor.Color()
	bgColor := skin.BgColor()

	var style tcell.Style
	style = style.Bold(true)
	style = style.Foreground(fgColor)
	style = style.Background(bgColor)

	for row := range c.GetRowCount() {
		c.GetCell(row, 0).SetTextColor(sectionColor)
		c.GetCell(row, 0).SetBackgroundColor(bgColor)
		c.GetCell(row, 1).SetStyle(style)
	}
}
