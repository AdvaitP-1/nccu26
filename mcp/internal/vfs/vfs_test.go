package vfs

import (
	"sort"
	"sync"
	"testing"

	"github.com/nccuhacks/nccu26/mcp/internal/models"
)

func snapshot(path, lang, content string) models.FileSnapshot {
	return models.FileSnapshot{Path: path, Language: lang, Content: content}
}

func TestPropose_StoresAndReturnsState(t *testing.T) {
	m := NewManager()
	m.Propose("agent-a", "s1", "t1", []models.FileSnapshot{snapshot("a.py", "python", "x")})
	m.Propose("agent-b", "s2", "t2", []models.FileSnapshot{snapshot("b.py", "python", "y")})

	state := m.State()
	if state.TotalAgents != 2 {
		t.Fatalf("expected 2 agents, got %d", state.TotalAgents)
	}
	if state.TotalFiles != 2 {
		t.Fatalf("expected 2 files, got %d", state.TotalFiles)
	}
}

func TestPropose_ReplacesExisting(t *testing.T) {
	m := NewManager()
	m.Propose("a", "", "", []models.FileSnapshot{snapshot("1.py", "python", "v1")})
	m.Propose("a", "", "", []models.FileSnapshot{snapshot("2.py", "python", "v2"), snapshot("3.py", "python", "v3")})

	files, err := m.FilesForAgent("a")
	if err != nil {
		t.Fatal(err)
	}
	if len(files) != 2 {
		t.Fatalf("expected 2 files after replace, got %d", len(files))
	}
}

func TestAddFile_Appends(t *testing.T) {
	m := NewManager()
	m.AddFile("a", "s", "t", snapshot("1.py", "python", "v1"))
	m.AddFile("a", "s", "t", snapshot("2.py", "python", "v2"))

	files, _ := m.FilesForAgent("a")
	if len(files) != 2 {
		t.Fatalf("expected 2 files, got %d", len(files))
	}
}

func TestFilesForAgent_ErrorOnMissing(t *testing.T) {
	m := NewManager()
	_, err := m.FilesForAgent("nonexistent")
	if err == nil {
		t.Fatal("expected error for missing agent")
	}
}

func TestClear_RemovesSingleAgent(t *testing.T) {
	m := NewManager()
	m.Propose("a", "", "", []models.FileSnapshot{snapshot("f.py", "python", "x")})
	m.Propose("b", "", "", []models.FileSnapshot{snapshot("g.py", "python", "y")})
	m.Clear("a")

	if m.AgentCount() != 1 {
		t.Fatalf("expected 1 agent after clear, got %d", m.AgentCount())
	}
	_, err := m.FilesForAgent("a")
	if err == nil {
		t.Fatal("expected error for cleared agent")
	}
}

func TestClearAll(t *testing.T) {
	m := NewManager()
	m.Propose("a", "", "", nil)
	m.Propose("b", "", "", nil)
	m.ClearAll()

	if m.AgentCount() != 0 {
		t.Fatalf("expected 0 agents after ClearAll, got %d", m.AgentCount())
	}
}

func TestChangeSetsForAnalysis_MapsCorrectly(t *testing.T) {
	m := NewManager()
	m.Propose("a", "", "", []models.FileSnapshot{snapshot("f.py", "python", "content-a")})
	m.Propose("b", "", "", []models.FileSnapshot{snapshot("f.py", "python", "content-b")})

	cs := m.ChangeSetsForAnalysis()
	if len(cs) != 2 {
		t.Fatalf("expected 2 changesets, got %d", len(cs))
	}
	ids := []string{cs[0].AgentID, cs[1].AgentID}
	sort.Strings(ids)
	if ids[0] != "a" || ids[1] != "b" {
		t.Fatalf("unexpected agent IDs: %v", ids)
	}
}

func TestManyAgents(t *testing.T) {
	m := NewManager()
	for i := range 20 {
		id := string(rune('a' + i))
		m.Propose(id, "", "", []models.FileSnapshot{snapshot("f.py", "python", "x")})
	}
	if m.AgentCount() != 20 {
		t.Fatalf("expected 20 agents, got %d", m.AgentCount())
	}
	cs := m.ChangeSetsForAnalysis()
	if len(cs) != 20 {
		t.Fatalf("expected 20 changesets, got %d", len(cs))
	}
}

func TestConcurrentAccess(t *testing.T) {
	m := NewManager()
	var wg sync.WaitGroup
	for i := range 50 {
		wg.Add(1)
		go func(idx int) {
			defer wg.Done()
			id := string(rune('A' + idx%26))
			m.Propose(id, "", "", []models.FileSnapshot{snapshot("f.py", "python", "x")})
			_ = m.State()
			_ = m.ChangeSetsForAnalysis()
		}(i)
	}
	wg.Wait()
	// Should not panic; state should be consistent.
	state := m.State()
	if state.TotalAgents < 1 {
		t.Fatal("expected at least 1 agent after concurrent writes")
	}
}

func TestEmptyVFSState(t *testing.T) {
	m := NewManager()
	state := m.State()
	if state.TotalAgents != 0 || state.TotalFiles != 0 || len(state.PendingChanges) != 0 {
		t.Fatal("empty VFS should return zero values")
	}
}

func TestUpdatedAtAdvances(t *testing.T) {
	m := NewManager()
	m.AddFile("a", "", "", snapshot("1.py", "python", "v1"))
	first := m.pending["a"].UpdatedAt
	m.AddFile("a", "", "", snapshot("2.py", "python", "v2"))
	second := m.pending["a"].UpdatedAt
	if !second.After(first) && !second.Equal(first) {
		t.Fatal("UpdatedAt should advance on AddFile")
	}
}
