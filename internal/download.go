package internal

import (
	"fmt"
	"go-idm/types"
	// "log/slog"
	// "github.com/davecgh/go-spew/spew"
)

func AddDownload(download types.Download, queueId int) error {
	State.Queues[queueId].DownloadIds = append(State.Queues[queueId].DownloadIds, download.Id)
	State.Downloads[download.Id] = &download
	State.Downloads[download.Id].QueueId = queueId
	return nil
}

func Delete(id int) {
	queueId := State.Downloads[id].QueueId
	delete(State.Downloads, id)
	index := -1
	for i, downloadId := range State.Queues[queueId].DownloadIds {
		if downloadId == id {
			index = i
			break
		}
	}
	if index != -1 {
		State.Queues[queueId].DownloadIds = append(State.Queues[queueId].DownloadIds[:index], State.Queues[queueId].DownloadIds[index+1:]...)
	} else {
		fmt.Println("Value", queueId, "not found in the slice.")
	}
}

func Pause(id int) {
	State.Downloads[id].Status = types.Failed
}

func Resume(id int) {
	State.Downloads[id].Status = types.InProgress
}

func FindDownload2(id int) types.Download {
	return *State.Downloads[id]
}

func FindQueue(id int) (*types.Queue, error) {
	for i, _ := range State.Queues {
		if State.Queues[i].Id == id {
			return State.Queues[i], nil
		}
	}
	return nil, fmt.Errorf("queue with ID %d not found", id)
}
