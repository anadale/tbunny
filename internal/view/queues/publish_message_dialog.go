package queues

import (
	"tbunny/internal/model"
	"tbunny/internal/rmq"
	"tbunny/internal/ui"

	"github.com/rivo/tview"
)

type PublishMessageFn func(vhost, queue string, props map[string]any, payload string, payloadEncoding rmq.PayloadEncoding)

func ShowPublishMessageDialog(app model.App, vhost, queue string, okFn PublishMessageFn) {
	f := ui.NewModalForm()

	f.AddDropDown("Delivery mode:", []string{rmq.MessageDeliveryModeNonPersistent.String(), rmq.MessageDeliveryModePersistent.String()}, 0, nil)
	headersField := ui.NewArguments().SetLabel("Headers:").SetKeyPlaceholder("Header name").SetValuePlaceholder("Header value")
	f.AddFormItem(headersField)
	propertiesField := ui.NewProperties().SetLabel("Properties:")
	f.AddFormItem(propertiesField)
	f.AddTextArea("Payload:", "", 59, 8, 0, nil)
	f.AddDropDown("Payload encoding:", []string{string(rmq.PayloadEncodingString), string(rmq.PayloadEncodingBase64)}, 0, nil)

	f.AddButtons([]string{"Cancel", "Publish"})

	deliveryModeField := f.GetFormItem(0).(*tview.DropDown)
	payloadField := f.GetFormItem(3).(*tview.TextArea)
	payloadEncodingField := f.GetFormItem(4).(*tview.DropDown)

	payloadField.SetPlaceholder("Message payload")

	f.SetDoneFunc(func(buttonIndex int, buttonLabel string) {
		if buttonIndex != 1 {
			app.DismissModal()
			return
		}

		_, deliveryModeText := deliveryModeField.GetCurrentOption()
		_, payloadEncoding := payloadEncodingField.GetCurrentOption()

		deliveryMode, err := rmq.ParseDeliveryMode(deliveryModeText)
		if err != nil {
			f.SetFocus(0)
			return
		}

		headers := headersField.GetValue()
		props := propertiesField.GetValue()
		props["delivery_mode"] = deliveryMode

		if len(headers) > 0 {
			props["headers"] = headers
		}

		okFn(vhost, queue, props, payloadField.GetText(), rmq.PayloadEncoding(payloadEncoding))
	})

	f.SetTitle("Publish message")

	const modalHeight = 17

	headersHeight := headersField.GetFieldHeight()
	propertiesHeight := propertiesField.GetFieldHeight()

	modal := ui.NewModalDialog(f, 80, modalHeight+headersHeight+propertiesHeight)
	app.ShowModal(modal)

	resize := func() {
		modal.Resize(80, modalHeight+headersHeight+propertiesHeight)
	}

	headersField.SetRowsChangedFunc(func(height int) { headersHeight = height; resize() })
	propertiesField.SetRowsChangedFunc(func(height int) { propertiesHeight = height; resize() })
}
