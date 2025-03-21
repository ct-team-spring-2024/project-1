package tab3

import (
	"fmt"
	"strconv"

	"github.com/gdamore/tcell/v2"
	"github.com/rivo/tview"
)

type Queue struct {
	Name             string
	DownloadFolder   string
	MaxConcurrent    int
	BandwidthLimit   int
	ActiveTimeWindow string
	MaxRetry         int
}

var staticQueues = []Queue{
	{Name: "Default", DownloadFolder: "~/Downloads", MaxConcurrent: 3, BandwidthLimit: 0, ActiveTimeWindow: "", MaxRetry: 0},
	{Name: "Queue1", DownloadFolder: "~/Documents", MaxConcurrent: 2, BandwidthLimit: 500, ActiveTimeWindow: "17:38-06:00", MaxRetry: 1},
}

type Tab3 struct {
	Pages  *tview.Pages
	Table  *tview.Table
	Footer *tview.TextView
	Queues []Queue
}

func NewTab3() *Tab3 {
	tab := &Tab3{
		Pages:  tview.NewPages(),
		Table:  tview.NewTable().SetBorders(true),
		Footer: tview.NewTextView().SetText("N: New   E: Edit   D: Delete").SetTextAlign(tview.AlignCenter),
		Queues: make([]Queue, len(staticQueues)),
	}
	copy(tab.Queues, staticQueues)
	tab.setupTableHeader()
	tab.Table.SetSelectable(true, false)
	tab.refreshTable()
	tab.setupInputCapture()
	layout := tview.NewFlex().SetDirection(tview.FlexRow).
		AddItem(tab.Table, 0, 1, true).
		AddItem(tab.Footer, 1, 0, false)
	tab.Pages.AddPage("main", layout, true, true)
	return tab
}

func (tab *Tab3) setupTableHeader() {
	headers := []string{"Name", "Folder", "Max Concurrent", "Bandwidth (KB/s)", "Time Window", "Max Retry"}
	for i, h := range headers {
		tab.Table.SetCell(0, i, tview.NewTableCell(h).SetSelectable(false))
	}
}

func (tab *Tab3) refreshTable() {
	rowCount := tab.Table.GetRowCount()
	for i := rowCount - 1; i > 0; i-- {
		tab.Table.RemoveRow(i)
	}
	for i, q := range tab.Queues {
		r := i + 1
		tab.Table.SetCell(r, 0, tview.NewTableCell(q.Name))
		tab.Table.SetCell(r, 1, tview.NewTableCell(q.DownloadFolder))
		tab.Table.SetCell(r, 2, tview.NewTableCell(fmt.Sprintf("%d", q.MaxConcurrent)))
		tab.Table.SetCell(r, 3, tview.NewTableCell(fmt.Sprintf("%d", q.BandwidthLimit)))
		tab.Table.SetCell(r, 4, tview.NewTableCell(q.ActiveTimeWindow))
		tab.Table.SetCell(r, 5, tview.NewTableCell(fmt.Sprintf("%d", q.MaxRetry)))
	}
}

func (tab *Tab3) setupInputCapture() {
	tab.Table.SetInputCapture(func(event *tcell.EventKey) *tcell.EventKey {
		row, _ := tab.Table.GetSelection()
		switch event.Rune() {
		case 'n', 'N':
			tab.showNewQueueForm()
			return nil
		case 'e', 'E':
			if row > 0 && row-1 < len(tab.Queues) {
				tab.showEditQueueForm(row - 1)
			}
			return nil
		case 'd', 'D':
			if row > 0 && row-1 < len(tab.Queues) {
				tab.showDeleteConfirm(row - 1)
			}
			return nil
		}
		return event
	})
}

func (tab *Tab3) showNewQueueForm() {
	form := tview.NewForm()
	var name, folder, timeWindow string
	var maxConcurrent, bandwidth, maxRetry int
	form.
		AddInputField("Name", "", 20, nil, func(text string) { name = text }).
		AddInputField("Folder", "", 20, nil, func(text string) { folder = text }).
		AddInputField("Max Concurrent", "1", 5, nil, func(text string) {
			if v, err := strconv.Atoi(text); err == nil {
				maxConcurrent = v
			}
		}).
		AddInputField("Bandwidth (KB/s)", "0", 5, nil, func(text string) {
			if v, err := strconv.Atoi(text); err == nil {
				bandwidth = v
			}
		}).
		AddInputField("Time Window", "", 20, nil, func(text string) { timeWindow = text }).
		AddInputField("Max Retry", "0", 5, nil, func(text string) {
			if v, err := strconv.Atoi(text); err == nil {
				maxRetry = v
			}
		}).
		AddButton("Save", func() {
			newQueue := Queue{
				Name:             name,
				DownloadFolder:   folder,
				MaxConcurrent:    maxConcurrent,
				BandwidthLimit:   bandwidth,
				ActiveTimeWindow: timeWindow,
				MaxRetry:         maxRetry,
			}
			tab.Queues = append(tab.Queues, newQueue)
			tab.refreshTable()
			tab.Pages.RemovePage("modal")
		}).
		AddButton("Cancel", func() {
			tab.Pages.RemovePage("modal")
		})
	form.SetBorder(true).SetTitle("New Queue").SetTitleAlign(tview.AlignLeft)
	modal := tview.NewFlex().AddItem(form, 0, 1, true)
	tab.Pages.AddPage("modal", modal, true, true)
}

func (tab *Tab3) showEditQueueForm(index int) {
	q := tab.Queues[index]
	form := tview.NewForm()
	var folder, timeWindow string
	var maxConcurrent, bandwidth, maxRetry int
	folder = q.DownloadFolder
	timeWindow = q.ActiveTimeWindow
	maxConcurrent = q.MaxConcurrent
	bandwidth = q.BandwidthLimit
	maxRetry = q.MaxRetry
	form.
		AddInputField("Folder", folder, 20, nil, func(text string) { folder = text }).
		AddInputField("Max Concurrent", fmt.Sprintf("%d", maxConcurrent), 5, nil, func(text string) {
			if v, err := strconv.Atoi(text); err == nil {
				maxConcurrent = v
			}
		}).
		AddInputField("Bandwidth (KB/s)", fmt.Sprintf("%d", bandwidth), 5, nil, func(text string) {
			if v, err := strconv.Atoi(text); err == nil {
				bandwidth = v
			}
		}).
		AddInputField("Time Window", timeWindow, 20, nil, func(text string) { timeWindow = text }).
		AddInputField("Max Retry", fmt.Sprintf("%d", maxRetry), 5, nil, func(text string) {
			if v, err := strconv.Atoi(text); err == nil {
				maxRetry = v
			}
		}).
		AddButton("Save", func() {
			tab.Queues[index].DownloadFolder = folder
			tab.Queues[index].MaxConcurrent = maxConcurrent
			tab.Queues[index].BandwidthLimit = bandwidth
			tab.Queues[index].ActiveTimeWindow = timeWindow
			tab.Queues[index].MaxRetry = maxRetry
			tab.refreshTable()
			tab.Pages.RemovePage("modal")
		}).
		AddButton("Cancel", func() {
			tab.Pages.RemovePage("modal")
		})
	form.SetBorder(true).SetTitle("Edit Queue: " + q.Name).SetTitleAlign(tview.AlignLeft)
	modal := tview.NewFlex().AddItem(form, 0, 1, true)
	tab.Pages.AddPage("modal", modal, true, true)
}

func (tab *Tab3) showDeleteConfirm(index int) {
	q := tab.Queues[index]
	modal := tview.NewModal().
		SetText("Are you sure you want to delete the queue \"" + q.Name + "\"?\nAll related downloads will be stopped and removed.").
		AddButtons([]string{"Yes", "No"}).
		SetDoneFunc(func(buttonIndex int, buttonLabel string) {
			if buttonLabel == "Yes" {
				fmt.Printf("Queue \"%s\" deleted; related downloads stopped and removed.\n", q.Name)
				tab.Queues = append(tab.Queues[:index], tab.Queues[index+1:]...)
				tab.refreshTable()
			}
			tab.Pages.RemovePage("modal")
		})
	tab.Pages.AddPage("modal", modal, true, true)
}
