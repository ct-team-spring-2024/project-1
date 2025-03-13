package internal

import (
	"go-idm/types"
)

// all function passes must have pointers , otherwise a copy of it will be appended
func AddQueue(queue types.Queue) {
	State.Queues[queue.Id] = &queue
}

func getInProgressDownloads(queue *types.Queue) (int, []types.Download) {
	cnt := 0
	result := make([]types.Download, 0)
	for _, downloadId := range queue.DownloadIds {
		if State.Downloads[downloadId].Status == types.InProgress {
			cnt++
			result = append(result, *State.Downloads[downloadId])
		}
	}
	return cnt, result
}
