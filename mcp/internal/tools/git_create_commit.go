package tools

import (
	"context"
	"encoding/json"
	"log/slog"

	"github.com/mark3labs/mcp-go/mcp"
	"github.com/mark3labs/mcp-go/server"
	"github.com/nccuhacks/nccu26/mcp/internal/models"
)

// CreateCommitTool groups merged file results from a push into one
// coherent multi-file Git commit.
func CreateCommitTool(d Deps) ToolEntry {
	tool := mcp.NewTool(
		"create_commit",
		mcp.WithDescription(
			"Creates a grouped multi-file Git commit from the current file "+
				"state of a push. All files from the push are combined into a "+
				"single coherent commit. Call prepare_commit first to verify "+
				"readiness.",
		),
		mcp.WithString("push_id", mcp.Required(), mcp.Description("The push/changeset ID to commit")),
		mcp.WithString("message", mcp.Description("Commit message (auto-generated if omitted)")),
	)

	handler := func(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
		if d.GitService == nil {
			return mcp.NewToolResultError("git service not configured"), nil
		}

		pushID, _ := req.Params.Arguments["push_id"].(string)
		message, _ := req.Params.Arguments["message"].(string)

		if pushID == "" {
			return mcp.NewToolResultError("push_id is required"), nil
		}

		resp, err := d.GitService.CreateGroupedCommit(ctx, models.GroupedCommitRequest{
			PushID:  pushID,
			Message: message,
		})
		if err != nil {
			slog.Error("grouped commit failed", "error", err)
			return mcp.NewToolResultError("commit failed: " + err.Error()), nil
		}

		data, _ := json.MarshalIndent(resp, "", "  ")
		return mcp.NewToolResultText(string(data)), nil
	}

	return ToolEntry{Tool: tool, Handler: server.ToolHandlerFunc(handler)}
}
