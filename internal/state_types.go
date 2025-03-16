package internal

type IDMEType int
const (
	AddQueueEvent IDMEType = iota
	ModifyQueueEvent
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
}

type ModifyQueueEventData struct {
	queueId         int
	newMaxBandwidth int
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

func NewAddQueueEvent() IDMEvent {
    return IDMEvent{
	EType: AddQueueEvent,
	Data: AddQueueEventData{
	},
    }
}

func NewModifyQueueEvent(queueId int, newMaxBandwidth int) IDMEvent {
    return IDMEvent{
	EType: ModifyQueueEvent,
	Data: ModifyQueueEventData{
		queueId: queueId,
		newMaxBandwidth: newMaxBandwidth,
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
