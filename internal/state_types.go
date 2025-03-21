package internal

import (
	"github.com/google/uuid"
	"go-idm/types"
)

type IDMEType int

const (
	AddQueueEvent IDMEType = iota
	ModifyQueueEvent
	AddDownloadEvent
	PauseDownloadEvent
	ResumeDownloadEvent
	RetryDownloadEvent
	DeleteDownloadEvent
)

type IDMEvent struct {
	EType IDMEType
	Data  interface{}
}

type AddQueueEventData struct {
	QueueId            int
	MaxInProgressCount int
	MaxRetriesCount    int
	Destination        string
	ActiveInterval     types.TimeInterval
	MaxBandwidth       int64
}

type ModifyQueueEventData struct {
	queueId             int
	newMaxBandwidth     *int64
	newActiveInterval   *types.TimeInterval
	newQueueDestination string
	newMaxinProgressCnt int
}

type AddDownloadEventData struct {
	Id       int
	Url      string
	Filename string
	QueueId  int
}

type PauseDownloadEventData struct {
	DownloadID int
}

type ResumeDownloadEventData struct {
	DownloadID int
}

type DeleteDownloadEventData struct {
	DownloadID int
}

func generateUniqueId() int {
	uuidValue := uuid.New()
	return int(uuidValue.ID())
}

func NewAddQueueEvent(
	id *int,
	maxInProgressCount int,
	maxRetriesCount int,
	destination string,
	activeInterval types.TimeInterval,
	maxBandwidth int64,
) IDMEvent {
	var finalId int
	if id == nil {
		finalId = generateUniqueId()
	} else {
		finalId = *id
	}

	return IDMEvent{
		EType: AddQueueEvent,
		Data: AddQueueEventData{
			QueueId:            finalId,
			MaxInProgressCount: maxInProgressCount,
			MaxRetriesCount:    maxRetriesCount,
			Destination:        destination,
			ActiveInterval:     activeInterval,
			MaxBandwidth:       maxBandwidth,
		},
	}
}

func NewModifyQueueEvent(queueId int, newMaxBandwidth *int64, newActiveInterval *types.TimeInterval, newQueueDestination string, maxInProgressCount int) IDMEvent {
	return IDMEvent{
		EType: ModifyQueueEvent,
		Data: ModifyQueueEventData{
			queueId:             queueId,
			newMaxBandwidth:     newMaxBandwidth,
			newActiveInterval:   newActiveInterval,
			newQueueDestination: newQueueDestination,
			newMaxinProgressCnt: maxInProgressCount,
		},
	}
}

func NewAddDownloadEvent(
	id *int,
	url string,
	filename string,
	queueId int,
) IDMEvent {
	var finalId int
	if id == nil {
		finalId = generateUniqueId()
	} else {
		finalId = *id
	}

	return IDMEvent{
		EType: AddDownloadEvent,
		Data: AddDownloadEventData{
			Id:       finalId,
			Url:      url,
			Filename: filename,
			QueueId:  queueId,
		},
	}
}

func NewPauseDownloadEvent(downloadID int) IDMEvent {
	return IDMEvent{
		EType: PauseDownloadEvent,
		Data: PauseDownloadEventData{
			DownloadID: downloadID,
		},
	}
}

func NewResumeDownloadEvent(downloadID int) IDMEvent {
	return IDMEvent{
		EType: ResumeDownloadEvent,
		Data: ResumeDownloadEventData{
			DownloadID: downloadID,
		},
	}
}

func NewDeleteDownloadEvent(downloadID int) IDMEvent {
	return IDMEvent{
		EType: DeleteDownloadEvent,
		Data: DeleteDownloadEventData{
			DownloadID: downloadID,
		},
	}
}
