package tools

import (
	"context"
	"encoding/json"
	"log/slog"

	"github.com/mark3labs/mcp-go/mcp"
	"github.com/mark3labs/mcp-go/server"
	"github.com/nccuhacks/nccu26/mcp/internal/models"
)

// ApplyMergeResultTool accepts an externally decided merge result and
// applies it to the file tree — creates a merge node, promotes it to
// head, and supersedes resolved nodes.
func ApplyMergeResultTool(d Deps) ToolEntry {
	tool := mcp.NewTool(
		"apply_merge_result",
		mcp.WithDescription(
			"Applies an externally decided merge result for one file on one "+
				"branch.  Creates a merge node with the provided content, promotes "+
				"it to head, and supersedes the previously active nodes.",
		),
		mcp.WithString("branch_name", mcp.Required(), mcp.Description("Git branch name")),
		mcp.WithString("file_path", mcp.Required(), mcp.Description("File path the merge applies to")),
		mcp.WithString("merged_content", mcp.Required(), mcp.Description("The externally decided merged file content")),
		mcp.WithString("superseded_node_ids_json", mcp.Description(
			"Optional JSON array of node IDs to supersede. If omitted, all active nodes are superseded.",
		)),
	)

	handler := func(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
		if d.GitService == nil {
			return mcp.NewToolResultError("git service not configured"), nil
		}

		branchName, _ := req.Params.Arguments["branch_name"].(string)
		filePath, _ := req.Params.Arguments["file_path"].(string)
		mergedContent, _ := req.Params.Arguments["merged_content"].(string)
		supersededJSON, _ := req.Params.Arguments["superseded_node_ids_json"].(string)

		if branchName == "" || filePath == "" || mergedContent == "" {
			return mcp.NewToolResultError("branch_name, file_path, and merged_content are required"), nil
		}

		var supersededIDs []string
		if supersededJSON != "" {
			if err := json.Unmarshal([]byte(supersededJSON), &supersededIDs); err != nil {
				return mcp.NewToolResultError("invalid superseded_node_ids_json: " + err.Error()), nil
			}
		}

		resp, err := d.GitService.ApplyMergeResult(ctx, models.ApplyMergeResultRequest{
			BranchName:        branchName,
			FilePath:          filePath,
			MergedContent:     mergedContent,
			SupersededNodeIDs: supersededIDs,
		})
		if err != nil {
			slog.Error("apply merge result failed", "error", err)
			return mcp.NewToolResultError("apply merge result failed: " + err.Error()), nil
		}

		data, _ := json.MarshalIndent(resp, "", "  ")
		return mcp.NewToolResultText(string(data)), nil
	}

	return ToolEntry{Tool: tool, Handler: server.ToolHandlerFunc(handler)}
}
