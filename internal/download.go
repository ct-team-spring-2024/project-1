package internal

import (
	"fmt"

	"log/slog"
)

type DownloadStatus int

const (
	Completed DownloadStatus = iota
	Failed

	// paused by the user
	Paused

	// download is in progress
	InProgress

	// created, but downloading has not started
	Created
)

type Download struct {
	id     int
	url    string
	status DownloadStatus
	currentRetriesCnt int
	queue  Queue
}

func NewDownload(id int) Download {
	return Download{
		id: id,
	}
}

func AddDownload(download Download) {
	queueId := download.queue.id
	var foundQueue *Queue
	for _, queue := range State.Queues {
		if queue.id == queueId {
			foundQueue = queue
			break
		}
	}
	slog.Info(fmt.Sprintf("add download => %+v", foundQueue))
	foundQueue.downloads = append(foundQueue.downloads, download)
	slog.Info(fmt.Sprintf("add download => %+v", foundQueue))
}

func Delete(download Download) {
	//TODO: Delete the file from the system too .
	i, j := findDownload(download)
	newDownloads := append(State.Queues[i].downloads[:j], State.Queues[i].downloads[j+1:]...)
	State.Queues[i].downloads = newDownloads
}

func pause(download Download) {
	//TODO : add a function to tell the Network layer to stop downloading
	i, j := findDownload(download)
	State.Queues[i].downloads[j].status = Failed
}

func resume(download Download) {
	i, j := findDownload(download)
	//Might need to add to downloads if status is not checked
	State.Queues[i].downloads[j].status = InProgress
}

func findDownload(download Download) (i, j int) {
	for k, _ := range State.Queues {
		if State.Queues[k].id == download.queue.id {
			for m, _ := range State.Queues[i].downloads {
				if State.Queues[k].downloads[m].id == download.id {
					i, j = k, m

				}
			}
		}
	}
	return
}
