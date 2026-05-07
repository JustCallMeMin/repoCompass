package postgres

import (
	"context"
	"crypto/sha256"
	"database/sql"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"time"

	"github.com/JustCallMeMin/repoCompass/backend/internal/assessment"
	"github.com/JustCallMeMin/repoCompass/backend/internal/auth"
	ghintegration "github.com/JustCallMeMin/repoCompass/backend/internal/integration/github"
	"github.com/JustCallMeMin/repoCompass/backend/internal/scan"
)

// HashSessionToken returns a stable SHA-256 hash for session token storage.
func HashSessionToken(token string) string {
	sum := sha256.Sum256([]byte(token))
	return hex.EncodeToString(sum[:])
}

// SaveUser creates or updates one product API user.
func (s *Store) SaveUser(ctx context.Context, u auth.User) error {
	_, err := s.db.ExecContext(ctx, `
		INSERT INTO users (id, github_id, login, name, avatar_url, email, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5, $6, COALESCE($7, NOW()), COALESCE($8, NOW()))
		ON CONFLICT (id) DO UPDATE SET
			github_id = EXCLUDED.github_id,
			login = EXCLUDED.login,
			name = EXCLUDED.name,
			avatar_url = EXCLUDED.avatar_url,
			email = EXCLUDED.email,
			updated_at = NOW()
	`, u.ID, u.GitHubID, u.Login, u.Name, u.AvatarURL, u.Email, nullableTime(u.CreatedAt), nullableTime(u.UpdatedAt))
	if err != nil {
		return fmt.Errorf("postgres: save user: %w", err)
	}
	return nil
}

// GetUser returns one user by ID.
func (s *Store) GetUser(ctx context.Context, id string) (auth.User, error) {
	var u auth.User
	err := s.db.QueryRowContext(ctx, `
		SELECT id, github_id, login, name, avatar_url, email, created_at, updated_at
		FROM users WHERE id = $1
	`, id).Scan(&u.ID, &u.GitHubID, &u.Login, &u.Name, &u.AvatarURL, &u.Email, &u.CreatedAt, &u.UpdatedAt)
	if errors.Is(err, sql.ErrNoRows) {
		return auth.User{}, fmt.Errorf("postgres: user not found: %s", id)
	}
	if err != nil {
		return auth.User{}, fmt.Errorf("postgres: get user: %w", err)
	}
	return u, nil
}

// SaveSession creates or updates one login session.
func (s *Store) SaveSession(ctx context.Context, sess auth.Session) error {
	_, err := s.db.ExecContext(ctx, `
		INSERT INTO sessions (id, user_id, token_hash, expires_at, created_at, revoked_at)
		VALUES ($1, $2, $3, $4, COALESCE($5, NOW()), $6)
		ON CONFLICT (id) DO UPDATE SET
			token_hash = EXCLUDED.token_hash,
			expires_at = EXCLUDED.expires_at,
			revoked_at = EXCLUDED.revoked_at
	`, sess.ID, sess.UserID, sess.TokenHash, sess.ExpiresAt, nullableTime(sess.CreatedAt), sess.RevokedAt)
	if err != nil {
		return fmt.Errorf("postgres: save session: %w", err)
	}
	return nil
}

// GetSessionByTokenHash returns an active session by token hash.
func (s *Store) GetSessionByTokenHash(ctx context.Context, tokenHash string) (auth.Session, error) {
	var sess auth.Session
	err := s.db.QueryRowContext(ctx, `
		SELECT id, user_id, token_hash, expires_at, created_at, revoked_at
		FROM sessions
		WHERE token_hash = $1 AND revoked_at IS NULL AND expires_at > NOW()
	`, tokenHash).Scan(&sess.ID, &sess.UserID, &sess.TokenHash, &sess.ExpiresAt, &sess.CreatedAt, &sess.RevokedAt)
	if errors.Is(err, sql.ErrNoRows) {
		return auth.Session{}, fmt.Errorf("postgres: session not found")
	}
	if err != nil {
		return auth.Session{}, fmt.Errorf("postgres: get session: %w", err)
	}
	return sess, nil
}

// RevokeSession marks one session revoked.
func (s *Store) RevokeSession(ctx context.Context, id string) error {
	result, err := s.db.ExecContext(ctx, `UPDATE sessions SET revoked_at = NOW() WHERE id = $1`, id)
	if err != nil {
		return fmt.Errorf("postgres: revoke session: %w", err)
	}
	if rows, _ := result.RowsAffected(); rows == 0 {
		return fmt.Errorf("postgres: session not found: %s", id)
	}
	return nil
}

// SaveGitHubIntegration creates or updates a GitHub integration.
func (s *Store) SaveGitHubIntegration(ctx context.Context, value ghintegration.Integration) error {
	_, err := s.db.ExecContext(ctx, `
		INSERT INTO github_integrations (id, organization_id, repository_id, owner, repo, clone_url, active, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7, COALESCE($8, NOW()), COALESCE($9, NOW()))
		ON CONFLICT (organization_id, owner, repo) DO UPDATE SET
			repository_id = EXCLUDED.repository_id,
			clone_url = EXCLUDED.clone_url,
			active = EXCLUDED.active,
			updated_at = NOW()
	`, value.ID, value.OrganizationID, value.RepositoryID, value.Owner, value.Repo, value.CloneURL, value.Active, nullableTime(value.CreatedAt), nullableTime(value.UpdatedAt))
	if err != nil {
		return fmt.Errorf("postgres: save github integration: %w", err)
	}
	return nil
}

// FindGitHubIntegration returns an integration by org and GitHub full name.
func (s *Store) FindGitHubIntegration(ctx context.Context, orgID, owner, repoName string) (ghintegration.Integration, error) {
	var value ghintegration.Integration
	err := s.db.QueryRowContext(ctx, `
		SELECT id, organization_id, repository_id, owner, repo, clone_url, active, created_at, updated_at
		FROM github_integrations
		WHERE organization_id = $1 AND owner = $2 AND repo = $3 AND active = TRUE
	`, orgID, owner, repoName).Scan(&value.ID, &value.OrganizationID, &value.RepositoryID, &value.Owner, &value.Repo, &value.CloneURL, &value.Active, &value.CreatedAt, &value.UpdatedAt)
	if errors.Is(err, sql.ErrNoRows) {
		return ghintegration.Integration{}, fmt.Errorf("postgres: github integration not found")
	}
	if err != nil {
		return ghintegration.Integration{}, fmt.Errorf("postgres: find github integration: %w", err)
	}
	return value, nil
}

// SaveWebhookEvent persists one GitHub webhook delivery.
func (s *Store) SaveWebhookEvent(ctx context.Context, event ghintegration.WebhookEvent) error {
	if len(event.Payload) == 0 {
		event.Payload = []byte(`{}`)
	}
	if !json.Valid(event.Payload) {
		event.Payload = []byte(`{}`)
	}
	_, err := s.db.ExecContext(ctx, `
		INSERT INTO github_webhook_events (id, delivery_id, event_type, repository_full_name, repository_clone_url, payload, status, error_message, created_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, COALESCE($9, NOW()))
	`, event.ID, event.DeliveryID, event.EventType, event.RepositoryFullName, event.RepositoryCloneURL, event.Payload, event.Status, event.ErrorMessage, nullableTime(event.CreatedAt))
	if err != nil {
		return fmt.Errorf("postgres: save webhook event: %w", err)
	}
	return nil
}

// SaveScanJob persists one scan job.
func (s *Store) SaveScanJob(ctx context.Context, job ghintegration.ScanJob) error {
	_, err := s.db.ExecContext(ctx, `
		INSERT INTO scan_jobs (id, repository_id, source_type, source_url, status, scan_id, webhook_event_id, error_message, created_at, started_at, completed_at)
		VALUES ($1, $2, $3, $4, $5, NULLIF($6, ''), NULLIF($7, ''), $8, COALESCE($9, NOW()), $10, $11)
		ON CONFLICT (id) DO UPDATE SET
			status = EXCLUDED.status,
			scan_id = EXCLUDED.scan_id,
			error_message = EXCLUDED.error_message,
			started_at = EXCLUDED.started_at,
			completed_at = EXCLUDED.completed_at
	`, job.ID, job.RepositoryID, job.SourceType, job.SourceURL, job.Status, job.ScanID, job.WebhookEventID, job.ErrorMessage, nullableTime(job.CreatedAt), job.StartedAt, job.CompletedAt)
	if err != nil {
		return fmt.Errorf("postgres: save scan job: %w", err)
	}
	return nil
}

// ClaimNextQueuedScanJob marks the oldest queued job as running and returns it.
func (s *Store) ClaimNextQueuedScanJob(ctx context.Context) (ghintegration.ScanJob, error) {
	var job ghintegration.ScanJob
	err := s.db.QueryRowContext(ctx, `
		UPDATE scan_jobs
		SET status = 'running', started_at = NOW()
		WHERE id = (
			SELECT id FROM scan_jobs
			WHERE status = 'queued'
			ORDER BY created_at ASC
			LIMIT 1
			FOR UPDATE SKIP LOCKED
		)
		RETURNING id, repository_id, source_type, source_url, status, COALESCE(scan_id, ''), COALESCE(webhook_event_id, ''), error_message, created_at, started_at, completed_at
	`).Scan(&job.ID, &job.RepositoryID, &job.SourceType, &job.SourceURL, &job.Status, &job.ScanID, &job.WebhookEventID, &job.ErrorMessage, &job.CreatedAt, &job.StartedAt, &job.CompletedAt)
	if errors.Is(err, sql.ErrNoRows) {
		return ghintegration.ScanJob{}, fmt.Errorf("postgres: scan job not found")
	}
	if err != nil {
		return ghintegration.ScanJob{}, fmt.Errorf("postgres: claim scan job: %w", err)
	}
	return job, nil
}

// CompleteScanJob marks a running job completed.
func (s *Store) CompleteScanJob(ctx context.Context, jobID, scanID string) error {
	_, err := s.db.ExecContext(ctx, `
		UPDATE scan_jobs
		SET status = 'completed', scan_id = $2, completed_at = NOW(), error_message = ''
		WHERE id = $1
	`, jobID, scanID)
	return err
}

// FailScanJob marks a running job failed.
func (s *Store) FailScanJob(ctx context.Context, jobID, message string) error {
	_, err := s.db.ExecContext(ctx, `
		UPDATE scan_jobs
		SET status = 'failed', error_message = $2, completed_at = NOW()
		WHERE id = $1
	`, jobID, message)
	return err
}

// GetScan returns one scan row by ID.
func (s *Store) GetScan(ctx context.Context, scanID string) (scan.Scan, error) {
	var value scan.Scan
	var status string
	err := s.db.QueryRowContext(ctx, `
		SELECT id, snapshot_id, status, start_time, end_time, error_details
		FROM scans WHERE id = $1
	`, scanID).Scan(&value.ID, &value.SnapshotID, &status, &value.StartTime, &value.EndTime, &value.ErrorDetails)
	if errors.Is(err, sql.ErrNoRows) {
		return scan.Scan{}, fmt.Errorf("postgres: scan not found: %s", scanID)
	}
	if err != nil {
		return scan.Scan{}, fmt.Errorf("postgres: get scan: %w", err)
	}
	value.Status = scan.Status(status)
	return value, nil
}

// GetAssessment returns one persisted assessment by scan ID.
func (s *Store) GetAssessment(ctx context.Context, scanID string) (assessment.Assessment, error) {
	var value assessment.Assessment
	var severityRaw, categoryScoresRaw, categoryBreakdownRaw []byte
	err := s.db.QueryRowContext(ctx, `
		SELECT overall_score, label, finding_count, severity_counts, category_scores, category_breakdown
		FROM assessments WHERE scan_id = $1
	`, scanID).Scan(&value.OverallScore, &value.Label, &value.FindingCount, &severityRaw, &categoryScoresRaw, &categoryBreakdownRaw)
	if errors.Is(err, sql.ErrNoRows) {
		return assessment.Assessment{}, fmt.Errorf("postgres: assessment not found: %s", scanID)
	}
	if err != nil {
		return assessment.Assessment{}, fmt.Errorf("postgres: get assessment: %w", err)
	}
	_ = json.Unmarshal(severityRaw, &value.SeverityCounts)
	_ = json.Unmarshal(categoryScoresRaw, &value.CategoryScores)
	_ = json.Unmarshal(categoryBreakdownRaw, &value.CategoryBreakdown)
	return value, nil
}

// ListReports returns report metadata rows for one scan.
func (s *Store) ListReports(ctx context.Context, scanID string) ([]map[string]any, error) {
	rows, err := s.db.QueryContext(ctx, `
		SELECT id, scan_id, format, content_type, metadata, created_at
		FROM reports
		WHERE scan_id = $1
		ORDER BY format ASC
	`, scanID)
	if err != nil {
		return nil, fmt.Errorf("postgres: list reports: %w", err)
	}
	defer rows.Close()

	var reports []map[string]any
	for rows.Next() {
		var id int64
		var scanIDValue, format, contentType string
		var metadataRaw []byte
		var createdAt time.Time
		if err := rows.Scan(&id, &scanIDValue, &format, &contentType, &metadataRaw, &createdAt); err != nil {
			return nil, fmt.Errorf("postgres: scan report: %w", err)
		}
		var metadata map[string]any
		_ = json.Unmarshal(metadataRaw, &metadata)
		reports = append(reports, map[string]any{
			"id":           id,
			"scan_id":      scanIDValue,
			"format":       format,
			"content_type": contentType,
			"metadata":     metadata,
			"created_at":   createdAt,
		})
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("postgres: reports rows: %w", err)
	}
	return reports, nil
}

func nullableTime(value time.Time) any {
	if value.IsZero() {
		return nil
	}
	return value
}
