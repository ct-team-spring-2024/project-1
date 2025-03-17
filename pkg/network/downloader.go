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
	"time"
)

type EType int

const (
	Startt EType = iota
	Stopp
	Pause
	Resume
	Reconfig
)

type DMEvent struct {
	EType EType
	Data  interface{}
}

type ReconfigDMData struct {
	newMaxBandwidth int
}

func NewReconfigDMEvent(newMaxBandwidth int) DMEvent {
	return DMEvent{
		EType: Reconfig,
		Data: ReconfigDMData{
			newMaxBandwidth: newMaxBandwidth,
		},
	}
}

type REType int

const (
	Completed REType = iota
	Failure
	InProgress
)

type DMREvent struct {
	EType REType
	// only for inprogress
	CurrentChunksByteOffset map[int]int
}

type DownloadManager struct {
	EventsChan        chan DMEvent
	ResponseEventChan chan DMREvent
	ChunksByteOffset  map[int]int
}

type DownloadResult struct {
	IsDone bool
	Err    error
}

type CMEType int

const (
	start CMEType = iota
	finish
	pause
	resume
	reconfig
)

type CMEvent struct {
	EType CMEType
	Data  interface{}
}

type ReconfigCMData struct {
	newMaxBandwidth int
}

func NewReconfigCMEvent(newMaxBandwidth int) CMEvent {
	return CMEvent{
		EType: reconfig,
		Data: ReconfigCMData{
			newMaxBandwidth: newMaxBandwidth,
		},
	}
}

type CMREType int

type CMREvent struct {
	EType   CMREType
	chunkId int
	// Only for inProgress
	chunkByteOffset int
}

const (
	inProgress CMREType = iota
	paused
	failed
	finished
	terminated
)

// TODO: On newConfig, all the current chunk places are stored in state.
//
//	Then given the new download, the chnuks managers are recreated.
//
// TODO: Because InProgress are frequent, maybe using buffered channel would help.
func AsyncStartDownload(download *types.Download, queue *types.Queue, chIn <-chan DMEvent, chOut chan<- DMREvent) {
	url := download.Url

	absolutePath := filepath.Join(queue.Destination, download.Filename)

	// Send a HEAD request to get the file size
	resp, err := http.Head(url)
	// defer resp.Body.Close()

	headError := false
	if err != nil {
		headError = true
		slog.Error(fmt.Sprintf("HTTP HEAD failed: %v", err))
	}

	if headError || resp.Header.Get("Accept-Ranges") != "bytes" {
		fmt.Println("Server does not support range requests. Downloading the entire file.")
		download.ServerAcceptsRanges = false
		// TODO: chIn to the downloader?

		//	downloadEntireFile(url, absolutePath, chIn, chOut)
		//	return
	}

	fileSize := resp.ContentLength
	if fileSize <= 0 {
		chOut <- DMREvent{
			EType: Failure,
		}
		slog.Error(fmt.Sprintf("invalid file size: %d", fileSize))
		return
	}
	// Server supports range requests, proceed with chunked download

	numChunks := 3
	//If server did not accept ranges , just sets the chunks number to one , the rest is basically the same
	if !download.ServerAcceptsRanges {
		numChunks = 1
	}
	chunkSize := fileSize / int64(numChunks)
	rateLimit := int64(queue.MaxBandwidth / numChunks)
	chunksByteOffset := make(map[int]int)
	// TODO: chunksByteOffset should be given as a parameter to the function
	for i := 0; i < numChunks; i++ {
		chunksByteOffset[i] = 0
	}
	tempFiles := make([]*os.File, numChunks)
	tempFilePaths := make([]string, numChunks)

	// Starting Chunk Managers
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
				EType: Failure,
			}
			slog.Error(fmt.Sprintf("failed to create temp file: %v", err))
			return
		}
		tempFiles[i] = tempFile
		tempFilePaths[i] = tempFile.Name()

		go downloadChunk(download.Url, start, end, tempFile, i, &inputCh, &respCh, rateLimit)

	}
	doneChannels := make([]bool, numChunks)

	ticker := time.NewTicker(500 * time.Millisecond)
	defer ticker.Stop()
	go func() {
		for {

			select {
			// incoming event from caller

			case event := <-chIn:
				fmt.Println("Recieved Resume or Pause msg")
				switch event.EType {
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
				case Pause:
					for _, chInCM := range inputChannels {
						slog.Info(fmt.Sprintf("Sent message to chunks"))
						*chInCM <- CMEvent{
							EType: pause,
						}
					}
				case Resume:
					fmt.Println("recieved resume msg")
					for _, chInCM := range inputChannels {
						*chInCM <- CMEvent{
							EType: resume,
						}
					}
				case Reconfig:
					data := event.Data.(ReconfigDMData)
					for i := 0; i < numChunks; i++ {
						*inputChannels[i] <- NewReconfigCMEvent(data.newMaxBandwidth)
					}
				}
			// incoming event from Chunk Managers
			case cmrevent := <-respCh:
				switch cmrevent.EType {
				case inProgress:
					chunksByteOffset[cmrevent.chunkId] = cmrevent.chunkByteOffset
					chOut <- DMREvent{
						EType:                   InProgress,
						CurrentChunksByteOffset: chunksByteOffset,
					}
				case failed:
					// stoping all chunks
					for _, chInCM := range inputChannels {
						*chInCM <- CMEvent{
							EType: finish,
						}
					}
					waitForTerminatedEvents(respCh, numChunks, 5*time.Second)
					chOut <- DMREvent{
						EType: Failure,
					}
					slog.Error(fmt.Sprintf("Failure in chunk managers"))
					return
				case finished:
					doneChannels[cmrevent.chunkId] = true
					fmt.Printf("Chunk number %v is finished\n", cmrevent.chunkId)
				default:
					slog.Error(fmt.Sprintf("unhandled type => %v", cmrevent))
				}
			case <-ticker.C:
				// This block runs every 500 milliseconds
				fmt.Println("Woke up")

				done := true
				for i := 0; i < numChunks; i++ {
					if !doneChannels[i] {
						done = false
						break
					}
				}
				if done {
					fmt.Println("Process finished")
					createFinalFile(absolutePath, tempFilePaths, chOut)
					return
				}
				//See if default not existing causes problems
				//	default:

			}
		}
	}()
}

// TODO inputCh should be monitored
// TODO Why channel type is pointer????
func downloadChunk(url string, start, end int64, tempFile *os.File,
	chunkID int, inputCh *chan CMEvent, responseCh *chan CMREvent,
	rateLimit int64) {
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
	bufferSizeInBytes := int64(32 * 1024)
	reader := bufio.NewReader(resp.Body)
	buffer := make([]byte, bufferSizeInBytes) // 32 KB buffer
	totalBytesRead := 0

	// Set up a ticker for rate limiting
	ticker := time.NewTicker(time.Second / time.Duration(rateLimit/bufferSizeInBytes)) // Adjust based on buffer size
	defer ticker.Stop()
	updateTicker := func(newRateLimit int64) {
		ticker.Stop() // Stop the existing ticker

		rateLimit = newRateLimit
		ticker = time.NewTicker(time.Second / time.Duration(rateLimit/bufferSizeInBytes))
		slog.Debug(fmt.Sprintf("Updated ticker interval to %v bytes/sec", rateLimit))
	}
	//Not needed , do it in the same goroutine
	//Goroutine to handle reconfiguration events
	// go func() {
	// 	for event := range *inputCh {

	// 	}
	// }()
	var isPaused = false
outerLoop:
	for {
		select {
		case event := <-*inputCh:
			if event.EType == pause {
				//wait for the next event to come
				slog.Info("Stopped the chunk downloader")
				isPaused = true

			}
			if event.EType == resume {
				slog.Info("Pausing chunk")
				isPaused = false
			}
			if event.EType == reconfig {
				data, ok := event.Data.(ReconfigCMData)
				if !ok {
					slog.Warn("Invalid reconfig event data")
					continue
				}
				newRateLimit := int64(data.newMaxBandwidth)
				slog.Debug(fmt.Sprintf("Received reconfig event with new rate limit: %v bytes/sec", newRateLimit))
				updateTicker(newRateLimit)
			}

		case <-ticker.C:
			if !isPaused {
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
					slog.Debug(fmt.Sprintf("Wrote %d bytes in chunk %v\n", n, chunkID))

					totalBytesRead += n
					// TODO if it blocks, then the download speed will be affected!
					*responseCh <- CMREvent{
						EType:           inProgress,
						chunkId:         chunkID,
						chunkByteOffset: totalBytesRead, // OK??
					}
				}

				if err == io.EOF {
					*responseCh <- CMREvent{EType: finished, chunkId: chunkID}
					break outerLoop
				}
			}
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
			EType: Failure,
		}
		slog.Error(fmt.Sprintf("HTTP GET failed: %v", err))
		return
	}
	defer resp.Body.Close()

	file, err := os.Create(filePath)
	if err != nil {
		chOut <- DMREvent{
			EType: Failure,
		}
		slog.Error(fmt.Sprintf("failed to create file: %v", err))
		return
	}
	defer file.Close()
	reader := bufio.NewReader(resp.Body)
	buffer := make([]byte, 32*1024) // 32 KB buffer
	for {
		select {
		case <-chIn:
		//do stuff
		default:

			n, err := reader.Read(buffer)
			if err != nil && err != io.EOF {
				chOut <- DMREvent{
					EType: Failure,
				}
				slog.Error(fmt.Sprintf("Error reading data: %v\n", err))
				return
			}

			if n > 0 {
				_, err := file.Write(buffer[:n])
				if err != nil {
					chOut <- DMREvent{
						EType: Failure,
					}
					slog.Error(fmt.Sprintf("Error writing to temp file: %v\n", err))
					return
				}
			}

			if err == io.EOF {
				return
			}

		}
	}
}

func createFinalFile(absolutePath string, tempFilePaths []string, chOut chan<- DMREvent) {
	// Merge the temporary files into the final file
	fmt.Printf("creating file %v\n", absolutePath)
	finalFile, err := os.Create(absolutePath)
	if err != nil {
		chOut <- DMREvent{
			EType: Failure,
		}
		slog.Error(fmt.Sprintf("failed to create final file: %v %v", absolutePath, tempFilePaths))
		slog.Error(fmt.Sprintf("failed to create final file: %v", err))
		return
	}
	defer finalFile.Close()
	fmt.Println("writing to final file")
	for _, tempFilePath := range tempFilePaths {
		tempFile, err := os.Open(tempFilePath)
		if err != nil {
			chOut <- DMREvent{
				EType: Failure,
			}
			slog.Error(fmt.Sprintf("failed to open temp file: %v", err))
			return
		}

		_, err = io.Copy(finalFile, tempFile)
		tempFile.Close()
		if err != nil {
			chOut <- DMREvent{
				EType: Failure,
			}
			slog.Error(fmt.Sprintf("failed to merge temp file: %v", err))
			return
		}
		// Remove the temporary file
		os.Remove(tempFilePath)
	}
	fmt.Println("Download completed!")
	chOut <- DMREvent{
		EType: Completed,
	}
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
				fmt.Printf("Received terminated event #%d (ID: %d)\n", count, event.chunkId)
				if count >= numOfChunks {
					return nil
				}
			}
		case <-timer.C:
			return fmt.Errorf("timeout waiting for terminated events")
		}
	}
}
