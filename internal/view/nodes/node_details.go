package nodes

import (
	"fmt"
	"strings"
	"tbunny/internal/model"
	"tbunny/internal/skins"
	"tbunny/internal/ui"
	"tbunny/internal/view"

	"github.com/gdamore/tcell/v2"
	rabbithole "github.com/michaelklishin/rabbit-hole/v3"
	"github.com/rivo/tview"
)

const nodeDetailsTitleFmt = " [fg:bg:b]%s[fg:bg:-]([hilite:bg:b]%s[fg:bg:-]) "

const (
	// Minimum inner width for the four-gauge columnar layout.
	minWidthForFourGauges = 105
)

// NodeDetails is a cluster-aware refreshable view that shows a health dashboard for a single node.
type NodeDetails struct {
	*view.ClusterAwareRefreshableView[*tview.Flex]

	name string
	node *rabbithole.NodeInfo
	skin *skins.Skin

	useScrollableMode bool
	relayoutPending   bool

	// Top section: four gauge panels.
	gaugesPanel *tview.Flex

	// Memory gauge.
	memUsed  *view.LabelAndValue
	memLimit *view.LabelAndValue
	memAlarm *view.LabelAndValue
	memBar   *tview.TextView

	// Disk gauge.
	diskFree  *view.LabelAndValue
	diskLimit *view.LabelAndValue
	diskAlarm *view.LabelAndValue
	diskBar   *tview.TextView

	// File Descriptors gauge.
	fdUsed  *view.LabelAndValue
	fdTotal *view.LabelAndValue
	fdBar   *tview.TextView

	// Processes gauge.
	procUsed  *view.LabelAndValue
	procTotal *view.LabelAndValue
	procBar   *tview.TextView

	// Bottom section: three detail columns.
	detailsPanel *tview.Flex

	// Overview column.
	ovType       *view.LabelAndValue
	ovRunning    *view.LabelAndValue
	ovUptime     *view.LabelAndValue
	ovOsPid      *view.LabelAndValue
	ovRunQueue   *view.LabelAndValue
	ovProcessors *view.LabelAndValue

	// Connections column.
	connCreated *view.LabelAndValue
	connClosed  *view.LabelAndValue
	chanCreated *view.LabelAndValue
	chanClosed  *view.LabelAndValue

	// Queue Stats (inside connections column).
	qsDeclared *view.LabelAndValue
	qsCreated  *view.LabelAndValue
	qsDeleted  *view.LabelAndValue

	// I/O column.
	ioReadCount  *view.LabelAndValue
	ioReadBytes  *view.LabelAndValue
	ioWriteCount *view.LabelAndValue
	ioWriteBytes *view.LabelAndValue

	// Section headers.
	memoryHeader    *tview.TextView
	diskHeader      *tview.TextView
	fdHeader        *tview.TextView
	processesHeader *tview.TextView
	overviewHeader  *tview.TextView
	connHeader      *tview.TextView
	queueHeader     *tview.TextView
	ioHeader        *tview.TextView

	// Separators and containers that need background styling.
	separators []*tview.TextView
	containers []*tview.Box

	// For scrollable (narrow) mode.
	scrollableView *tview.TextView
}

// NewNodeDetails creates and returns a detail view for the given node name.
func NewNodeDetails(name string) *NodeDetails {
	flex := tview.NewFlex().SetDirection(tview.FlexRow)
	flex.SetBorder(true).SetBorderPadding(1, 0, 1, 1)

	strategy := view.NewLiveUpdateStrategy()

	v := &NodeDetails{
		ClusterAwareRefreshableView: view.NewClusterAwareRefreshableView[*tview.Flex]("Node Details", flex, strategy),
		name:                        name,
	}

	v.SetUpdateFn(v.performUpdate)
	v.AddBindingKeysFn(v.bindScrollKeys)

	return v
}

func (v *NodeDetails) Init(app model.App) error {
	err := v.ClusterAwareRefreshableView.Init(app)
	if err != nil {
		return err
	}

	v.skin = skins.Current()

	// Start in scrollable mode as a safe fallback.
	// The actual mode is determined on the first update.
	v.useScrollableMode = true
	v.createLayout()
	v.updateTitle()

	// Detect terminal resize and switch layout mode without waiting for the next data refresh.
	v.Ui().SetDrawFunc(func(screen tcell.Screen, x, y, width, height int) (int, int, int, int) {
		innerX, innerY, innerW, innerH := v.Ui().GetInnerRect()
		if v.node != nil && (innerW < minWidthForFourGauges) != v.useScrollableMode && !v.relayoutPending {
			v.relayoutPending = true
			v.App().QueueUpdateDraw(func() {
				v.relayoutPending = false
				_, _, w, _ := v.Ui().GetInnerRect()
				previousMode := v.useScrollableMode
				v.useScrollableMode = w < minWidthForFourGauges
				if previousMode != v.useScrollableMode {
					v.createLayout()
					if v.useScrollableMode {
						v.updateScrollableView()
					} else {
						v.updateColumnarView()
					}
				}
			})
		}
		return innerX, innerY, innerW, innerH
	})

	return nil
}

func (v *NodeDetails) determineLayoutMode() {
	_, _, width, _ := v.Ui().GetInnerRect()

	if width == 0 {
		v.useScrollableMode = false
		return
	}

	v.useScrollableMode = width < minWidthForFourGauges
}

func (v *NodeDetails) performUpdate(view.UpdateKind) {
	node, err := v.Cluster().GetNode(v.name)
	if err != nil {
		v.App().StatusLine().Errorf("Failed to fetch node details: %v", err)
		return
	}

	v.App().QueueUpdateDraw(func() {
		v.node = node
		previousMode := v.useScrollableMode
		v.determineLayoutMode()

		if previousMode != v.useScrollableMode {
			v.createLayout()
		}

		if v.useScrollableMode {
			v.updateScrollableView()
		} else {
			v.updateColumnarView()
		}
	})
}

func (v *NodeDetails) updateTitle() {
	title := view.SkinTitle(fmt.Sprintf(nodeDetailsTitleFmt, v.Name(), v.name))
	v.Ui().SetTitle(title)
}

func (v *NodeDetails) bindScrollKeys(km ui.KeyMap) {
	scroll := func(delta int) ui.ActionHandler {
		return func(e *tcell.EventKey) *tcell.EventKey {
			if !v.useScrollableMode || v.scrollableView == nil {
				return e
			}
			row, col := v.scrollableView.GetScrollOffset()
			v.scrollableView.ScrollTo(row+delta, col)
			return nil
		}
	}

	km.Add(tcell.KeyUp, ui.NewHiddenKeyAction("Scroll Up", scroll(-1)))
	km.Add(tcell.KeyDown, ui.NewHiddenKeyAction("Scroll Down", scroll(1)))
	km.Add(tcell.KeyPgUp, ui.NewHiddenKeyAction("Page Up", scroll(-10)))
	km.Add(tcell.KeyPgDn, ui.NewHiddenKeyAction("Page Down", scroll(10)))
	km.Add(tcell.KeyHome, ui.NewHiddenKeyAction("Scroll to Top", func(e *tcell.EventKey) *tcell.EventKey {
		if !v.useScrollableMode || v.scrollableView == nil {
			return e
		}
		v.scrollableView.ScrollToBeginning()
		return nil
	}))
	km.Add(tcell.KeyEnd, ui.NewHiddenKeyAction("Scroll to End", func(e *tcell.EventKey) *tcell.EventKey {
		if !v.useScrollableMode || v.scrollableView == nil {
			return e
		}
		v.scrollableView.ScrollToEnd()
		return nil
	}))
}

func (v *NodeDetails) createLayout() {
	v.Ui().Clear()

	if v.useScrollableMode {
		v.createScrollableLayout()
	} else {
		v.createFourGaugeLayout()
	}

	v.applyStyles()
}

func (v *NodeDetails) createScrollableLayout() {
	v.scrollableView = tview.NewTextView()
	v.scrollableView.SetDynamicColors(true)
	v.scrollableView.SetScrollable(true)
	v.scrollableView.SetWordWrap(false)

	v.Ui().AddItem(v.scrollableView, 0, 1, false)
}

func (v *NodeDetails) createFourGaugeLayout() {
	v.separators = nil
	v.containers = nil

	// Top row: four gauge panels centered horizontally.
	v.gaugesPanel = tview.NewFlex().SetDirection(tview.FlexColumn)
	leftPad1 := tview.NewFlex()
	rightPad1 := tview.NewFlex()
	v.containers = append(v.containers, leftPad1.Box, rightPad1.Box)
	gaugeSep1 := tview.NewBox()
	gaugeSep2 := tview.NewBox()
	gaugeSep3 := tview.NewBox()
	v.containers = append(v.containers, gaugeSep1, gaugeSep2, gaugeSep3)

	v.gaugesPanel.AddItem(leftPad1, 0, 1, false)
	v.gaugesPanel.AddItem(v.createMemoryGauge(), 25, 0, false)
	v.gaugesPanel.AddItem(gaugeSep1, 2, 0, false)
	v.gaugesPanel.AddItem(v.createDiskGauge(), 25, 0, false)
	v.gaugesPanel.AddItem(gaugeSep2, 2, 0, false)
	v.gaugesPanel.AddItem(v.createFDGauge(), 25, 0, false)
	v.gaugesPanel.AddItem(gaugeSep3, 2, 0, false)
	v.gaugesPanel.AddItem(v.createProcessesGauge(), 25, 0, false)
	v.gaugesPanel.AddItem(rightPad1, 0, 1, false)

	// Bottom row: three detail columns centered horizontally.
	v.detailsPanel = tview.NewFlex().SetDirection(tview.FlexColumn)
	leftPad2 := tview.NewFlex()
	rightPad2 := tview.NewFlex()
	v.containers = append(v.containers, leftPad2.Box, rightPad2.Box)
	v.detailsPanel.AddItem(leftPad2, 0, 1, false)
	v.detailsPanel.AddItem(v.createOverviewSection(), 28, 0, false)
	v.detailsPanel.AddItem(v.createConnectionsSection(), 28, 0, false)
	v.detailsPanel.AddItem(v.createIOSection(), 28, 0, false)
	v.detailsPanel.AddItem(rightPad2, 0, 1, false)

	// Gauges: 1 (header) + 1 (sep) + 3 (grid rows) + 1 (bar) = 6 max (Memory/Disk with alarm).
	v.Ui().AddItem(v.gaugesPanel, 6, 0, false)
	blankBox := tview.NewBox()
	v.containers = append(v.containers, blankBox)
	v.Ui().AddItem(blankBox, 1, 0, false)
	v.Ui().AddItem(v.detailsPanel, 0, 1, false)
}

func (v *NodeDetails) createSeparator(width int) *tview.TextView {
	sep := tview.NewTextView().SetText(strings.Repeat("─", width))
	v.separators = append(v.separators, sep)
	return sep
}

func (v *NodeDetails) createMemoryGauge() *tview.Flex {
	panel := tview.NewFlex().SetDirection(tview.FlexRow)

	v.memoryHeader = tview.NewTextView().SetText("Memory").SetDynamicColors(true)
	panel.AddItem(v.memoryHeader, 1, 0, false)
	panel.AddItem(v.createSeparator(23), 1, 0, false)

	v.memUsed = view.NewLabelAndValue("Used:")
	v.memLimit = view.NewLabelAndValue("Limit:")
	v.memAlarm = view.NewLabelAndValue("Alarm:")
	v.memBar = tview.NewTextView().SetDynamicColors(true)

	grid := tview.NewGrid().SetColumns(11, -1).SetRows(1, 1, 1).SetGap(0, 0)
	v.memUsed.AddToGrid(grid, 0, 0)
	v.memLimit.AddToGrid(grid, 1, 0)
	v.memAlarm.AddToGrid(grid, 2, 0)

	panel.AddItem(grid, 3, 0, false)
	panel.AddItem(v.memBar, 1, 0, false)

	return panel
}

func (v *NodeDetails) createDiskGauge() *tview.Flex {
	panel := tview.NewFlex().SetDirection(tview.FlexRow)

	v.diskHeader = tview.NewTextView().SetText("Disk").SetDynamicColors(true)
	panel.AddItem(v.diskHeader, 1, 0, false)
	panel.AddItem(v.createSeparator(23), 1, 0, false)

	v.diskFree = view.NewLabelAndValue("Free:")
	v.diskLimit = view.NewLabelAndValue("Limit:")
	v.diskAlarm = view.NewLabelAndValue("Alarm:")
	v.diskBar = tview.NewTextView().SetDynamicColors(true)

	grid := tview.NewGrid().SetColumns(11, -1).SetRows(1, 1, 1).SetGap(0, 0)
	v.diskFree.AddToGrid(grid, 0, 0)
	v.diskLimit.AddToGrid(grid, 1, 0)
	v.diskAlarm.AddToGrid(grid, 2, 0)

	panel.AddItem(grid, 3, 0, false)
	panel.AddItem(v.diskBar, 1, 0, false)

	return panel
}

func (v *NodeDetails) createFDGauge() *tview.Flex {
	panel := tview.NewFlex().SetDirection(tview.FlexRow)

	v.fdHeader = tview.NewTextView().SetText("File Descriptors").SetDynamicColors(true)
	panel.AddItem(v.fdHeader, 1, 0, false)
	panel.AddItem(v.createSeparator(23), 1, 0, false)

	v.fdUsed = view.NewLabelAndValue("Used:")
	v.fdTotal = view.NewLabelAndValue("Total:")
	v.fdBar = tview.NewTextView().SetDynamicColors(true)

	grid := tview.NewGrid().SetColumns(11, -1).SetRows(1, 1, 1).SetGap(0, 0)
	v.fdUsed.AddToGrid(grid, 0, 0)
	v.fdTotal.AddToGrid(grid, 1, 0)
	// row 2 left empty to align bar with Memory/Disk gauges

	panel.AddItem(grid, 3, 0, false)
	panel.AddItem(v.fdBar, 1, 0, false)

	return panel
}

func (v *NodeDetails) createProcessesGauge() *tview.Flex {
	panel := tview.NewFlex().SetDirection(tview.FlexRow)

	v.processesHeader = tview.NewTextView().SetText("Processes").SetDynamicColors(true)
	panel.AddItem(v.processesHeader, 1, 0, false)
	panel.AddItem(v.createSeparator(23), 1, 0, false)

	v.procUsed = view.NewLabelAndValue("Used:")
	v.procTotal = view.NewLabelAndValue("Total:")
	v.procBar = tview.NewTextView().SetDynamicColors(true)

	grid := tview.NewGrid().SetColumns(11, -1).SetRows(1, 1, 1).SetGap(0, 0)
	v.procUsed.AddToGrid(grid, 0, 0)
	v.procTotal.AddToGrid(grid, 1, 0)
	// row 2 left empty to align bar with Memory/Disk gauges

	panel.AddItem(grid, 3, 0, false)
	panel.AddItem(v.procBar, 1, 0, false)

	return panel
}

func (v *NodeDetails) createOverviewSection() *tview.Flex {
	section := tview.NewFlex().SetDirection(tview.FlexRow)

	v.overviewHeader = tview.NewTextView().SetText("Overview").SetDynamicColors(true)
	section.AddItem(v.overviewHeader, 1, 0, false)
	section.AddItem(v.createSeparator(26), 1, 0, false)

	grid := tview.NewGrid().SetColumns(13, -1).SetRows(1, 1, 1, 1, 1, 1).SetGap(0, 0)

	v.ovType = view.NewLabelAndValue("Type:")
	v.ovRunning = view.NewLabelAndValue("Running:")
	v.ovUptime = view.NewLabelAndValue("Uptime:")
	v.ovOsPid = view.NewLabelAndValue("OS PID:")
	v.ovRunQueue = view.NewLabelAndValue("Run queue:")
	v.ovProcessors = view.NewLabelAndValue("Processors:")

	v.ovType.AddToGrid(grid, 0, 0)
	v.ovRunning.AddToGrid(grid, 1, 0)
	v.ovUptime.AddToGrid(grid, 2, 0)
	v.ovOsPid.AddToGrid(grid, 3, 0)
	v.ovRunQueue.AddToGrid(grid, 4, 0)
	v.ovProcessors.AddToGrid(grid, 5, 0)

	section.AddItem(grid, 6, 0, false)

	return section
}

func (v *NodeDetails) createConnectionsSection() *tview.Flex {
	section := tview.NewFlex().SetDirection(tview.FlexRow)

	// Connections sub-section.
	v.connHeader = tview.NewTextView().SetText("Connections").SetDynamicColors(true)
	section.AddItem(v.connHeader, 1, 0, false)
	section.AddItem(v.createSeparator(26), 1, 0, false)

	connGrid := tview.NewGrid().SetColumns(13, -1).SetRows(1, 1, 1, 1).SetGap(0, 0)

	v.connCreated = view.NewLabelAndValue("Created:")
	v.connClosed = view.NewLabelAndValue("Closed:")
	v.chanCreated = view.NewLabelAndValue("Chan open:")
	v.chanClosed = view.NewLabelAndValue("Chan close:")

	v.connCreated.AddToGrid(connGrid, 0, 0)
	v.connClosed.AddToGrid(connGrid, 1, 0)
	v.chanCreated.AddToGrid(connGrid, 2, 0)
	v.chanClosed.AddToGrid(connGrid, 3, 0)

	section.AddItem(connGrid, 4, 0, false)

	// Empty line between Connections and Queue Stats.
	connQsSpacer := tview.NewBox()
	v.containers = append(v.containers, connQsSpacer)
	section.AddItem(connQsSpacer, 1, 0, false)

	// Queue Stats sub-section.
	v.queueHeader = tview.NewTextView().SetText("Queue Stats").SetDynamicColors(true)
	section.AddItem(v.queueHeader, 1, 0, false)
	section.AddItem(v.createSeparator(26), 1, 0, false)

	qsGrid := tview.NewGrid().SetColumns(13, -1).SetRows(1, 1, 1).SetGap(0, 0)

	v.qsDeclared = view.NewLabelAndValue("Declared:")
	v.qsCreated = view.NewLabelAndValue("Created:")
	v.qsDeleted = view.NewLabelAndValue("Deleted:")

	v.qsDeclared.AddToGrid(qsGrid, 0, 0)
	v.qsCreated.AddToGrid(qsGrid, 1, 0)
	v.qsDeleted.AddToGrid(qsGrid, 2, 0)

	section.AddItem(qsGrid, 3, 0, false)

	return section
}

func (v *NodeDetails) createIOSection() *tview.Flex {
	section := tview.NewFlex().SetDirection(tview.FlexRow)

	v.ioHeader = tview.NewTextView().SetText("I/O").SetDynamicColors(true)
	section.AddItem(v.ioHeader, 1, 0, false)
	section.AddItem(v.createSeparator(26), 1, 0, false)

	grid := tview.NewGrid().SetColumns(13, -1).SetRows(1, 1, 1, 1).SetGap(0, 0)

	v.ioReadCount = view.NewLabelAndValue("Read cnt:")
	v.ioReadBytes = view.NewLabelAndValue("Read bytes:")
	v.ioWriteCount = view.NewLabelAndValue("Write cnt:")
	v.ioWriteBytes = view.NewLabelAndValue("Write bytes:")

	v.ioReadCount.AddToGrid(grid, 0, 0)
	v.ioReadBytes.AddToGrid(grid, 1, 0)
	v.ioWriteCount.AddToGrid(grid, 2, 0)
	v.ioWriteBytes.AddToGrid(grid, 3, 0)

	section.AddItem(grid, 4, 0, false)

	return section
}

func (v *NodeDetails) applyStyles() {
	v.updateTitle()

	skin := v.skin.Views.Stats
	bgColor := skin.BgColor.Color()

	v.Ui().SetBackgroundColor(bgColor)

	if v.useScrollableMode {
		v.scrollableView.SetTextColor(skin.ValueFgColor.Color())
		v.scrollableView.SetBackgroundColor(bgColor)
	} else {
		labelStyle := tcell.StyleDefault.Foreground(skin.LabelFgColor.Color()).Background(bgColor)
		valueStyle := tcell.StyleDefault.Foreground(skin.ValueFgColor.Color()).Background(bgColor)
		captionStyle := tcell.StyleDefault.Foreground(skin.CaptionFgColor.Color()).Background(bgColor)

		// Memory gauge.
		v.memUsed.ApplyStyles(labelStyle, valueStyle)
		v.memLimit.ApplyStyles(labelStyle, valueStyle)
		v.memAlarm.ApplyStyles(labelStyle, valueStyle)
		v.memBar.SetBackgroundColor(bgColor)

		// Disk gauge.
		v.diskFree.ApplyStyles(labelStyle, valueStyle)
		v.diskLimit.ApplyStyles(labelStyle, valueStyle)
		v.diskAlarm.ApplyStyles(labelStyle, valueStyle)
		v.diskBar.SetBackgroundColor(bgColor)

		// FD gauge.
		v.fdUsed.ApplyStyles(labelStyle, valueStyle)
		v.fdTotal.ApplyStyles(labelStyle, valueStyle)
		v.fdBar.SetBackgroundColor(bgColor)

		// Processes gauge.
		v.procUsed.ApplyStyles(labelStyle, valueStyle)
		v.procTotal.ApplyStyles(labelStyle, valueStyle)
		v.procBar.SetBackgroundColor(bgColor)

		// Overview.
		v.ovType.ApplyStyles(labelStyle, valueStyle)
		v.ovRunning.ApplyStyles(labelStyle, valueStyle)
		v.ovUptime.ApplyStyles(labelStyle, valueStyle)
		v.ovOsPid.ApplyStyles(labelStyle, valueStyle)
		v.ovRunQueue.ApplyStyles(labelStyle, valueStyle)
		v.ovProcessors.ApplyStyles(labelStyle, valueStyle)

		// Connections.
		v.connCreated.ApplyStyles(labelStyle, valueStyle)
		v.connClosed.ApplyStyles(labelStyle, valueStyle)
		v.chanCreated.ApplyStyles(labelStyle, valueStyle)
		v.chanClosed.ApplyStyles(labelStyle, valueStyle)

		// Queue Stats.
		v.qsDeclared.ApplyStyles(labelStyle, valueStyle)
		v.qsCreated.ApplyStyles(labelStyle, valueStyle)
		v.qsDeleted.ApplyStyles(labelStyle, valueStyle)

		// I/O.
		v.ioReadCount.ApplyStyles(labelStyle, valueStyle)
		v.ioReadBytes.ApplyStyles(labelStyle, valueStyle)
		v.ioWriteCount.ApplyStyles(labelStyle, valueStyle)
		v.ioWriteBytes.ApplyStyles(labelStyle, valueStyle)

		// Section headers.
		v.memoryHeader.SetTextStyle(captionStyle)
		v.diskHeader.SetTextStyle(captionStyle)
		v.fdHeader.SetTextStyle(captionStyle)
		v.processesHeader.SetTextStyle(captionStyle)
		v.overviewHeader.SetTextStyle(captionStyle)
		v.connHeader.SetTextStyle(captionStyle)
		v.queueHeader.SetTextStyle(captionStyle)
		v.ioHeader.SetTextStyle(captionStyle)

		// Separators.
		sepStyle := tcell.StyleDefault.Foreground(skin.CaptionFgColor.Color()).Background(bgColor)
		for _, sep := range v.separators {
			sep.SetTextStyle(sepStyle)
		}

		// Padding and container widgets.
		for _, c := range v.containers {
			c.SetBackgroundColor(bgColor)
		}

		v.gaugesPanel.SetBackgroundColor(bgColor)
		v.detailsPanel.SetBackgroundColor(bgColor)
	}
}

func (v *NodeDetails) updateColumnarView() {
	n := v.node

	// Memory gauge.
	memPct := 0.0
	if n.MemLimit > 0 {
		memPct = float64(n.MemUsed) / float64(n.MemLimit) * 100
	}
	v.memUsed.SetBytes(int64(n.MemUsed))
	v.memLimit.SetBytes(int64(n.MemLimit))
	v.memAlarm.SetBool(n.MemAlarm)
	color := view.GetStateColor(memPct, v.skin)
	v.memBar.SetText(color + view.RenderProgressBar(memPct, 20, view.DefaultProgressBarStyle) + "[-]")

	// Disk gauge.
	// Disk: percentage = how full the disk is. We don't have total, so use alarm threshold approach.
	// For disk, "free" is what matters — show how close free is to the limit (lower = worse).
	diskPct := 0.0
	if n.DiskFreeLimit > 0 {
		if n.DiskFree == 0 {
			diskPct = 100
		} else {
			// Ratio of limit to free: as free approaches limit, percentage grows toward 100%.
			diskPct = float64(n.DiskFreeLimit) / float64(n.DiskFree) * 100
			if diskPct > 100 {
				diskPct = 100
			}
		}
	}
	v.diskFree.SetBytes(int64(n.DiskFree))
	v.diskLimit.SetBytes(int64(n.DiskFreeLimit))
	v.diskAlarm.SetBool(n.DiskFreeAlarm)
	color = view.GetStateColor(diskPct, v.skin)
	v.diskBar.SetText(color + view.RenderProgressBar(diskPct, 20, view.DefaultProgressBarStyle) + "[-]")

	// File Descriptors gauge.
	fdPct := 0.0
	if n.FdTotal > 0 {
		fdPct = float64(n.FdUsed) / float64(n.FdTotal) * 100
	}
	v.fdUsed.SetText(fmt.Sprintf("%d", n.FdUsed))
	v.fdTotal.SetText(fmt.Sprintf("%d", n.FdTotal))
	color = view.GetStateColor(fdPct, v.skin)
	v.fdBar.SetText(color + view.RenderProgressBar(fdPct, 20, view.DefaultProgressBarStyle) + "[-]")

	// Processes gauge.
	procPct := 0.0
	if n.ProcTotal > 0 {
		procPct = float64(n.ProcUsed) / float64(n.ProcTotal) * 100
	}
	v.procUsed.SetText(fmt.Sprintf("%d", n.ProcUsed))
	v.procTotal.SetText(fmt.Sprintf("%d", n.ProcTotal))
	color = view.GetStateColor(procPct, v.skin)
	v.procBar.SetText(color + view.RenderProgressBar(procPct, 20, view.DefaultProgressBarStyle) + "[-]")

	// Overview column.
	v.ovType.SetText(n.NodeType)
	v.ovRunning.SetBool(n.IsRunning)
	v.ovUptime.SetText(formatUptime(n.Uptime))
	v.ovOsPid.SetText(string(n.OsPid))
	v.ovRunQueue.SetText(fmt.Sprintf("%d", n.RunQueueLength))
	v.ovProcessors.SetText(fmt.Sprintf("%d", n.Processors))

	// Connections column.
	v.connCreated.SetText(fmt.Sprintf("%d", n.ConnectionCreated))
	v.connClosed.SetText(fmt.Sprintf("%d", n.ConnectionClosed))
	v.chanCreated.SetText(fmt.Sprintf("%d", n.ChannelCreated))
	v.chanClosed.SetText(fmt.Sprintf("%d", n.ChannelClosed))

	// Queue Stats.
	v.qsDeclared.SetText(fmt.Sprintf("%d", n.QueueDeclared))
	v.qsCreated.SetText(fmt.Sprintf("%d", n.QueueCreated))
	v.qsDeleted.SetText(fmt.Sprintf("%d", n.QueueDeleted))

	// I/O column.
	v.ioReadCount.SetText(fmt.Sprintf("%d", n.IOReadCount))
	v.ioReadBytes.SetBytes(int64(n.IOReadBytes))
	v.ioWriteCount.SetText(fmt.Sprintf("%d", n.IOWriteCount))
	v.ioWriteBytes.SetBytes(int64(n.IOWriteBytes))
}

func (v *NodeDetails) updateScrollableView() {
	n := v.node
	b := new(strings.Builder)

	// Memory.
	memPct := 0.0
	if n.MemLimit > 0 {
		memPct = float64(n.MemUsed) / float64(n.MemLimit) * 100
	}
	view.WriteTextSection(b, "Memory", []view.TextRow{
		{Label: "Used:", Value: view.FormatBytes(n.MemUsed)},
		{Label: "Limit:", Value: view.FormatBytes(n.MemLimit)},
		{Label: "Alarm:", Value: view.FormatBool(n.MemAlarm)},
		{Label: "Usage:", Value: view.GetStateColor(memPct, v.skin) + view.RenderProgressBar(memPct, 20, view.DefaultProgressBarStyle) + "[-]"},
	})

	// Disk.
	diskPct := 0.0
	if n.DiskFreeLimit > 0 {
		if n.DiskFree == 0 {
			diskPct = 100
		} else {
			diskPct = float64(n.DiskFreeLimit) / float64(n.DiskFree) * 100
			if diskPct > 100 {
				diskPct = 100
			}
		}
	}
	view.WriteTextSection(b, "Disk", []view.TextRow{
		{Label: "Free:", Value: view.FormatBytes(n.DiskFree)},
		{Label: "Limit:", Value: view.FormatBytes(n.DiskFreeLimit)},
		{Label: "Alarm:", Value: view.FormatBool(n.DiskFreeAlarm)},
		{Label: "Usage:", Value: view.GetStateColor(diskPct, v.skin) + view.RenderProgressBar(diskPct, 20, view.DefaultProgressBarStyle) + "[-]"},
	})

	// File Descriptors.
	fdPct := 0.0
	if n.FdTotal > 0 {
		fdPct = float64(n.FdUsed) / float64(n.FdTotal) * 100
	}
	view.WriteTextSection(b, "File Descriptors", []view.TextRow{
		{Label: "Used:", Value: fmt.Sprintf("%d", n.FdUsed)},
		{Label: "Total:", Value: fmt.Sprintf("%d", n.FdTotal)},
		{Label: "Usage:", Value: view.GetStateColor(fdPct, v.skin) + view.RenderProgressBar(fdPct, 20, view.DefaultProgressBarStyle) + "[-]"},
	})

	// Processes.
	procPct := 0.0
	if n.ProcTotal > 0 {
		procPct = float64(n.ProcUsed) / float64(n.ProcTotal) * 100
	}
	view.WriteTextSection(b, "Processes", []view.TextRow{
		{Label: "Used:", Value: fmt.Sprintf("%d", n.ProcUsed)},
		{Label: "Total:", Value: fmt.Sprintf("%d", n.ProcTotal)},
		{Label: "Usage:", Value: view.GetStateColor(procPct, v.skin) + view.RenderProgressBar(procPct, 20, view.DefaultProgressBarStyle) + "[-]"},
	})

	// Overview.
	view.WriteTextSection(b, "Overview", []view.TextRow{
		{Label: "Name:", Value: n.Name},
		{Label: "Type:", Value: n.NodeType},
		{Label: "Running:", Value: view.FormatBool(n.IsRunning)},
		{Label: "Uptime:", Value: formatUptime(n.Uptime)},
		{Label: "OS PID:", Value: string(n.OsPid)},
		{Label: "Run queue:", Value: fmt.Sprintf("%d", n.RunQueueLength)},
		{Label: "Processors:", Value: fmt.Sprintf("%d", n.Processors)},
	})

	// Connections.
	view.WriteTextSection(b, "Connections", []view.TextRow{
		{Label: "Created:", Value: fmt.Sprintf("%d", n.ConnectionCreated)},
		{Label: "Closed:", Value: fmt.Sprintf("%d", n.ConnectionClosed)},
		{Label: "Chan open:", Value: fmt.Sprintf("%d", n.ChannelCreated)},
		{Label: "Chan close:", Value: fmt.Sprintf("%d", n.ChannelClosed)},
	})

	// Queue Stats.
	view.WriteTextSection(b, "Queue Stats", []view.TextRow{
		{Label: "Declared:", Value: fmt.Sprintf("%d", n.QueueDeclared)},
		{Label: "Created:", Value: fmt.Sprintf("%d", n.QueueCreated)},
		{Label: "Deleted:", Value: fmt.Sprintf("%d", n.QueueDeleted)},
	})

	// I/O.
	view.WriteTextSection(b, "I/O", []view.TextRow{
		{Label: "Read count:", Value: fmt.Sprintf("%d", n.IOReadCount)},
		{Label: "Read bytes:", Value: view.FormatBytes(int64(n.IOReadBytes))},
		{Label: "Write count:", Value: fmt.Sprintf("%d", n.IOWriteCount)},
		{Label: "Write bytes:", Value: view.FormatBytes(int64(n.IOWriteBytes))},
	})

	styledContent := view.SkinStatsContent(b.String(), &v.skin.Views.Stats)
	v.scrollableView.SetText(styledContent)
}
