package api

import (
	"net/http"

	"github.com/JustCallMeMin/repoCompass/backend/internal/org"
)

// handleListPolicies lists all policies for an organization.
func (s *Server) handleListPolicies(w http.ResponseWriter, r *http.Request) {
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

	policies, err := s.orgs.ListPoliciesByOrg(r.Context(), orgID)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "db_error", "failed to list policies")
		return
	}
	if policies == nil {
		policies = []org.Policy{}
	}
	writeData(w, r, http.StatusOK, policies)
}

// handleOrgInsights returns aggregated org-level insights.
func (s *Server) handleOrgInsights(w http.ResponseWriter, r *http.Request) {
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

	if s.insights == nil {
		writeError(w, http.StatusNotImplemented, "not_implemented", "insights store not configured")
		return
	}

	result, err := s.insights.GetOrganizationInsights(r.Context(), orgID)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "insights_query_failed", err.Error())
		return
	}

	writeData(w, r, http.StatusOK, result)
}
