package internal

import (
	"fmt"
	"go-idm/types"
	"log/slog"
)

func AddDownload(download types.Download) {
	queueId := download.Queue.Id
	var foundQueue *types.Queue
	for _, queue := range State.Queues {
		if queue.Id == queueId {
			foundQueue = queue
			break
		}
	}
	slog.Info(fmt.Sprintf("add download => %+v", foundQueue))
	foundQueue.Downloads = append(foundQueue.Downloads, download)
	slog.Info(fmt.Sprintf("add download => %+v", foundQueue))
}

func Delete(download types.Download) {
	//TODO: Delete the file from the system too .
	i, j := findDownload(download.Id)
	newDownloads := append(State.Queues[i].Downloads[:j], State.Queues[i].Downloads[j+1:]...)
	State.Queues[i].Downloads = newDownloads
}

func pause(download types.Download) {
	//TODO : add a function to tell the Network layer to stop downloading
	i, j := findDownload(download.Id)
	State.Queues[i].Downloads[j].Status = types.Failed
}

func resume(download types.Download) {
	i, j := findDownload(download.Id)
	//Might need to add to downloads if status is not checked
	State.Queues[i].Downloads[j].Status = types.InProgress
}

func findDownload(id int) (i, j int) {
	for k, _ := range State.Queues {
		for m, _ := range State.Queues[i].Downloads {
			if State.Queues[k].Downloads[m].Id == id {
				return k, m
			}
		}

	}
	return -1, -1 // if not found
}
