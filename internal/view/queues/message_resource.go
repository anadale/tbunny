package queues

import (
	"strconv"
	"tbunny/internal/rmq"
	"tbunny/internal/view"
)

type MessageResource struct {
	*rmq.FetchedMessage

	index int
}

func (r *MessageResource) GetName() string {
	return strconv.Itoa(r.index)
}

func (r *MessageResource) GetDisplayName() string {
	return "Message " + strconv.Itoa(r.index)
}

func (r *MessageResource) GetTableRowID() string {
	return strconv.Itoa(r.index)
}

func (r *MessageResource) GetTableColumnValue(columnName string) string {
	switch columnName {
	case "index":
		return strconv.Itoa(r.index)
	case "contentType":
		return r.Properties.ContentType
	case "deliveryMode":
		return r.Properties.DeliveryMode.String()
	case "exchange":
		return view.ExchangeDisplayName(r.Exchange)
	case "length":
		return view.FormatBytes(r.PayloadBytes)
	case "routingKey":
		return r.RoutingKey
	}

	return ""
}
