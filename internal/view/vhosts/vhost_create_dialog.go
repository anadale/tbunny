package vhosts

import (
	"tbunny/internal/model"
	"tbunny/internal/ui"

	"github.com/rivo/tview"
)

// CreateVHostFn is a callback function for creating a virtual host.
type CreateVHostFn func(name, description, tags, queueType string, tracing bool)

func ShowCreateVHostDialog(app model.App, okFn CreateVHostFn) {
	f := ui.NewModalForm()

	f.AddInputField("Virtual host name:", "", 30, nil, nil)
	f.AddInputField("Description:", "", 30, nil, nil)
	f.AddInputField("Tags:", "", 30, nil, nil)
	f.AddDropDown("Default Queue Type:", []string{"Classic", "Quorum", "Stream"}, 0, nil)
	f.AddCheckbox("Enable Tracing:", false, nil)

	nameField := f.GetFormItem(0).(*tview.InputField)
	nameField.SetPlaceholder("Enter virtual host name")

	f.AddButtons([]string{"Cancel", "Create"})
	f.SetDoneFunc(func(buttonIndex int, buttonLabel string) {
		if buttonIndex != 1 {
			app.DismissModal()
			return
		}

		name := nameField.GetText()
		if name == "" {
			return
		}

		description := f.GetFormItem(1).(*tview.InputField).GetText()
		tags := f.GetFormItem(2).(*tview.InputField).GetText()
		_, queueType := f.GetFormItem(3).(*tview.DropDown).GetCurrentOption()
		tracing := f.GetFormItem(4).(*tview.Checkbox).IsChecked()

		okFn(name, description, tags, queueType, tracing)
	})

	f.SetTitle("Create virtual host")

	modal := ui.NewModalDialog(f, 60, 11)
	app.ShowModal(modal)
}
