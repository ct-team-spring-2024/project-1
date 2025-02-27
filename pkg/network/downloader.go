package network

import (
	"fmt"
	"go-idm/types"
	"io"
	"net/http"
	"os"
	"path/filePath"
)

type DownloadResult struct {
	IsDone bool
	Err    error
}

func SyncStartDownload(download types.Download) DownloadResult {
	url := download.Url
	abosolutePath := filepath.Join(download.Queue.Destination, download.Filename)
	out, err := os.Create(abosolutePath)
	if err != nil {
		return DownloadResult{false, fmt.Errorf("failed to create file: %w", err)}

	}
	defer out.Close()
	resp, err := http.Get(url)
	if err != nil {
		return DownloadResult{false, fmt.Errorf("failed to create file: %w", err)}
	}
	defer resp.Body.Close()

	// Check if the response status is OK
	if resp.StatusCode != http.StatusOK {
		return DownloadResult{false, fmt.Errorf("failed to create file: %w", err)}
	}

	// Stream the response body to the file
	_, err = io.Copy(out, resp.Body)
	if err != nil {
		return DownloadResult{false, fmt.Errorf("failed to create file: %w", err)}
	}

	return DownloadResult{true, nil}

}
