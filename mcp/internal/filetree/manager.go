// Package filetree manages the per-file diff tree lifecycle.
//
// Each file on each branch has its own tree of proposed diffs.  This
// package provides the logic for:
//   - looking up the current head/root for a file
//   - listing active child nodes
//   - promoting a merge result to head
//   - transitioning node statuses
//
// It operates on the shared storage layer and does not own any state itself.
package filetree

import (
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/nccuhacks/nccu26/mcp/internal/diff"
	"github.com/nccuhacks/nccu26/mcp/internal/models"
	"github.com/nccuhacks/nccu26/mcp/internal/storage"
)

// Manager provides operations on per-file diff trees.
type Manager struct {
	store  *storage.Store
	engine *diff.Engine
}

// NewManager creates a Manager backed by the given store and diff engine.
func NewManager(store *storage.Store, engine *diff.Engine) *Manager {
	return &Manager{store: store, engine: engine}
}

// ActiveChildren returns all active FileNodes for a BranchFile.
func (m *Manager) ActiveChildren(branchFileID string) []*models.FileNode {
	return m.store.ActiveNodesForFile(branchFileID)
}

// AllNodes returns every FileNode for a BranchFile.
func (m *Manager) AllNodes(branchFileID string) []*models.FileNode {
	return m.store.AllNodesForFile(branchFileID)
}

// ReconstructContent applies a node's diff to the file's base content
// and returns the proposed file content.
func (m *Manager) ReconstructContent(bf *models.BranchFile, node *models.FileNode) (string, error) {
	blob, ok := m.store.GetDiffBlob(node.DiffBlobID)
	if !ok {
		return "", fmt.Errorf("diff blob %q not found for node %q", node.DiffBlobID, node.NodeID)
	}
	return m.engine.ApplyPatch(bf.BaseContent, blob.Payload)
}

// AddNode creates a new active FileNode for a BranchFile from a proposed
// content change.  Returns the created node and diff blob.
func (m *Manager) AddNode(bf *models.BranchFile, userID, pushID, newContent string) (*models.FileNode, *models.DiffBlob, error) {
	if !m.engine.HasChanges(bf.BaseContent, newContent) {
		return nil, nil, fmt.Errorf("no changes between base and proposed content for %q", bf.FilePath)
	}

	patchText := m.engine.CreatePatch(bf.BaseContent, newContent)

	now := time.Now().UTC()

	blob := &models.DiffBlob{
		DiffBlobID:  uuid.New().String(),
		Format:      diff.Format,
		Payload:     patchText,
		ContentHash: diff.ContentHash(newContent),
		CreatedAt:   now,
	}
	m.store.PutDiffBlob(blob)

	node := &models.FileNode{
		NodeID:       uuid.New().String(),
		BranchFileID: bf.BranchFileID,
		ParentNodeID: bf.CurrentHeadNodeID,
		UserID:       userID,
		PushID:       pushID,
		DiffBlobID:   blob.DiffBlobID,
		Status:       models.NodeStatusActive,
		CreatedAt:    now,
		UpdatedAt:    now,
	}
	m.store.PutFileNode(node)

	return node, blob, nil
}

// PromoteHead sets a merge-result node as the new head and updates the
// base content. All previously active sibling nodes are transitioned to
// the given supersededStatus (typically "merged" or "superseded").
func (m *Manager) PromoteHead(
	bf *models.BranchFile,
	mergeNode *models.FileNode,
	mergedContent string,
	supersededStatus models.NodeStatus,
) error {
	activeNodes := m.store.ActiveNodesForFile(bf.BranchFileID)

	now := time.Now().UTC()
	for _, n := range activeNodes {
		if n.NodeID == mergeNode.NodeID {
			continue
		}
		n.Status = supersededStatus
		n.UpdatedAt = now
		m.store.UpdateFileNode(n)
	}

	mergeNode.Status = models.NodeStatusMerged
	mergeNode.UpdatedAt = now
	m.store.UpdateFileNode(mergeNode)

	return m.store.SetBranchFileHead(bf.BranchFileID, mergeNode.NodeID, mergedContent)
}

// CollectDiffPayloads returns the raw patch payloads for all active nodes
// on a BranchFile, in creation order.
func (m *Manager) CollectDiffPayloads(branchFileID string) ([]string, error) {
	nodes := m.store.ActiveNodesForFile(branchFileID)
	payloads := make([]string, 0, len(nodes))
	for _, n := range nodes {
		blob, ok := m.store.GetDiffBlob(n.DiffBlobID)
		if !ok {
			return nil, fmt.Errorf("diff blob %q missing for node %q", n.DiffBlobID, n.NodeID)
		}
		payloads = append(payloads, blob.Payload)
	}
	return payloads, nil
}
