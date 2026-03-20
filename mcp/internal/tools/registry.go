// Package tools defines MCP tool handlers and their registration.
//
// Each tool lives in its own file for readability.  The All function
// wires them together with shared dependencies and returns a slice
// ready for server.AddTool().
package tools

import (
	"github.com/mark3labs/mcp-go/mcp"
	"github.com/mark3labs/mcp-go/server"
	"github.com/nccuhacks/nccu26/mcp/internal/analysisclient"
	"github.com/nccuhacks/nccu26/mcp/internal/commits"
	"github.com/nccuhacks/nccu26/mcp/internal/policy"
	svc "github.com/nccuhacks/nccu26/mcp/internal/service"
	"github.com/nccuhacks/nccu26/mcp/internal/vfs"
)

// ToolEntry pairs a tool definition with its handler.
type ToolEntry struct {
	Tool    mcp.Tool
	Handler server.ToolHandlerFunc
}

// Deps groups shared dependencies injected into every tool handler.
type Deps struct {
	VFS        *vfs.Manager
	Analysis   *analysisclient.Client
	Policy     *policy.Evaluator
	Commits    *commits.Coordinator
	GitService *svc.GitService
}

// All returns every MCP tool the server should expose.
func All(d Deps) []ToolEntry {
	entries := []ToolEntry{
		GetVFSStateTool(d),
		IdentifyOverlapsTool(d),
		RequestMicroCommitTool(d),
	}

	if d.GitService != nil {
		entries = append(entries,
			GitHealthTool(d),
			RegisterPushTool(d),
			GetBranchFileStateTool(d),
			PrepareMergeContextTool(d),
			ApplyMergeResultTool(d),
			PrepareCommitTool(d),
			CreateCommitTool(d),
			PushCommitTool(d),
			GetCommitStatusTool(d),
		)
	}

	return entries
}
