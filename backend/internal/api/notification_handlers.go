package api

import (
	"net/http"
	"strings"
	"time"

	"github.com/JustCallMeMin/repoCompass/backend/internal/notification"
)

// handleListNotifications lists recent org notifications visible to the actor.
func (s *Server) handleListNotifications(w http.ResponseWriter, r *http.Request) {
	if s.notifications == nil {
		writeRequestError(w, r, http.StatusNotImplemented, "not_implemented", "notification storage not configured")
		return
	}
	orgID := strings.TrimSpace(r.PathValue("organization_id"))
	if !s.authorize(w, r, orgID, "read") {
		return
	}
	items, err := s.notifications.ListNotifications(r.Context(), orgID, s.authenticatedActorID(r), parseLimit(r, 50))
	if err != nil {
		writeRequestError(w, r, http.StatusInternalServerError, "notification_query_failed", "failed to list notifications")
		return
	}
	if items == nil {
		items = []notification.Notification{}
	}
	writeData(w, r, http.StatusOK, items)
}

// handleMarkNotificationRead marks one notification as read by the actor.
func (s *Server) handleMarkNotificationRead(w http.ResponseWriter, r *http.Request) {
	if s.notifications == nil {
		writeRequestError(w, r, http.StatusNotImplemented, "not_implemented", "notification storage not configured")
		return
	}
	orgID := strings.TrimSpace(r.PathValue("organization_id"))
	if !s.authorize(w, r, orgID, "read") {
		return
	}
	id := strings.TrimSpace(r.PathValue("notification_id"))
	if id == "" {
		writeRequestError(w, r, http.StatusBadRequest, "invalid_request", "notification_id is required")
		return
	}
	if err := s.notifications.MarkNotificationRead(r.Context(), id, s.authenticatedActorID(r), time.Now().UTC()); err != nil {
		writeRequestError(w, r, http.StatusNotFound, "not_found", "notification not found")
		return
	}
	writeData(w, r, http.StatusOK, map[string]string{"status": "read"})
}

// handleGetNotificationPreference returns the actor's notification preferences.
func (s *Server) handleGetNotificationPreference(w http.ResponseWriter, r *http.Request) {
	if s.notifications == nil {
		writeRequestError(w, r, http.StatusNotImplemented, "not_implemented", "notification storage not configured")
		return
	}
	orgID := strings.TrimSpace(r.PathValue("organization_id"))
	if !s.authorize(w, r, orgID, "read") {
		return
	}
	pref, err := s.notifications.GetNotificationPreference(r.Context(), orgID, s.authenticatedActorID(r))
	if err != nil {
		writeRequestError(w, r, http.StatusInternalServerError, "notification_preference_query_failed", "failed to fetch notification preference")
		return
	}
	writeData(w, r, http.StatusOK, pref)
}

// handleSaveNotificationPreference updates the actor's notification preferences.
func (s *Server) handleSaveNotificationPreference(w http.ResponseWriter, r *http.Request) {
	if s.notifications == nil {
		writeRequestError(w, r, http.StatusNotImplemented, "not_implemented", "notification storage not configured")
		return
	}
	orgID := strings.TrimSpace(r.PathValue("organization_id"))
	if !s.authorize(w, r, orgID, "read") {
		return
	}
	var pref notification.Preference
	if err := decodeJSON(r, &pref); err != nil {
		writeRequestError(w, r, http.StatusBadRequest, "invalid_json", err.Error())
		return
	}
	pref.OrganizationID = orgID
	pref.UserID = s.authenticatedActorID(r)
	if len(pref.Channels) == 0 {
		pref.Channels = []notification.NotificationChannel{notification.ChannelInApp}
	}
	for _, channel := range pref.Channels {
		if channel != notification.ChannelInApp && channel != notification.ChannelWebhook && channel != notification.ChannelEmail {
			writeRequestError(w, r, http.StatusBadRequest, "invalid_request", "unsupported notification channel")
			return
		}
	}
	pref.UpdatedAt = time.Now().UTC()
	if err := s.notifications.SaveNotificationPreference(r.Context(), pref); err != nil {
		writeRequestError(w, r, http.StatusInternalServerError, "notification_preference_write_failed", "failed to save notification preference")
		return
	}
	writeData(w, r, http.StatusOK, pref)
}
