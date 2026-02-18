package users

import (
	"strings"

	rabbithole "github.com/michaelklishin/rabbit-hole/v3"
)

type Resource struct {
	rabbithole.UserInfo
}

func (r *Resource) GetName() string {
	return r.Name
}

func (r *Resource) GetDisplayName() string {
	return "users " + r.Name
}

func (r *Resource) GetTableRowID() string {
	return r.Name
}

func (r *Resource) GetTableColumnValue(columnName string) string {
	switch columnName {
	case "name":
		return r.Name
	case "tags":
		return strings.Join(r.Tags, ", ")
	}

	return ""
}
