package internal

import (
	"encoding/json"
	"fmt"
	"io"
	"log/slog"
	"os"
)

var Encoder *json.Encoder
var fileDir string

func CreateEncoder(file io.Writer, address string) *json.Encoder {
	fileDir = address
	encoder := json.NewEncoder(file)
	encoder.SetIndent("", "  ") // Pretty-print JSON
	Encoder = encoder
	return encoder
}
func SaveState() {
	State.mu.Lock()
	err := os.Truncate(fileDir, 0)
	if err != nil {
		slog.Error(fmt.Sprintf("Error truncating state file: %s", err))
	}

	err = Encoder.Encode(State)
	if err != nil {
		slog.Error(fmt.Sprintf("Error encoding state: %s", err))
	}
	State.mu.Unlock()
}
func LoadState(stateAddress string) (*AppState, error) {
	var state *AppState
	file, err := os.Open(stateAddress)
	if err != nil {
		slog.Info(fmt.Sprintf("Error opening the file : %v", err))
		return state, err
	}
	defer file.Close()
	decoder := json.NewDecoder(file)
	if err := decoder.Decode(&state); err != nil {
		fmt.Println("Error decoding JSON:", err)
		return state, err
	}

	return state, nil
}

// TODO : I think the state should be locked for this , but locking caused errors before
func SaveFile() {
	//State.mu.Lock()
	//defer State.mu.Unlock()
	jsonData, err := json.MarshalIndent(State, "", "  ")
	if err != nil {
		fmt.Println("Error encoding JSON:", err)
		return
	}

	// Overwrite JSON file
	err = os.WriteFile("data.json", jsonData, 0644)
	if err != nil {
		fmt.Println("Error writing file:", err)
		return
	}

	fmt.Println("JSON file overwritten successfully!")
	//State.mu.Unlock()
}
