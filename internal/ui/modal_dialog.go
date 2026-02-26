package ui

import (
	"tbunny/internal/skins"

	"github.com/rivo/tview"
)

type ModalDialog struct {
	*tview.Flex

	centerRow *tview.Flex
	primitive tview.Primitive
}

type Skinnable interface {
	ApplySkin(skin *skins.Skin)
}

func NewModalDialog(primitive tview.Primitive, width, height int) *ModalDialog {
	d := ModalDialog{
		Flex:      tview.NewFlex(),
		centerRow: tview.NewFlex().SetDirection(tview.FlexRow),
		primitive: primitive,
	}

	if skinnable, ok := primitive.(Skinnable); ok {
		skinnable.ApplySkin(skins.Current())
	}

	d.centerRow.
		AddItem(nil, 0, 1, false).
		AddItem(primitive, height, 1, true).
		AddItem(nil, 0, 1, false)

	d.AddItem(nil, 0, 1, false).
		AddItem(d.centerRow, width, 1, true).
		AddItem(nil, 0, 1, false)

	return &d
}

func (d *ModalDialog) Resize(width, height int) {
	d.centerRow.ResizeItem(d.primitive, height, 1)
	d.ResizeItem(d.centerRow, width, 1)
}
