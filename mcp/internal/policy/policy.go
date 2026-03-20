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

// Decision is the outcome of a policy evaluation.
type Decision struct {
	Allowed bool
	Reason  string
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

// Evaluate inspects the analysis response and returns a commit decision.
func (e *Evaluator) Evaluate(resp *models.AnalyzeOverlapsResponse) Decision {
	if resp == nil {
		return Decision{Allowed: true, Reason: "no analysis data — allowing by default"}
	}

	// Rule 1: block on critical overlaps.
	if e.BlockOnCritical {
		for _, o := range resp.Overlaps {
			if strings.EqualFold(o.Severity, "critical") {
				return Decision{
					Allowed: false,
					Reason: fmt.Sprintf(
						"blocked: critical overlap on %s (symbol %q between %s and %s)",
						o.FilePath, o.SymbolName, o.AgentA, o.AgentB,
					),
				}
			}
		}
	}

	// Rule 2: block on high file-level risk.
	for _, fr := range resp.FileRisks {
		if fr.RiskScore > e.RiskThreshold {
			return Decision{
				Allowed: false,
				Reason: fmt.Sprintf(
					"blocked: file %s risk score %d exceeds threshold %d",
					fr.FilePath, fr.RiskScore, e.RiskThreshold,
				),
			}
		}
	}

	return Decision{Allowed: true, Reason: "all checks passed"}
}
