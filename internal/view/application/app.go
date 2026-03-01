package application

import (
	"context"
	"fmt"
	"log/slog"
	"os"
	"os/signal"
	"syscall"
	"tbunny/internal/cluster"
	"tbunny/internal/config"
	"tbunny/internal/model"
	"tbunny/internal/skins"
	"tbunny/internal/sl"
	"tbunny/internal/ui"
	"tbunny/internal/view/clusters"
	"tbunny/internal/view/connections"
	"tbunny/internal/view/exchanges"
	"tbunny/internal/view/queues"
	"tbunny/internal/view/users"
	"tbunny/internal/view/vhosts"
	"time"

	"github.com/gdamore/tcell/v2"
	"github.com/rivo/tview"
)

const (
	mainPageName   = "main"
	splashPageName = "splash"
)

type App struct {
	*tview.Application

	statusLine *model.StatusLine

	cluster *cluster.Cluster
	config  *config.Config

	Version string

	cancelFn context.CancelFunc

	actions ui.KeyMap

	// Toggles
	headerVisible bool
	crumbsVisible bool
	disableKeys   bool

	// Views
	main         *tview.Pages
	header       *Header
	content      *ViewStack
	crumbs       *ui.Crumbs
	statusLineUi *ui.StatusLine
}

func NewApp(version string) *App {
	a := App{
		Application:   tview.NewApplication(),
		statusLine:    model.NewStatusLine(model.DefaultStatusLineDelay),
		config:        config.Current(),
		Version:       version,
		headerVisible: true,
		crumbsVisible: true,
	}

	a.content = NewViewStack(&a)
	a.main = tview.NewPages()
	a.header = NewHeader(&a)
	a.crumbs = ui.NewCrumbs()
	a.statusLineUi = ui.NewStatusLine(&a)

	cluster.AddListener(&a)
	config.AddListener(&a)
	skins.AddListener(&a)

	flex := tview.NewFlex().SetDirection(tview.FlexRow)
	a.main.AddPage(mainPageName, flex, true, false)
	a.main.AddPage(splashPageName, ui.NewSplash(skins.Current(), a.Version), true, true)

	a.SkinChanged(skins.Current())

	return &a
}

func (a *App) StatusLine() *model.StatusLine {
	return a.statusLine
}

func (a *App) Actions() model.KeyMap {
	return a.actions
}

func (a *App) Init() error {
	ctx := context.Background()
	ctx, a.cancelFn = context.WithCancel(ctx)

	// Add listeners for crumbs and menu
	a.content.AddListener(a.crumbs)

	a.SetRoot(a.main, true).EnableMouse(a.config.UI.EnableMouse)
	a.SetInputCapture(a.keyboard)
	a.bindKeys()

	// Initialize screen components
	go a.statusLineUi.Watch(ctx, a.statusLine.Channel())

	// Handle the current cluster
	activeCluster := cluster.Current()
	if activeCluster != nil {
		a.ClusterChanged(activeCluster)
		a.OpenClusterDefaultView()
	} else {
		a.OpenClustersView()
	}

	a.layout()
	a.initSignals()

	return nil
}

func (a *App) Run() error {
	go func() {
		<-time.After(a.config.UI.SplashDuration)

		a.QueueUpdateDraw(func() {
			a.main.SwitchToPage(mainPageName)
		})
	}()

	return a.Application.Run()
}

func (a *App) ClusterChanged(cluster *cluster.Cluster) {
	if a.cluster != nil {
		a.cluster.RemoveListener(a)
	}

	a.cluster = cluster

	if a.cluster != nil {
		a.cluster.AddListener(a)
	}

	a.bindKeys()
}

func (a *App) ClusterConnectionLost(*cluster.Cluster) {
	a.statusLine.Error("Lost connection to cluster")
}

func (a *App) ClusterConnectionRestored(*cluster.Cluster) {
	a.statusLine.Info("Connection to cluster restored")
}

func (a *App) ConfigChanged(cfg *config.Config) {
	a.config = cfg

	a.EnableMouse(cfg.UI.EnableMouse)
}

func (a *App) SkinChanged(skin *skins.Skin) {
	bgColor := skin.BgColor()

	a.main.SetBackgroundColor(bgColor)
	a.contentFlex().SetBackgroundColor(bgColor)
}

func (a *App) contentFlex() *tview.Flex {
	if f, ok := a.main.GetPage(mainPageName).(*tview.Flex); ok {
		return f
	}

	slog.Error("Main panel not found")

	return nil
}

func (a *App) QueueUpdateDraw(f func()) {
	go func() {
		a.Application.QueueUpdateDraw(f)
	}()
}

func (a *App) layout() {
	f := a.contentFlex()

	f.Clear()

	if a.headerVisible {
		f.AddItem(a.header, 7, 1, false)
	}

	f.AddItem(a.content.Primitive(), 0, 10, true)

	if a.crumbsVisible {
		f.AddItem(a.crumbs, 1, 1, false)
	}

	f.AddItem(a.statusLineUi, 1, 1, false)
}

func (*App) initSignals() {
	sig := make(chan os.Signal, 1)
	signal.Notify(sig, syscall.SIGHUP)

	go func(sig chan os.Signal) {
		<-sig
		os.Exit(0)
	}(sig)
}

func (a *App) bindKeys() {
	m := ui.KeyMap{
		tcell.KeyEscape: ui.NewKeyActionWithGroup("Back/Clear", a.closeViewCmd, false, 0),
		ui.KeyHelp:      ui.NewKeyActionWithGroup("Help", a.helpCmd, false, 1),
		tcell.KeyCtrlC:  ui.NewKeyActionWithGroup("Quit", a.quitCmd, false, 2),
		tcell.KeyCtrlE:  ui.NewKeyActionWithGroup("Toggle header", a.toggleHeaderCmd, false, 3),
		tcell.KeyCtrlG:  ui.NewKeyActionWithGroup("Toggle crumbs", a.toggleCrumbsCmd, false, 3),
	}

	if a.cluster != nil {
		m.Add(ui.KeyShiftQ, ui.NewKeyActionWithGroup("Queues", a.goToQueuesCmd, false, 5))
		m.Add(ui.KeyShiftE, ui.NewKeyActionWithGroup("Exchanges", a.goToExchangesCmd, false, 5))
		m.Add(ui.KeyShiftV, ui.NewKeyActionWithGroup("Virtual hosts", a.goToVHostsCmd, false, 5))
		m.Add(ui.KeyShiftL, ui.NewKeyActionWithGroup("Clusters", a.goToClustersCmd, false, 5))
		m.Add(ui.KeyShiftO, ui.NewKeyActionWithGroup("Connections", a.goToConnectionsCmd, false, 5))
		m.Add(ui.KeyShiftU, ui.NewKeyActionWithGroup("Users", a.goToUsersCmd, false, 5))
	}

	a.actions = m
}

func (a *App) toggleHeaderCmd(*tcell.EventKey) *tcell.EventKey {
	a.headerVisible = !a.headerVisible
	a.layout()

	return nil
}

func (a *App) toggleCrumbsCmd(*tcell.EventKey) *tcell.EventKey {
	a.crumbsVisible = !a.crumbsVisible
	a.layout()

	return nil
}

func (a *App) quitCmd(*tcell.EventKey) *tcell.EventKey {
	a.Application.Stop()
	os.Exit(0)

	return nil
}

func (a *App) helpCmd(*tcell.EventKey) *tcell.EventKey {
	if a.content.Empty() {
		return nil
	}

	top := a.content.Top()
	if top.Name() == helpViewName {
		a.CloseLastView()
		return nil
	}

	_ = a.openView(NewHelp(a), false)

	return nil
}

func (a *App) closeViewCmd(*tcell.EventKey) *tcell.EventKey {
	a.CloseLastView()

	return nil
}

func (a *App) goToQueuesCmd(*tcell.EventKey) *tcell.EventKey {
	err := a.ReplaceOpenViews(queues.NewQueues())
	if err != nil {
		a.statusLine.Error(fmt.Sprintf("Failed to load queues: %s", err))
	}

	return nil
}

func (a *App) goToExchangesCmd(*tcell.EventKey) *tcell.EventKey {
	err := a.ReplaceOpenViews(exchanges.NewExchanges())
	if err != nil {
		a.statusLine.Error(fmt.Sprintf("Failed to load exchanges: %s", err))
	}

	return nil
}

func (a *App) goToVHostsCmd(*tcell.EventKey) *tcell.EventKey {
	err := a.ReplaceOpenViews(vhosts.NewVHosts())
	if err != nil {
		a.statusLine.Error(fmt.Sprintf("Failed to load virtual hosts: %s", err))
	}

	return nil
}

func (a *App) goToConnectionsCmd(*tcell.EventKey) *tcell.EventKey {
	err := a.ReplaceOpenViews(connections.NewConnections())
	if err != nil {
		a.statusLine.Error(fmt.Sprintf("Failed to load connections: %s", err))
	}

	return nil
}

func (a *App) goToUsersCmd(*tcell.EventKey) *tcell.EventKey {
	err := a.ReplaceOpenViews(users.NewView())
	if err != nil {
		a.statusLine.Error(fmt.Sprintf("Failed to load users: %s", err))
	}

	return nil
}

func (a *App) goToClustersCmd(*tcell.EventKey) *tcell.EventKey {
	a.OpenClustersView()

	return nil
}

func (a *App) DisableKeys() {
	a.disableKeys = true
}

func (a *App) EnableKeys() {
	a.disableKeys = false
}

func (a *App) keyboard(event *tcell.EventKey) *tcell.EventKey {
	if a.disableKeys {
		return nil
	}

	ui.RecordLastKey(event)
	if k, ok := a.actions[ui.AsKey(event)]; ok && !a.content.IsTopDialog() {
		return k.Action(event)
	}

	return event
}

func (a *App) AddView(v model.View) error {
	return a.openView(v, false)
}

func (a *App) ReplaceOpenViews(v model.View) error {
	return a.openView(v, true)
}

func (a *App) CloseLastView() {
	if !a.content.Last() {
		a.content.Pop()
	}
}

func (a *App) openView(v model.View, clearStack bool) error {
	if err := v.Init(a); err != nil {
		slog.Error("View init failed",
			sl.Error, err,
			sl.Component, v.Name())
		return err
	}

	if clearStack {
		a.content.Clear()
	}
	a.content.Push(v)

	return nil
}

func (a *App) OpenClusterDefaultView() {
	if a.cluster == nil {
		return
	}

	err := a.ReplaceOpenViews(queues.NewQueues())
	if err != nil {
		a.statusLine.Error(fmt.Sprintf("Failed to load queues: %s", err))
	}
}

func (a *App) OpenClustersView() {
	err := a.ReplaceOpenViews(clusters.NewClusters())
	if err != nil {
		a.statusLine.Error(fmt.Sprintf("Failed to load clusters: %s", err))
	}
}

func (a *App) ShowModal(modal tview.Primitive) {
	a.content.ShowModal(modal)
}

func (a *App) DismissModal() {
	a.content.DismissModal()
}
