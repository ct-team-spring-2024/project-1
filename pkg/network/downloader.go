package network

import (
	"fmt"
	"go-idm/types"
	"io"
	"net/http"
	"os"
	"path/filePath"
	"sync"
)

type DownloadResult struct {
	IsDone bool
	Err    error
}

func SyncStartDownload(download types.Download) DownloadResult {
	url := download.Url
	abosolutePath := filepath.Join(download.Queue.Destination, download.Filename)
	file, err := os.Create(abosolutePath)
	if err != nil {
		return DownloadResult{false, fmt.Errorf("failed to create file: %w", err)}

	}
	defer file.Close()
	resp, err := http.Get(url)
	if err != nil {
		return DownloadResult{false, fmt.Errorf("failed to create file: %w", err)}
	}
	defer resp.Body.Close()
	fileSize := resp.ContentLength

	// Check if the response status is OK
	if resp.StatusCode != http.StatusOK {
		return DownloadResult{false, fmt.Errorf("http connection failed: %w", err)}
	}

	// Stream the response body to the file
	_, err = io.Copy(file, resp.Body)
	if err != nil {
		return DownloadResult{false, fmt.Errorf("failed to write file to path: %v", file)}
	}
	defer file.Close()
	//Chunk set to 4 for now, change later
	numChunks := 4
	chunkSize := fileSize / int64(numChunks)

	var wg sync.WaitGroup
	var mu sync.Mutex

	// Download each chunk concurrently
	for i := 0; i < int(numChunks); i++ {
		start := int64(i) * chunkSize
		end := start + chunkSize - 1
		if i == int(numChunks)-1 {
			end = fileSize - 1 // Last chunk gets the remaining bytes
		}

		wg.Add(1)
		go downloadChunk(url, start, end, &wg, file, &mu)
	}

	// Wait for all chunks to finish downloading
	wg.Wait()
	fmt.Println("Download completed!")

	return DownloadResult{true, nil}

}
func downloadChunk(url string, start, end int64, wg *sync.WaitGroup, file *os.File, mu *sync.Mutex) {
	defer wg.Done()

	// Create a new HTTP request with a range header
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		fmt.Printf("Error creating request for chunk %d-%d: %v\n", start, end, err)
		return
	}
	req.Header.Set("Range", fmt.Sprintf("bytes=%d-%d", start, end))

	// Send the request
	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		fmt.Printf("Error downloading chunk %d-%d: %v\n", start, end, err)
		return
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusPartialContent {
		fmt.Printf("Server does not support partial content for chunk %d-%d: %s\n", start, end, resp.Status)
		return
	}

	// Read the chunk data
	chunkData, err := io.ReadAll(resp.Body)
	if err != nil {
		fmt.Printf("Error reading chunk %d-%d: %v\n", start, end, err)
		return
	}

	// Write the chunk data to the correct position in the file
	mu.Lock()
	_, err = file.WriteAt(chunkData, start)
	mu.Unlock()
	if err != nil {
		fmt.Printf("Error writing chunk %d-%d: %v\n", start, end, err)
		return
	}
	fmt.Printf("downloading chunk %d-%d\n", start, end)

	fmt.Printf("Downloaded chunk %d-%d\n", start, end)
}
