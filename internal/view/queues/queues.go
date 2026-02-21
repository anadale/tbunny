package queues

import (
	"fmt"
	"log/slog"
	"net/url"
	"slices"
	"tbunny/internal/rmq"
	"tbunny/internal/sl"
	"tbunny/internal/ui"
	"tbunny/internal/ui/dialog"
	"tbunny/internal/utils"
	"tbunny/internal/view"
	"tbunny/internal/view/bindings"
	"tbunny/internal/view/vhosts"

	"github.com/gdamore/tcell/v2"
	rabbithole "github.com/michaelklishin/rabbit-hole/v3"
	"github.com/rivo/tview"
)

type Queues struct {
	view.ClusterAwareResourceView[*QueueResource]
}

func NewQueues() view.ResourceView[*QueueResource] {
	q := Queues{
		bindings.NewBindingsExtender[*QueueResource](
			vhosts.NewVHostExtender[*QueueResource](
				view.NewClusterAwareResourceTableView[*QueueResource]("Queues", view.NewLiveUpdateStrategy()),
			),
			bindings.QueueSubject),
	}

	q.SetResourceProvider(&q)
	q.AddBindingKeysFn(q.bindKeys)

	return &q
}

func (q *Queues) GetResources() ([]*QueueResource, error) {
	queues, err := q.getQueues()
	if err != nil {
		return nil, fmt.Errorf("failed to list queues: %w", err)
	}

	rows := utils.Map(queues, func(i rabbithole.QueueInfo) *QueueResource {
		return &QueueResource{i}
	})

	return rows, nil
}

func (q *Queues) getQueues() (queues []rabbithole.QueueInfo, err error) {
	c := q.Cluster()
	vhost := c.ActiveVirtualHost()

	slog.Debug("Fetching queues", sl.Component, q.Name(), sl.Cluster, c.Name(), sl.VirtualHost, vhost)

	if vhost == "" {
		queues, err = c.ListQueues()
	} else {
		queues, err = c.ListQueuesIn(vhost)
	}

	if err != nil {
		slog.Error("Failed to fetch queues", sl.Error, err, sl.Component, q.Name(), sl.Cluster, c.Name(), sl.VirtualHost, vhost)
	}

	return queues, err
}

func (q *Queues) GetColumns() []ui.TableColumn {
	c := []ui.TableColumn{
		{Name: "name", Title: "NAME", Expansion: 2},
		{Name: "type", Title: "TYPE"},
		{Name: "features", Title: "FEATURES", Align: tview.AlignRight},
		{Name: "msgReady", Title: "MR", Align: tview.AlignRight},
		{Name: "msgUnacked", Title: "MU", Align: tview.AlignRight},
		{Name: "msgTotal", Title: "MT", Align: tview.AlignRight},
		{Name: "msgRateIn", Title: "MR/S", Align: tview.AlignRight},
		{Name: "msgRateDelivered", Title: "MD/S", Align: tview.AlignRight},
		{Name: "msgRateAcked", Title: "MA/S", Align: tview.AlignRight},
		//{Name: "node", Title: "NODE", Expansion: 3},
	}

	if q.Cluster().ActiveVirtualHost() == "" {
		c = append(c, []ui.TableColumn{{Name: "vhost", Title: "VHOST"}}...)
	}

	return c
}

func (q *Queues) CanDeleteResources() bool {
	return true
}

func (q *Queues) DeleteResource(resource *QueueResource) error {
	_, err := q.Cluster().DeleteQueue(resource.Vhost, resource.Name, rabbithole.QueueDeleteOptions{})

	return err
}

func (q *Queues) bindKeys(km ui.KeyMap) {
	if q.Cluster().IsAvailable() {
		km.Add(tcell.KeyEnter, ui.NewHiddenKeyAction("Show details", q.showDetailsCmd))
		km.Add(ui.KeyC, ui.NewKeyAction("Create", q.createQueueCmd))
		km.Add(ui.KeyM, ui.NewKeyAction("Get messages", q.getMessagesCmd))
		km.Add(ui.KeyP, ui.NewKeyAction("Publish message", q.publishMessageCmd))
		km.Add(ui.KeyV, ui.NewKeyAction("Move messages", q.moveMessagesCmd))
		km.Add(tcell.KeyCtrlP, ui.NewKeyAction("Purge", q.purgeQueueCmd))
	}
}

func (q *Queues) getMessagesCmd(*tcell.EventKey) *tcell.EventKey {
	if queue, ok := q.GetSelectedResource(); ok {
		ShowGetMessagesDialog(q.App(), queue, q.getMessages)
	}

	return nil
}

func (q *Queues) getMessages(queue *QueueResource, ackMode rmq.AckMode, encoding rmq.RequestedMessageEncoding, count int) {
	q.App().StatusLine().Info(fmt.Sprintf("Getting messages from %s...", queue.GetDisplayName()))

	messages, err := q.Cluster().GetQueueMessages(queue.Vhost, queue.Name, ackMode, encoding, count)
	if err != nil {
		q.App().StatusLine().Error(fmt.Sprintf("Failed to get messages: %s", err))
		return
	}

	messagesView := NewMessages(messages, queue.Name, queue.Vhost)

	err = q.App().AddView(messagesView)
	if err != nil {
		q.App().StatusLine().Error(fmt.Sprintf("Failed to load messages: %s", err))
	}
}

func (q *Queues) showDetailsCmd(*tcell.EventKey) *tcell.EventKey {
	queue, ok := q.GetSelectedResource()
	if !ok {
		return nil
	}

	details := NewQueueDetails(queue.Name, queue.Vhost)

	err := q.App().AddView(details)
	if err != nil {
		q.App().StatusLine().Error(fmt.Sprintf("Failed to load queue details: %s", err))
	}

	return nil
}

func (q *Queues) createQueueCmd(*tcell.EventKey) *tcell.EventKey {
	ShowCreateQueueDialog(q.App(), q.createQueue)

	return nil
}

func (q *Queues) createQueue(queueType, vhost, name string, durable bool, args map[string]any) {
	q.App().StatusLine().Info(fmt.Sprintf("Creating queue %s", name))

	settings := rabbithole.QueueSettings{Durable: durable}
	if queueType != "" {
		if args == nil {
			args = map[string]any{"x-queue-type": queueType}
		} else {
			args["x-queue-type"] = queueType
		}
	}

	if args != nil {
		settings.Arguments = args
	}

	_, err := q.Cluster().DeclareQueue(vhost, name, settings)
	if err != nil {
		q.App().StatusLine().Error(fmt.Sprintf("Failed to create queue %s: %s", name, err.Error()))
		return
	}

	q.RequestUpdate(view.PartialUpdate)

	q.App().DismissModal()
}

func (q *Queues) purgeQueueCmd(*tcell.EventKey) *tcell.EventKey {
	queue, ok := q.GetSelectedResource()
	if !ok {
		return nil
	}

	msg := fmt.Sprintf("Purge %s?", queue.GetDisplayName())
	dialogSkin := q.App().Skin().Dialog

	modal := dialog.CreateConfirmDialog(
		&dialogSkin,
		"Confirm Purge",
		msg,
		func() {
			q.App().StatusLine().Info(fmt.Sprintf("Purging %s...", queue.GetDisplayName()))

			_, err := q.Cluster().PurgeQueue(queue.Vhost, queue.Name)
			if err != nil {
				q.App().StatusLine().Error(err.Error())
			} else {
				q.RequestUpdate(view.PartialUpdate)
			}
		},
		func() {
			q.App().DismissModal()
		})

	q.App().ShowModal(modal)

	return nil
}

func (q *Queues) publishMessageCmd(*tcell.EventKey) *tcell.EventKey {
	if queue, ok := q.GetSelectedResource(); ok {
		ShowPublishMessageDialog(q.App(), queue.Vhost, queue.Name, q.publishMessage)
	}

	return nil
}

func (q *Queues) publishMessage(vhost, queue string, props map[string]any, payload string, payloadEncoding rmq.PayloadEncoding) {
	opts := rabbithole.PublishOptions{
		RoutingKey:      queue,
		Properties:      props,
		Payload:         payload,
		PayloadEncoding: string(payloadEncoding),
	}

	q.App().StatusLine().Info(fmt.Sprintf("Publishing message to %s", queue))

	_, err := q.Cluster().PublishToExchange(vhost, "amq.default", opts)
	if err != nil {
		q.App().StatusLine().Error(fmt.Sprintf("Failed to publish message: %s", err.Error()))
		return
	}

	q.App().StatusLine().Info(fmt.Sprintf("Message published successfully to %s", queue))

	q.App().DismissModal()
	q.RequestUpdate(view.PartialUpdate)
}

func (q *Queues) moveMessagesCmd(*tcell.EventKey) *tcell.EventKey {
	queue, ok := q.GetSelectedResource()
	if !ok {
		return nil
	}

	node, err := q.Cluster().GetNode(queue.Node)
	if err != nil {
		q.App().StatusLine().Error(fmt.Sprintf("Failed to get node info: %s", err))
		return nil
	}

	if !slices.ContainsFunc(node.ErlangApps, func(app rabbithole.ErlangApp) bool { return app.Name == "rabbitmq_shovel" }) {
		q.App().StatusLine().Error("Shovel plugin is not enabled on the selected node")
		return nil
	}

	destinationQueues, err := q.Cluster().ListQueuesIn(queue.Vhost)
	if err != nil {
		q.App().StatusLine().Error(fmt.Sprintf("Failed to list queues: %s", err))
		return nil
	}

	destinationQueueNames := utils.FilterMap(
		destinationQueues,
		func(q rabbithole.QueueInfo) bool { return q.Name != queue.Name },
		func(q rabbithole.QueueInfo) string { return q.Name })

	ShowMoveMessagesDialog(q.App(), queue.Vhost, queue.Name, destinationQueueNames, q.moveMessages)

	return nil
}

func (q *Queues) moveMessages(vhost, sourceQueue, destinationQueue string) {
	uri := rabbithole.URISet{fmt.Sprintf("amqp:///%s", url.PathEscape(vhost))}
	sd := rabbithole.ShovelDefinition{
		SourceURI:        uri,
		DestinationURI:   uri,
		AckMode:          "on-confirm",
		DeleteAfter:      rabbithole.DeleteAfter("queue-length"),
		SourceProtocol:   "amqp091",
		SourceQueue:      sourceQueue,
		DestinationQueue: destinationQueue,
	}

	_, err := q.Cluster().DeclareShovel(vhost, fmt.Sprintf("move-from-%s", sourceQueue), sd)
	if err != nil {
		q.App().StatusLine().Error(fmt.Sprintf("Failed to create shovel: %s", err.Error()))
		return
	}

	q.App().DismissModal()
	q.RequestUpdate(view.PartialUpdate)
}
