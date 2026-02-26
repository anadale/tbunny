package dialog

import (
	"tbunny/internal/skins"

	"github.com/gdamore/tcell/v2"
	"github.com/rivo/tview"
)

func CreateConfirmDialog(skin *skins.Skin, title, msg string, confirmFn func(), closeFn func()) tview.Primitive {
	dialogSkin := skin.Dialog

	abs := tcell.StyleDefault.
		Foreground(dialogSkin.ButtonFocusFgColor.Color()).
		Background(dialogSkin.ButtonFocusBgColor.Color())

	modal := tview.NewModal().
		SetButtonBackgroundColor(dialogSkin.ButtonBgColor.Color()).
		SetButtonTextColor(dialogSkin.ButtonFgColor.Color()).
		SetBackgroundColor(dialogSkin.BgColor.Color()).
		SetTextColor(dialogSkin.FgColor.Color()).
		SetButtonActivatedStyle(abs)

	modal.SetText(msg).SetTitle(" " + title + " ")
	modal.AddButtons([]string{"Cancel", "OK"})

	modal.SetDoneFunc(func(buttonIndex int, buttonLabel string) {
		if buttonIndex == 1 {
			confirmFn()
		}

		closeFn()
	})

	return modal
}
