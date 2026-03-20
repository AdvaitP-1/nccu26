package tools

import (
	"context"
	"encoding/json"

	"github.com/mark3labs/mcp-go/mcp"
	"github.com/mark3labs/mcp-go/server"
)

// GetBranchFileStateTool returns the tracked state for one file on one branch.
func GetBranchFileStateTool(d Deps) ToolEntry {
	tool := mcp.NewTool(
		"get_branch_file_state",
		mcp.WithDescription(
			"Returns the tracked per-file tree state for a specific file on a "+
				"specific branch, including active diff nodes and the current head.",
		),
		mcp.WithString("branch_name", mcp.Required(), mcp.Description("Git branch name")),
		mcp.WithString("file_path", mcp.Required(), mcp.Description("File path relative to repo root")),
	)

	handler := func(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
		if d.GitService == nil {
			return mcp.NewToolResultError("git service not configured"), nil
		}

		branchName, _ := req.Params.Arguments["branch_name"].(string)
		filePath, _ := req.Params.Arguments["file_path"].(string)

		if branchName == "" || filePath == "" {
			return mcp.NewToolResultError("branch_name and file_path are required"), nil
		}

		resp, err := d.GitService.GetFileState(branchName, filePath)
		if err != nil {
			return mcp.NewToolResultError(err.Error()), nil
		}

		data, _ := json.MarshalIndent(resp, "", "  ")
		return mcp.NewToolResultText(string(data)), nil
	}

	return ToolEntry{Tool: tool, Handler: server.ToolHandlerFunc(handler)}
}
