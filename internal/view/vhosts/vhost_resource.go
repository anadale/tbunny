package vhosts

import (
	"fmt"
	"tbunny/internal/view"

	rabbithole "github.com/michaelklishin/rabbit-hole/v3"
)

type VHostResource struct {
	rabbithole.VhostInfo

	active bool
}

func (r *VHostResource) GetName() string {
	return r.Name
}

func (r *VHostResource) GetDisplayName() string {
	return "virtual host " + r.Name
}

func (r *VHostResource) GetTableRowID() string {
	return r.Name
}

func (r *VHostResource) GetTableColumnValue(columnName string) string {
	if columnName == "name" {
		if r.active {
			return view.VhostDisplayName(r.Name) + " (*)"
		}
		return view.VhostDisplayName(r.Name)
	}

	if r.Name == "" {
		return ""
	}

	switch columnName {
	case "msgReady":
		return fmt.Sprintf("%d", r.MessagesReady)
	case "msgRateReady":
		return fmt.Sprintf("%.2f", r.MessagesReadyDetails.Rate)
	case "msgUnacked":
		return fmt.Sprintf("%d", r.MessagesUnacknowledged)
	case "msgRateUnacked":
		return fmt.Sprintf("%.2f", r.MessagesUnacknowledgedDetails.Rate)
	case "msgTotal":
		return fmt.Sprintf("%d", r.Messages)
	case "msgRateTotal":
		return fmt.Sprintf("%.2f", r.MessagesDetails.Rate)
	default:
		return ""
	}
}
