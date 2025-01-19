package cmd

import (
	"log/slog"
	"os"
	"os/signal"

	"github.com/spf13/cobra"

	"github.com/codingLayce/tunnel-server/server"
)

var RootCmd = &cobra.Command{
	Short: "Start a Tunnel server",
	Run: func(_ *cobra.Command, _ []string) {
		srv := server.NewServer(":19917")

		err := srv.Start()
		if err != nil {
			slog.Error("Cannot start server", "error", err)
			os.Exit(1)
		}
		slog.Info("Tunnel server started")

		signalChan := make(chan os.Signal)
		signal.Notify(signalChan)

		select {
		case <-signalChan:
			slog.Info("Received signal. Stopping server")
			srv.Stop()
		case <-srv.Done():
			slog.Error("Server stopped it self")
		}
		slog.Info("Tunnel server stopped")
	},
}

func Exec() {
	RootCmd.Execute()
}
