// Package history contains read models for persisted scan history.
package history

import (
	"context"
	"time"
)

// Reader defines history read operations for persisted scan data.
type Reader interface {
	ListScanHistory(ctx context.Context, repositoryID string, limit int) ([]ScanSummary, error)
	LatestScan(ctx context.Context, repositoryID string) (ScanSummary, error)
	ListFindings(ctx context.Context, scanID string) ([]FindingDetail, error)
	ListMetricTrend(ctx context.Context, repositoryID string, metricKey string, limit int) ([]MetricPoint, error)
}

// ScanSummary is one persisted scan row plus joined repository assessment data.
type ScanSummary struct {
	ScanID       string     `json:"scan_id"`
	RepositoryID string     `json:"repository_id"`
	SnapshotID   string     `json:"snapshot_id"`
	Status       string     `json:"status"`
	StartedAt    *time.Time `json:"started_at,omitempty"`
	CompletedAt  *time.Time `json:"completed_at,omitempty"`
	Score        int        `json:"score"`
	Label        string     `json:"label"`
	FindingCount int        `json:"finding_count"`
}

// FindingDetail is one persisted finding row with evidence and recommendations.
type FindingDetail struct {
	ID              string                 `json:"id"`
	ScanID          string                 `json:"scan_id"`
	RuleID          string                 `json:"rule_id"`
	AnalyzerID      string                 `json:"analyzer_id"`
	Severity        string                 `json:"severity"`
	Title           string                 `json:"title"`
	Message         string                 `json:"message"`
	Category        string                 `json:"category"`
	Status          string                 `json:"status"`
	Evidence        []EvidenceDetail       `json:"evidence"`
	Recommendations []RecommendationDetail `json:"recommendations"`
}

// EvidenceDetail is one stored fact supporting a finding.
type EvidenceDetail struct {
	Type    string `json:"type"`
	Message string `json:"message"`
	Path    string `json:"path,omitempty"`
	Value   string `json:"value,omitempty"`
}

// RecommendationDetail is one stored remediation action for a finding.
type RecommendationDetail struct {
	Title     string `json:"title"`
	Action    string `json:"action"`
	Rationale string `json:"rationale"`
	Priority  string `json:"priority"`
}

// MetricPoint is one metric snapshot for a repository scan.
type MetricPoint struct {
	ScanID       string     `json:"scan_id"`
	MetricKey    string     `json:"metric_key"`
	Value        float64    `json:"value"`
	CapturedAt   time.Time  `json:"captured_at"`
	CompletedAt  *time.Time `json:"completed_at,omitempty"`
	RepositoryID string     `json:"repository_id"`
}

// SnapshotSummary is the latest persisted snapshot metadata for a repository.
type SnapshotSummary struct {
	SnapshotID    string    `json:"snapshot_id"`
	RepositoryID  string    `json:"repository_id"`
	SourceType    string    `json:"source_type"`
	BranchName    string    `json:"branch_name,omitempty"`
	CommitSHA     string    `json:"commit_sha,omitempty"`
	TreeReference string    `json:"tree_reference,omitempty"`
	CapturedAt    time.Time `json:"captured_at"`
}
