package diff

import (
	"strings"
	"testing"
)

func TestCreateAndApplyPatch_SimpleChange(t *testing.T) {
	e := NewEngine()

	old := "line1\nline2\nline3\n"
	new := "line1\nline2_modified\nline3\n"

	patch := e.CreatePatch(old, new)
	if patch == "" {
		t.Fatal("expected non-empty patch")
	}

	result, err := e.ApplyPatch(old, patch)
	if err != nil {
		t.Fatalf("apply patch: %v", err)
	}
	if result != new {
		t.Errorf("got %q, want %q", result, new)
	}
}

func TestCreateAndApplyPatch_Insertion(t *testing.T) {
	e := NewEngine()

	old := "def hello():\n    pass\n"
	new := "def hello():\n    print('hello')\n    pass\n"

	patch := e.CreatePatch(old, new)
	result, err := e.ApplyPatch(old, patch)
	if err != nil {
		t.Fatalf("apply patch: %v", err)
	}
	if result != new {
		t.Errorf("got %q, want %q", result, new)
	}
}

func TestCreateAndApplyPatch_Deletion(t *testing.T) {
	e := NewEngine()

	old := "a\nb\nc\nd\n"
	new := "a\nd\n"

	patch := e.CreatePatch(old, new)
	result, err := e.ApplyPatch(old, patch)
	if err != nil {
		t.Fatalf("apply patch: %v", err)
	}
	if result != new {
		t.Errorf("got %q, want %q", result, new)
	}
}

func TestCreateAndApplyPatch_EmptyToContent(t *testing.T) {
	e := NewEngine()

	old := ""
	new := "brand new file\n"

	patch := e.CreatePatch(old, new)
	result, err := e.ApplyPatch(old, patch)
	if err != nil {
		t.Fatalf("apply patch: %v", err)
	}
	if result != new {
		t.Errorf("got %q, want %q", result, new)
	}
}

func TestCreateAndApplyPatch_NoChange(t *testing.T) {
	e := NewEngine()

	content := "unchanged\n"
	patch := e.CreatePatch(content, content)

	result, err := e.ApplyPatch(content, patch)
	if err != nil {
		t.Fatalf("apply patch: %v", err)
	}
	if result != content {
		t.Errorf("got %q, want %q", result, content)
	}
}

func TestHasChanges(t *testing.T) {
	e := NewEngine()

	if e.HasChanges("same", "same") {
		t.Error("expected no changes")
	}
	if !e.HasChanges("old", "new") {
		t.Error("expected changes")
	}
}

func TestContentHash_Deterministic(t *testing.T) {
	h1 := ContentHash("hello")
	h2 := ContentHash("hello")
	if h1 != h2 {
		t.Errorf("hashes differ: %q vs %q", h1, h2)
	}
	h3 := ContentHash("world")
	if h1 == h3 {
		t.Error("different content should have different hashes")
	}
}

func TestApplyPatch_InvalidPatchText(t *testing.T) {
	e := NewEngine()

	_, err := e.ApplyPatch("base", "not a valid patch %%% @@@")
	if err == nil {
		t.Error("expected error for invalid patch text")
	}
}

func TestCreateAndApplyPatch_LargeFile(t *testing.T) {
	e := NewEngine()

	var lines []string
	for i := 0; i < 500; i++ {
		lines = append(lines, "line content here")
	}
	old := strings.Join(lines, "\n") + "\n"

	lines[250] = "MODIFIED LINE 250"
	lines = append(lines[:400], append([]string{"INSERTED LINE"}, lines[400:]...)...)
	new := strings.Join(lines, "\n") + "\n"

	patch := e.CreatePatch(old, new)
	result, err := e.ApplyPatch(old, patch)
	if err != nil {
		t.Fatalf("apply patch: %v", err)
	}
	if result != new {
		t.Error("large file patch did not produce expected result")
	}
}
