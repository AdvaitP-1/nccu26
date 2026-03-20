package tools

import (
	"context"
	"encoding/json"
	"log/slog"
	"time"

	"github.com/mark3labs/mcp-go/mcp"
	"github.com/mark3labs/mcp-go/server"
)

// IdentifyOverlapsTool gathers pending VFS changes and calls the backend
// /analyze/overlaps endpoint.  The analysis result is returned verbatim so
// that the orchestrator (or agent) can reason about structural conflicts.
func IdentifyOverlapsTool(d Deps) ToolEntry {
	tool := mcp.NewTool(
		"identify_overlaps",
		mcp.WithDescription(
			"Analyses all pending VFS changes for structural overlaps. "+
				"Calls the backend analysis service and returns per-symbol "+
				"overlaps and per-file risk scores.",
		),
	)

	handler := func(_ context.Context, _ mcp.CallToolRequest) (*mcp.CallToolResult, error) {
		changesets := d.VFS.ChangeSetsForAnalysis()
		if len(changesets) < 2 {
			return mcp.NewToolResultText(
				`{"overlaps":[],"file_risks":[],"note":"fewer than 2 agents — nothing to compare"}`,
			), nil
		}

		// Use a detached context — the SSE transport may cancel the
		// original request context before the backend responds.
		bgCtx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
		defer cancel()

		resp, err := d.Analysis.AnalyzeOverlaps(bgCtx, changesets)
		if err != nil {
			slog.Error("backend analysis failed", "error", err)
			return mcp.NewToolResultError("backend analysis failed: " + err.Error()), nil
		}

		data, err := json.MarshalIndent(resp, "", "  ")
		if err != nil {
			return mcp.NewToolResultError("failed to serialise analysis result"), nil
		}

		return mcp.NewToolResultText(string(data)), nil
	}

	return ToolEntry{Tool: tool, Handler: server.ToolHandlerFunc(handler)}
}
