package ui

import (
	"tbunny/internal/skins"

	"github.com/gdamore/tcell/v2"
	"github.com/rivo/tview"
)

type ModalForm struct {
	*tview.Form

	done func(buttonIndex int, buttonLabel string)
}

func NewModalForm() *ModalForm {
	f := ModalForm{Form: tview.NewForm()}

	f.SetItemPadding(0)
	f.SetTitleAlign(tview.AlignCenter)
	f.SetBorder(true)
	f.SetButtonsAlign(tview.AlignCenter)

	f.SetCancelFunc(func() {
		if f.done != nil {
			f.done(-1, "")
		}
	})

	return &f
}

// ApplySkin applies the provided skin to the form.
func (f *ModalForm) ApplySkin(skin *skins.Dialog) *ModalForm {
	f.SetButtonBackgroundColor(skin.ButtonBgColor.Color()).
		SetButtonTextColor(skin.ButtonFgColor.Color()).
		SetLabelColor(skin.LabelFgColor.Color()).
		SetFieldBackgroundColor(skin.BgColor.Color()).
		SetFieldTextColor(skin.FieldFgColor.Color())

	for i := 0; i < f.GetFormItemCount(); i++ {
		item := f.GetFormItem(i)

		if dropDown, ok := item.(*tview.DropDown); ok {
			dropDown.SetListStyles(
				tcell.StyleDefault.Foreground(skin.DropdownFgColor.Color()).Background(skin.DropdownBgColor.Color()),
				tcell.StyleDefault.Foreground(skin.DropdownFocusFgColor.Color()).Background(skin.DropdownFocusBgColor.Color()))
		}

		if inputField, ok := item.(*tview.InputField); ok {
			inputField.SetAutocompleteStyles(
				skin.DropdownBgColor.Color(),
				tcell.StyleDefault.Foreground(skin.DropdownFgColor.Color()).Background(skin.DropdownBgColor.Color()),
				tcell.StyleDefault.Foreground(skin.DropdownFocusFgColor.Color()).Background(skin.DropdownFocusBgColor.Color()))
		}

		if arguments, ok := item.(*Arguments); ok {
			arguments.SetListStyles(
				tcell.StyleDefault.Foreground(skin.DropdownFgColor.Color()).Background(skin.DropdownBgColor.Color()),
				tcell.StyleDefault.Foreground(skin.DropdownFocusFgColor.Color()).Background(skin.DropdownFocusBgColor.Color()))
		}

		if properties, ok := item.(*Properties); ok {
			properties.SetListStyles(
				tcell.StyleDefault.Foreground(skin.DropdownFgColor.Color()).Background(skin.DropdownBgColor.Color()),
				tcell.StyleDefault.Foreground(skin.DropdownFocusFgColor.Color()).Background(skin.DropdownFocusBgColor.Color()))
		}
	}

	return f
}

// SetDoneFunc sets a handler which is called when one of the buttons was
// pressed. It receives the index of the button as well as its label text. The
// handler is also called when the users presses the Escape key. The index will
// then be negative and the label text an empty string.
func (f *ModalForm) SetDoneFunc(handler func(buttonIndex int, buttonLabel string)) *ModalForm {
	f.done = handler

	return f
}

// SetTitle sets the title of the window.
func (f *ModalForm) SetTitle(title string) *ModalForm {
	f.Form.SetTitle(" " + title + " ")

	return f
}

// AddButtons adds buttons to the window. There must be at least one button and
// a "done" handler so the window can be closed again.
func (f *ModalForm) AddButtons(labels []string) *ModalForm {
	for index, label := range labels {
		func(i int, l string) {
			f.AddButton(label, func() {
				if f.done != nil {
					f.done(i, l)
				}
			})

			button := f.GetButton(f.GetButtonCount() - 1)
			button.SetInputCapture(func(event *tcell.EventKey) *tcell.EventKey {
				switch event.Key() {
				case tcell.KeyDown, tcell.KeyRight:
					return tcell.NewEventKey(tcell.KeyTab, 0, tcell.ModNone)
				case tcell.KeyUp, tcell.KeyLeft:
					return tcell.NewEventKey(tcell.KeyBacktab, 0, tcell.ModNone)
				default:
					break
				}
				return event
			})
		}(index, label)
	}

	return f
}

// ClearButtons removes all buttons from the window.
func (f *ModalForm) ClearButtons() *ModalForm {
	f.ClearButtons()

	return f
}
