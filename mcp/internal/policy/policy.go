// Package policy evaluates whether a micro-commit should be allowed.
//
// Rules (v1):
//   - Block if any overlap severity is "critical" (when BlockOnCritical is set).
//   - Block if any file's risk score exceeds the configured threshold.
//   - Otherwise allow.
//
// All decision logic lives here so it can be tested in isolation.
package policy

import (
	"fmt"
	"strings"

	"github.com/nccuhacks/nccu26/mcp/internal/models"
)

// Decision is the structured outcome of a policy evaluation.
type Decision struct {
	Allowed             bool           `json:"allowed"`
	Reasons             []string       `json:"reasons"`
	BlockingFiles       []string       `json:"blocking_files,omitempty"`
	MaxSeverity         string         `json:"max_severity,omitempty"`
	EvaluatedThresholds map[string]int `json:"evaluated_thresholds"`
}

// Evaluator holds the thresholds used by Evaluate.
type Evaluator struct {
	RiskThreshold   int
	BlockOnCritical bool
}

// NewEvaluator builds an Evaluator from explicit values.
func NewEvaluator(riskThreshold int, blockOnCritical bool) *Evaluator {
	return &Evaluator{
		RiskThreshold:   riskThreshold,
		BlockOnCritical: blockOnCritical,
	}
}

// Evaluate inspects the analysis response and returns a structured commit
// decision.  It collects *all* blocking reasons rather than short-circuiting
// on the first, so callers get complete diagnostic information.
func (e *Evaluator) Evaluate(resp *models.AnalyzeOverlapsResponse) Decision {
	d := Decision{
		Allowed: true,
		EvaluatedThresholds: map[string]int{
			"risk_threshold": e.RiskThreshold,
		},
	}

	if resp == nil {
		d.Reasons = []string{"no analysis data — allowing by default"}
		return d
	}

	// Track global max severity across all overlaps.
	maxRank := 0
	severityRank := map[string]int{"critical": 4, "high": 3, "medium": 2, "low": 1}

	// Rule 1: block on critical overlaps.
	blockingFilesSet := map[string]bool{}
	if e.BlockOnCritical {
		for _, o := range resp.Overlaps {
			sev := strings.ToLower(o.Severity)
			if rank, ok := severityRank[sev]; ok && rank > maxRank {
				maxRank = rank
			}
			if sev == "critical" {
				d.Allowed = false
				reason := fmt.Sprintf(
					"critical overlap on %s (symbol %q between %s and %s)",
					o.FilePath, o.SymbolName, o.AgentA, o.AgentB,
				)
				d.Reasons = append(d.Reasons, reason)
				blockingFilesSet[o.FilePath] = true
			}
		}
	} else {
		for _, o := range resp.Overlaps {
			sev := strings.ToLower(o.Severity)
			if rank, ok := severityRank[sev]; ok && rank > maxRank {
				maxRank = rank
			}
		}
	}

	// Rule 2: block on high file-level risk.
	for _, fr := range resp.FileRisks {
		if fr.RiskScore > e.RiskThreshold {
			d.Allowed = false
			reason := fmt.Sprintf(
				"file %s risk score %d exceeds threshold %d",
				fr.FilePath, fr.RiskScore, e.RiskThreshold,
			)
			d.Reasons = append(d.Reasons, reason)
			blockingFilesSet[fr.FilePath] = true
		}
	}

	// Populate structured output.
	for f := range blockingFilesSet {
		d.BlockingFiles = append(d.BlockingFiles, f)
	}

	for sev, rank := range severityRank {
		if rank == maxRank && maxRank > 0 {
			d.MaxSeverity = sev
			break
		}
	}

	if d.Allowed {
		d.Reasons = []string{"all checks passed"}
	}

	return d
}
