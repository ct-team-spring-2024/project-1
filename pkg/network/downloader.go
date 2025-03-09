package network

import (
	"bufio"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"sync"
	"test/types"
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
	finished
)

type CMEvent struct {
	event CEtpye
}

type CMREvent struct {
	responseEvent CMREtype
	id            int
}

func SyncStartDownload(download types.Download, inputChannel chan int, responseCh chan int) DownloadResult {
	url := download.Url
	absolutePath := filepath.Join(download.Queue.Destination, download.Filename)

	// Send a HEAD request to get the file size
	resp, err := http.Head(url)
	if err != nil {
		responseCh <- -1
		return DownloadResult{false, fmt.Errorf("HTTP HEAD failed: %w", err)}
	}
	defer resp.Body.Close()

	fileSize := resp.ContentLength
	if fileSize <= 0 {
		responseCh <- -1
		return DownloadResult{false, fmt.Errorf("invalid file size: %d", fileSize)}
	}

	// Check if the server supports range requests
	if resp.Header.Get("Accept-Ranges") != "bytes" {
		fmt.Println("Server does not support range requests. Downloading the entire file.")
		inputCh := make(chan int)
		return downloadEntireFile(url, absolutePath, &inputCh, responseCh)
	}

	// Server supports range requests, proceed with chunked download
	numChunks := 3
	chunkSize := fileSize / int64(numChunks)

	var wg sync.WaitGroup
	tempFiles := make([]*os.File, numChunks)
	tempFilePaths := make([]string, numChunks)

	inputChannels := make([]*chan CMEvent, 0)
	//respChannels := make([]*chan CMREvent, 0)
	// Download each chunk concurrently
	respCh := make(chan CMREvent)
	for i := 0; i < numChunks; i++ {
		start := int64(i) * chunkSize
		end := start + chunkSize - 1
		if i == numChunks-1 {
			end = fileSize - 1 // Last chunk gets the remaining bytes
		}

		inputCh := make(chan CMEvent)
		inputChannels = append(inputChannels, &inputCh)

		// Create a temporary file for this chunk
		tempFile, err := os.CreateTemp("", fmt.Sprintf("chunk-%d-", i))
		if err != nil {
			responseCh <- -1
			return DownloadResult{false, fmt.Errorf("failed to create temp file: %w", err)}
		}
		tempFiles[i] = tempFile
		tempFilePaths[i] = tempFile.Name()

		wg.Add(1)
		go downloadChunk(download.Url, start, end, tempFile, i, &wg, &inputCh, &respCh)

	}
	doneChannels := make([]bool, numChunks)
outerLoop:
	for {
		select {
		case <-inputChannel:
			//do stuff
		case v := <-respCh:
			i := v.id
			fmt.Printf("Chunk number %v is finished", i)
			doneChannels[i] = true

		default:
			time.Sleep(500 * time.Millisecond)
			fmt.Println("Woke up")
			done := true
			for i := 0; i < numChunks; i++ {
				if !doneChannels[i] {
					done = false
				}
			}
			if done {
				fmt.Print("Proccess finished")
				break outerLoop

			}

		}
	}
	// Wait for all chunks to finish downloading
	//wg.Wait()

	// Merge the temporary files into the final file
	fmt.Println("creating file")
	finalFile, err := os.Create(absolutePath)
	if err != nil {
		responseCh <- -1
		return DownloadResult{false, fmt.Errorf("failed to create final file: %w", err)}
	}
	defer finalFile.Close()
	fmt.Println("writing to final file")
	for _, tempFilePath := range tempFilePaths {
		tempFile, err := os.Open(tempFilePath)
		if err != nil {
			responseCh <- -1
			return DownloadResult{false, fmt.Errorf("failed to open temp file: %w", err)}
		}

		_, err = io.Copy(finalFile, tempFile)
		tempFile.Close()
		if err != nil {
			responseCh <- -1
			return DownloadResult{false, fmt.Errorf("failed to merge temp file: %w", err)}
		}

		// Remove the temporary file
		os.Remove(tempFilePath)
	}

	fmt.Println("Download completed!")
	return DownloadResult{true, nil}

}

func downloadChunk(url string, start, end int64, tempFile *os.File, chunkID int, wg *sync.WaitGroup, inputCh *chan CMEvent, responseCh *chan CMREvent) {
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

	// Use a buffered reader to read the response body
	reader := bufio.NewReader(resp.Body)
	buffer := make([]byte, 32*1024) // 32 KB buffer

	for {
		n, err := reader.Read(buffer)
		if err != nil && err != io.EOF {
			fmt.Printf("Error reading data: %v\n", err)
			return
		}

		if n > 0 {
			_, err := tempFile.Write(buffer[:n])
			fmt.Printf("Wrote %d bytes in chunk %v\n", n, chunkID)
			if err != nil {
				fmt.Printf("Error writing to temp file: %v\n", err)
				return
			}
		}

		if err == io.EOF {
			*responseCh <- CMREvent{id: chunkID}
			break
		}
	}

	fmt.Printf("Chunk %d downloaded successfully.\n", chunkID)
	return
}

func downloadEntireFile(url, filePath string, inputCh *chan int, responseCh chan int) DownloadResult {
	resp, err := http.Get(url)
	if err != nil {
		responseCh <- -1
		return DownloadResult{false, fmt.Errorf("HTTP GET failed: %w", err)}
	}
	defer resp.Body.Close()

	file, err := os.Create(filePath)
	if err != nil {
		responseCh <- -1
		return DownloadResult{false, fmt.Errorf("failed to create file: %w", err)}
	}
	defer file.Close()
	for {
		select {

		case <-*inputCh:
		//do stuff
		default:
			reader := bufio.NewReader(resp.Body)
			buffer := make([]byte, 32*1024) // 32 KB buffer

			n, err := reader.Read(buffer)
			if err != nil && err != io.EOF {
				fmt.Printf("Error reading data: %v\n", err)
				return DownloadResult{IsDone: false}
			}

			if n > 0 {
				_, err := file.Write(buffer[:n])
				if err != nil {
					fmt.Printf("Error writing to temp file: %v\n", err)
					return DownloadResult{IsDone: false}
				}
			}

			if err == io.EOF {
				break
			}

		}

		return DownloadResult{true, nil}
	}
}
