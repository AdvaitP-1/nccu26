// Package storage provides an in-memory store for the branch-scoped
// per-file diff tree system.
//
// The backing store uses sync.RWMutex-guarded maps, matching the pattern
// established by the VFS package.  A persistent backend (SQLite, Postgres)
// can replace the maps later without touching callers.
package storage

import (
	"fmt"
	"sync"
	"time"

	"github.com/nccuhacks/nccu26/mcp/internal/models"
)

// Store is the central in-memory repository for git orchestration state.
type Store struct {
	mu sync.RWMutex

	branches    map[string]*models.Branch     // keyed by branch_id
	branchFiles map[string]*models.BranchFile // keyed by branch_file_id
	fileNodes   map[string]*models.FileNode   // keyed by node_id
	diffBlobs   map[string]*models.DiffBlob   // keyed by diff_blob_id
	pushSets    map[string]*models.PushSet    // keyed by push_id
	commits     map[string]*models.CommitRecord // keyed by commit_id

	// Secondary indices for efficient lookups.
	branchByName      map[string]string            // git_branch_name → branch_id
	branchFileByKey   map[string]string            // "branch_id:file_path" → branch_file_id
	nodesByBranchFile map[string][]string           // branch_file_id → []node_id
	nodesByPush       map[string][]string           // push_id → []node_id
}

// New creates a ready-to-use Store.
func New() *Store {
	return &Store{
		branches:          make(map[string]*models.Branch),
		branchFiles:       make(map[string]*models.BranchFile),
		fileNodes:         make(map[string]*models.FileNode),
		diffBlobs:         make(map[string]*models.DiffBlob),
		pushSets:          make(map[string]*models.PushSet),
		commits:           make(map[string]*models.CommitRecord),
		branchByName:      make(map[string]string),
		branchFileByKey:   make(map[string]string),
		nodesByBranchFile: make(map[string][]string),
		nodesByPush:       make(map[string][]string),
	}
}

func branchFileKey(branchID, filePath string) string {
	return branchID + ":" + filePath
}

// ---------------------------------------------------------------------------
// Branch
// ---------------------------------------------------------------------------

func (s *Store) PutBranch(b *models.Branch) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.branches[b.BranchID] = b
	s.branchByName[b.GitBranchName] = b.BranchID
}

func (s *Store) GetBranch(id string) (*models.Branch, bool) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	b, ok := s.branches[id]
	return b, ok
}

func (s *Store) GetBranchByName(name string) (*models.Branch, bool) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	id, ok := s.branchByName[name]
	if !ok {
		return nil, false
	}
	b, ok := s.branches[id]
	return b, ok
}

// ---------------------------------------------------------------------------
// BranchFile
// ---------------------------------------------------------------------------

func (s *Store) PutBranchFile(bf *models.BranchFile) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.branchFiles[bf.BranchFileID] = bf
	s.branchFileByKey[branchFileKey(bf.BranchID, bf.FilePath)] = bf.BranchFileID
}

func (s *Store) GetBranchFile(id string) (*models.BranchFile, bool) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	bf, ok := s.branchFiles[id]
	return bf, ok
}

func (s *Store) FindBranchFile(branchID, filePath string) (*models.BranchFile, bool) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	id, ok := s.branchFileByKey[branchFileKey(branchID, filePath)]
	if !ok {
		return nil, false
	}
	bf, ok := s.branchFiles[id]
	return bf, ok
}

// ListBranchFiles returns all BranchFile entries for a given branch.
func (s *Store) ListBranchFiles(branchID string) []*models.BranchFile {
	s.mu.RLock()
	defer s.mu.RUnlock()
	var out []*models.BranchFile
	for _, bf := range s.branchFiles {
		if bf.BranchID == branchID {
			out = append(out, bf)
		}
	}
	return out
}

// ---------------------------------------------------------------------------
// FileNode
// ---------------------------------------------------------------------------

func (s *Store) PutFileNode(n *models.FileNode) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.fileNodes[n.NodeID] = n
	s.nodesByBranchFile[n.BranchFileID] = append(s.nodesByBranchFile[n.BranchFileID], n.NodeID)
	if n.PushID != "" {
		s.nodesByPush[n.PushID] = append(s.nodesByPush[n.PushID], n.NodeID)
	}
}

func (s *Store) GetFileNode(id string) (*models.FileNode, bool) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	n, ok := s.fileNodes[id]
	return n, ok
}

// UpdateFileNode replaces a node in place. Caller is responsible for
// setting UpdatedAt.
func (s *Store) UpdateFileNode(n *models.FileNode) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.fileNodes[n.NodeID] = n
}

// ActiveNodesForFile returns all nodes with status "active" for a BranchFile.
func (s *Store) ActiveNodesForFile(branchFileID string) []*models.FileNode {
	s.mu.RLock()
	defer s.mu.RUnlock()
	var out []*models.FileNode
	for _, nid := range s.nodesByBranchFile[branchFileID] {
		n := s.fileNodes[nid]
		if n != nil && n.Status == models.NodeStatusActive {
			out = append(out, n)
		}
	}
	return out
}

// AllNodesForFile returns every node for a BranchFile regardless of status.
func (s *Store) AllNodesForFile(branchFileID string) []*models.FileNode {
	s.mu.RLock()
	defer s.mu.RUnlock()
	var out []*models.FileNode
	for _, nid := range s.nodesByBranchFile[branchFileID] {
		if n, ok := s.fileNodes[nid]; ok {
			out = append(out, n)
		}
	}
	return out
}

// NodesForPush returns all nodes created by a given push.
func (s *Store) NodesForPush(pushID string) []*models.FileNode {
	s.mu.RLock()
	defer s.mu.RUnlock()
	var out []*models.FileNode
	for _, nid := range s.nodesByPush[pushID] {
		if n, ok := s.fileNodes[nid]; ok {
			out = append(out, n)
		}
	}
	return out
}

// ---------------------------------------------------------------------------
// DiffBlob
// ---------------------------------------------------------------------------

func (s *Store) PutDiffBlob(d *models.DiffBlob) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.diffBlobs[d.DiffBlobID] = d
}

func (s *Store) GetDiffBlob(id string) (*models.DiffBlob, bool) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	d, ok := s.diffBlobs[id]
	return d, ok
}

// ---------------------------------------------------------------------------
// PushSet
// ---------------------------------------------------------------------------

func (s *Store) PutPushSet(p *models.PushSet) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.pushSets[p.PushID] = p
}

func (s *Store) GetPushSet(id string) (*models.PushSet, bool) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	p, ok := s.pushSets[id]
	return p, ok
}

func (s *Store) UpdatePushSet(p *models.PushSet) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.pushSets[p.PushID] = p
}

// ---------------------------------------------------------------------------
// CommitRecord
// ---------------------------------------------------------------------------

func (s *Store) PutCommitRecord(c *models.CommitRecord) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.commits[c.CommitID] = c
}

func (s *Store) GetCommitRecord(id string) (*models.CommitRecord, bool) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	c, ok := s.commits[id]
	return c, ok
}

func (s *Store) UpdateCommitRecord(c *models.CommitRecord) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.commits[c.CommitID] = c
}

// MergedNodesForPush returns all merged/active nodes belonging to a push,
// grouped by branch_file_id. Used during grouped commit assembly.
func (s *Store) MergedNodesForPush(pushID string) map[string]*models.FileNode {
	s.mu.RLock()
	defer s.mu.RUnlock()
	out := make(map[string]*models.FileNode)
	for _, nid := range s.nodesByPush[pushID] {
		n := s.fileNodes[nid]
		if n == nil {
			continue
		}
		if n.Status == models.NodeStatusMerged || n.Status == models.NodeStatusActive {
			out[n.BranchFileID] = n
		}
	}
	return out
}

// Stats returns summary counts for health/debug endpoints.
func (s *Store) Stats() map[string]int {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return map[string]int{
		"branches":     len(s.branches),
		"branch_files": len(s.branchFiles),
		"file_nodes":   len(s.fileNodes),
		"diff_blobs":   len(s.diffBlobs),
		"push_sets":    len(s.pushSets),
		"commits":      len(s.commits),
	}
}

// FindOrCreateBranch finds a branch by name or creates it with sensible
// defaults. Returns the branch and whether it was newly created.
func (s *Store) FindOrCreateBranch(name, branchID, remoteName string) (*models.Branch, bool) {
	s.mu.Lock()
	defer s.mu.Unlock()

	if existingID, ok := s.branchByName[name]; ok {
		return s.branches[existingID], false
	}

	now := time.Now().UTC()
	b := &models.Branch{
		BranchID:      branchID,
		GitBranchName: name,
		RemoteName:    remoteName,
		CreatedAt:     now,
		UpdatedAt:     now,
	}
	s.branches[b.BranchID] = b
	s.branchByName[name] = b.BranchID
	return b, true
}

// FindOrCreateBranchFile finds a BranchFile by branch+path or creates it.
func (s *Store) FindOrCreateBranchFile(branchID, filePath, bfID, baseContent string) (*models.BranchFile, bool) {
	s.mu.Lock()
	defer s.mu.Unlock()

	key := branchFileKey(branchID, filePath)
	if existingID, ok := s.branchFileByKey[key]; ok {
		return s.branchFiles[existingID], false
	}

	now := time.Now().UTC()
	bf := &models.BranchFile{
		BranchFileID: bfID,
		BranchID:     branchID,
		FilePath:     filePath,
		BaseContent:  baseContent,
		CreatedAt:    now,
		UpdatedAt:    now,
	}
	s.branchFiles[bf.BranchFileID] = bf
	s.branchFileByKey[key] = bf.BranchFileID
	return bf, true
}

// SetBranchFileHead updates the head node and base content for a BranchFile.
func (s *Store) SetBranchFileHead(bfID, headNodeID, newBaseContent string) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	bf, ok := s.branchFiles[bfID]
	if !ok {
		return fmt.Errorf("branch file %q not found", bfID)
	}
	bf.CurrentHeadNodeID = headNodeID
	bf.BaseContent = newBaseContent
	bf.UpdatedAt = time.Now().UTC()
	return nil
}
