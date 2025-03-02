package internal

import (
	"go-idm/types"
	// "log/slog"
	// "fmt"

	"github.com/davecgh/go-spew/spew"
)

type EType int
const (
	start EType = iota
	stop
)

type DMEvent struct {
	etype EType
	// TODO Add data fields for the event
}

type DownloadManager struct {
	eventsChan chan DMEvent
}

type AppState struct {
	Queues []*types.Queue
	downloadManagers map[int]DownloadManager
}

var State *AppState

func InitState() {
	State = &AppState{
		Queues: make([]*types.Queue, 0, 10),
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

func UpdateState() {
	var inProgressCandidates []types.Download
	inProgressCandidates = findInPrpgressCandidates(State)
	spew.Dump(inProgressCandidates)
	for _, d := range inProgressCandidates {
		updateDownloadStatus(d.Id, types.InProgress)
		// pass to network
		// result := network.SyncStartDownload(v)
		// switch e := result.Err; {
		// case e == nil:
		//	updateDownloadStatus(v.Id, types.Completed)
		// default:
		//	updateDownloadStatus(v.Id, types.Failed)
		// }
		// create download manager
		createDownloadManager(d.Id)
	}
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
	// create a download manager and goroutine for it
	downloadManager := DownloadManager{
		eventsChan: make(chan DMEvent, 0),
	}
	go DownloadManagerHandler(downloadId, downloadManager.eventsChan)
}
