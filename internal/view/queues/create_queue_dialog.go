package queues

import (
	"slices"
	"strings"
	"tbunny/internal/model"
	"tbunny/internal/ui"
	"tbunny/internal/utils"

	rabbithole "github.com/michaelklishin/rabbit-hole/v3"
	"github.com/rivo/tview"
)

type CreateQueueFn func(queueType, vhost, name string, durable bool, args map[string]any)

func ShowCreateQueueDialog(app model.App, okFn CreateQueueFn) {
	f := ui.NewModalForm()

	virtualHostNames := utils.Map(app.Cluster().VirtualHosts(), func(vh rabbithole.VhostInfo) string { return vh.Name })
	activeVhostNameIndex := max(0, slices.Index(virtualHostNames, app.Cluster().ActiveVirtualHost()))

	f.AddInputField("Name:", "", 30, nil, nil)
	f.AddDropDown("Type:", []string{"Default for virtual host", "Classic", "Quorum", "Stream"}, 0, nil)
	f.AddDropDown("Virtual host:", virtualHostNames, activeVhostNameIndex, nil)
	f.AddDropDown("Durability:", []string{"Durable", "Transient"}, 0, nil)

	args := ui.NewArguments().SetLabel("Arguments:")
	f.AddFormItem(args)

	f.AddButtons([]string{"Cancel", "Create"})

	nameField := f.GetFormItem(0).(*tview.InputField)
	queueTypeField := f.GetFormItem(1).(*tview.DropDown)
	vhostField := f.GetFormItem(2).(*tview.DropDown)
	durabilityField := f.GetFormItem(3).(*tview.DropDown)

	f.SetDoneFunc(func(buttonIndex int, buttonLabel string) {
		if buttonIndex != 1 {
			app.DismissModal()
			return
		}

		name := nameField.GetText()
		if name == "" {
			return
		}

		queueTypeIndex, queueType := queueTypeField.GetCurrentOption()
		if queueTypeIndex == 0 {
			queueType = ""
		} else {
			queueType = strings.ToLower(queueType)
		}

		_, vhost := vhostField.GetCurrentOption()
		_, durability := durabilityField.GetCurrentOption()

		okFn(queueType, vhost, name, durability == "Durable", args.GetValue())
	})

	f.SetTitle("Create queue")

	modal := ui.NewModalDialog(f, 80, 10+args.GetFieldHeight())
	app.ShowModal(modal)

	args.SetRowsChangedFunc(func(height int) { modal.Resize(80, 10+height) })
}
