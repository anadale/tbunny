package dialogs

import (
	"strings"
	"tbunny/internal/cluster"
	"tbunny/internal/ui"

	"github.com/rivo/tview"
)

const directConnectionUriFieldLabel = "URI:"

func createDirectConnectionFields(f *ui.ModalForm) int {
	uriField := tview.NewInputField().
		SetLabel(directConnectionUriFieldLabel).
		SetFieldWidth(30).
		SetPlaceholder("http://localhost:15672")

	f.AddFormItem(uriField)

	return 1
}

func collectDirectConnectionParameters(f *ui.ModalForm) (*cluster.DirectConnectionParameters, bool) {
	uriField := f.GetFormItemByLabel(directConnectionUriFieldLabel).(*tview.InputField)

	uri := strings.TrimSpace(uriField.GetText())
	if !validateUri(uri) {
		f.SetFocus(f.GetFormItemIndex(directConnectionUriFieldLabel))
		return nil, false
	}

	return &cluster.DirectConnectionParameters{
		Uri: uri,
	}, true
}
