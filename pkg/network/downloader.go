package network

import (
	"bufio"
	"fmt"
	"go-idm/types"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"sync"
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

func SyncStartDownload(download types.Download, inputCh chan int, responseCh chan int) DownloadResult {
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
		return downloadEntireFile(url, absolutePath, responseCh)
	}

	// Server supports range requests, proceed with chunked download
	numChunks := 3
	chunkSize := fileSize / int64(numChunks)

	var wg sync.WaitGroup
	tempFiles := make([]*os.File, numChunks)
	tempFilePaths := make([]string, numChunks)

	// Download each chunk concurrently
	for i := 0; i < numChunks; i++ {
		start := int64(i) * chunkSize
		end := start + chunkSize - 1
		if i == numChunks-1 {
			end = fileSize - 1 // Last chunk gets the remaining bytes
		}

		// Create a temporary file for this chunk
		tempFile, err := os.CreateTemp("", fmt.Sprintf("chunk-%d-", i))
		if err != nil {
			responseCh <- -1
			return DownloadResult{false, fmt.Errorf("failed to create temp file: %w", err)}
		}
		tempFiles[i] = tempFile
		tempFilePaths[i] = tempFile.Name()

		wg.Add(1)
		go func(start, end int64, chunkID int, tempFile *os.File) {
			defer wg.Done()
			downloadChunk(url, start, end, tempFile, chunkID)
		}(start, end, i, tempFile)
	}

	// Wait for all chunks to finish downloading
	wg.Wait()

	// Merge the temporary files into the final file
	finalFile, err := os.Create(absolutePath)
	if err != nil {
		responseCh <- -1
		return DownloadResult{false, fmt.Errorf("failed to create final file: %w", err)}
	}
	defer finalFile.Close()

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

func downloadChunk(url string, start, end int64, tempFile *os.File, chunkID int) {
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
			if err != nil {
				fmt.Printf("Error writing to temp file: %v\n", err)
				return
			}
		}

		if err == io.EOF {
			break
		}
	}

	fmt.Printf("Chunk %d downloaded successfully.\n", chunkID)
}

func downloadEntireFile(url, filePath string, responseCh chan int) DownloadResult {
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

	_, err = io.Copy(file, resp.Body)
	if err != nil {
		responseCh <- -1
		return DownloadResult{false, fmt.Errorf("failed to write file: %w", err)}
	}

	return DownloadResult{true, nil}
}
