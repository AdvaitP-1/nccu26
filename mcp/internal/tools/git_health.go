package tools

import (
	"context"
	"encoding/json"

	"github.com/mark3labs/mcp-go/mcp"
	"github.com/mark3labs/mcp-go/server"
)

// GitHealthTool returns the health status of the Git subsystem.
func GitHealthTool(d Deps) ToolEntry {
	tool := mcp.NewTool(
		"git_health",
		mcp.WithDescription(
			"Returns the health status of the Git orchestration subsystem, "+
				"including repository state and entity counts.",
		),
	)

	handler := func(ctx context.Context, _ mcp.CallToolRequest) (*mcp.CallToolResult, error) {
		if d.GitService == nil {
			return mcp.NewToolResultError("git service not configured"), nil
		}

		status := d.GitService.Health(ctx)
		data, err := json.MarshalIndent(status, "", "  ")
		if err != nil {
			return mcp.NewToolResultError("failed to serialise health status"), nil
		}
		return mcp.NewToolResultText(string(data)), nil
	}

	return ToolEntry{Tool: tool, Handler: server.ToolHandlerFunc(handler)}
}
