package main

import (
	"fmt"
	"log/slog"
	"os"
	"time"

	"github.com/gdamore/tcell/v2"
	"github.com/rivo/tview"

	"go-idm/internal"
	"go-idm/ui/tab1"
	"go-idm/ui/tab2"
	"go-idm/ui/tab3"
)


func main() {
	// Logging
	logFile, err := os.OpenFile("out.log", os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		slog.Error("Failed to open log file", "error", err)
		return
	}
	defer logFile.Close()
	opts := &slog.HandlerOptions{
		Level: slog.LevelDebug,
	}
	handler := slog.NewTextHandler(logFile, opts)
	logger := slog.New(handler)
	slog.SetDefault(logger)
	slog.Info("Logging initialized")

	// Setup Logic
	internal.InitState()

	go func() {
		var eventBuffer []internal.IDMEvent
		ticker := time.NewTicker(200 * time.Millisecond)
		defer ticker.Stop()

		for {
			select {
			case event := <-internal.LogicIn:
				slog.Info("JAN gOOZ?")
				eventBuffer = append(eventBuffer, event)
			case <-ticker.C:
				if len(eventBuffer) > 0 {
					slog.Info("JAN?")
					internal.UpdateState(eventBuffer)
					slog.Info(fmt.Sprintf("eventBuffer KIR KHAR %+v %v", eventBuffer, eventBuffer[0]))
					eventBuffer = nil
				}
			}
		}
		slog.Info("DEAD")
	}()

	// UI
	app := tview.NewApplication()

	layout, pages, t2 := buildUILayout(app)
	setGlobalKeyBindings(layout, pages, app)
	go autoRefresh(t2, app)

	defer func() {
		if err := recover(); err != nil {
			slog.Error(fmt.Sprintf("Panic occurred: %v", err))
			app.Stop() // Ensure the terminal is reset
		}
	}()

	// Your application logic here
	if err := app.SetRoot(layout, true).Run(); err != nil {
		slog.Error(fmt.Sprintf("Application error: %v", err))
	}
	if err := app.SetRoot(layout, true).Run(); err != nil {
		panic(err)
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
