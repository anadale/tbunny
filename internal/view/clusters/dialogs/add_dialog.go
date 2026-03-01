package dialogs

import (
	"strings"
	"tbunny/internal/cluster"
	"tbunny/internal/model"
	"tbunny/internal/ui"

	"github.com/rivo/tview"
)

type AddClusterFn func(name string, parameters cluster.ConnectionParameters)

const (
	directTypeOption = "Direct"
	k8sTypeOption    = "Kubernetes"
	usernameLabel    = "Username:"
	passwordLabel    = "Password:"
)

func ShowAddClusterDialog(app model.App, okFn AddClusterFn) {
	f := ui.NewModalForm()

	f.AddInputField("Name:", "", 30, nil, nil)
	f.AddDropDown("Type:", getAllowedTypes(), 0, nil)

	f.AddButtons([]string{"Cancel", "Create"})

	nameField := f.GetFormItem(0).(*tview.InputField)
	typeField := f.GetFormItem(1).(*tview.DropDown)

	nameField.SetPlaceholder("localhost")

	typeFieldsCount := createDirectConnectionFields(f)
	createUsernameAndPasswordFields(f)

	f.SetDoneFunc(func(buttonIndex int, buttonLabel string) {
		if buttonIndex != 1 {
			app.DismissModal()
			return
		}

		name := strings.TrimSpace(nameField.GetText())
		if !validateClusterName(name) {
			f.SetFocus(0)
			return
		}

		usernameField := f.GetFormItemByLabel(usernameLabel).(*tview.InputField)
		passwordField := f.GetFormItemByLabel(passwordLabel).(*tview.InputField)

		username := strings.TrimSpace(usernameField.GetText())
		if username == "" {
			f.SetFocus(f.GetFormItemIndex(usernameLabel))
		}

		password := passwordField.GetText()
		if password == "" {
			password = "guest"
		}

		var params cluster.ConnectionParameters

		_, t := typeField.GetCurrentOption()

		if t == directTypeOption {
			directParams, ok := collectDirectConnectionParameters(f)
			if !ok {
				return
			}

			params = cluster.ConnectionParameters{
				Direct:   directParams,
				Username: username,
				Password: password,
			}
		} else if t == k8sTypeOption {
			k8sParams, ok := collectKubernetesConnectionParameters(f)
			if !ok {
				return
			}

			params = cluster.ConnectionParameters{
				K8s:      k8sParams,
				Username: username,
				Password: password,
			}
		}

		okFn(name, params)
	})

	f.SetTitle("Add cluster")

	const formWidth = 60
	const baseFormHeight = 10

	modal := ui.NewModalDialog(f, formWidth, baseFormHeight+typeFieldsCount)

	typeField.SetSelectedFunc(func(text string, index int) {
		for f.GetFormItemCount() > 2 {
			f.RemoveFormItem(2)
		}

		var typeFieldsCount int

		if text == directTypeOption {
			typeFieldsCount = createDirectConnectionFields(f)
		} else if text == k8sTypeOption {
			typeFieldsCount = createKubernetesConnectionFields(f)
		}

		createUsernameAndPasswordFields(f)

		modal.ApplySkin()
		modal.Resize(formWidth, baseFormHeight+typeFieldsCount)
	})

	app.ShowModal(modal)
}

func getAllowedTypes() []string {
	t := make([]string, 0, 2)

	t = append(t, directTypeOption)

	if kubernetesIsAvailable() {
		t = append(t, k8sTypeOption)
	}

	return t
}

func createUsernameAndPasswordFields(f *ui.ModalForm) {
	f.AddInputField(usernameLabel, "", 30, nil, nil)
	f.AddInputField(passwordLabel, "", 30, nil, nil)

	usernameField := f.GetFormItemByLabel(usernameLabel).(*tview.InputField)
	passwordField := f.GetFormItemByLabel(passwordLabel).(*tview.InputField)

	usernameField.SetPlaceholder("guest")
	passwordField.SetPlaceholder("guest")
}
