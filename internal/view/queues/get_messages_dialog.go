package queues

import (
	"strconv"
	"tbunny/internal/model"
	"tbunny/internal/rmq"
	"tbunny/internal/ui"

	"github.com/rivo/tview"
)

type GetMessagesFn func(queue *QueueResource, ackMode rmq.AckMode, encoding rmq.RequestedMessageEncoding, count int)

func ShowGetMessagesDialog(app model.App, queue *QueueResource, okFn GetMessagesFn) {
	skin := app.Skin().Dialog

	f := ui.NewModalForm()

	f.AddDropDown("Ack Mode:", []string{"Nack message requeue true", "Automatic ack", "Reject requeue true", "Reject requeue false"}, 0, nil)
	f.AddDropDown("Encoding:", []string{"Auto string / base64", "base64"}, 0, nil)
	f.AddInputField("Count:", "1", 30, tview.InputFieldInteger, nil)

	f.AddButtons([]string{"Cancel", "Get Message(s)"})

	ackModeField := f.GetFormItem(0).(*tview.DropDown)
	encodingField := f.GetFormItem(1).(*tview.DropDown)
	countField := f.GetFormItem(2).(*tview.InputField)

	f.SetDoneFunc(func(buttonIndex int, buttonLabel string) {
		if buttonIndex != 1 {
			app.DismissModal()
			return
		}

		ackModeIndex, _ := ackModeField.GetCurrentOption()
		encodingModeIndex, _ := encodingField.GetCurrentOption()
		countValue := countField.GetText()

		var ackMode rmq.AckMode
		var encoding rmq.RequestedMessageEncoding

		switch ackModeIndex {
		case 0:
			ackMode = rmq.AckModeAckRequeueTrue
		case 1:
			ackMode = rmq.AckModeAckRequeueFalse
		case 2:
			ackMode = rmq.AckModeRejectRequeueTrue
		case 3:
			ackMode = rmq.AckModeRejectRequeueFalse
		}

		switch encodingModeIndex {
		case 0:
			encoding = rmq.RequestedMessageEncodingAuto
		case 1:
			encoding = rmq.RequestedMessageEncodingBase64
		}

		count, err := strconv.Atoi(countValue)
		if err != nil {
			f.SetFocus(2)
			return
		}

		okFn(queue, ackMode, encoding, count)
	})

	f.SetTitle("Get messages").ApplySkin(&skin)

	modal := ui.NewModalDialog(f, 80, 9)
	app.ShowModal(modal)
}
