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
}

// Provider defines the interface for fetching organization-level insights.
type Provider interface {
	GetOrganizationInsights(ctx context.Context, orgID string) (OrganizationInsights, error)
}
