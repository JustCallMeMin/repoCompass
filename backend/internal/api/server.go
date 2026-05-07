// Package api exposes the RepoCompass product HTTP API.
package api

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"log/slog"
	"net/http"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/JustCallMeMin/repoCompass/backend/internal/auth"
	"github.com/JustCallMeMin/repoCompass/backend/internal/history"
	"github.com/JustCallMeMin/repoCompass/backend/internal/insights"
	ghintegration "github.com/JustCallMeMin/repoCompass/backend/internal/integration/github"
	"github.com/JustCallMeMin/repoCompass/backend/internal/org"
	"github.com/JustCallMeMin/repoCompass/backend/internal/repository"
	"github.com/JustCallMeMin/repoCompass/backend/internal/scan"
)

// HistoryReader reads persisted scan history and findings for API responses.
type HistoryReader interface {
	ListScanHistory(ctx context.Context, repositoryID string, limit int) ([]history.ScanSummary, error)
	ListFindings(ctx context.Context, scanID string) ([]history.FindingDetail, error)
	ListMetricTrend(ctx context.Context, repositoryID string, metricKey string, limit int) ([]history.MetricPoint, error)
}

// GitHubCloner prepares public GitHub repositories as local scan paths.
type GitHubCloner interface {
	Clone(ctx context.Context, repo ghintegration.RepositoryURL) (ghintegration.Checkout, error)
}

// InsightsReader fetches organization-level aggregated statistics.
type InsightsReader interface {
	GetOrganizationInsights(ctx context.Context, orgID string) (insights.OrganizationInsights, error)
}

// Server handles RepoCompass API routes.
type Server struct {
	runner              scan.ScanRunner
	history             HistoryReader
	github              GitHubCloner
	logger              *slog.Logger
	githubWebhookSecret string
	orgs                OrgStore
	authSvc             *auth.AuthorizationService
	insights            InsightsReader
	scanLimiter         *scanRateLimiter
}

// NewServer creates an API server with required dependencies.
func NewServer(runner scan.ScanRunner, history HistoryReader, github GitHubCloner, orgs OrgStore, logger *slog.Logger) *Server {
	if logger == nil {
		logger = slog.Default()
	}
	var authSvc *auth.AuthorizationService
	if orgs != nil {
		authSvc = auth.NewAuthorizationService(orgs)
	}
	return &Server{
		runner:  runner,
		history: history,
		github:  github,
		orgs:    orgs,
		authSvc: authSvc,
		logger:  logger,
		scanLimiter: newScanRateLimiter(scanRateLimit{
			limit:  20,
			window: time.Minute,
		}),
	}
}

// SetGitHubWebhookSecret configures optional GitHub webhook HMAC validation.
func (s *Server) SetGitHubWebhookSecret(secret string) {
	s.githubWebhookSecret = secret
}

// Handler returns the HTTP handler for all API routes.
func (s *Server) Handler() http.Handler {
	mux := http.NewServeMux()
	mux.HandleFunc("GET /healthz", s.handleHealth)
	mux.HandleFunc("POST /api/v1/scans", s.handleCreateScan)
	mux.HandleFunc("GET /api/v1/repositories/{repository_id}/scans", s.handleRepositoryScans)
	mux.HandleFunc("GET /api/v1/repositories/{repository_id}/metrics", s.handleRepositoryMetrics)
	mux.HandleFunc("GET /api/v1/scans/{scan_id}/findings", s.handleScanFindings)
	mux.HandleFunc("POST /api/v1/integrations/github/webhook", s.handleGitHubWebhook)

	// Organization Management
	mux.HandleFunc("GET /api/v1/organizations", s.handleListOrganizations)
	mux.HandleFunc("GET /api/v1/organizations/{organization_id}", s.handleGetOrganization)
	mux.HandleFunc("GET /api/v1/organizations/{organization_id}/members", s.handleListMemberships)
	mux.HandleFunc("POST /api/v1/organizations/{organization_id}/members", s.handleAddMember)
	mux.HandleFunc("DELETE /api/v1/organizations/{organization_id}/members/{user_id}", s.handleRemoveMember)
	mux.HandleFunc("GET /api/v1/organizations/{organization_id}/repositories", s.handleListOrgRepositories)

	// Organization Policies
	mux.HandleFunc("GET /api/v1/policies/{policy_id}", s.handleGetPolicy)
	mux.HandleFunc("POST /api/v1/organizations/{organization_id}/policies", s.handleSavePolicy)
	mux.HandleFunc("PUT /api/v1/organizations/{organization_id}/policies/{policy_id}", s.handleSavePolicy)
	mux.HandleFunc("GET /api/v1/organizations/{organization_id}/policies", s.handleListPolicies)

	// Operational Metrics & Insights
	mux.HandleFunc("GET /api/v1/metrics", s.handleMetrics)
	mux.HandleFunc("GET /api/v1/organizations/{organization_id}/insights", s.handleOrgInsights)

	return requestIDMiddleware(mux)
}

func (s *Server) handleHealth(w http.ResponseWriter, r *http.Request) {
	writeJSON(w, http.StatusOK, map[string]string{"status": "ok"})
}

func (s *Server) handleCreateScan(w http.ResponseWriter, r *http.Request) {
	if s.runner == nil {
		writeError(w, http.StatusInternalServerError, "runner_unavailable", "scan runner is not configured")
		return
	}

	userID := actorIDFromRequest(r)
	if s.scanLimiter != nil && !s.scanLimiter.allow(userID, time.Now()) {
		writeError(w, http.StatusTooManyRequests, "rate_limited", "scan request rate limit exceeded")
		return
	}

	var request createScanRequest
	if err := decodeJSON(r, &request); err != nil {
		writeError(w, http.StatusBadRequest, "invalid_json", err.Error())
		return
	}

	source, cleanup, err := s.resolveScanSource(r.Context(), request)
	if err != nil {
		writeError(w, http.StatusBadRequest, "invalid_scan_source", err.Error())
		return
	}
	if cleanup != nil {
		defer cleanup()
	}

	orgID := r.Header.Get("X-Organization-Id")
	if orgID == "" {
		orgID = org.DefaultPersonalOrgID
	}
	source.OrganizationID = orgID

	if s.authSvc != nil {
		if err := s.authSvc.CheckEditRepo(r.Context(), userID, orgID); err != nil {
			writeError(w, http.StatusForbidden, "forbidden", err.Error())
			return
		}
	}

	result, err := s.runner.Run(r.Context(), scan.RunRequest{Source: source})
	if err != nil {
		s.logger.ErrorContext(r.Context(), "api scan failed", "request_id", requestIDFromContext(r.Context()), "error", err)
		writeError(w, http.StatusInternalServerError, "scan_failed", err.Error())
		return
	}

	writeJSON(w, http.StatusAccepted, scanResponseFromResult(result))
}

func (s *Server) resolveScanSource(ctx context.Context, request createScanRequest) (repository.RepositorySource, func(), error) {
	switch request.SourceType {
	case "local":
		if strings.TrimSpace(request.Path) == "" {
			return repository.RepositorySource{}, nil, fmt.Errorf("path is required for local scans")
		}
		return repository.RepositorySource{
			Type: repository.SourceTypeLocal,
			Path: request.Path,
		}, nil, nil
	case "github":
		if s.github == nil {
			return repository.RepositorySource{}, nil, fmt.Errorf("github integration is not configured")
		}
		repo, err := ghintegration.ParseRepositoryURL(request.URL)
		if err != nil {
			return repository.RepositorySource{}, nil, err
		}
		checkout, err := s.github.Clone(ctx, repo)
		if err != nil {
			return repository.RepositorySource{}, nil, err
		}
		return repository.RepositorySource{
			Type: repository.SourceTypeLocal,
			Path: checkout.Path,
		}, checkout.Cleanup, nil
	default:
		return repository.RepositorySource{}, nil, fmt.Errorf("source_type must be local or github")
	}
}

func (s *Server) handleRepositoryScans(w http.ResponseWriter, r *http.Request) {
	if s.history == nil {
		writeError(w, http.StatusInternalServerError, "history_unavailable", "history store is not configured")
		return
	}
	repositoryID := r.PathValue("repository_id")

	orgID := r.Header.Get("X-Organization-Id")
	if orgID == "" {
		orgID = org.DefaultPersonalOrgID
	}
	if s.authSvc != nil {
		userID := actorIDFromRequest(r)
		if err := s.authSvc.CheckAccessRepo(r.Context(), userID, orgID); err != nil {
			writeError(w, http.StatusForbidden, "forbidden", err.Error())
			return
		}
	}
	// T6-015: verify the repository belongs to the claimed org (prevent cross-org data leak).
	if s.orgs != nil {
		repo, err := s.orgs.GetRepository(r.Context(), repositoryID)
		if err != nil {
			writeError(w, http.StatusNotFound, "not_found", "repository not found")
			return
		}
		if repo.OrganizationID != orgID {
			writeError(w, http.StatusForbidden, "forbidden", "repository does not belong to the requested organization")
			return
		}
	}

	limit := parseLimit(r, 20)
	items, err := s.history.ListScanHistory(r.Context(), repositoryID, limit)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "history_query_failed", err.Error())
		return
	}
	if items == nil {
		items = []history.ScanSummary{}
	}
	writeJSON(w, http.StatusOK, items)
}

func (s *Server) handleScanFindings(w http.ResponseWriter, r *http.Request) {
	if s.history == nil {
		writeError(w, http.StatusInternalServerError, "history_unavailable", "history store is not configured")
		return
	}
	items, err := s.history.ListFindings(r.Context(), r.PathValue("scan_id"))
	if err != nil {
		writeError(w, http.StatusInternalServerError, "findings_query_failed", err.Error())
		return
	}
	if items == nil {
		items = []history.FindingDetail{}
	}
	writeJSON(w, http.StatusOK, items)
}

func (s *Server) handleRepositoryMetrics(w http.ResponseWriter, r *http.Request) {
	if s.history == nil {
		writeError(w, http.StatusInternalServerError, "history_unavailable", "history store is not configured")
		return
	}
	metricKey := r.URL.Query().Get("metric_key")
	if metricKey == "" {
		metricKey = "assessment.overall_score"
	}
	repositoryID := r.PathValue("repository_id")

	orgID := r.Header.Get("X-Organization-Id")
	if orgID == "" {
		orgID = org.DefaultPersonalOrgID
	}
	if s.authSvc != nil {
		userID := actorIDFromRequest(r)
		if err := s.authSvc.CheckAccessRepo(r.Context(), userID, orgID); err != nil {
			writeError(w, http.StatusForbidden, "forbidden", err.Error())
			return
		}
	}
	// T6-015: verify the repository belongs to the claimed org.
	if s.orgs != nil {
		repo, err := s.orgs.GetRepository(r.Context(), repositoryID)
		if err != nil {
			writeError(w, http.StatusNotFound, "not_found", "repository not found")
			return
		}
		if repo.OrganizationID != orgID {
			writeError(w, http.StatusForbidden, "forbidden", "repository does not belong to the requested organization")
			return
		}
	}

	limit := parseLimit(r, 20)
	items, err := s.history.ListMetricTrend(r.Context(), repositoryID, metricKey, limit)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "metrics_query_failed", err.Error())
		return
	}
	if items == nil {
		items = []history.MetricPoint{}
	}
	writeJSON(w, http.StatusOK, items)
}

func (s *Server) handleGitHubWebhook(w http.ResponseWriter, r *http.Request) {
	event, payload, err := ghintegration.ReadWebhook(r, s.githubWebhookSecret)
	if err != nil {
		status := http.StatusBadRequest
		if errors.Is(err, ghintegration.ErrInvalidSignature) {
			status = http.StatusUnauthorized
		}
		writeError(w, status, "github_webhook_invalid", err.Error())
		return
	}

	switch event {
	case "ping":
		writeJSON(w, http.StatusOK, map[string]string{"status": "ok", "event": event})
	case "push":
		writeJSON(w, http.StatusAccepted, map[string]string{
			"status":     "accepted",
			"event":      event,
			"repository": payload.Repository.FullName,
		})
	default:
		writeJSON(w, http.StatusAccepted, map[string]string{"status": "ignored", "event": event})
	}
}

type createScanRequest struct {
	SourceType string `json:"source_type"`
	Path       string `json:"path,omitempty"`
	URL        string `json:"url,omitempty"`
}

type scanResponse struct {
	ScanID             string `json:"scan_id"`
	RepositoryID       string `json:"repository_id"`
	SnapshotID         string `json:"snapshot_id"`
	Status             string `json:"status"`
	AnalyzersProcessed int    `json:"analyzers_processed"`
	FindingCount       int    `json:"finding_count"`
	AssessmentScore    int    `json:"assessment_score"`
}

func scanResponseFromResult(result scan.RunResult) scanResponse {
	return scanResponse{
		ScanID:             result.Scan.ID,
		RepositoryID:       result.Repository.ID,
		SnapshotID:         result.Snapshot.ID,
		Status:             string(result.Scan.Status),
		AnalyzersProcessed: result.Summary.AnalyzersProcessed,
		FindingCount:       result.Summary.FindingCount,
		AssessmentScore:    result.Summary.AssessmentScore,
	}
}

func decodeJSON(r *http.Request, target any) error {
	defer r.Body.Close()
	decoder := json.NewDecoder(r.Body)
	decoder.DisallowUnknownFields()
	if err := decoder.Decode(target); err != nil {
		return err
	}
	return nil
}

func writeJSON(w http.ResponseWriter, status int, value any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(value)
}

func writeError(w http.ResponseWriter, status int, code, message string) {
	writeJSON(w, status, map[string]any{
		"error": map[string]string{
			"code":    code,
			"message": message,
		},
	})
}

func parseLimit(r *http.Request, fallback int) int {
	raw := r.URL.Query().Get("limit")
	if raw == "" {
		return fallback
	}
	limit, err := strconv.Atoi(raw)
	if err != nil || limit <= 0 {
		return fallback
	}
	return limit
}

func requestIDMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		requestID := r.Header.Get("X-Request-ID")
		if requestID == "" {
			requestID = newRequestID()
		}
		w.Header().Set("X-Request-ID", requestID)
		w.Header().Set("Access-Control-Allow-Origin", "*")
		w.Header().Set("Access-Control-Allow-Headers", "Content-Type, X-Request-ID, X-Organization-Id, X-User-Id")
		w.Header().Set("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, OPTIONS")
		if r.Method == http.MethodOptions {
			w.WriteHeader(http.StatusNoContent)
			return
		}
		ctx := context.WithValue(r.Context(), correlationKey{}, correlationAttrs{
			requestID: requestID,
			method:    r.Method,
			path:      r.URL.Path,
		})
		slog.InfoContext(ctx, "request",
			slog.String("request_id", requestID),
			slog.String("method", r.Method),
			slog.String("path", r.URL.Path),
		)
		next.ServeHTTP(w, r.WithContext(ctx))
	})
}

const defaultActorID = "mock_user"

func actorIDFromRequest(r *http.Request) string {
	userID := strings.TrimSpace(r.Header.Get("X-User-Id"))
	if userID == "" {
		return defaultActorID
	}
	return userID
}

// correlationKey is the context key for request correlation attributes.
type correlationKey struct{}

// correlationAttrs holds per-request structured log fields propagated via context.
type correlationAttrs struct {
	requestID string
	method    string
	path      string
}

func requestIDFromContext(ctx context.Context) string {
	attrs, ok := ctx.Value(correlationKey{}).(correlationAttrs)
	if !ok {
		return ""
	}
	return attrs.requestID
}

func newRequestID() string {
	var b [8]byte
	if _, err := rand.Read(b[:]); err != nil {
		return fmt.Sprintf("req_%d", time.Now().UnixNano())
	}
	return "req_" + hex.EncodeToString(b[:])
}

type scanRateLimit struct {
	limit  int
	window time.Duration
}

type scanRateLimiter struct {
	mu     sync.Mutex
	config scanRateLimit
	hits   map[string][]time.Time
}

func newScanRateLimiter(config scanRateLimit) *scanRateLimiter {
	if config.limit <= 0 {
		config.limit = 20
	}
	if config.window <= 0 {
		config.window = time.Minute
	}
	return &scanRateLimiter{
		config: config,
		hits:   make(map[string][]time.Time),
	}
}

func (l *scanRateLimiter) allow(key string, now time.Time) bool {
	l.mu.Lock()
	defer l.mu.Unlock()

	cutoff := now.Add(-l.config.window)
	recent := l.hits[key][:0]
	for _, hit := range l.hits[key] {
		if hit.After(cutoff) {
			recent = append(recent, hit)
		}
	}
	if len(recent) >= l.config.limit {
		l.hits[key] = recent
		return false
	}
	recent = append(recent, now)
	l.hits[key] = recent
	return true
}
