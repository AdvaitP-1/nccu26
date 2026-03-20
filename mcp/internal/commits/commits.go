// Package commits coordinates the micro-commit flow.
//
// For v1 the actual git integration is intentionally stubbed.  The
// interface and models are real and production-shaped so that a proper
// git implementation (or a call to an external commit service) can be
// dropped in without changing callers.
//
// Stub behaviour:
//   - Generates a deterministic pseudo commit-ID from the agent ID + timestamp.
//   - Returns success immediately — no disk or network I/O.
//
// When real git integration is added, replace the body of (*Coordinator).Commit
// and leave the rest untouched.
package commits

import (
	"crypto/sha256"
	"fmt"
	"log/slog"
	"time"

	"github.com/nccuhacks/nccu26/mcp/internal/models"
)

// Coordinator manages the micro-commit lifecycle.
type Coordinator struct {
	logger *slog.Logger
}

// NewCoordinator builds a ready-to-use Coordinator.
func NewCoordinator() *Coordinator {
	return &Coordinator{
		logger: slog.Default().With("component", "commits"),
	}
}

// Commit attempts to persist a set of approved changes.
//
// The caller is responsible for having already run policy evaluation;
// this method does NOT re-check safety.
func (c *Coordinator) Commit(req models.CommitRequest) models.CommitResult {
	// STUB: generate a synthetic commit ID.
	// Real implementation would stage files, create a git commit, etc.
	commitID := pseudoCommitID(req.AgentID)

	c.logger.Info("micro-commit recorded (stub)",
		"agent_id", req.AgentID,
		"files", len(req.Files),
		"commit_id", commitID,
		"message", req.Message,
	)

	return models.CommitResult{
		Allowed:  true,
		CommitID: commitID,
		Reason:   "commit accepted (stub — real git integration pending)",
	}
}

func pseudoCommitID(agentID string) string {
	h := sha256.Sum256([]byte(fmt.Sprintf("%s:%d", agentID, time.Now().UnixNano())))
	return fmt.Sprintf("%x", h[:8])
}
