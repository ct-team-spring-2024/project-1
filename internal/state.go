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
	Queues           map[int]*types.Queue
	Downloads        map[int]*types.Download
	downloadManagers map[int]*network.DownloadManager
}

var State *AppState

func InitState() {
	State = &AppState{
		Queues: make(map[int]*types.Queue),
		Downloads: make(map[int]*types.Download),
		downloadManagers: make(map[int]*network.DownloadManager),
	}
}

func checkToBeInProgress(id int) bool {
	d := State.Downloads[id]
	q := State.Queues[d.QueueId]
	if d.Status == types.Created {
		return true
	}
	if d.Status == types.Failed && d.CurrentRetriesCnt < q.MaxRetriesCount {
		return true
	}
	return false
}

func findInPrpgressCandidates() []types.Download {
	result := make([]types.Download, 0)
	for i := range State.Queues {
		queue := State.Queues[i]
		inProgressCnt, _ := getInProgressDownloads(queue)
		remainingInProgress := queue.MaxInProgressCount - inProgressCnt
		for _, downloadId := range queue.DownloadIds {
			if remainingInProgress == 0 {
				break
			}
			if checkToBeInProgress(downloadId) {
				result = append(result, *State.Downloads[downloadId])
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
	inProgressCandidates = findInPrpgressCandidates()
	slog.Info(fmt.Sprintf("candidates => %+v", inProgressCandidates))
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
	State.Downloads[id].Status = status
	if status == types.InProgress {
		State.Queues[State.Downloads[id].QueueId].CurrentInProgressCount++
	}
	if status == types.Failed {
		State.Downloads[id].CurrentRetriesCnt++
	}
}

func createDownloadManager(downloadId int) {
	downloadManager := network.DownloadManager{
		EventsChan:        make(chan network.DMEvent),
		ResponseEventChan: make(chan network.DMREvent),
	}
	State.downloadManagers[downloadId] = &downloadManager
}
