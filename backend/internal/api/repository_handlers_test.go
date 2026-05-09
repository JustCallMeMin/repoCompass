package api

import (
	"bytes"
	"context"
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"net/http"
	"net/http/httptest"
	"testing"

	ghintegration "github.com/JustCallMeMin/repoCompass/backend/internal/integration/github"
	"github.com/JustCallMeMin/repoCompass/backend/internal/org"
	"github.com/JustCallMeMin/repoCompass/backend/internal/repository"
)

func TestGitHubOAuthLoginPersistsState(t *testing.T) {
	t.Setenv("GITHUB_OAUTH_CLIENT_ID", "client-id")
	t.Setenv("GITHUB_OAUTH_REDIRECT_URL", "http://localhost:8080/api/v1/auth/github/callback")
	store := newFakeOrgStore()
	handler := NewServer(nil, nil, nil, store, nil).Handler()

	req := httptest.NewRequest(http.MethodGet, "/api/v1/auth/github/login?format=json", nil)
	rw := httptest.NewRecorder()
	handler.ServeHTTP(rw, req)

	if rw.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d: %s", rw.Code, rw.Body.String())
	}
	if len(store.oauthStates) != 1 {
		t.Fatalf("expected one oauth state, got %d", len(store.oauthStates))
	}
	if !bytes.Contains(rw.Body.Bytes(), []byte("https://github.com/login/oauth/authorize")) {
		t.Fatalf("expected authorization url: %s", rw.Body.String())
	}
}

func TestGitHubWebhookRejectsMissingSecret(t *testing.T) {
	server := NewServer(nil, nil, nil, newFakeOrgStore(), nil)
	handler := server.Handler()

	req := httptest.NewRequest(http.MethodPost, "/api/v1/integrations/github/webhook", bytes.NewReader([]byte(`{}`)))
	req.Header.Set("X-GitHub-Event", "ping")
	req.Header.Set("X-GitHub-Delivery", "delivery_missing_secret")
	rw := httptest.NewRecorder()
	handler.ServeHTTP(rw, req)

	if rw.Code != http.StatusServiceUnavailable {
		t.Fatalf("expected 503, got %d: %s", rw.Code, rw.Body.String())
	}
}

func TestGitHubWebhookRejectsInvalidSignature(t *testing.T) {
	server := NewServer(nil, nil, nil, newFakeOrgStore(), nil)
	server.SetGitHubWebhookSecret("secret")
	handler := server.Handler()

	req := httptest.NewRequest(http.MethodPost, "/api/v1/integrations/github/webhook", bytes.NewReader([]byte(`{}`)))
	req.Header.Set("X-GitHub-Event", "ping")
	req.Header.Set("X-GitHub-Delivery", "delivery_bad_sig")
	req.Header.Set("X-Hub-Signature-256", "sha256=bad")
	rw := httptest.NewRecorder()
	handler.ServeHTTP(rw, req)

	if rw.Code != http.StatusUnauthorized {
		t.Fatalf("expected 401, got %d: %s", rw.Code, rw.Body.String())
	}
}

func TestM4RepositoryAndScanEndpointsUseEnvelope(t *testing.T) {
	store := newFakeOrgStore()
	store.repos = []repository.Repository{{ID: "repo_1", Name: "alpha", FullName: "owner/alpha", LocalPath: "./testdata", Provider: repository.ProviderLocal, Status: repository.StatusActive, OrganizationID: org.DefaultPersonalOrgID}}
	store.memberships = []org.Membership{{OrganizationID: org.DefaultPersonalOrgID, UserID: "mock_user", Role: org.RoleOwner}}
	runner := fakeRunner{result: testRunResult("scan_1", "repo_1")}
	handler := newDevHeaderTestServer(&runner, &fakeHistoryReader{}, nil, store).Handler()

	for _, tc := range []struct {
		method string
		path   string
		body   string
		status int
		want   string
	}{
		{method: http.MethodGet, path: "/api/v1/health", status: http.StatusOK, want: "request_id"},
		{method: http.MethodGet, path: "/api/v1/repositories", status: http.StatusOK, want: "alpha"},
		{method: http.MethodGet, path: "/api/v1/repositories/repo_1", status: http.StatusOK, want: "owner/alpha"},
		{method: http.MethodPost, path: "/api/v1/repositories/repo_1/scans", status: http.StatusAccepted, want: "scan_1"},
		{method: http.MethodGet, path: "/api/v1/scans/scan_1", status: http.StatusOK, want: "scan_123"},
		{method: http.MethodGet, path: "/api/v1/scans/scan_1/assessment", status: http.StatusOK, want: "82"},
		{method: http.MethodGet, path: "/api/v1/scans/scan_1/reports", status: http.StatusOK, want: "json"},
	} {
		t.Run(tc.method+" "+tc.path, func(t *testing.T) {
			req := httptest.NewRequest(tc.method, tc.path, bytes.NewBufferString(tc.body))
			rw := httptest.NewRecorder()
			handler.ServeHTTP(rw, req)
			if rw.Code != tc.status {
				t.Fatalf("expected %d, got %d: %s", tc.status, rw.Code, rw.Body.String())
			}
			if !bytes.Contains(rw.Body.Bytes(), []byte(`"data"`)) || !bytes.Contains(rw.Body.Bytes(), []byte(`"meta"`)) {
				t.Fatalf("expected envelope, got %s", rw.Body.String())
			}
			if !bytes.Contains(rw.Body.Bytes(), []byte(tc.want)) {
				t.Fatalf("expected %q in response: %s", tc.want, rw.Body.String())
			}
		})
	}
}

func TestGitHubWebhookPushPersistsEventAndQueuesJob(t *testing.T) {
	store := newFakeOrgStore()
	store.memberships = []org.Membership{{OrganizationID: org.DefaultPersonalOrgID, UserID: "mock_user", Role: org.RoleOwner}}
	integrations := &fakeGitHubIntegrationStore{}
	server := newDevHeaderTestServer(nil, nil, nil, store)
	server.integrations = integrations
	server.SetGitHubWebhookSecret("secret")
	handler := server.Handler()

	body := []byte(`{"repository":{"full_name":"owner/repo","clone_url":"https://github.com/owner/repo"}}`)
	req := httptest.NewRequest(http.MethodPost, "/api/v1/integrations/github/webhook", bytes.NewReader(body))
	req.Header.Set("X-GitHub-Event", "push")
	req.Header.Set("X-GitHub-Delivery", "delivery_1")
	req.Header.Set("X-Hub-Signature-256", signWebhook(body, "secret"))
	rw := httptest.NewRecorder()

	handler.ServeHTTP(rw, req)

	if rw.Code != http.StatusAccepted {
		t.Fatalf("expected 202, got %d: %s", rw.Code, rw.Body.String())
	}
	if len(integrations.events) != 1 {
		t.Fatalf("expected one event, got %d", len(integrations.events))
	}
	if len(integrations.jobs) != 1 || integrations.jobs[0].Status != ghintegration.JobStatusQueued {
		t.Fatalf("expected queued job, got %#v", integrations.jobs)
	}
	if len(store.repos) != 1 || store.repos[0].Provider != repository.Provider("github") {
		t.Fatalf("expected github repo to be saved, got %#v", store.repos)
	}
}

func signWebhook(body []byte, secret string) string {
	mac := hmac.New(sha256.New, []byte(secret))
	_, _ = mac.Write(body)
	return "sha256=" + hex.EncodeToString(mac.Sum(nil))
}

type fakeGitHubIntegrationStore struct {
	integrations []ghintegration.Integration
	events       []ghintegration.WebhookEvent
	jobs         []ghintegration.ScanJob
}

func (f *fakeGitHubIntegrationStore) SaveGitHubIntegration(_ context.Context, value ghintegration.Integration) error {
	f.integrations = append(f.integrations, value)
	return nil
}

func (f *fakeGitHubIntegrationStore) FindGitHubIntegration(_ context.Context, orgID, owner, repoName string) (ghintegration.Integration, error) {
	for _, value := range f.integrations {
		if value.OrganizationID == orgID && value.Owner == owner && value.Repo == repoName {
			return value, nil
		}
	}
	return ghintegration.Integration{}, errNotFound("integration", owner+"/"+repoName)
}

func (f *fakeGitHubIntegrationStore) SaveWebhookEvent(_ context.Context, event ghintegration.WebhookEvent) error {
	for _, existing := range f.events {
		if existing.DeliveryID == event.DeliveryID {
			return nil
		}
	}
	f.events = append(f.events, event)
	return nil
}

func (f *fakeGitHubIntegrationStore) SaveScanJob(_ context.Context, job ghintegration.ScanJob) error {
	for index, existing := range f.jobs {
		if existing.ID == job.ID {
			f.jobs[index] = job
			return nil
		}
	}
	f.jobs = append(f.jobs, job)
	return nil
}

func (f *fakeGitHubIntegrationStore) ClaimNextQueuedScanJob(context.Context) (ghintegration.ScanJob, error) {
	if len(f.jobs) == 0 {
		return ghintegration.ScanJob{}, errNotFound("job", "queued")
	}
	f.jobs[0].Status = ghintegration.JobStatusRunning
	return f.jobs[0], nil
}

func (f *fakeGitHubIntegrationStore) CompleteScanJob(_ context.Context, jobID, scanID string) error {
	f.jobs[0].Status = ghintegration.JobStatusCompleted
	f.jobs[0].ScanID = scanID
	return nil
}

func (f *fakeGitHubIntegrationStore) FailScanJob(_ context.Context, jobID, message string) error {
	f.jobs[0].Status = ghintegration.JobStatusFailed
	f.jobs[0].ErrorMessage = message
	return nil
}
