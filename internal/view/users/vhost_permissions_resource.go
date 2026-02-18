package users

import rabbithole "github.com/michaelklishin/rabbit-hole/v3"

type VhostPermissionsResource struct {
	rabbithole.PermissionInfo
}

func (r *VhostPermissionsResource) GetName() string {
	return r.Vhost
}

func (r *VhostPermissionsResource) GetDisplayName() string {
	return "permissions for vhost " + r.Vhost
}

func (r *VhostPermissionsResource) GetTableRowID() string {
	return r.Vhost
}

func (r *VhostPermissionsResource) GetTableColumnValue(columnName string) string {
	switch columnName {
	case "vhost":
		return r.Vhost
	case "configure":
		return r.Configure
	case "write":
		return r.Write
	case "read":
		return r.Read
	default:
		return ""
	}
}
