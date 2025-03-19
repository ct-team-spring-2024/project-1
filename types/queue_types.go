package types

import (
	"time"
)

type TimeInterval struct {
	Start time.Time
	End   time.Time
}
func (ti TimeInterval) IsTimeInInterval(t time.Time) bool {
	return !t.Before(ti.Start) && !t.After(ti.End)
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
