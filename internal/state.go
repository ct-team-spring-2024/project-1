package internal

type AppState struct {
	Queues []*Queue
}

var State *AppState

func InitState() {
	State = &AppState{
		Queues: make([]*Queue, 0, 10),
	}
}

func checkToBeInProgress(d Download) bool {
	if d.status == Created {
		return true
	}
	if d.status == Failed && d.currentRetriesCnt < d.queue.maxRetriesCount {
		return true
	}
	return false
}

func findInPrpgressCandidates(state *AppState) []Download {
	result := make([]Download, 0)
	for _, q := range state.Queues {
		inProgressCnt, inProgressDownloads := getInProgressDownloads(*q)
		result = append(result, inProgressDownloads...)
		remainingInProgress := q.maxInProgressCount - inProgressCnt
		for _, d := range q.downloads {
			if remainingInProgress == 0 {
				break
			}
			if checkToBeInProgress(d) {
				result = append(result, d)
				remainingInProgress--
			}
		}
	}
	return result

}

func UpdateState() {
	var inProgressCandidates []Download
	inProgressCandidates = findInPrpgressCandidates(State)
	for _, v := range inProgressCandidates {
		updateDownloadStatus(v.id, InProgress)
		//pass to network

	}

}
func updateDownloadStatus(id int, status DownloadStatus) {
	for _, q := range State.Queues {
		for _, d := range q.downloads {
			if d.id == id {
				d.status = status
				if status == InProgress {
					q.currentInProgressCount++
				}
			}
		}
	}
}
