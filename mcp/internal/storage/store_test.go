package storage

import (
	"testing"

	"github.com/nccuhacks/nccu26/mcp/internal/models"
)

func TestFindOrCreateBranch(t *testing.T) {
	s := New()

	b1, created := s.FindOrCreateBranch("feature/auth", "id-1", "origin")
	if !created {
		t.Error("expected branch to be created")
	}
	if b1.GitBranchName != "feature/auth" {
		t.Errorf("branch name = %q", b1.GitBranchName)
	}

	b2, created2 := s.FindOrCreateBranch("feature/auth", "id-2", "origin")
	if created2 {
		t.Error("expected branch to already exist")
	}
	if b2.BranchID != b1.BranchID {
		t.Error("expected same branch ID")
	}
}

func TestFindOrCreateBranchFile(t *testing.T) {
	s := New()

	bf1, created := s.FindOrCreateBranchFile("branch-1", "auth.py", "bf-1", "base")
	if !created {
		t.Error("expected branch file to be created")
	}

	bf2, created2 := s.FindOrCreateBranchFile("branch-1", "auth.py", "bf-2", "new base")
	if created2 {
		t.Error("expected branch file to already exist")
	}
	if bf2.BranchFileID != bf1.BranchFileID {
		t.Error("expected same branch file ID")
	}
	if bf2.BaseContent != "base" {
		t.Error("base content should not change on duplicate create")
	}

	// Different branch, same file path → separate BranchFile.
	bf3, created3 := s.FindOrCreateBranchFile("branch-2", "auth.py", "bf-3", "other base")
	if !created3 {
		t.Error("expected different branch file")
	}
	if bf3.BranchFileID == bf1.BranchFileID {
		t.Error("same path on different branch should have different ID")
	}
}

func TestActiveNodesForFile(t *testing.T) {
	s := New()

	active := &models.FileNode{
		NodeID:       "n1",
		BranchFileID: "bf-1",
		Status:       models.NodeStatusActive,
	}
	merged := &models.FileNode{
		NodeID:       "n2",
		BranchFileID: "bf-1",
		Status:       models.NodeStatusMerged,
	}
	s.PutFileNode(active)
	s.PutFileNode(merged)

	nodes := s.ActiveNodesForFile("bf-1")
	if len(nodes) != 1 {
		t.Fatalf("got %d active nodes, want 1", len(nodes))
	}
	if nodes[0].NodeID != "n1" {
		t.Errorf("active node = %q, want n1", nodes[0].NodeID)
	}
}

func TestNodesForPush(t *testing.T) {
	s := New()

	s.PutFileNode(&models.FileNode{NodeID: "n1", BranchFileID: "bf-1", PushID: "push-1", Status: models.NodeStatusActive})
	s.PutFileNode(&models.FileNode{NodeID: "n2", BranchFileID: "bf-2", PushID: "push-1", Status: models.NodeStatusActive})
	s.PutFileNode(&models.FileNode{NodeID: "n3", BranchFileID: "bf-1", PushID: "push-2", Status: models.NodeStatusActive})

	nodes := s.NodesForPush("push-1")
	if len(nodes) != 2 {
		t.Errorf("got %d nodes for push-1, want 2", len(nodes))
	}
}

func TestSetBranchFileHead(t *testing.T) {
	s := New()
	s.FindOrCreateBranchFile("b1", "f.py", "bf-1", "old content")

	err := s.SetBranchFileHead("bf-1", "node-1", "new content")
	if err != nil {
		t.Fatalf("SetBranchFileHead: %v", err)
	}

	bf, ok := s.GetBranchFile("bf-1")
	if !ok {
		t.Fatal("branch file not found")
	}
	if bf.CurrentHeadNodeID != "node-1" {
		t.Errorf("head = %q, want node-1", bf.CurrentHeadNodeID)
	}
	if bf.BaseContent != "new content" {
		t.Errorf("base content = %q", bf.BaseContent)
	}

	err = s.SetBranchFileHead("nonexistent", "n", "c")
	if err == nil {
		t.Error("expected error for missing branch file")
	}
}

func TestStats(t *testing.T) {
	s := New()
	stats := s.Stats()
	for _, key := range []string{"branches", "branch_files", "file_nodes", "diff_blobs", "push_sets", "commits"} {
		if _, ok := stats[key]; !ok {
			t.Errorf("missing stat key: %s", key)
		}
	}
}
