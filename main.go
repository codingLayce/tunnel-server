package main

import (
	"log/slog"

	"github.com/codingLayce/tunnel-server/cmd"
)

func main() {
	slog.SetLogLoggerLevel(slog.LevelInfo)
	cmd.Exec()
}
