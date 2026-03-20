// Package tools defines MCP tool handlers and their registration.
//
// Each tool lives in its own file for readability.  The Registry function
// wires them all together with the shared dependencies (VFS, analysis
// client, policy, commits) and returns a slice ready for server.AddTool().
package tools

import (
	"github.com/mark3labs/mcp-go/mcp"
	"github.com/mark3labs/mcp-go/server"
	"github.com/nccuhacks/nccu26/mcp/internal/analysisclient"
	"github.com/nccuhacks/nccu26/mcp/internal/commits"
	"github.com/nccuhacks/nccu26/mcp/internal/policy"
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
}

// All returns every MCP tool the server should expose.
func All(d Deps) []ToolEntry {
	return []ToolEntry{
		GetVFSStateTool(d),
		IdentifyOverlapsTool(d),
		RequestMicroCommitTool(d),
	}
}
