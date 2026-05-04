// Package findings contains the core finding domain model.
package findings

import (
	"fmt"
	"path/filepath"

	"github.com/JustCallMeMin/repoCompass/backend/internal/rules"
)

// FindingStatus describes the lifecycle state of a finding.
type FindingStatus string

const (
	FindingStatusOpen FindingStatus = "open"
)

// EvidenceType describes the kind of fact supporting a finding.
type EvidenceType string

const (
	EvidenceTypeFileExists     EvidenceType = "file_exists"
	EvidenceTypeFileMissing    EvidenceType = "file_missing"
	EvidenceTypePatternMatch   EvidenceType = "pattern_match"
	EvidenceTypePatternMissing EvidenceType = "pattern_missing"
	EvidenceTypeMetadata       EvidenceType = "metadata"
)

// RecommendationPriority describes action priority for a recommendation.
type RecommendationPriority string

const (
	RecommendationPriorityHigh   RecommendationPriority = "high"
	RecommendationPriorityMedium RecommendationPriority = "medium"
	RecommendationPriorityLow    RecommendationPriority = "low"
)

// Evidence is one deterministic fact that supports a finding.
type Evidence struct {
	Type    EvidenceType
	Message string
	Path    string
	Value   string
}

// Recommendation is one actionable next step linked to a finding.
type Recommendation struct {
	FindingID string
	Title     string
	Action    string
	Rationale string
	Priority  RecommendationPriority
}

// Finding is a verifiable repository issue produced by an analyzer.
type Finding struct {
	ID              string
	RuleID          string
	AnalyzerID      string
	Severity        rules.Severity
	Title           string
	Message         string
	Category        rules.Category
	Status          FindingStatus
	Evidence        []Evidence
	Recommendations []Recommendation
}

// NewEvidence creates one evidence item.
func NewEvidence(evidenceType EvidenceType, message string, path string, value string) Evidence {
	return Evidence{
		Type:    evidenceType,
		Message: message,
		Path:    path,
		Value:   value,
	}
}

// NewRecommendation creates one recommendation item.
func NewRecommendation(
	findingID string,
	title string,
	action string,
	rationale string,
	priority RecommendationPriority,
) Recommendation {
	return Recommendation{
		FindingID: findingID,
		Title:     title,
		Action:    action,
		Rationale: rationale,
		Priority:  priority,
	}
}

// NewFinding creates a finding with the default open status.
func NewFinding(
	id string,
	ruleID string,
	analyzerID string,
	severity rules.Severity,
	category rules.Category,
	title string,
	message string,
	evidence []Evidence,
) Finding {
	return Finding{
		ID:         id,
		RuleID:     ruleID,
		AnalyzerID: analyzerID,
		Severity:   severity,
		Title:      title,
		Message:    message,
		Category:   category,
		Status:     FindingStatusOpen,
		Evidence:   evidence,
	}
}

// Validate checks whether evidence has required fields for its type.
func (e Evidence) Validate() error {
	if !e.Type.Valid() {
		return fmt.Errorf("evidence type %q is invalid", e.Type)
	}
	if e.Message == "" {
		return fmt.Errorf("evidence message is empty")
	}
	if e.requiresPath() && e.Path == "" {
		return fmt.Errorf("evidence path is empty")
	}
	if e.Path != "" && filepath.IsAbs(e.Path) {
		return fmt.Errorf("evidence path must be repository-relative")
	}
	if e.requiresValue() && e.Value == "" {
		return fmt.Errorf("evidence value is empty")
	}
	return nil
}

// Validate checks whether recommendation has required actionable fields.
func (r Recommendation) Validate() error {
	if r.FindingID == "" {
		return fmt.Errorf("recommendation finding id is empty")
	}
	if r.Title == "" {
		return fmt.Errorf("recommendation title is empty")
	}
	if r.Action == "" {
		return fmt.Errorf("recommendation action is empty")
	}
	if r.Rationale == "" {
		return fmt.Errorf("recommendation rationale is empty")
	}
	if !r.Priority.Valid() {
		return fmt.Errorf("recommendation priority %q is invalid", r.Priority)
	}
	return nil
}

// Validate checks whether a finding has required stable fields.
func (f Finding) Validate() error {
	if f.ID == "" {
		return fmt.Errorf("finding id is empty")
	}
	if f.RuleID == "" {
		return fmt.Errorf("finding rule id is empty")
	}
	if f.AnalyzerID == "" {
		return fmt.Errorf("finding analyzer id is empty")
	}
	if !f.Severity.Valid() {
		return fmt.Errorf("finding severity %q is invalid", f.Severity)
	}
	if f.Title == "" {
		return fmt.Errorf("finding title is empty")
	}
	if f.Message == "" {
		return fmt.Errorf("finding message is empty")
	}
	if !f.Category.Valid() {
		return fmt.Errorf("finding category %q is invalid", f.Category)
	}
	if !f.Status.Valid() {
		return fmt.Errorf("finding status %q is invalid", f.Status)
	}
	if len(f.Evidence) == 0 {
		return fmt.Errorf("finding evidence is empty")
	}
	for _, evidence := range f.Evidence {
		if err := evidence.Validate(); err != nil {
			return err
		}
	}
	for _, recommendation := range f.Recommendations {
		if err := recommendation.Validate(); err != nil {
			return err
		}
	}
	return nil
}

// Valid reports whether the evidence type is supported.
func (e EvidenceType) Valid() bool {
	switch e {
	case EvidenceTypeFileExists, EvidenceTypeFileMissing, EvidenceTypePatternMatch, EvidenceTypePatternMissing, EvidenceTypeMetadata:
		return true
	default:
		return false
	}
}

// Valid reports whether the finding status is supported.
func (s FindingStatus) Valid() bool {
	switch s {
	case FindingStatusOpen:
		return true
	default:
		return false
	}
}

// Valid reports whether recommendation priority is supported.
func (p RecommendationPriority) Valid() bool {
	switch p {
	case RecommendationPriorityHigh, RecommendationPriorityMedium, RecommendationPriorityLow:
		return true
	default:
		return false
	}
}

func (e Evidence) requiresPath() bool {
	switch e.Type {
	case EvidenceTypeFileExists, EvidenceTypeFileMissing, EvidenceTypePatternMatch, EvidenceTypePatternMissing:
		return true
	default:
		return false
	}
}

func (e Evidence) requiresValue() bool {
	switch e.Type {
	case EvidenceTypePatternMatch, EvidenceTypePatternMissing, EvidenceTypeMetadata:
		return true
	default:
		return false
	}
}
