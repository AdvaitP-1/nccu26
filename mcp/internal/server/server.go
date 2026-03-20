// Package server wires up the MCP server with all registered tools.
//
// It creates the mark3labs/mcp-go MCPServer, registers every tool from
// the tools package, and exposes an SSE transport suitable for remote
// invocation (e.g. from IBM watsonx Orchestrate).
package server

import (
	"log/slog"

	mcpserver "github.com/mark3labs/mcp-go/server"
	"github.com/nccuhacks/nccu26/mcp/internal/analysisclient"
	"github.com/nccuhacks/nccu26/mcp/internal/commits"
	"github.com/nccuhacks/nccu26/mcp/internal/config"
	"github.com/nccuhacks/nccu26/mcp/internal/policy"
	"github.com/nccuhacks/nccu26/mcp/internal/tools"
	"github.com/nccuhacks/nccu26/mcp/internal/vfs"
)

// Run initialises all dependencies, registers MCP tools, and starts the
// SSE server.  It blocks until the server exits.
func Run(cfg config.Config) error {
	// -- Dependencies --
	vfsMgr := vfs.NewManager()
	analysisCli := analysisclient.New(cfg.BackendBaseURL, cfg.BackendTimeout)
	policyEval := policy.NewEvaluator(cfg.RiskThreshold, cfg.BlockOnCritical)
	commitCoord := commits.NewCoordinator()

	deps := tools.Deps{
		VFS:      vfsMgr,
		Analysis: analysisCli,
		Policy:   policyEval,
		Commits:  commitCoord,
	}

	// -- MCP Server --
	mcpSrv := mcpserver.NewMCPServer(
		"pm-assistant-mcp",
		"0.1.0",
		mcpserver.WithToolCapabilities(true),
	)

	for _, entry := range tools.All(deps) {
		mcpSrv.AddTool(entry.Tool, entry.Handler)
	}

	slog.Info("registered MCP tools",
		"tools", []string{"get_vfs_state", "identify_overlaps", "request_micro_commit"},
	)

	// -- Transport --
	// SSE is the transport of choice for remote MCP servers.
	// IBM watsonx Orchestrate will connect to this endpoint.
	baseURL := "http://localhost" + cfg.ServerAddr
	sse := mcpserver.NewSSEServer(mcpSrv,
		mcpserver.WithBaseURL(baseURL),
	)

	slog.Info("starting MCP SSE server", "addr", cfg.ServerAddr, "base_url", baseURL)
	return sse.Start(cfg.ServerAddr)
}

// ToolNames is a convenience for logging / health endpoints.
func ToolNames() []string {
	return []string{
		"get_vfs_state",
		"identify_overlaps",
		"request_micro_commit",
	}
}
