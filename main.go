package main

import (
	"log/slog"

	"tunnel-server/cmd"
)

func main() {
	slog.SetLogLoggerLevel(slog.LevelDebug)
	cmd.Exec()
}
