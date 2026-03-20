// Package vfs manages the in-memory virtual file system (shadow workspace).
//
// Every proposed file change from an agent is stored here until it is either
// committed (micro-commit) or discarded.  The VFS is the single source of
// truth for "what agents intend to write" before anything touches a real repo.
//
// For v1 the backing store is a sync.RWMutex-guarded map.  The interface is
// intentionally narrow so that a persistent backend (e.g. Redis, SQLite) can
// replace the map later without touching callers.
package vfs

import (
	"fmt"
	"sync"
	"time"

	"github.com/nccuhacks/nccu26/mcp/internal/models"
)

// Manager is the public API for VFS operations.
type Manager struct {
	mu       sync.RWMutex
	pending  map[string]*models.PendingChange // keyed by agent_id
}

// NewManager returns a ready-to-use VFS manager.
func NewManager() *Manager {
	return &Manager{
		pending: make(map[string]*models.PendingChange),
	}
}

// Propose stores or replaces the pending file set for an agent.
func (m *Manager) Propose(agentID, sessionID string, files []models.FileSnapshot) {
	m.mu.Lock()
	defer m.mu.Unlock()

	m.pending[agentID] = &models.PendingChange{
		AgentID:   agentID,
		SessionID: sessionID,
		Files:     files,
		CreatedAt: time.Now().UTC(),
	}
}

// AddFile appends a single file to an agent's pending set.
// Creates the pending entry if it does not exist yet.
func (m *Manager) AddFile(agentID, sessionID string, file models.FileSnapshot) {
	m.mu.Lock()
	defer m.mu.Unlock()

	pc, ok := m.pending[agentID]
	if !ok {
		pc = &models.PendingChange{
			AgentID:   agentID,
			SessionID: sessionID,
			CreatedAt: time.Now().UTC(),
		}
		m.pending[agentID] = pc
	}
	pc.Files = append(pc.Files, file)
}

// State returns a snapshot of the entire VFS.
func (m *Manager) State() models.VFSState {
	m.mu.RLock()
	defer m.mu.RUnlock()

	changes := make([]models.PendingChange, 0, len(m.pending))
	totalFiles := 0
	for _, pc := range m.pending {
		changes = append(changes, *pc)
		totalFiles += len(pc.Files)
	}

	return models.VFSState{
		PendingChanges: changes,
		TotalFiles:     totalFiles,
		TotalAgents:    len(m.pending),
	}
}

// ChangeSetsForAnalysis converts the current VFS into the shape expected
// by the backend's /analyze/overlaps endpoint.
func (m *Manager) ChangeSetsForAnalysis() []models.AnalysisChangeSet {
	m.mu.RLock()
	defer m.mu.RUnlock()

	out := make([]models.AnalysisChangeSet, 0, len(m.pending))
	for _, pc := range m.pending {
		files := make([]models.AnalysisFileInput, len(pc.Files))
		for i, f := range pc.Files {
			files[i] = models.AnalysisFileInput{
				Path:     f.Path,
				Language: f.Language,
				Content:  f.Content,
			}
		}
		out = append(out, models.AnalysisChangeSet{
			AgentID: pc.AgentID,
			Files:   files,
		})
	}
	return out
}

// FilesForAgent returns the pending files belonging to a specific agent,
// or an error if no pending changes exist for that agent.
func (m *Manager) FilesForAgent(agentID string) ([]models.FileSnapshot, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()

	pc, ok := m.pending[agentID]
	if !ok {
		return nil, fmt.Errorf("no pending changes for agent %q", agentID)
	}
	cp := make([]models.FileSnapshot, len(pc.Files))
	copy(cp, pc.Files)
	return cp, nil
}

// Clear removes an agent's pending changes (e.g. after a successful commit).
func (m *Manager) Clear(agentID string) {
	m.mu.Lock()
	defer m.mu.Unlock()
	delete(m.pending, agentID)
}

// ClearAll wipes the entire VFS.  Primarily useful in tests.
func (m *Manager) ClearAll() {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.pending = make(map[string]*models.PendingChange)
}
