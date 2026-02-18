package clusters

import (
	"tbunny/internal/cluster"
)

type ClusterResource struct {
	*cluster.Config

	name   string
	active bool
}

func NewClusterResource(name string, cfg *cluster.Config, active bool) *ClusterResource {
	return &ClusterResource{
		Config: cfg,
		name:   name,
		active: active,
	}
}

func (c *ClusterResource) GetName() string {
	return c.name
}

func (c *ClusterResource) GetDisplayName() string {
	return "cluster " + c.name
}

func (c *ClusterResource) GetTableRowID() string {
	return c.name
}

func (c *ClusterResource) GetTableColumnValue(columnName string) string {
	switch columnName {
	case "name":
		if c.active {
			return c.name + "(*)"
		}
		return c.name

	case "uri":
		return c.Connection.Uri

	case "username":
		return c.Connection.Username

	default:
		return ""
	}
}
