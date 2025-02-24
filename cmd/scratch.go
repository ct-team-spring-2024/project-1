package main

import (
	"fmt"
	"go-idm/internal"
	"log/slog"
)

func main() {
	internal.InitState()
	slog.Info(fmt.Sprintf("state => %v", internal.State))
}
