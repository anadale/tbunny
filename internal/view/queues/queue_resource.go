package queues

import (
	"fmt"
	"strings"
	"tbunny/internal/view/bindings"

	rabbithole "github.com/michaelklishin/rabbit-hole/v3"
)

type QueueResource struct {
	rabbithole.QueueInfo
}

func (r *QueueResource) GetBindingDetails() bindings.BindingDetails {
	return bindings.BindingDetails{
		Subject: r.Name,
		Vhost:   r.Vhost,
	}
}

func (r *QueueResource) GetName() string {
	return r.Name
}

func (r *QueueResource) GetDisplayName() string {
	return "queue " + r.Name
}

func (r *QueueResource) GetTableRowID() string {
	return fmt.Sprintf("%s-%s", r.Vhost, r.Name)
}

func (r *QueueResource) GetTableColumnValue(columnName string) string {
	switch columnName {
	case "vhost":
		return r.Vhost
	case "name":
		return r.Name
	case "node":
		return r.Node
	case "type":
		return r.Type
	case "features":
		return r.getFeatures()
	case "msgReady":
		return fmt.Sprintf("%d", r.MessagesReady)
	case "msgUnacked":
		return fmt.Sprintf("%d", r.MessagesUnacknowledged)
	case "msgTotal":
		return fmt.Sprintf("%d", r.Messages)
	case "msgRateIn":
		if r.MessageStats == nil {
			return ""
		}
		return fmt.Sprintf("%.2f", r.MessageStats.PublishDetails.Rate)
	case "msgRateDelivered":
		if r.MessageStats == nil {
			return ""
		}
		return fmt.Sprintf("%.2f", r.MessageStats.DeliverDetails.Rate)
	case "msgRateAcked":
		if r.MessageStats == nil {
			return ""
		}
		return fmt.Sprintf("%.2f", r.MessageStats.AckDetails.Rate)
	default:
		return ""
	}
}

func (r *QueueResource) getFeatures() string {
	var parts []string

	if r.Durable {
		parts = append(parts, "D")
	}

	if r.AutoDelete {
		parts = append(parts, "AD")
	}

	if r.Exclusive {
		parts = append(parts, "E")
	}

	if len(r.Arguments) > 0 {
		parts = append(parts, "Args")
	}

	return strings.Join(parts, " ")
}
