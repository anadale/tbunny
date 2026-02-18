package bindings

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

const (
	QueueSubject    SubjectType = "queue"
	ExchangeSubject SubjectType = "exchange"
)

type SubjectType string

type Bindings struct {
	*view.ClusterAwareResourceTableView[*BindingResource]

	subjectType SubjectType
	subject     string
	vhost       string
}

func NewBindings(subjectType SubjectType, subject, vhost string) view.ClusterAwareResourceView[*BindingResource] {
	b := Bindings{
		ClusterAwareResourceTableView: view.NewClusterAwareResourceTableView[*BindingResource]("Bindings", view.NewLiveUpdateStrategy()),
		subjectType:                   subjectType,
		subject:                       subject,
		vhost:                         vhost,
	}

	b.SetPath(view.VhostDisplayName(vhost) + " ⏵ " + subject)

	b.SetResourceProvider(&b)
	b.AddBindingKeysFn(b.bindKeys)

	return &b
}

func (b *Bindings) GetResources() ([]*BindingResource, error) {
	bindings, err := b.getBindings()
	if err != nil {
		return nil, fmt.Errorf("failed to get bindings: %w", err)
	}

	rows := utils.Map(bindings, func(i rabbithole.BindingInfo) *BindingResource {
		return &BindingResource{i}
	})

	return rows, nil
}

func (b *Bindings) getBindings() (bindings []rabbithole.BindingInfo, err error) {
	c := b.Cluster()

	if b.subjectType == QueueSubject {
		slog.Debug("Fetching queue bindings", sl.Component, b.Name(), sl.Cluster, c.Name(), sl.VirtualHost, b.vhost, sl.Resource, b.subject)
		bindings, err = c.ListQueueBindings(b.vhost, b.subject)
	} else {
		slog.Debug("Fetching exchange bindings", sl.Component, b.Name(), sl.Cluster, c.Name(), sl.VirtualHost, b.vhost, sl.Resource, b.subject)
		bindings, err = c.ListExchangeBindingsWithSource(b.vhost, b.subject)
	}

	if err != nil {
		slog.Error("Failed to fetch bindings", sl.Error, err, sl.Component, b.Name(), sl.Cluster, c.Name(), sl.VirtualHost, b.vhost, sl.Resource, b.subject)
	}

	return bindings, err
}

func (b *Bindings) GetColumns() []ui.TableColumn {
	c := make([]ui.TableColumn, 0, 4)

	if b.subjectType == QueueSubject {
		c = append(c, ui.TableColumn{Name: "source", Title: "FROM", Expansion: 2})
	} else {
		c = append(c, ui.TableColumn{Name: "destination", Title: "TO", Expansion: 2})
		c = append(c, ui.TableColumn{Name: "destinationType", Title: "TYPE"})
	}

	c = append(c, ui.TableColumn{Name: "routingKey", Title: "ROUTING KEY"})
	c = append(c, ui.TableColumn{Name: "features", Title: "FEATURES", Align: tview.AlignRight})

	return c
}

func (b *Bindings) CanDeleteResources() bool {
	return true
}

func (b *Bindings) DeleteResource(resource *BindingResource) error {
	_, err := b.Cluster().DeleteBinding(resource.Vhost, resource.BindingInfo)

	return err
}

func (b *Bindings) bindKeys(km ui.KeyMap) {
	km.Add(ui.KeyC, ui.NewKeyAction("Create", b.createBindingCmd))
}

func (b *Bindings) createBindingCmd(*tcell.EventKey) *tcell.EventKey {
	ShowCreateBindingDialog(b.App(), b.subjectType, b.fetchExchanges, b.fetchQueues, b.createBinding)

	return nil
}

func (b *Bindings) createBinding(otherType SubjectType, otherName, routingKey string, args map[string]any) {
	b.App().StatusLine().Info("Creating binding...")

	info := rabbithole.BindingInfo{
		RoutingKey: routingKey,
		Arguments:  args,
	}

	if b.subjectType == QueueSubject {
		info.Source = otherName
		info.Destination = b.subject
		info.DestinationType = string(b.subjectType)
	} else {
		info.Source = b.subject
		info.Destination = otherName
		info.DestinationType = string(otherType)
	}

	_, err := b.Cluster().DeclareBinding(b.vhost, info)
	if err != nil {
		b.App().StatusLine().Error(fmt.Sprintf("Failed to create binding: %s", err.Error()))
		return
	}

	b.RequestUpdate(view.PartialUpdate)
	b.App().DismissModal()
}

func (b *Bindings) fetchExchanges() ([]string, error) {
	items, err := b.Cluster().ListExchangesIn(b.vhost)
	if err != nil {
		return nil, err
	}

	exchanges := utils.FilterMap(items, func(item rabbithole.ExchangeInfo) bool { return item.Name != "" }, func(item rabbithole.ExchangeInfo) string { return item.Name })

	return exchanges, nil
}

func (b *Bindings) fetchQueues() ([]string, error) {
	items, err := b.Cluster().ListQueuesIn(b.vhost)
	if err != nil {
		return nil, err
	}

	queues := utils.Map(items, func(item rabbithole.QueueInfo) string { return item.Name })

	return queues, nil
}
