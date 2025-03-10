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
	Downloads              map[int]*Download
	MaxInProgressCount     int
	CurrentInProgressCount int
	MaxRetriesCount        int
	Destination            string
	ActiveInterval         TimeInterval
	MaxBandwidth           float32 // In Byte
}

func NewQueue(id int) Queue {
	return Queue{
		Id:        id,
		Downloads: make(map[int]*Download),
	}
}
