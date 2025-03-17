package main

import (
	"fmt"
	"log/slog"
	"os"

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
	internal.AddDownload(d1, d1.Id)
	internal.AddDownload(d2, d2.QueueId)
	internal.AddDownload(d3, d3.QueueId)
	slog.Info("Initial State =>")
	spew.Dump(internal.State)
	internal.UpdaterWithCount(1, make(map[int][]internal.IDMEvent))

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
	// q.Destination = "C:/Users/Asus/Documents/GitHub/project-1/files"
	q.Destination = "./files"
	internal.AddQueue(q)
	d.Filename = "downloaded.bin"
	d.Url = "http://127.0.0.1:8080"
	internal.AddDownload(d, d.Id)

	//spew.Dump(internal.State)
	// TODO: because of not having DI, we cannot
	// mock the network component, so testing is limited.
	internal.UpdaterWithCount(120, make(map[int][]internal.IDMEvent))
}
func t3() {
	opts := &slog.HandlerOptions{
		Level: slog.LevelInfo,
	}
	handler := slog.NewTextHandler(os.Stdout, opts)
	logger := slog.New(handler)
	slog.SetDefault(logger)

	internal.InitState()
	q := types.NewQueue(0)
	d1 := types.NewDownload(0, q)
	d2 := types.NewDownload(1, q)
	q.MaxInProgressCount = 1
	// q.Destination = "C:/Users/Asus/Documents/GitHub/project-1/files"
	q.Destination = "./files"
	internal.AddQueue(q)
	d1.Filename = "downloaded-1.bin"
	d2.Filename = "downloaded-2.bin"
	d1.Url = "http://127.0.0.1:8080"
	d2.Url = "http://127.0.0.1:8080"
	internal.AddDownload(d1, q.Id)
	internal.AddDownload(d2, q.Id)

	//spew.Dump(internal.State)
	// TODO: because of not having DI, we cannot
	// mock the network component, so testing is limited.
	slog.Info("Initial State =>")
	spew.Dump(internal.State)
	internal.UpdaterWithCount(250, make(map[int][]internal.IDMEvent))
}

func t4ChangingConfiguration() {
	opts := &slog.HandlerOptions{
		Level: slog.LevelInfo,
	}
	handler := slog.NewTextHandler(os.Stdout, opts)
	logger := slog.New(handler)
	slog.SetDefault(logger)

	internal.InitState()
	q := types.NewQueue(0)
	d1 := types.NewDownload(0, q)
	q.MaxInProgressCount = 1
	q.Destination = "C:/Users/Asus/Documents/GitHub/project-1/files"
	//	q.Destination = "./files"
	q.MaxBandwidth = 10 * 1024 * 1024
	internal.AddQueue(q)
	d1.Filename = "downloaded.bin"
	d1.Url = "http://127.0.0.1:8080"
	internal.AddDownload(d1, q.Id)

	slog.Info("Initial State =>")
	spew.Dump(internal.State)
	eventsMap := make(map[int][]internal.IDMEvent)
	eventsMap[10] = []internal.IDMEvent{internal.NewModifyQueueEvent(q.Id, 5*1024*1024)}
	slog.Info(fmt.Sprintf("GOOZ %+v", eventsMap))
	internal.UpdaterWithCount(250, eventsMap)
}
func t5TestingPauseAndResume() {
	opts := &slog.HandlerOptions{
		Level: slog.LevelInfo,
	}
	handler := slog.NewTextHandler(os.Stdout, opts)
	logger := slog.New(handler)
	slog.SetDefault(logger)

	internal.InitState()
	q := types.NewQueue(0)
	d1 := types.NewDownload(0, q)
	q.MaxInProgressCount = 1
	q.Destination = "C:/Users/Asus/Documents/GitHub/project-1/files"
	//	q.Destination = "./files"
	q.MaxBandwidth = 10 * 1024 * 1024
	internal.AddQueue(q)
	d1.Filename = "downloaded.bin"
	d1.Url = "http://127.0.0.1:8080"
	internal.AddDownload(d1, q.Id)

	slog.Info("Initial State =>")
	spew.Dump(internal.State)
	eventsMap := make(map[int][]internal.IDMEvent)
	eventsMap[10] = []internal.IDMEvent{internal.NewPauseDownloadEvent(0)}
	eventsMap[20] = []internal.IDMEvent{internal.NewResumeDownloadEvent(0)}
	slog.Info(fmt.Sprintf("GOOZ %+v", eventsMap))

	internal.UpdaterWithCount(250, eventsMap)
}

func main() {
	// t1()
	//t4ChangingConfiguration()
	t5TestingPauseAndResume()
}
