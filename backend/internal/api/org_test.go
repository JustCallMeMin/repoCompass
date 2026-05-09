package api

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"net/http/httptest"
	"os"
	"testing"
	"time"

	"github.com/JustCallMeMin/repoCompass/backend/internal/auth"
	"github.com/JustCallMeMin/repoCompass/backend/internal/insights"
	"github.com/JustCallMeMin/repoCompass/backend/internal/org"
	"github.com/JustCallMeMin/repoCompass/backend/internal/repository"
	"github.com/JustCallMeMin/repoCompass/backend/internal/scan"
)

// fakeOrgStore is an in-memory implementation of OrgStore for unit tests.
type fakeOrgStore struct {
	orgs        map[string]org.Organization
	memberships []org.Membership
	policies    []org.Policy
	repos       []repository.Repository
	users       map[string]auth.User
	sessions    map[string]auth.Session
	oauthStates map[string]auth.OAuthState
}

func newDevHeaderTestServer(runner scan.ScanRunner, history HistoryReader, github GitHubCloner, orgs OrgStore) *Server {
	_ = os.Setenv("REPOCOMPASS_ALLOW_MOCK_USER", "true")
	server := NewServer(runner, history, github, orgs, nil)
	server.SetDevHeaderAuth(true)
	return server
}

func newFakeOrgStore() *fakeOrgStore {
	return &fakeOrgStore{
		orgs:        make(map[string]org.Organization),
		users:       make(map[string]auth.User),
		sessions:    make(map[string]auth.Session),
		oauthStates: make(map[string]auth.OAuthState),
	}
}

func (f *fakeOrgStore) SaveUser(_ context.Context, user auth.User) error {
	f.users[user.ID] = user
	return nil
}

func (f *fakeOrgStore) GetUser(_ context.Context, id string) (auth.User, error) {
	user, ok := f.users[id]
	if !ok {
		return auth.User{}, errNotFound("user", id)
	}
	return user, nil
}

func (f *fakeOrgStore) SaveSession(_ context.Context, session auth.Session) error {
	f.sessions[session.TokenHash] = session
	return nil
}

func (f *fakeOrgStore) GetSessionByTokenHash(_ context.Context, tokenHash string) (auth.Session, error) {
	session, ok := f.sessions[tokenHash]
	if !ok || session.RevokedAt != nil || time.Now().After(session.ExpiresAt) {
		return auth.Session{}, errNotFound("session", tokenHash)
	}
	return session, nil
}

func (f *fakeOrgStore) RevokeSession(_ context.Context, id string) error {
	for key, session := range f.sessions {
		if session.ID == id {
			now := time.Now()
			session.RevokedAt = &now
			f.sessions[key] = session
			return nil
		}
	}
	return errNotFound("session", id)
}

func (f *fakeOrgStore) SaveOAuthState(_ context.Context, state auth.OAuthState) error {
	f.oauthStates[state.State] = state
	return nil
}

func (f *fakeOrgStore) ConsumeOAuthState(_ context.Context, provider, state string, now time.Time) (auth.OAuthState, error) {
	value, ok := f.oauthStates[state]
	if !ok || value.Provider != provider || value.ConsumedAt != nil || !value.ExpiresAt.After(now) {
		return auth.OAuthState{}, errors.New("oauth state not found or expired")
	}
	value.ConsumedAt = &now
	f.oauthStates[state] = value
	return value, nil
}

func (f *fakeOrgStore) GetOrganization(_ context.Context, id string) (org.Organization, error) {
	o, ok := f.orgs[id]
	if !ok {
		return org.Organization{}, errNotFound("organization", id)
	}
	return o, nil
}

func (f *fakeOrgStore) SaveOrganization(_ context.Context, o org.Organization) error {
	f.orgs[o.ID] = o
	return nil
}

func (f *fakeOrgStore) ListOrganizations(_ context.Context) ([]org.Organization, error) {
	result := make([]org.Organization, 0, len(f.orgs))
	for _, o := range f.orgs {
		result = append(result, o)
	}
	return result, nil
}

func (f *fakeOrgStore) GetMembershipsByUser(_ context.Context, userID string) ([]org.Membership, error) {
	var result []org.Membership
	for _, m := range f.memberships {
		if m.UserID == userID {
			result = append(result, m)
		}
	}
	return result, nil
}

func (f *fakeOrgStore) GetMembershipsByOrg(_ context.Context, orgID string) ([]org.Membership, error) {
	var result []org.Membership
	for _, m := range f.memberships {
		if m.OrganizationID == orgID {
			result = append(result, m)
		}
	}
	return result, nil
}

func (f *fakeOrgStore) SaveMembership(_ context.Context, m org.Membership) error {
	f.memberships = append(f.memberships, m)
	return nil
}

func (f *fakeOrgStore) DeleteMembership(_ context.Context, orgID, userID string) error {
	for i, m := range f.memberships {
		if m.OrganizationID == orgID && m.UserID == userID {
			f.memberships = append(f.memberships[:i], f.memberships[i+1:]...)
			return nil
		}
	}
	return fmt.Errorf("membership not found")
}

func (f *fakeOrgStore) GetPolicy(_ context.Context, id string) (org.Policy, error) {
	for _, p := range f.policies {
		if p.ID == id {
			return p, nil
		}
	}
	return org.Policy{}, errNotFound("policy", id)
}

func (f *fakeOrgStore) SavePolicy(_ context.Context, p org.Policy) error {
	for i, existing := range f.policies {
		if existing.ID == p.ID {
			f.policies[i] = p
			return nil
		}
	}
	f.policies = append(f.policies, p)
	return nil
}

func (f *fakeOrgStore) ListPoliciesByOrg(_ context.Context, orgID string) ([]org.Policy, error) {
	var result []org.Policy
	for _, p := range f.policies {
		if p.OrganizationID == orgID {
			result = append(result, p)
		}
	}
	return result, nil
}

func (f *fakeOrgStore) GetRepository(_ context.Context, id string) (repository.Repository, error) {
	for _, r := range f.repos {
		if r.ID == id {
			return r, nil
		}
	}
	return repository.Repository{}, fmt.Errorf("repository %s not found", id)
}

func (f *fakeOrgStore) ListRepositoriesByOrg(_ context.Context, orgID string) ([]repository.Repository, error) {
	var result []repository.Repository
	for _, r := range f.repos {
		if r.OrganizationID == orgID {
			result = append(result, r)
		}
	}
	return result, nil
}

func (f *fakeOrgStore) SaveRepository(_ context.Context, repo repository.Repository) error {
	for i, existing := range f.repos {
		if existing.ID == repo.ID {
			f.repos[i] = repo
			return nil
		}
	}
	f.repos = append(f.repos, repo)
	return nil
}

func errNotFound(resource, id string) error {
	return &notFoundError{resource: resource, id: id}
}

type notFoundError struct{ resource, id string }

func (e *notFoundError) Error() string {
	return e.resource + " " + e.id + " not found"
}

// fakeInsightsReader is a test double for InsightsReader.
type fakeInsightsReader struct {
	data insights.OrganizationInsights
}

func (f *fakeInsightsReader) GetOrganizationInsights(_ context.Context, orgID string) (insights.OrganizationInsights, error) {
	f.data.OrganizationID = orgID
	return f.data, nil
}

// --- Tests ---

func TestListOrganizationsReturnsOrgForMockUser(t *testing.T) {
	store := newFakeOrgStore()
	o := org.Organization{ID: "org_1", Name: "Test Org"}
	store.orgs["org_1"] = o
	store.memberships = []org.Membership{{OrganizationID: "org_1", UserID: "mock_user", Role: org.RoleOwner}}

	handler := newDevHeaderTestServer(nil, nil, nil, store).Handler()
	req := httptest.NewRequest(http.MethodGet, "/api/v1/organizations", nil)
	rw := httptest.NewRecorder()
	handler.ServeHTTP(rw, req)

	if rw.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d: %s", rw.Code, rw.Body.String())
	}
	if !bytes.Contains(rw.Body.Bytes(), []byte("Test Org")) {
		t.Fatalf("expected response to contain 'Test Org', got: %s", rw.Body.String())
	}
}

func TestGetOrganizationNotFound(t *testing.T) {
	store := newFakeOrgStore()
	handler := newDevHeaderTestServer(nil, nil, nil, store).Handler()

	req := httptest.NewRequest(http.MethodGet, "/api/v1/organizations/nonexistent", nil)
	rw := httptest.NewRecorder()
	handler.ServeHTTP(rw, req)

	if rw.Code != http.StatusNotFound {
		t.Fatalf("expected 404, got %d", rw.Code)
	}
}

func TestAddMemberAndListMembers(t *testing.T) {
	store := newFakeOrgStore()
	store.orgs["org_1"] = org.Organization{ID: "org_1", Name: "Org"}
	store.memberships = []org.Membership{{OrganizationID: "org_1", UserID: "mock_user", Role: org.RoleOwner}}
	handler := newDevHeaderTestServer(nil, nil, nil, store).Handler()

	body := bytes.NewBufferString(`{"user_id":"user_42","role":"member"}`)
	req := httptest.NewRequest(http.MethodPost, "/api/v1/organizations/org_1/members", body)
	rw := httptest.NewRecorder()
	handler.ServeHTTP(rw, req)

	if rw.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d: %s", rw.Code, rw.Body.String())
	}

	// List members
	req2 := httptest.NewRequest(http.MethodGet, "/api/v1/organizations/org_1/members", nil)
	rw2 := httptest.NewRecorder()
	handler.ServeHTTP(rw2, req2)

	if rw2.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d: %s", rw2.Code, rw2.Body.String())
	}
	if !bytes.Contains(rw2.Body.Bytes(), []byte("user_42")) {
		t.Fatalf("expected 'user_42' in response, got: %s", rw2.Body.String())
	}
}

func TestListPoliciesByOrg(t *testing.T) {
	store := newFakeOrgStore()
	store.policies = []org.Policy{
		{ID: "pol_1", OrganizationID: "org_1", Name: "policy_a"},
		{ID: "pol_2", OrganizationID: "org_2", Name: "policy_b"},
	}
	// mock_user must be a member of org_1 for auth to pass
	store.memberships = []org.Membership{
		{OrganizationID: "org_1", UserID: "mock_user", Role: org.RoleAdmin},
	}

	handler := newDevHeaderTestServer(nil, nil, nil, store).Handler()
	req := httptest.NewRequest(http.MethodGet, "/api/v1/organizations/org_1/policies", nil)
	rw := httptest.NewRecorder()
	handler.ServeHTTP(rw, req)

	if rw.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d: %s", rw.Code, rw.Body.String())
	}
	if !bytes.Contains(rw.Body.Bytes(), []byte("policy_a")) {
		t.Fatalf("expected policy_a in response: %s", rw.Body.String())
	}
	if bytes.Contains(rw.Body.Bytes(), []byte("policy_b")) {
		t.Fatalf("should NOT contain org_2 policy in org_1 response: %s", rw.Body.String())
	}
}

func TestOrgInsightsEndpoint(t *testing.T) {
	insightsReader := &fakeInsightsReader{
		data: insights.OrganizationInsights{
			AverageScore:      77,
			TotalRepositories: 5,
			TotalScans:        20,
		},
	}

	// Wire insights into server manually (field is unexported, use embedded test approach)
	s := &Server{insights: insightsReader}
	handler := s.Handler()
	req := httptest.NewRequest(http.MethodGet, "/api/v1/organizations/org_1/insights", nil)
	rw := httptest.NewRecorder()
	handler.ServeHTTP(rw, req)

	if rw.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d: %s", rw.Code, rw.Body.String())
	}

	var resp map[string]any
	if err := json.Unmarshal(rw.Body.Bytes(), &resp); err != nil {
		t.Fatalf("decode response: %v", err)
	}
	data, _ := resp["data"].(map[string]any)
	if data == nil {
		t.Fatalf("expected data field: %s", rw.Body.String())
	}
}

func TestMetricsEndpoint(t *testing.T) {
	handler := NewServer(nil, nil, nil, nil, nil).Handler()
	req := httptest.NewRequest(http.MethodGet, "/api/v1/metrics", nil)
	rw := httptest.NewRecorder()
	handler.ServeHTTP(rw, req)

	if rw.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d: %s", rw.Code, rw.Body.String())
	}
	if !bytes.Contains(rw.Body.Bytes(), []byte("uptime_seconds")) {
		t.Fatalf("expected uptime_seconds in metrics: %s", rw.Body.String())
	}
	if !bytes.Contains(rw.Body.Bytes(), []byte("goroutines")) {
		t.Fatalf("expected goroutines in metrics: %s", rw.Body.String())
	}
}

func TestRemoveMemberRequiresAdminRole(t *testing.T) {
	store := newFakeOrgStore()
	store.orgs["org_1"] = org.Organization{ID: "org_1", Name: "Org"}
	// mock_user is admin; target is user_42
	store.memberships = []org.Membership{
		{OrganizationID: "org_1", UserID: "mock_user", Role: org.RoleAdmin},
		{OrganizationID: "org_1", UserID: "user_42", Role: org.RoleMember},
	}
	handler := newDevHeaderTestServer(nil, nil, nil, store).Handler()

	req := httptest.NewRequest(http.MethodDelete, "/api/v1/organizations/org_1/members/user_42", nil)
	rw := httptest.NewRecorder()
	handler.ServeHTTP(rw, req)

	if rw.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d: %s", rw.Code, rw.Body.String())
	}
	assertJSONField(t, rw.Body.Bytes(), "status", "removed")

	// Verify member is gone
	if len(store.memberships) != 1 {
		t.Fatalf("expected 1 remaining member, got %d", len(store.memberships))
	}
}

func TestRemoveMemberForbiddenForNonMember(t *testing.T) {
	store := newFakeOrgStore()
	store.orgs["org_1"] = org.Organization{ID: "org_1", Name: "Org"}
	// mock_user is a plain member (not admin/owner) — cannot remove others
	store.memberships = []org.Membership{
		{OrganizationID: "org_1", UserID: "mock_user", Role: org.RoleMember},
		{OrganizationID: "org_1", UserID: "user_42", Role: org.RoleMember},
	}
	handler := newDevHeaderTestServer(nil, nil, nil, store).Handler()

	req := httptest.NewRequest(http.MethodDelete, "/api/v1/organizations/org_1/members/user_42", nil)
	rw := httptest.NewRecorder()
	handler.ServeHTTP(rw, req)

	if rw.Code != http.StatusForbidden {
		t.Fatalf("expected 403, got %d: %s", rw.Code, rw.Body.String())
	}
}

// TestOrgRepositoriesListEndpoint verifies org repo list returns only repos in the org (T6-021).
func TestOrgRepositoriesListEndpoint(t *testing.T) {
	store := newFakeOrgStore()
	store.memberships = []org.Membership{
		{OrganizationID: "org_1", UserID: "mock_user", Role: org.RoleAdmin},
	}
	store.repos = []repository.Repository{
		{ID: "repo_1", Name: "alpha", OrganizationID: "org_1"},
		{ID: "repo_2", Name: "beta", OrganizationID: "org_2"},
	}

	handler := newDevHeaderTestServer(nil, nil, nil, store).Handler()
	req := httptest.NewRequest(http.MethodGet, "/api/v1/organizations/org_1/repositories", nil)
	rw := httptest.NewRecorder()
	handler.ServeHTTP(rw, req)

	if rw.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d: %s", rw.Code, rw.Body.String())
	}
	if !bytes.Contains(rw.Body.Bytes(), []byte("alpha")) {
		t.Fatalf("expected 'alpha' in response: %s", rw.Body.String())
	}
	if bytes.Contains(rw.Body.Bytes(), []byte("beta")) {
		t.Fatalf("should NOT contain org_2 repo 'beta': %s", rw.Body.String())
	}
}
