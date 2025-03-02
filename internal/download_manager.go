package internal

import (
	"fmt"
	"log/slog"
)

func DownloadManagerHandler(downloadId int, ch chan DMEvent) {
	slog.Info(fmt.Sprintf("DownloadManagerHandler for download %d started", downloadId))
}
