package users

import (
	"tbunny/internal/model"
	"tbunny/internal/ui"

	"github.com/rivo/tview"
)

type VhostPermissionsFn func(vhost, configure, write, read string)

func ShowCreateVhostPermissionsDialog(app model.App, availableVhosts []string, okFn VhostPermissionsFn) {
	skin := app.Skin().Dialog
	f := ui.NewModalForm()

	f.AddDropDown("Virtual host:", availableVhosts, 0, nil)
	f.AddInputField("Configure regexp:", ".*", 30, nil, nil)
	f.AddInputField("Write regexp:", ".*", 30, nil, nil)
	f.AddInputField("Read regexp:", ".*", 30, nil, nil)

	f.AddButtons([]string{"Cancel", "Create"})

	vhostField := f.GetFormItem(0).(*tview.DropDown)
	configureField := f.GetFormItem(1).(*tview.InputField)
	writeField := f.GetFormItem(2).(*tview.InputField)
	readField := f.GetFormItem(3).(*tview.InputField)

	f.SetDoneFunc(func(buttonIndex int, buttonLabel string) {
		if buttonIndex != 1 {
			app.DismissModal()
			return
		}

		_, vhost := vhostField.GetCurrentOption()
		configure := configureField.GetText()
		write := writeField.GetText()
		read := readField.GetText()

		okFn(vhost, configure, write, read)
	})

	f.SetTitle("Create permissions").ApplySkin(&skin)

	modal := ui.NewModalDialog(f, 60, 10)
	app.ShowModal(modal)
}

func ShowEditVhostPermissionsDialog(app model.App, vhost, configure, write, read string, okFn VhostPermissionsFn) {
	skin := app.Skin().Dialog
	f := ui.NewModalForm()

	f.AddInputField("Virtual host:", vhost, 30, nil, nil)
	f.AddInputField("Configure regexp:", configure, 30, nil, nil)
	f.AddInputField("Write regexp:", write, 30, nil, nil)
	f.AddInputField("Read regexp:", read, 30, nil, nil)

	f.AddButtons([]string{"Cancel", "Update"})

	vhostField := f.GetFormItem(0).(*tview.InputField)
	configureField := f.GetFormItem(1).(*tview.InputField)
	writeField := f.GetFormItem(2).(*tview.InputField)
	readField := f.GetFormItem(3).(*tview.InputField)

	vhostField.SetDisabled(true)

	f.SetDoneFunc(func(buttonIndex int, buttonLabel string) {
		if buttonIndex != 1 {
			app.DismissModal()
			return
		}

		configure = configureField.GetText()
		write = writeField.GetText()
		read = readField.GetText()

		okFn(vhost, configure, write, read)
	})

	f.SetTitle("Edit permissions").ApplySkin(&skin)

	modal := ui.NewModalDialog(f, 60, 10)
	app.ShowModal(modal)
}
