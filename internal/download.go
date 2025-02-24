package internal

type DownloadStatus int
const (
	Completed DownloadStatus = iota
	Failed
	Paused
	InProgress
	Created
)

type Download struct {
	url string
	status DownloadStatus
	queue Queue
}
