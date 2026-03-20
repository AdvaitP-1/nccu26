package service

import (
	"context"
	"testing"

	"github.com/nccuhacks/nccu26/mcp/internal/diff"
	"github.com/nccuhacks/nccu26/mcp/internal/filetree"
	"github.com/nccuhacks/nccu26/mcp/internal/models"
	"github.com/nccuhacks/nccu26/mcp/internal/storage"
	"github.com/nccuhacks/nccu26/mcp/internal/vfs"
)

func newTestService() *GitService {
	store := storage.New()
	engine := diff.NewEngine()
	tree := filetree.NewManager(store, engine)
	return NewGitService(store, engine, tree, nil, nil)
}

func newTestServiceWithVFS(vfsMgr *vfs.Manager) *GitService {
	store := storage.New()
	engine := diff.NewEngine()
	tree := filetree.NewManager(store, engine)
	return NewGitService(store, engine, tree, nil, vfsMgr)
}

func TestIngestPush_SingleFile(t *testing.T) {
	svc := newTestService()
	ctx := context.Background()

	resp, err := svc.IngestPush(ctx, models.IngestPushRequest{
		BranchName: "feature/auth",
		UserID:     "dev-A",
		Files: []models.PushFileChange{
			{
				FilePath:    "auth.py",
				BaseContent: "def login():\n    pass\n",
				NewContent:  "def login():\n    return True\n",
			},
		},
	})
	if err != nil {
		t.Fatalf("IngestPush: %v", err)
	}

	if resp.PushID == "" {
		t.Error("expected non-empty push ID")
	}
	if len(resp.NodeIDs) != 1 {
		t.Errorf("got %d nodes, want 1", len(resp.NodeIDs))
	}
}

func TestIngestPush_MultipleFiles(t *testing.T) {
	svc := newTestService()
	ctx := context.Background()

	resp, err := svc.IngestPush(ctx, models.IngestPushRequest{
		BranchName: "feature/auth",
		UserID:     "dev-A",
		Files: []models.PushFileChange{
			{FilePath: "auth.py", BaseContent: "old auth\n", NewContent: "new auth\n"},
			{FilePath: "user.py", BaseContent: "old user\n", NewContent: "new user\n"},
		},
	})
	if err != nil {
		t.Fatalf("IngestPush: %v", err)
	}
	if len(resp.NodeIDs) != 2 {
		t.Errorf("got %d nodes, want 2", len(resp.NodeIDs))
	}
}

func TestIngestPush_Validation(t *testing.T) {
	svc := newTestService()
	ctx := context.Background()

	_, err := svc.IngestPush(ctx, models.IngestPushRequest{})
	if err == nil {
		t.Error("expected error for empty request")
	}

	_, err = svc.IngestPush(ctx, models.IngestPushRequest{
		BranchName: "main",
		UserID:     "dev",
	})
	if err == nil {
		t.Error("expected error for empty files")
	}
}

func TestGetFileState(t *testing.T) {
	svc := newTestService()
	ctx := context.Background()

	svc.IngestPush(ctx, models.IngestPushRequest{
		BranchName: "feature/ui",
		UserID:     "dev-A",
		Files: []models.PushFileChange{
			{FilePath: "app.tsx", BaseContent: "old\n", NewContent: "new\n"},
		},
	})

	state, err := svc.GetFileState("feature/ui", "app.tsx")
	if err != nil {
		t.Fatalf("GetFileState: %v", err)
	}
	if state.BranchFile.FilePath != "app.tsx" {
		t.Errorf("file path = %q", state.BranchFile.FilePath)
	}
	if len(state.ActiveNodes) != 1 {
		t.Errorf("active nodes = %d, want 1", len(state.ActiveNodes))
	}
}

func TestGetFileState_NotFound(t *testing.T) {
	svc := newTestService()

	_, err := svc.GetFileState("nonexistent", "file.py")
	if err == nil {
		t.Error("expected error for missing branch")
	}
}

func TestPrepareMergeContext(t *testing.T) {
	svc := newTestService()
	ctx := context.Background()

	svc.IngestPush(ctx, models.IngestPushRequest{
		BranchName: "feature/auth",
		UserID:     "dev-A",
		Files: []models.PushFileChange{
			{FilePath: "auth.py", BaseContent: "base\n", NewContent: "changed-A\n"},
		},
	})
	svc.IngestPush(ctx, models.IngestPushRequest{
		BranchName: "feature/auth",
		UserID:     "dev-B",
		Files: []models.PushFileChange{
			{FilePath: "auth.py", BaseContent: "base\n", NewContent: "changed-B\n"},
		},
	})

	resp, err := svc.PrepareMergeContext("feature/auth", "auth.py")
	if err != nil {
		t.Fatalf("PrepareMergeContext: %v", err)
	}

	if resp.BaseContent != "base\n" {
		t.Errorf("base content = %q, want %q", resp.BaseContent, "base\n")
	}
	if len(resp.ActiveNodes) != 2 {
		t.Errorf("active nodes = %d, want 2", len(resp.ActiveNodes))
	}
	if len(resp.Candidates) != 2 {
		t.Errorf("candidates = %d, want 2", len(resp.Candidates))
	}
	if len(resp.DiffPayloads) != 2 {
		t.Errorf("diff payloads = %d, want 2", len(resp.DiffPayloads))
	}

	// Each candidate should have reconstructed content.
	for _, c := range resp.Candidates {
		if c.ReconstructedContent == "" {
			t.Errorf("candidate %q has empty content", c.NodeID)
		}
		if c.ReconstructedContent == "base\n" {
			t.Errorf("candidate %q still has base content", c.NodeID)
		}
	}
}

func TestApplyMergeResult(t *testing.T) {
	svc := newTestService()
	ctx := context.Background()

	svc.IngestPush(ctx, models.IngestPushRequest{
		BranchName: "feature/auth",
		UserID:     "dev-A",
		Files: []models.PushFileChange{
			{FilePath: "auth.py", BaseContent: "base\n", NewContent: "changed\n"},
		},
	})

	resp, err := svc.ApplyMergeResult(ctx, models.ApplyMergeResultRequest{
		BranchName:    "feature/auth",
		FilePath:      "auth.py",
		MergedContent: "externally decided content\n",
	})
	if err != nil {
		t.Fatalf("ApplyMergeResult: %v", err)
	}

	if resp.MergedNodeID == "" {
		t.Error("expected merged node ID")
	}
	if resp.NodesAffected != 1 {
		t.Errorf("nodes affected = %d, want 1", resp.NodesAffected)
	}

	// After apply, file state should show no active nodes.
	state, _ := svc.GetFileState("feature/auth", "auth.py")
	if len(state.ActiveNodes) != 0 {
		t.Errorf("active nodes after merge = %d, want 0", len(state.ActiveNodes))
	}
	if state.BranchFile.BaseContent != "externally decided content\n" {
		t.Errorf("base content not updated after merge")
	}
}

func TestApplyMergeResult_SelectiveSupersede(t *testing.T) {
	svc := newTestService()
	ctx := context.Background()

	push1, _ := svc.IngestPush(ctx, models.IngestPushRequest{
		BranchName: "feature/auth",
		UserID:     "dev-A",
		Files: []models.PushFileChange{
			{FilePath: "auth.py", BaseContent: "base\n", NewContent: "A\n"},
		},
	})
	svc.IngestPush(ctx, models.IngestPushRequest{
		BranchName: "feature/auth",
		UserID:     "dev-B",
		Files: []models.PushFileChange{
			{FilePath: "auth.py", BaseContent: "base\n", NewContent: "B\n"},
		},
	})

	// Only supersede dev-A's node.
	resp, err := svc.ApplyMergeResult(ctx, models.ApplyMergeResultRequest{
		BranchName:        "feature/auth",
		FilePath:          "auth.py",
		MergedContent:     "merged-A\n",
		SupersededNodeIDs: push1.NodeIDs,
	})
	if err != nil {
		t.Fatalf("ApplyMergeResult: %v", err)
	}
	if resp.NodesAffected != 1 {
		t.Errorf("nodes affected = %d, want 1", resp.NodesAffected)
	}
}

func TestPrepareCommit(t *testing.T) {
	svc := newTestService()
	ctx := context.Background()

	pushResp, _ := svc.IngestPush(ctx, models.IngestPushRequest{
		BranchName: "feature/auth",
		UserID:     "dev-A",
		Files: []models.PushFileChange{
			{FilePath: "auth.py", BaseContent: "base\n", NewContent: "changed\n"},
			{FilePath: "user.py", BaseContent: "base\n", NewContent: "changed\n"},
		},
	})

	// Before any merge, files should be pending.
	prep, err := svc.PrepareCommit(pushResp.PushID)
	if err != nil {
		t.Fatalf("PrepareCommit: %v", err)
	}
	if prep.AllResolved {
		t.Error("expected AllResolved=false before merge")
	}
	if len(prep.PendingFiles) != 2 {
		t.Errorf("pending files = %d, want 2", len(prep.PendingFiles))
	}
	if len(prep.Files) != 2 {
		t.Errorf("files = %d, want 2", len(prep.Files))
	}

	// Apply merge results for both files.
	svc.ApplyMergeResult(ctx, models.ApplyMergeResultRequest{
		BranchName: "feature/auth", FilePath: "auth.py", MergedContent: "final-auth\n",
	})
	svc.ApplyMergeResult(ctx, models.ApplyMergeResultRequest{
		BranchName: "feature/auth", FilePath: "user.py", MergedContent: "final-user\n",
	})

	// Now all files should be resolved.
	prep2, err := svc.PrepareCommit(pushResp.PushID)
	if err != nil {
		t.Fatalf("PrepareCommit after merge: %v", err)
	}
	if !prep2.AllResolved {
		t.Error("expected AllResolved=true after all merges")
	}
	if len(prep2.PendingFiles) != 0 {
		t.Errorf("pending files after merge = %d, want 0", len(prep2.PendingFiles))
	}
}

func TestCreateGroupedCommit_NoGitExecutor(t *testing.T) {
	svc := newTestService()
	ctx := context.Background()

	pushResp, _ := svc.IngestPush(ctx, models.IngestPushRequest{
		BranchName: "feature/auth",
		UserID:     "dev-A",
		Files: []models.PushFileChange{
			{FilePath: "auth.py", BaseContent: "base\n", NewContent: "changed\n"},
			{FilePath: "user.py", BaseContent: "base\n", NewContent: "changed\n"},
		},
	})

	commitResp, err := svc.CreateGroupedCommit(ctx, models.GroupedCommitRequest{
		PushID:  pushResp.PushID,
		Message: "test grouped commit",
	})
	if err != nil {
		t.Fatalf("CreateGroupedCommit: %v", err)
	}

	if commitResp.CommitID == "" {
		t.Error("expected commit ID")
	}
	if commitResp.FilesIncluded != 2 {
		t.Errorf("files included = %d, want 2", commitResp.FilesIncluded)
	}
	if commitResp.Status != string(models.CommitStatusCreated) {
		t.Errorf("status = %q, want %q", commitResp.Status, models.CommitStatusCreated)
	}

	record, err := svc.GetCommitRecord(commitResp.CommitID)
	if err != nil {
		t.Fatalf("GetCommitRecord: %v", err)
	}
	if record.PushID != pushResp.PushID {
		t.Errorf("commit push_id = %q, want %q", record.PushID, pushResp.PushID)
	}
	if len(record.FileNodeIDs) != 2 {
		t.Errorf("file node IDs = %d, want 2", len(record.FileNodeIDs))
	}
}

func TestHealth(t *testing.T) {
	svc := newTestService()
	ctx := context.Background()

	status := svc.Health(ctx)
	if status.Status != "no_repo" {
		t.Errorf("status = %q, want no_repo (no git executor)", status.Status)
	}
	if status.Stats == nil {
		t.Error("expected stats map")
	}
}

func TestGitFlow_MirrorsAndClearsVFS(t *testing.T) {
	vfsMgr := vfs.NewManager()
	svc := newTestServiceWithVFS(vfsMgr)
	ctx := context.Background()

	pushResp, err := svc.IngestPush(ctx, models.IngestPushRequest{
		BranchName: "feature/auth",
		UserID:     "dev-A",
		Files: []models.PushFileChange{
			{FilePath: "auth.py", BaseContent: "base\n", NewContent: "changed-auth\n"},
			{FilePath: "user.py", BaseContent: "base\n", NewContent: "changed-user\n"},
		},
	})
	if err != nil {
		t.Fatalf("IngestPush: %v", err)
	}

	vfsState := vfsMgr.State()
	if vfsState.TotalAgents != 1 {
		t.Fatalf("VFS total agents = %d, want 1", vfsState.TotalAgents)
	}
	if vfsState.TotalFiles != 2 {
		t.Fatalf("VFS total files = %d, want 2", vfsState.TotalFiles)
	}

	svc.ApplyMergeResult(ctx, models.ApplyMergeResultRequest{
		BranchName: "feature/auth", FilePath: "auth.py", MergedContent: "final-auth\n",
	})
	svc.ApplyMergeResult(ctx, models.ApplyMergeResultRequest{
		BranchName: "feature/auth", FilePath: "user.py", MergedContent: "final-user\n",
	})

	_, err = svc.CreateGroupedCommit(ctx, models.GroupedCommitRequest{
		PushID:  pushResp.PushID,
		Message: "test grouped commit",
	})
	if err != nil {
		t.Fatalf("CreateGroupedCommit: %v", err)
	}

	vfsState = vfsMgr.State()
	if vfsState.TotalAgents != 0 {
		t.Fatalf("VFS total agents after commit = %d, want 0", vfsState.TotalAgents)
	}
	if vfsState.TotalFiles != 0 {
		t.Fatalf("VFS total files after commit = %d, want 0", vfsState.TotalFiles)
	}
}
