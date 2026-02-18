package queues

import (
	"fmt"
	"strings"
	"tbunny/internal/model"
	"tbunny/internal/rmq"
	"tbunny/internal/skins"
	"tbunny/internal/ui"
	"tbunny/internal/utils"
	"tbunny/internal/view"
	"time"

	"github.com/gdamore/tcell/v2"
	"github.com/go-faster/jx"
	"github.com/rivo/tview"
	"golang.design/x/clipboard"
)

const MessageViewTitleFmt = " [fg:bg:b]%s [count:bg:b]%d[fg:bg:-][fg:bg:-] "

type MessageView struct {
	*view.View[*tview.TextView]

	message        *MessageResource
	skin           *skins.Skin
	showProperties bool
	wrap           bool
}

func NewMessageView(message *MessageResource) *MessageView {
	tv := tview.NewTextView()

	v := MessageView{
		View:    view.NewView[*tview.TextView]("Message", tv),
		message: message,
	}

	tv.SetScrollable(true).SetDynamicColors(true)
	tv.SetWrap(false).SetWordWrap(false).SetRegions(true)
	tv.SetBorderPadding(0, 0, 1, 1)
	tv.SetBorder(true)
	tv.SetTitle("Message")

	v.AddBindingKeysFn(v.bindKeys)

	return &v
}

func (v *MessageView) Init(app model.App) (err error) {
	err = v.View.Init(app)
	if err != nil {
		return err
	}

	skm := v.App().SkinManager()

	v.skin = skm.Skin
	skm.AddListener(v)

	return nil
}

func (v *MessageView) Start() {
	v.View.Start()

	v.updateText()
	v.updateTitle()
}

func (v *MessageView) SkinChanged(skin *skins.Skin) {
	v.skin = skin
	v.updateTitle()
}

func (v *MessageView) updateTitle() {
	var title string

	title = view.SkinTitle(fmt.Sprintf(MessageViewTitleFmt, v.Name(), v.message.index), &v.skin.Frame)

	v.Ui().SetTitle(title)
}

func (v *MessageView) updateText() {
	cs := fmt.Sprintf("%s:%s:-", v.skin.Views.Stats.CaptionFgColor.String(), v.skin.Views.Stats.CaptionBgColor.String())
	div := fmt.Sprintf("[%s]%s[-]\n", cs, strings.Repeat("─", 60))

	var b strings.Builder

	if v.showProperties {
		b.WriteString(fmt.Sprintf("[%s]Properties[-]\n", cs))
		b.WriteString(div)
		b.WriteString(v.formatProperties())

		if len(v.message.Properties.Headers) > 0 {
			b.WriteString("\n")
			b.WriteString(fmt.Sprintf("[%s]Headers[-]\n", cs))
			b.WriteString(div)
			b.WriteString(v.formatKeyValuePairs(utils.NewKeyValuePairsFromMap(v.message.Properties.Headers).Sort()))
		}

		b.WriteString("\n")
		b.WriteString(fmt.Sprintf("[%s]Payload[-]\n", cs))
		b.WriteString(div)
	}

	b.WriteString(v.formatPayload())

	v.Ui().SetText(b.String())
}

func (v *MessageView) formatProperties() string {
	props := v.message.Properties

	var pairs utils.KeyValuePairs

	pairs = pairs.Add("Exchange", view.ExchangeDisplayName(v.message.Exchange))
	pairs = pairs.Add("Routing key", v.message.RoutingKey)
	pairs = pairs.Add("Redelivered", v.message.Redelivered)

	if props.AppId != "" {
		pairs = pairs.Add("App ID", props.AppId)
	}

	if props.ContentEncoding != "" {
		pairs = pairs.Add("Content encoding", props.ContentEncoding)
	}

	if props.ContentType != "" {
		pairs = pairs.Add("Content type", props.ContentType)
	}

	if props.CorrelationId != "" {
		pairs = pairs.Add("Correlation ID", props.CorrelationId)
	}

	if props.DeliveryMode != rmq.MessageDeliveryModeUnknown {
		pairs = pairs.Add("Delivery mode", props.DeliveryMode.String())
	}

	if props.Expiration != "" {
		pairs = pairs.Add("Expiration", props.Expiration)
	}

	if props.MessageId != "" {
		pairs = pairs.Add("Message ID", props.MessageId)
	}

	if props.Priority != 0 {
		pairs = pairs.Add("Priority", props.Priority)
	}

	if props.ReplyTo != "" {
		pairs = pairs.Add("Reply to", props.ReplyTo)
	}

	if props.Timestamp != 0 {
		t := time.UnixMilli(props.Timestamp)
		pairs = pairs.Add("Timestamp", t.Format(time.DateTime))
	}

	if props.Type != "" {
		pairs = pairs.Add("Type", props.Type)
	}

	if props.UserId != "" {
		pairs = pairs.Add("User ID", props.UserId)
	}

	return v.formatKeyValuePairs(pairs)
}

func (v *MessageView) formatKeyValuePairs(pairs utils.KeyValuePairs) string {
	ls := fmt.Sprintf("%s:%s:-", v.skin.Views.Stats.LabelFgColor.String(), v.skin.Views.Stats.BgColor.String())

	var b strings.Builder
	maxKeyLen := 0

	for _, kvp := range pairs {
		if len(kvp.Key) > maxKeyLen {
			maxKeyLen = len(kvp.Key)
		}
	}

	for _, kvp := range pairs {
		b.WriteString(fmt.Sprintf("[%s]%s:[-]%s", ls, kvp.Key, strings.Repeat(" ", maxKeyLen-len(kvp.Key)+1)))
		v.formatKvpValue(kvp.Value, &b)
		b.WriteString("\n")
	}

	return b.String()
}

func (v *MessageView) formatKvpValue(value any, b *strings.Builder) {
	switch typed := value.(type) {
	case string:
		if strings.Contains(typed, "\n") {
			b.WriteString("\n")
		}
		b.WriteString(fmt.Sprintf("[%s]%s[-]", v.skin.Views.Stats.ValueFgColor.String(), tview.Escape(typed)))
	case int, int8, int16, int32, int64:
		b.WriteString(fmt.Sprintf("[%s]%d[-]", v.skin.Views.Stats.ValueFgColor.String(), typed))
	case float32, float64:
		b.WriteString(fmt.Sprintf("[%s]%g[-]", v.skin.Views.Stats.ValueFgColor.String(), typed))
	case bool:
		b.WriteString(fmt.Sprintf("[%s]%t[-]", v.skin.Views.Stats.ValueFgColor.String(), typed))
	case []any:
		v.formatKvpList(typed, b)
	}
}

func (v *MessageView) formatKvpList(list []any, b *strings.Builder) {
	b.WriteString(fmt.Sprintf("[%s][", v.skin.Views.Json.BracketColor.String()))

	length := len(list)

	for i, value := range list {
		v.formatKvpValue(value, b)

		if i < length-1 {
			b.WriteString(fmt.Sprintf("[%s], ", v.skin.Views.Json.PunctuationColor.String()))
		}
	}

	b.WriteString(fmt.Sprintf("[%s]]", v.skin.Views.Json.BracketColor.String()))
}

func (v *MessageView) formatPayload() string {
	if v.message.PayloadEncoding == rmq.PayloadEncodingBase64 {
		return v.message.Payload
	}

	if payload, err := v.formatJsonPayload(); err == nil {
		return payload
	}

	return v.message.Payload
}

func (v *MessageView) formatJsonPayload() (string, error) {
	d := jx.DecodeStr(v.message.Payload)
	f := view.NewJxFormatter(v.skin)

	return f.Format(d)
}

func (v *MessageView) bindKeys(km ui.KeyMap) {
	km.Add(ui.KeyP, ui.NewKeyAction("Toggle headers", v.togglePropertiesCmd))

	if utils.IsClipboardSupported() {
		km.Add(ui.KeyC, ui.NewKeyAction("Copy payload to clipboard", v.copyPayloadCmd))
	}

	km.Add(ui.KeyW, ui.NewKeyAction("Toggle wrap", v.toggleWrapCmd))
}

func (v *MessageView) togglePropertiesCmd(*tcell.EventKey) *tcell.EventKey {
	v.showProperties = !v.showProperties
	v.updateText()

	return nil
}

func (v *MessageView) copyPayloadCmd(*tcell.EventKey) *tcell.EventKey {
	clipboard.Write(clipboard.FmtText, []byte(v.message.Payload))

	return nil
}

func (v *MessageView) toggleWrapCmd(*tcell.EventKey) *tcell.EventKey {
	v.wrap = !v.wrap
	v.Ui().SetWrap(v.wrap)

	return nil
}
