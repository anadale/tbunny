package bindings

import (
	"fmt"

	rabbithole "github.com/michaelklishin/rabbit-hole/v3"
)

const defaultExchangeBindingTitle = "(Default exchange binding)"

type BindingResource struct {
	rabbithole.BindingInfo
}

func (r *BindingResource) GetName() string {
	return r.Source + "/" + r.Destination
}

func (r *BindingResource) GetDisplayName() string {
	if r.RoutingKey != "" {
		return fmt.Sprintf("binding from exchange \"%s\" to %s \"%s\" with routing key \"%s\"", r.Source, r.DestinationType, r.Destination, r.RoutingKey)
	}

	return fmt.Sprintf("binding from exchange \"%s\" to %s \"%s\"", r.Source, r.DestinationType, r.Destination)
}

func (r *BindingResource) GetTableRowID() string {
	return r.Vhost + "-" + r.Source + "-" + r.Destination + "-" + r.RoutingKey
}

func (r *BindingResource) GetTableColumnValue(columnName string) string {
	switch columnName {
	case "source":
		if r.Source == "" {
			return defaultExchangeBindingTitle
		}
		return r.Source
	case "destination":
		if r.Destination == "" {
			return defaultExchangeBindingTitle
		}
		return r.Destination
	case "destinationType":
		return r.DestinationType
	case "routingKey":
		return r.RoutingKey
	case "features":
		if len(r.Arguments) > 0 {
			return "Args"
		}
		return ""
	default:
		return ""
	}
}
