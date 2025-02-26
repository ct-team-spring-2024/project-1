package internal

import "reflect"

func delete(download Download) {
	//TODO: Delete the file from the system too .
	for i, _ := range State.queues {
		if reflect.DeepEqual(State.queues[i], download.queue) {
			for j, _ := range State.queues[i].downloads {
				if reflect.DeepEqual(State.queues[i].downloads[j], download) {
					newQueue := append(State.queues[i].downloads[:j], State.queues[i].downloads[j+1:]...)
					State.queues[i].downloads = newQueue
				}

			}
		}
	}
}
func pause(download Download) {
	//TODO : add a function to tell the Network layer to stop downloading
	i, j := findDownload(download)
	State.queues[i].downloads[j].status = Failed
}
func resume(download Download) {
	i, j := findDownload(download)
	//Might need to add to downloads if status is not checked
	State.queues[i].downloads[j].status = InProgress
}
func findDownload(download Download) (i, j int) {
	for k, _ := range State.queues {
		if reflect.DeepEqual(State.queues[k], download.queue) {
			for m, _ := range State.queues[i].downloads {
				if reflect.DeepEqual(State.queues[i].downloads[j], download) {
					i, j = k, m

				}
			}
		}
	}
	return
}
