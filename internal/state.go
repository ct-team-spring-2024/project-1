package internal

import (
	"fmt"
	"go-idm/pkg/network"
	"go-idm/types"
	"log/slog"
	"sync"
	"time"

	"github.com/davecgh/go-spew/spew"
)

type AppState struct {
	mu sync.Mutex
	Queues []*types.Queue
	downloadManagers map[int]*network.DownloadManager
}

var State *AppState

func InitState() {
	State = &AppState{
		Queues: make([]*types.Queue, 0, 10),
		downloadManagers: make(map[int]*network.DownloadManager),
	}
}

func checkToBeInProgress(d types.Download) bool {
	if d.Status == types.Created {
		return true
	}
	if d.Status == types.Failed && d.CurrentRetriesCnt < d.Queue.MaxRetriesCount {
		return true
	}
	return false
}

func findInPrpgressCandidates(state *AppState) []types.Download {
	result := make([]types.Download, 0)
	for _, q := range state.Queues {
		inProgressCnt, _ := getInProgressDownloads(*q)
		remainingInProgress := q.MaxInProgressCount - inProgressCnt
		for _, d := range q.Downloads {
			if remainingInProgress == 0 {
				break
			}
			if checkToBeInProgress(*d) {
				result = append(result, *d)
				remainingInProgress--
			}
		}
	}
	return result

}

func UpdaterWithCount(step int) {
	for i := 0; i < step; i++ {
		// Some Queue/Download are added
		updateState()
		time.Sleep(1 * time.Second)
		slog.Info(fmt.Sprintf("================================================================================"))
		slog.Info(fmt.Sprintf("State after step %d => ", i))
		spew.Dump(State)
	}
}

func updateState() {
	State.mu.Lock()
	var inProgressCandidates []types.Download
	inProgressCandidates = findInPrpgressCandidates(State)
	spew.Dump(inProgressCandidates)
	for _, d := range inProgressCandidates {
		updateDownloadStatus(d.Id, types.InProgress)
		createDownloadManager(d.Id)
		spew.Dump(State.downloadManagers)
		go network.AsyncStartDownload(d, State.downloadManagers[d.Id].EventsChan, State.downloadManagers[d.Id].ResponseEventChan)
		// go DownloadManagerHandler(d.Id, State.downloadManagers[d.Id].eventsChan, State.downloadManagers[d.Id].responseEventChan)
		// State.downloadManagers[d.Id].EventsChan <- network.DMEvent{
		//	Etype: network.Startt,
		// }
	}
	State.mu.Unlock()
}

func updateDownloadStatus(id int, status types.DownloadStatus) {
	for _, q := range State.Queues {
		for _, d := range q.Downloads {
			if d.Id == id {
				d.Status = status
				if status == types.InProgress {
					q.CurrentInProgressCount++
				}
				if status == types.Failed {
					d.CurrentRetriesCnt++ // TODO: Fails, because d is a copy
				}
			}
		}
	}
}

func createDownloadManager(downloadId int) {
	downloadManager := network.DownloadManager{
		EventsChan: make(chan network.DMEvent),
		ResponseEventChan: make(chan network.DMREvent),
	}
	State.downloadManagers[downloadId] = &downloadManager
}
