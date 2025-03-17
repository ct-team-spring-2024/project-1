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
	// TODO Now download managers are killed if not running.
	//      But stroing the bytes offset means they should be not.
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

	result := false
	d := State.Downloads[id]
	q := State.Queues[d.QueueId]
	if d.Status == types.Created {
		result = true
	}
	if d.Status == types.Failed && d.CurrentRetriesCnt < q.MaxRetriesCount {
		result = true
	}

	return result
}

func findInPrpgressCandidates() []types.Download {
	State.mu.Lock()

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

	State.mu.Unlock()
	return result
}

func UpdaterWithCount(step int, events map[int][]IDMEvent) {
	for i := 0; i < step; i++ {
		// Some Queue/Download are added
		updateState(events[i])
		time.Sleep(1 * time.Second)
		if i % 5 == 0 {
			slog.Info(fmt.Sprintf("================================================================================"))
			slog.Info(fmt.Sprintf("State after step %d => ", i))
			spew.Dump(State)
		}
	}
}

func updateState(events []IDMEvent) {
	// 1. Process new events sent by user.
	//    Send the new configuration using the chIn.
	for _, e := range events {
		switch e.EType {
		case AddQueueEvent:
		case ModifyQueueEvent:
			data := e.Data.(ModifyQueueEventData)
			slog.Info(fmt.Sprintf("Modify Event => %v", data))
			dm := State.downloadManagers[data.queueId]
			dm.EventsChan <- network.NewReconfigDMEvent(data.newMaxBandwidth) // TODO follow!!!
			slog.Info("Modify Event End")
		case PauseDownloadEvent:
		case ResumeDownloadEvent:
		case DeleteDownloadEvent:
		}
	}
	// 2. Fire up new candidates
	var inProgressCandidates []types.Download
	inProgressCandidates = findInPrpgressCandidates()
	for _, d := range inProgressCandidates {
		updateDownloadStatus(d.Id, types.InProgress)
		createDownloadManager(d.Id)
		chIn, chOut := getChannel(d.Id)
		queue := getQueue(d.Id)
		slog.Debug(fmt.Sprintf("candidates => %v %v %v %v", d, *queue, chIn, chOut))
		go network.AsyncStartDownload(d, *queue, chIn, chOut)
		// setup listener for each of the generated downloads.
		go func() {
			for responseEvent := range chOut {
				// Motherfucker! this log caused the concurrent access error.
				// (because the log uses the responseEvent in a async way, but it is also processed after ward!)
				// slog.Info(fmt.Sprintf("Response Event for %d => %+v", d.Id, responseEvent))
				switch responseEvent.EType {
				case network.Completed:
					slog.Debug(fmt.Sprintf("DMR : Completed %d", d.Id))
					updateDownloadStatus(d.Id, types.Completed)
				case network.Failure:
					slog.Debug(fmt.Sprintf("DMR : Failure %d", d.Id))
					updateDownloadStatus(d.Id, types.Failed)
				case network.InProgress:
					slog.Debug(fmt.Sprintf("DMR : InProgress %d", d.Id))
					updateDMChunksByteOffset(d.Id, responseEvent.CurrentChunksByteOffset)
				}
			}
		}()
	}
}

// TODO: This is unneccessery contention! Because we are storing the chans in a
//       map, then we have to lock.
func getChannel(id int) (chan network.DMEvent, chan network.DMREvent) {
	State.mu.Lock()

	r1, r2 := State.downloadManagers[id].EventsChan, State.downloadManagers[id].ResponseEventChan

	State.mu.Unlock()

	return r1, r2
}

func getQueue(id int) *types.Queue {
	State.mu.Lock()

	download := State.Downloads[id]
	result := State.Queues[download.QueueId]

	State.mu.Unlock()

	return result
}

func updateDownloadStatus(id int, status types.DownloadStatus) {
	State.mu.Lock()

	oldStatus := State.Downloads[id].Status
	State.Downloads[id].Status = status

	slog.Info(fmt.Sprintf("id %d %v %v", id, status, oldStatus))
	spew.Dump(State)
	// 1: Update CurrentInProgressCount
	if oldStatus == types.InProgress && (status == types.Failed || status == types.Completed) {
		State.Queues[State.Downloads[id].QueueId].CurrentInProgressCount--
	}
	if status == types.InProgress {
		State.Queues[State.Downloads[id].QueueId].CurrentInProgressCount++
	}
	// 1: Update CurrentRetriesCnt
	if status == types.Failed {
		State.Downloads[id].CurrentRetriesCnt++
	}

	slog.Info("Update done")

	State.mu.Unlock()
}

func updateDMChunksByteOffset(downloadId int, currentChunksByteOffset map[int]int) {
	State.mu.Lock()

	State.downloadManagers[downloadId].ChunksByteOffset = currentChunksByteOffset

	State.mu.Unlock()
}

func createDownloadManager(downloadId int) {
	State.mu.Lock()

	downloadManager := network.DownloadManager{
		EventsChan:        make(chan network.DMEvent),
		ResponseEventChan: make(chan network.DMREvent),
	}
	State.downloadManagers[downloadId] = &downloadManager


	State.mu.Unlock()
}
