// Package agents provides a thread-safe registry of known agent metadata.
//
// The registry is the source of truth for agent types, display names, and
// connection status.  It is populated when agents register pushes through
// VFS or via explicit registration from the dashboard.
package agents

import (
	"sync"
	"time"
)

// Type classifies an agent's role in the orchestration topology.
type Type string

const (
	TypeManager  Type = "manager"
	TypeCoder    Type = "coder"
	TypeMerge    Type = "merge"
	TypeReviewer Type = "reviewer"
	TypeExternal Type = "external"
)

// Status represents the current lifecycle state of an agent.
type Status string

const (
	StatusActive       Status = "active"
	StatusIdle         Status = "idle"
	StatusDisconnected Status = "disconnected"
)

// Info holds all metadata for a single registered agent.
type Info struct {
	ID          string            `json:"id"`
	Type        Type              `json:"type"`
	DisplayName string            `json:"display_name"`
	Status      Status            `json:"status"`
	Meta        map[string]string `json:"meta,omitempty"`
	ConnectedAt time.Time         `json:"connected_at"`
	LastSeenAt  time.Time         `json:"last_seen_at"`
}

// Registry is a thread-safe store of agent metadata, independent of VFS.
type Registry struct {
	mu     sync.RWMutex
	agents map[string]*Info
}

// NewRegistry creates a ready-to-use empty registry.
func NewRegistry() *Registry {
	return &Registry{agents: make(map[string]*Info)}
}

// Register adds or updates an agent entry.  If the agent already exists,
// its type, display name, and status are updated in place.
func (r *Registry) Register(id string, agentType Type, displayName string, meta map[string]string) {
	r.mu.Lock()
	defer r.mu.Unlock()

	now := time.Now().UTC()
	if existing, ok := r.agents[id]; ok {
		existing.Type = agentType
		existing.DisplayName = displayName
		existing.Status = StatusActive
		existing.LastSeenAt = now
		if meta != nil {
			existing.Meta = meta
		}
		return
	}

	r.agents[id] = &Info{
		ID:          id,
		Type:        agentType,
		DisplayName: displayName,
		Status:      StatusActive,
		Meta:        meta,
		ConnectedAt: now,
		LastSeenAt:  now,
	}
}

// Touch updates LastSeenAt and marks the agent active.
func (r *Registry) Touch(id string) {
	r.mu.Lock()
	defer r.mu.Unlock()
	if a, ok := r.agents[id]; ok {
		a.LastSeenAt = time.Now().UTC()
		a.Status = StatusActive
	}
}

// SetStatus explicitly changes an agent's status.
func (r *Registry) SetStatus(id string, status Status) {
	r.mu.Lock()
	defer r.mu.Unlock()
	if a, ok := r.agents[id]; ok {
		a.Status = status
	}
}

// Get returns a copy of the agent info, or nil if not found.
func (r *Registry) Get(id string) *Info {
	r.mu.RLock()
	defer r.mu.RUnlock()
	if a, ok := r.agents[id]; ok {
		cp := *a
		return &cp
	}
	return nil
}

// All returns a snapshot of every registered agent.
func (r *Registry) All() []Info {
	r.mu.RLock()
	defer r.mu.RUnlock()
	out := make([]Info, 0, len(r.agents))
	for _, a := range r.agents {
		out = append(out, *a)
	}
	return out
}

// Remove deletes an agent from the registry.
func (r *Registry) Remove(id string) {
	r.mu.Lock()
	defer r.mu.Unlock()
	delete(r.agents, id)
}

// Clear wipes the entire registry.
func (r *Registry) Clear() {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.agents = make(map[string]*Info)
}

// Count returns the number of registered agents.
func (r *Registry) Count() int {
	r.mu.RLock()
	defer r.mu.RUnlock()
	return len(r.agents)
}
