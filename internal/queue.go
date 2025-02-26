package internal

import (
	"time"
)

type TimeInterval struct {
	start time.Time
	end time.Time
}
type Queue struct {
	id int
	downloads []Download
	maxSpeed float32     // In Byte
	maxDownloadCount int
	destination string
	activeInterval TimeInterval
	maxBandwidth float32 // In Byte
}

func NewQueue(id int) Queue{
	return Queue{
		id: 0,
	}
}

func AddQueue(queue Queue) {
	State.Queues = append(State.Queues, &queue)
}
