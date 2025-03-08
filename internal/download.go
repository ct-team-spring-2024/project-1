package internal

import (
	"fmt"
	"go-idm/types"
)

func AddDownload(download types.Download) error {
	queueId := download.Queue.Id
	var foundQueue *types.Queue
	for _, queue := range State.Queues {
		if queue.Id == queueId {
			foundQueue = queue
			break
		}
	}
	if (foundQueue == nil) {
		return fmt.Errorf("queue with ID %d not found", queueId)
	}
	foundQueue.Downloads = append(foundQueue.Downloads, &download)
	return nil
}

func Delete(download types.Download) {
	//TODO: Delete the file from the system too .
	i, j := FindDownload(download.Id)
	newDownloads := append(State.Queues[i].Downloads[:j], State.Queues[i].Downloads[j+1:]...)
	State.Queues[i].Downloads = newDownloads
}

func pause(download types.Download) {
	//TODO : add a function to tell the Network layer to stop downloading
	i, j := FindDownload(download.Id)
	State.Queues[i].Downloads[j].Status = types.Failed
}

func resume(download types.Download) {
	i, j := FindDownload(download.Id)
	//Might need to add to downloads if status is not checked
	State.Queues[i].Downloads[j].Status = types.InProgress
}


func FindDownload(id int) (i, j int) {
	for k := range State.Queues {
		for m := range State.Queues[k].Downloads {
			if State.Queues[k].Downloads[m].Id == id {
				return k, m
			}
		}
	}
	return -1, -1 // if not found
}

func FindDownload2(id int) types.Download {
	for k := range State.Queues {
		for m := range State.Queues[k].Downloads {
			if State.Queues[k].Downloads[m].Id == id {
				return *State.Queues[k].Downloads[m]
			}
		}
	}
	return types.Download{}
}
