package main

import (
	"log/slog"
	"github.com/davecgh/go-spew/spew"

	"go-idm/internal"
	"go-idm/types"
)


func main() {
	internal.InitState()
	q := types.NewQueue(0)
	d1 := types.NewDownload(0, q)
	d2 := types.NewDownload(1, q)
	d3 := types.NewDownload(2, q)
	q.MaxInProgressCount = 1
	internal.AddQueue(q)
	internal.AddDownload(d1)
	internal.AddDownload(d2)
	internal.AddDownload(d3)
	// slog.Info(fmt.Sprintf("SSS => %+v", internal.State.Queues[0]))

	// slog.Info(fmt.Sprintf("SSS => %+v", internal.State.Queues[0]))
	slog.Info("Initial State =>")
	spew.Dump(internal.State)
	internal.UpdaterWithCount(1)
}
