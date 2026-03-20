// Command server starts the Product Manager Assistant MCP server.
//
// Usage:
//
//	go run ./cmd/server
//
// Environment variables (see internal/config for full list):
//
//	MCP_SERVER_ADDR      — listen address       (default ":9090")
//	MCP_BACKEND_URL      — backend base URL     (default "http://localhost:8000")
//	MCP_RISK_THRESHOLD   — max safe risk score  (default 70)
//	MCP_BLOCK_ON_CRITICAL — block critical?     (default true)
//	MCP_LOG_LEVEL        — "debug" for verbose  (default "info")
package main

import (
	"log/slog"
	"os"

	"github.com/nccuhacks/nccu26/mcp/internal/config"
	"github.com/nccuhacks/nccu26/mcp/internal/logging"
	"github.com/nccuhacks/nccu26/mcp/internal/server"
)

func main() {
	logging.Init()

	cfg := config.Load()
	slog.Info("configuration loaded",
		"server_addr", cfg.ServerAddr,
		"http_addr", cfg.HTTPAddr,
		"backend_url", cfg.BackendBaseURL,
		"risk_threshold", cfg.RiskThreshold,
		"block_on_critical", cfg.BlockOnCritical,
		"git_repo_path", cfg.GitRepoPath,
		"git_remote", cfg.GitRemote,
	)

	if err := server.Run(cfg); err != nil {
		slog.Error("server exited with error", "error", err)
		os.Exit(1)
	}
}
