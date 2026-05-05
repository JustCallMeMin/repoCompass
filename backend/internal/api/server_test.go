package api

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/JustCallMeMin/repoCompass/backend/internal/assessment"
	"github.com/JustCallMeMin/repoCompass/backend/internal/history"
	ghintegration "github.com/JustCallMeMin/repoCompass/backend/internal/integration/github"
	"github.com/JustCallMeMin/repoCompass/backend/internal/repository"
	"github.com/JustCallMeMin/repoCompass/backend/internal/scan"
	"github.com/JustCallMeMin/repoCompass/backend/internal/snapshot"
)

func TestHealthEndpoint(t *testing.T) {
	handler := NewServer(nil, nil, nil, nil).Handler()
	response := httptest.NewRecorder()
	request := httptest.NewRequest(http.MethodGet, "/healthz", nil)

	handler.ServeHTTP(response, request)

	if response.Code != http.StatusOK {
		t.Fatalf("unexpected status: got %d", response.Code)
	}
	assertJSONField(t, response.Body.Bytes(), "status", "ok")
	if response.Header().Get("X-Request-ID") == "" {
		t.Fatal("expected request ID header")
	}
}

func TestCORSPreflight(t *testing.T) {
	handler := NewServer(nil, nil, nil, nil).Handler()
	response := httptest.NewRecorder()
	request := httptest.NewRequest(http.MethodOptions, "/api/v1/scans", nil)

	handler.ServeHTTP(response, request)

	if response.Code != http.StatusNoContent {
		t.Fatalf("unexpected status: got %d", response.Code)
	}
	if response.Header().Get("Access-Control-Allow-Origin") != "*" {
		t.Fatalf("expected CORS origin header")
	}
	if response.Header().Get("Access-Control-Allow-Methods") == "" {
		t.Fatalf("expected CORS methods header")
	}
}

func TestCreateLocalScan(t *testing.T) {
	runner := fakeRunner{
		result: scan.RunResult{
			Scan:       scan.Scan{ID: "scan_123", Status: scan.StatusCompleted},
			Repository: repository.Repository{ID: "repo_123"},
			Snapshot:   snapshot.RepositorySnapshot{ID: "snap_123"},
			Summary: scan.Summary{
				AnalyzersProcessed: 4,
				FindingCount:       2,
				AssessmentScore:    82,
			},
			Assessment: assessment.Assessment{OverallScore: 82},
		},
	}
	handler := NewServer(&runner, nil, nil, nil).Handler()
	body := bytes.NewBufferString(`{"source_type":"local","path":"./repo"}`)
	request := httptest.NewRequest(http.MethodPost, "/api/v1/scans", body)
	response := httptest.NewRecorder()

	handler.ServeHTTP(response, request)

	if response.Code != http.StatusAccepted {
		t.Fatalf("unexpected status: got %d body=%s", response.Code, response.Body.String())
	}
	if runner.request.Source.Type != repository.SourceTypeLocal {
		t.Fatalf("unexpected source type: got %q", runner.request.Source.Type)
	}
	if runner.request.Source.Path != "./repo" {
		t.Fatalf("unexpected path: got %q", runner.request.Source.Path)
	}
	assertJSONField(t, response.Body.Bytes(), "scan_id", "scan_123")
	assertJSONField(t, response.Body.Bytes(), "repository_id", "repo_123")
}

func TestCreateScanRejectsInvalidLocalRequest(t *testing.T) {
	handler := NewServer(&fakeRunner{}, nil, nil, nil).Handler()
	request := httptest.NewRequest(http.MethodPost, "/api/v1/scans", bytes.NewBufferString(`{"source_type":"local"}`))
	response := httptest.NewRecorder()

	handler.ServeHTTP(response, request)

	if response.Code != http.StatusBadRequest {
		t.Fatalf("unexpected status: got %d", response.Code)
	}
	assertNestedJSONField(t, response.Body.Bytes(), "error", "code", "invalid_scan_source")
}

func TestCreateGitHubScanUsesCheckoutPath(t *testing.T) {
	runner := fakeRunner{result: scan.RunResult{
		Scan:       scan.Scan{ID: "scan_123", Status: scan.StatusCompleted},
		Repository: repository.Repository{ID: "repo_123"},
		Snapshot:   snapshot.RepositorySnapshot{ID: "snap_123"},
	}}
	cloner := fakeGitHubCloner{checkout: ghintegration.Checkout{Path: "/tmp/repo", Cleanup: func() {}}}
	handler := NewServer(&runner, nil, &cloner, nil).Handler()
	body := bytes.NewBufferString(`{"source_type":"github","url":"https://github.com/owner/repo"}`)
	request := httptest.NewRequest(http.MethodPost, "/api/v1/scans", body)
	response := httptest.NewRecorder()

	handler.ServeHTTP(response, request)

	if response.Code != http.StatusAccepted {
		t.Fatalf("unexpected status: got %d body=%s", response.Code, response.Body.String())
	}
	if cloner.repo.Owner != "owner" || cloner.repo.Name != "repo" {
		t.Fatalf("unexpected github repo: %#v", cloner.repo)
	}
	if runner.request.Source.Path != "/tmp/repo" {
		t.Fatalf("unexpected checkout path: got %q", runner.request.Source.Path)
	}
}

func TestHistoryEndpoints(t *testing.T) {
	store := fakeHistoryReader{
		scans:    []history.ScanSummary{{ScanID: "scan_123", RepositoryID: "repo_123"}},
		findings: []history.FindingDetail{{ID: "finding_123", ScanID: "scan_123"}},
		metrics:  []history.MetricPoint{{ScanID: "scan_123", RepositoryID: "repo_123", MetricKey: "onboarding_score", Value: 82}},
	}
	handler := NewServer(nil, &store, nil, nil).Handler()

	for _, tc := range []struct {
		name string
		path string
		want string
	}{
		{name: "history", path: "/api/v1/repositories/repo_123/scans", want: "scan_123"},
		{name: "findings", path: "/api/v1/scans/scan_123/findings", want: "finding_123"},
		{name: "metrics", path: "/api/v1/repositories/repo_123/metrics", want: "onboarding_score"},
	} {
		t.Run(tc.name, func(t *testing.T) {
			request := httptest.NewRequest(http.MethodGet, tc.path, nil)
			response := httptest.NewRecorder()

			handler.ServeHTTP(response, request)

			if response.Code != http.StatusOK {
				t.Fatalf("unexpected status: got %d", response.Code)
			}
			if !bytes.Contains(response.Body.Bytes(), []byte(tc.want)) {
				t.Fatalf("expected response to contain %q, got %s", tc.want, response.Body.String())
			}
		})
	}
}

func assertJSONField(t *testing.T, body []byte, key, want string) {
	t.Helper()
	var payload map[string]any
	if err := json.Unmarshal(body, &payload); err != nil {
		t.Fatalf("decode response: %v", err)
	}
	got, _ := payload[key].(string)
	if got != want {
		t.Fatalf("unexpected %s: got %q want %q", key, got, want)
	}
}

func assertNestedJSONField(t *testing.T, body []byte, parent, key, want string) {
	t.Helper()
	var payload map[string]map[string]string
	if err := json.Unmarshal(body, &payload); err != nil {
		t.Fatalf("decode response: %v", err)
	}
	if got := payload[parent][key]; got != want {
		t.Fatalf("unexpected %s.%s: got %q want %q", parent, key, got, want)
	}
}

type fakeRunner struct {
	request scan.RunRequest
	result  scan.RunResult
	err     error
}

func (r *fakeRunner) Run(_ context.Context, request scan.RunRequest) (scan.RunResult, error) {
	r.request = request
	return r.result, r.err
}

type fakeGitHubCloner struct {
	repo     ghintegration.RepositoryURL
	checkout ghintegration.Checkout
	err      error
}

func (c *fakeGitHubCloner) Clone(_ context.Context, repo ghintegration.RepositoryURL) (ghintegration.Checkout, error) {
	c.repo = repo
	return c.checkout, c.err
}

type fakeHistoryReader struct {
	scans    []history.ScanSummary
	findings []history.FindingDetail
	metrics  []history.MetricPoint
}

func (r *fakeHistoryReader) ListScanHistory(context.Context, string, int) ([]history.ScanSummary, error) {
	return r.scans, nil
}

func (r *fakeHistoryReader) ListFindings(context.Context, string) ([]history.FindingDetail, error) {
	return r.findings, nil
}

func (r *fakeHistoryReader) ListMetricTrend(context.Context, string, string, int) ([]history.MetricPoint, error) {
	return r.metrics, nil
}
