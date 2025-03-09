package network

import (
	"bufio"
	"fmt"
	"go-idm/types"
	"io"
	"log/slog"
	"net/http"
	// "net/url"
	"os"
	"path/filepath"
	"sync"
	"time"
)

type EType int

const (
	Startt EType = iota
	Stopp
)

type REType int

const (
	Completed REType = iota
	Failure
)

type DMEvent struct {
	Etype EType
	// TODO Add data fields for the event
}

type DMREvent struct {
	Etype REType
}

type DownloadManager struct {
	EventsChan        chan DMEvent
	ResponseEventChan chan DMREvent
}

type DownloadResult struct {
	IsDone bool
	Err    error
}

type CMEtype int
type CMREtype int

const (
	start CMEtype = iota
	finish
)

const (
	inProgress CMREtype = iota
	paused
	failed
	finished
	terminated
)

type CMEvent struct {
	EType CMEtype
}

type CMREvent struct {
	EType CMREtype
	id    int
}

// TODO: On newConfig, all the current chunk places are stored in state.
//	Then give the new download, the chnuks managers are recreated.
func AsyncStartDownload(download types.Download, chIn <-chan DMEvent, chOut chan<- DMREvent) {
	url := download.Url
	absolutePath := filepath.Join(download.Queue.Destination, download.Filename)

	// Send a HEAD request to get the file size
	resp, err := http.Head(url)
	// defer resp.Body.Close()

	headError := false
	slog.Info(fmt.Sprintf("ggg => %v %v", resp, err))
	if err != nil {
		headError = true
		slog.Error(fmt.Sprintf("HTTP HEAD failed: %v", err))
	}
	if headError || resp.Header.Get("Accept-Ranges") != "bytes" {
		fmt.Println("Server does not support range requests. Downloading the entire file.")
		downloadEntireFile(url, absolutePath, chIn, chOut)
		return
	}

	fileSize := resp.ContentLength
	if fileSize <= 0 {
		chOut <- DMREvent{
			Etype: Failure,
		}
		slog.Error(fmt.Sprintf("invalid file size: %d", fileSize))
		return
	}




	// Server supports range requests, proceed with chunked download
	numChunks := 3
	chunkSize := fileSize / int64(numChunks)

	var wg sync.WaitGroup
	tempFiles := make([]*os.File, numChunks)
	tempFilePaths := make([]string, numChunks)

	inputChannels := make([]*chan CMEvent, 0)
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
			chOut <- DMREvent{
				Etype: Failure,
			}
			slog.Error(fmt.Sprintf("failed to create temp file: %v", err))
			return
		}
		tempFiles[i] = tempFile
		tempFilePaths[i] = tempFile.Name()

		wg.Add(1)
		go downloadChunk(download.Url, start, end, tempFile, i, &wg, &inputCh, &respCh)

	}
	doneChannels := make([]bool, numChunks)
	for {
		select {
		case event := <-chIn:
			// For each of the possible inputs, sent correct messages to the
			// chunkManagers.
			switch event.Etype {
			case Startt:
				// NOTHING
			case Stopp:
				// send stop to all chunk managers
				for _, chInCM := range inputChannels {
					*chInCM <- CMEvent{
						EType: finish,
					}
				}
				// TODO: should we wait for stop successfully message from them?
				waitForTerminatedEvents(respCh, numChunks, 5*time.Second)
			}
		case cmrevent := <-respCh:
			switch cmrevent.EType {
			case failed:
				// stoping all chunks
				for _, chInCM := range inputChannels {
					*chInCM <- CMEvent{
						EType: finish,
					}
				}
				waitForTerminatedEvents(respCh, numChunks, 5*time.Second)
				chOut <- DMREvent{
					Etype: Failure,
				}
				slog.Error(fmt.Sprintf("Failure in chunk managers"))
				return
			case finished:
				doneChannels[cmrevent.id] = true
				fmt.Printf("Chunk number %v is finished", cmrevent.id)
			default:
				slog.Error(fmt.Sprintf("unhandled type => %v", cmrevent))
			}
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
				createFinalFile(absolutePath, tempFilePaths, chOut)
				return
			}
		}
	}
}

// TODO inputCh should be monitored
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
			slog.Debug(fmt.Sprintf("Wrote %d bytes in chunk %v\n", n, chunkID))
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

// TODO: buggy now! dosen't cancel sync download in case of new chIn message
// TODO sent completed on the chOut
func downloadEntireFile(rawurl, filePath string, chIn <-chan DMEvent, chOut chan<- DMREvent) {
	// parsedUrl, err := url.Parse(rawurl)
	resp, err := http.Get(rawurl)
	slog.Info(fmt.Sprintf("raw => %s", rawurl))
	if err != nil {
		chOut <- DMREvent{
			Etype: Failure,
		}
		slog.Error(fmt.Sprintf("HTTP GET failed: %v", err))
		return
	}
	defer resp.Body.Close()

	file, err := os.Create(filePath)
	if err != nil {
		chOut <- DMREvent{
			Etype: Failure,
		}
		slog.Error(fmt.Sprintf("failed to create file: %v", err))
		return
	}
	defer file.Close()
	for {
		select {
		case <-chIn:
		//do stuff
		default:
			reader := bufio.NewReader(resp.Body)
			buffer := make([]byte, 32*1024) // 32 KB buffer

			n, err := reader.Read(buffer)
			if err != nil && err != io.EOF {
				chOut <- DMREvent{
					Etype: Failure,
				}
				slog.Error(fmt.Sprintf("Error reading data: %v\n", err))
				return
			}

			if n > 0 {
				_, err := file.Write(buffer[:n])
				if err != nil {
					chOut <- DMREvent{
						Etype: Failure,
					}
					slog.Error(fmt.Sprintf("Error writing to temp file: %v\n", err))
					return
				}
			}

			if err == io.EOF {
				break
			}

		}
	}
}

func createFinalFile(absolutePath string, tempFilePaths []string, chOut chan<- DMREvent) {
	// Merge the temporary files into the final file
	fmt.Println("creating file")
	finalFile, err := os.Create(absolutePath)
	if err != nil {
		chOut <- DMREvent{
			Etype: Failure,
		}
		slog.Error(fmt.Sprintf("failed to create final file: %v", err))
		return
	}
	defer finalFile.Close()
	fmt.Println("writing to final file")
	for _, tempFilePath := range tempFilePaths {
		tempFile, err := os.Open(tempFilePath)
		if err != nil {
			chOut <- DMREvent{
				Etype: Failure,
			}
			slog.Error(fmt.Sprintf("failed to open temp file: %v", err))
			return
		}

		_, err = io.Copy(finalFile, tempFile)
		tempFile.Close()
		if err != nil {
			chOut <- DMREvent{
				Etype: Failure,
			}
			slog.Error(fmt.Sprintf("failed to merge temp file: %v", err))
			return
		}
		// Remove the temporary file
		os.Remove(tempFilePath)
	}
	fmt.Println("Download completed!")
}

func waitForTerminatedEvents(respCh <-chan CMREvent, numOfChunks int, timeout time.Duration) error {
	count := 0
	timer := time.NewTimer(timeout)
	defer timer.Stop()

	for {
		select {
		case event, ok := <-respCh:
			if !ok {
				return fmt.Errorf("channel closed before receiving required events")
			}
			if event.EType == terminated {
				count++
				fmt.Printf("Received terminated event #%d (ID: %d)\n", count, event.id)
				if count >= numOfChunks {
					return nil
				}
			}
		case <-timer.C:
			return fmt.Errorf("timeout waiting for terminated events")
		}
	}
}
