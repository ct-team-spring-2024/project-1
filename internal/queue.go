package internal

import (
	"go-idm/types"
	"go-idm/pkg/network"
	"time"
)

func AddQueue(queue types.Queue) {
	State.Queues[queue.Id] = &queue

	State.downloadTickers[queue.Id] = &network.DownloadTicker{
		Ticker: time.NewTicker(time.Second / time.Duration(queue.MaxBandwidth/network.BufferSizeInBytes)),
	}
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
