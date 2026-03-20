// Package models defines the domain types shared across the MCP internals.
//
// These types map to the backend's request/response schemas and to the
// VFS state exposed through MCP tools.
package models

import "time"

// ---------------------------------------------------------------------------
// VFS
// ---------------------------------------------------------------------------

// FileSnapshot is a single proposed file version held in the VFS.
type FileSnapshot struct {
	Path     string `json:"path"`
	Language string `json:"language"`
	Content  string `json:"content"`
}

// PendingChange groups file snapshots under an agent/session scope.
type PendingChange struct {
	AgentID   string         `json:"agent_id"`
	SessionID string         `json:"session_id,omitempty"`
	TaskID    string         `json:"task_id,omitempty"`
	Files     []FileSnapshot `json:"files"`
	CreatedAt time.Time      `json:"created_at"`
	UpdatedAt time.Time      `json:"updated_at"`
}

// VFSState is the full in-memory workspace state returned by get_vfs_state.
type VFSState struct {
	PendingChanges []PendingChange `json:"pending_changes"`
	TotalFiles     int             `json:"total_files"`
	TotalAgents    int             `json:"total_agents"`
}

// ---------------------------------------------------------------------------
// Backend analysis (mirrors backend schemas)
// ---------------------------------------------------------------------------

// AnalysisFileInput is a single file sent to the backend for parsing.
type AnalysisFileInput struct {
	Path     string `json:"path"`
	Language string `json:"language"`
	Content  string `json:"content"`
}

// AnalysisChangeSet groups files under an agent ID for the backend.
type AnalysisChangeSet struct {
	AgentID string              `json:"agent_id"`
	Files   []AnalysisFileInput `json:"files"`
}

// AnalyzeOverlapsRequest is the payload sent to POST /analyze/overlaps.
type AnalyzeOverlapsRequest struct {
	Changesets []AnalysisChangeSet `json:"changesets"`
}

// Overlap describes a single structural conflict between two agents.
type Overlap struct {
	FilePath   string `json:"file_path"`
	SymbolName string `json:"symbol_name"`
	SymbolKind string `json:"symbol_kind"`
	AgentA     string `json:"agent_a"`
	AgentB     string `json:"agent_b"`
	Severity   string `json:"severity"` // critical | high | medium | low
	Reason     string `json:"reason"`
	StartLineA int    `json:"start_line_a"`
	EndLineA   int    `json:"end_line_a"`
	StartLineB int    `json:"start_line_b"`
	EndLineB   int    `json:"end_line_b"`
}

// FileRisk is a per-file risk summary returned by the backend.
type FileRisk struct {
	FilePath             string   `json:"file_path"`
	RiskScore            int      `json:"risk_score"`
	StabilityScore       int      `json:"stability_score"`
	OverlapCount         int      `json:"overlap_count"`
	Contributors         []string `json:"contributors"`
	ContributorsCount    int      `json:"contributors_count"`
	PairwiseOverlapCount int      `json:"pairwise_overlap_count"`
	MaxSeverity          *string  `json:"max_severity"`
	IsHotspot            bool     `json:"is_hotspot"`
	Summary              string   `json:"summary"`
}

// AnalyzeOverlapsResponse is the payload returned from POST /analyze/overlaps.
type AnalyzeOverlapsResponse struct {
	Overlaps  []Overlap  `json:"overlaps"`
	FileRisks []FileRisk `json:"file_risks"`
}

// ---------------------------------------------------------------------------
// Micro-commit
// ---------------------------------------------------------------------------

// CommitRequest describes a set of changes an agent wants to commit.
type CommitRequest struct {
	AgentID   string         `json:"agent_id"`
	SessionID string         `json:"session_id,omitempty"`
	Files     []FileSnapshot `json:"files"`
	Message   string         `json:"message"`
}

// CommitResult is returned after a micro-commit attempt.
type CommitResult struct {
	Allowed  bool   `json:"allowed"`
	CommitID string `json:"commit_id,omitempty"`
	Reason   string `json:"reason"`
}
