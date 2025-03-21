package main

import (
	"log"
	"time"

	"github.com/gdamore/tcell/v2"

	"go-idm/ui/tab1"
	"go-idm/ui/tab2"
	"go-idm/ui/tab3"

	"github.com/rivo/tview"
)

func main() {
	app := tview.NewApplication()

	// Initating the tabs
	t1 := tab1.NewTab1()
	t2 := tab2.NewTab2()
	t3 := tab3.NewTab3()

	// Add each tab as a page
	pages := tview.NewPages().
		AddPage("tab1", t1.Form, true, false).
		AddPage("tab2", t2.Table, true, true).
		AddPage("tab3", t3.Table, true, false)

	// Header and footer
	header := tview.NewTextView().SetText("[F1] Add Download   [F2] Downloads   [F3] Queues").SetTextAlign(tview.AlignLeft)
	footer := tview.NewTextView().SetText("Escape: Quit").SetTextAlign(tview.AlignLeft)

	// Designing the header, footer and the layout for pages
	layout := tview.NewFlex().SetDirection(tview.FlexRow).
		AddItem(header, 1, 0, false).
		AddItem(pages, 0, 1, true).
		AddItem(footer, 1, 0, false)

	// Setting keystrokes to their respective tabs
	layout.SetInputCapture(func(event *tcell.EventKey) *tcell.EventKey {
		switch event.Key() {
		case tcell.KeyF1:
			pages.SwitchToPage("tab1")
		case tcell.KeyF2:
			pages.SwitchToPage("tab2")
		case tcell.KeyF3:
			pages.SwitchToPage("tab3")
		case tcell.KeyEscape:
			app.Stop()
		}
		return event
	})

	// Coroutine Auto-refresh for the Downloads List
	go func() {
		ticker := time.NewTicker(2 * time.Second)
		defer ticker.Stop()
		for range ticker.C {
			app.QueueUpdateDraw(func() {
				// TODO: Implement the auto-refresh with correct logic here
				if t2.Table.GetRowCount() > 1 {
					t2.Table.GetCell(1, 2).SetText("Updated")
				}
			})
		}
	}()

	// Log for any errors
	if err := app.SetRoot(layout, true).Run(); err != nil {
		log.Fatal(err)
	}
}
