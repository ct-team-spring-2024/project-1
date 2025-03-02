package main

import (
	"fmt"
	"go-idm/internal"
	"go-idm/types"
	"log/slog"
	"time"

	"github.com/davecgh/go-spew/spew"
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
	spew.Dump(internal.State)
	for i := 0; i < 1; i++ {
		// Some Queue/Download are added
		internal.UpdateState()
		time.Sleep(1 * time.Second)
		slog.Info(fmt.Sprintf("State after step %d => ", i))
		spew.Dump(internal.State)
	}
}
