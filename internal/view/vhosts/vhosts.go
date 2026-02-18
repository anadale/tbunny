package vhosts

import (
	"fmt"
	"slices"
	"strings"
	"tbunny/internal/cluster"
	"tbunny/internal/model"
	"tbunny/internal/ui"
	"tbunny/internal/utils"
	"tbunny/internal/view"

	"github.com/gdamore/tcell/v2"
	rabbithole "github.com/michaelklishin/rabbit-hole/v3"
	"github.com/rivo/tview"
)

type VHosts struct {
	view.ClusterAwareResourceView[*VHostResource]
}

func NewVHosts() view.ResourceView[*VHostResource] {
	v := VHosts{
		view.NewClusterAwareResourceTableView[*VHostResource]("Virtual hosts", view.NewLiveUpdateStrategy()),
	}

	v.SetResourceProvider(&v)
	v.AddBindingKeysFn(v.bindKeys)

	return &v
}

func (v *VHosts) Init(app model.App) error {
	if err := v.ClusterAwareResourceView.Init(app); err != nil {
		return err
	}

	v.Cluster().AddListener(v)

	return nil
}

func (v *VHosts) GetResources() ([]*VHostResource, error) {
	activeVhost := v.Cluster().ActiveVirtualHost()
	clusterVhosts := v.Cluster().VirtualHosts()

	vhosts := make([]rabbithole.VhostInfo, len(clusterVhosts)+1)
	vhosts[0] = v.createAllVhost()
	copy(vhosts[1:], clusterVhosts)

	slices.SortFunc(vhosts, func(i, j rabbithole.VhostInfo) int {
		return strings.Compare(i.Name, j.Name)
	})

	rows := utils.Map(vhosts, func(i rabbithole.VhostInfo) *VHostResource {
		return &VHostResource{i, i.Name == activeVhost}
	})

	return rows, nil
}

func (v *VHosts) createAllVhost() rabbithole.VhostInfo {
	return rabbithole.VhostInfo{
		Name: "",
	}
}
func (v *VHosts) GetColumns() []ui.TableColumn {
	c := []ui.TableColumn{
		{Name: "name", Title: "NAME", Expansion: 2},
		{Name: "msgReady", Title: "MR", Align: tview.AlignRight},
		{Name: "msgUnacked", Title: "MU", Align: tview.AlignRight},
		{Name: "msgTotal", Title: "MT", Align: tview.AlignRight},
		{Name: "msgRateReady", Title: "MR/S", Align: tview.AlignRight},
		{Name: "msgRateUnacked", Title: "MU/S", Align: tview.AlignRight},
		{Name: "msgRateDelivered", Title: "MT/S", Align: tview.AlignRight},
	}

	return c
}

func (v *VHosts) CanDeleteResources() bool {
	return true
}

func (v *VHosts) DeleteResource(resource *VHostResource) error {
	v.App().StatusLine().Info(fmt.Sprintf("Deleting virtual host %s", view.VhostDisplayName(resource.Name)))

	_, err := v.Cluster().DeleteVhost(resource.Name)
	if err != nil {
		return err
	}

	v.Cluster().Refresh()

	return nil
}

func (v *VHosts) ClusterVirtualHostsChanged(*cluster.Cluster) {
	v.RequestUpdate(view.PartialUpdate)
}

func (v *VHosts) ClusterActiveVirtualHostChanged(*cluster.Cluster) {
	v.RequestUpdate(view.PartialUpdate)
}

func (v *VHosts) bindKeys(km ui.KeyMap) {
	km.Add(tcell.KeyEnter, ui.NewKeyActionWithGroup("Switch virtual host", v.selectVHostCmd, false, 0))
	km.Add(ui.KeyC, ui.NewKeyAction("Create", v.createVHostCmd))
}

func (v *VHosts) selectVHostCmd(*tcell.EventKey) *tcell.EventKey {
	row, ok := v.GetSelectedResource()
	if !ok {
		return nil
	}

	vhost := row.Name

	v.App().StatusLine().Info(fmt.Sprintf("Switching to virtual host %s", view.VhostDisplayName(vhost)))
	v.Cluster().SetActiveVirtualHost(vhost)
	v.App().OpenClusterDefaultView()

	return nil
}

func (v *VHosts) createVHostCmd(*tcell.EventKey) *tcell.EventKey {
	ShowCreateVHostDialog(v.App(), v.createVHost)
	return nil
}

func (v *VHosts) createVHost(name, description, tags, queueType string, tracing bool) {
	v.App().StatusLine().Info(fmt.Sprintf("Creating virtual host %s", view.VhostDisplayName(name)))

	var vhostTags rabbithole.VhostTags

	for _, s := range strings.Split(tags, ",") {
		vhostTags = append(vhostTags, strings.TrimSpace(s))
	}

	_, err := v.Cluster().PutVhost(
		name,
		rabbithole.VhostSettings{
			Description:      description,
			Tags:             vhostTags,
			DefaultQueueType: strings.ToLower(queueType),
			Tracing:          tracing,
		})
	if err != nil {
		v.App().StatusLine().Error(fmt.Sprintf("Failed to create virtual host %s: %s", view.VhostDisplayName(name), err.Error()))
		return
	}

	v.Cluster().Refresh()

	v.App().DismissModal()
}
