package connections

import (
	"fmt"
	"strconv"
	"tbunny/internal/view"

	rabbithole "github.com/michaelklishin/rabbit-hole/v3"
)

type ConnectionResource struct {
	rabbithole.ConnectionInfo
}

func (r *ConnectionResource) GetName() string {
	return fmt.Sprintf("%s:%d - %s", r.PeerHost, r.PeerPort, r.ConnectionName())
}

func (r *ConnectionResource) GetDisplayName() string {
	if r.Name == "" {
		return fmt.Sprintf("Connection from host %s:%d", r.PeerHost, r.PeerPort)
	}

	return fmt.Sprintf("Connection %s from host %s:%d", r.ConnectionName(), r.PeerHost, r.PeerPort)
}

func (r *ConnectionResource) GetTableRowID() string {
	return fmt.Sprintf("%s:%d-%s", r.PeerHost, r.PeerPort, r.ConnectionName())
}

func (r *ConnectionResource) ConnectionName() string {
	if name, ok := r.ClientProperties["connection_name"]; ok {
		return name.(string)
	}

	return ""
}

func (r *ConnectionResource) GetTableColumnValue(columnName string) string {
	switch columnName {
	case "client":
		return fmt.Sprintf("%s:%d", r.PeerHost, r.PeerPort)
	case "name":
		return r.ConnectionName()
	case "vhost":
		return r.Vhost
	case "username":
		return r.User
	case "state":
		return r.State
	case "channels":
		return strconv.Itoa(r.Channels)
	case "protocol":
		return r.Protocol
	case "fromClient":
		return view.FormatRate(r.RecvOctDetails.Rate)
	case "toClient":
		return view.FormatRate(r.SendOctDetails.Rate)
	case "tls":
		return view.FormatBool(r.UsesTLS)
	case "node":
		return r.Node
	default:
		return ""
	}
}
