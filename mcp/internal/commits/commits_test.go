package commits

import (
	"testing"

	"github.com/nccuhacks/nccu26/mcp/internal/models"
)

func TestCommitReturnsStructuredResult(t *testing.T) {
	c := NewCoordinator()
	result := c.Commit(models.CommitRequest{
		AgentID: "agent-1",
		Files:   []models.FileSnapshot{{Path: "f.py", Language: "python", Content: "x"}},
		Message: "test commit",
	})
	if !result.Allowed {
		t.Fatal("stub commit should always succeed")
	}
	if result.CommitID == "" {
		t.Fatal("expected non-empty commit ID")
	}
	if result.Reason == "" {
		t.Fatal("expected non-empty reason")
	}
}

func TestCommitIDsAreUnique(t *testing.T) {
	c := NewCoordinator()
	ids := map[string]bool{}
	for range 100 {
		r := c.Commit(models.CommitRequest{AgentID: "a", Message: "m"})
		if ids[r.CommitID] {
			t.Fatalf("duplicate commit ID: %s", r.CommitID)
		}
		ids[r.CommitID] = true
	}
}

func TestCommitWithEmptyFiles(t *testing.T) {
	c := NewCoordinator()
	result := c.Commit(models.CommitRequest{AgentID: "a", Files: nil, Message: "empty"})
	if !result.Allowed {
		t.Fatal("empty commit should still succeed in stub")
	}
}
