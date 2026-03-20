package tools

import (
	"context"
	"encoding/json"

	"github.com/mark3labs/mcp-go/mcp"
	"github.com/mark3labs/mcp-go/server"
)

// GetVFSStateTool returns the tool definition and handler for get_vfs_state.
//
// This tool exposes the current VFS (shadow workspace) to the caller,
// showing all pending agent changes that have not yet been committed.
func GetVFSStateTool(d Deps) ToolEntry {
	tool := mcp.NewTool(
		"get_vfs_state",
		mcp.WithDescription(
			"Returns the current state of the virtual file system (shadow workspace), "+
				"including all pending file changes grouped by agent.",
		),
	)

	handler := func(_ context.Context, _ mcp.CallToolRequest) (*mcp.CallToolResult, error) {
		state := d.VFS.State()

		data, err := json.MarshalIndent(state, "", "  ")
		if err != nil {
			return mcp.NewToolResultError("failed to serialise VFS state"), nil
		}

		return mcp.NewToolResultText(string(data)), nil
	}

	return ToolEntry{Tool: tool, Handler: server.ToolHandlerFunc(handler)}
}
