// Package logging provides a thin structured-logging setup for the MCP server.
package logging

import (
	"log/slog"
	"os"
)

// Init configures the process-wide default slog logger.
// Call once from main before any other work.
func Init() {
	level := slog.LevelInfo
	if os.Getenv("MCP_LOG_LEVEL") == "debug" {
		level = slog.LevelDebug
	}

	handler := slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{Level: level})
	slog.SetDefault(slog.New(handler))
}
