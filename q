[1mdiff --git a/internal/state.go b/internal/state.go[m
[1mindex 8cec13d..6716d41 100644[m
[1m--- a/internal/state.go[m
[1m+++ b/internal/state.go[m
[36m@@ -95,15 +95,6 @@[m [mfunc updateState() {[m
 	spew.Dump(inProgressCandidates)[m
 	for _, d := range inProgressCandidates {[m
 		updateDownloadStatus(d.Id, types.InProgress)[m
[31m-		// pass to network[m
[31m-		// result := network.SyncStartDownload(v)[m
[31m-		// switch e := result.Err; {[m
[31m-		// case e == nil:[m
[31m-		//	updateDownloadStatus(v.Id, types.Completed)[m
[31m-		// default:[m
[31m-		//	updateDownloadStatus(v.Id, types.Failed)[m
[31m-		// }[m
[31m-		// create download manager[m
 		createDownloadManager(d.Id)[m
 		spew.Dump(State.downloadManagers)[m
 		go DownloadManagerHandler(d.Id, State.downloadManagers[d.Id].eventsChan, State.downloadManagers[d.Id].responseEventChan)[m
[1mdiff --git a/pkg/network/downloader.go b/pkg/network/downloader.go[m
[1mindex 73a444f..0244bca 100644[m
[1m--- a/pkg/network/downloader.go[m
[1m+++ b/pkg/network/downloader.go[m
[36m@@ -1,10 +1,10 @@[m
 package network[m
 [m
 import ([m
[32m+[m	[32m"bufio"[m
 	"fmt"[m
 	"go-idm/types"[m
 	"io"[m
[31m-	"log/slog"[m
 	"net/http"[m
 	"os"[m
 	"path/filepath"[m
[36m@@ -16,6 +16,7 @@[m [mtype DownloadResult struct {[m
 	IsDone bool[m
 	Err    error[m
 }[m
[32m+[m
 type CEtpye int[m
 type CMREtype int[m
 [m
[36m@@ -23,117 +24,146 @@[m [mconst ([m
 	start CEtpye = iota[m
 	finish[m
 )[m
[32m+[m
 const ([m
 	inProgress CMREtype = iota[m
 	paused[m
 	failed[m
[32m+[m	[32mfinished[m
 )[m
 [m
 type CMEvent struct {[m
 	event CEtpye[m
 }[m
[32m+[m
 type CMREvent struct {[m
 	responseEvent CMREtype[m
[32m+[m	[32mid            int[m
 }[m
 [m
[31m-func SyncStartDownload(download types.Download, inputCh chan int, responseCh chan int) DownloadResult {[m
[32m+[m[32mfunc SyncStartDownload(download types.Download, inputChannel chan int, responseCh chan int) DownloadResult {[m
 	url := download.Url[m
[31m-	abosolutePath := filepath.Join(download.Queue.Destination, download.Filename)[m
[31m-	file, err := os.Create(abosolutePath)[m
[31m-	if err != nil {[m
[31m-		responseCh <- -1[m
[31m-		return DownloadResult{false, fmt.Errorf("failed to create file: %w", err)}[m
[31m-	}[m
[31m-	defer file.Close()[m
[32m+[m	[32mabsolutePath := filepath.Join(download.Queue.Destination, download.Filename)[m
[32m+[m
[32m+[m	[32m// Send a HEAD request to get the file size[m
 	resp, err := http.Head(url)[m
 	if err != nil {[m
 		responseCh <- -1[m
[31m-		return DownloadResult{false, fmt.Errorf("HTTP Get failed: %w", err)}[m
[32m+[m		[32mreturn DownloadResult{false, fmt.Errorf("HTTP HEAD failed: %w", err)}[m
 	}[m
[31m-	//defer resp.Body.Close()[m
[31m-	fileSize := resp.ContentLength[m
[32m+[m	[32mdefer resp.Body.Close()[m
 [m
[31m-	// Check if the response status is OK[m
[31m-	if resp.StatusCode != http.StatusOK {[m
[32m+[m	[32mfileSize := resp.ContentLength[m
[32m+[m	[32mif fileSize <= 0 {[m
 		responseCh <- -1[m
[31m-		return DownloadResult{false, fmt.Errorf("http connection failed: %w", err)}[m
[32m+[m		[32mreturn DownloadResult{false, fmt.Errorf("invalid file size: %d", fileSize)}[m
 	}[m
 [m
[31m-	//Chunk set to 4 for now, change later[m
[31m-	if resp.StatusCode != http.StatusPartialContent {[m
[31m-		//server does not accept partial content[m
[31m-		req, err := http.NewRequest("GET", url, nil)[m
[31m-		if err != nil {[m
[31m-			return DownloadResult{false, fmt.Errorf("failed to create request: %w", err)}[m
[31m-		}[m
[31m-		client := &http.Client{}[m
[31m-		resp, err := client.Do(req)[m
[31m-		file, err := os.OpenFile(abosolutePath, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)[m
[31m-[m
[31m-		if err != nil {[m
[31m-			//do sth[m
[31m-		}[m
[31m-		for {[m
[31m-			select {[m
[31m-			case inputEvent, ok := <-inputCh:[m
[31m-				if !ok {[m
[31m-					slog.Info("channel closed , exiting")[m
[31m-				}[m
[31m-				fmt.Println("executing the order &v", inputEvent)[m
[31m-				file.Seek(download.CurrnetDownloadOffsets[0], 0)[m
[31m-			default:[m
[31m-				io.Copy(file, resp.Body)[m
[31m-				if err != nil {[m
[31m-					responseCh <- -1[m
[31m-					return DownloadResult{false, fmt.Errorf("failed to write file to path: %v", file)}[m
[31m-				}[m
[31m-				return DownloadResult{true, nil}[m
[31m-			}[m
[31m-		}[m
[31m-[m
[32m+[m	[32m// Check if the server supports range requests[m
[32m+[m	[32mif resp.Header.Get("Accept-Ranges") != "bytes" {[m
[32m+[m		[32mfmt.Println("Server does not support range requests. Downloading the entire file.")[m
[32m+[m		[32minputCh := make(chan int)[m
[32m+[m		[32mreturn downloadEntireFile(url, absolutePath, &inputCh, responseCh)[m
 	}[m
 [m
[31m-	numChunks := 4[m
[32m+[m	[32m// Server supports range requests, proceed with chunked download[m
[32m+[m	[32mnumChunks := 3[m
 	chunkSize := fileSize / int64(numChunks)[m
 [m
 	var wg sync.WaitGroup[m
[31m-	var mu sync.Mutex[m
[31m-	chunkInputChannels := make([]*chan CMEvent, 0)[m
[31m-	chunkResponseChannels := make([]*chan CMREvent, 0)[m
[31m-	for i := 0; i < int(numChunks); i++ {[m
[32m+[m	[32mtempFiles := make([]*os.File, numChunks)[m
[32m+[m	[32mtempFilePaths := make([]string, numChunks)[m
[32m+[m
[32m+[m	[32minputChannels := make([]*chan CMEvent, 0)[m
[32m+[m	[32m//respChannels := make([]*chan CMREvent, 0)[m
[32m+[m	[32m// Download each chunk concurrently[m
[32m+[m	[32mrespCh := make(chan CMREvent)[m
[32m+[m	[32mfor i := 0; i < numChunks; i++ {[m
 		start := int64(i) * chunkSize[m
 		end := start + chunkSize - 1[m
[31m-		if i == int(numChunks)-1 {[m
[32m+[m		[32mif i == numChunks-1 {[m
 			end = fileSize - 1 // Last chunk gets the remaining bytes[m
 		}[m
[31m-		download.CurrnetDownloadOffsets[i] = start[m
[31m-		chunkIn := make(chan CMEvent)[m
[31m-		chunkInputChannels = append(chunkInputChannels, &chunkIn)[m
[31m-		chunkRes := make(chan CMREvent)[m
[31m-		chunkResponseChannels = append(chunkResponseChannels, &chunkRes)[m
[32m+[m
[32m+[m		[32minputCh := make(chan CMEvent)[m
[32m+[m		[32minputChannels = append(inputChannels, &inputCh)[m
[32m+[m
[32m+[m		[32m// Create a temporary file for this chunk[m
[32m+[m		[32mtempFile, err := os.CreateTemp("", fmt.Sprintf("chunk-%d-", i))[m
[32m+[m		[32mif err != nil {[m
[32m+[m			[32mresponseCh <- -1[m
[32m+[m			[32mreturn DownloadResult{false, fmt.Errorf("failed to create temp file: %w", err)}[m
[32m+[m		[32m}[m
[32m+[m		[32mtempFiles[i] = tempFile[m
[32m+[m		[32mtempFilePaths[i] = tempFile.Name()[m
 [m
 		wg.Add(1)[m
[31m-		go downloadChunk(url, start, end, &wg, file, &mu, chunkIn, chunkRes, download, i)[m
[31m-	}[m
[32m+[m		[32mgo downloadChunk(download.Url, start, end, tempFile, i, &wg, &inputCh, &respCh)[m
 [m
[32m+[m	[32m}[m
[32m+[m	[32mdoneChannels := make([]bool, numChunks)[m
[32m+[m[32mouterLoop:[m
 	for {[m
 		select {[m
[31m-		case <-inputCh:[m
[31m-			fmt.Printf("executing the order that came from inputCh : %v", inputCh)[m
[32m+[m		[32mcase <-inputChannel:[m
[32m+[m			[32m//do stuff[m
[32m+[m		[32mcase v := <-respCh:[m
[32m+[m			[32mi := v.id[m
[32m+[m			[32mfmt.Printf("Chunk number %v is finished", i)[m
[32m+[m			[32mdoneChannels[i] = true[m
[32m+[m
 		default:[m
 			time.Sleep(500 * time.Millisecond)[m
[32m+[m			[32mfmt.Println("Woke up")[m
[32m+[m			[32mdone := true[m
[32m+[m			[32mfor i := 0; i < numChunks; i++ {[m
[32m+[m				[32mif !doneChannels[i] {[m
[32m+[m					[32mdone = false[m
[32m+[m				[32m}[m
[32m+[m			[32m}[m
[32m+[m			[32mif done {[m
[32m+[m				[32mfmt.Print("Proccess finished")[m
[32m+[m				[32mbreak outerLoop[m
[32m+[m
[32m+[m			[32m}[m
[32m+[m
 		}[m
 	}[m
[31m-[m
 	// Wait for all chunks to finish downloading[m
[31m-	wg.Wait()[m
[31m-	fmt.Println("Download completed!")[m
[32m+[m	[32m//wg.Wait()[m
[32m+[m
[32m+[m	[32m// Merge the temporary files into the final file[m
[32m+[m	[32mfmt.Println("creating file")[m
[32m+[m	[32mfinalFile, err := os.Create(absolutePath)[m
[32m+[m	[32mif err != nil {[m
[32m+[m		[32mresponseCh <- -1[m
[32m+[m		[32mreturn DownloadResult{false, fmt.Errorf("failed to create final file: %w", err)}[m
[32m+[m	[32m}[m
[32m+[m	[32mdefer finalFile.Close()[m
[32m+[m	[32mfmt.Println("writing to final file")[m
[32m+[m	[32mfor _, tempFilePath := range tempFilePaths {[m
[32m+[m		[32mtempFile, err := os.Open(tempFilePath)[m
[32m+[m		[32mif err != nil {[m
[32m+[m			[32mresponseCh <- -1[m
[32m+[m			[32mreturn DownloadResult{false, fmt.Errorf("failed to open temp file: %w", err)}[m
[32m+[m		[32m}[m
 [m
[32m+[m		[32m_, err = io.Copy(finalFile, tempFile)[m
[32m+[m		[32mtempFile.Close()[m
[32m+[m		[32mif err != nil {[m
[32m+[m			[32mresponseCh <- -1[m
[32m+[m			[32mreturn DownloadResult{false, fmt.Errorf("failed to merge temp file: %w", err)}[m
[32m+[m		[32m}[m
[32m+[m
[32m+[m		[32m// Remove the temporary file[m
[32m+[m		[32mos.Remove(tempFilePath)[m
[32m+[m	[32m}[m
[32m+[m
[32m+[m	[32mfmt.Println("Download completed!")[m
 	return DownloadResult{true, nil}[m
 }[m
[31m-func downloadChunk(url string, start, end int64, wg *sync.WaitGroup, file *os.File, mu *sync.Mutex, inputCh chan CMEvent, responseCh chan CMREvent, download types.Download, chunckId int) {[m
[31m-	defer wg.Done()[m
 [m
[32m+[m[32mfunc downloadChunk(url string, start, end int64, tempFile *os.File, chunkID int, wg *sync.WaitGroup, inputCh *chan CMEvent, responseCh *chan CMREvent) {[m
 	// Create a new HTTP request with a range header[m
 	req, err := http.NewRequest("GET", url, nil)[m
 	if err != nil {[m
[36m@@ -145,27 +175,85 @@[m [mfunc downloadChunk(url string, start, end int64, wg *sync.WaitGroup, file *os.Fi[m
 	// Send the request[m
 	client := &http.Client{}[m
 	resp, err := client.Do(req)[m
[31m-[m
 	if err != nil {[m
 		fmt.Printf("Error downloading chunk %d-%d: %v\n", start, end, err)[m
 		return[m
 	}[m
 	defer resp.Body.Close()[m
[32m+[m
[32m+[m	[32m// Use a buffered reader to read the response body[m
[32m+[m	[32mreader := bufio.NewReader(resp.Body)[m
[32m+[m	[32mbuffer := make([]byte, 32*1024) // 32 KB buffer[m
[32m+[m
 	for {[m
[31m-		select {[m
[31m-		case <-inputCh:[m
[31m-			fmt.Printf("executing the order came from : %v", inputCh)[m
[31m-			file.Seek(download.CurrnetDownloadOffsets[chunckId], 0)[m
[31m-		default:[m
[31m-			mu.Lock()[m
[31m-			_, err := io.Copy(file, resp.Body)[m
[32m+[m		[32mn, err := reader.Read(buffer)[m
[32m+[m		[32mif err != nil && err != io.EOF {[m
[32m+[m			[32mfmt.Printf("Error reading data: %v\n", err)[m
[32m+[m			[32mreturn[m
[32m+[m		[32m}[m
[32m+[m
[32m+[m		[32mif n > 0 {[m
[32m+[m			[32m_, err := tempFile.Write(buffer[:n])[m
[32m+[m			[32mfmt.Printf("Wrote %d bytes in chunk %v\n", n, chunkID)[m
 			if err != nil {[m
[31m-				fmt.Printf("Error writing chunk %d-%d: %v\n", start, end, err)[m
[32m+[m				[32mfmt.Printf("Error writing to temp file: %v\n", err)[m
 				return[m
 			}[m
[31m-			mu.Unlock()[m
[31m-			//remove if not neccessary[m
[31m-			time.Sleep(100 * time.Millisecond)[m
 		}[m
[32m+[m
[32m+[m		[32mif err == io.EOF {[m
[32m+[m			[32m*responseCh <- CMREvent{id: chunkID}[m
[32m+[m			[32mbreak[m
[32m+[m		[32m}[m
[32m+[m	[32m}[m
[32m+[m
[32m+[m	[32mfmt.Printf("Chunk %d downloaded successfully.\n", chunkID)[m
[32m+[m	[32mreturn[m
[32m+[m[32m}[m
[32m+[m
[32m+[m[32mfunc downloadEntireFile(url, filePath string, inputCh *chan int, responseCh chan int) DownloadResult {[m
[32m+[m	[32mresp, err := http.Get(url)[m
[32m+[m	[32mif err != nil {[m
[32m+[m		[32mresponseCh <- -1[m
[32m+[m		[32mreturn DownloadResult{false, fmt.Errorf("HTTP GET failed: %w", err)}[m
[32m+[m	[32m}[m
[32m+[m	[32mdefer resp.Body.Close()[m
[32m+[m
[32m+[m	[32mfile, err := os.Create(filePath)[m
[32m+[m	[32mif err != nil {[m
[32m+[m		[32mresponseCh <- -1[m
[32m+[m		[32mreturn DownloadResult{false, fmt.Errorf("failed to create file: %w", err)}[m
[32m+[m	[32m}[m
[32m+[m	[32mdefer file.Close()[m
[32m+[m	[32mfor {[m
[32m+[m		[32mselect {[m
[32m+[m
[32m+[m		[32mcase <-*inputCh:[m
[32m+[m		[32m//do stuff[m
[32m+[m		[32mdefault:[m
[32m+[m			[32mreader := bufio.NewReader(resp.Body)[m
[32m+[m			[32mbuffer := make([]byte, 32*1024) // 32 KB buffer[m
[32m+[m
[32m+[m			[32mn, err := reader.Read(buffer)[m
[32m+[m			[32mif err != nil && err != io.EOF {[m
[32m+[m				[32mfmt.Printf("Error reading data: %v\n", err)[m
[32m+[m				[32mreturn DownloadResult{IsDone: false}[m
[32m+[m			[32m}[m
[32m+[m
[32m+[m			[32mif n > 0 {[m
[32m+[m				[32m_, err := file.Write(buffer[:n])[m
[32m+[m				[32mif err != nil {[m
[32m+[m					[32mfmt.Printf("Error writing to temp file: %v\n", err)[m
[32m+[m					[32mreturn DownloadResult{IsDone: false}[m
[32m+[m				[32m}[m
[32m+[m			[32m}[m
[32m+[m
[32m+[m			[32mif err == io.EOF {[m
[32m+[m				[32mbreak[m
[32m+[m			[32m}[m
[32m+[m
[32m+[m		[32m}[m
[32m+[m
[32m+[m		[32mreturn DownloadResult{true, nil}[m
 	}[m
 }[m
