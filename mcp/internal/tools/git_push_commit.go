package tools

import (
	"context"
	"encoding/json"
	"log/slog"

	"github.com/mark3labs/mcp-go/mcp"
	"github.com/mark3labs/mcp-go/server"
	"github.com/nccuhacks/nccu26/mcp/internal/models"
)

// PushCommitTool pushes a previously created commit to the remote.
func PushCommitTool(d Deps) ToolEntry {
	tool := mcp.NewTool(
		"push_commit",
		mcp.WithDescription(
			"Pushes a previously created commit to the remote Git repository. "+
				"Includes retry logic for transient failures.",
		),
		mcp.WithString("commit_id", mcp.Required(), mcp.Description("The commit record ID to push")),
	)

	handler := func(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
		if d.GitService == nil {
			return mcp.NewToolResultError("git service not configured"), nil
		}

		commitID, _ := req.Params.Arguments["commit_id"].(string)
		if commitID == "" {
			return mcp.NewToolResultError("commit_id is required"), nil
		}

		resp, err := d.GitService.PushCommit(ctx, models.GitPushRequest{
			CommitID: commitID,
		})
		if err != nil {
			slog.Error("push failed", "error", err)
			return mcp.NewToolResultError("push failed: " + err.Error()), nil
		}

		data, _ := json.MarshalIndent(resp, "", "  ")
		return mcp.NewToolResultText(string(data)), nil
	}

	return ToolEntry{Tool: tool, Handler: server.ToolHandlerFunc(handler)}
}
