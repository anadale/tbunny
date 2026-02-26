package connections

import (
	"fmt"
	"slices"
	"strings"
	"tbunny/internal/model"
	"tbunny/internal/skins"
	"tbunny/internal/utils"
	"tbunny/internal/view"
	"time"

	rabbithole "github.com/michaelklishin/rabbit-hole/v3"
	"github.com/rivo/tview"
)

const ConnectionDetailsTitleFmt = " [fg:bg:b]%s[fg:bg:-]([hilite:bg:b]%s[fg:bg:-]) "

type ConnectionDetails struct {
	*view.ClusterAwareRefreshableView[*tview.TextView]

	name       string
	connection *rabbithole.ConnectionInfo
	skin       *skins.Skin
}

func NewConnectionDetails(name string) *ConnectionDetails {
	textView := tview.NewTextView()
	textView.SetBorderPadding(1, 0, 1, 1)
	textView.SetBorder(true)
	textView.SetDynamicColors(true)
	textView.SetScrollable(true)
	textView.SetWordWrap(false)

	strategy := view.NewLiveUpdateStrategy()

	v := &ConnectionDetails{
		ClusterAwareRefreshableView: view.NewClusterAwareRefreshableView("Connection Details", textView, strategy),
		name:                        name,
	}

	v.SetUpdateFn(v.performUpdate)

	return v
}

func (v *ConnectionDetails) Init(app model.App) error {
	err := v.ClusterAwareRefreshableView.Init(app)
	if err != nil {
		return err
	}

	v.skin = skins.Current()
	v.updateTitle()

	return nil
}

func (v *ConnectionDetails) performUpdate(view.UpdateKind) {
	connection, err := v.Cluster().GetConnection(v.name)
	if err != nil {
		v.App().StatusLine().Error(fmt.Sprintf("Failed to fetch connection details: %v", err))
		return
	}

	v.connection = connection

	v.App().QueueUpdateDraw(func() {
		v.updateContent()
	})
}

func (v *ConnectionDetails) updateTitle() {
	title := view.SkinTitle(fmt.Sprintf(ConnectionDetailsTitleFmt, v.Name(), v.name))

	v.Ui().SetTitle(title)
}

func (v *ConnectionDetails) updateContent() {
	var b strings.Builder

	v.formatDetails(&b)
	v.formatClientProperties(&b)

	styledContent := view.SkinStatsContent(b.String(), &v.skin.Views.Stats)

	v.Ui().SetText(styledContent)
}

type labelAndFormatter struct {
	label string
	fn    detailsFormatter
}

type detailsFormatter func(connection *rabbithole.ConnectionInfo) string

var detailsFormatters = []labelAndFormatter{
	{label: "Node:", fn: func(connection *rabbithole.ConnectionInfo) string { return connection.Node }},
	{label: "Client-provided name:", fn: func(connection *rabbithole.ConnectionInfo) string {
		if name, ok := connection.ClientProperties["connection_name"]; ok {
			return name.(string)
		}

		return ""
	}},
	{label: "User name:", fn: func(connection *rabbithole.ConnectionInfo) string { return connection.User }},
	{label: "Protocol:", fn: func(connection *rabbithole.ConnectionInfo) string { return connection.Protocol }},
	{label: "Connected at:", fn: func(connection *rabbithole.ConnectionInfo) string {
		t := time.UnixMilli(int64(connection.ConnectedAt))
		return t.Format(time.DateTime)
	}},
	{label: "SASL auth mechanism:", fn: func(connection *rabbithole.ConnectionInfo) string {
		if connection.SSLProtocol != "" {
			return connection.SSLProtocol
		}

		return "PLAIN"
	}},
	{label: "State:", fn: func(connection *rabbithole.ConnectionInfo) string { return connection.State }},
	{label: "Heartbeat:", fn: func(connection *rabbithole.ConnectionInfo) string { return fmt.Sprintf("%ds", connection.Timeout) }},
	{label: "Frame max:", fn: func(connection *rabbithole.ConnectionInfo) string { return view.FormatBytes(connection.FrameMax) }},
}

func (v *ConnectionDetails) formatDetails(b *strings.Builder) {
	maxLabelLength := 0

	for _, lf := range detailsFormatters {
		if len(lf.label) > maxLabelLength {
			maxLabelLength = len(lf.label)
		}
	}

	b.WriteString("[caption]Details[-]\n\n")
	for _, lf := range detailsFormatters {
		b.WriteString(fmt.Sprintf("[label]%s[-] [value]%s[-]\n", utils.PadRight(lf.label, maxLabelLength), lf.fn(v.connection)))
	}
}

func (v *ConnectionDetails) formatClientProperties(b *strings.Builder) {
	b.WriteString("\n[caption]Client properties[-]\n\n")
	v.formatMap(b, v.connection.ClientProperties, 0)
}

func (v *ConnectionDetails) formatMap(b *strings.Builder, m map[string]any, padding int) {
	if len(m) == 0 {
		b.WriteString("\n")
		return
	}

	maxKeyLength := 0
	keys := make([]string, 0, len(m))

	for k := range m {
		keys = append(keys, k)
		if len(k) > maxKeyLength {
			maxKeyLength = len(k)
		}
	}

	slices.Sort(keys)

	paddingStr := strings.Repeat(" ", padding)

	for i, key := range keys {
		if i > 0 {
			b.WriteString(paddingStr)
		}

		b.WriteString(fmt.Sprintf("[label]%s:%s[-] ", key, strings.Repeat(" ", maxKeyLength-len(key))))

		switch value := m[key].(type) {
		case map[string]any:
			v.formatMap(b, value, padding+maxKeyLength+2)
		default:
			b.WriteString(fmt.Sprintf("[value]%v[-]\n", value))
		}
	}
}
