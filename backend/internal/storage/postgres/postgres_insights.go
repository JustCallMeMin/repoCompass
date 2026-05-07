package postgres

import (
	"context"
	"database/sql"
	"errors"
	"fmt"

	"github.com/JustCallMeMin/repoCompass/backend/internal/insights"
)

// GetOrganizationInsights returns aggregated statistics for an organization's repositories and scans.
func (s *Store) GetOrganizationInsights(ctx context.Context, orgID string) (insights.OrganizationInsights, error) {
	query := `
		SELECT 
			COUNT(DISTINCT r.id) AS total_repositories,
			COUNT(DISTINCT s.id) AS total_scans,
			COALESCE(AVG(a.overall_score), 0) AS average_score
		FROM repositories r
		LEFT JOIN repository_snapshots rs ON rs.repository_id = r.id
		LEFT JOIN scans s ON s.snapshot_id = rs.id AND s.status = 'completed'
		LEFT JOIN assessments a ON a.scan_id = s.id
		WHERE r.organization_id = $1
	`
	
	var result insights.OrganizationInsights
	result.OrganizationID = orgID
	
	var avgScore float64
	err := s.db.QueryRowContext(ctx, query, orgID).Scan(
		&result.TotalRepositories,
		&result.TotalScans,
		&avgScore,
	)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return result, nil // Return zero values
		}
		return result, fmt.Errorf("postgres: GetOrganizationInsights: %w", err)
	}
	
	result.AverageScore = int(avgScore)
	
	return result, nil
}
