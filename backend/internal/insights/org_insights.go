package insights

import (
	"context"
)

// OrganizationInsights contains aggregated benchmark and health data for an organization.
type OrganizationInsights struct {
	OrganizationID    string `json:"organization_id"`
	AverageScore      int    `json:"average_score"`
	TotalRepositories int    `json:"total_repositories"`
	TotalScans        int    `json:"total_scans"`
	HighRiskCount     int    `json:"high_risk_count"`
	StaleScanCount    int    `json:"stale_scan_count"`
	TopRepositoryID   string `json:"top_repository_id,omitempty"`
	LowestRepositoryID string `json:"lowest_repository_id,omitempty"`
	Insights          []Insight `json:"insights"`
}

// Insight is a deterministic organization-level recommendation.
type Insight struct {
	Severity      string `json:"severity"`
	Title         string `json:"title"`
	Explanation   string `json:"explanation"`
	NextAction    string `json:"next_action"`
	RepositoryID  string `json:"repository_id,omitempty"`
	PolicyID      string `json:"policy_id,omitempty"`
}

// Provider defines the interface for fetching organization-level insights.
type Provider interface {
	GetOrganizationInsights(ctx context.Context, orgID string) (OrganizationInsights, error)
}
