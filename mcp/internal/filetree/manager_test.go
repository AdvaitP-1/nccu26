package filetree

import (
	"testing"

	"github.com/nccuhacks/nccu26/mcp/internal/diff"
	"github.com/nccuhacks/nccu26/mcp/internal/models"
	"github.com/nccuhacks/nccu26/mcp/internal/storage"
)

func setup() (*Manager, *storage.Store) {
	store := storage.New()
	engine := diff.NewEngine()
	mgr := NewManager(store, engine)
	return mgr, store
}

func makeBranchFile(store *storage.Store, branchID, filePath, baseContent string) *models.BranchFile {
	bf, _ := store.FindOrCreateBranchFile(branchID, filePath, "bf-"+filePath, baseContent)
	return bf
}

func TestAddNode_CreatesNodeAndBlob(t *testing.T) {
	mgr, store := setup()
	bf := makeBranchFile(store, "branch-1", "auth.py", "def login():\n    pass\n")

	node, blob, err := mgr.AddNode(bf, "dev-A", "push-1", "def login():\n    return True\n")
	if err != nil {
		t.Fatalf("AddNode: %v", err)
	}

	if node.NodeID == "" {
		t.Error("expected non-empty node ID")
	}
	if node.BranchFileID != bf.BranchFileID {
		t.Errorf("node.BranchFileID = %q, want %q", node.BranchFileID, bf.BranchFileID)
	}
	if node.Status != models.NodeStatusActive {
		t.Errorf("node.Status = %q, want %q", node.Status, models.NodeStatusActive)
	}
	if node.PushID != "push-1" {
		t.Errorf("node.PushID = %q, want %q", node.PushID, "push-1")
	}
	if blob.DiffBlobID == "" {
		t.Error("expected non-empty blob ID")
	}
	if blob.Format != diff.Format {
		t.Errorf("blob.Format = %q, want %q", blob.Format, diff.Format)
	}

	// Verify the blob was persisted.
	retrieved, ok := store.GetDiffBlob(blob.DiffBlobID)
	if !ok {
		t.Fatal("blob not found in store")
	}
	if retrieved.Payload != blob.Payload {
		t.Error("persisted blob payload mismatch")
	}
}

func TestAddNode_NoChange_ReturnsError(t *testing.T) {
	mgr, store := setup()
	base := "unchanged\n"
	bf := makeBranchFile(store, "branch-1", "file.txt", base)

	_, _, err := mgr.AddNode(bf, "dev-A", "push-1", base)
	if err == nil {
		t.Error("expected error when content is unchanged")
	}
}

func TestReconstructContent(t *testing.T) {
	mgr, store := setup()
	base := "original\n"
	bf := makeBranchFile(store, "branch-1", "file.txt", base)

	proposed := "modified\n"
	node, _, err := mgr.AddNode(bf, "dev-A", "push-1", proposed)
	if err != nil {
		t.Fatalf("AddNode: %v", err)
	}

	reconstructed, err := mgr.ReconstructContent(bf, node)
	if err != nil {
		t.Fatalf("ReconstructContent: %v", err)
	}
	if reconstructed != proposed {
		t.Errorf("got %q, want %q", reconstructed, proposed)
	}
}

func TestActiveChildren(t *testing.T) {
	mgr, store := setup()
	bf := makeBranchFile(store, "branch-1", "file.py", "base\n")

	mgr.AddNode(bf, "dev-A", "push-1", "change-A\n")
	mgr.AddNode(bf, "dev-B", "push-2", "change-B\n")
	mgr.AddNode(bf, "dev-C", "push-3", "change-C\n")

	active := mgr.ActiveChildren(bf.BranchFileID)
	if len(active) != 3 {
		t.Errorf("got %d active children, want 3", len(active))
	}
}

func TestPromoteHead(t *testing.T) {
	mgr, store := setup()
	base := "base content\n"
	bf := makeBranchFile(store, "branch-1", "file.py", base)

	nodeA, _, _ := mgr.AddNode(bf, "dev-A", "push-1", "change-A\n")
	nodeB, _, _ := mgr.AddNode(bf, "dev-B", "push-2", "change-B\n")

	mergedContent := "merged content\n"
	mergeNode, _, _ := mgr.AddNode(bf, "system", "", mergedContent)

	err := mgr.PromoteHead(bf, mergeNode, mergedContent, models.NodeStatusSuperseded)
	if err != nil {
		t.Fatalf("PromoteHead: %v", err)
	}

	// Verify head was updated.
	updatedBF, ok := store.GetBranchFile(bf.BranchFileID)
	if !ok {
		t.Fatal("branch file not found after promote")
	}
	if updatedBF.CurrentHeadNodeID != mergeNode.NodeID {
		t.Errorf("head = %q, want %q", updatedBF.CurrentHeadNodeID, mergeNode.NodeID)
	}
	if updatedBF.BaseContent != mergedContent {
		t.Errorf("base content not updated")
	}

	// Verify siblings were superseded.
	updatedA, _ := store.GetFileNode(nodeA.NodeID)
	if updatedA.Status != models.NodeStatusSuperseded {
		t.Errorf("nodeA.Status = %q, want superseded", updatedA.Status)
	}
	updatedB, _ := store.GetFileNode(nodeB.NodeID)
	if updatedB.Status != models.NodeStatusSuperseded {
		t.Errorf("nodeB.Status = %q, want superseded", updatedB.Status)
	}

	// Verify merge node is marked as merged.
	updatedMerge, _ := store.GetFileNode(mergeNode.NodeID)
	if updatedMerge.Status != models.NodeStatusMerged {
		t.Errorf("mergeNode.Status = %q, want merged", updatedMerge.Status)
	}

	// No active children remain.
	active := mgr.ActiveChildren(bf.BranchFileID)
	if len(active) != 0 {
		t.Errorf("got %d active children after promote, want 0", len(active))
	}
}

func TestCollectDiffPayloads(t *testing.T) {
	mgr, store := setup()
	bf := makeBranchFile(store, "branch-1", "file.py", "base\n")

	mgr.AddNode(bf, "dev-A", "push-1", "change-A\n")
	mgr.AddNode(bf, "dev-B", "push-2", "change-B\n")

	payloads, err := mgr.CollectDiffPayloads(bf.BranchFileID)
	if err != nil {
		t.Fatalf("CollectDiffPayloads: %v", err)
	}
	if len(payloads) != 2 {
		t.Errorf("got %d payloads, want 2", len(payloads))
	}
	for i, p := range payloads {
		if p == "" {
			t.Errorf("payload %d is empty", i)
		}
	}
}

func TestAddNode_SetsParentToCurrentHead(t *testing.T) {
	mgr, store := setup()
	bf := makeBranchFile(store, "branch-1", "file.py", "base\n")

	nodeA, _, _ := mgr.AddNode(bf, "dev-A", "push-1", "v1\n")
	if nodeA.ParentNodeID != "" {
		t.Errorf("first node parent = %q, want empty", nodeA.ParentNodeID)
	}

	// Simulate a merge promoting nodeA as head.
	mgr.PromoteHead(bf, nodeA, "v1\n", models.NodeStatusSuperseded)

	// Refresh bf from store.
	bf, _ = store.GetBranchFile(bf.BranchFileID)

	nodeB, _, _ := mgr.AddNode(bf, "dev-B", "push-2", "v2\n")
	if nodeB.ParentNodeID != nodeA.NodeID {
		t.Errorf("second node parent = %q, want %q", nodeB.ParentNodeID, nodeA.NodeID)
	}
}

