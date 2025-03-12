package internal

import (
	"fmt"
	"go-idm/types"
)

func AddDownload(download *types.Download, QueueId int) error {
	queue, err := FindQueue(QueueId)
	if err != nil {
		return err
	}
	queue.Downloads[download.Id] = download
	download.QueueId = QueueId

	return nil
}

func Delete(id int) {
	i, _ := FindDownload(id)

	delete(State.Queues[i].Downloads, id)
}

func pause(id int) {
	i, j := FindDownload(id)
	State.Queues[i].Downloads[j].Status = types.Failed
}

func resume(id int) {
	i, j := FindDownload(id)
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
	return -1, -1
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
func FindQueue(id int) (*types.Queue, error) {
	for i, _ := range State.Queues {
		if State.Queues[i].Id == id {
			return State.Queues[i], nil
		}
	}
	return nil, fmt.Errorf("queue with ID %d not found", id)
}
