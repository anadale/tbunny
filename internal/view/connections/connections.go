package connections

import (
	"fmt"
	"log/slog"
	"tbunny/internal/sl"
	"tbunny/internal/ui"
	"tbunny/internal/utils"
	"tbunny/internal/view"
	"tbunny/internal/view/vhosts"

	"github.com/gdamore/tcell/v2"
	rabbithole "github.com/michaelklishin/rabbit-hole/v3"
	"github.com/rivo/tview"
)

type Connections struct {
	view.ClusterAwareResourceView[*ConnectionResource]

	wideMode bool
}

func NewConnections() *Connections {
	v := &Connections{
		vhosts.NewVHostExtender[*ConnectionResource](
			view.NewClusterAwareResourceTableView[*ConnectionResource]("Connections", view.NewLiveUpdateStrategy()),
		),
		false,
	}

	v.SetResourceProvider(v)
	v.AddBindingKeysFn(v.bindKeys)

	return v
}

func (v *Connections) GetResources() ([]*ConnectionResource, error) {
	exchanges, err := v.getConnections()
	if err != nil {
		return nil, fmt.Errorf("failed to list connections: %w", err)
	}

	vhost := v.Cluster().ActiveVirtualHost()

	rows := utils.FilterMap(
		exchanges,
		func(c rabbithole.ConnectionInfo) bool { return vhost == "" || c.Vhost == vhost },
		func(c rabbithole.ConnectionInfo) *ConnectionResource {
			return &ConnectionResource{c}
		})

	return rows, nil
}

func (v *Connections) getConnections() (exchanges []rabbithole.ConnectionInfo, err error) {
	c := v.Cluster()

	slog.Debug("Fetching connections", sl.Component, v.Name(), sl.Cluster, c.Name())

	connections, err := c.ListConnections()

	if err != nil {
		slog.Error("Failed to fetch connections", sl.Error, err, sl.Component, v.Name(), sl.Cluster, c.Name())
	}

	return connections, err
}

func (v *Connections) GetColumns() []ui.TableColumn {
	c := []ui.TableColumn{
		{Name: "client", Title: "CLIENT"},
		{Name: "name", Title: "NAME", Expansion: 1},
	}

	if v.wideMode {
		c = append(c, ui.TableColumn{Name: "username", Title: "USER NAME"})
	}

	c = append(c, ui.TableColumn{Name: "state", Title: "STATE"})

	if v.wideMode {
		c = append(c, []ui.TableColumn{
			{Name: "tls", Title: "SSL/TLS", Align: tview.AlignCenter},
			{Name: "protocol", Title: "PROTOCOL"},
		}...)
	}

	c = append(c, []ui.TableColumn{
		{Name: "channels", Title: "CHANNELS", Align: tview.AlignRight},
		{Name: "fromClient", Title: "FROM CLIENT", Align: tview.AlignRight},
		{Name: "toClient", Title: "TO CLIENT", Align: tview.AlignRight},
	}...)

	if v.Cluster().ActiveVirtualHost() == "" {
		c = append(c, ui.TableColumn{Name: "vhost", Title: "VHOST"})
	}

	if v.wideMode {
		c = append(c, ui.TableColumn{Name: "node", Title: "NODE"})
	}

	return c
}

func (v *Connections) CanDeleteResources() bool {
	return true
}

func (v *Connections) DeleteResource(resource *ConnectionResource) error {
	_, err := v.Cluster().CloseConnection(resource.Name)

	return err
}

func (v *Connections) bindKeys(km ui.KeyMap) {
	km.Add(tcell.KeyCtrlW, ui.NewKeyAction("Toggle wide mode", v.toggleWideModeCmd))
}

func (v *Connections) toggleWideModeCmd(*tcell.EventKey) *tcell.EventKey {
	v.wideMode = !v.wideMode

	v.RequestUpdate(view.FullUpdate)

	return nil
}
