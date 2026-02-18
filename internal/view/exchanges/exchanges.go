package exchanges

import (
	"fmt"
	"log/slog"
	"tbunny/internal/sl"
	"tbunny/internal/ui"
	"tbunny/internal/utils"
	"tbunny/internal/view"
	"tbunny/internal/view/bindings"
	"tbunny/internal/view/vhosts"

	"github.com/gdamore/tcell/v2"
	rabbithole "github.com/michaelklishin/rabbit-hole/v3"
	"github.com/rivo/tview"
)

type Exchanges struct {
	view.ClusterAwareResourceView[*ExchangeResource]
}

func NewExchanges() view.ClusterAwareResourceView[*ExchangeResource] {
	e := Exchanges{
		bindings.NewBindingsExtender[*ExchangeResource](
			vhosts.NewVHostExtender[*ExchangeResource](
				view.NewClusterAwareResourceTableView[*ExchangeResource]("Exchanges", view.NewLiveUpdateStrategy()),
			),
			bindings.ExchangeSubject),
	}

	e.SetResourceProvider(&e)
	e.AddBindingKeysFn(e.bindKeys)

	return &e
}

func (e *Exchanges) GetResources() ([]*ExchangeResource, error) {
	exchanges, err := e.getExchanges()
	if err != nil {
		return nil, fmt.Errorf("failed to list exchanges: %w", err)
	}

	rows := utils.Map(exchanges, func(i rabbithole.ExchangeInfo) *ExchangeResource {
		return &ExchangeResource{i}
	})

	return rows, nil
}

func (e *Exchanges) getExchanges() (exchanges []rabbithole.ExchangeInfo, err error) {
	c := e.Cluster()
	vhost := c.ActiveVirtualHost()

	slog.Debug("Fetching exchanges", sl.Component, e.Name(), sl.Cluster, c.Name(), sl.VirtualHost, vhost)

	if vhost == "" {
		exchanges, err = c.ListExchanges()
	} else {
		exchanges, err = c.ListExchangesIn(vhost)
	}

	if err != nil {
		slog.Error("Failed to fetch exchanges", sl.Error, err, sl.Component, e.Name(), sl.Cluster, c.Name(), sl.VirtualHost, vhost)
	}

	return exchanges, err
}

func (e *Exchanges) GetColumns() []ui.TableColumn {
	c := []ui.TableColumn{
		{Name: "name", Title: "NAME", Expansion: 2},
		{Name: "type", Title: "TYPE"},
		{Name: "features", Title: "FEATURES", Align: tview.AlignRight},
		{Name: "msgRateIn", Title: "MI/S", Align: tview.AlignRight},
		{Name: "msgRateOut", Title: "MO/S", Align: tview.AlignRight},
	}

	if e.Cluster().ActiveVirtualHost() == "" {
		c = append(c, []ui.TableColumn{{Name: "vhost", Title: "VHOST"}}...)
	}

	return c
}

func (e *Exchanges) CanDeleteResources() bool {
	return true
}

func (e *Exchanges) DeleteResource(resource *ExchangeResource) error {
	_, err := e.Cluster().DeleteExchange(resource.Vhost, resource.Name)

	return err
}

func (e *Exchanges) bindKeys(km ui.KeyMap) {
	km.Add(ui.KeyC, ui.NewKeyAction("Create", e.createExchangeCmd))
}

func (e *Exchanges) createExchangeCmd(*tcell.EventKey) *tcell.EventKey {
	ShowCreateExchangeDialog(e.App(), e.createExchange)

	return nil
}

func (e *Exchanges) createExchange(exchangeType, vhost, name string, durable, autoDelete bool, args map[string]any) {
	e.App().StatusLine().Info(fmt.Sprintf("Creating exchange %s", name))

	settings := rabbithole.ExchangeSettings{
		Type:       exchangeType,
		Durable:    durable,
		AutoDelete: autoDelete,
		Arguments:  args}

	_, err := e.Cluster().DeclareExchange(vhost, name, settings)
	if err != nil {
		e.App().StatusLine().Error(fmt.Sprintf("Failed to create exchange %s: %s", name, err.Error()))
		return
	}

	e.RequestUpdate(view.PartialUpdate)

	e.App().DismissModal()
}
