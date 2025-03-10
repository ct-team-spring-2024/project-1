package internal

import (
	"go-idm/types"
)

// all function passes must have pointers , otherwise a copy of it will be appended
func AddQueue(queue *types.Queue) {
	State.Queues = append(State.Queues, queue)
}

func getInProgressDownloads(queue *types.Queue) (int, []types.Download) {
	cnt := 0
	result := make([]types.Download, 0)
	for _, d := range queue.Downloads {
		if d.Status == types.InProgress {
			cnt++
			result = append(result, *d)
		}
	}
	return cnt, result
}
