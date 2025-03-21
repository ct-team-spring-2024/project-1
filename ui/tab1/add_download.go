package tab1

import (
	"fmt"

	"github.com/gdamore/tcell/v2"
	"github.com/rivo/tview"
)

// List of a static set of queue names for testing
var staticQueues = []string{"Default", "Queue1", "Queue2"}

// Initiating the Add Download form
type Tab1 struct {
	Form *tview.Form
}

func NewTab1() *Tab1 {
	// Creating the form and adding the fields
	form := tview.NewForm().
		AddInputField("URL", "", 40, nil, nil).
		AddDropDown("Queue", staticQueues, 0, nil).
		AddInputField("Output File", "", 40, nil, nil)

	// Adding the buttons and their functionalities
	form.AddButton("OK", func() {
		url := form.GetFormItemByLabel("URL").(*tview.InputField).GetText()
		outputFile := form.GetFormItemByLabel("Output File").(*tview.InputField).GetText()

		// Check if the required fields are empty
		if url == "" {
			form.GetFormItemByLabel("URL").(*tview.InputField).SetText("This field is required")
			form.SetFocus(0) // Move focus back to the URL field.
		} else if outputFile == "" {
			form.GetFormItemByLabel("Output File").(*tview.InputField).SetText("This field is required")
			form.SetFocus(2) // Move focus to the Output File field.
		} else {
			fmt.Println("Download added!") // For testing
			// Add the suitable functionality here
			form.GetFormItemByLabel("URL").(*tview.InputField).SetText("")
			form.GetFormItemByLabel("Queue").(*tview.DropDown).SetCurrentOption(0)
			form.GetFormItemByLabel("Output File").(*tview.InputField).SetText("")
			// TODO: Achieve the value of each field and use the provided function to send them to the backend
			form.SetFocus(0) // Move focus back to the URL field
		}
	}).AddButton("Cancel", func() {
		// Clear the form and move focus back to the URL field
		form.GetFormItemByLabel("URL").(*tview.InputField).SetText("")
		form.GetFormItemByLabel("Queue").(*tview.DropDown).SetCurrentOption(0)
		form.GetFormItemByLabel("Output File").(*tview.InputField).SetText("")
		form.SetFocus(0) // Move focus back to the URL field
	})

	form.SetBorder(true).SetTitle("Add Download").SetTitleAlign(tview.AlignLeft)

	// Use SelectedFunc on the dropdown to move focus when an option is selected
	dropdown := form.GetFormItemByLabel("Queue").(*tview.DropDown)
	dropdown.SetSelectedFunc(func(text string, index int) {
		// Immediately move focus to the "Output File" field after selection.
		form.SetFocus(2)
	})

	// Override InputCapture for the Output File field to catch Backtab
	outputField := form.GetFormItemByLabel("Output File").(*tview.InputField)
	outputField.SetInputCapture(func(event *tcell.EventKey) *tcell.EventKey {
		if event.Key() == tcell.KeyBacktab {
			form.SetFocus(1) // Move focus to the dropdown
			return nil
		}
		return event
	})

	// Set a global InputCapture for non-button fields
	form.SetInputCapture(func(event *tcell.EventKey) *tcell.EventKey {
		currentFocus, _ := form.GetFocusedItemIndex()
		switch event.Key() {
		case tcell.KeyEnter:
			if currentFocus == 0 { // URL field
				form.SetFocus(1) // Move to Queue
				return nil
			} else if currentFocus == 2 { // Output File field
				form.SetFocus(3) // Move to OK button
				return nil
			}
		case tcell.KeyBacktab:
			// If we're in the URL field (index 0), do nothing
			if currentFocus == 0 {
				return nil
			} else if currentFocus == 1 { // Queue field
				form.SetFocus(0) // Move to URL
				return nil
			}
			// For Output File, Backtab is handled in its own InputCapture
		}
		return event
	})

	// Defining the InputCapture for the buttons
	// In our form, the first button is at index 3 and the second at index 4.
	okButton := form.GetButton(0)
	okButton.SetInputCapture(func(event *tcell.EventKey) *tcell.EventKey {
		switch event.Key() {
		case tcell.KeyBacktab:
			form.SetFocus(2) // Move focus back to Output File.
			return nil
		case tcell.KeyRight:
			form.SetFocus(4) // Move focus to Cancel button.
			return nil
		}
		return event
	})

	cancelButton := form.GetButton(1)
	cancelButton.SetInputCapture(func(event *tcell.EventKey) *tcell.EventKey {
		switch event.Key() {
		case tcell.KeyLeft:
			form.SetFocus(3) // Move focus to OK button.
			return nil
		case tcell.KeyBacktab:
			form.SetFocus(2) // Move focus to Output File.
			return nil
		}
		return event
	})

	return &Tab1{Form: form}
}
