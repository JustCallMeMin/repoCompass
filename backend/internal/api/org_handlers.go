package api

import (
	"context"
	"net/http"

	"github.com/JustCallMeMin/repoCompass/backend/internal/org"
	"github.com/JustCallMeMin/repoCompass/backend/internal/repository"
)

// OrgStore defines the necessary storage operations for organization handlers.
type OrgStore interface {
	GetOrganization(ctx context.Context, id string) (org.Organization, error)
	ListOrganizations(ctx context.Context) ([]org.Organization, error)
	GetMembershipsByUser(ctx context.Context, userID string) ([]org.Membership, error)
	GetMembershipsByOrg(ctx context.Context, orgID string) ([]org.Membership, error)
	SaveMembership(ctx context.Context, m org.Membership) error
	DeleteMembership(ctx context.Context, orgID, userID string) error
	GetPolicy(ctx context.Context, id string) (org.Policy, error)
	SavePolicy(ctx context.Context, p org.Policy) error
	ListPoliciesByOrg(ctx context.Context, orgID string) ([]org.Policy, error)
	// Repository scope verification (T6-015)
	GetRepository(ctx context.Context, id string) (repository.Repository, error)
	ListRepositoriesByOrg(ctx context.Context, orgID string) ([]repository.Repository, error)
}

func (s *Server) handleListOrganizations(w http.ResponseWriter, r *http.Request) {
	if s.orgs == nil {
		writeError(w, http.StatusNotImplemented, "not_implemented", "organization storage not configured")
		return
	}

	memberships, err := s.orgs.GetMembershipsByUser(r.Context(), actorIDFromRequest(r))
	if err != nil {
		writeError(w, http.StatusInternalServerError, "db_error", "failed to fetch user memberships")
		return
	}

	var accessibleOrgs []org.Organization
	for _, m := range memberships {
		o, err := s.orgs.GetOrganization(r.Context(), m.OrganizationID)
		if err == nil {
			accessibleOrgs = append(accessibleOrgs, o)
		}
	}

	writeJSON(w, http.StatusOK, map[string]any{"data": accessibleOrgs})
}

func (s *Server) handleGetOrganization(w http.ResponseWriter, r *http.Request) {
	if s.orgs == nil {
		writeError(w, http.StatusNotImplemented, "not_implemented", "organization storage not configured")
		return
	}

	orgID := r.PathValue("organization_id")
	if orgID == "" {
		writeError(w, http.StatusBadRequest, "invalid_request", "organization_id path variable missing")
		return
	}

	o, err := s.orgs.GetOrganization(r.Context(), orgID)
	if err != nil {
		writeError(w, http.StatusNotFound, "not_found", "organization not found")
		return
	}

	writeJSON(w, http.StatusOK, map[string]any{"data": o})
}

func (s *Server) handleListMemberships(w http.ResponseWriter, r *http.Request) {
	if s.orgs == nil {
		writeError(w, http.StatusNotImplemented, "not_implemented", "organization storage not configured")
		return
	}

	orgID := r.PathValue("organization_id")
	memberships, err := s.orgs.GetMembershipsByOrg(r.Context(), orgID)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "db_error", "failed to fetch org memberships")
		return
	}

	writeJSON(w, http.StatusOK, map[string]any{"data": memberships})
}

func (s *Server) handleAddMember(w http.ResponseWriter, r *http.Request) {
	if s.orgs == nil {
		writeError(w, http.StatusNotImplemented, "not_implemented", "organization storage not configured")
		return
	}

	orgID := r.PathValue("organization_id")
	if orgID == "" {
		writeError(w, http.StatusBadRequest, "invalid_request", "organization_id missing")
		return
	}

	if s.authSvc != nil {
		if err := s.authSvc.CheckManageOrg(r.Context(), actorIDFromRequest(r), orgID); err != nil {
			writeError(w, http.StatusForbidden, "forbidden", "owner or admin role required to add members")
			return
		}
	}

	var req struct {
		UserID string   `json:"user_id"`
		Role   org.Role `json:"role"`
	}
	if err := decodeJSON(r, &req); err != nil {
		writeError(w, http.StatusBadRequest, "invalid_json", err.Error())
		return
	}

	mem := org.Membership{
		OrganizationID: orgID,
		UserID:         req.UserID,
		Role:           req.Role,
	}

	if err := s.orgs.SaveMembership(r.Context(), mem); err != nil {
		writeError(w, http.StatusInternalServerError, "db_error", "failed to save membership")
		return
	}

	writeJSON(w, http.StatusOK, map[string]string{"status": "success"})
}

func (s *Server) handleRemoveMember(w http.ResponseWriter, r *http.Request) {
	if s.orgs == nil {
		writeError(w, http.StatusNotImplemented, "not_implemented", "organization storage not configured")
		return
	}

	orgID := r.PathValue("organization_id")
	targetUserID := r.PathValue("user_id")
	if orgID == "" || targetUserID == "" {
		writeError(w, http.StatusBadRequest, "invalid_request", "organization_id and user_id are required")
		return
	}

	// Only admin/owner can remove members
	if s.authSvc != nil {
		if err := s.authSvc.CheckManageOrg(r.Context(), actorIDFromRequest(r), orgID); err != nil {
			writeError(w, http.StatusForbidden, "forbidden", "admin or owner role required to remove members")
			return
		}
	}

	if err := s.orgs.DeleteMembership(r.Context(), orgID, targetUserID); err != nil {
		writeError(w, http.StatusInternalServerError, "db_error", "failed to remove member")
		return
	}

	writeJSON(w, http.StatusOK, map[string]string{"status": "removed"})
}

func (s *Server) handleGetPolicy(w http.ResponseWriter, r *http.Request) {
	if s.orgs == nil {
		writeError(w, http.StatusNotImplemented, "not_implemented", "organization storage not configured")
		return
	}
	policyID := r.PathValue("policy_id")
	if policyID == "" {
		writeError(w, http.StatusBadRequest, "invalid_request", "policy_id missing")
		return
	}

	p, err := s.orgs.GetPolicy(r.Context(), policyID)
	if err != nil {
		writeError(w, http.StatusNotFound, "not_found", "policy not found")
		return
	}

	// Verify the caller has access to the organization
	if s.authSvc != nil {
		if err := s.authSvc.CheckAccessRepo(r.Context(), actorIDFromRequest(r), p.OrganizationID); err != nil {
			writeError(w, http.StatusForbidden, "forbidden", "access denied to organization policies")
			return
		}
	}

	writeJSON(w, http.StatusOK, map[string]any{"data": p})
}

func (s *Server) handleSavePolicy(w http.ResponseWriter, r *http.Request) {
	if s.orgs == nil {
		writeError(w, http.StatusNotImplemented, "not_implemented", "organization storage not configured")
		return
	}

	orgID := r.PathValue("organization_id")
	if orgID == "" {
		writeError(w, http.StatusBadRequest, "invalid_request", "organization_id missing")
		return
	}

	if s.authSvc != nil {
		if err := s.authSvc.CheckManageOrg(r.Context(), actorIDFromRequest(r), orgID); err != nil {
			writeError(w, http.StatusForbidden, "forbidden", "manage org access required")
			return
		}
	}

	var req org.Policy
	if err := decodeJSON(r, &req); err != nil {
		writeError(w, http.StatusBadRequest, "invalid_json", err.Error())
		return
	}
	req.OrganizationID = orgID
	if req.ID == "" {
		writeError(w, http.StatusBadRequest, "invalid_request", "policy id missing")
		return
	}

	if err := s.orgs.SavePolicy(r.Context(), req); err != nil {
		writeError(w, http.StatusInternalServerError, "db_error", "failed to save policy")
		return
	}

	writeJSON(w, http.StatusOK, map[string]string{"status": "success"})
}

// handleListOrgRepositories lists repositories belonging to an organization (T6-021).
func (s *Server) handleListOrgRepositories(w http.ResponseWriter, r *http.Request) {
	if s.orgs == nil {
		writeError(w, http.StatusNotImplemented, "not_implemented", "organization storage not configured")
		return
	}

	orgID := r.PathValue("organization_id")
	if orgID == "" {
		writeError(w, http.StatusBadRequest, "invalid_request", "organization_id missing")
		return
	}

	if s.authSvc != nil {
		if err := s.authSvc.CheckAccessRepo(r.Context(), actorIDFromRequest(r), orgID); err != nil {
			writeError(w, http.StatusForbidden, "forbidden", err.Error())
			return
		}
	}

	repos, err := s.orgs.ListRepositoriesByOrg(r.Context(), orgID)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "db_error", "failed to list repositories")
		return
	}
	if repos == nil {
		repos = []repository.Repository{}
	}
	writeJSON(w, http.StatusOK, map[string]any{"data": repos})
}
