package tools

import (
	"context"
	"encoding/json"

	"github.com/mark3labs/mcp-go/mcp"
	"github.com/mark3labs/mcp-go/server"
)

// PrepareCommitTool returns what would be included in a grouped commit
// for a push, and whether all files have been resolved.  This is a dry
// run — it does not create the commit.
func PrepareCommitTool(d Deps) ToolEntry {
	tool := mcp.NewTool(
		"prepare_commit",
		mcp.WithDescription(
			"Dry-run for a grouped commit.  Returns the list of files that "+
				"would be included, their current content, and whether all files "+
				"in the push have been resolved (no remaining active diff nodes). "+
				"Call create_commit to actually execute.",
		),
		mcp.WithString("push_id", mcp.Required(), mcp.Description("The push/changeset ID to inspect")),
	)

	handler := func(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
		if d.GitService == nil {
			return mcp.NewToolResultError("git service not configured"), nil
		}

		pushID, _ := req.Params.Arguments["push_id"].(string)
		if pushID == "" {
			return mcp.NewToolResultError("push_id is required"), nil
		}

		resp, err := d.GitService.PrepareCommit(pushID)
		if err != nil {
			return mcp.NewToolResultError(err.Error()), nil
		}

		data, _ := json.MarshalIndent(resp, "", "  ")
		return mcp.NewToolResultText(string(data)), nil
	}

	return ToolEntry{Tool: tool, Handler: server.ToolHandlerFunc(handler)}
}
