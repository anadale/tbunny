package bindings

import (
	"strings"
	"tbunny/internal/model"
	"tbunny/internal/ui"
	"tbunny/internal/utils"

	"github.com/rivo/tview"
)

type CreateBindingFn func(otherType SubjectType, otherName, routingKey string, args map[string]any)
type AutocompleteFn func() ([]string, error)

func ShowCreateBindingDialog(app model.App, subjectType SubjectType, exchangesFn, queuesFn AutocompleteFn, okFn CreateBindingFn) {
	skin := app.Skin().Dialog

	f := ui.NewModalForm()

	var exchanges []string
	var queues []string

	fetchExchanges := func() []string {
		if exchanges != nil {
			return exchanges
		}

		var err error
		exchanges, err = exchangesFn()
		if err != nil {
			return nil
		}

		return exchanges
	}

	fetchQueues := func() []string {
		if queues != nil {
			return queues
		}

		var err error
		queues, err = queuesFn()
		if err != nil {
			return nil
		}

		return queues
	}

	makeAutocomplete := func(fetchFunc func() []string) func(string) []string {
		return func(text string) []string {
			items := fetchFunc()
			return utils.Filter(items, func(item string) bool { return strings.HasPrefix(strings.ToLower(item), strings.ToLower(text)) })
		}
	}
	makeAutocompleted := func(inputField *tview.InputField) func(string, int, int) bool {
		return func(text string, index, source int) bool {
			if source != tview.AutocompletedNavigate {
				inputField.SetText(text)
			}

			return source == tview.AutocompletedEnter || source == tview.AutocompletedClick
		}
	}

	var otherType SubjectType
	var otherNameField *tview.InputField

	fieldsCount := 2

	if subjectType == ExchangeSubject {
		f.AddDropDown("To:", []string{"Queue", "Exchange"}, 0, nil)
		f.AddInputField("Queue name:", "", 30, nil, nil)

		otherTypeField := f.GetFormItem(0).(*tview.DropDown)
		otherNameField = f.GetFormItem(1).(*tview.InputField)

		otherNameField.SetAutocompletedFunc(makeAutocompleted(otherNameField))
		otherNameField.SetAutocompleteFunc(makeAutocomplete(fetchQueues))

		otherTypeField.SetSelectedFunc(func(text string, index int) {
			if index == 0 {
				otherType = QueueSubject
				otherNameField.SetLabel("Queue name:")
				otherNameField.SetText("")
				otherNameField.SetAutocompleteFunc(makeAutocomplete(fetchQueues))
			} else {
				otherType = ExchangeSubject
				otherNameField.SetLabel("Exchange name:")
				otherNameField.SetText("")
				otherNameField.SetAutocompleteFunc(makeAutocomplete(fetchExchanges))
			}
		})

		fieldsCount++
	} else {
		f.AddInputField("From exchange:", "", 30, nil, nil)

		otherType = ExchangeSubject
		otherNameField = f.GetFormItem(0).(*tview.InputField)

		otherNameField.SetAutocompletedFunc(makeAutocompleted(otherNameField))
		otherNameField.SetAutocompleteFunc(makeAutocomplete(fetchExchanges))
	}

	f.AddInputField("Routing key:", "", 30, nil, nil)

	routingKeyField := f.GetFormItemByLabel("Routing key:").(*tview.InputField)
	argsField := ui.NewArguments().SetLabel("Arguments:")
	f.AddFormItem(argsField)

	f.AddButtons([]string{"Cancel", "Create"})

	f.SetDoneFunc(func(buttonIndex int, buttonLabel string) {
		if buttonIndex != 1 {
			app.DismissModal()
			return
		}

		otherName := otherNameField.GetText()
		if otherName == "" {
			return
		}

		routingKey := routingKeyField.GetText()

		okFn(otherType, otherName, routingKey, argsField.GetValue())
	})

	f.SetTitle("Create binding").ApplySkin(&skin)

	modal := ui.NewModalDialog(f, 80, 6+fieldsCount+argsField.GetFieldHeight())
	app.ShowModal(modal)

	argsField.SetRowsChangedFunc(func(height int) { modal.Resize(80, 6+fieldsCount+height) })
}
