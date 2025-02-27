package main

import (
	"go-idm/internal"
	"time"
)


func main() {
	for {
		// Some Queue/Download are added
		internal.UpdateState()
		time.Sleep(1 * time.Second)
	}
}
