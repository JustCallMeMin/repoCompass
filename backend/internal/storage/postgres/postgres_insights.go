package postgres

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"time"

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
	if err := s.decorateOrganizationInsights(ctx, orgID, &result); err != nil {
		return result, err
	}

	return result, nil
}

func (s *Store) decorateOrganizationInsights(ctx context.Context, orgID string, result *insights.OrganizationInsights) error {
	rows, err := s.db.QueryContext(ctx, `
		WITH latest AS (
			SELECT DISTINCT ON (r.id)
				r.id AS repository_id,
				COALESCE(a.overall_score, 0) AS score,
				s.end_time
			FROM repositories r
			LEFT JOIN repository_snapshots rs ON rs.repository_id = r.id
			LEFT JOIN scans s ON s.snapshot_id = rs.id AND s.status = 'completed'
			LEFT JOIN assessments a ON a.scan_id = s.id
			WHERE r.organization_id = $1
			ORDER BY r.id, s.end_time DESC NULLS LAST
		)
		SELECT repository_id, score, end_time
		FROM latest
		ORDER BY score ASC, repository_id ASC
	`, orgID)
	if err != nil {
		return fmt.Errorf("postgres: organization insight details: %w", err)
	}
	defer rows.Close()
	var lowScoreID, highScoreID string
	var lowScore, highScore int
	first := true
	for rows.Next() {
		var repoID string
		var score int
		var endedAt sql.NullTime
		if err := rows.Scan(&repoID, &score, &endedAt); err != nil {
			return fmt.Errorf("postgres: scan organization insight detail: %w", err)
		}
		if first {
			lowScoreID, lowScore = repoID, score
			first = false
		}
		highScoreID, highScore = repoID, score
		if score > 0 && score < 70 {
			result.HighRiskCount++
		}
		if !endedAt.Valid || endedAt.Time.Before(time.Now().UTC().AddDate(0, 0, -14)) {
			result.StaleScanCount++
		}
	}
	if err := rows.Err(); err != nil {
		return fmt.Errorf("postgres: organization insight rows: %w", err)
	}
	result.LowestRepositoryID = lowScoreID
	result.TopRepositoryID = highScoreID
	if result.HighRiskCount > 0 {
		result.Insights = append(result.Insights, insights.Insight{
			Severity:     "critical",
			Title:        "Repositories below baseline",
			Explanation:  fmt.Sprintf("%d repositories have latest scores below 70.", result.HighRiskCount),
			NextAction:   "Open the lowest ranked repository and address high severity findings first.",
			RepositoryID: lowScoreID,
		})
	}
	if result.StaleScanCount > 0 {
		result.Insights = append(result.Insights, insights.Insight{
			Severity:    "warning",
			Title:       "Stale or missing scans",
			Explanation: fmt.Sprintf("%d repositories have no completed scan in the last 14 days.", result.StaleScanCount),
			NextAction:  "Trigger fresh scans for stale repositories before comparing benchmark scores.",
		})
	}
	if result.TotalRepositories > 1 && highScore-lowScore >= 20 {
		result.Insights = append(result.Insights, insights.Insight{
			Severity:     "info",
			Title:        "Large score spread",
			Explanation:  fmt.Sprintf("The org score spread is %d points between best and lowest repository.", highScore-lowScore),
			NextAction:   "Use the top repository as onboarding baseline for lower ranked repositories.",
			RepositoryID: lowScoreID,
		})
	}
	return nil
}
