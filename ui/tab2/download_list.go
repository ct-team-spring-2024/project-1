package tab2

import (
	"fmt"

	"github.com/rivo/tview"
)

// Initiate the Downloads List table
type Download struct {
	URL       string
	QueueName string
	Status    string
	Progress  float64
	Speed     float64
}

// For testing, we define a static list of downloads
var staticDownloads = []Download{
	{URL: "http://example.com/file1", QueueName: "Default", Status: "Completed", Progress: 100, Speed: 500},
	{URL: "http://example.com/file2", QueueName: "Default", Status: "Downloading", Progress: 50, Speed: 250},
}

// Initiating the Downloads List table
type Tab2 struct {
	Table *tview.Table
}

func NewTab2() *Tab2 {
	// Create the table and set its properities
	table := tview.NewTable().SetBorders(true)
	headers := []string{"URL", "Queue", "Status", "Progress", "Speed (KB/s)"}
	for i, h := range headers {
		table.SetCell(0, i, tview.NewTableCell(h).SetSelectable(false))
	}

	// Set the table to be selectable, but only by row.
	table.SetSelectable(true, false)

	// TODO: Add a function to update the table with new downloads
	// For now, we just add the static downloads
	for i, d := range staticDownloads {
		table.SetCell(i+1, 0, tview.NewTableCell(d.URL))
		table.SetCell(i+1, 1, tview.NewTableCell(d.QueueName))
		table.SetCell(i+1, 2, tview.NewTableCell(d.Status))
		table.SetCell(i+1, 3, tview.NewTableCell(fmt.Sprintf("%.0f%%", d.Progress)))
		table.SetCell(i+1, 4, tview.NewTableCell(fmt.Sprintf("%.2f", d.Speed)))
	}
	return &Tab2{Table: table}
}
