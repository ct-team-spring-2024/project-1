package main

import (
	"os"
	"log/slog"
	"github.com/davecgh/go-spew/spew"

	"go-idm/internal"
	"go-idm/types"
)


func t1() {
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
	slog.Info("Initial State =>")
	spew.Dump(internal.State)
	internal.UpdaterWithCount(1)

}
func t2() {
	opts := &slog.HandlerOptions{
		Level: slog.LevelInfo,
	}
	handler := slog.NewTextHandler(os.Stdout, opts)
	logger := slog.New(handler)
	slog.SetDefault(logger)

	internal.InitState()
	q := types.NewQueue(0)
	d := types.NewDownload(0, q)
	q.MaxInProgressCount = 1
	q.Destination = "./files"
	internal.AddQueue(q)
	d.Filename = "downloaded.bin"
	d.Url = "http://127.0.0.1:8080"
	internal.AddDownload(d)
	spew.Dump(internal.State)
	// TODO: because of not having DI, we cannot
	// mock the network component, so testing is limited.
	internal.UpdaterWithCount(120)
}
func main() {
	t2()
}
