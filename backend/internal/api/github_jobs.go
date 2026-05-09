package api

import (
	"context"
	"fmt"
	"strings"
	"time"

	ghintegration "github.com/JustCallMeMin/repoCompass/backend/internal/integration/github"
	"github.com/JustCallMeMin/repoCompass/backend/internal/org"
	"github.com/JustCallMeMin/repoCompass/backend/internal/repository"
	"github.com/JustCallMeMin/repoCompass/backend/internal/scan"
)

// RepositoryWriter persists repository metadata created by integration intake.
type RepositoryWriter interface {
	SaveRepository(ctx context.Context, repo repository.Repository) error
}

// acceptGitHubNonPush persists a valid GitHub delivery that does not enqueue a scan.
func (s *Server) acceptGitHubNonPush(ctx context.Context, value ghintegration.WebhookRequest, status string) (string, error) {
	if s.integrations == nil {
		return "", fmt.Errorf("github integration store is not configured")
	}
	now := time.Now().UTC()
	event := ghintegration.WebhookEvent{
		ID:                 "ghevt_" + stableHash(value.DeliveryID),
		DeliveryID:         value.DeliveryID,
		EventType:          value.Event,
		RepositoryFullName: strings.TrimSpace(value.Payload.Repository.FullName),
		RepositoryCloneURL: strings.TrimSpace(value.Payload.Repository.CloneURL),
		Payload:            value.Body,
		Status:             status,
		CreatedAt:          now,
	}
	if err := s.integrations.SaveWebhookEvent(ctx, event); err != nil {
		return "", err
	}
	return event.ID, nil
}

// acceptGitHubPush persists a push event and queues a scan job.
func (s *Server) acceptGitHubPush(ctx context.Context, value ghintegration.WebhookRequest) (map[string]string, error) {
	if s.integrations == nil {
		return nil, fmt.Errorf("github integration store is not configured")
	}
	fullName := strings.TrimSpace(value.Payload.Repository.FullName)
	cloneURL := strings.TrimSpace(value.Payload.Repository.CloneURL)
	if fullName == "" || cloneURL == "" {
		return nil, fmt.Errorf("repository full_name and clone_url are required")
	}
	owner, repoName, ok := strings.Cut(fullName, "/")
	if !ok || owner == "" || repoName == "" {
		return nil, fmt.Errorf("repository full_name must be owner/repo")
	}
	now := time.Now().UTC()
	event := ghintegration.WebhookEvent{
		ID:                 "ghevt_" + stableHash(value.DeliveryID),
		DeliveryID:         value.DeliveryID,
		EventType:          value.Event,
		RepositoryFullName: fullName,
		RepositoryCloneURL: cloneURL,
		Payload:            value.Body,
		Status:             ghintegration.WebhookStatusAccepted,
		CreatedAt:          now,
	}
	if err := s.integrations.SaveWebhookEvent(ctx, event); err != nil {
		return nil, err
	}

	integration, err := s.integrations.FindGitHubIntegration(ctx, org.DefaultPersonalOrgID, owner, repoName)
	if err != nil {
		repo := repository.Repository{
			ID:             "repo_github_" + stableHash(fullName),
			Name:           repoName,
			OwnerName:      owner,
			FullName:       fullName,
			URL:            cloneURL,
			Provider:       repository.Provider("github"),
			Status:         repository.StatusActive,
			OrganizationID: org.DefaultPersonalOrgID,
		}
		if writer, ok := s.orgs.(RepositoryWriter); ok {
			if err := writer.SaveRepository(ctx, repo); err != nil {
				return nil, err
			}
		}
		integration = ghintegration.Integration{
			ID:             "ghint_" + stableHash(fullName),
			OrganizationID: org.DefaultPersonalOrgID,
			RepositoryID:   repo.ID,
			Owner:          owner,
			Repo:           repoName,
			CloneURL:       cloneURL,
			Active:         true,
			CreatedAt:      now,
			UpdatedAt:      now,
		}
		if err := s.integrations.SaveGitHubIntegration(ctx, integration); err != nil {
			return nil, err
		}
	}

	job := ghintegration.ScanJob{
		ID:             "job_" + stableHash(event.ID),
		RepositoryID:   integration.RepositoryID,
		SourceType:     "github",
		SourceURL:      cloneURL,
		Status:         ghintegration.JobStatusQueued,
		WebhookEventID: event.ID,
		CreatedAt:      now,
	}
	if err := s.integrations.SaveScanJob(ctx, job); err != nil {
		return nil, err
	}
	return map[string]string{"event_id": event.ID, "job_id": job.ID, "status": job.Status}, nil
}

// RunNextScanJob executes one queued scan job.
func (s *Server) RunNextScanJob(ctx context.Context) error {
	if s.integrations == nil || s.runner == nil {
		return fmt.Errorf("scan job dependencies are not configured")
	}
	job, err := s.integrations.ClaimNextQueuedScanJob(ctx)
	if err != nil {
		return err
	}
	source := repository.RepositorySource{Type: repository.SourceTypeLocal, Path: job.SourceURL, URL: job.SourceURL, OrganizationID: org.DefaultPersonalOrgID}
	var cleanup func()
	if job.SourceType == "github" {
		if s.github == nil {
			_ = s.integrations.FailScanJob(ctx, job.ID, "github cloner is not configured")
			return fmt.Errorf("github cloner is not configured")
		}
		repoURL, err := ghintegration.ParseRepositoryURL(job.SourceURL)
		if err != nil {
			_ = s.integrations.FailScanJob(ctx, job.ID, err.Error())
			return err
		}
		checkout, err := s.github.Clone(ctx, repoURL)
		if err != nil {
			_ = s.integrations.FailScanJob(ctx, job.ID, err.Error())
			return err
		}
		cleanup = checkout.Cleanup
		source.Path = checkout.Path
	}
	if cleanup != nil {
		defer cleanup()
	}
	result, err := s.runner.Run(ctx, scan.RunRequest{Source: source})
	if err != nil {
		_ = s.integrations.FailScanJob(ctx, job.ID, err.Error())
		return err
	}
	return s.integrations.CompleteScanJob(ctx, job.ID, result.Scan.ID)
}
