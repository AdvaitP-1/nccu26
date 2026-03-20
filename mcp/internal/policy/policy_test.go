package policy

import (
	"testing"

	"github.com/nccuhacks/nccu26/mcp/internal/models"
)

func overlap(sev string, file string, a, b string) models.Overlap {
	return models.Overlap{
		FilePath:   file,
		SymbolName: "fn",
		SymbolKind: "function",
		AgentA:     a,
		AgentB:     b,
		Severity:   sev,
		Reason:     "test",
	}
}

func fileRisk(file string, score int) models.FileRisk {
	return models.FileRisk{
		FilePath:  file,
		RiskScore: score,
	}
}

func TestBlockOnCriticalOverlap(t *testing.T) {
	e := NewEvaluator(70, true)
	resp := &models.AnalyzeOverlapsResponse{
		Overlaps: []models.Overlap{overlap("critical", "f.py", "a", "b")},
	}
	d := e.Evaluate(resp)
	if d.Allowed {
		t.Fatal("expected blocked on critical overlap")
	}
	if len(d.Reasons) == 0 {
		t.Fatal("expected at least one reason")
	}
	if len(d.BlockingFiles) == 0 {
		t.Fatal("expected blocking files")
	}
}

func TestBlockOnFileRiskAboveThreshold(t *testing.T) {
	e := NewEvaluator(70, true)
	resp := &models.AnalyzeOverlapsResponse{
		FileRisks: []models.FileRisk{fileRisk("hot.py", 85)},
	}
	d := e.Evaluate(resp)
	if d.Allowed {
		t.Fatal("expected blocked on high risk")
	}
	if d.EvaluatedThresholds["risk_threshold"] != 70 {
		t.Fatal("expected threshold in output")
	}
}

func TestAllowSafeChangeset(t *testing.T) {
	e := NewEvaluator(70, true)
	resp := &models.AnalyzeOverlapsResponse{
		Overlaps:  []models.Overlap{overlap("medium", "f.py", "a", "b")},
		FileRisks: []models.FileRisk{fileRisk("f.py", 15)},
	}
	d := e.Evaluate(resp)
	if !d.Allowed {
		t.Fatalf("expected allowed, got blocked: %v", d.Reasons)
	}
}

func TestNilResponse_Allowed(t *testing.T) {
	e := NewEvaluator(70, true)
	d := e.Evaluate(nil)
	if !d.Allowed {
		t.Fatal("nil response should allow")
	}
}

func TestCriticalDisabled_StillBlocksOnRisk(t *testing.T) {
	e := NewEvaluator(50, false) // BlockOnCritical = false
	resp := &models.AnalyzeOverlapsResponse{
		Overlaps:  []models.Overlap{overlap("critical", "f.py", "a", "b")},
		FileRisks: []models.FileRisk{fileRisk("f.py", 60)},
	}
	d := e.Evaluate(resp)
	if d.Allowed {
		t.Fatal("expected blocked on risk even with critical disabled")
	}
}

func TestCriticalDisabled_LowRisk_Allowed(t *testing.T) {
	e := NewEvaluator(70, false)
	resp := &models.AnalyzeOverlapsResponse{
		Overlaps:  []models.Overlap{overlap("critical", "f.py", "a", "b")},
		FileRisks: []models.FileRisk{fileRisk("f.py", 30)},
	}
	d := e.Evaluate(resp)
	if !d.Allowed {
		t.Fatal("expected allowed when critical disabled and risk is low")
	}
}

func TestMultipleBlockingReasons(t *testing.T) {
	e := NewEvaluator(30, true)
	resp := &models.AnalyzeOverlapsResponse{
		Overlaps: []models.Overlap{
			overlap("critical", "a.py", "x", "y"),
			overlap("critical", "b.py", "x", "z"),
		},
		FileRisks: []models.FileRisk{
			fileRisk("a.py", 40),
			fileRisk("b.py", 50),
		},
	}
	d := e.Evaluate(resp)
	if d.Allowed {
		t.Fatal("expected blocked")
	}
	// Two critical reasons + two risk reasons = at least 4
	if len(d.Reasons) < 4 {
		t.Fatalf("expected >= 4 reasons, got %d: %v", len(d.Reasons), d.Reasons)
	}
	if len(d.BlockingFiles) < 2 {
		t.Fatalf("expected >= 2 blocking files, got %d", len(d.BlockingFiles))
	}
}

func TestMaxSeverityTracked(t *testing.T) {
	e := NewEvaluator(100, true)
	resp := &models.AnalyzeOverlapsResponse{
		Overlaps: []models.Overlap{
			overlap("low", "f.py", "a", "b"),
			overlap("high", "f.py", "a", "c"),
			overlap("medium", "f.py", "b", "c"),
		},
	}
	d := e.Evaluate(resp)
	if d.MaxSeverity != "high" {
		t.Fatalf("expected max severity 'high', got %q", d.MaxSeverity)
	}
}

func TestEmptyOverlaps_Allowed(t *testing.T) {
	e := NewEvaluator(70, true)
	resp := &models.AnalyzeOverlapsResponse{}
	d := e.Evaluate(resp)
	if !d.Allowed {
		t.Fatal("empty response should allow")
	}
}
