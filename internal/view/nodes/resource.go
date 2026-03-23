package nodes

import (
	"fmt"

	"tbunny/internal/view"

	rabbithole "github.com/michaelklishin/rabbit-hole/v3"
)

// Resource wraps rabbithole.NodeInfo to implement the view.Resource interface.
type Resource struct {
	rabbithole.NodeInfo
}

func (r *Resource) GetName() string {
	return r.Name
}

func (r *Resource) GetDisplayName() string {
	return "node " + r.Name
}

func (r *Resource) GetTableRowID() string {
	return r.Name
}

func (r *Resource) GetTableColumnValue(columnName string) string {
	switch columnName {
	case "name":
		return r.Name
	case "type":
		return r.NodeType
	case "running":
		return view.FormatBool(r.IsRunning)
	case "os_pid":
		return string(r.OsPid)
	case "fd_used":
		return fmt.Sprintf("%d / %d", r.FdUsed, r.FdTotal)
	case "mem_used":
		return view.FormatBytes(r.MemUsed)
	case "mem_limit":
		return view.FormatBytes(r.MemLimit)
	case "mem_alarm":
		return view.FormatBool(r.MemAlarm)
	case "disk_free":
		return view.FormatBytes(r.DiskFree)
	case "disk_alarm":
		return view.FormatBool(r.DiskFreeAlarm)
	case "proc_used":
		return fmt.Sprintf("%d / %d", r.ProcUsed, r.ProcTotal)
	case "uptime":
		return formatUptime(r.Uptime)
	}

	return ""
}

// formatUptime formats a duration given in milliseconds as a human-readable string.
// e.g. "2d 5h 30m 12s" or "45m 3s".
func formatUptime(ms uint64) string {
	totalSeconds := ms / 1000

	days := totalSeconds / 86400
	hours := (totalSeconds % 86400) / 3600
	minutes := (totalSeconds % 3600) / 60
	seconds := totalSeconds % 60

	if days > 0 {
		return fmt.Sprintf("%dd %dh %dm %ds", days, hours, minutes, seconds)
	}

	if hours > 0 {
		return fmt.Sprintf("%dh %dm %ds", hours, minutes, seconds)
	}

	if minutes > 0 {
		return fmt.Sprintf("%dm %ds", minutes, seconds)
	}

	return fmt.Sprintf("%ds", seconds)
}
