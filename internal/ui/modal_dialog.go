package ui

import "github.com/rivo/tview"

type ModalDialog struct {
	*tview.Flex

	centerRow *tview.Flex
	primitive tview.Primitive
}

func NewModalDialog(primitive tview.Primitive, width, height int) *ModalDialog {
	d := ModalDialog{
		Flex:      tview.NewFlex(),
		centerRow: tview.NewFlex().SetDirection(tview.FlexRow),
		primitive: primitive,
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
