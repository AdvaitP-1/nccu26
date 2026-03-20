package tools

import (
	"context"
	"encoding/json"
	"fmt"
	"log/slog"
	"time"

	"github.com/mark3labs/mcp-go/mcp"
	"github.com/mark3labs/mcp-go/server"
	"github.com/nccuhacks/nccu26/mcp/internal/models"
)

// RequestMicroCommitTool evaluates whether the calling agent's pending
// changes are safe to commit and, if so, executes the commit flow.
//
// Parameters (via MCP tool arguments):
//   - agent_id  (required) — which agent is requesting.
//   - message   (optional) — commit message; defaults to a generic one.
func RequestMicroCommitTool(d Deps) ToolEntry {
	tool := mcp.NewTool(
		"request_micro_commit",
		mcp.WithDescription(
			"Requests a micro-commit for the calling agent's pending changes. "+
				"Runs overlap analysis and policy evaluation first; blocks the "+
				"commit if risk is too high.",
		),
		mcp.WithString("agent_id", mcp.Required(), mcp.Description("The agent requesting the commit")),
		mcp.WithString("message", mcp.Description("Commit message (optional)")),
	)

	handler := func(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
		agentID, _ := req.Params.Arguments["agent_id"].(string)
		if agentID == "" {
			return mcp.NewToolResultError("agent_id is required"), nil
		}

		message, _ := req.Params.Arguments["message"].(string)
		if message == "" {
			message = fmt.Sprintf("micro-commit by %s", agentID)
		}

		// 1. Verify the agent has pending work.
		files, err := d.VFS.FilesForAgent(agentID)
		if err != nil {
			return mcp.NewToolResultError(err.Error()), nil
		}

		// 2. Run overlap analysis across the full VFS.
		changesets := d.VFS.ChangeSetsForAnalysis()

		var analysisResp *models.AnalyzeOverlapsResponse
		if len(changesets) >= 2 {
			bgCtx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
			defer cancel()
			analysisResp, err = d.Analysis.AnalyzeOverlaps(bgCtx, changesets)
			if err != nil {
				slog.Error("analysis failed during micro-commit", "error", err)
				return mcp.NewToolResultError("analysis failed: " + err.Error()), nil
			}
		}

		// 3. Evaluate policy — returns structured decision with all reasons.
		decision := d.Policy.Evaluate(analysisResp)
		if !decision.Allowed {
			payload, _ := json.MarshalIndent(decision, "", "  ")
			return mcp.NewToolResultText(string(payload)), nil
		}

		// 4. Execute commit.
		result := d.Commits.Commit(models.CommitRequest{
			AgentID: agentID,
			Files:   files,
			Message: message,
		})

		// 5. On success, clear the agent's VFS entries.
		if result.Allowed {
			d.VFS.Clear(agentID)
		}

		payload, _ := json.MarshalIndent(result, "", "  ")
		return mcp.NewToolResultText(string(payload)), nil
	}

	return ToolEntry{Tool: tool, Handler: server.ToolHandlerFunc(handler)}
}
