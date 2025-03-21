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

type DownloadTicker struct {
	Ticker *time.Ticker
	//	TickerMu sync.Mutex
}

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
	downloadId int
}

type TerminateDMData struct {
	downloadId int
}

func NewReconfigDMEvent(downloadId int) DMEvent {
	return DMEvent{
		EType: Reconfig,
		Data: ReconfigDMData{
			downloadId: downloadId,
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
	SetTempFileAddress
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
type SetTempFileAddressDMRData struct {
	downloadId        int
	TempFileAddresses map[int]string
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
func NewSetTempFileAddressDMREvent(downloadId int, tempFileAddress map[int]string) DMREvent {
	return DMREvent{
		EType: SetTempFileAddress,
		Data: SetTempFileAddressDMRData{
			downloadId:        downloadId,
			TempFileAddresses: tempFileAddress,
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
}

type GetStatusCMData struct {
}

type TerminateCMData struct {
}

func NewReconfigCMEvent() CMEvent {
	return CMEvent{
		EType: reconfig,
		Data:  ReconfigCMData{},
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

var BufferSizeInBytes int64 = int64(32 * 1024)

func AsyncStartDownload(download types.Download, queue types.Queue,
	downloadTicker *DownloadTicker,
	chIn <-chan DMEvent, chOut chan<- DMREvent, beginFromFirst bool) {
	DMStatus := "inProgress"
	var numChunks int
	var chInCM []chan CMEvent
	var chOutCM []chan CMREvent
	var tempFiles []*os.File
	var tempFilePaths map[int]string
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
		var fileSize int64
		if err != nil {
			fileSize = 0
			headError = true
			slog.Error(fmt.Sprintf("HTTP HEAD failed: %v", err))
		} else {
			fileSize = resp.ContentLength
		}
		if fileSize <= 0 {
			DMStatus = "failed"
			slog.Error(fmt.Sprintf("invalid file size: %d", fileSize))
			return
		}
		if headError || resp.Header.Get("Accept-Ranges") != "bytes" {
			fmt.Println("Server does not support range requests.")
			acceptsRanges = false
		}

		numChunks = 3
		if !acceptsRanges {
			numChunks = 1
		}
		slog.Info(fmt.Sprintf("Num Chunks is => %d", numChunks))

		chunkSize := fileSize / int64(numChunks)
		chunksByteOffset = make(map[int]int)
		// TODO: chunksByteOffset should be given as a parameter to the function
		for i := 0; i < numChunks; i++ {
			chunksByteOffset[i] = 0
		}
		tempFiles = make([]*os.File, numChunks)
		tempFilePaths = make(map[int]string)

		// Starting Chunk Managers
		chInCM = make([]chan CMEvent, 0)
		chOutCM = make([]chan CMREvent, 0)
		//TODO : this should be changes to something else , checking the first offset is not quite right.
		if acceptsRanges && download.CurrnetDownloadOffsets[0] != 0 && !beginFromFirst {
			for i := 0; i < numChunks; i++ {

				index := int64(i) * chunkSize
				end := index + chunkSize - 1
				start := int64(download.CurrnetDownloadOffsets[i]) + index
				if i == numChunks-1 {
					end = fileSize - 1 // Last chunk gets the remaining bytes
				}

				inputCh := make(chan CMEvent)
				outputCh := make(chan CMREvent)
				chInCM = append(chInCM, inputCh)
				chOutCM = append(chOutCM, outputCh)

				// Create a temporary file for this chunk
				tempFile, err := os.OpenFile(download.TempFileAddresses[i], os.O_RDWR, 0666)
				//	tempFile, err := os.Open(download.TempFileAddresses[i])
				if err != nil {
					DMStatus = "failed"
					slog.Error(fmt.Sprintf("failed to open temp file: %v", err))
					return
				}
				tempFiles[i] = tempFile
				tempFilePaths[i] = tempFile.Name()
				go downloadChunk(download.Url, start, end, download.CurrnetDownloadOffsets[i], acceptsRanges, tempFile, i, downloadTicker, inputCh, outputCh)
			}
			doneChannels = make([]bool, numChunks)
			ticker = time.NewTicker(500 * time.Millisecond)

		} else {
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
				//	filePath := filepath.Join("C:/Users/Asus/Documents/GitHub/project-1/temp", fmt.Sprintf("chunk-%d-", i))
				//	tempFile, err := os.OpenFile(filePath, os.O_CREATE|os.O_WRONLY, 0666)
				//	tempFile, err := os.Create(filePath)
				tempFile, err := os.CreateTemp("", fmt.Sprintf("chunk-%d-", i))
				if err != nil {
					DMStatus = "failed"
					slog.Error(fmt.Sprintf("failed to create temp file: %v", err))
					return
				}
				tempFiles[i] = tempFile
				tempFilePaths[i] = tempFile.Name()
				go downloadChunk(download.Url, start, end, 0, acceptsRanges, tempFile, i, downloadTicker, inputCh, outputCh)
			}
			doneChannels = make([]bool, numChunks)
			ticker = time.NewTicker(500 * time.Millisecond)
		}

	}

	initFunc()
	chOut <- NewSetTempFileAddressDMREvent(download.Id, tempFilePaths)
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
							chOut <- NewInProgressDMREvent(download.Id, chunksByteOffset)
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
	for i, _ := range tempFiles {
		defer tempFiles[i].Close()
	}
	for {
		select {
		case event := <-chIn:
			switch event.EType {
			case Terminatee:
				for i := 0; i < numChunks; i++ {
					chInCM[i] <- NewTerminateCMEvent()
				}
				DMStatus = "terminated"
				fmt.Println("terminated")
				return
			case Reconfig:
				// data := event.Data.(ReconfigDMData)
				for i := 0; i < numChunks; i++ {
					chInCM[i] <- NewReconfigCMEvent()
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
					fmt.Println("Sent resume msg to chunks")
				}
			}
		}
	}
}

// States:
// inProgress
// paused
// failed
// downloaded
// terminated

// When terminated, we can shut down everything and don't listen to chIn

// When downloaded or terminated, the Download thread can be stopped.
// When inProgress, Download loop should continue.
// When paused or failed, wait until the status is changed and download can be continued.
func downloadChunk(url string, start, end int64, lastTimeIndex int, acceptsRanges bool, tempFile *os.File,
	chunkID int, downloadTicker *DownloadTicker,
	chIn chan CMEvent, chOut chan CMREvent) {
	CMStatus := "inProgress"
	var reader *bufio.Reader
	var buffer []byte
	var totalBytesRead int
	var resp *http.Response

	// updateTicker := func(newRateLimit int64) {
	//	tickerMu.Lock()
	//	ticker.Stop() // Stop the existing ticker
	//	rateLimit = newRateLimit
	//	ticker = time.NewTicker(time.Second / time.Duration(rateLimit/BufferSizeInBytes))
	//	slog.Info(fmt.Sprintf("Updated ticker interval to %v bytes/sec", rateLimit))
	//	tickerMu.Unlock()
	// }
	initFunc := func() {

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

		reader = bufio.NewReader(resp.Body)
		buffer = make([]byte, BufferSizeInBytes) // 32 KB buffer
		totalBytesRead = 0
	}

	initFunc()
	go func() {
		for CMStatus != "downloaded" && CMStatus != "terminated" {
			for CMStatus != "inProgress" {
				time.Sleep(1 * time.Second)
			}
			if CMStatus == "downloaded" || CMStatus == "terminated" {
				return
			}
			defer resp.Body.Close()

			// TODO: ticker is thread safe , lock unneccessary
			//	downloadTicker.TickerMu.Lock()
			<-downloadTicker.Ticker.C

			//downloadTicker.TickerMu.Unlock()

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
	}()

	// Listening for message
	for {
		select {
		case e := <-chIn:
			switch e.EType {
			case reconfig:
				// data, ok := e.Data.(ReconfigCMData)
				// if !ok {
				//	slog.Warn("Invalid reconfig event data")
				//	continue
				// }
				// slog.Debug(fmt.Sprintf("Received reconfig event with new rate limit: %v bytes/sec", newRateLimit))
				// updateTicker(newRateLimit)
			case pause:
				CMStatus = "paused"
			case resume:
				CMStatus = "inProgress"
			case getStatus:
				go func() {
					switch CMStatus {
					case "inProgress":
						chOut <- NewInProgressCMREvent(chunkID, totalBytesRead+lastTimeIndex)
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

func createFinalFile(absolutePath string, tempFilePaths map[int]string, chOut chan<- DMREvent) {
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
