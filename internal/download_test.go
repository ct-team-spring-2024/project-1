package internal

import (
	"testing"
	// "fmt"
	// "log/slog"
)

func Test_Add(t *testing.T) {
	InitState()
	q := Queue{id: 0}
	AddQueue(q)
	d := Download{
		id:    0,
		queue: q,
	}
	AddDownload(d)

	act := len(State.Queues[0].downloads)
	exp := 1
	if exp != act {
		t.Fatalf("download length is not correct. exp = %d, act = %d", exp, act)
	}
	// State.queues = append(State.queues, )
}

func Test_Delete(t *testing.T) {
	InitState()
	q := Queue{id: 0}
	AddQueue(q)
	d := Download{
		id:    0,
		queue: q,
	}
	AddDownload(d)

	dd := Download{
		id: 0,
	}
	Delete(dd)
	act := len(State.Queues[0].downloads)
	exp := 0
	if exp != act {
		t.Fatalf("download length is not correct. exp = %d, act = %d", exp, act)
	}
}
