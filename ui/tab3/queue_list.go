package tab3

import (
	"fmt"

	"github.com/rivo/tview"
)

// Initiate the Queue struct
type Queue struct {
	Name             string
	DownloadFolder   string
	MaxConcurrent    int
	BandwidthLimit   int // in KB/s
	ActiveTimeWindow string
	MaxRetry         int
}

// For testing, we define static queue data
var staticQueues = []Queue{
	{Name: "Default", DownloadFolder: "~/Downloads", MaxConcurrent: 3, BandwidthLimit: 0, ActiveTimeWindow: "", MaxRetry: 0},
	{Name: "Queue1", DownloadFolder: "~/Documents", MaxConcurrent: 2, BandwidthLimit: 500, ActiveTimeWindow: "17:38-06:00", MaxRetry: 1},
}

// Queue table struct
type Tab3 struct {
	Table *tview.Table
}

func NewTab3() *Tab3 {
	// Create a table and set its properties
	table := tview.NewTable().SetBorders(true)
	headers := []string{"Name", "Folder", "Max Concurrent", "Bandwidth (KB/s)", "Time Window", "Max Retry"}
	for i, h := range headers {
		table.SetCell(0, i, tview.NewTableCell(h).SetSelectable(false))
	}
	// Make the table selectable in x axis
	table.SetSelectable(true, false)

	// TODO: replace the staticQueue with dynamicQueue and implement the logic to update the table
	for i, q := range staticQueues {
		table.SetCell(i+1, 0, tview.NewTableCell(q.Name))
		table.SetCell(i+1, 1, tview.NewTableCell(q.DownloadFolder))
		table.SetCell(i+1, 2, tview.NewTableCell(fmt.Sprintf("%d", q.MaxConcurrent)))
		table.SetCell(i+1, 3, tview.NewTableCell(fmt.Sprintf("%d", q.BandwidthLimit)))
		table.SetCell(i+1, 4, tview.NewTableCell(q.ActiveTimeWindow))
		table.SetCell(i+1, 5, tview.NewTableCell(fmt.Sprintf("%d", q.MaxRetry)))
	}
	return &Tab3{Table: table}
}
