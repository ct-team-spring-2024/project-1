package types

type DownloadStatus int

const (
	Completed DownloadStatus = iota
	Failed

	// paused by the user
	Paused

	// download is in progress
	InProgress

	// created, but downloading has not started
	Created
)

type Download struct {
	Id                     int
	Url                    string
	Filename               string
	Status                 DownloadStatus
	CurrentRetriesCnt      int
	QueueId                int
	CurrnetDownloadOffsets map[int]int
	TempFileAddresses      map[int]string
}

func NewDownload(id int, q Queue) Download {
	return Download{
		Id:      id,
		QueueId: q.Id,
		Status:  Created,
	}
}
