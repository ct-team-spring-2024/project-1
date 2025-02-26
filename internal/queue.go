package internal

import (
	"time"
)

type TimeInterval struct {
	start time.Time
	end time.Time
}
type Queue struct {
	downloads []Download
	maxSpeed float32     // In Byte
	maxDownloadCount int
	destination string
	activeInterval TimeInterval
	maxBandwidth float32 // In Byte
}
