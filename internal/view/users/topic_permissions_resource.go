package users

import (
	"fmt"
	"tbunny/internal/view"

	rabbithole "github.com/michaelklishin/rabbit-hole/v3"
)

type TopicPermissionsResource struct {
	rabbithole.TopicPermissionInfo
}

func (r *TopicPermissionsResource) GetName() string {
	return r.Vhost + ":" + r.Exchange
}

func (r *TopicPermissionsResource) GetDisplayName() string {
	return fmt.Sprintf("permissions for topic %s in vhost %s", r.Exchange, r.Vhost)
}

func (r *TopicPermissionsResource) GetTableRowID() string {
	return r.Vhost + ":" + r.Exchange
}

func (r *TopicPermissionsResource) GetTableColumnValue(columnName string) string {
	switch columnName {
	case "vhost":
		return r.Vhost
	case "exchange":
		return view.ExchangeDisplayName(r.Exchange)
	case "write":
		return r.Write
	case "read":
		return r.Read
	default:
		return ""
	}
}
