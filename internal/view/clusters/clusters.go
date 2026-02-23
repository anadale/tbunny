package clusters

import (
	"fmt"
	"slices"
	"strings"
	"tbunny/internal/cluster"
	"tbunny/internal/model"
	"tbunny/internal/ui"
	"tbunny/internal/view"

	"github.com/gdamore/tcell/v2"
)

type Clusters struct {
	view.ResourceView[*ClusterResource]
}

func NewClusters() model.View {
	c := Clusters{
		ResourceView: view.NewResourceTableView[*ClusterResource]("Clusters", view.NewManualUpdateStrategy()),
	}

	c.SetResourceProvider(&c)
	c.AddBindingKeysFn(c.bindKeys)
	c.SetEnterAction("Switch cluster", c.selectCluster)

	return &c
}

func (c *Clusters) GetResources() ([]*ClusterResource, error) {
	configClusters := c.App().ClusterManager().GetClusters()
	currentCluster := c.App().ClusterManager().GetActiveClusterName()

	rows := make([]*ClusterResource, 0, len(configClusters))

	for name, cfg := range configClusters {
		rows = append(rows, NewClusterResource(name, cfg, currentCluster == name))
	}

	slices.SortFunc(rows, func(a, b *ClusterResource) int { return strings.Compare(a.name, b.name) })

	return rows, nil
}

func (c *Clusters) GetColumns() []ui.TableColumn {
	return []ui.TableColumn{
		{Name: "name", Title: "NAME", Expansion: 2},
		{Name: "uri", Title: "URI", Expansion: 1},
		{Name: "username", Title: "USER"},
	}
}

func (c *Clusters) CanDeleteResources() bool {
	return true
}

func (c *Clusters) DeleteResource(cluster *ClusterResource) error {
	if cluster.active {
		return fmt.Errorf("unable to delete active cluster %s", cluster.name)
	}

	err := c.App().ClusterManager().Delete(cluster.name)
	if err != nil {
		return err
	}

	c.RequestUpdate(view.PartialUpdate)

	return nil
}

func (c *Clusters) bindKeys(km ui.KeyMap) {
	km.Add(ui.KeyA, ui.NewKeyAction("Add cluster", c.addClusterCmd))
}

func (c *Clusters) selectCluster(row *ClusterResource) {
	name := row.GetName()

	c.App().StatusLine().Info(fmt.Sprintf("Switching to cluster %s", name))
	c.switchToCluster(name)
}

func (c *Clusters) addClusterCmd(*tcell.EventKey) *tcell.EventKey {
	ShowAddClusterDialog(c.App(), c.addCluster)

	return nil
}

func (c *Clusters) addCluster(name string, parameters cluster.ConnectionParameters) {
	c.App().StatusLine().Info(fmt.Sprintf("Adding cluster %s...", name))

	clm := c.App().ClusterManager()

	err := clm.Create(name, parameters)
	if err != nil {
		c.App().StatusLine().Error(fmt.Sprintf("Failed to add cluster: %s", err))
		return
	}

	c.App().DismissModal()
	c.switchToCluster(name)
}

func (c *Clusters) switchToCluster(name string) {
	clm := c.App().ClusterManager()

	c.App().DisableKeys()

	go func() {
		_, err := clm.ConnectToCluster(name)
		c.App().EnableKeys()

		if err != nil {
			c.App().StatusLine().Error(err.Error())
			return
		}

		c.App().QueueUpdateDraw(func() {
			c.App().OpenClusterDefaultView()
		})
	}()
}
