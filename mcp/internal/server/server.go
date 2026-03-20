// Package server wires up the MCP server with all registered tools and
// optionally starts an HTTP API server for the Git state/execution layer.
//
// It creates the mark3labs/mcp-go MCPServer, registers every tool from
// the tools package, and exposes an SSE transport suitable for remote
// invocation by external systems.
package server

import (
	"log/slog"
	"net/http"

	mcpserver "github.com/mark3labs/mcp-go/server"
	"github.com/nccuhacks/nccu26/mcp/internal/analysisclient"
	"github.com/nccuhacks/nccu26/mcp/internal/commits"
	"github.com/nccuhacks/nccu26/mcp/internal/config"
	"github.com/nccuhacks/nccu26/mcp/internal/diff"
	"github.com/nccuhacks/nccu26/mcp/internal/filetree"
	"github.com/nccuhacks/nccu26/mcp/internal/gitcontrol"
	"github.com/nccuhacks/nccu26/mcp/internal/httpapi"
	"github.com/nccuhacks/nccu26/mcp/internal/policy"
	"github.com/nccuhacks/nccu26/mcp/internal/service"
	"github.com/nccuhacks/nccu26/mcp/internal/storage"
	"github.com/nccuhacks/nccu26/mcp/internal/tools"
	"github.com/nccuhacks/nccu26/mcp/internal/vfs"
)

// Run initialises all dependencies, registers MCP tools, and starts the
// SSE server.  It blocks until the server exits.
func Run(cfg config.Config) error {
	// -- Existing dependencies --
	vfsMgr := vfs.NewManager()
	analysisCli := analysisclient.New(cfg.BackendBaseURL, cfg.BackendTimeout)
	policyEval := policy.NewEvaluator(cfg.RiskThreshold, cfg.BlockOnCritical)
	commitCoord := commits.NewCoordinator()

	// -- Git state/execution dependencies --
	store := storage.New()
	engine := diff.NewEngine()
	treeMgr := filetree.NewManager(store, engine)

	var gitExec *gitcontrol.Executor
	if cfg.GitRepoPath != "" {
		gitExec = gitcontrol.NewExecutor(cfg.GitRepoPath)
		slog.Info("git executor configured", "repo_path", cfg.GitRepoPath)
	} else {
		slog.Info("git executor not configured — state-tracking mode only")
	}

	gitSvc := service.NewGitService(store, engine, treeMgr, gitExec)

	deps := tools.Deps{
		VFS:        vfsMgr,
		Analysis:   analysisCli,
		Policy:     policyEval,
		Commits:    commitCoord,
		GitService: gitSvc,
	}

	// -- MCP Server --
	mcpSrv := mcpserver.NewMCPServer(
		"pm-assistant-mcp",
		"0.2.0",
		mcpserver.WithToolCapabilities(true),
	)

	allTools := tools.All(deps)
	toolNames := make([]string, len(allTools))
	for i, entry := range allTools {
		mcpSrv.AddTool(entry.Tool, entry.Handler)
		toolNames[i] = entry.Tool.Name
	}

	slog.Info("registered MCP tools", "tools", toolNames)

	// -- HTTP API server (git endpoints) --
	if cfg.HTTPAddr != "" {
		mux := http.NewServeMux()
		handler := httpapi.NewHandler(gitSvc)
		handler.RegisterRoutes(mux)

		go func() {
			slog.Info("starting HTTP API server", "addr", cfg.HTTPAddr)
			if err := http.ListenAndServe(cfg.HTTPAddr, mux); err != nil {
				slog.Error("HTTP API server error", "error", err)
			}
		}()
	}

	// -- SSE Transport --
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
		"git_health",
		"register_push",
		"get_branch_file_state",
		"prepare_merge_context",
		"apply_merge_result",
		"prepare_commit",
		"create_commit",
		"push_commit",
		"get_commit_status",
	}
}
