package analysisclient

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/nccuhacks/nccu26/mcp/internal/models"
)

func TestAnalyzeOverlaps_Success(t *testing.T) {
	expected := models.AnalyzeOverlapsResponse{
		Overlaps: []models.Overlap{
			{
				FilePath:   "f.py",
				SymbolName: "fn",
				Severity:   "critical",
				AgentA:     "a",
				AgentB:     "b",
			},
		},
		FileRisks: []models.FileRisk{
			{
				FilePath:          "f.py",
				RiskScore:         40,
				StabilityScore:    60,
				OverlapCount:      1,
				Contributors:      []string{"a", "b"},
				ContributorsCount: 2,
				IsHotspot:         false,
			},
		},
	}

	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/analyze/overlaps" {
			t.Fatalf("unexpected path: %s", r.URL.Path)
		}
		if r.Method != http.MethodPost {
			t.Fatalf("unexpected method: %s", r.Method)
		}
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(expected)
	}))
	defer srv.Close()

	c := New(srv.URL, 5*time.Second)
	resp, err := c.AnalyzeOverlaps(context.Background(), []models.AnalysisChangeSet{
		{AgentID: "a", Files: []models.AnalysisFileInput{{Path: "f.py"}}},
		{AgentID: "b", Files: []models.AnalysisFileInput{{Path: "f.py"}}},
	})
	if err != nil {
		t.Fatal(err)
	}
	if len(resp.Overlaps) != 1 {
		t.Fatalf("expected 1 overlap, got %d", len(resp.Overlaps))
	}
	if resp.Overlaps[0].Severity != "critical" {
		t.Fatalf("expected critical severity, got %s", resp.Overlaps[0].Severity)
	}
}

func TestAnalyzeOverlaps_BackendError(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte("internal error"))
	}))
	defer srv.Close()

	c := New(srv.URL, 5*time.Second)
	_, err := c.AnalyzeOverlaps(context.Background(), nil)
	if err == nil {
		t.Fatal("expected error on 500 response")
	}
}

func TestAnalyzeOverlaps_Timeout(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		time.Sleep(200 * time.Millisecond)
		w.WriteHeader(http.StatusOK)
	}))
	defer srv.Close()

	c := New(srv.URL, 50*time.Millisecond)
	_, err := c.AnalyzeOverlaps(context.Background(), nil)
	if err == nil {
		t.Fatal("expected timeout error")
	}
}

func TestAnalyzeOverlaps_InvalidJSON(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.Write([]byte("not json"))
	}))
	defer srv.Close()

	c := New(srv.URL, 5*time.Second)
	_, err := c.AnalyzeOverlaps(context.Background(), nil)
	if err == nil {
		t.Fatal("expected decode error")
	}
}

func TestHealth_Success(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/health" {
			t.Fatalf("unexpected path: %s", r.URL.Path)
		}
		w.WriteHeader(http.StatusOK)
	}))
	defer srv.Close()

	c := New(srv.URL, 5*time.Second)
	if err := c.Health(context.Background()); err != nil {
		t.Fatal(err)
	}
}

func TestHealth_BackendDown(t *testing.T) {
	c := New("http://127.0.0.1:1", 100*time.Millisecond)
	err := c.Health(context.Background())
	if err == nil {
		t.Fatal("expected error when backend is unreachable")
	}
}
