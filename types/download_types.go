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
	Id                int
	Url               string
	Filename          string
	Status            DownloadStatus
	CurrentRetriesCnt int
	Queue             Queue
}

func NewDownload(id int) Download {
	return Download{
		Id: id,
	}
}
