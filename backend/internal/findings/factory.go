package findings

import (
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"strings"
	"unicode"

	"github.com/JustCallMeMin/repoCompass/backend/internal/rules"
)

// FindingSpec contains normalized input for building a validated finding.
type FindingSpec struct {
	AnalyzerID      string
	RuleID          string
	Target          string
	Severity        rules.Severity
	Category        rules.Category
	Title           string
	Message         string
	Evidence        []Evidence
	Recommendations []Recommendation
}

// BuildFinding creates and validates a finding from a spec.
func BuildFinding(spec FindingSpec) (Finding, error) {
	id := BuildFindingID(spec.AnalyzerID, spec.RuleID, spec.Target)
	finding := NewFinding(
		id,
		spec.RuleID,
		spec.AnalyzerID,
		spec.Severity,
		spec.Category,
		spec.Title,
		spec.Message,
		spec.Evidence,
	)
	finding.Recommendations = spec.Recommendations
	if err := finding.Validate(); err != nil {
		return Finding{}, err
	}
	return finding, nil
}

// BuildFindingID returns a stable finding ID from analyzer, rule, and target.
func BuildFindingID(analyzerID string, ruleID string, target string) string {
	sum := sha256.Sum256([]byte(target))
	return fmt.Sprintf("finding_%s_%s_%s", slug(analyzerID), slug(ruleID), hex.EncodeToString(sum[:])[:12])
}

// FileMissingEvidence creates file_missing evidence for a repository-relative path.
func FileMissingEvidence(path string, message string) Evidence {
	return NewEvidence(EvidenceTypeFileMissing, message, path, "")
}

// FileExistsEvidence creates file_exists evidence for a repository-relative path.
func FileExistsEvidence(path string, message string) Evidence {
	return NewEvidence(EvidenceTypeFileExists, message, path, "")
}

// PatternMissingEvidence creates pattern_missing evidence for a repository-relative path.
func PatternMissingEvidence(path string, value string, message string) Evidence {
	return NewEvidence(EvidenceTypePatternMissing, message, path, value)
}

// RecommendationForFinding creates a recommendation linked to a finding.
func RecommendationForFinding(
	findingID string,
	title string,
	action string,
	rationale string,
	priority RecommendationPriority,
) Recommendation {
	return NewRecommendation(findingID, title, action, rationale, priority)
}

func slug(value string) string {
	var b strings.Builder
	lastUnderscore := false
	for _, r := range value {
		if unicode.IsLetter(r) || unicode.IsDigit(r) {
			b.WriteRune(unicode.ToLower(r))
			lastUnderscore = false
			continue
		}
		if !lastUnderscore {
			b.WriteRune('_')
			lastUnderscore = true
		}
	}
	return strings.Trim(b.String(), "_")
}
