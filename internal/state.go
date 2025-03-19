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
	mu        sync.Mutex
	Queues    map[int]*types.Queue
	Downloads map[int]*types.Download
	// TODO Now download managers are killed if not running.
	//      But stroing the bytes offset means they should be not.
	downloadManagers map[int]*network.DownloadManager
}

var State *AppState

func InitState() {
	State = &AppState{
		Queues:           make(map[int]*types.Queue),
		Downloads:        make(map[int]*types.Download),
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

func findResetCandidates() []types.Download {
	State.mu.Lock()

	result := make([]types.Download, 0)
	for i := range State.Queues {
		q := State.Queues[i]
		inProgressCnt, _ := getInProgressDownloads(q)
		extraInProgress := inProgressCnt - q.MaxInProgressCount
		for _, downloadId := range q.DownloadIds {
			d := State.Downloads[downloadId]
			if extraInProgress > 0 && d.Status == types.InProgress {
				result = append(result, *d)
				extraInProgress--
				continue
			}
			// check active interval
			if d.Status == types.InProgress && !q.ActiveInterval.IsTimeInInterval(time.Now()) {
				result = append(result, *d)
				extraInProgress--
				continue
			}
		}
	}

	State.mu.Unlock()

	return result
}

func findInProgressCandidates() []types.Download {
	State.mu.Lock()

	result := make([]types.Download, 0)
	for i := range State.Queues {
		q := State.Queues[i]
		inProgressCnt, _ := getInProgressDownloads(q)
		remainingInProgress := q.MaxInProgressCount - inProgressCnt
		for _, downloadId := range q.DownloadIds {
			d := State.Downloads[downloadId]
			now := time.Now()
			if remainingInProgress > 0 &&
				d.Status == types.Created &&
				q.ActiveInterval.IsTimeInInterval(now) {
				result = append(result, *d)
				remainingInProgress--
				continue
			}
			if remainingInProgress > 0 &&
				d.Status == types.Failed &&
				d.CurrentRetriesCnt < q.MaxRetriesCount &&
				q.ActiveInterval.IsTimeInInterval(now) {
				result = append(result, *d)
				remainingInProgress--
				continue
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
		if i%3 == 0 {
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
			slog.Info(fmt.Sprintf("Modify Event => %+v", data))
			if data.newMaxBandwidth != nil {
				State.Queues[data.queueId].MaxBandwidth = *data.newMaxBandwidth
				for _, downloadId := range State.Queues[data.queueId].DownloadIds {
					// TODO: the downloads may not be active!
					dm := State.downloadManagers[downloadId]
					if dm != nil {
						dm.EventsChan <- network.NewReconfigDMEvent(downloadId, *data.newMaxBandwidth)
					}
				}
			}
			if data.newActiveInterval != nil {
				State.Queues[data.queueId].ActiveInterval = *data.newActiveInterval
			}
			slog.Info("Modify Event End")
		case PauseDownloadEvent:
			slog.Info("Pausing the download ----------------------------------------------------------------")
			data := e.Data.(PauseDownloadEventData)
			id := data.DownloadID
			slog.Info(fmt.Sprintf("Download id is %v", id))
			updateDownloadStatus(id, types.Paused)
			State.downloadManagers[id].EventsChan <- network.DMEvent{EType: network.Pause}
		case ResumeDownloadEvent:
			data := e.Data.(ResumeDownloadEventData)
			id := data.DownloadID
			updateDownloadStatus(id, types.InProgress)
			fmt.Println("finished updating status")
			State.downloadManagers[id].EventsChan <- network.DMEvent{EType: network.Resume}

		case DeleteDownloadEvent:
		}
	}
	// 2. Fire up new candidates
	inProgressCandidates := findInProgressCandidates()
	//slog.Info(fmt.Sprintf("inProgressCandidates => %+v", inProgressCandidates))
	for _, d := range inProgressCandidates {
		updateDownloadStatus(d.Id, types.InProgress)
		createDownloadManager(d.Id)
		chIn, chOut := getChannel(d.Id)
		queue := getQueue(d.Id)
		go network.AsyncStartDownload(d, *queue, chIn, chOut)
		// because the exact order of changing states are important,
		// we cannot use the pull based approach.
		// However, again, DM will not get terminated until we send the terminate message for it.
		go func() {
			for responseEvent := range chOut {
				// Motherfucker! this log caused the concurrent access error.
				// (because the log uses the responseEvent in a async way, but it is also processed after ward!)
				// slog.Info(fmt.Sprintf("Response Event for %d => %+v", d.Id, responseEvent))
				switch responseEvent.EType {
				case network.Completed:
					slog.Debug(fmt.Sprintf("DMR : Completed %d", d.Id))
					updateDownloadStatus(d.Id, types.Completed)
					return
				case network.Failure:
					slog.Debug(fmt.Sprintf("DMR : Failure %d", d.Id))
					updateDownloadStatus(d.Id, types.Failed)
					return
				case network.InProgress:
					data := responseEvent.Data.(network.InProgressDMRData)
					slog.Debug(fmt.Sprintf("DMR : InProgress %d", d.Id))
					updateDMChunksByteOffset(d.Id, data.CurrentChunksByteOffset)
				}
			}
		}()
	}
	// 3. Reset candidates:
	//      The InProgress Downloads that don't abide the current configuration.
	//      The DownloadStatus will be changed to Created.
	resetCandidates := findResetCandidates()
	//	slog.Info(fmt.Sprintf("resetCaadindtes => %+v", resetCandidates))
	for _, d := range resetCandidates {
		updateDownloadStatus(d.Id, types.Created)
		// send message to stop downloading.
		chIn, _ := getChannel(d.Id)
		chIn <- network.NewTerminateDMEvent(d.Id)
	}
	// 4. Ask for updates from the DMs.
	//    Update the state accordingly.
	//    This is done in the DMWatcher
}

// TODO: This is unneccessery contention! Because we are storing the chans in a
//
//	map, then we have to lock.
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

func updateDownloadStatus(id int, newStatus types.DownloadStatus) {
	State.mu.Lock()
	oldStatus := State.Downloads[id].Status
	State.Downloads[id].Status = newStatus
	// 1: Update CurrentInProgressCount
	// if oldStatus == newStatus {
	// 	panic("old status same is new status??")
	// }
	if oldStatus == types.InProgress {
		State.Queues[State.Downloads[id].QueueId].CurrentInProgressCount--
	}
	if oldStatus == types.InProgress && newStatus == types.Paused {
		fmt.Println("ran here 3")
		State.Queues[State.Downloads[id].QueueId].CurrentInProgressCount--

	}
	if newStatus == types.InProgress {
		State.Queues[State.Downloads[id].QueueId].CurrentInProgressCount++
	}
	// 1: Update CurrentRetriesCnt
	if newStatus == types.Failed {
		State.Downloads[id].CurrentRetriesCnt++
	}

	State.mu.Unlock()
}

func updateDMChunksByteOffset(downloadId int, currentChunksByteOffset map[int]int) {
	State.mu.Lock()

	State.downloadManagers[downloadId].ChunksByteOffset = currentChunksByteOffset
	State.Downloads[downloadId].CurrnetDownloadOffsets = currentChunksByteOffset

	State.mu.Unlock()
}

func createDownloadManager(downloadId int) {
	State.mu.Lock()

	downloadManager := network.DownloadManager{
		EventsChan:        make(chan network.DMEvent),
		ResponseEventChan: make(chan network.DMREvent),
	}
	State.downloadManagers[downloadId] = &downloadManager
	spew.Dump(State.downloadManagers)

	State.mu.Unlock()
}
