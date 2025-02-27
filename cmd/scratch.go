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
	d := types.NewDownload(0)
	internal.AddDownload(d)
	slog.Info(fmt.Sprintf("state 2 => %+v", internal.State.Queues[0]))
	internal.Delete(d)
	slog.Info(fmt.Sprintf("state 3 => %+v", internal.State.Queues[0]))
}
