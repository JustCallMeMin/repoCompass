package api

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"strings"
	"time"

	"github.com/JustCallMeMin/repoCompass/backend/internal/audit"
	"github.com/JustCallMeMin/repoCompass/backend/internal/notification"
	"github.com/JustCallMeMin/repoCompass/backend/internal/org"
	"github.com/JustCallMeMin/repoCompass/backend/internal/repository"
)

// OrgStore defines the necessary storage operations for organization handlers.
type OrgStore interface {
	SaveOrganization(ctx context.Context, o org.Organization) error
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

// NotificationStore defines storage used by M6 notification endpoints.
type NotificationStore interface {
	SaveNotificationPreference(ctx context.Context, pref notification.Preference) error
	GetNotificationPreference(ctx context.Context, orgID, userID string) (notification.Preference, error)
	SaveNotification(ctx context.Context, n notification.Notification) error
	ListNotifications(ctx context.Context, orgID, userID string, limit int) ([]notification.Notification, error)
	MarkNotificationRead(ctx context.Context, id, userID string, readAt time.Time) error
	SaveNotificationDelivery(ctx context.Context, d notification.Delivery) error
}

// AuditStore defines durable audit event storage.
type AuditStore interface {
	SaveAuditEntry(ctx context.Context, entry audit.Entry) error
}

func (s *Server) handleListOrganizations(w http.ResponseWriter, r *http.Request) {
	if s.orgs == nil {
		writeError(w, http.StatusNotImplemented, "not_implemented", "organization storage not configured")
		return
	}

	userID := s.authenticatedActorID(r)
	if userID == "" {
		writeRequestError(w, r, http.StatusUnauthorized, "unauthorized", "session is required")
		return
	}
	memberships, err := s.orgs.GetMembershipsByUser(r.Context(), userID)
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

	writeData(w, r, http.StatusOK, accessibleOrgs)
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
	if !s.authorize(w, r, orgID, "read") {
		return
	}

	writeData(w, r, http.StatusOK, o)
}

// handleCreateOrganization creates an organization and makes the actor owner.
func (s *Server) handleCreateOrganization(w http.ResponseWriter, r *http.Request) {
	if s.orgs == nil {
		writeRequestError(w, r, http.StatusNotImplemented, "not_implemented", "organization storage not configured")
		return
	}
	var req struct {
		ID   string `json:"id"`
		Name string `json:"name"`
	}
	if err := decodeJSON(r, &req); err != nil {
		writeRequestError(w, r, http.StatusBadRequest, "invalid_json", err.Error())
		return
	}
	req.ID = strings.TrimSpace(req.ID)
	req.Name = strings.TrimSpace(req.Name)
	if req.ID == "" || req.Name == "" {
		writeRequestError(w, r, http.StatusBadRequest, "invalid_request", "id and name are required")
		return
	}
	now := time.Now().UTC()
	value := org.Organization{ID: req.ID, Name: req.Name, CreatedAt: now, UpdatedAt: now}
	if err := s.orgs.SaveOrganization(r.Context(), value); err != nil {
		writeRequestError(w, r, http.StatusInternalServerError, "db_error", "failed to save organization")
		return
	}
	actor := s.authenticatedActorID(r)
	_ = s.orgs.SaveMembership(r.Context(), org.Membership{OrganizationID: value.ID, UserID: actor, Role: org.RoleOwner, CreatedAt: now, UpdatedAt: now})
	s.recordAudit(r, audit.EventOrgCreated, value.ID, "org", value.ID, map[string]any{"name": value.Name})
	s.emitNotification(r, notification.EventOrgCreated, value.ID, "org", value.ID, "info", "Organization created", value.Name)
	writeData(w, r, http.StatusCreated, value)
}

// handleUpdateOrganization updates organization metadata.
func (s *Server) handleUpdateOrganization(w http.ResponseWriter, r *http.Request) {
	if s.orgs == nil {
		writeRequestError(w, r, http.StatusNotImplemented, "not_implemented", "organization storage not configured")
		return
	}
	orgID := strings.TrimSpace(r.PathValue("organization_id"))
	if orgID == "" {
		writeRequestError(w, r, http.StatusBadRequest, "invalid_request", "organization_id missing")
		return
	}
	if !s.authorize(w, r, orgID, "manage") {
		return
	}
	current, err := s.orgs.GetOrganization(r.Context(), orgID)
	if err != nil {
		writeRequestError(w, r, http.StatusNotFound, "not_found", "organization not found")
		return
	}
	var req struct {
		Name string `json:"name"`
	}
	if err := decodeJSON(r, &req); err != nil {
		writeRequestError(w, r, http.StatusBadRequest, "invalid_json", err.Error())
		return
	}
	current.Name = strings.TrimSpace(req.Name)
	if current.Name == "" {
		writeRequestError(w, r, http.StatusBadRequest, "invalid_request", "name is required")
		return
	}
	current.UpdatedAt = time.Now().UTC()
	if err := s.orgs.SaveOrganization(r.Context(), current); err != nil {
		writeRequestError(w, r, http.StatusInternalServerError, "db_error", "failed to update organization")
		return
	}
	s.recordAudit(r, audit.EventOrgUpdated, orgID, "org", orgID, map[string]any{"name": current.Name})
	s.emitNotification(r, notification.EventOrgUpdated, orgID, "org", orgID, "info", "Organization updated", current.Name)
	writeData(w, r, http.StatusOK, current)
}

func (s *Server) handleListMemberships(w http.ResponseWriter, r *http.Request) {
	if s.orgs == nil {
		writeError(w, http.StatusNotImplemented, "not_implemented", "organization storage not configured")
		return
	}

	orgID := r.PathValue("organization_id")
	if !s.authorize(w, r, orgID, "read") {
		return
	}
	memberships, err := s.orgs.GetMembershipsByOrg(r.Context(), orgID)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "db_error", "failed to fetch org memberships")
		return
	}

	writeData(w, r, http.StatusOK, memberships)
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

	if !s.authorize(w, r, orgID, "manage") {
		return
	}

	var req struct {
		UserID string   `json:"user_id"`
		Role   org.Role `json:"role"`
	}
	if err := decodeJSON(r, &req); err != nil {
		writeError(w, http.StatusBadRequest, "invalid_json", err.Error())
		return
	}
	req.UserID = strings.TrimSpace(req.UserID)
	if req.UserID == "" || !org.ValidRole(req.Role) {
		writeRequestError(w, r, http.StatusBadRequest, "invalid_request", "valid user_id and role are required")
		return
	}
	now := time.Now().UTC()

	mem := org.Membership{
		OrganizationID: orgID,
		UserID:         req.UserID,
		Role:           req.Role,
		CreatedAt:      now,
		UpdatedAt:      now,
	}

	if err := s.orgs.SaveMembership(r.Context(), mem); err != nil {
		writeError(w, http.StatusInternalServerError, "db_error", "failed to save membership")
		return
	}

	s.recordAudit(r, audit.EventMemberAdded, orgID, "membership", req.UserID, map[string]any{"role": req.Role})
	s.emitNotification(r, notification.EventMemberAdded, orgID, "membership", req.UserID, "info", "Member added", req.UserID+" joined as "+string(req.Role))
	writeData(w, r, http.StatusOK, mem)
}

// handleUpdateMember changes a member role with last-owner protection.
func (s *Server) handleUpdateMember(w http.ResponseWriter, r *http.Request) {
	if s.orgs == nil {
		writeRequestError(w, r, http.StatusNotImplemented, "not_implemented", "organization storage not configured")
		return
	}
	orgID := strings.TrimSpace(r.PathValue("organization_id"))
	userID := strings.TrimSpace(r.PathValue("user_id"))
	if orgID == "" || userID == "" {
		writeRequestError(w, r, http.StatusBadRequest, "invalid_request", "organization_id and user_id are required")
		return
	}
	if !s.authorize(w, r, orgID, "manage") {
		return
	}
	var req struct {
		Role org.Role `json:"role"`
	}
	if err := decodeJSON(r, &req); err != nil {
		writeRequestError(w, r, http.StatusBadRequest, "invalid_json", err.Error())
		return
	}
	if !org.ValidRole(req.Role) {
		writeRequestError(w, r, http.StatusBadRequest, "invalid_request", "valid role is required")
		return
	}
	if err := s.ensureOwnerProtection(r.Context(), orgID, userID, req.Role); err != nil {
		writeRequestError(w, r, http.StatusBadRequest, "owner_protection", err.Error())
		return
	}
	now := time.Now().UTC()
	mem := org.Membership{OrganizationID: orgID, UserID: userID, Role: req.Role, CreatedAt: now, UpdatedAt: now}
	if err := s.orgs.SaveMembership(r.Context(), mem); err != nil {
		writeRequestError(w, r, http.StatusInternalServerError, "db_error", "failed to update membership")
		return
	}
	s.recordAudit(r, audit.EventMemberRoleChanged, orgID, "membership", userID, map[string]any{"role": req.Role})
	s.emitNotification(r, notification.EventMemberRoleChanged, orgID, "membership", userID, "info", "Member role changed", userID+" is now "+string(req.Role))
	writeData(w, r, http.StatusOK, mem)
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
	if !s.authorize(w, r, orgID, "manage") {
		return
	}
	if err := s.ensureOwnerProtection(r.Context(), orgID, targetUserID, ""); err != nil {
		writeRequestError(w, r, http.StatusBadRequest, "owner_protection", err.Error())
		return
	}

	if err := s.orgs.DeleteMembership(r.Context(), orgID, targetUserID); err != nil {
		writeError(w, http.StatusInternalServerError, "db_error", "failed to remove member")
		return
	}

	s.recordAudit(r, audit.EventMemberRemoved, orgID, "membership", targetUserID, nil)
	s.emitNotification(r, notification.EventMemberRemoved, orgID, "membership", targetUserID, "warning", "Member removed", targetUserID+" was removed")
	writeData(w, r, http.StatusOK, map[string]string{"status": "removed"})
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
	if !s.authorize(w, r, p.OrganizationID, "read") {
		return
	}

	writeData(w, r, http.StatusOK, p)
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

	if !s.authorize(w, r, orgID, "manage") {
		return
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
	if err := validatePolicy(req); err != nil {
		writeRequestError(w, r, http.StatusBadRequest, "invalid_policy", err.Error())
		return
	}
	if req.Status == "" {
		req.Status = "active"
	}
	if req.Version <= 0 {
		req.Version = 1
	}
	now := time.Now().UTC()
	if req.CreatedAt.IsZero() {
		req.CreatedAt = now
	}
	req.UpdatedAt = now

	if err := s.orgs.SavePolicy(r.Context(), req); err != nil {
		writeError(w, http.StatusInternalServerError, "db_error", "failed to save policy")
		return
	}

	s.recordAudit(r, audit.EventPolicyUpdated, orgID, "policy", req.ID, map[string]any{"name": req.Name, "status": req.Status})
	s.emitNotification(r, notification.EventPolicyUpdated, orgID, "policy", req.ID, "info", "Policy updated", req.Name)
	writeData(w, r, http.StatusOK, req)
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

	if !s.authorize(w, r, orgID, "read") {
		return
	}

	repos, err := s.orgs.ListRepositoriesByOrg(r.Context(), orgID)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "db_error", "failed to list repositories")
		return
	}
	if repos == nil {
		repos = []repository.Repository{}
	}
	writeData(w, r, http.StatusOK, repos)
}

// validatePolicy checks the minimum policy contract accepted by M6.
func validatePolicy(p org.Policy) error {
	p.Name = strings.TrimSpace(p.Name)
	if p.Name == "" {
		return errBadRequest("policy name is required")
	}
	if p.Status != "" && p.Status != "active" && p.Status != "draft" && p.Status != "disabled" {
		return errBadRequest("policy status must be active, draft, or disabled")
	}
	if len(p.Rules) == 0 {
		return errBadRequest("policy rules are required")
	}
	var raw map[string]any
	if err := json.Unmarshal(p.Rules, &raw); err != nil {
		return errBadRequest("policy rules must be a JSON object")
	}
	return nil
}

// ensureOwnerProtection prevents removing or demoting the last owner.
func (s *Server) ensureOwnerProtection(ctx context.Context, orgID, targetUserID string, nextRole org.Role) error {
	members, err := s.orgs.GetMembershipsByOrg(ctx, orgID)
	if err != nil {
		return err
	}
	owners := 0
	targetIsOwner := false
	for _, member := range members {
		if member.Role == org.RoleOwner {
			owners++
		}
		if member.UserID == targetUserID && member.Role == org.RoleOwner {
			targetIsOwner = true
		}
	}
	if targetIsOwner && owners <= 1 && nextRole != org.RoleOwner {
		return errBadRequest("cannot remove or demote the last organization owner")
	}
	return nil
}

// recordAudit writes a best-effort durable audit event.
func (s *Server) recordAudit(r *http.Request, typ audit.EventType, orgID, resourceType, resourceID string, details map[string]any) {
	if s.audit == nil {
		return
	}
	now := time.Now().UTC()
	_ = s.audit.SaveAuditEntry(r.Context(), audit.Entry{
		ID:             "aud_" + stableHash(string(typ)+orgID+resourceID+now.Format(time.RFC3339Nano)),
		Type:           typ,
		OrganizationID: orgID,
		ActorID:        s.authenticatedActorID(r),
		ResourceType:   resourceType,
		ResourceID:     resourceID,
		Details:        details,
		OccurredAt:     now,
	})
}

// emitNotification creates in-app notification and configured channel deliveries.
func (s *Server) emitNotification(r *http.Request, typ notification.EventType, orgID, resourceType, resourceID, severity, title, message string) {
	if s.notifications == nil {
		return
	}
	now := time.Now().UTC()
	n := notification.Notification{
		ID:             "ntf_" + stableHash(string(typ)+orgID+resourceID+now.Format(time.RFC3339Nano)),
		OrganizationID: orgID,
		Type:           typ,
		Severity:       severity,
		Title:          title,
		Message:        message,
		ResourceType:   resourceType,
		ResourceID:     resourceID,
		CreatedAt:      now,
	}
	_ = s.notifications.SaveNotification(r.Context(), n)
	_ = s.notifications.SaveNotificationDelivery(r.Context(), notification.Delivery{
		ID:             "del_" + stableHash(n.ID+"in_app"),
		NotificationID: n.ID,
		Channel:        notification.ChannelInApp,
		Status:         "sent",
		AttemptedAt:    now,
	})
	pref, err := s.notifications.GetNotificationPreference(r.Context(), orgID, s.authenticatedActorID(r))
	if err != nil {
		return
	}
	for _, channel := range pref.Channels {
		switch channel {
		case notification.ChannelWebhook:
			s.deliverWebhookNotification(r, n, pref.WebhookURL)
		case notification.ChannelEmail:
			_ = s.notifications.SaveNotificationDelivery(r.Context(), notification.Delivery{
				ID:             "del_" + stableHash(n.ID+"email"),
				NotificationID: n.ID,
				Channel:        notification.ChannelEmail,
				Status:         "stubbed",
				Target:         pref.Email,
				AttemptedAt:    time.Now().UTC(),
			})
		}
	}
}

// deliverWebhookNotification sends a notification to a configured webhook URL.
func (s *Server) deliverWebhookNotification(r *http.Request, n notification.Notification, target string) {
	if strings.TrimSpace(target) == "" {
		_ = s.notifications.SaveNotificationDelivery(r.Context(), notification.Delivery{
			ID:             "del_" + stableHash(n.ID+"webhook-missing"),
			NotificationID: n.ID,
			Channel:        notification.ChannelWebhook,
			Status:         "failed",
			ErrorMessage:   "webhook_url is not configured",
			AttemptedAt:    time.Now().UTC(),
		})
		return
	}
	payload, err := json.Marshal(n)
	if err != nil {
		return
	}
	ctx, cancel := context.WithTimeout(r.Context(), 3*time.Second)
	defer cancel()
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, target, bytes.NewReader(payload))
	if err != nil {
		return
	}
	req.Header.Set("Content-Type", "application/json")
	resp, err := http.DefaultClient.Do(req)
	status := "sent"
	errorMessage := ""
	if err != nil {
		status = "failed"
		errorMessage = err.Error()
	} else {
		defer resp.Body.Close()
		if resp.StatusCode < 200 || resp.StatusCode >= 300 {
			status = "failed"
			errorMessage = resp.Status
		}
	}
	_ = s.notifications.SaveNotificationDelivery(r.Context(), notification.Delivery{
		ID:             "del_" + stableHash(n.ID+"webhook"+status),
		NotificationID: n.ID,
		Channel:        notification.ChannelWebhook,
		Status:         status,
		Target:         target,
		ErrorMessage:   errorMessage,
		AttemptedAt:    time.Now().UTC(),
	})
}

type badRequestError struct{ message string }

// Error returns the user-facing validation message.
func (e badRequestError) Error() string { return e.message }

// errBadRequest creates a validation error.
func errBadRequest(message string) error { return badRequestError{message: message} }
