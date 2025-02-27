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
	i, j := findDownload(download)
	newDownloads := append(State.Queues[i].Downloads[:j], State.Queues[i].Downloads[j+1:]...)
	State.Queues[i].Downloads = newDownloads
}

func pause(download types.Download) {
	//TODO : add a function to tell the Network layer to stop downloading
	i, j := findDownload(download)
	State.Queues[i].Downloads[j].Status = types.Failed
}

func resume(download types.Download) {
	i, j := findDownload(download)
	//Might need to add to downloads if status is not checked
	State.Queues[i].Downloads[j].Status = types.InProgress
}

func findDownload(download types.Download) (i, j int) {
	for k, _ := range State.Queues {
		if State.Queues[k].Id == download.Queue.Id {
			for m, _ := range State.Queues[i].Downloads {
				if State.Queues[k].Downloads[m].Id == download.Id {
					i, j = k, m

				}
			}
		}
	}
	return
}
