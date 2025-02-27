package internal

import (
	"time"
)

type TimeInterval struct {
	start time.Time
	end   time.Time
}
type Queue struct {
	id                     int
	downloads              []Download
	maxInProgressCount     int
	currentInProgressCount int
	maxRetriesCount        int
	destination            string
	activeInterval         TimeInterval
	maxBandwidth           float32 // In Byte
}

func NewQueue(id int) Queue {
	return Queue{
		id: 0,
	}
}

func AddQueue(queue Queue) {
	State.Queues = append(State.Queues, &queue)
}

func getInProgressDownloads(queue Queue) (int, []Download) {
	cnt := 0
	result := make([]Download, 0)
	for _, d := range queue.downloads {
		if d.status == InProgress {
			cnt++
			result = append(result, d)
		}
	}
	return cnt, result
}
