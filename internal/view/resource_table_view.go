package view

import (
	"fmt"
	"strings"
	"sync"
	"tbunny/internal/skins"
	"tbunny/internal/ui"
	"tbunny/internal/ui/dialog"

	"github.com/gdamore/tcell/v2"
)

const (
	titleNameFragmentFmt   = "[fg:bg:b]%s"
	titlePathFragmentFmt   = "([hilite:bg:b]%s[fg:bg:-])"
	titleCountFragmentFmt  = "[fg:bg:-][[count:bg:b]%d[fg:bg:-]][fg:bg:-]"
	titleFilterFragmentFmt = " [fg:bg:-]<[filter:bg:b]/%s[fg:bg:-]>[fg:bg:-]"
)

type ResourceTableView[R Resource] struct {
	*RefreshableView[*ui.Table[R]]

	resourceProvider ResourceProvider[R]
	path             string
	enterActionTitle string
	enterActionFn    func(R)
	filter           string
	resources        []R
	mx               sync.RWMutex
}

func NewResourceTableView[R Resource](name string, strategy UpdateStrategy) *ResourceTableView[R] {
	r := ResourceTableView[R]{
		RefreshableView: NewRefreshableView[*ui.Table[R]](name, ui.NewTable[R](), strategy),
	}

	r.AddBindingKeysFn(r.bindKeys)
	r.SetUpdateFn(r.performUpdate)

	return &r
}

func (b *ResourceTableView[R]) GetSelectedResource() (R, bool) {
	return b.Ui().GetSelectedRow()
}

func (b *ResourceTableView[R]) SetResourceProvider(rp ResourceProvider[R]) {
	b.resourceProvider = rp
}

func (b *ResourceTableView[R]) SetPath(path string) {
	b.path = path
}

func (b *ResourceTableView[R]) SetEnterAction(title string, fn func(R)) {
	b.enterActionTitle = title
	b.enterActionFn = fn
}

func (b *ResourceTableView[R]) AddBindingKeysFn(fn ui.BindingKeysFn) {
	b.RefreshableView.AddBindingKeysFn(fn)
}

func (b *ResourceTableView[R]) Start() {
	b.RefreshableView.Start()

	skins.AddListener(b)
	b.SkinChanged(skins.Current())
}

func (b *ResourceTableView[R]) Stop() {
	b.RefreshableView.Stop()
	skins.RemoveListener(b)
}

func (b *ResourceTableView[R]) SkinChanged(skin *skins.Skin) {
	b.Ui().ApplySkin(skin)
	b.updateTitle()
}

func (b *ResourceTableView[R]) Filter(filter string) {
	b.mx.Lock()
	b.filter = filter
	b.mx.Unlock()

	b.App().QueueUpdateDraw(b.filterAndSet)
}

func (b *ResourceTableView[R]) Clear() bool {
	if b.filter == "" {
		return false
	}

	b.filter = ""
	b.App().QueueUpdateDraw(b.filterAndSet)

	return true
}

func (b *ResourceTableView[R]) performUpdate(kind UpdateKind) {
	rp := b.resourceProviderWithCheck()

	rows, err := rp.GetResources()
	if err != nil {
		b.app.StatusLine().Error(err.Error())
	}

	b.mx.Lock()
	if err == nil {
		b.resources = rows
	}
	b.mx.Unlock()

	b.App().QueueUpdateDraw(func() {
		table := b.Ui()

		if kind == FullUpdate {
			table.Reset()

			table.SetColumns(rp.GetColumns())
			b.RefreshActions()
		}

		b.filterAndSet()
	})
}

func (b *ResourceTableView[R]) filterAndSet() {
	b.mx.RLock()
	defer b.mx.RUnlock()

	var rows []R

	if b.filter != "" {
		rows = make([]R, 0, len(b.resources))
		columns := b.resourceProvider.GetColumns()
		lowerFilter := strings.ToLower(b.filter)

	r:
		for _, row := range b.resources {
			for _, column := range columns {
				value := row.GetTableColumnValue(column.Name)
				if value == "" {
					continue
				}

				if strings.Contains(strings.ToLower(value), lowerFilter) {
					rows = append(rows, row)
					continue r
				}
			}
		}
	} else {
		rows = b.resources
	}

	b.Ui().SetRows(rows)
	b.updateTitle()
}

func (b *ResourceTableView[R]) bindKeys(km ui.KeyMap) {
	if b.enterActionTitle != "" {
		km.Add(tcell.KeyEnter, ui.NewKeyAction(b.enterActionTitle, b.enterCmd))
	}

	if b.resourceProviderWithCheck().CanDeleteResources() {
		km.Add(tcell.KeyCtrlD, ui.NewKeyAction("Delete", b.deleteCmd))
	}
}

func (b *ResourceTableView[R]) enterCmd(*tcell.EventKey) *tcell.EventKey {
	if row, ok := b.GetSelectedResource(); ok {
		b.enterActionFn(row)
	}

	return nil
}

func (b *ResourceTableView[R]) deleteCmd(*tcell.EventKey) *tcell.EventKey {
	row, ok := b.GetSelectedResource()
	if !ok {
		return nil
	}

	b.Stop()
	defer b.Start()

	displayName := row.GetDisplayName()
	msg := fmt.Sprintf("Delete %s?", displayName)

	modal := dialog.CreateConfirmDialog(
		skins.Current(),
		"Confirm Delete",
		msg,
		func() {
			b.App().StatusLine().Infof("Deleting %s...", displayName)

			err := b.resourceProviderWithCheck().DeleteResource(row)
			if err != nil {
				b.App().StatusLine().Errorf("Failed to delete %s: %s", displayName, err.Error())
			} else {
				b.App().StatusLine().Infof("Deleted %s", displayName)
				b.RequestUpdate(PartialUpdate)
			}
		},
		func() {
			b.App().DismissModal()
		})

	b.App().ShowModal(modal)

	return nil
}

func (b *ResourceTableView[R]) openFilterCmd(*tcell.EventKey) *tcell.EventKey {
	b.App().OpenFilter(b)

	return nil
}

func (b *ResourceTableView[R]) updateTitle() {
	count := b.Ui().GetRowCount()
	if count > 0 {
		count--
	}

	var sb strings.Builder

	sb.WriteString(" ")
	sb.WriteString(fmt.Sprintf(titleNameFragmentFmt, b.Name()))

	if b.path != "" {
		sb.WriteString(fmt.Sprintf(titlePathFragmentFmt, b.path))
	}

	sb.WriteString(fmt.Sprintf(titleCountFragmentFmt, count))

	if b.filter != "" {
		sb.WriteString(fmt.Sprintf(titleFilterFragmentFmt, b.filter))
	}

	sb.WriteString(" ")

	b.Ui().SetTitle(SkinTitle(sb.String()))
}

func (b *ResourceTableView[R]) resourceProviderWithCheck() ResourceProvider[R] {
	if b.resourceProvider == nil {
		panic("Resource provider not set")
	}

	return b.resourceProvider
}
