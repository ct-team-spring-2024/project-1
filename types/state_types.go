package types

type AppState struct {
	Queues []*Queue
}

var State *AppState

func InitState() {
	State = &AppState{
		Queues: make([]*Queue, 0, 10),
	}
}
