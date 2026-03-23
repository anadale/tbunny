package nodes

import (
	"log/slog"
	"tbunny/internal/model"
	"tbunny/internal/sl"
	"tbunny/internal/ui"
	"tbunny/internal/utils"
	"tbunny/internal/view"

	rabbithole "github.com/michaelklishin/rabbit-hole/v3"
)

// View is a cluster-aware resource table view listing all RabbitMQ nodes.
type View struct {
	view.ClusterAwareResourceView[*Resource]
}

// NewView creates and returns a new Nodes list view.
func NewView() model.View {
	v := &View{
		view.NewClusterAwareResourceTableView[*Resource]("Nodes", view.NewLiveUpdateStrategy()),
	}

	v.SetResourceProvider(v)
	v.SetEnterAction("Show details", v.showDetails)

	return v
}

func (v *View) GetColumns() []ui.TableColumn {
	return []ui.TableColumn{
		{Name: "name", Title: "NAME", Expansion: 2},
		{Name: "type", Title: "TYPE"},
		{Name: "running", Title: "RUNNING"},
		{Name: "mem_used", Title: "MEMORY"},
		{Name: "mem_alarm", Title: "MEM ALARM"},
		{Name: "disk_free", Title: "DISK FREE"},
		{Name: "disk_alarm", Title: "DISK ALARM"},
		{Name: "fd_used", Title: "FILE DESC"},
		{Name: "proc_used", Title: "PROCESSES"},
		{Name: "uptime", Title: "UPTIME"},
	}
}

func (v *View) GetResources() ([]*Resource, error) {
	nodes, err := v.getNodes()
	if err != nil {
		return nil, err
	}

	rows := utils.Map(nodes, func(n rabbithole.NodeInfo) *Resource { return &Resource{n} })

	return rows, nil
}

func (v *View) getNodes() ([]rabbithole.NodeInfo, error) {
	c := v.Cluster()

	slog.Debug("Fetching nodes", sl.Component, v.Name(), sl.Cluster, c.Name())

	nodes, err := c.ListNodes()
	if err != nil {
		slog.Error("Failed to fetch nodes", sl.Error, err, sl.Component, v.Name(), sl.Cluster, c.Name())
	}

	return nodes, err
}

func (v *View) CanDeleteResources() bool {
	return false
}

func (v *View) DeleteResource(_ *Resource) error {
	return nil
}

func (v *View) showDetails(node *Resource) {
	details := NewNodeDetails(node.Name)

	v.App().AddView(details)
}
