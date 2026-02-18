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

type TopicsPermissionsView struct {
	*view.ClusterAwareResourceTableView[*TopicPermissionsResource]

	user string
}

func NewTopicsPermissionsView(user string) *TopicsPermissionsView {
	v := TopicsPermissionsView{
		view.NewClusterAwareResourceTableView[*TopicPermissionsResource]("Topics permissions", view.NewLiveUpdateStrategy()),
		user,
	}

	v.SetPath(user)
	v.SetResourceProvider(&v)
	v.AddBindingKeysFn(v.bindKeys)

	return &v
}

func (v *TopicsPermissionsView) GetColumns() []ui.TableColumn {
	return []ui.TableColumn{
		{Name: "vhost", Title: "VIRTUAL HOST", Expansion: 2},
		{Name: "exchange", Title: "EXCHANGE", Expansion: 2},
		{Name: "write", Title: "WRITE", Align: tview.AlignRight},
		{Name: "read", Title: "READ", Align: tview.AlignRight},
	}
}

func (v *TopicsPermissionsView) GetResources() ([]*TopicPermissionsResource, error) {
	permissions, err := v.getPermissions()
	if err != nil {
		return nil, err
	}

	rows := utils.Map(permissions, func(p rabbithole.TopicPermissionInfo) *TopicPermissionsResource { return &TopicPermissionsResource{p} })

	return rows, nil
}

func (v *TopicsPermissionsView) getPermissions() ([]rabbithole.TopicPermissionInfo, error) {
	c := v.Cluster()

	slog.Debug("Fetching users's topics permissions", sl.Component, v.Name(), sl.Cluster, c.Name(), sl.User, v.user)

	permissions, err := v.Cluster().ListTopicPermissionsOf(v.user)
	if err != nil {
		slog.Error("Failed to fetch users's topics permissions", sl.Error, err, sl.Component, v.Name(), sl.Cluster, c.Name(), sl.User, v.user)
	}

	return permissions, err
}

func (v *TopicsPermissionsView) CanDeleteResources() bool {
	return true
}

func (v *TopicsPermissionsView) DeleteResource(resource *TopicPermissionsResource) error {
	_, err := v.Cluster().DeleteTopicPermissionsIn(resource.Vhost, resource.User, resource.Exchange)
	if err != nil {
		slog.Error("Failed to delete topic permissions", sl.Error, err, sl.Component, v.Name(), sl.Cluster, v.Cluster().Name(), sl.User, resource.User)
		return err
	}

	return nil
}

func (v *TopicsPermissionsView) bindKeys(km ui.KeyMap) {
	km.Add(ui.KeyC, ui.NewKeyAction("Create", v.createPermissionsCmd))
	km.Add(ui.KeyE, ui.NewKeyAction("Edit", v.editPermissionsCmd))
}

func (v *TopicsPermissionsView) createPermissionsCmd(*tcell.EventKey) *tcell.EventKey {
	vhosts := utils.Map(v.Cluster().VirtualHosts(), func(vh rabbithole.VhostInfo) string { return vh.Name })

	if len(vhosts) == 0 {
		v.App().StatusLine().Error("No vhosts available to create permissions for")
		return nil
	}

	ShowCreateTopicPermissionsDialog(v.App(), vhosts, v.fetchExchanges, func(vhost, exchange, write, read string) {
		v.App().StatusLine().Info(fmt.Sprintf("Creating permissions for exchange %s in vhost %s", exchange, vhost))
		v.setPermissions(vhost, exchange, write, read)
	})

	return nil
}

func (v *TopicsPermissionsView) editPermissionsCmd(*tcell.EventKey) *tcell.EventKey {
	r, ok := v.GetSelectedResource()
	if !ok {
		return nil
	}

	ShowEditTopicPermissionsDialog(v.App(), r.Vhost, view.ExchangeDisplayName(r.Exchange), r.Write, r.Read, func(vhost, exchange, write, read string) {
		v.App().StatusLine().Info(fmt.Sprintf("Updating permissions for exchange %s in vhost %s", exchange, vhost))
		v.setPermissions(vhost, exchange, write, read)
	})

	return nil
}

func (v *TopicsPermissionsView) setPermissions(vhost, exchange, write, read string) {
	if exchange == view.ExchangeDisplayName("") {
		exchange = ""
	}

	permissions := rabbithole.TopicPermissions{
		Exchange: exchange,
		Write:    write,
		Read:     read,
	}

	_, err := v.Cluster().UpdateTopicPermissionsIn(vhost, v.user, permissions)
	if err != nil {
		slog.Error("Failed to set topic permissions", sl.Error, err, sl.Component, v.Name(), sl.Cluster, v.Cluster().Name(), sl.User, v.user, sl.VirtualHost, vhost)
		v.App().StatusLine().Error(fmt.Sprintf("Failed to set topic permissions for exchange %s in vhost %s", view.ExchangeDisplayName(exchange), vhost))

		return
	}

	v.App().StatusLine().Clear()
	v.App().DismissModal()
	v.RequestUpdate(view.PartialUpdate)
}

func (v *TopicsPermissionsView) fetchExchanges(vhost string) []string {
	exchanges, err := v.Cluster().ListExchangesIn(vhost)
	if err != nil {
		return nil
	}

	return utils.Map(exchanges, func(e rabbithole.ExchangeInfo) string { return view.ExchangeDisplayName(e.Name) })
}
