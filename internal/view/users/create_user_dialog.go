package users

import (
	"strings"
	"tbunny/internal/model"
	"tbunny/internal/ui"

	"github.com/rivo/tview"
)

type CreateUserFn func(name, password, tags string)

func ShowCreateUserDialog(app model.App, okFn CreateUserFn) {
	f := ui.NewModalForm()

	f.AddInputField("User name:", "", 30, nil, nil)
	f.AddPasswordField("Password:", "", 30, '*', nil)
	f.AddPasswordField("Confirm password:", "", 30, '*', nil)
	f.AddInputField("Tags:", "", 30, nil, nil)

	f.AddButtons([]string{"Cancel", "Create"})

	nameField := f.GetFormItem(0).(*tview.InputField)
	passwordField := f.GetFormItem(1).(*tview.InputField)
	confirmPasswordField := f.GetFormItem(2).(*tview.InputField)
	tagsField := f.GetFormItem(3).(*tview.InputField)

	passwordField.SetPlaceholder("Leave empty if users cannot login")
	confirmPasswordField.SetPlaceholder("Leave empty if users cannot login")
	tagsField.SetPlaceholder("Comma-separated list of tags")

	f.SetDoneFunc(func(buttonIndex int, buttonLabel string) {
		if buttonIndex != 1 {
			app.DismissModal()
			return
		}

		name := strings.TrimSpace(nameField.GetText())
		if name == "" {
			f.SetFocus(0)
			return
		}

		password := passwordField.GetText()
		confirmPassword := confirmPasswordField.GetText()
		tags := tagsField.GetText()

		if password != confirmPassword {
			f.SetFocus(1)
			return
		}

		okFn(name, password, strings.TrimSpace(tags))
	})

	f.SetTitle("Create users")

	modal := ui.NewModalDialog(f, 60, 10)
	app.ShowModal(modal)
}
