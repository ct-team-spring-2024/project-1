package internal

import (
	"go-idm/types"
	"testing"
	// "fmt"
	// "log/slog"

	// "github.com/davecgh/go-spew/spew"
)

func Test_Add(t *testing.T) {
	InitState()
	q := types.NewQueue(0)
	AddQueue(q)
	d := types.Download{
		Id:      0,
		QueueId: q.Id,
	}
	AddDownload(d, d.QueueId)

	act := len(State.Queues[0].DownloadIds)
	exp := 1
	if exp != act {
		t.Fatalf("download length is not correct. exp = %d, act = %d", exp, act)
	}
}

func Test_Delete(t *testing.T) {
	InitState()
	q := types.NewQueue(0)
	AddQueue(q)
	d := types.Download{
		Id:      0,
		QueueId: q.Id,
	}
	AddDownload(d, d.QueueId)

	dd := types.Download{
		Id: 0,
	}
	Delete(dd.Id)
	act := len(State.Queues[0].DownloadIds)
	exp := 0
	if exp != act {
		t.Fatalf("download length is not correct. exp = %d, act = %d", exp, act)
	}
}
