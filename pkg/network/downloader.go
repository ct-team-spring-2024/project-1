package network

import (
	"fmt"
	"go-idm/types"
	"io"
	"log/slog"
	"net/http"
	"os"
	"path/filepath"
	"sync"
	"time"
)

type DownloadResult struct {
	IsDone bool
	Err    error
}
type CEtpye int
type CMREtype int

const (
	start CEtpye = iota
	finish
)
const (
	inProgress CMREtype = iota
	paused
	failed
)

type CMEvent struct {
	event CEtpye
}
type CMREvent struct {
	responseEvent CMREtype
}

func SyncStartDownload(download types.Download, inputCh chan int, responseCh chan int) DownloadResult {
	url := download.Url
	abosolutePath := filepath.Join(download.Queue.Destination, download.Filename)
	file, err := os.Create(abosolutePath)
	if err != nil {
		responseCh <- -1
		return DownloadResult{false, fmt.Errorf("failed to create file: %w", err)}
	}
	defer file.Close()
	resp, err := http.Head(url)
	if err != nil {
		responseCh <- -1
		return DownloadResult{false, fmt.Errorf("HTTP Get failed: %w", err)}
	}
	//defer resp.Body.Close()
	fileSize := resp.ContentLength

	// Check if the response status is OK
	if resp.StatusCode != http.StatusOK {
		responseCh <- -1
		return DownloadResult{false, fmt.Errorf("http connection failed: %w", err)}
	}

	//Chunk set to 4 for now, change later
	if resp.StatusCode != http.StatusPartialContent {
		//server does not accept partial content
		req, err := http.NewRequest("GET", url, nil)
		if err != nil {
			return DownloadResult{false, fmt.Errorf("failed to create request: %w", err)}
		}
		client := &http.Client{}
		resp, err := client.Do(req)
		file, err := os.OpenFile(abosolutePath, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)

		if err != nil {
			//do sth
		}
		for {
			select {
			case inputEvent, ok := <-inputCh:
				if !ok {
					slog.Info("channel closed , exiting")
				}
				fmt.Println("executing the order &v", inputEvent)
				file.Seek(download.CurrnetDownloadOffsets[0], 0)
			default:
				io.Copy(file, resp.Body)
				if err != nil {
					responseCh <- -1
					return DownloadResult{false, fmt.Errorf("failed to write file to path: %v", file)}
				}
				return DownloadResult{true, nil}
			}
		}

	}

	numChunks := 4
	chunkSize := fileSize / int64(numChunks)

	var wg sync.WaitGroup
	var mu sync.Mutex
	chunkInputChannels := make([]*chan CMEvent, 0)
	chunkResponseChannels := make([]*chan CMREvent, 0)
	for i := 0; i < int(numChunks); i++ {
		start := int64(i) * chunkSize
		end := start + chunkSize - 1
		if i == int(numChunks)-1 {
			end = fileSize - 1 // Last chunk gets the remaining bytes
		}
		download.CurrnetDownloadOffsets[i] = start
		chunkIn := make(chan CMEvent)
		chunkInputChannels = append(chunkInputChannels, &chunkIn)
		chunkRes := make(chan CMREvent)
		chunkResponseChannels = append(chunkResponseChannels, &chunkRes)

		wg.Add(1)
		go downloadChunk(url, start, end, &wg, file, &mu, chunkIn, chunkRes, download, i)
	}

	for {
		select {
		case <-inputCh:
			fmt.Printf("executing the order that came from inputCh : %v", inputCh)
		default:
			time.Sleep(500 * time.Millisecond)
		}
	}

	// Wait for all chunks to finish downloading
	wg.Wait()
	fmt.Println("Download completed!")

	return DownloadResult{true, nil}
}
func downloadChunk(url string, start, end int64, wg *sync.WaitGroup, file *os.File, mu *sync.Mutex, inputCh chan CMEvent, responseCh chan CMREvent, download types.Download, chunckId int) {
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
	for {
		select {
		case <-inputCh:
			fmt.Printf("executing the order came from : %v", inputCh)
			file.Seek(download.CurrnetDownloadOffsets[chunckId], 0)
		default:
			mu.Lock()
			_, err := io.Copy(file, resp.Body)
			if err != nil {
				fmt.Printf("Error writing chunk %d-%d: %v\n", start, end, err)
				return
			}
			mu.Unlock()
			//remove if not neccessary
			time.Sleep(100 * time.Millisecond)
		}
	}
}
