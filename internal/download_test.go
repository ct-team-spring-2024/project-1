package internal

import (
	"testing"
	"go-idm/types"
	// "fmt"
	// "log/slog"
)

func Test_Add(t *testing.T) {
	InitState()
	q := types.Queue{Id: 0}
	AddQueue(q)
	d := types.Download{
		Id:    0,
		Queue: q,
	}
	AddDownload(d)

	act := len(State.Queues[0].Downloads)
	exp := 1
	if exp != act {
		t.Fatalf("download length is not correct. exp = %d, act = %d", exp, act)
	}
	// State.queues = append(State.queues, )
}

func Test_Delete(t *testing.T) {
	InitState()
	q := types.Queue{Id: 0}
	AddQueue(q)
	d := types.Download{
		Id:    0,
		Queue: q,
	}
	AddDownload(d)

	dd := types.Download{
		Id: 0,
	}
	Delete(dd)
	act := len(State.Queues[0].Downloads)
	exp := 0
	if exp != act {
		t.Fatalf("download length is not correct. exp = %d, act = %d", exp, act)
	}
}
