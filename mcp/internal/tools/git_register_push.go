package tools

import (
	"context"
	"encoding/json"
	"log/slog"

	"github.com/mark3labs/mcp-go/mcp"
	"github.com/mark3labs/mcp-go/server"
	"github.com/nccuhacks/nccu26/mcp/internal/models"
)

// RegisterPushTool registers a push (changeset) with multiple file changes.
// One push creates multiple file-level diff nodes across per-file trees.
func RegisterPushTool(d Deps) ToolEntry {
	tool := mcp.NewTool(
		"register_push",
		mcp.WithDescription(
			"Registers a developer/agent push containing one or more file changes. "+
				"Creates diff nodes in each file's per-file tree on the specified branch.",
		),
		mcp.WithString("branch_name", mcp.Required(), mcp.Description("Target Git branch name")),
		mcp.WithString("user_id", mcp.Required(), mcp.Description("Developer or agent identifier")),
		mcp.WithString("files_json", mcp.Required(), mcp.Description(
			`JSON array of file changes: [{"file_path":"...","base_content":"...","new_content":"..."}]`,
		)),
		mcp.WithString("message", mcp.Description("Optional push description")),
	)

	handler := func(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
		if d.GitService == nil {
			return mcp.NewToolResultError("git service not configured"), nil
		}

		branchName, _ := req.Params.Arguments["branch_name"].(string)
		userID, _ := req.Params.Arguments["user_id"].(string)
		filesJSON, _ := req.Params.Arguments["files_json"].(string)
		message, _ := req.Params.Arguments["message"].(string)

		if branchName == "" || userID == "" || filesJSON == "" {
			return mcp.NewToolResultError("branch_name, user_id, and files_json are required"), nil
		}

		var files []models.PushFileChange
		if err := json.Unmarshal([]byte(filesJSON), &files); err != nil {
			return mcp.NewToolResultError("invalid files_json: " + err.Error()), nil
		}

		resp, err := d.GitService.IngestPush(ctx, models.IngestPushRequest{
			BranchName: branchName,
			UserID:     userID,
			Files:      files,
			Message:    message,
		})
		if err != nil {
			slog.Error("push ingestion failed", "error", err)
			return mcp.NewToolResultError("push ingestion failed: " + err.Error()), nil
		}

		data, _ := json.MarshalIndent(resp, "", "  ")
		return mcp.NewToolResultText(string(data)), nil
	}

	return ToolEntry{Tool: tool, Handler: server.ToolHandlerFunc(handler)}
}
