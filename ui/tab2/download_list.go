package tab2

import (
	"fmt"

	"github.com/gdamore/tcell/v2"
	"github.com/rivo/tview"
)

type Download struct {
	URL       string
	QueueName string
	Status    string
	Progress  float64
	Speed     float64
}

var staticDownloads = []Download{
	{URL: "http://example.com/file1", QueueName: "Default", Status: "Completed", Progress: 100, Speed: 500},
	{URL: "http://example.com/file2", QueueName: "Default", Status: "Downloading", Progress: 50, Speed: 250},
}

type Tab2 struct {
	Table  *tview.Table
	Layout *tview.Flex
	Footer *tview.TextView
}

func NewTab2() *Tab2 {
	table := createDownloadTable(staticDownloads)
	setDownloadInputCapture(table)
	footer := createFooter()
	layout := tview.NewFlex().SetDirection(tview.FlexRow).
		AddItem(table, 0, 1, true).
		AddItem(footer, 1, 0, false)
	return &Tab2{
		Table:  table,
		Layout: layout,
		Footer: footer,
	}
}

func createDownloadTable(downloads []Download) *tview.Table {
	table := tview.NewTable().SetBorders(true)
	headers := []string{"URL", "Queue", "Status", "Progress", "Speed (KB/s)"}
	for i, header := range headers {
		table.SetCell(0, i, tview.NewTableCell(header).SetSelectable(false))
	}
	table.SetSelectable(true, false)
	for i, d := range downloads {
		row := i + 1
		table.SetCell(row, 0, tview.NewTableCell(d.URL))
		table.SetCell(row, 1, tview.NewTableCell(d.QueueName))
		table.SetCell(row, 2, tview.NewTableCell(d.Status))
		table.SetCell(row, 3, tview.NewTableCell(fmt.Sprintf("%.0f%%", d.Progress)))
		table.SetCell(row, 4, tview.NewTableCell(fmt.Sprintf("%.2f", d.Speed)))
	}
	return table
}

func setDownloadInputCapture(table *tview.Table) {
	table.SetInputCapture(func(event *tcell.EventKey) *tcell.EventKey {
		row, _ := table.GetSelection()
		if row <= 0 {
			return event
		}
		switch event.Rune() {
		case 'r', 'R':
			table.GetCell(row, 2).SetText("Retrying")
		case 'p', 'P':
			cell := table.GetCell(row, 2)
			if cell.Text == "Paused" {
				cell.SetText("Downloading")
			} else {
				cell.SetText("Paused")
			}
		case 'm', 'M':
			table.RemoveRow(row)
		}
		return event
	})
}

func createFooter() *tview.TextView {
	return tview.NewTextView().
		SetText("R: Retry   P: Pause/Resume   M: Remove").
		SetTextAlign(tview.AlignCenter)
}
