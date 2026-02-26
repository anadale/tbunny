package queues

import (
	"fmt"
	"strings"
	"tbunny/internal/model"
	"tbunny/internal/skins"
	"tbunny/internal/view"

	"github.com/gdamore/tcell/v2"
	rabbithole "github.com/michaelklishin/rabbit-hole/v3"
	"github.com/rivo/tview"
)

const QueueDetailsTitleFmt = " [fg:bg:b]%s[fg:bg:-]([hilite:bg:b]%s[fg:bg:-]) "

const (
	// Minimum width for four-column mode (4 columns x 28 + padding).
	minWidthForFourColumns = 117
)

type QueueDetails struct {
	*view.ClusterAwareRefreshableView[*tview.Flex]

	name      string
	vhost     string
	queueInfo *rabbithole.DetailedQueueInfo

	skin *skins.Skin

	useScrollableMode bool

	// Panels for four-column mode.
	col1Panel      *tview.Flex
	col2Panel      *tview.Flex
	col3Panel      *tview.Flex
	col4Panel      *tview.Flex
	argumentsPanel *tview.Flex
	policyPanel    *tview.Flex
	nodePanel      *tview.Flex

	// Column 1 - Messages.
	msgReady      *labelAndValue
	msgUnacked    *labelAndValue
	msgTotal      *labelAndValue
	msgInMemory   *labelAndValue
	msgPersistent *labelAndValue
	msgTransient  *labelAndValue
	msgPagedOut   *labelAndValue
	msgBytes      *labelAndValue

	// Arguments section (colspan).
	argumentsView *tview.TextView

	// Column 2 - Message Rates.
	ratePublish          *labelAndValue
	rateDeliverManualAck *labelAndValue
	rateDeliverAutoAck   *labelAndValue
	rateConsumerAck      *labelAndValue
	rateRedelivered      *labelAndValue
	rateGetManualAck     *labelAndValue
	rateGetAutoAck       *labelAndValue

	// Policy section (colspan).
	policyView *tview.TextView

	// Column 3 - State & Resources.
	stateStatus          *labelAndValue
	stateConsumers       *labelAndValue
	stateCapacity        *labelAndValue
	stateProcessMemory   *labelAndValue
	stateActiveConsumers *labelAndValue

	// Column 4 - Configuration.
	cfgType         *labelAndValue
	cfgDurable      *labelAndValue
	cfgAutoDelete   *labelAndValue
	cfgExclusive    *labelAndValue
	cfgQueueVersion *labelAndValue

	// Node.
	nodeView *tview.TextView

	// Section headers.
	messagesHeader  *tview.TextView
	ratesHeader     *tview.TextView
	stateHeader     *tview.TextView
	configHeader    *tview.TextView
	argumentsHeader *tview.TextView
	policyHeader    *tview.TextView
	nodeHeader      *tview.TextView

	// For scrollable mode.
	scrollableView *tview.TextView
}

func NewQueueDetails(name, vhost string) *QueueDetails {
	flex := tview.NewFlex().SetDirection(tview.FlexRow)
	flex.SetBorder(true).SetBorderPadding(1, 0, 1, 1)

	strategy := view.NewLiveUpdateStrategy()

	q := QueueDetails{
		ClusterAwareRefreshableView: view.NewClusterAwareRefreshableView[*tview.Flex]("Queue Details", flex, strategy),
		name:                        name,
		vhost:                       vhost,
	}

	q.SetUpdateFn(q.performUpdate)

	return &q
}

func (q *QueueDetails) Init(app model.App) (err error) {
	err = q.ClusterAwareRefreshableView.Init(app)
	if err != nil {
		return err
	}

	q.skin = skins.Current()

	// Start in scrollable mode as a safe fallback.
	// The actual mode is determined on the first update.
	q.useScrollableMode = true
	q.createLayout()
	q.updateTitle()

	return nil
}

func (q *QueueDetails) determineLayoutMode() {
	// Getting dimensions of the main flex container
	_, _, width, _ := q.Ui().GetInnerRect()

	if width == 0 {
		q.useScrollableMode = false
		return
	}

	q.useScrollableMode = width < minWidthForFourColumns
}

func (q *QueueDetails) performUpdate(view.UpdateKind) {
	i, err := q.Cluster().GetQueue(q.vhost, q.name)
	if err != nil {
		q.App().StatusLine().Error(fmt.Sprintf("Не удалось получить данные очереди: %v", err.Error()))
		return
	}

	q.queueInfo = i

	q.App().QueueUpdateDraw(func() {
		// Check whether the display mode has changed.
		previousMode := q.useScrollableMode
		q.determineLayoutMode()

		// If the mode changed, rebuild the layout.
		if previousMode != q.useScrollableMode {
			q.createLayout()
		}

		// Update data in the selected mode.
		if q.useScrollableMode {
			q.updateScrollableView()
		} else {
			q.updateColumnarView()
		}
	})
}

func (q *QueueDetails) updateScrollableView() {
	qi := q.queueInfo

	content := q.formatQueueInfoAsText(qi)
	styledContent := view.SkinStatsContent(content, &q.skin.Views.Stats)

	q.scrollableView.SetText(styledContent)
}

func (q *QueueDetails) updateColumnarView() {
	qi := q.queueInfo

	// Messages
	q.msgReady.SetCount(qi.MessagesReady)
	q.msgUnacked.SetCount(qi.MessagesUnacknowledged)
	q.msgTotal.SetCount(qi.Messages)
	q.msgInMemory.SetCount(qi.MessagesRAM)
	q.msgPersistent.SetCount(qi.MessagesPersistent)

	transient := qi.Messages - qi.MessagesPersistent
	if transient < 0 {
		transient = 0
	}
	q.msgTransient.SetCount(transient)

	pagedOut := qi.Messages - qi.MessagesRAM
	if pagedOut < 0 {
		pagedOut = 0
	}
	q.msgPagedOut.SetCount(pagedOut)
	q.msgBytes.SetBytes(qi.MessagesBytes)

	// Message Rates
	if qi.MessageStats != nil {
		q.ratePublish.SetRate(qi.MessageStats.PublishDetails.Rate)
		q.rateDeliverManualAck.SetRate(qi.MessageStats.DeliverDetails.Rate)
		q.rateDeliverAutoAck.SetRate(qi.MessageStats.DeliverNoAckDetails.Rate)
		q.rateConsumerAck.SetRate(qi.MessageStats.AckDetails.Rate)
		q.rateRedelivered.SetRate(qi.MessageStats.RedeliverDetails.Rate)
		q.rateGetManualAck.SetRate(qi.MessageStats.GetDetails.Rate)
		q.rateGetAutoAck.SetRate(qi.MessageStats.GetNoAckDetails.Rate)
	} else {
		q.ratePublish.SetRate(0)
		q.rateDeliverManualAck.SetRate(0)
		q.rateDeliverAutoAck.SetRate(0)
		q.rateConsumerAck.SetRate(0)
		q.rateRedelivered.SetRate(0)
		q.rateGetManualAck.SetRate(0)
		q.rateGetAutoAck.SetRate(0)
	}

	// Configuration
	q.cfgType.SetText(qi.Type)
	q.cfgDurable.SetBool(qi.Durable)
	q.cfgAutoDelete.SetBool(bool(qi.AutoDelete))
	q.cfgExclusive.SetBool(qi.Exclusive)

	if queueVersion, ok := qi.Arguments["x-queue-version"].(float64); ok {
		q.cfgQueueVersion.SetText(fmt.Sprintf("%.0f", queueVersion))
	} else {
		q.cfgQueueVersion.SetNotApplicable()
	}

	// Arguments
	q.updateArguments(qi.Arguments)

	// Policy
	if qi.Policy != "" {
		q.policyView.SetText(qi.Policy)
	} else {
		q.policyView.SetText("N/A")
	}

	// State & Resources
	if qi.Status != "" {
		q.stateStatus.SetText(qi.Status)
	} else {
		q.stateStatus.SetNotApplicable()
	}
	q.stateConsumers.SetCount(qi.Consumers)

	if qi.ConsumerUtilisation > 0 {
		q.stateCapacity.SetText(fmt.Sprintf("%.0f%%", qi.ConsumerUtilisation*100))
	} else {
		q.stateCapacity.SetNotApplicable()
	}

	q.stateProcessMemory.SetBytes(qi.Memory)
	q.stateActiveConsumers.SetCount(int(qi.ActiveConsumers))

	// Node
	if qi.Node != "" {
		q.nodeView.SetText(qi.Node)
	} else {
		q.nodeView.SetText("N/A")
	}
}

func (q *QueueDetails) updateTitle() {
	title := view.SkinTitle(fmt.Sprintf(QueueDetailsTitleFmt, q.Name(), fmt.Sprintf("%s:%s", q.vhost, q.name)))

	q.Ui().SetTitle(title)
}

func (q *QueueDetails) updateArguments(args map[string]interface{}) {
	if len(args) == 0 {
		q.argumentsView.SetText("N/A")
		return
	}

	var builder strings.Builder
	for key, value := range args {
		// Skip arguments already shown in Configuration.
		if key == "x-queue-version" || key == "x-queue-type" {
			continue
		}

		if builder.Len() > 0 {
			builder.WriteString("\n")
		}

		builder.WriteString(fmt.Sprintf("%s: %v", key, value))
	}

	if builder.Len() == 0 {
		q.argumentsView.SetText("N/A")
	} else {
		q.argumentsView.SetText(builder.String())
	}
}

func (q *QueueDetails) createLayout() {
	// Clear the current layout.
	q.Ui().Clear()

	if q.useScrollableMode {
		q.createScrollableLayout()
	} else {
		q.createFourColumnLayout()
	}

	q.applyStyles()
}

func (q *QueueDetails) createScrollableLayout() {
	q.scrollableView = tview.NewTextView()
	q.scrollableView.SetDynamicColors(true)
	q.scrollableView.SetScrollable(true)
	q.scrollableView.SetWordWrap(false)

	// Handle scrolling keys.
	q.scrollableView.SetInputCapture(func(event *tcell.EventKey) *tcell.EventKey {
		switch event.Key() {
		case tcell.KeyUp:
			row, col := q.scrollableView.GetScrollOffset()
			q.scrollableView.ScrollTo(row-1, col)
			return nil
		case tcell.KeyDown:
			row, col := q.scrollableView.GetScrollOffset()
			q.scrollableView.ScrollTo(row+1, col)
			return nil
		case tcell.KeyPgUp:
			row, col := q.scrollableView.GetScrollOffset()
			q.scrollableView.ScrollTo(row-10, col)
			return nil
		case tcell.KeyPgDn:
			row, col := q.scrollableView.GetScrollOffset()
			q.scrollableView.ScrollTo(row+10, col)
			return nil
		case tcell.KeyHome:
			q.scrollableView.ScrollToBeginning()
			return nil
		case tcell.KeyEnd:
			q.scrollableView.ScrollToEnd()
			return nil
		default:
		}
		return event
	})

	q.Ui().AddItem(q.scrollableView, 0, 1, false)
}

func (q *QueueDetails) createFourColumnLayout() {
	// Create a container for centering.
	mainFlex := tview.NewFlex().SetDirection(tview.FlexColumn)

	// Four columns.
	q.col1Panel = q.createCol1Panel() // Messages
	q.col2Panel = q.createCol2Panel() // Message Rates
	q.col3Panel = q.createCol3Panel() // State & Resources
	q.col4Panel = q.createCol4Panel() // Configuration

	// Add empty flex items for centering.
	mainFlex.AddItem(tview.NewFlex(), 0, 1, false) // Left padding.
	mainFlex.AddItem(q.col1Panel, 28, 0, false)
	mainFlex.AddItem(q.col2Panel, 28, 0, false)
	mainFlex.AddItem(q.col3Panel, 28, 0, false)
	mainFlex.AddItem(q.col4Panel, 28, 0, false)
	mainFlex.AddItem(tview.NewFlex(), 0, 1, false) // Right padding.

	// Arguments section with colspan (28*4 = 112).
	q.argumentsPanel = q.createArgumentsPanel()
	argumentsFlex := tview.NewFlex().SetDirection(tview.FlexColumn)
	argumentsFlex.AddItem(tview.NewFlex(), 0, 1, false)
	argumentsFlex.AddItem(q.argumentsPanel, 112, 0, false)
	argumentsFlex.AddItem(tview.NewFlex(), 0, 1, false)

	// Policy section with colspan (28*4 = 112).
	q.policyPanel = q.createPolicyPanel()
	policyFlex := tview.NewFlex().SetDirection(tview.FlexColumn)
	policyFlex.AddItem(tview.NewFlex(), 0, 1, false)
	policyFlex.AddItem(q.policyPanel, 112, 0, false)
	policyFlex.AddItem(tview.NewFlex(), 0, 1, false)

	// Node section with colspan (28*4 = 112).
	q.nodePanel = q.createNodePanel()
	nodeFlex := tview.NewFlex().SetDirection(tview.FlexColumn)
	nodeFlex.AddItem(tview.NewFlex(), 0, 1, false)
	nodeFlex.AddItem(q.nodePanel, 112, 0, false)
	nodeFlex.AddItem(tview.NewFlex(), 0, 1, false)

	// Add all sections.
	// mainFlex: max column height = 10 (Messages).
	q.Ui().AddItem(mainFlex, 10, 0, false)
	q.Ui().AddItem(tview.NewBox(), 1, 0, false) // Blank line before Arguments.
	q.Ui().AddItem(argumentsFlex, 0, 1, false)
	q.Ui().AddItem(policyFlex, 4, 0, false)
	q.Ui().AddItem(nodeFlex, 3, 0, false)
}

func (q *QueueDetails) createCol1Panel() *tview.Flex {
	panel := tview.NewFlex().SetDirection(tview.FlexRow)

	// Messages: 1 (header) + 1 (separator) + 8 (grid) = 10.
	messagesSection := q.createMessagesSection()
	panel.AddItem(messagesSection, 10, 0, false)

	return panel
}

func (q *QueueDetails) createCol2Panel() *tview.Flex {
	panel := tview.NewFlex().SetDirection(tview.FlexRow)

	// Message Rates: 1 (header) + 1 (separator) + 7 (grid) = 9.
	ratesSection := q.createMessageRatesSection()
	panel.AddItem(ratesSection, 9, 0, false)

	return panel
}

func (q *QueueDetails) createCol3Panel() *tview.Flex {
	panel := tview.NewFlex().SetDirection(tview.FlexRow)

	// State & Resources: 1 (header) + 1 (separator) + 5 (grid) = 7.
	stateSection := q.createStateSection()
	panel.AddItem(stateSection, 7, 0, false)

	return panel
}

func (q *QueueDetails) createCol4Panel() *tview.Flex {
	panel := tview.NewFlex().SetDirection(tview.FlexRow)

	// Configuration: 1 (header) + 1 (separator) + 5 (grid) = 7.
	configSection := q.createConfigurationSection()
	panel.AddItem(configSection, 7, 0, false)

	return panel
}

func (q *QueueDetails) createMessagesSection() *tview.Flex {
	section := tview.NewFlex().SetDirection(tview.FlexRow)

	q.messagesHeader = tview.NewTextView().SetText("Messages").SetDynamicColors(true)
	section.AddItem(q.messagesHeader, 1, 0, false)

	separator := tview.NewTextView().SetText(strings.Repeat("─", 26))
	section.AddItem(separator, 1, 0, false)

	grid := tview.NewGrid().
		SetColumns(13, -1).
		SetRows(1, 1, 1, 1, 1, 1, 1, 1). // Fixed height for 8 rows.
		SetGap(0, 0)

	q.msgReady = createLabelAndValue("Ready:")
	q.msgUnacked = createLabelAndValue("Unacked:")
	q.msgTotal = createLabelAndValue("Total:")
	q.msgInMemory = createLabelAndValue("In memory:")
	q.msgPersistent = createLabelAndValue("Persist:")
	q.msgTransient = createLabelAndValue("Transient:")
	q.msgPagedOut = createLabelAndValue("Paged Out:")
	q.msgBytes = createLabelAndValue("Msg bytes:")

	q.msgReady.AddToGrid(grid, 0, 0)
	q.msgUnacked.AddToGrid(grid, 1, 0)
	q.msgTotal.AddToGrid(grid, 2, 0)
	q.msgInMemory.AddToGrid(grid, 3, 0)
	q.msgPersistent.AddToGrid(grid, 4, 0)
	q.msgTransient.AddToGrid(grid, 5, 0)
	q.msgPagedOut.AddToGrid(grid, 6, 0)
	q.msgBytes.AddToGrid(grid, 7, 0)

	section.AddItem(grid, 8, 0, false) // Fixed section height = 8 rows.

	return section
}

func (q *QueueDetails) createArgumentsPanel() *tview.Flex {
	panel := tview.NewFlex().SetDirection(tview.FlexRow)

	q.argumentsHeader = tview.NewTextView().SetText("Arguments").SetDynamicColors(true)
	panel.AddItem(q.argumentsHeader, 1, 0, false)

	separator := tview.NewTextView().SetText(strings.Repeat("─", 110))
	panel.AddItem(separator, 1, 0, false)

	q.argumentsView = tview.NewTextView().SetDynamicColors(true).SetWrap(true)
	panel.AddItem(q.argumentsView, 0, 1, false)

	return panel
}

func (q *QueueDetails) createMessageRatesSection() *tview.Flex {
	section := tview.NewFlex().SetDirection(tview.FlexRow)

	q.ratesHeader = tview.NewTextView().SetText("Rates (msg/s)").SetDynamicColors(true)
	section.AddItem(q.ratesHeader, 1, 0, false)

	separator := tview.NewTextView().SetText(strings.Repeat("─", 26))
	section.AddItem(separator, 1, 0, false)

	grid := tview.NewGrid().
		SetColumns(13, -1).
		SetRows(1, 1, 1, 1, 1, 1, 1). // Fixed height for 7 rows.
		SetGap(0, 0)

	q.ratePublish = createLabelAndValue("Publish:")
	q.rateDeliverManualAck = createLabelAndValue("Deliv (man):")
	q.rateDeliverAutoAck = createLabelAndValue("Deliv (aut):")
	q.rateConsumerAck = createLabelAndValue("Cons ack:")
	q.rateRedelivered = createLabelAndValue("Redeliv:")
	q.rateGetManualAck = createLabelAndValue("Get (man):")
	q.rateGetAutoAck = createLabelAndValue("Get (aut):")

	q.ratePublish.AddToGrid(grid, 0, 0)
	q.rateDeliverManualAck.AddToGrid(grid, 1, 0)
	q.rateDeliverAutoAck.AddToGrid(grid, 2, 0)
	q.rateConsumerAck.AddToGrid(grid, 3, 0)
	q.rateRedelivered.AddToGrid(grid, 4, 0)
	q.rateGetManualAck.AddToGrid(grid, 5, 0)
	q.rateGetAutoAck.AddToGrid(grid, 6, 0)

	section.AddItem(grid, 7, 0, false) // Fixed section height = 7 rows.

	return section
}

func (q *QueueDetails) createPolicyPanel() *tview.Flex {
	panel := tview.NewFlex().SetDirection(tview.FlexRow)

	q.policyHeader = tview.NewTextView().SetText("Policy").SetDynamicColors(true)
	panel.AddItem(q.policyHeader, 1, 0, false)

	separator := tview.NewTextView().SetText(strings.Repeat("─", 110))
	panel.AddItem(separator, 1, 0, false)

	q.policyView = tview.NewTextView().SetDynamicColors(true).SetMaxLines(1)
	panel.AddItem(q.policyView, 1, 0, false)

	return panel
}

func (q *QueueDetails) createStateSection() *tview.Flex {
	section := tview.NewFlex().SetDirection(tview.FlexRow)

	q.stateHeader = tview.NewTextView().SetText("State & Resources").SetDynamicColors(true)
	section.AddItem(q.stateHeader, 1, 0, false)

	separator := tview.NewTextView().SetText(strings.Repeat("─", 26))
	section.AddItem(separator, 1, 0, false)

	grid := tview.NewGrid().
		SetColumns(13, -1).
		SetRows(1, 1, 1, 1, 1). // Fixed height for 5 rows.
		SetGap(0, 0)

	q.stateStatus = createLabelAndValue("State:")
	q.stateConsumers = createLabelAndValue("Consumers:")
	q.stateCapacity = createLabelAndValue("Capacity:")
	q.stateProcessMemory = createLabelAndValue("Proc mem:")
	q.stateActiveConsumers = createLabelAndValue("Active:")

	q.stateStatus.AddToGrid(grid, 0, 0)
	q.stateConsumers.AddToGrid(grid, 1, 0)
	q.stateCapacity.AddToGrid(grid, 2, 0)
	q.stateProcessMemory.AddToGrid(grid, 3, 0)
	q.stateActiveConsumers.AddToGrid(grid, 4, 0)

	section.AddItem(grid, 5, 0, false) // Fixed section height = 5 rows.

	return section
}

func (q *QueueDetails) createConfigurationSection() *tview.Flex {
	section := tview.NewFlex().SetDirection(tview.FlexRow)

	q.configHeader = tview.NewTextView().SetText("Configuration").SetDynamicColors(true)
	section.AddItem(q.configHeader, 1, 0, false)

	separator := tview.NewTextView().SetText(strings.Repeat("─", 26))
	section.AddItem(separator, 1, 0, false)

	grid := tview.NewGrid().
		SetColumns(13, -1).
		SetRows(1, 1, 1, 1, 1). // Fixed height for 5 rows.
		SetGap(0, 0)

	q.cfgType = createLabelAndValue("Type:")
	q.cfgDurable = createLabelAndValue("Durable:")
	q.cfgAutoDelete = createLabelAndValue("Auto del:")
	q.cfgExclusive = createLabelAndValue("Exclusive:")
	q.cfgQueueVersion = createLabelAndValue("Version:")

	q.cfgType.AddToGrid(grid, 0, 0)
	q.cfgDurable.AddToGrid(grid, 1, 0)
	q.cfgAutoDelete.AddToGrid(grid, 2, 0)
	q.cfgExclusive.AddToGrid(grid, 3, 0)
	q.cfgQueueVersion.AddToGrid(grid, 4, 0)

	section.AddItem(grid, 5, 0, false) // Fixed section height = 5 rows.

	return section
}

func (q *QueueDetails) createNodePanel() *tview.Flex {
	panel := tview.NewFlex().SetDirection(tview.FlexRow)

	q.nodeHeader = tview.NewTextView().SetText("Node").SetDynamicColors(true)
	panel.AddItem(q.nodeHeader, 1, 0, false)

	separator := tview.NewTextView().SetText(strings.Repeat("─", 110))
	panel.AddItem(separator, 1, 0, false)

	q.nodeView = tview.NewTextView().SetDynamicColors(true).SetMaxLines(1)
	panel.AddItem(q.nodeView, 1, 0, false)

	return panel
}

func (q *QueueDetails) applyStyles() {
	q.updateTitle()

	skin := q.skin.Views.Stats
	bgColor := skin.BgColor.Color()

	q.Ui().SetBackgroundColor(bgColor)

	if q.useScrollableMode {
		q.scrollableView.SetTextColor(skin.ValueFgColor.Color())
		q.scrollableView.SetBackgroundColor(bgColor)
	} else {
		labelStyle := tcell.StyleDefault.Foreground(skin.LabelFgColor.Color()).Background(bgColor)
		valueStyle := tcell.StyleDefault.Foreground(skin.ValueFgColor.Color()).Background(bgColor)
		captionStyle := tcell.StyleDefault.Foreground(skin.CaptionFgColor.Color()).Background(bgColor)

		// Messages
		q.msgReady.ApplyStyles(labelStyle, valueStyle)
		q.msgUnacked.ApplyStyles(labelStyle, valueStyle)
		q.msgTotal.ApplyStyles(labelStyle, valueStyle)
		q.msgInMemory.ApplyStyles(labelStyle, valueStyle)
		q.msgPersistent.ApplyStyles(labelStyle, valueStyle)
		q.msgTransient.ApplyStyles(labelStyle, valueStyle)
		q.msgPagedOut.ApplyStyles(labelStyle, valueStyle)
		q.msgBytes.ApplyStyles(labelStyle, valueStyle)

		// Configuration
		q.cfgType.ApplyStyles(labelStyle, valueStyle)
		q.cfgDurable.ApplyStyles(labelStyle, valueStyle)
		q.cfgAutoDelete.ApplyStyles(labelStyle, valueStyle)
		q.cfgExclusive.ApplyStyles(labelStyle, valueStyle)
		q.cfgQueueVersion.ApplyStyles(labelStyle, valueStyle)

		// Arguments
		q.argumentsView.SetTextColor(skin.ValueFgColor.Color())
		q.argumentsView.SetBackgroundColor(bgColor)

		// Policy
		q.policyView.SetTextColor(skin.ValueFgColor.Color())
		q.policyView.SetBackgroundColor(bgColor)

		// Message Rates
		q.ratePublish.ApplyStyles(labelStyle, valueStyle)
		q.rateDeliverManualAck.ApplyStyles(labelStyle, valueStyle)
		q.rateDeliverAutoAck.ApplyStyles(labelStyle, valueStyle)
		q.rateConsumerAck.ApplyStyles(labelStyle, valueStyle)
		q.rateRedelivered.ApplyStyles(labelStyle, valueStyle)
		q.rateGetManualAck.ApplyStyles(labelStyle, valueStyle)
		q.rateGetAutoAck.ApplyStyles(labelStyle, valueStyle)

		// State & Resources
		q.stateStatus.ApplyStyles(labelStyle, valueStyle)
		q.stateConsumers.ApplyStyles(labelStyle, valueStyle)
		q.stateCapacity.ApplyStyles(labelStyle, valueStyle)
		q.stateProcessMemory.ApplyStyles(labelStyle, valueStyle)
		q.stateActiveConsumers.ApplyStyles(labelStyle, valueStyle)

		// Node
		q.nodeView.SetTextColor(skin.ValueFgColor.Color())
		q.nodeView.SetBackgroundColor(bgColor)

		// Apply styles to section headers.
		q.messagesHeader.SetTextStyle(captionStyle)
		q.ratesHeader.SetTextStyle(captionStyle)
		q.stateHeader.SetTextStyle(captionStyle)
		q.configHeader.SetTextStyle(captionStyle)
		q.argumentsHeader.SetTextStyle(captionStyle)
		q.policyHeader.SetTextStyle(captionStyle)
		q.nodeHeader.SetTextStyle(captionStyle)
	}
}

func (q *QueueDetails) formatQueueInfoAsText(qi *rabbithole.DetailedQueueInfo) string {
	var b strings.Builder

	// Messages
	b.WriteString("[caption]Messages[-]\n")
	b.WriteString("──────────────────────────────\n")
	b.WriteString(fmt.Sprintf("[label]Ready:[-]               [value]%d[-]\n", qi.MessagesReady))
	b.WriteString(fmt.Sprintf("[label]Unacked:[-]             [value]%d[-]\n", qi.MessagesUnacknowledged))
	b.WriteString(fmt.Sprintf("[label]Total:[-]               [value]%d[-]\n", qi.Messages))
	b.WriteString(fmt.Sprintf("[label]In memory:[-]           [value]%d[-]\n", qi.MessagesRAM))
	b.WriteString(fmt.Sprintf("[label]Persistent:[-]          [value]%d[-]\n", qi.MessagesPersistent))

	transient := qi.Messages - qi.MessagesPersistent
	if transient < 0 {
		transient = 0
	}
	b.WriteString(fmt.Sprintf("[label]Transient:[-]           [value]%d[-]\n", transient))

	pagedOut := qi.Messages - qi.MessagesRAM
	if pagedOut < 0 {
		pagedOut = 0
	}
	b.WriteString(fmt.Sprintf("[label]Paged Out:[-]           [value]%d[-]\n", pagedOut))
	b.WriteString(fmt.Sprintf("[label]Message bytes:[-]       [value]%s[-]\n", view.FormatBytes(qi.MessagesBytes)))
	b.WriteString("\n")

	// Message Rates
	b.WriteString("[caption]Rates (msg/s)[-]\n")
	b.WriteString("──────────────────────────────\n")
	if qi.MessageStats != nil {
		b.WriteString(fmt.Sprintf("[label]Publish:[-]             [value]%.2f[-]\n", qi.MessageStats.PublishDetails.Rate))
		b.WriteString(fmt.Sprintf("[label]Deliver (manual):[-]    [value]%.2f[-]\n", qi.MessageStats.DeliverDetails.Rate))
		b.WriteString(fmt.Sprintf("[label]Deliver (auto):[-]      [value]%.2f[-]\n", qi.MessageStats.DeliverNoAckDetails.Rate))
		b.WriteString(fmt.Sprintf("[label]Consumer ack:[-]        [value]%.2f[-]\n", qi.MessageStats.AckDetails.Rate))
		b.WriteString(fmt.Sprintf("[label]Redelivered:[-]         [value]%.2f[-]\n", qi.MessageStats.RedeliverDetails.Rate))
		b.WriteString(fmt.Sprintf("[label]Get (manual):[-]        [value]%.2f[-]\n", qi.MessageStats.GetDetails.Rate))
		b.WriteString(fmt.Sprintf("[label]Get (auto):[-]          [value]%.2f[-]\n", qi.MessageStats.GetNoAckDetails.Rate))
	} else {
		b.WriteString("[label]Publish:[-]             [value]0.00[-]\n")
		b.WriteString("[label]Deliver (manual):[-]    [value]0.00[-]\n")
		b.WriteString("[label]Deliver (auto):[-]      [value]0.00[-]\n")
		b.WriteString("[label]Consumer ack:[-]        [value]0.00[-]\n")
		b.WriteString("[label]Redelivered:[-]         [value]0.00[-]\n")
		b.WriteString("[label]Get (manual):[-]        [value]0.00[-]\n")
		b.WriteString("[label]Get (auto):[-]          [value]0.00[-]\n")
	}
	b.WriteString("\n")

	// Configuration
	b.WriteString("[caption]Configuration[-]\n")
	b.WriteString("──────────────────────────────\n")
	b.WriteString(fmt.Sprintf("[label]Type:[-]                [value]%s[-]\n", qi.Type))
	b.WriteString(fmt.Sprintf("[label]Durable:[-]             [value]%t[-]\n", qi.Durable))
	b.WriteString(fmt.Sprintf("[label]Auto delete:[-]         [value]%t[-]\n", qi.AutoDelete))
	b.WriteString(fmt.Sprintf("[label]Exclusive:[-]           [value]%t[-]\n", qi.Exclusive))

	if queueVersion, ok := qi.Arguments["x-queue-version"].(float64); ok {
		b.WriteString(fmt.Sprintf("[label]Queue version:[-]       [value]%.0f[-]\n", queueVersion))
	} else {
		b.WriteString("[label]Queue version:[-]       [value]N/A[-]\n")
	}
	b.WriteString("\n")

	// Arguments
	b.WriteString("[caption]Arguments[-]\n")
	b.WriteString("──────────────────────────────\n")
	if len(qi.Arguments) == 0 {
		b.WriteString("[value]N/A[-]\n")
	} else {
		hasArgs := false
		for key, value := range qi.Arguments {
			// Skip arguments already shown in Configuration.
			if key == "x-queue-version" || key == "x-queue-type" {
				continue
			}
			b.WriteString(fmt.Sprintf("[value]%s: %v[-]\n", key, value))
			hasArgs = true
		}
		if !hasArgs {
			b.WriteString("[value]N/A[-]\n")
		}
	}
	b.WriteString("\n")

	// Policy
	b.WriteString("[caption]Policy[-]\n")
	b.WriteString("──────────────────────────────\n")
	if qi.Policy != "" {
		b.WriteString(fmt.Sprintf("[value]%s[-]\n", qi.Policy))
	} else {
		b.WriteString("[value]N/A[-]\n")
	}
	b.WriteString("\n")

	// State & Resources
	b.WriteString("[caption]State & Resources[-]\n")
	b.WriteString("──────────────────────────────\n")
	if qi.Status != "" {
		b.WriteString(fmt.Sprintf("[label]State:[-]               [value]%s[-]\n", qi.Status))
	} else {
		b.WriteString("[label]State:[-]               [value]N/A[-]\n")
	}
	b.WriteString(fmt.Sprintf("[label]Consumers:[-]           [value]%d[-]\n", qi.Consumers))

	if qi.ConsumerUtilisation > 0 {
		b.WriteString(fmt.Sprintf("[label]Consumer capacity:[-]   [value]%.0f%%[-]\n", qi.ConsumerUtilisation*100))
	} else {
		b.WriteString("[label]Consumer capacity:[-]   [value]N/A[-]\n")
	}

	b.WriteString(fmt.Sprintf("[label]Process memory:[-]      [value]%s[-]\n", view.FormatBytes(qi.Memory)))
	b.WriteString(fmt.Sprintf("[label]Active consumers:[-]    [value]%d[-]\n", qi.ActiveConsumers))
	b.WriteString("\n")

	// Node
	b.WriteString("[caption]Node[-]\n")
	b.WriteString("──────────────────────────────────────────────\n")
	if qi.Node != "" {
		b.WriteString(fmt.Sprintf("[value]%s[-]\n", qi.Node))
	} else {
		b.WriteString("[value]N/A[-]\n")
	}

	return b.String()
}
