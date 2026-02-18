package users

import (
	"fmt"
	"log/slog"
	"tbunny/internal/sl"
	"tbunny/internal/ui"
	"tbunny/internal/utils"
	"tbunny/internal/view"

	"github.com/gdamore/tcell/v2"
	rabbithole "github.com/michaelklishin/rabbit-hole/v3"
	"github.com/rivo/tview"
)

type VhostsPermissionsView struct {
	*view.ClusterAwareResourceTableView[*VhostPermissionsResource]

	user string
}

func NewVhostsPermissionsView(user string) *VhostsPermissionsView {
	v := VhostsPermissionsView{
		view.NewClusterAwareResourceTableView[*VhostPermissionsResource]("Vhost permissions", view.NewLiveUpdateStrategy()),
		user,
	}

	v.SetPath(user)
	v.SetResourceProvider(&v)
	v.AddBindingKeysFn(v.bindKeys)

	return &v
}

func (v *VhostsPermissionsView) GetColumns() []ui.TableColumn {
	return []ui.TableColumn{
		{Name: "vhost", Title: "VIRTUAL HOST", Expansion: 2},
		{Name: "configure", Title: "CONFIGURE", Align: tview.AlignRight},
		{Name: "write", Title: "WRITE", Align: tview.AlignRight},
		{Name: "read", Title: "READ", Align: tview.AlignRight},
	}
}

func (v *VhostsPermissionsView) GetResources() ([]*VhostPermissionsResource, error) {
	permissions, err := v.getPermissions()
	if err != nil {
		return nil, err
	}

	rows := utils.Map(permissions, func(p rabbithole.PermissionInfo) *VhostPermissionsResource { return &VhostPermissionsResource{p} })

	return rows, nil
}

func (v *VhostsPermissionsView) getPermissions() ([]rabbithole.PermissionInfo, error) {
	c := v.Cluster()

	slog.Debug("Fetching users permissions", sl.Component, v.Name(), sl.Cluster, c.Name(), sl.User, v.user)

	permissions, err := v.Cluster().ListPermissionsOf(v.user)
	if err != nil {
		slog.Error("Failed to fetch users permissions", sl.Error, err, sl.Component, v.Name(), sl.Cluster, c.Name(), sl.User, v.user)
	}

	return permissions, err
}

func (v *VhostsPermissionsView) CanDeleteResources() bool {
	return true
}

func (v *VhostsPermissionsView) DeleteResource(resource *VhostPermissionsResource) error {
	_, err := v.Cluster().ClearPermissionsIn(resource.Vhost, resource.User)
	if err != nil {
		slog.Error("Failed to clear permissions", sl.Error, err, sl.Component, v.Name(), sl.Cluster, v.Cluster().Name(), sl.User, resource.User)
		return err
	}

	return nil
}

func (v *VhostsPermissionsView) bindKeys(km ui.KeyMap) {
	km.Add(ui.KeyC, ui.NewKeyAction("Create", v.createPermissionsCmd))
	km.Add(ui.KeyE, ui.NewKeyAction("Edit", v.editPermissionsCmd))
}

func (v *VhostsPermissionsView) createPermissionsCmd(*tcell.EventKey) *tcell.EventKey {
	existingVhosts := make(map[string]bool)
	for _, row := range v.Ui().Rows() {
		existingVhosts[row.Vhost] = true
	}

	vhosts := v.Cluster().VirtualHosts()
	availableVhosts := make([]string, 0, len(vhosts))
	for _, vhost := range vhosts {
		if !existingVhosts[vhost.Name] {
			availableVhosts = append(availableVhosts, vhost.Name)
		}
	}

	if len(availableVhosts) > 0 {
		ShowCreateVhostPermissionsDialog(v.App(), availableVhosts, func(vhost, configure, write, read string) {
			v.App().StatusLine().Info(fmt.Sprintf("Creating permissions for vhost %s", vhost))
			v.setPermissions(vhost, configure, write, read)
		})
	} else {
		v.App().StatusLine().Error("No vhosts available to create permissions for")
	}

	return nil
}

func (v *VhostsPermissionsView) editPermissionsCmd(*tcell.EventKey) *tcell.EventKey {
	r, ok := v.GetSelectedResource()
	if !ok {
		return nil
	}

	ShowEditVhostPermissionsDialog(v.App(), r.Vhost, r.Configure, r.Write, r.Read, func(vhost, configure, write, read string) {
		v.App().StatusLine().Info(fmt.Sprintf("Updating permissions for vhost %s", vhost))
		v.setPermissions(vhost, configure, write, read)
	})

	return nil
}

func (v *VhostsPermissionsView) setPermissions(vhost, configure, write, read string) {
	permissions := rabbithole.Permissions{
		Configure: configure,
		Write:     write,
		Read:      read,
	}

	_, err := v.Cluster().UpdatePermissionsIn(vhost, v.user, permissions)
	if err != nil {
		slog.Error("Failed to set permissions", sl.Error, err, sl.Component, v.Name(), sl.Cluster, v.Cluster().Name(), sl.User, v.user, sl.VirtualHost, vhost)
		v.App().StatusLine().Error(fmt.Sprintf("Failed to set permissions for vhost %s", vhost))

		return
	}

	v.App().StatusLine().Clear()
	v.App().DismissModal()
	v.RequestUpdate(view.PartialUpdate)
}
