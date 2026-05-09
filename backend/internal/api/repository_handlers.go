package api

import (
	"context"
	"crypto/rand"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"os"
	"strings"
	"time"

	"github.com/JustCallMeMin/repoCompass/backend/internal/auth"
	ghintegration "github.com/JustCallMeMin/repoCompass/backend/internal/integration/github"
	"github.com/JustCallMeMin/repoCompass/backend/internal/org"
	"github.com/JustCallMeMin/repoCompass/backend/internal/repository"
	"github.com/JustCallMeMin/repoCompass/backend/internal/scan"
)

// handleListRepositories lists repositories visible in the requested organization.
func (s *Server) handleListRepositories(w http.ResponseWriter, r *http.Request) {
	if s.orgs == nil {
		writeRequestError(w, r, http.StatusInternalServerError, "repository_store_unavailable", "repository store is not configured")
		return
	}
	orgID := organizationIDFromRequest(r)
	if !s.authorize(w, r, orgID, "read") {
		return
	}
	repos, err := s.orgs.ListRepositoriesByOrg(r.Context(), orgID)
	if err != nil {
		writeRequestError(w, r, http.StatusInternalServerError, "repository_query_failed", "failed to list repositories")
		return
	}
	if repos == nil {
		repos = []repository.Repository{}
	}
	writeData(w, r, http.StatusOK, repos)
}

// handleGetRepository returns one repository after org-scope verification.
func (s *Server) handleGetRepository(w http.ResponseWriter, r *http.Request) {
	repo, ok := s.repositoryFromPath(w, r)
	if !ok {
		return
	}
	writeData(w, r, http.StatusOK, repo)
}

// handleCreateRepositoryScan triggers a persisted scan for an existing repository.
func (s *Server) handleCreateRepositoryScan(w http.ResponseWriter, r *http.Request) {
	if s.runner == nil {
		writeRequestError(w, r, http.StatusInternalServerError, "runner_unavailable", "scan runner is not configured")
		return
	}
	repo, ok := s.repositoryFromPath(w, r)
	if !ok {
		return
	}
	source, err := sourceFromRepository(repo)
	if err != nil {
		writeRequestError(w, r, http.StatusBadRequest, "invalid_scan_source", err.Error())
		return
	}
	var cleanup func()
	if strings.Contains(string(repo.Provider), "github") && repo.URL != "" {
		if s.github == nil {
			writeRequestError(w, r, http.StatusInternalServerError, "github_unavailable", "github integration is not configured")
			return
		}
		parsed, err := ghintegration.ParseRepositoryURL(repo.URL)
		if err != nil {
			writeRequestError(w, r, http.StatusBadRequest, "invalid_scan_source", err.Error())
			return
		}
		checkout, err := s.github.Clone(r.Context(), parsed)
		if err != nil {
			writeRequestError(w, r, http.StatusBadRequest, "invalid_scan_source", err.Error())
			return
		}
		source.Path = checkout.Path
		cleanup = checkout.Cleanup
	}
	if cleanup != nil {
		defer cleanup()
	}
	ctx, cancel := context.WithTimeout(r.Context(), 2*time.Minute)
	defer cancel()
	result, err := s.runner.Run(ctx, scan.RunRequest{Source: source})
	if err != nil {
		writeRequestError(w, r, http.StatusInternalServerError, "scan_failed", err.Error())
		return
	}
	writeData(w, r, http.StatusAccepted, scanResponseFromResult(result))
}

// handleGetScan returns one persisted scan row.
func (s *Server) handleGetScan(w http.ResponseWriter, r *http.Request) {
	if s.history == nil {
		writeRequestError(w, r, http.StatusInternalServerError, "history_unavailable", "history store is not configured")
		return
	}
	scanID := strings.TrimSpace(r.PathValue("scan_id"))
	if scanID == "" {
		writeRequestError(w, r, http.StatusBadRequest, "invalid_request", "scan_id path variable missing")
		return
	}
	value, err := s.history.GetScan(r.Context(), scanID)
	if err != nil {
		writeRequestError(w, r, http.StatusNotFound, "not_found", "scan not found")
		return
	}
	writeData(w, r, http.StatusOK, value)
}

// handleGetAssessment returns one persisted assessment.
func (s *Server) handleGetAssessment(w http.ResponseWriter, r *http.Request) {
	if s.history == nil {
		writeRequestError(w, r, http.StatusInternalServerError, "history_unavailable", "history store is not configured")
		return
	}
	value, err := s.history.GetAssessment(r.Context(), r.PathValue("scan_id"))
	if err != nil {
		writeRequestError(w, r, http.StatusNotFound, "not_found", "assessment not found")
		return
	}
	writeData(w, r, http.StatusOK, value)
}

// handleListReports returns persisted report metadata.
func (s *Server) handleListReports(w http.ResponseWriter, r *http.Request) {
	if s.history == nil {
		writeRequestError(w, r, http.StatusInternalServerError, "history_unavailable", "history store is not configured")
		return
	}
	reports, err := s.history.ListReports(r.Context(), r.PathValue("scan_id"))
	if err != nil {
		writeRequestError(w, r, http.StatusInternalServerError, "reports_query_failed", "failed to list reports")
		return
	}
	if reports == nil {
		reports = []map[string]any{}
	}
	writeData(w, r, http.StatusOK, reports)
}

// handleGitHubLogin starts a GitHub OAuth browser flow.
func (s *Server) handleGitHubLogin(w http.ResponseWriter, r *http.Request) {
	if s.sessions == nil {
		writeRequestError(w, r, http.StatusInternalServerError, "session_store_unavailable", "session store is not configured")
		return
	}
	clientID := os.Getenv("GITHUB_OAUTH_CLIENT_ID")
	redirectURL := os.Getenv("GITHUB_OAUTH_REDIRECT_URL")
	if clientID == "" || redirectURL == "" {
		writeRequestError(w, r, http.StatusServiceUnavailable, "oauth_unconfigured", "GitHub OAuth is not configured")
		return
	}
	state := newOpaqueToken()
	now := time.Now().UTC()
	if err := s.sessions.SaveOAuthState(r.Context(), auth.OAuthState{
		State:     state,
		Provider:  "github",
		ExpiresAt: now.Add(10 * time.Minute),
		CreatedAt: now,
	}); err != nil {
		writeRequestError(w, r, http.StatusInternalServerError, "oauth_state_write_failed", "failed to start GitHub OAuth")
		return
	}
	target := "https://github.com/login/oauth/authorize?client_id=" + url.QueryEscape(clientID) +
		"&redirect_uri=" + url.QueryEscape(redirectURL) +
		"&scope=read:user%20user:email" +
		"&state=" + url.QueryEscape(state)
	if r.URL.Query().Get("format") != "json" {
		http.Redirect(w, r, target, http.StatusFound)
		return
	}
	writeData(w, r, http.StatusOK, map[string]string{"authorization_url": target, "state": state})
}

// handleGitHubCallback completes a GitHub OAuth flow for local/dev use.
func (s *Server) handleGitHubCallback(w http.ResponseWriter, r *http.Request) {
	if s.sessions == nil {
		writeRequestError(w, r, http.StatusInternalServerError, "session_store_unavailable", "session store is not configured")
		return
	}
	clientID := os.Getenv("GITHUB_OAUTH_CLIENT_ID")
	clientSecret := os.Getenv("GITHUB_OAUTH_CLIENT_SECRET")
	redirectURL := os.Getenv("GITHUB_OAUTH_REDIRECT_URL")
	if clientID == "" || clientSecret == "" || redirectURL == "" {
		writeRequestError(w, r, http.StatusServiceUnavailable, "oauth_unconfigured", "GitHub OAuth is not configured")
		return
	}
	code := strings.TrimSpace(r.URL.Query().Get("code"))
	if code == "" {
		writeRequestError(w, r, http.StatusBadRequest, "invalid_request", "code query parameter is required")
		return
	}
	state := strings.TrimSpace(r.URL.Query().Get("state"))
	if state == "" {
		writeRequestError(w, r, http.StatusBadRequest, "invalid_request", "state query parameter is required")
		return
	}
	if _, err := s.sessions.ConsumeOAuthState(r.Context(), "github", state, time.Now().UTC()); err != nil {
		writeRequestError(w, r, http.StatusUnauthorized, "oauth_state_invalid", "OAuth state is invalid or expired")
		return
	}
	githubUser, err := fetchGitHubOAuthUser(r.Context(), clientID, clientSecret, redirectURL, code)
	if err != nil {
		writeRequestError(w, r, http.StatusBadGateway, "oauth_exchange_failed", "failed to authenticate with GitHub")
		return
	}
	now := time.Now().UTC()
	user := auth.User{
		ID:        "usr_github_" + githubUser.ID,
		GitHubID:  githubUser.ID,
		Login:     githubUser.Login,
		Name:      githubUser.Name,
		AvatarURL: githubUser.AvatarURL,
		Email:     githubUser.Email,
		CreatedAt: now,
		UpdatedAt: now,
	}
	if err := s.sessions.SaveUser(r.Context(), user); err != nil {
		writeRequestError(w, r, http.StatusInternalServerError, "session_write_failed", "failed to save user")
		return
	}
	if s.orgs != nil {
		_ = s.orgs.SaveMembership(r.Context(), org.Membership{
			OrganizationID: org.DefaultPersonalOrgID,
			UserID:         user.ID,
			Role:           org.RoleOwner,
			CreatedAt:      now,
			UpdatedAt:      now,
		})
	}
	token := newOpaqueToken()
	session := auth.Session{
		ID:        "sess_" + stableHash(token),
		UserID:    user.ID,
		TokenHash: hashToken(token),
		ExpiresAt: now.Add(24 * time.Hour),
		CreatedAt: now,
	}
	if err := s.sessions.SaveSession(r.Context(), session); err != nil {
		writeRequestError(w, r, http.StatusInternalServerError, "session_write_failed", "failed to save session")
		return
	}
	http.SetCookie(w, &http.Cookie{
		Name:     "repocompass_session",
		Value:    token,
		Path:     "/",
		HttpOnly: true,
		SameSite: http.SameSiteLaxMode,
		Secure:   os.Getenv("APP_ENV") == "production",
		Expires:  session.ExpiresAt,
	})
	writeData(w, r, http.StatusOK, map[string]any{"user": user, "session": session, "token": token})
}

type githubOAuthUser struct {
	ID        string
	Login     string
	Name      string
	AvatarURL string
	Email     string
}

// fetchGitHubOAuthUser exchanges an OAuth code and fetches the real GitHub user.
func fetchGitHubOAuthUser(ctx context.Context, clientID, clientSecret, redirectURL, code string) (githubOAuthUser, error) {
	form := url.Values{}
	form.Set("client_id", clientID)
	form.Set("client_secret", clientSecret)
	form.Set("redirect_uri", redirectURL)
	form.Set("code", code)
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, "https://github.com/login/oauth/access_token", strings.NewReader(form.Encode()))
	if err != nil {
		return githubOAuthUser{}, err
	}
	req.Header.Set("Accept", "application/json")
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return githubOAuthUser{}, err
	}
	defer resp.Body.Close()
	body, err := io.ReadAll(io.LimitReader(resp.Body, 1<<20))
	if err != nil {
		return githubOAuthUser{}, err
	}
	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return githubOAuthUser{}, fmt.Errorf("github token exchange failed: %s", resp.Status)
	}
	var tokenResp struct {
		AccessToken string `json:"access_token"`
		Error       string `json:"error"`
		Description string `json:"error_description"`
	}
	if err := json.Unmarshal(body, &tokenResp); err != nil {
		return githubOAuthUser{}, err
	}
	if tokenResp.AccessToken == "" {
		if tokenResp.Error != "" {
			return githubOAuthUser{}, fmt.Errorf("github oauth error: %s", tokenResp.Error)
		}
		return githubOAuthUser{}, fmt.Errorf("github oauth response missing access_token")
	}
	userReq, err := http.NewRequestWithContext(ctx, http.MethodGet, "https://api.github.com/user", nil)
	if err != nil {
		return githubOAuthUser{}, err
	}
	userReq.Header.Set("Accept", "application/vnd.github+json")
	userReq.Header.Set("Authorization", "Bearer "+tokenResp.AccessToken)
	userResp, err := http.DefaultClient.Do(userReq)
	if err != nil {
		return githubOAuthUser{}, err
	}
	defer userResp.Body.Close()
	userBody, err := io.ReadAll(io.LimitReader(userResp.Body, 1<<20))
	if err != nil {
		return githubOAuthUser{}, err
	}
	if userResp.StatusCode < 200 || userResp.StatusCode >= 300 {
		return githubOAuthUser{}, fmt.Errorf("github user fetch failed: %s", userResp.Status)
	}
	var raw struct {
		ID        int64  `json:"id"`
		Login     string `json:"login"`
		Name      string `json:"name"`
		AvatarURL string `json:"avatar_url"`
		Email     string `json:"email"`
	}
	if err := json.Unmarshal(userBody, &raw); err != nil {
		return githubOAuthUser{}, err
	}
	if raw.ID == 0 || raw.Login == "" {
		return githubOAuthUser{}, fmt.Errorf("github user response missing id/login")
	}
	return githubOAuthUser{
		ID:        fmt.Sprintf("%d", raw.ID),
		Login:     raw.Login,
		Name:      raw.Name,
		AvatarURL: raw.AvatarURL,
		Email:     raw.Email,
	}, nil
}

// handleCurrentSession returns the authenticated session actor.
func (s *Server) handleCurrentSession(w http.ResponseWriter, r *http.Request) {
	userID := s.authenticatedActorID(r)
	if userID == "" {
		writeRequestError(w, r, http.StatusUnauthorized, "unauthorized", "session is required")
		return
	}
	writeData(w, r, http.StatusOK, map[string]string{"user_id": userID, "organization_id": organizationIDFromRequest(r)})
}

// handleLogout revokes the current session when backed by persistent sessions.
func (s *Server) handleLogout(w http.ResponseWriter, r *http.Request) {
	if s.sessions == nil {
		writeData(w, r, http.StatusOK, map[string]string{"status": "ok"})
		return
	}
	token := bearerOrCookieToken(r)
	if token == "" {
		writeRequestError(w, r, http.StatusUnauthorized, "unauthorized", "session token is required")
		return
	}
	session, err := s.sessions.GetSessionByTokenHash(r.Context(), hashToken(token))
	if err != nil {
		writeRequestError(w, r, http.StatusUnauthorized, "unauthorized", "session not found")
		return
	}
	if err := s.sessions.RevokeSession(r.Context(), session.ID); err != nil {
		writeRequestError(w, r, http.StatusInternalServerError, "session_revoke_failed", "failed to revoke session")
		return
	}
	writeData(w, r, http.StatusOK, map[string]string{"status": "revoked"})
}

func (s *Server) repositoryFromPath(w http.ResponseWriter, r *http.Request) (repository.Repository, bool) {
	if s.orgs == nil {
		writeRequestError(w, r, http.StatusInternalServerError, "repository_store_unavailable", "repository store is not configured")
		return repository.Repository{}, false
	}
	repositoryID := strings.TrimSpace(r.PathValue("repository_id"))
	if repositoryID == "" {
		writeRequestError(w, r, http.StatusBadRequest, "invalid_request", "repository_id path variable missing")
		return repository.Repository{}, false
	}
	repo, err := s.orgs.GetRepository(r.Context(), repositoryID)
	if err != nil {
		writeRequestError(w, r, http.StatusNotFound, "not_found", "repository not found")
		return repository.Repository{}, false
	}
	if !s.authorize(w, r, repo.OrganizationID, "read") {
		return repository.Repository{}, false
	}
	return repo, true
}

func (s *Server) authorize(w http.ResponseWriter, r *http.Request, orgID, action string) bool {
	if s.authSvc == nil {
		return true
	}
	userID := s.authenticatedActorID(r)
	if userID == "" {
		writeRequestError(w, r, http.StatusUnauthorized, "unauthorized", "session is required")
		return false
	}
	var err error
	if action == "manage" {
		err = s.authSvc.CheckManageOrg(r.Context(), userID, orgID)
	} else if action == "write" {
		err = s.authSvc.CheckEditRepo(r.Context(), userID, orgID)
	} else {
		err = s.authSvc.CheckAccessRepo(r.Context(), userID, orgID)
	}
	if err != nil {
		writeRequestError(w, r, http.StatusForbidden, "forbidden", err.Error())
		return false
	}
	return true
}

func (s *Server) authenticatedActorID(r *http.Request) string {
	if s.devHeaderAuth {
		return actorIDFromRequest(r)
	}
	token := bearerOrCookieToken(r)
	if token == "" || s.sessions == nil {
		return ""
	}
	session, err := s.sessions.GetSessionByTokenHash(r.Context(), hashToken(token))
	if err != nil {
		return ""
	}
	return session.UserID
}

func sourceFromRepository(repo repository.Repository) (repository.RepositorySource, error) {
	if repo.LocalPath != "" {
		return repository.RepositorySource{Type: repository.SourceTypeLocal, Path: repo.LocalPath, OrganizationID: repo.OrganizationID}, nil
	}
	if repo.URL != "" && strings.Contains(string(repo.Provider), "github") {
		return repository.RepositorySource{Type: repository.SourceTypeLocal, URL: repo.URL, Path: repo.URL, OrganizationID: repo.OrganizationID}, nil
	}
	return repository.RepositorySource{}, fmt.Errorf("repository has no scan source")
}

func organizationIDFromRequest(r *http.Request) string {
	orgID := strings.TrimSpace(r.Header.Get("X-Organization-Id"))
	if orgID == "" {
		return org.DefaultPersonalOrgID
	}
	return orgID
}

func bearerOrCookieToken(r *http.Request) string {
	authz := strings.TrimSpace(r.Header.Get("Authorization"))
	if strings.HasPrefix(strings.ToLower(authz), "bearer ") {
		return strings.TrimSpace(authz[7:])
	}
	cookie, err := r.Cookie("repocompass_session")
	if err == nil {
		return cookie.Value
	}
	return ""
}

func hashToken(token string) string {
	sum := sha256.Sum256([]byte(token))
	return hex.EncodeToString(sum[:])
}

func stableHash(value string) string {
	sum := sha256.Sum256([]byte(value))
	return hex.EncodeToString(sum[:8])
}

func newOpaqueToken() string {
	var b [24]byte
	if _, err := rand.Read(b[:]); err != nil {
		return fmt.Sprintf("tok_%d", time.Now().UnixNano())
	}
	return hex.EncodeToString(b[:])
}
