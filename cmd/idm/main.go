package main

import (
	"fmt"
	"log/slog"
	"os"
	"time"

	"github.com/davecgh/go-spew/spew"

	"go-idm/internal"
	"go-idm/types"
)

// same as two. DELETE
func t1() {
	internal.InitState()
	q := types.NewQueue(0)
	d1 := types.NewDownload(0, q)
	d2 := types.NewDownload(1, q)
	d3 := types.NewDownload(2, q)
	now := time.Now()
	q.MaxInProgressCount = 1
	q.Destination = "./files"
	q.MaxBandwidth = 50 * 1024 * 1024
	q.ActiveInterval = types.TimeInterval{
		Start: now.Add(-10 * time.Minute),
		End:   now.Add(10 * time.Minute),
	}
	d1.Filename = "d1.bin"
	d2.Filename = "d2.bin"
	d3.Filename = "d3.bin"
	d1.Url = "http://127.0.0.1:8080"
	d2.Url = "http://127.0.0.1:8080"
	d3.Url = "http://127.0.0.1:8080"
	internal.AddQueue(q)
	internal.AddDownload(d1, d1.QueueId)
	internal.AddDownload(d2, d2.QueueId)
	internal.AddDownload(d3, d3.QueueId)
	slog.Info("Initial State =>")
	spew.Dump(internal.State)
	internal.UpdaterWithCount(1, make(map[int][]internal.IDMEvent))

}

// 1 download
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
	now := time.Now()
	q.MaxInProgressCount = 1
	// q.Destination = "C:/Users/Asus/Documents/GitHub/project-1/files"
	q.Destination = "./files"
	q.MaxBandwidth = 15 * 1024 * 1024
	q.ActiveInterval = types.TimeInterval{
		Start: now.Add(-10 * time.Minute),
		End:   now.Add(10 * time.Minute),
	}
	internal.AddQueue(q)
	d.Filename = "downloaded.bin"
	d.Url = "http://127.0.0.1:8080"
	internal.AddDownload(d, d.Id)

	//spew.Dump(internal.State)
	// TODO: because of not having DI, we cannot
	// mock the network component, so testing is limited.
	internal.UpdaterWithCount(120, make(map[int][]internal.IDMEvent))
}

// 2 downloads
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
	now := time.Now()
	q.MaxInProgressCount = 2
	// q.Destination = "C:/Users/Asus/Documents/GitHub/project-1/files"
	q.Destination = "./files"
	q.MaxBandwidth = 15 * 1024 * 1024
	q.ActiveInterval = types.TimeInterval{
		Start: now.Add(-10 * time.Minute),
		End:   now.Add(10 * time.Minute),
	}
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
	now := time.Now()
	q.MaxInProgressCount = 1
	// q.Destination = "C:/Users/Asus/Documents/GitHub/project-1/files"
	q.Destination = "./files"
	q.MaxBandwidth = 9 * 1024 * 1024
	q.ActiveInterval = types.TimeInterval{
		Start: now.Add(-10 * time.Minute),
		End:   now.Add(10 * time.Minute),
	}
	internal.AddQueue(q)
	d1.Filename = "downloaded.bin"
	d1.Url = "http://127.0.0.1:8080"
	internal.AddDownload(d1, q.Id)

	slog.Info("Initial State =>")
	spew.Dump(internal.State)
	eventsMap := make(map[int][]internal.IDMEvent)
	newBandwidth := 3 * 1024 * 1024
	eventsMap[10] = []internal.IDMEvent{internal.NewModifyQueueEvent(q.Id, &newBandwidth, nil)}
	slog.Info(fmt.Sprintf("GOOZ %+v", eventsMap))
	internal.UpdaterWithCount(250, eventsMap)
}

func tActiveInterval() {
	opts := &slog.HandlerOptions{
		Level: slog.LevelInfo,
	}
	handler := slog.NewTextHandler(os.Stdout, opts)
	logger := slog.New(handler)
	slog.SetDefault(logger)

	internal.InitState()
	q := types.NewQueue(0)
	d1 := types.NewDownload(0, q)
	now := time.Now()
	q.MaxInProgressCount = 1
	q.ActiveInterval = types.TimeInterval{
		Start: now.Add(-10 * time.Minute),
		End:   now.Add(10 * time.Minute),
	}
	q.Destination = "C:/Users/Asus/Documents/GitHub/project-1/files"
	q.Destination = "./files"
	q.MaxBandwidth = 9 * 1024 * 1024
	internal.AddQueue(q)
	d1.Filename = "downloaded.bin"
	d1.Url = "http://127.0.0.1:8080"
	internal.AddDownload(d1, q.Id)

	slog.Info("Initial State =>")
	spew.Dump(internal.State)
	eventsMap := make(map[int][]internal.IDMEvent)
	eventsMap[10] = []internal.IDMEvent{internal.NewModifyQueueEvent(q.Id, nil, &types.TimeInterval{
		Start: now.Add(10 * time.Minute),
		End:   now.Add(20 * time.Minute),
	})}
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
	now := time.Now()
	q.ActiveInterval = types.TimeInterval{
		Start: now.Add(-10 * time.Minute),
		End:   now.Add(10 * time.Minute),
	}
	q.Destination = "C:/Users/Asus/Documents/GitHub/project-1/files"
	//	q.Destination = "./files"
	q.MaxBandwidth = 28 * 1024 * 1024
	internal.AddQueue(q)
	d1.Filename = "downloaded.mp4"
	d1.Url = "https://dl33.deserver.top/www2/serial/Daredevil.Born.Again/s01/Daredevil.Born.Again.S01E04.REPACK.720p.WEB-DL.SoftSub.DigiMoviez.mkv?md5=8pKAOCubgbXPCqJFKHnCXw&expires=1742751594"
	internal.AddDownload(d1, q.Id)

	slog.Info("Initial State =>")
	spew.Dump(internal.State)
	eventsMap := make(map[int][]internal.IDMEvent)
	eventsMap[9] = []internal.IDMEvent{internal.NewPauseDownloadEvent(0)}
	eventsMap[18] = []internal.IDMEvent{internal.NewResumeDownloadEvent(0)}
	slog.Info(fmt.Sprintf("GOOZ %+v", eventsMap))

	internal.UpdaterWithCount(2000, eventsMap)
}

func main() {
	// t4ChangingConfiguration()
	//tActiveInterval()
	t5TestingPauseAndResume()
}
