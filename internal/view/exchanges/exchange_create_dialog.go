package exchanges

import (
	"slices"
	"strings"
	"tbunny/internal/cluster"
	"tbunny/internal/model"
	"tbunny/internal/ui"
	"tbunny/internal/utils"

	rabbithole "github.com/michaelklishin/rabbit-hole/v3"
	"github.com/rivo/tview"
)

type CreateExchangeFn func(exchangeType, vhost, name string, durable, autoDelete bool, args map[string]any)

func ShowCreateExchangeDialog(app model.App, okFn CreateExchangeFn) {
	f := ui.NewModalForm()
	c := cluster.Current()

	virtualHostNames := utils.Map(c.VirtualHosts(), func(vh rabbithole.VhostInfo) string { return vh.Name })
	activeVhostNameIndex := max(0, slices.Index(virtualHostNames, c.ActiveVirtualHost()))

	f.AddInputField("Name:", "", 30, nil, nil)
	f.AddDropDown("Type:", []string{"Direct", "Fanout", "Headers", "Topic", "x-local-random"}, 0, nil)
	f.AddDropDown("Virtual host:", virtualHostNames, activeVhostNameIndex, nil)
	f.AddDropDown("Durability:", []string{"Durable", "Transient"}, 0, nil)
	f.AddCheckbox("Auto-delete:", false, nil)

	argsField := ui.NewArguments().SetLabel("Arguments:")
	f.AddFormItem(argsField)

	f.AddButtons([]string{"Cancel", "Create"})

	nameField := f.GetFormItem(0).(*tview.InputField)
	exchangeTypeField := f.GetFormItem(1).(*tview.DropDown)
	vhostField := f.GetFormItem(2).(*tview.DropDown)
	durabilityField := f.GetFormItem(3).(*tview.DropDown)
	autoDeleteField := f.GetFormItem(4).(*tview.Checkbox)

	f.SetDoneFunc(func(buttonIndex int, buttonLabel string) {
		if buttonIndex != 1 {
			app.DismissModal()
			return
		}

		name := nameField.GetText()
		if name == "" {
			return
		}

		_, exchangeType := exchangeTypeField.GetCurrentOption()
		_, vhost := vhostField.GetCurrentOption()
		_, durability := durabilityField.GetCurrentOption()
		autoDelete := autoDeleteField.IsChecked()

		okFn(strings.ToLower(exchangeType), vhost, name, durability == "Durable", autoDelete, argsField.GetValue())
	})

	f.SetTitle("Create exchange")

	modal := ui.NewModalDialog(f, 80, 11+argsField.GetFieldHeight())
	app.ShowModal(modal)

	argsField.SetRowsChangedFunc(func(height int) { modal.Resize(80, 11+height) })
}
