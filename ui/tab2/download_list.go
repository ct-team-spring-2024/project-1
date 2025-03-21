package tab2

import (
	"go-idm/internal"
	"go-idm/types"
	"strconv"

	"github.com/gdamore/tcell/v2"
	"github.com/rivo/tview"
)

var staticDownloads []types.Download

type Tab2 struct {
	Table  *tview.Table
	Layout *tview.Flex
	Footer *tview.TextView
}

func NewTab2() *Tab2 {
	staticDownloads := internal.GetDownloads()
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

func createDownloadTable(downloads []types.Download) *tview.Table {
	table := tview.NewTable().SetBorders(true)
	headers := []string{"URL", "Queue", "Status", "Progress", "Speed (KB/s)"}
	for i, header := range headers {
		table.SetCell(0, i, tview.NewTableCell(header).SetSelectable(false))
	}
	table.SetSelectable(true, false)
	f := func(x int) string {
		result := strconv.Itoa(x)
		return result
	}
	for i, d := range downloads {
		row := i + 1
		table.SetCell(row, 0, tview.NewTableCell(d.Url))
		table.SetCell(row, 1, tview.NewTableCell(f(d.QueueId)))
		table.SetCell(row, 2, tview.NewTableCell(f(int(d.Status))))
		table.SetCell(row, 3, tview.NewTableCell("UNKNOWN"))
		table.SetCell(row, 4, tview.NewTableCell("UNKNONW"))
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
