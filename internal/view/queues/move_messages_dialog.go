package queues

import (
	"fmt"
	"strings"
	"tbunny/internal/model"
	"tbunny/internal/ui"

	"github.com/rivo/tview"
)

type MoveMessagesFn func(vhost, sourceQueue, destinationQueue string)

func ShowMoveMessagesDialog(app model.App, vhost, sourceQueue string, destinationQueues []string, okFn MoveMessagesFn) {
	skin := app.Skin().Dialog
	f := ui.NewModalForm()

	f.AddInputField("Destination queue:", "", 30, nil, nil)
	f.AddButtons([]string{"Cancel", "Move"})

	destinationField := f.GetFormItem(0).(*tview.InputField)
	destinationField.SetAutocompleteFunc(func(text string) (items []string) {
		if len(text) == 0 {
			return destinationQueues
		}

		for _, queue := range destinationQueues {
			if strings.HasPrefix(queue, text) {
				items = append(items, queue)
			}
		}

		return
	})

	f.SetDoneFunc(func(buttonIndex int, buttonLabel string) {
		if buttonIndex != 1 {
			app.DismissModal()
			return
		}

		destinationQueue := destinationField.GetText()
		if destinationQueue == "" {
			return
		}

		okFn(vhost, sourceQueue, destinationQueue)
	})

	f.SetTitle(fmt.Sprintf("Move messages from %s", sourceQueue)).ApplySkin(&skin)

	modal := ui.NewModalDialog(f, 60, 7)
	app.ShowModal(modal)
}
