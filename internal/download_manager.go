package internal

import (
	"fmt"
	"log/slog"
)

func DownloadManagerHandler(downloadId int, ch <-chan DMEvent) {
	slog.Info(fmt.Sprintf("DownloadManagerHandler for download %d started", downloadId))
	// Listen on the channel for incoming events
	for event := range ch {
		switch event.etype {
		case start:
			slog.Info(fmt.Sprintf("Download %d: Received START event\n", downloadId))
		case stop:
			slog.Info(fmt.Sprintf("Download %d: Received STOP event\n", downloadId))
		default:
			slog.Info(fmt.Sprintf("Download %d: Received UNKNOWN event\n", downloadId))
		}
	}
	slog.Info(fmt.Sprintf("DownloadManagerHandler for download %d stopped", downloadId))
}
