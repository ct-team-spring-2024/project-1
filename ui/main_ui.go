package main

import (
	"log"
	"time"

	"github.com/gdamore/tcell/v2"
	"github.com/rivo/tview"

	"go-idm/ui/tab1"
	"go-idm/ui/tab2"
	"go-idm/ui/tab3"
)

func main() {
	app := tview.NewApplication()
	layout, pages, t2 := buildUILayout(app)
	setGlobalKeyBindings(layout, pages, app)
	go autoRefresh(t2, app)

	if err := app.SetRoot(layout, true).Run(); err != nil {
		log.Fatal(err)
	}
}

func buildUILayout(app *tview.Application) (layout *tview.Flex, pages *tview.Pages, t2 *tab2.Tab2) {
	t1 := tab1.NewTab1()
	t2 = tab2.NewTab2()
	t3 := tab3.NewTab3()

	pages = tview.NewPages().
		AddPage("tab1", t1.Form, true, false).
		AddPage("tab2", t2.Layout, true, true).
		AddPage("tab3", t3.Pages, true, false)

	header := tview.NewTextView().
		SetText("[F1] Add Download   [F2] Downloads   [F3] Queues   [Esc] Quit").
		SetTextAlign(tview.AlignLeft)

	layout = tview.NewFlex().SetDirection(tview.FlexRow).
		AddItem(header, 1, 0, false).
		AddItem(pages, 0, 1, true)

	return layout, pages, t2
}

func setGlobalKeyBindings(layout *tview.Flex, pages *tview.Pages, app *tview.Application) {
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
}

func autoRefresh(t2 *tab2.Tab2, app *tview.Application) {
	ticker := time.NewTicker(2 * time.Second)
	defer ticker.Stop()
	for range ticker.C {
		app.QueueUpdateDraw(func() {
			if t2.Table.GetRowCount() > 1 {
				t2.Table.GetCell(1, 2).SetText("Updated")
			}
		})
	}
}
