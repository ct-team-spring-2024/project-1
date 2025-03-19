package network

import (
	"bufio"
	"fmt"
	"go-idm/types"
	"io"
	"log/slog"
	"net/http"
	"sync"

	// "net/url"
	"os"
	"path/filepath"
	"time"
)

type DownloadManager struct {
	EventsChan        chan DMEvent
	ResponseEventChan chan DMREvent
	ChunksByteOffset  map[int]int
}

type EType int

const (
	Reconfig EType = iota
	Pause
	Resume
	Terminatee
)

type DMEvent struct {
	EType EType
	Data  interface{}
}

type ReconfigDMData struct {
	downloadId      int
	newMaxBandwidth int
}

type TerminateDMData struct {
	downloadId int
}

func NewReconfigDMEvent(downloadId int, newMaxBandwidth int) DMEvent {
	return DMEvent{
		EType: Reconfig,
		Data: ReconfigDMData{
			downloadId:      downloadId,
			newMaxBandwidth: newMaxBandwidth,
		},
	}
}

func NewTerminateDMEvent(downloadId int) DMEvent {
	return DMEvent{
		EType: Terminatee,
		Data: TerminateDMData{
			downloadId: downloadId,
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
	Data  interface{}
}

type CompletedDMRData struct {
	downloadId int
}
type FailureDMRData struct {
	downloadId int
}
type InProgressDMRData struct {
	downloadId              int
	CurrentChunksByteOffset map[int]int
}

func NewCompletedDMREvent(downloadId int) DMREvent {
	return DMREvent{
		EType: Completed,
		Data: CompletedDMRData{
			downloadId: downloadId,
		},
	}
}

func NewFailureDMREvent(downloadId int) DMREvent {
	return DMREvent{
		EType: Failure,
		Data: FailureDMRData{
			downloadId: downloadId,
		},
	}
}

func NewInProgressDMREvent(downloadId int, currentChunksByteOffset map[int]int) DMREvent {
	return DMREvent{
		EType: InProgress,
		Data: InProgressDMRData{
			downloadId:              downloadId,
			CurrentChunksByteOffset: currentChunksByteOffset,
		},
	}
}

type CMEType int

const (
	reconfig CMEType = iota
	pause
	resume
	getStatus
	terminate
)

type CMEvent struct {
	EType CMEType
	Data  interface{}
}

type ReconfigCMData struct {
	newMaxBandwidth int
}

type GetStatusCMData struct {
}

type TerminateCMData struct {
}

func NewReconfigCMEvent(newMaxBandwidth int) CMEvent {
	return CMEvent{
		EType: reconfig,
		Data: ReconfigCMData{
			newMaxBandwidth: newMaxBandwidth,
		},
	}
}

func NewGetStatusCMEvent() CMEvent {
	return CMEvent{
		EType: getStatus,
		Data:  GetStatusCMData{},
	}
}

func NewTerminateCMEvent() CMEvent {
	return CMEvent{
		EType: terminate,
		Data:  TerminateCMData{},
	}
}
func NewPauseCMEvent() CMEvent {
	return CMEvent{
		EType: pause,
	}
}
func NewResumeCMEvent() CMEvent {
	return CMEvent{
		EType: resume,
	}
}

type CMREType int

const (
	inProgress CMREType = iota
	stopped
	paused
	failed
	downloaded
)

type CMREvent struct {
	EType CMREType
	Data  interface{}
}

type InProgressCMRData struct {
	chunkId         int
	chunkByteOffset int
}

type StoppedCMRData struct {
	chunkId int
}

type FailedCMRData struct {
	chunkId int
}
type PausedCMRData struct {
	chunkId int
}
type DownloadedCMRData struct {
	chunkId int
}

func NewInProgressCMREvent(chunkId int, chunkByteOffset int) CMREvent {
	return CMREvent{
		EType: inProgress,
		Data: InProgressCMRData{
			chunkId:         chunkId,
			chunkByteOffset: chunkByteOffset,
		},
	}
}

func NewStoppedCMREvent(chunkId int) CMREvent {
	return CMREvent{
		EType: stopped,
		Data: StoppedCMRData{
			chunkId: chunkId,
		},
	}
}

func NewFailedCMREvent(chunkId int) CMREvent {
	return CMREvent{
		EType: failed,
		Data: FailedCMRData{
			chunkId: chunkId,
		},
	}
}
func NewPausedCMREvnet(chunkId int) CMREvent {
	return CMREvent{
		EType: paused,
		Data: PausedCMRData{
			chunkId: chunkId,
		},
	}
}

func NewDownloadedCMREvent(chunkId int) CMREvent {
	return CMREvent{
		EType: downloaded,
		Data: DownloadedCMRData{
			chunkId: chunkId,
		},
	}
}

// TODO: On newConfig, all the current chunk places are stored in state.
//
//	Then given the new download, the chnuks managers are recreated.
//
// TODO: Because InProgress are frequent, maybe using buffered channel would help.
func AsyncStartDownload(download types.Download, queue types.Queue, chIn <-chan DMEvent, chOut chan<- DMREvent) {
	DMStatus := "inProgress"
	var numChunks int
	var chInCM []chan CMEvent
	var chOutCM []chan CMREvent
	var tempFiles []*os.File
	var tempFilePaths []string
	var absolutePath string
	var chunksByteOffset map[int]int
	var doneChannels []bool
	var ticker *time.Ticker
	initFunc := func() {
		url := download.Url
		absolutePath = filepath.Join(queue.Destination, download.Filename)

		// Send a HEAD request to get the file size
		resp, err := http.Head(url)

		acceptsRanges := true
		headError := false
		if err != nil {
			headError = true
			slog.Error(fmt.Sprintf("HTTP HEAD failed: %v", err))
		}
		if headError || resp.Header.Get("Accept-Ranges") != "bytes" {
			fmt.Println("Server does not support range requests. Downloading the entire file.")
			acceptsRanges = false
		}
		fileSize := resp.ContentLength
		if fileSize <= 0 {
			DMStatus = "failed"
			slog.Error(fmt.Sprintf("invalid file size: %d", fileSize))
			return
		}

		numChunks = 3
		if !acceptsRanges {
			numChunks = 1
		}
		slog.Info(fmt.Sprintf("Num Chunks is => %d", numChunks))

		chunkSize := fileSize / int64(numChunks)
		rateLimit := int64(queue.MaxBandwidth / numChunks)
		chunksByteOffset = make(map[int]int)
		// TODO: chunksByteOffset should be given as a parameter to the function
		for i := 0; i < numChunks; i++ {
			chunksByteOffset[i] = 0
		}
		tempFiles = make([]*os.File, numChunks)
		tempFilePaths = make([]string, numChunks)

		// Starting Chunk Managers
		chInCM = make([]chan CMEvent, 0)
		chOutCM = make([]chan CMREvent, 0)

		for i := 0; i < numChunks; i++ {
			start := int64(i) * chunkSize
			end := start + chunkSize - 1
			if i == numChunks-1 {
				end = fileSize - 1 // Last chunk gets the remaining bytes
			}

			inputCh := make(chan CMEvent)
			outputCh := make(chan CMREvent)
			chInCM = append(chInCM, inputCh)
			chOutCM = append(chOutCM, outputCh)

			// Create a temporary file for this chunk
			tempFile, err := os.CreateTemp("", fmt.Sprintf("chunk-%d-", i))
			if err != nil {
				DMStatus = "failed"
				slog.Error(fmt.Sprintf("failed to create temp file: %v", err))
				return
			}
			tempFiles[i] = tempFile
			tempFilePaths[i] = tempFile.Name()
			go downloadChunk(download.Url, start, end, acceptsRanges, tempFile, i, rateLimit, inputCh, outputCh)
		}
		doneChannels = make([]bool, numChunks)
		ticker = time.NewTicker(500 * time.Millisecond)
	}

	initFunc()
	if DMStatus == "failed" {
		chOut <- NewFailureDMREvent(download.Id)
		return
	}
	// Rule:
	// - Each CM is running. The state of CM can be fetched.
	//   If terminate message is sent, the coroutine will be freed.

	// Check CMs after each tick
	go func() {
		for DMStatus == "inProgress" {
			select {
			case <-ticker.C:
				fmt.Println("Woke up")
				// 1. Check the updates
				for i := 0; i < numChunks; i++ {
					select {
					case msg := <-chOutCM[i]:
						switch msg.EType {
						case inProgress:
							data := msg.Data.(InProgressCMRData)
							chunksByteOffset[data.chunkId] = data.chunkByteOffset
							select {
							case chOut <- NewInProgressDMREvent(download.Id, chunksByteOffset):
								slog.Info("updated chunk offsets")
							default:
								fmt.Print("dfkdfldkfdlfk")
							}
							slog.Debug(fmt.Sprintf("InProgress: chunkId=%d, chunkByteOffset=%d\n", data.chunkId, data.chunkByteOffset))

						case stopped:
							data := msg.Data.(StoppedCMRData)
							slog.Debug(fmt.Sprintf("Stopped: chunkId=%d\n", data.chunkId))
						case failed:
							data := msg.Data.(FailedCMRData)
							slog.Debug(fmt.Sprintf("Failed: chunkId=%d\n", data.chunkId))
						case downloaded:
							data := msg.Data.(DownloadedCMRData)
							slog.Debug(fmt.Sprintf("Downloaded: chunkId=%d\n", data.chunkId))
							doneChannels[data.chunkId] = true
						default:
							slog.Debug("Unknown event type")
						}
					default:
						// No message available on the channel
						fmt.Printf("No message available on chOutCM[%d]\n", i)
					}
				}
				// 2: Send get status + check if completed
				done := true
				for i := 0; i < numChunks; i++ {
					if !doneChannels[i] {
						done = false
						//too many goroutines , this is unneccesary
						go func() {
							chInCM[i] <- NewGetStatusCMEvent()
							return
						}()
					}
				}
				if done {
					DMStatus = "completed"
					createFinalFile(absolutePath, tempFilePaths, chOut)
					return
				}
			}
		}
	}()
	// Listen for incomming messages
	for {
		select {
		case event := <-chIn:
			switch event.EType {
			case Terminatee:
				for i := 0; i < numChunks; i++ {
					chInCM[i] <- NewTerminateCMEvent()
				}
				DMStatus = "terminated"
				return
			case Reconfig:
				data := event.Data.(ReconfigDMData)
				for i := 0; i < numChunks; i++ {
					chInCM[i] <- NewReconfigCMEvent(data.newMaxBandwidth / numChunks)
				}
			case Pause:
				for _, ch := range chInCM {
					slog.Info(fmt.Sprintf("Sent message to chunks"))
					ch <- NewPauseCMEvent()
					fmt.Println("Sent pause msg to chunks")
				}
			case Resume:
				for _, ch := range chInCM {
					slog.Info(fmt.Sprintf("Sent message to chunks"))
					ch <- NewResumeCMEvent()
					fmt.Println("Sent pause msg to chunks")
				}
			}
		}
	}
}

// TODO inputCh should be monitored
// TODO Why channel type is pointer????
func downloadChunk(url string, start, end int64, acceptsRanges bool, tempFile *os.File,
	chunkID int, rateLimit int64,
	chIn chan CMEvent, chOut chan CMREvent) {
	CMStatus := "inProgress"
	var tickerMu sync.Mutex
	var ticker *time.Ticker
	var reader *bufio.Reader
	var buffer []byte
	var totalBytesRead int
	var bufferSizeInBytes int64
	var resp *http.Response

	updateTicker := func(newRateLimit int64) {
		tickerMu.Lock()
		ticker.Stop() // Stop the existing ticker
		rateLimit = newRateLimit
		ticker = time.NewTicker(time.Second / time.Duration(rateLimit/bufferSizeInBytes))
		slog.Info(fmt.Sprintf("Updated ticker interval to %v bytes/sec", rateLimit))
		tickerMu.Unlock()
	}
	initFunc := func() {
		slog.Info(fmt.Sprintf("downloadChunk args %v %v", rateLimit, url))

		req, err := http.NewRequest("GET", url, nil)
		if err != nil {
			fmt.Printf("Error creating request for chunk %d-%d: %v\n", start, end, err)
			CMStatus = "failed"
			return
		}
		//If the server accpets ranges the request should have the bytes header
		if acceptsRanges {
			req.Header.Set("Range", fmt.Sprintf("bytes=%d-%d", start, end))
		}

		client := &http.Client{}
		resp, err = client.Do(req)
		if err != nil {
			fmt.Printf("Error downloading chunk %d-%d: %v\n", start, end, err)
			CMStatus = "failed"
			return
		}

		bufferSizeInBytes = int64(32 * 1024)
		reader = bufio.NewReader(resp.Body)
		buffer = make([]byte, bufferSizeInBytes) // 32 KB buffer
		totalBytesRead = 0

		ticker = time.NewTicker(time.Second / time.Duration(rateLimit/bufferSizeInBytes)) // Adjust based on buffer size
	}

	initFunc()

	// Downloading
	// go func() {
	//	for CMStatus == "inProgress" {
	//		//ticker is thread safe , lock unneccessary
	//		tickerMu.Lock()
	//		<-ticker.C
	//		tickerMu.Unlock()

	//		n, err := reader.Read(buffer)
	//		if err != nil && err != io.EOF {
	//			fmt.Printf("Error reading data: %v\n", err)
	//			CMStatus = "failed"
	//			return
	//		}

	//		if n > 0 {
	//			_, err := tempFile.Write(buffer[:n])
	//			if err != nil {
	//				fmt.Printf("Error writing to temp file: %v\n", err)
	//				CMStatus = "failed"
	//				return
	//			}
	//			slog.Debug(fmt.Sprintf("Wrote %d bytes in chunk %v\n", n, chunkID))

	//			totalBytesRead += n
	//		}

	//		if err == io.EOF {
	//			CMStatus = "downloaded"
	//			resp.Body.Close()
	//			break
	//		}
	//	}
	// }()

	// Listening for message
	for {
		select {
		case <-ticker.C:
			if CMStatus == "inProgress" {
				n, err := reader.Read(buffer)
				if err != nil && err != io.EOF {
					fmt.Printf("Error reading data: %v\n", err)
					CMStatus = "failed"
					return
				}

				if n > 0 {
					_, err := tempFile.Write(buffer[:n])
					if err != nil {
						fmt.Printf("Error writing to temp file: %v\n", err)
						CMStatus = "failed"
						return
					}
					slog.Debug(fmt.Sprintf("Wrote %d bytes in chunk %v\n", n, chunkID))

					totalBytesRead += n
				}

				if err == io.EOF {
					CMStatus = "downloaded"
					resp.Body.Close()
					break
				}
			}
		case e := <-chIn:
			switch e.EType {
			case reconfig:
				data, ok := e.Data.(ReconfigCMData)
				if !ok {
					slog.Warn("Invalid reconfig event data")
					continue
				}
				newRateLimit := int64(data.newMaxBandwidth)
				slog.Debug(fmt.Sprintf("Received reconfig event with new rate limit: %v bytes/sec", newRateLimit))
				updateTicker(newRateLimit)
			case pause:
				CMStatus = "paused"
			case resume:
				CMStatus = "inProgress"
			case getStatus:
				go func() {
					fmt.Println("sending status")
					switch CMStatus {
					case "inProgress":
						chOut <- NewInProgressCMREvent(chunkID, totalBytesRead)
					case "downloaded":
						chOut <- NewDownloadedCMREvent(chunkID)
					case "stopped":
						chOut <- NewStoppedCMREvent(chunkID)
					case "failed":
						chOut <- NewFailedCMREvent(chunkID)
					case "paused":
						chOut <- NewPausedCMREvnet(chunkID)
					}
				}()
			case terminate:
				CMStatus = "stopped"
				return
			}
		}
	}
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
	slog.Info(fmt.Sprintf("creating file %v\n", absolutePath))
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
