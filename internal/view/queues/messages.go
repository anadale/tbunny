package queues

import (
	"fmt"
	"tbunny/internal/model"
	"tbunny/internal/rmq"
	"tbunny/internal/ui"
	"tbunny/internal/utils"
	"tbunny/internal/view"

	"github.com/gdamore/tcell/v2"
	"github.com/rivo/tview"
)

type Messages struct {
	view.ResourceView[*MessageResource]

	messages []*MessageResource
}

func NewMessages(messages []*rmq.FetchedMessage, queue, vhost string) model.View {
	v := Messages{
		ResourceView: view.NewResourceTableView[*MessageResource]("Messages", view.NewManualUpdateStrategy()),
		messages: utils.MapWithIndex(messages, func(idx int, fm *rmq.FetchedMessage) *MessageResource {
			return &MessageResource{fm, idx}
		}),
	}

	v.SetPath(view.VhostDisplayName(vhost) + " ⏵ " + queue)
	v.SetResourceProvider(&v)
	v.AddBindingKeysFn(v.bindKeys)

	return &v
}

func (v *Messages) GetResources() ([]*MessageResource, error) {
	return v.messages, nil
}

func (v *Messages) GetColumns() []ui.TableColumn {
	c := []ui.TableColumn{
		{Name: "index", Title: "IDX"},
		{Name: "exchange", Title: "EXCHANGE", Expansion: 1},
		{Name: "routingKey", Title: "ROUTING KEY", Expansion: 1},
		{Name: "deliveryMode", Title: "MODE"},
		{Name: "contentType", Title: "CONTENT TYPE"},
		{Name: "length", Title: "LENGTH", Align: tview.AlignRight},
	}

	return c
}

func (v *Messages) CanDeleteResources() bool {
	return false
}

func (v *Messages) DeleteResource(*MessageResource) error {
	return nil
}

func (v *Messages) bindKeys(km ui.KeyMap) {
	km.Add(tcell.KeyEnter, ui.NewKeyActionWithGroup("View message", v.showMessageCmd, false, 0))
}

func (v *Messages) showMessageCmd(*tcell.EventKey) *tcell.EventKey {
	row, ok := v.GetSelectedResource()
	if !ok {
		return nil
	}

	messageView := NewMessageView(row)

	err := v.App().AddView(messageView)
	if err != nil {
		v.App().StatusLine().Error(fmt.Sprintf("Failed to load message details: %s", err))
	}

	return nil
}
