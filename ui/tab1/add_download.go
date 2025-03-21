package tab1

import (
	"log/slog"

	"github.com/gdamore/tcell/v2"
	"github.com/rivo/tview"
)

var staticQueues = []string{"Default", "Queue1", "Queue2"}

type Tab1 struct {
	Form *tview.Form
}

func NewTab1() *Tab1 {
	form := tview.NewForm().
		AddInputField("URL", "", 40, nil, nil).
		AddDropDown("Queue", staticQueues, 0, nil).
		AddInputField("Output File", "", 40, nil, nil)

	form.
		AddButton("OK", okButtonHandler(form)).
		AddButton("Cancel", cancelButtonHandler(form))
	form.SetBorder(true).SetTitle("Add Download").SetTitleAlign(tview.AlignLeft)

	setupDropdown(form)
	setupURLFieldInputCapture(form)
	setupOutputFieldInputCapture(form)
	setupButtonInputCaptures(form)

	return &Tab1{Form: form}
}

func okButtonHandler(form *tview.Form) func() {
	return func() {
		url := form.GetFormItemByLabel("URL").(*tview.InputField).GetText()
		outputFile := form.GetFormItemByLabel("Output File").(*tview.InputField).GetText()

		if url == "" {
			form.GetFormItemByLabel("URL").(*tview.InputField).SetText("This field is required")
			form.SetFocus(0)
		} else if outputFile == "" {
			form.GetFormItemByLabel("Output File").(*tview.InputField).SetText("This field is required")
			form.SetFocus(2)
		} else {
			slog.Info("FK")
			url := form.GetFormItemByLabel("URL").(*tview.InputField).GetText()
			queueIndex, _ := form.GetFormItemByLabel("Queue").(*tview.DropDown).GetCurrentOption()
			queueName := staticQueues[queueIndex]
			outputFile := form.GetFormItemByLabel("Output File").(*tview.InputField).GetText()
			slog.Info("Submitted Data",
				"url", url,
				"queue", queueName,
				"output_file", outputFile,
			)

			form.GetFormItemByLabel("URL").(*tview.InputField).SetText("")
			form.GetFormItemByLabel("Queue").(*tview.DropDown).SetCurrentOption(0)
			form.GetFormItemByLabel("Output File").(*tview.InputField).SetText("")
			form.SetFocus(0)
		}
	}
}

func cancelButtonHandler(form *tview.Form) func() {
	return func() {
		form.GetFormItemByLabel("URL").(*tview.InputField).SetText("")
		form.GetFormItemByLabel("Queue").(*tview.DropDown).SetCurrentOption(0)
		form.GetFormItemByLabel("Output File").(*tview.InputField).SetText("")
		form.SetFocus(0)
	}
}

func setupDropdown(form *tview.Form) {
	dropdown := form.GetFormItemByLabel("Queue").(*tview.DropDown)
	dropdown.SetSelectedFunc(func(text string, index int) {
		form.SetFocus(2)
	})
}

func setupURLFieldInputCapture(form *tview.Form) {
	urlField := form.GetFormItemByLabel("URL").(*tview.InputField)
	urlField.SetInputCapture(func(event *tcell.EventKey) *tcell.EventKey {
		switch event.Key() {
		case tcell.KeyEnter:
			form.SetFocus(1)
			return nil
		case tcell.KeyBacktab:
			return nil
		}
		return event
	})
}

func setupOutputFieldInputCapture(form *tview.Form) {
	outputField := form.GetFormItemByLabel("Output File").(*tview.InputField)
	outputField.SetInputCapture(func(event *tcell.EventKey) *tcell.EventKey {
		switch event.Key() {
		case tcell.KeyEnter:
			form.SetFocus(3)
			return nil
		case tcell.KeyBacktab:
			form.SetFocus(1)
			return nil
		}
		return event
	})
}

func setupButtonInputCaptures(form *tview.Form) {
	okButton := form.GetButton(0)
	okButton.SetInputCapture(func(event *tcell.EventKey) *tcell.EventKey {
		switch event.Key() {
		case tcell.KeyBacktab:
			form.SetFocus(2)
			return nil
		case tcell.KeyRight:
			form.SetFocus(4)
			return nil
		}
		return event
	})
	cancelButton := form.GetButton(1)
	cancelButton.SetInputCapture(func(event *tcell.EventKey) *tcell.EventKey {
		switch event.Key() {
		case tcell.KeyLeft:
			form.SetFocus(3)
			return nil
		case tcell.KeyBacktab:
			form.SetFocus(2)
			return nil
		}
		return event
	})
}
