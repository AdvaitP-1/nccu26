package tools

import (
	"context"
	"encoding/json"

	"github.com/mark3labs/mcp-go/mcp"
	"github.com/mark3labs/mcp-go/server"
)

// GetCommitStatusTool returns a commit record and its status.
func GetCommitStatusTool(d Deps) ToolEntry {
	tool := mcp.NewTool(
		"get_commit_status",
		mcp.WithDescription(
			"Returns the commit record for a given commit ID, including its "+
				"status, Git SHA, associated push, and file nodes.",
		),
		mcp.WithString("commit_id", mcp.Required(), mcp.Description("The commit record ID")),
	)

	handler := func(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
		if d.GitService == nil {
			return mcp.NewToolResultError("git service not configured"), nil
		}

		commitID, _ := req.Params.Arguments["commit_id"].(string)
		if commitID == "" {
			return mcp.NewToolResultError("commit_id is required"), nil
		}

		record, err := d.GitService.GetCommitRecord(commitID)
		if err != nil {
			return mcp.NewToolResultError(err.Error()), nil
		}

		data, _ := json.MarshalIndent(record, "", "  ")
		return mcp.NewToolResultText(string(data)), nil
	}

	return ToolEntry{Tool: tool, Handler: server.ToolHandlerFunc(handler)}
}
