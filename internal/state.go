package internal

type AppState struct {
	queues []Queue
}

var State *AppState

func InitState() {
	State = &AppState{
		queues: make([]Queue, 0, 10),
	}
}
