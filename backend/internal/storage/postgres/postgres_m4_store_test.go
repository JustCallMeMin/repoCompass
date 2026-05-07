package postgres_test

import (
	"context"
	"testing"
	"time"

	"github.com/JustCallMeMin/repoCompass/backend/internal/auth"
	ghintegration "github.com/JustCallMeMin/repoCompass/backend/internal/integration/github"
	"github.com/JustCallMeMin/repoCompass/backend/internal/org"
	"github.com/JustCallMeMin/repoCompass/backend/internal/repository"
	pgstore "github.com/JustCallMeMin/repoCompass/backend/internal/storage/postgres"
)

func TestPostgresStore_M4SessionAndGitHubJobCRUD(t *testing.T) {
	store := openDB(t)
	ctx := context.Background()
	now := time.Now().UTC().Truncate(time.Second)
	orgID := testID("org-m4")
	repoID := testID("repo-m4")

	if err := store.SaveOrganization(ctx, org.Organization{ID: orgID, Name: "M4 Org", CreatedAt: now, UpdatedAt: now}); err != nil {
		t.Fatalf("SaveOrganization: %v", err)
	}
	if err := store.SaveRepository(ctx, repository.Repository{
		ID:             repoID,
		Name:           "repo",
		OwnerName:      "owner",
		FullName:       "owner/repo",
		URL:            "https://github.com/owner/repo",
		Provider:       repository.Provider("github"),
		Status:         repository.StatusActive,
		OrganizationID: orgID,
	}); err != nil {
		t.Fatalf("SaveRepository: %v", err)
	}

	user := auth.User{ID: testID("user-m4"), GitHubID: testID("gh"), Login: "octo", CreatedAt: now, UpdatedAt: now}
	if err := store.SaveUser(ctx, user); err != nil {
		t.Fatalf("SaveUser: %v", err)
	}
	session := auth.Session{ID: testID("sess-m4"), UserID: user.ID, TokenHash: pgstore.HashSessionToken("token"), ExpiresAt: now.Add(time.Hour), CreatedAt: now}
	if err := store.SaveSession(ctx, session); err != nil {
		t.Fatalf("SaveSession: %v", err)
	}
	gotSession, err := store.GetSessionByTokenHash(ctx, pgstore.HashSessionToken("token"))
	if err != nil {
		t.Fatalf("GetSessionByTokenHash: %v", err)
	}
	if gotSession.UserID != user.ID {
		t.Fatalf("unexpected session user: %s", gotSession.UserID)
	}

	integration := ghintegration.Integration{ID: testID("ghint"), OrganizationID: orgID, RepositoryID: repoID, Owner: "owner", Repo: "repo", CloneURL: "https://github.com/owner/repo", Active: true, CreatedAt: now, UpdatedAt: now}
	if err := store.SaveGitHubIntegration(ctx, integration); err != nil {
		t.Fatalf("SaveGitHubIntegration: %v", err)
	}
	if _, err := store.FindGitHubIntegration(ctx, orgID, "owner", "repo"); err != nil {
		t.Fatalf("FindGitHubIntegration: %v", err)
	}

	event := ghintegration.WebhookEvent{ID: testID("ghevt"), DeliveryID: testID("delivery"), EventType: "push", RepositoryFullName: "owner/repo", RepositoryCloneURL: "https://github.com/owner/repo", Payload: []byte(`{"ok":true}`), Status: ghintegration.WebhookStatusAccepted, CreatedAt: now}
	if err := store.SaveWebhookEvent(ctx, event); err != nil {
		t.Fatalf("SaveWebhookEvent: %v", err)
	}
	job := ghintegration.ScanJob{ID: testID("job"), RepositoryID: repoID, SourceType: "github", SourceURL: "https://github.com/owner/repo", Status: ghintegration.JobStatusQueued, WebhookEventID: event.ID, CreatedAt: now}
	if err := store.SaveScanJob(ctx, job); err != nil {
		t.Fatalf("SaveScanJob: %v", err)
	}
	claimed, err := store.ClaimNextQueuedScanJob(ctx)
	if err != nil {
		t.Fatalf("ClaimNextQueuedScanJob: %v", err)
	}
	if claimed.ID != job.ID || claimed.Status != ghintegration.JobStatusRunning {
		t.Fatalf("unexpected claimed job: %#v", claimed)
	}
	if err := store.FailScanJob(ctx, job.ID, "scan failed"); err != nil {
		t.Fatalf("FailScanJob: %v", err)
	}
}
