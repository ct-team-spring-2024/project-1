package types

import (
	"time"
)

type TimeInterval struct {
	start time.Time
	end   time.Time
}
type Queue struct {
	Id                     int
	DownloadIds            []int
	MaxInProgressCount     int
	CurrentInProgressCount int
	MaxRetriesCount        int
	Destination            string
	ActiveInterval         TimeInterval
	MaxBandwidth           int // In Byte
}

func NewQueue(id int) Queue {
	return Queue{
		Id:        id,
		DownloadIds: make([]int, 0),
	}
}
