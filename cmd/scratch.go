package main

import (
	"fmt"
	"go-idm/internal"
	"go-idm/types"
	"log/slog"
)

func main() {
	internal.InitState()
	slog.Info(fmt.Sprintf("state 0 => %+v", internal.State))
	q := types.NewQueue(0)
	internal.AddQueue(q)
	slog.Info(fmt.Sprintf("state 1 => %+v", internal.State.Queues[0]))
	d := types.NewDownload(0, q)
	internal.AddDownload(d, d.QueueId)
	slog.Info(fmt.Sprintf("state 2 => %+v", internal.State.Queues[0]))
	internal.Delete(d.Id)
	slog.Info(fmt.Sprintf("state 3 => %+v", internal.State.Queues[0]))
}
