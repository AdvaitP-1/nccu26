// Package models — git_models.go extends the domain types with models for
// the branch-scoped per-file diff tree system.
//
// Key concepts:
//   - Branch: a tracked Git branch
//   - BranchFile: one file on one specific branch (branch-scoped)
//   - FileNode: a proposed diff attached to a file's tree
//   - DiffBlob: the serialised patch payload
//   - PushSet: one push spanning multiple files
//   - CommitRecord: a produced Git commit
package models

import "time"

// ---------------------------------------------------------------------------
// Status enums
// ---------------------------------------------------------------------------

type NodeStatus string

const (
	NodeStatusActive     NodeStatus = "active"
	NodeStatusMerged     NodeStatus = "merged"
	NodeStatusSuperseded NodeStatus = "superseded"
	NodeStatusRejected   NodeStatus = "rejected"
)

type PushStatus string

const (
	PushStatusPending   PushStatus = "pending"
	PushStatusAnalyzed  PushStatus = "analyzed"
	PushStatusApproved  PushStatus = "approved"
	PushStatusCommitted PushStatus = "committed"
	PushStatusRejected  PushStatus = "rejected"
	PushStatusFailed    PushStatus = "failed"
)

type CommitStatus string

const (
	CommitStatusCreated    CommitStatus = "created"
	CommitStatusPushed     CommitStatus = "pushed"
	CommitStatusPushFailed CommitStatus = "push_failed"
)

// ---------------------------------------------------------------------------
// Core entities
// ---------------------------------------------------------------------------

// Branch represents one tracked Git branch.
type Branch struct {
	BranchID       string    `json:"branch_id"`
	GitBranchName  string    `json:"git_branch_name"`
	RemoteName     string    `json:"remote_name"`
	BaseCommitSHA  string    `json:"base_commit_sha"`
	CurrentHeadSHA string    `json:"current_head_sha"`
	CreatedAt      time.Time `json:"created_at"`
	UpdatedAt      time.Time `json:"updated_at"`
}

// BranchFile represents one file tracked on one specific branch.
// The same path on different branches is a different BranchFile.
type BranchFile struct {
	BranchFileID      string    `json:"branch_file_id"`
	BranchID          string    `json:"branch_id"`
	FilePath          string    `json:"file_path"`
	CurrentHeadNodeID string    `json:"current_head_node_id"`
	BaseContent       string    `json:"base_content"`
	CreatedAt         time.Time `json:"created_at"`
	UpdatedAt         time.Time `json:"updated_at"`
}

// FileNode represents one proposed diff node in a file's tree.
type FileNode struct {
	NodeID       string     `json:"node_id"`
	BranchFileID string     `json:"branch_file_id"`
	ParentNodeID string     `json:"parent_node_id"`
	UserID       string     `json:"user_id"`
	PushID       string     `json:"push_id"`
	DiffBlobID   string     `json:"diff_blob_id"`
	Status       NodeStatus `json:"status"`
	CreatedAt    time.Time  `json:"created_at"`
	UpdatedAt    time.Time  `json:"updated_at"`
}

// DiffBlob stores the actual patch/change payload.
type DiffBlob struct {
	DiffBlobID  string    `json:"diff_blob_id"`
	Format      string    `json:"format"`
	Payload     string    `json:"payload"`
	ContentHash string    `json:"content_hash"`
	CreatedAt   time.Time `json:"created_at"`
}

// PushSet represents one developer/agent push spanning multiple files.
type PushSet struct {
	PushID    string     `json:"push_id"`
	BranchID  string     `json:"branch_id"`
	UserID    string     `json:"user_id"`
	Status    PushStatus `json:"status"`
	NodeIDs   []string   `json:"node_ids"`
	CreatedAt time.Time  `json:"created_at"`
	UpdatedAt time.Time  `json:"updated_at"`
}

// CommitRecord tracks a produced Git commit.
type CommitRecord struct {
	CommitID     string       `json:"commit_id"`
	BranchID     string       `json:"branch_id"`
	PushID       string       `json:"push_id"`
	GitCommitSHA string       `json:"git_commit_sha"`
	Status       CommitStatus `json:"status"`
	Message      string       `json:"message"`
	FileNodeIDs  []string     `json:"file_node_ids"`
	CreatedAt    time.Time    `json:"created_at"`
	UpdatedAt    time.Time    `json:"updated_at"`
	PushedAt     *time.Time   `json:"pushed_at,omitempty"`
}

// ---------------------------------------------------------------------------
// Request / Response types for the git orchestration API
// ---------------------------------------------------------------------------

// PushFileChange describes one file's change within a push.
type PushFileChange struct {
	FilePath    string `json:"file_path"`
	Language    string `json:"language,omitempty"`
	BaseContent string `json:"base_content"`
	NewContent  string `json:"new_content"`
}

// IngestPushRequest is the input for push ingestion.
type IngestPushRequest struct {
	BranchName string           `json:"branch_name"`
	UserID     string           `json:"user_id"`
	Files      []PushFileChange `json:"files"`
	Message    string           `json:"message,omitempty"`
}

// IngestPushResponse is returned after ingesting a push.
type IngestPushResponse struct {
	PushID  string   `json:"push_id"`
	NodeIDs []string `json:"node_ids"`
}

// FileStateResponse describes the tracked state of one file on one branch.
type FileStateResponse struct {
	BranchFile  BranchFile `json:"branch_file"`
	ActiveNodes []FileNode `json:"active_nodes"`
	TotalNodes  int        `json:"total_nodes"`
}

// ---------------------------------------------------------------------------
// Merge context (data for external merge decision)
// ---------------------------------------------------------------------------

// MergeContextResponse provides all data needed by an external system to
// make a merge decision for one file on one branch.  This layer does NOT
// decide the merge — it only prepares the context.
type MergeContextResponse struct {
	BranchFile   BranchFile       `json:"branch_file"`
	BaseContent  string           `json:"base_content"`
	ActiveNodes  []FileNode       `json:"active_nodes"`
	Candidates   []MergeCandidate `json:"candidates"`
	DiffPayloads []string         `json:"diff_payloads"`
}

// MergeCandidate pairs a node with the fully reconstructed file content
// that would result from applying that node's diff to the base.
type MergeCandidate struct {
	NodeID               string `json:"node_id"`
	UserID               string `json:"user_id"`
	PushID               string `json:"push_id"`
	ReconstructedContent string `json:"reconstructed_content"`
}

// ---------------------------------------------------------------------------
// Apply merge result (externally decided)
// ---------------------------------------------------------------------------

// ApplyMergeResultRequest carries the externally decided merge result
// back to this layer for execution.
type ApplyMergeResultRequest struct {
	BranchName    string `json:"branch_name"`
	FilePath      string `json:"file_path"`
	MergedContent string `json:"merged_content"`
	// If empty, all active nodes for this file are superseded.
	SupersededNodeIDs []string `json:"superseded_node_ids,omitempty"`
}

// ApplyMergeResultResponse is returned after the merge result is applied.
type ApplyMergeResultResponse struct {
	MergedNodeID  string `json:"merged_node_id"`
	NodesAffected int    `json:"nodes_affected"`
	BranchFileID  string `json:"branch_file_id"`
}

// ---------------------------------------------------------------------------
// Prepare commit (dry run)
// ---------------------------------------------------------------------------

// PrepareCommitResponse shows what would be included in a commit for a
// push, and whether all files in the push have been resolved.
type PrepareCommitResponse struct {
	PushID       string            `json:"push_id"`
	BranchName   string            `json:"branch_name"`
	Files        []CommitFileEntry `json:"files"`
	AllResolved  bool              `json:"all_resolved"`
	PendingFiles []string          `json:"pending_files,omitempty"`
}

// CommitFileEntry describes one file that would be included in a commit.
type CommitFileEntry struct {
	FilePath string `json:"file_path"`
	Content  string `json:"content"`
	NodeID   string `json:"node_id"`
}

// ---------------------------------------------------------------------------
// Create commit / push commit
// ---------------------------------------------------------------------------

// GroupedCommitRequest is the input for grouped multi-file commit.
type GroupedCommitRequest struct {
	PushID  string `json:"push_id"`
	Message string `json:"message,omitempty"`
}

// GroupedCommitResponse is returned after creating a grouped commit.
type GroupedCommitResponse struct {
	CommitID      string `json:"commit_id"`
	GitCommitSHA  string `json:"git_commit_sha,omitempty"`
	FilesIncluded int    `json:"files_included"`
	Status        string `json:"status"`
}

// GitPushRequest is the input for pushing a commit to remote.
type GitPushRequest struct {
	CommitID string `json:"commit_id"`
}

// GitPushResponse is returned after pushing.
type GitPushResponse struct {
	CommitID string `json:"commit_id"`
	Status   string `json:"status"`
	Message  string `json:"message"`
}

// ---------------------------------------------------------------------------
// Backend merge endpoint types (mirrors backend POST /merge)
// ---------------------------------------------------------------------------

// BackendMergeRequest is the payload for POST /merge on the backend.
type BackendMergeRequest struct {
	BaseContent string   `json:"base_content"`
	Diffs       []string `json:"diffs"`
}

// BackendConflict describes a conflict reported by the backend merge.
type BackendConflict struct {
	DiffIndex int    `json:"diff_index"`
	Reason    string `json:"reason"`
}

// BackendMergeResponse is the response from POST /merge.
type BackendMergeResponse struct {
	Success       bool              `json:"success"`
	MergedContent string            `json:"merged_content"`
	Conflicts     []BackendConflict `json:"conflicts,omitempty"`
}
