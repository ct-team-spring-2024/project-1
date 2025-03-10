package internal

import (
	"fmt"
	"github.com/davecgh/go-spew/spew"
	"log/slog"
	"sync"
	"time"

	"go-idm/pkg/network"
	"go-idm/types"
)

type AppState struct {
	mu               sync.Mutex
	Queues           []*types.Queue
	downloadManagers map[int]*network.DownloadManager
}

var State *AppState

func InitState() {
	State = &AppState{
		Queues:           make([]*types.Queue, 0, 10),
		downloadManagers: make(map[int]*network.DownloadManager),
	}
}

func checkToBeInProgress(id int) bool {
	i, j := FindDownload(id)
	q := State.Queues[i]
	d := q.Downloads[j]
	if d.Status == types.Created {
		return true
	}
	if d.Status == types.Failed && d.CurrentRetriesCnt < q.MaxRetriesCount {
		return true
	}
	return false
}

func findInPrpgressCandidates(state *AppState) []types.Download {

	result := make([]types.Download, 0)
	for i, _ := range State.Queues {
		queue := State.Queues[i]
		inProgressCnt, _ := getInProgressDownloads(queue)
		remainingInProgress := queue.MaxInProgressCount - inProgressCnt
		for i, _ := range queue.Downloads {
			if remainingInProgress == 0 {
				break
			}
			if checkToBeInProgress(queue.Downloads[i].Id) {
				result = append(result, *queue.Downloads[i])
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
		if i%5 == 0 {
			slog.Info(fmt.Sprintf("================================================================================"))
			slog.Info(fmt.Sprintf("State after step %d => ", i))
			spew.Dump(State)
		}
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
		queue, _ := FindQueue(d.Id)
		go network.AsyncStartDownload(d, *queue, State.downloadManagers[d.Id].EventsChan, State.downloadManagers[d.Id].ResponseEventChan)
		// setup listener for each of the generated downloads.
		go func() {
			for responseEvent := range State.downloadManagers[d.Id].ResponseEventChan {
				switch responseEvent.Etype {
				case network.Completed:
					slog.Info(fmt.Sprintf("Response Event for %d => %s", d.Id, spew.Sdump(responseEvent)))
				case network.Failure:
					slog.Info(fmt.Sprintf("Response Event for %d => %s", d.Id, spew.Sdump(responseEvent)))
				}
			}
		}()
		// go DownloadManagerHandler(d.Id, State.downloadManagers[d.Id].eventsChan, State.downloadManagers[d.Id].responseEventChan)
		// State.downloadManagers[d.Id].EventsChan <- network.DMEvent{
		//	Etype: network.Startt,
		// }
	}
	State.mu.Unlock()
}

func updateDownloadStatus(id int, status types.DownloadStatus) {
	i, j := FindDownload(id)
	State.Queues[i].Downloads[j].Status = status
	if status == types.InProgress {
		State.Queues[i].CurrentInProgressCount++
	}
	if status == types.Failed {
		State.Queues[i].Downloads[j].CurrentRetriesCnt++
	}
	// for _, q := range State.Queues {
	// 	for _, d := range q.Downloads {
	// 		if d.Id == id {
	// 			d.Status = status
	// 			if status == types.InProgress {
	// 				q.CurrentInProgressCount++
	// 			}
	// 			if status == types.Failed {
	// 				d.CurrentRetriesCnt++ // TODO: Fails, because d is a copy
	// 			}
	// 		}
	// 	}
	// }
}

func createDownloadManager(downloadId int) {
	downloadManager := network.DownloadManager{
		EventsChan:        make(chan network.DMEvent),
		ResponseEventChan: make(chan network.DMREvent),
	}
	State.downloadManagers[downloadId] = &downloadManager
}
