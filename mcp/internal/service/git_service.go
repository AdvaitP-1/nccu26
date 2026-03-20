// Package service provides the Git state and execution service.
//
// This is NOT an orchestration layer.  It exposes composable capabilities
// that external systems (e.g. IBM watsonx Orchestrate) call in whatever
// sequence they decide.
//
// Responsibilities:
//   - register pushes and create per-file diff nodes
//   - retrieve file tree state
//   - prepare merge context (data only, no decision)
//   - apply externally decided merge results
//   - prepare commit dry-runs
//   - execute Git commits and pushes
//   - track lifecycle statuses
package service

import (
	"context"
	"fmt"
	"log/slog"
	"sort"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/nccuhacks/nccu26/mcp/internal/diff"
	"github.com/nccuhacks/nccu26/mcp/internal/filetree"
	"github.com/nccuhacks/nccu26/mcp/internal/gitcontrol"
	"github.com/nccuhacks/nccu26/mcp/internal/models"
	"github.com/nccuhacks/nccu26/mcp/internal/storage"
	"github.com/nccuhacks/nccu26/mcp/internal/vfs"
)

// GitService provides Git state management and execution capabilities.
// It does not make merge decisions or sequence operations — that is the
// caller's responsibility.
type GitService struct {
	store  *storage.Store
	engine *diff.Engine
	tree   *filetree.Manager
	git    *gitcontrol.Executor
	vfs    *vfs.Manager
	logger *slog.Logger
}

// NewGitService builds a ready-to-use service.
func NewGitService(
	store *storage.Store,
	engine *diff.Engine,
	tree *filetree.Manager,
	git *gitcontrol.Executor,
	vfsMgr *vfs.Manager,
) *GitService {
	return &GitService{
		store:  store,
		engine: engine,
		tree:   tree,
		git:    git,
		vfs:    vfsMgr,
		logger: slog.Default().With("component", "git_service"),
	}
}

// ---------------------------------------------------------------------------
// Health
// ---------------------------------------------------------------------------

// HealthStatus is a summary of the git subsystem state.
type HealthStatus struct {
	Status   string         `json:"status"`
	RepoPath string         `json:"repo_path"`
	HasRepo  bool           `json:"has_repo"`
	Stats    map[string]int `json:"stats"`
}

// Health returns the current health status.
func (s *GitService) Health(ctx context.Context) HealthStatus {
	hasRepo := false
	if s.git != nil {
		hasRepo = s.git.IsRepo(ctx)
	}
	status := "ok"
	if !hasRepo {
		status = "no_repo"
	}
	repoPath := ""
	if s.git != nil {
		repoPath = s.git.RepoPath()
	}
	return HealthStatus{
		Status:   status,
		RepoPath: repoPath,
		HasRepo:  hasRepo,
		Stats:    s.store.Stats(),
	}
}

// ---------------------------------------------------------------------------
// register_push
// ---------------------------------------------------------------------------

// IngestPush registers a push that may span multiple files.  For each
// changed file it finds-or-creates the BranchFile, computes a diff,
// creates a DiffBlob and FileNode, and links everything under one PushSet.
func (s *GitService) IngestPush(_ context.Context, req models.IngestPushRequest) (*models.IngestPushResponse, error) {
	if req.BranchName == "" {
		return nil, fmt.Errorf("branch_name is required")
	}
	if req.UserID == "" {
		return nil, fmt.Errorf("user_id is required")
	}
	if len(req.Files) == 0 {
		return nil, fmt.Errorf("at least one file change is required")
	}

	branch, _ := s.store.FindOrCreateBranch(req.BranchName, uuid.New().String(), "origin")

	pushID := uuid.New().String()
	now := time.Now().UTC()

	push := &models.PushSet{
		PushID:    pushID,
		BranchID:  branch.BranchID,
		UserID:    req.UserID,
		Status:    models.PushStatusPending,
		CreatedAt: now,
		UpdatedAt: now,
	}

	var nodeIDs []string
	vfsFiles := make([]models.FileSnapshot, 0, len(req.Files))

	for _, fc := range req.Files {
		if fc.FilePath == "" {
			return nil, fmt.Errorf("file_path is required for every file change")
		}

		bf, _ := s.store.FindOrCreateBranchFile(
			branch.BranchID,
			fc.FilePath,
			uuid.New().String(),
			fc.BaseContent,
		)

		node, _, err := s.tree.AddNode(bf, req.UserID, pushID, fc.NewContent)
		if err != nil {
			return nil, fmt.Errorf("add node for %q: %w", fc.FilePath, err)
		}

		nodeIDs = append(nodeIDs, node.NodeID)
		vfsFiles = append(vfsFiles, models.FileSnapshot{
			Path:     fc.FilePath,
			Language: fc.Language,
			Content:  fc.NewContent,
		})
	}

	push.NodeIDs = nodeIDs
	s.store.PutPushSet(push)
	if s.vfs != nil {
		s.vfs.Propose(vfsAgentID(req.BranchName, req.UserID), pushID, vfsFiles)
	}

	s.logger.Info("push ingested",
		"push_id", pushID,
		"branch", req.BranchName,
		"user", req.UserID,
		"files", len(req.Files),
		"nodes", len(nodeIDs),
	)

	return &models.IngestPushResponse{
		PushID:  pushID,
		NodeIDs: nodeIDs,
	}, nil
}

// ---------------------------------------------------------------------------
// get_branch_file_state
// ---------------------------------------------------------------------------

// GetFileState returns the tracked state for one file on one branch.
func (s *GitService) GetFileState(branchName, filePath string) (*models.FileStateResponse, error) {
	branch, ok := s.store.GetBranchByName(branchName)
	if !ok {
		return nil, fmt.Errorf("branch %q not found", branchName)
	}

	bf, ok := s.store.FindBranchFile(branch.BranchID, filePath)
	if !ok {
		return nil, fmt.Errorf("file %q not tracked on branch %q", filePath, branchName)
	}

	activeNodes := s.tree.ActiveChildren(bf.BranchFileID)
	allNodes := s.tree.AllNodes(bf.BranchFileID)

	return &models.FileStateResponse{
		BranchFile:  *bf,
		ActiveNodes: derefNodes(activeNodes),
		TotalNodes:  len(allNodes),
	}, nil
}

// ---------------------------------------------------------------------------
// prepare_merge_context
// ---------------------------------------------------------------------------

// PrepareMergeContext returns all the data an external system needs to
// make a merge decision for one file on one branch.
//
// It does NOT decide anything.  It returns:
//   - the current base content
//   - all active diff nodes
//   - reconstructed candidate content for each node
//   - raw diff payloads
func (s *GitService) PrepareMergeContext(branchName, filePath string) (*models.MergeContextResponse, error) {
	branch, ok := s.store.GetBranchByName(branchName)
	if !ok {
		return nil, fmt.Errorf("branch %q not found", branchName)
	}

	bf, ok := s.store.FindBranchFile(branch.BranchID, filePath)
	if !ok {
		return nil, fmt.Errorf("file %q not tracked on branch %q", filePath, branchName)
	}

	activeNodes := s.tree.ActiveChildren(bf.BranchFileID)
	if len(activeNodes) == 0 {
		return nil, fmt.Errorf("no active diff nodes for %q on %q", filePath, branchName)
	}

	candidates := make([]models.MergeCandidate, 0, len(activeNodes))
	for _, node := range activeNodes {
		content, err := s.tree.ReconstructContent(bf, node)
		if err != nil {
			s.logger.Warn("failed to reconstruct node content",
				"node_id", node.NodeID, "error", err)
			continue
		}
		candidates = append(candidates, models.MergeCandidate{
			NodeID:               node.NodeID,
			UserID:               node.UserID,
			PushID:               node.PushID,
			ReconstructedContent: content,
		})
	}

	payloads, err := s.tree.CollectDiffPayloads(bf.BranchFileID)
	if err != nil {
		return nil, fmt.Errorf("collect diff payloads: %w", err)
	}

	return &models.MergeContextResponse{
		BranchFile:   *bf,
		BaseContent:  bf.BaseContent,
		ActiveNodes:  derefNodes(activeNodes),
		Candidates:   candidates,
		DiffPayloads: payloads,
	}, nil
}

// ---------------------------------------------------------------------------
// apply_merge_result
// ---------------------------------------------------------------------------

// ApplyMergeResult accepts an externally decided merge result, creates a
// merge node, promotes it to head, and supersedes the resolved nodes.
func (s *GitService) ApplyMergeResult(_ context.Context, req models.ApplyMergeResultRequest) (*models.ApplyMergeResultResponse, error) {
	if req.BranchName == "" || req.FilePath == "" || req.MergedContent == "" {
		return nil, fmt.Errorf("branch_name, file_path, and merged_content are required")
	}

	branch, ok := s.store.GetBranchByName(req.BranchName)
	if !ok {
		return nil, fmt.Errorf("branch %q not found", req.BranchName)
	}

	bf, ok := s.store.FindBranchFile(branch.BranchID, req.FilePath)
	if !ok {
		return nil, fmt.Errorf("file %q not tracked on branch %q", req.FilePath, req.BranchName)
	}

	// Create a merge-result node from the externally provided content.
	mergeNode, _, err := s.tree.AddNode(bf, "system", "", req.MergedContent)
	if err != nil {
		return nil, fmt.Errorf("create merge node: %w", err)
	}

	// Determine which nodes to supersede.
	var nodesAffected int
	if len(req.SupersededNodeIDs) > 0 {
		now := time.Now().UTC()
		for _, nid := range req.SupersededNodeIDs {
			n, exists := s.store.GetFileNode(nid)
			if !exists || n.BranchFileID != bf.BranchFileID {
				continue
			}
			n.Status = models.NodeStatusSuperseded
			n.UpdatedAt = now
			s.store.UpdateFileNode(n)
			nodesAffected++
		}
		mergeNode.Status = models.NodeStatusMerged
		mergeNode.UpdatedAt = now
		s.store.UpdateFileNode(mergeNode)
		if err := s.store.SetBranchFileHead(bf.BranchFileID, mergeNode.NodeID, req.MergedContent); err != nil {
			return nil, fmt.Errorf("set head: %w", err)
		}
	} else {
		// Supersede ALL active nodes via PromoteHead.
		activeNodes := s.tree.ActiveChildren(bf.BranchFileID)
		nodesAffected = len(activeNodes)
		if err := s.tree.PromoteHead(bf, mergeNode, req.MergedContent, models.NodeStatusSuperseded); err != nil {
			return nil, fmt.Errorf("promote merge head: %w", err)
		}
	}

	s.logger.Info("merge result applied",
		"branch", req.BranchName,
		"file", req.FilePath,
		"merged_node", mergeNode.NodeID,
		"nodes_affected", nodesAffected,
	)

	return &models.ApplyMergeResultResponse{
		MergedNodeID:  mergeNode.NodeID,
		NodesAffected: nodesAffected,
		BranchFileID:  bf.BranchFileID,
	}, nil
}

// ---------------------------------------------------------------------------
// prepare_commit
// ---------------------------------------------------------------------------

// PrepareCommit returns what would be included in a grouped commit for
// a push.  It checks whether all files in the push have been resolved
// (no remaining active nodes).  This is a dry run — it does not execute.
func (s *GitService) PrepareCommit(pushID string) (*models.PrepareCommitResponse, error) {
	push, ok := s.store.GetPushSet(pushID)
	if !ok {
		return nil, fmt.Errorf("push %q not found", pushID)
	}

	branch, ok := s.store.GetBranch(push.BranchID)
	if !ok {
		return nil, fmt.Errorf("branch %q not found for push %q", push.BranchID, pushID)
	}

	var files []models.CommitFileEntry
	var pendingFiles []string
	seenPaths := make(map[string]bool)

	for _, nodeID := range push.NodeIDs {
		node, exists := s.store.GetFileNode(nodeID)
		if !exists {
			continue
		}
		bf, exists := s.store.GetBranchFile(node.BranchFileID)
		if !exists {
			continue
		}
		if seenPaths[bf.FilePath] {
			continue
		}
		seenPaths[bf.FilePath] = true

		activeForFile := s.tree.ActiveChildren(bf.BranchFileID)
		if len(activeForFile) > 0 {
			pendingFiles = append(pendingFiles, bf.FilePath)
		}

		files = append(files, models.CommitFileEntry{
			FilePath: bf.FilePath,
			Content:  bf.BaseContent,
			NodeID:   nodeID,
		})
	}

	sort.Slice(files, func(i, j int) bool {
		return files[i].FilePath < files[j].FilePath
	})

	return &models.PrepareCommitResponse{
		PushID:       pushID,
		BranchName:   branch.GitBranchName,
		Files:        files,
		AllResolved:  len(pendingFiles) == 0,
		PendingFiles: pendingFiles,
	}, nil
}

// ---------------------------------------------------------------------------
// create_commit
// ---------------------------------------------------------------------------

// CreateGroupedCommit executes a grouped multi-file Git commit for a push.
// It writes files to the working tree, stages, commits, and records the
// result.  The caller is responsible for ensuring files are in the desired
// state (via apply_merge_result) before calling this.
func (s *GitService) CreateGroupedCommit(ctx context.Context, req models.GroupedCommitRequest) (*models.GroupedCommitResponse, error) {
	push, ok := s.store.GetPushSet(req.PushID)
	if !ok {
		return nil, fmt.Errorf("push %q not found", req.PushID)
	}

	branch, ok := s.store.GetBranch(push.BranchID)
	if !ok {
		return nil, fmt.Errorf("branch %q not found for push %q", push.BranchID, req.PushID)
	}

	type fileEntry struct {
		path    string
		content string
		nodeID  string
	}
	var files []fileEntry

	for _, nodeID := range push.NodeIDs {
		node, exists := s.store.GetFileNode(nodeID)
		if !exists {
			continue
		}
		bf, exists := s.store.GetBranchFile(node.BranchFileID)
		if !exists {
			continue
		}
		files = append(files, fileEntry{
			path:    bf.FilePath,
			content: bf.BaseContent,
			nodeID:  nodeID,
		})
	}

	// Deduplicate by file path.
	seen := make(map[string]int)
	var deduped []fileEntry
	for _, f := range files {
		if idx, exists := seen[f.path]; exists {
			deduped[idx] = f
		} else {
			seen[f.path] = len(deduped)
			deduped = append(deduped, f)
		}
	}
	files = deduped

	if len(files) == 0 {
		return nil, fmt.Errorf("no files to commit for push %q", req.PushID)
	}

	sort.Slice(files, func(i, j int) bool {
		return files[i].path < files[j].path
	})

	message := req.Message
	if message == "" {
		var paths []string
		for _, f := range files {
			paths = append(paths, f.path)
		}
		message = fmt.Sprintf("grouped commit for push %s: %s", req.PushID[:8], strings.Join(paths, ", "))
	}

	commitID := uuid.New().String()
	now := time.Now().UTC()

	var nodeIDsForCommit []string
	for _, f := range files {
		nodeIDsForCommit = append(nodeIDsForCommit, f.nodeID)
	}

	var gitSHA string

	if s.git != nil && s.git.IsRepo(ctx) {
		if err := s.git.CheckoutBranch(ctx, branch.GitBranchName); err != nil {
			return nil, fmt.Errorf("checkout branch %q: %w", branch.GitBranchName, err)
		}

		var stagePaths []string
		for _, f := range files {
			if err := s.git.WriteFile(f.path, f.content); err != nil {
				return nil, fmt.Errorf("write file %q: %w", f.path, err)
			}
			stagePaths = append(stagePaths, f.path)
		}

		if err := s.git.StageFiles(ctx, stagePaths); err != nil {
			return nil, fmt.Errorf("stage files: %w", err)
		}

		sha, err := s.git.Commit(ctx, message)
		if err != nil {
			record := &models.CommitRecord{
				CommitID:    commitID,
				BranchID:    push.BranchID,
				PushID:      push.PushID,
				Status:      models.CommitStatusPushFailed,
				Message:     message,
				FileNodeIDs: nodeIDsForCommit,
				CreatedAt:   now,
			}
			s.store.PutCommitRecord(record)
			return nil, fmt.Errorf("git commit: %w", err)
		}
		gitSHA = sha
	}

	record := &models.CommitRecord{
		CommitID:     commitID,
		BranchID:     push.BranchID,
		PushID:       push.PushID,
		GitCommitSHA: gitSHA,
		Status:       models.CommitStatusCreated,
		Message:      message,
		FileNodeIDs:  nodeIDsForCommit,
		CreatedAt:    now,
	}
	s.store.PutCommitRecord(record)

	push.Status = models.PushStatusCommitted
	push.UpdatedAt = now
	s.store.UpdatePushSet(push)
	if s.vfs != nil {
		s.vfs.Clear(vfsAgentID(branch.GitBranchName, push.UserID))
	}

	s.logger.Info("grouped commit created",
		"commit_id", commitID,
		"push_id", req.PushID,
		"git_sha", gitSHA,
		"files", len(files),
	)

	return &models.GroupedCommitResponse{
		CommitID:      commitID,
		GitCommitSHA:  gitSHA,
		FilesIncluded: len(files),
		Status:        string(models.CommitStatusCreated),
	}, nil
}

// ---------------------------------------------------------------------------
// push_commit
// ---------------------------------------------------------------------------

// PushCommit pushes a previously created commit to the remote.
func (s *GitService) PushCommit(ctx context.Context, req models.GitPushRequest) (*models.GitPushResponse, error) {
	record, ok := s.store.GetCommitRecord(req.CommitID)
	if !ok {
		return nil, fmt.Errorf("commit %q not found", req.CommitID)
	}

	if record.Status == models.CommitStatusPushed {
		return &models.GitPushResponse{
			CommitID: req.CommitID,
			Status:   string(models.CommitStatusPushed),
			Message:  "already pushed",
		}, nil
	}

	branch, ok := s.store.GetBranch(record.BranchID)
	if !ok {
		return nil, fmt.Errorf("branch %q not found for commit %q", record.BranchID, req.CommitID)
	}

	if s.git == nil {
		return nil, fmt.Errorf("git executor not configured")
	}

	remote := branch.RemoteName
	if remote == "" {
		remote = "origin"
	}

	err := s.git.PushWithRetry(ctx, remote, branch.GitBranchName, 2)
	now := time.Now().UTC()
	if err != nil {
		record.Status = models.CommitStatusPushFailed
		record.UpdatedAt = now
		s.store.UpdateCommitRecord(record)
		return &models.GitPushResponse{
			CommitID: req.CommitID,
			Status:   string(models.CommitStatusPushFailed),
			Message:  err.Error(),
		}, nil
	}

	record.Status = models.CommitStatusPushed
	record.PushedAt = &now
	record.UpdatedAt = now
	s.store.UpdateCommitRecord(record)

	s.logger.Info("commit pushed",
		"commit_id", req.CommitID,
		"branch", branch.GitBranchName,
		"remote", remote,
	)

	return &models.GitPushResponse{
		CommitID: req.CommitID,
		Status:   string(models.CommitStatusPushed),
		Message:  "push successful",
	}, nil
}

// ---------------------------------------------------------------------------
// get_commit_status
// ---------------------------------------------------------------------------

// GetCommitRecord returns a stored commit record by ID.
func (s *GitService) GetCommitRecord(commitID string) (*models.CommitRecord, error) {
	r, ok := s.store.GetCommitRecord(commitID)
	if !ok {
		return nil, fmt.Errorf("commit %q not found", commitID)
	}
	return r, nil
}

// ---------------------------------------------------------------------------
// Helpers
// ---------------------------------------------------------------------------

func derefNodes(ptrs []*models.FileNode) []models.FileNode {
	out := make([]models.FileNode, len(ptrs))
	for i, p := range ptrs {
		out[i] = *p
	}
	return out
}

func vfsAgentID(branchName, userID string) string {
	return "git:" + branchName + ":" + userID
}
