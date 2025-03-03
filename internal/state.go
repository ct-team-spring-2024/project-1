package internal

import (
	"go-idm/types"
	"log/slog"
	"fmt"
	"time"

	"github.com/davecgh/go-spew/spew"
)

type EType int
const (
	start EType = iota
	stop
)

type REType int
const (
	completed EType = iota
	failure
)

type DMEvent struct {
	etype EType
	// TODO Add data fields for the event
}

type DMREvent struct {
	etype REType
}

type DownloadManager struct {
	eventsChan chan DMEvent
	responseEventChan chan DMREvent
}

type AppState struct {
	Queues []*types.Queue
	downloadManagers map[int]*DownloadManager
}

var State *AppState

func InitState() {
	State = &AppState{
		Queues: make([]*types.Queue, 0, 10),
		downloadManagers: make(map[int]*DownloadManager),
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
		spew.Dump(State.downloadManagers)
		go DownloadManagerHandler(d.Id, State.downloadManagers[d.Id].eventsChan)
		State.downloadManagers[d.Id].eventsChan <- DMEvent{
			etype: start,
		}
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

// create a download manager
func createDownloadManager(downloadId int) {
	downloadManager := DownloadManager{
		eventsChan: make(chan DMEvent),
		responseEventChan: make(chan DMREvent),
	}
	State.downloadManagers[downloadId] = &downloadManager
}
