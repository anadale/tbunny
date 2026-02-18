package dialog

import (
	"tbunny/internal/skins"

	"github.com/gdamore/tcell/v2"
	"github.com/rivo/tview"
)

func CreateConfirmDialog(skin *skins.Dialog, title, msg string, confirmFn func(), closeFn func()) tview.Primitive {
	abs := tcell.StyleDefault.
		Foreground(skin.ButtonFocusFgColor.Color()).
		Background(skin.ButtonFocusBgColor.Color())

	modal := tview.NewModal().
		SetButtonBackgroundColor(skin.ButtonBgColor.Color()).
		SetButtonTextColor(skin.ButtonFgColor.Color()).
		SetBackgroundColor(skin.BgColor.Color()).
		SetTextColor(skin.FgColor.Color()).
		SetButtonActivatedStyle(abs)

	modal.SetText(msg).SetTitle("<" + title + ">")
	modal.AddButtons([]string{"Cancel", "OK"})

	modal.SetDoneFunc(func(buttonIndex int, buttonLabel string) {
		if buttonIndex == 1 {
			confirmFn()
		}

		closeFn()
	})

	return modal
}
