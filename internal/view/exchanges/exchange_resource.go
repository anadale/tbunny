package exchanges

import (
	"fmt"
	"strings"
	"tbunny/internal/view/bindings"

	rabbithole "github.com/michaelklishin/rabbit-hole/v3"
)

type ExchangeResource struct {
	rabbithole.ExchangeInfo
}

func (r *ExchangeResource) GetBindingDetails() bindings.BindingDetails {
	return bindings.BindingDetails{
		Subject: r.Name,
		Vhost:   r.Vhost,
	}
}

func (r *ExchangeResource) GetName() string {
	return r.Name
}

func (r *ExchangeResource) GetDisplayName() string {
	return "exchange " + r.Name
}

func (r *ExchangeResource) GetTableRowID() string {
	return fmt.Sprintf("%s-%s", r.Vhost, r.Name)
}

func (r *ExchangeResource) GetTableColumnValue(columnName string) string {
	switch columnName {
	case "vhost":
		return r.Vhost
	case "name":
		if r.Name == "" {
			return "(AMQP default)"
		}
		return r.Name
	case "type":
		return r.Type
	case "features":
		return r.getFeatures()
	case "msgRateIn":
		if r.MessageStats == nil || r.MessageStats.PublishInDetails == nil {
			return ""
		}
		return fmt.Sprintf("%.2f", r.MessageStats.PublishInDetails.Rate)
	case "msgRateOut":
		if r.MessageStats == nil || r.MessageStats.PublishOutDetails == nil {
			return ""
		}
		return fmt.Sprintf("%.2f", r.MessageStats.PublishOutDetails.Rate)
	default:
		return ""
	}
}

func (r *ExchangeResource) getFeatures() string {
	var parts []string

	if r.Durable {
		parts = append(parts, "D")
	}

	if r.AutoDelete {
		parts = append(parts, "AD")
	}

	if r.Internal {
		parts = append(parts, "I")
	}

	if len(r.Arguments) > 0 {
		parts = append(parts, "Args")
	}

	return strings.Join(parts, " ")
}
