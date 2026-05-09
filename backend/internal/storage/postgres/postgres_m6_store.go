package postgres

import (
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"
	"time"

	"github.com/JustCallMeMin/repoCompass/backend/internal/audit"
	"github.com/JustCallMeMin/repoCompass/backend/internal/notification"
)

// SaveNotificationPreference creates or updates a user's notification preferences.
func (s *Store) SaveNotificationPreference(ctx context.Context, pref notification.Preference) error {
	channels, err := json.Marshal(pref.Channels)
	if err != nil {
		return fmt.Errorf("postgres: marshal notification channels: %w", err)
	}
	eventTypes, err := json.Marshal(pref.EventTypes)
	if err != nil {
		return fmt.Errorf("postgres: marshal notification event types: %w", err)
	}
	if pref.UpdatedAt.IsZero() {
		pref.UpdatedAt = time.Now().UTC()
	}
	_, err = s.db.ExecContext(ctx, `
		INSERT INTO notification_preferences (organization_id, user_id, channels, event_types, webhook_url, email, updated_at)
		VALUES ($1, $2, $3, $4, NULLIF($5, ''), NULLIF($6, ''), $7)
		ON CONFLICT (organization_id, user_id) DO UPDATE SET
			channels = EXCLUDED.channels,
			event_types = EXCLUDED.event_types,
			webhook_url = EXCLUDED.webhook_url,
			email = EXCLUDED.email,
			updated_at = EXCLUDED.updated_at
	`, pref.OrganizationID, pref.UserID, channels, eventTypes, pref.WebhookURL, pref.Email, pref.UpdatedAt)
	if err != nil {
		return fmt.Errorf("postgres: save notification preference: %w", err)
	}
	return nil
}

// GetNotificationPreference returns a user's notification preferences or defaults.
func (s *Store) GetNotificationPreference(ctx context.Context, orgID, userID string) (notification.Preference, error) {
	var pref notification.Preference
	var channelsRaw, eventTypesRaw []byte
	var webhookURL, email sql.NullString
	err := s.db.QueryRowContext(ctx, `
		SELECT organization_id, user_id, channels, event_types, webhook_url, email, updated_at
		FROM notification_preferences
		WHERE organization_id = $1 AND user_id = $2
	`, orgID, userID).Scan(&pref.OrganizationID, &pref.UserID, &channelsRaw, &eventTypesRaw, &webhookURL, &email, &pref.UpdatedAt)
	if errors.Is(err, sql.ErrNoRows) {
		return notification.Preference{
			OrganizationID: orgID,
			UserID:         userID,
			Channels:       []notification.NotificationChannel{notification.ChannelInApp},
			EventTypes:     []notification.EventType{},
			UpdatedAt:      time.Now().UTC(),
		}, nil
	}
	if err != nil {
		return pref, fmt.Errorf("postgres: get notification preference: %w", err)
	}
	if err := json.Unmarshal(channelsRaw, &pref.Channels); err != nil {
		return pref, fmt.Errorf("postgres: unmarshal notification channels: %w", err)
	}
	if err := json.Unmarshal(eventTypesRaw, &pref.EventTypes); err != nil {
		return pref, fmt.Errorf("postgres: unmarshal notification event types: %w", err)
	}
	pref.WebhookURL = webhookURL.String
	pref.Email = email.String
	return pref, nil
}

// SaveNotification persists an in-app notification.
func (s *Store) SaveNotification(ctx context.Context, n notification.Notification) error {
	metadata := []byte("{}")
	if n.Metadata != nil {
		var err error
		metadata, err = json.Marshal(n.Metadata)
		if err != nil {
			return fmt.Errorf("postgres: marshal notification metadata: %w", err)
		}
	}
	_, err := s.db.ExecContext(ctx, `
		INSERT INTO notifications (id, organization_id, user_id, type, severity, title, message, resource_type, resource_id, metadata, read_at, created_at)
		VALUES ($1, $2, NULLIF($3, ''), $4, $5, $6, $7, NULLIF($8, ''), NULLIF($9, ''), $10, $11, $12)
	`, n.ID, n.OrganizationID, n.UserID, string(n.Type), n.Severity, n.Title, n.Message, n.ResourceType, n.ResourceID, metadata, n.ReadAt, n.CreatedAt)
	if err != nil {
		return fmt.Errorf("postgres: save notification: %w", err)
	}
	return nil
}

// ListNotifications returns recent notifications for an org/user pair.
func (s *Store) ListNotifications(ctx context.Context, orgID, userID string, limit int) ([]notification.Notification, error) {
	if limit <= 0 || limit > 100 {
		limit = 50
	}
	rows, err := s.db.QueryContext(ctx, `
		SELECT id, organization_id, COALESCE(user_id, ''), type, severity, title, message,
		       COALESCE(resource_type, ''), COALESCE(resource_id, ''), metadata, read_at, created_at
		FROM notifications
		WHERE organization_id = $1 AND (user_id IS NULL OR user_id = $2)
		ORDER BY created_at DESC
		LIMIT $3
	`, orgID, userID, limit)
	if err != nil {
		return nil, fmt.Errorf("postgres: list notifications: %w", err)
	}
	defer rows.Close()
	var items []notification.Notification
	for rows.Next() {
		var n notification.Notification
		var typ string
		var metadataRaw []byte
		if err := rows.Scan(&n.ID, &n.OrganizationID, &n.UserID, &typ, &n.Severity, &n.Title, &n.Message, &n.ResourceType, &n.ResourceID, &metadataRaw, &n.ReadAt, &n.CreatedAt); err != nil {
			return nil, fmt.Errorf("postgres: scan notification: %w", err)
		}
		n.Type = notification.EventType(typ)
		_ = json.Unmarshal(metadataRaw, &n.Metadata)
		items = append(items, n)
	}
	return items, rows.Err()
}

// MarkNotificationRead records a read timestamp for one notification.
func (s *Store) MarkNotificationRead(ctx context.Context, id, userID string, readAt time.Time) error {
	result, err := s.db.ExecContext(ctx, `
		UPDATE notifications SET read_at = $3
		WHERE id = $1 AND (user_id IS NULL OR user_id = $2)
	`, id, userID, readAt)
	if err != nil {
		return fmt.Errorf("postgres: mark notification read: %w", err)
	}
	rows, _ := result.RowsAffected()
	if rows == 0 {
		return fmt.Errorf("postgres: notification not found: %s", id)
	}
	return nil
}

// SaveNotificationDelivery records a notification delivery attempt.
func (s *Store) SaveNotificationDelivery(ctx context.Context, d notification.Delivery) error {
	_, err := s.db.ExecContext(ctx, `
		INSERT INTO notification_deliveries (id, notification_id, channel, status, target, error_message, attempted_at)
		VALUES ($1, $2, $3, $4, NULLIF($5, ''), NULLIF($6, ''), $7)
	`, d.ID, d.NotificationID, string(d.Channel), d.Status, d.Target, d.ErrorMessage, d.AttemptedAt)
	if err != nil {
		return fmt.Errorf("postgres: save notification delivery: %w", err)
	}
	return nil
}

// SaveAuditEntry persists an audit event.
func (s *Store) SaveAuditEntry(ctx context.Context, entry audit.Entry) error {
	details := []byte("{}")
	if entry.Details != nil {
		var err error
		details, err = json.Marshal(entry.Details)
		if err != nil {
			return fmt.Errorf("postgres: marshal audit details: %w", err)
		}
	}
	_, err := s.db.ExecContext(ctx, `
		INSERT INTO audit_events (id, type, organization_id, actor_id, resource_type, resource_id, details, occurred_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8)
	`, entry.ID, string(entry.Type), entry.OrganizationID, entry.ActorID, entry.ResourceType, entry.ResourceID, details, entry.OccurredAt)
	if err != nil {
		return fmt.Errorf("postgres: save audit entry: %w", err)
	}
	return nil
}
