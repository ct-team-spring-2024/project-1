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

func checkToBeInProgress (d Download) bool {
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
		inProgressCnt, inProgressDownloads := getInProgressDownloads(q)
		result = append(result, inProgressDownloads...)
		remainingInProgress := q.maxInProgressCount - inProgressCnt
		for _, d := range q.downloads {
			if remainingInProgress == 0 {
				break
			}
		}

	}

}

func UpdateState() {
	var inProgressCandidates []Download
	inProgressCandidates = findInPrpgressCandidates(State)

}
