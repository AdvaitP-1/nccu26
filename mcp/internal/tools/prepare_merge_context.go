package tools

import (
	"context"
	"encoding/json"

	"github.com/mark3labs/mcp-go/mcp"
	"github.com/mark3labs/mcp-go/server"
)

// PrepareMergeContextTool returns the data an external system needs to
// make a merge decision.  This layer does NOT decide the merge.
func PrepareMergeContextTool(d Deps) ToolEntry {
	tool := mcp.NewTool(
		"prepare_merge_context",
		mcp.WithDescription(
			"Returns the base content, active diff nodes, reconstructed "+
				"candidate versions, and raw diff payloads for one file on "+
				"one branch.  The caller uses this data to make a merge decision "+
				"externally, then calls apply_merge_result.",
		),
		mcp.WithString("branch_name", mcp.Required(), mcp.Description("Git branch name")),
		mcp.WithString("file_path", mcp.Required(), mcp.Description("File path to prepare merge context for")),
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

		resp, err := d.GitService.PrepareMergeContext(branchName, filePath)
		if err != nil {
			return mcp.NewToolResultError(err.Error()), nil
		}

		data, _ := json.MarshalIndent(resp, "", "  ")
		return mcp.NewToolResultText(string(data)), nil
	}

	return ToolEntry{Tool: tool, Handler: server.ToolHandlerFunc(handler)}
}
