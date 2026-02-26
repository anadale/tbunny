package users

import (
	"tbunny/internal/model"
	"tbunny/internal/ui"

	"github.com/rivo/tview"
)

type TopicPermissionsFn func(vhost, exchange, write, read string)

func ShowCreateTopicPermissionsDialog(app model.App, vhosts []string, fetchExchanges func(vhost string) []string, okFn TopicPermissionsFn) {
	f := ui.NewModalForm()

	f.AddDropDown("Virtual host:", vhosts, 0, nil)
	f.AddDropDown("Exchange:", nil, 0, nil)
	f.AddInputField("Write regexp:", ".*", 30, nil, nil)
	f.AddInputField("Read regexp:", ".*", 30, nil, nil)

	f.AddButtons([]string{"Cancel", "Create"})

	vhostField := f.GetFormItem(0).(*tview.DropDown)
	exchangeField := f.GetFormItem(1).(*tview.DropDown)
	writeField := f.GetFormItem(2).(*tview.InputField)
	readField := f.GetFormItem(3).(*tview.InputField)

	fillExchanges := func(text string, index int) {
		exchanges := fetchExchanges(text)
		exchangeField.SetOptions(exchanges, nil)
		exchangeField.SetCurrentOption(0)
	}

	vhostField.SetSelectedFunc(fillExchanges)
	fillExchanges(vhosts[0], 0)

	f.SetDoneFunc(func(buttonIndex int, buttonLabel string) {
		if buttonIndex != 1 {
			app.DismissModal()
			return
		}

		_, vhost := vhostField.GetCurrentOption()
		_, exchange := exchangeField.GetCurrentOption()
		write := writeField.GetText()
		read := readField.GetText()

		okFn(vhost, exchange, write, read)
	})

	f.SetTitle("Create topic permissions")

	modal := ui.NewModalDialog(f, 60, 10)
	app.ShowModal(modal)
}

func ShowEditTopicPermissionsDialog(app model.App, vhost, exchange, write, read string, okFn TopicPermissionsFn) {
	f := ui.NewModalForm()

	f.AddInputField("Virtual host:", vhost, 30, nil, nil)
	f.AddInputField("Exchange:", exchange, 30, nil, nil)
	f.AddInputField("Write regexp:", write, 30, nil, nil)
	f.AddInputField("Read regexp:", read, 30, nil, nil)

	f.AddButtons([]string{"Cancel", "Update"})

	vhostField := f.GetFormItem(0).(*tview.InputField)
	exchangeField := f.GetFormItem(1).(*tview.InputField)
	writeField := f.GetFormItem(2).(*tview.InputField)
	readField := f.GetFormItem(3).(*tview.InputField)

	vhostField.SetDisabled(true)
	exchangeField.SetDisabled(true)

	f.SetDoneFunc(func(buttonIndex int, buttonLabel string) {
		if buttonIndex != 1 {
			app.DismissModal()
			return
		}

		write = writeField.GetText()
		read = readField.GetText()

		okFn(vhost, exchange, write, read)
	})

	f.SetTitle("Edit topic permissions")

	modal := ui.NewModalDialog(f, 60, 10)
	app.ShowModal(modal)
}
