package users

import (
	"strings"
	"tbunny/internal/model"
	"tbunny/internal/ui"

	"github.com/rivo/tview"
)

type EditUserFn func(changePassword bool, password, tags string)

func ShowEditUserDialog(app model.App, name, tags string, okFn EditUserFn) {
	f := ui.NewModalForm()

	f.AddCheckbox("Change password:", false, nil)
	f.AddPasswordField("Password:", "", 30, '*', nil)
	f.AddPasswordField("Confirm password:", "", 30, '*', nil)
	f.AddInputField("Tags:", tags, 30, nil, nil)

	f.AddButtons([]string{"Cancel", "Update"})

	changePasswordField := f.GetFormItem(0).(*tview.Checkbox)
	passwordField := f.GetFormItem(1).(*tview.InputField)
	confirmPasswordField := f.GetFormItem(2).(*tview.InputField)
	tagsField := f.GetFormItem(3).(*tview.InputField)

	const leaveEmpty = "Leave empty to disable login"
	const inputIgnored = "Input is ignored"

	changePasswordField.SetChangedFunc(func(checked bool) {
		if checked {
			passwordField.SetPlaceholder(leaveEmpty)
			confirmPasswordField.SetPlaceholder(leaveEmpty)
		} else {
			passwordField.SetPlaceholder(inputIgnored)
			confirmPasswordField.SetPlaceholder(inputIgnored)
		}
	})

	passwordField.SetPlaceholder(inputIgnored)
	confirmPasswordField.SetPlaceholder(inputIgnored)
	tagsField.SetPlaceholder("Comma-separated list of tags")

	f.SetDoneFunc(func(buttonIndex int, buttonLabel string) {
		if buttonIndex != 1 {
			app.DismissModal()
			return
		}

		changePassword := changePasswordField.IsChecked()
		password := passwordField.GetText()
		confirmPassword := confirmPasswordField.GetText()
		tags := tagsField.GetText()

		if changePassword && password != confirmPassword {
			f.SetFocus(1)
			return
		}

		okFn(changePassword, password, strings.TrimSpace(tags))
	})

	f.SetTitle("Edit users " + name)

	modal := ui.NewModalDialog(f, 60, 10)
	app.ShowModal(modal)
}
