// Package notification defines the event taxonomy and delivery abstraction
// for RepoCompass organization-scoped notifications.
package notification

import "time"

// EventType identifies a notification event.
type EventType string

const (
	// Organization lifecycle events
	EventOrgCreated       EventType = "org.created"
	EventOrgUpdated       EventType = "org.updated"
	EventMemberAdded      EventType = "org.member.added"
	EventMemberRemoved    EventType = "org.member.removed"
	EventMemberRoleChanged EventType = "org.member.role_changed"

	// Policy events
	EventPolicyCreated EventType = "org.policy.created"
	EventPolicyUpdated EventType = "org.policy.updated"

	// Scan/Assessment events
	EventScanCompleted   EventType = "scan.completed"
	EventScanFailed      EventType = "scan.failed"
	EventScoreDegraded   EventType = "scan.score.degraded"    // score dropped below MinimumScore
	EventScoreImproved   EventType = "scan.score.improved"    // score crossed threshold upward
)

// NotificationChannel identifies how a notification is delivered.
type NotificationChannel string

const (
	ChannelInApp  NotificationChannel = "in_app"
	ChannelEmail  NotificationChannel = "email"
	ChannelSlack  NotificationChannel = "slack"
	ChannelWebhook NotificationChannel = "webhook"
)

// Event is a notification payload produced by business operations.
type Event struct {
	Type           EventType      `json:"type"`
	OrganizationID string         `json:"organization_id"`
	ActorID        string         `json:"actor_id"`
	ResourceID     string         `json:"resource_id"`   // org/repo/policy ID relevant to event
	Metadata       map[string]any `json:"metadata,omitempty"`
	OccurredAt     time.Time      `json:"occurred_at"`
}

// Preference records a user's channel subscriptions per org.
type Preference struct {
	OrganizationID string                `json:"organization_id"`
	UserID         string                `json:"user_id"`
	Channels       []NotificationChannel `json:"channels"`
	EventTypes     []EventType           `json:"event_types"` // empty = all events
	UpdatedAt      time.Time             `json:"updated_at"`
}

// Deliverer is the interface for delivering notification events.
// Implementations may be in-app, email, Slack, or webhook-based.
type Deliverer interface {
	Deliver(event Event) error
}

// NoopDeliverer silently discards every notification.
// It is the default in non-production environments.
type NoopDeliverer struct{}

func (n NoopDeliverer) Deliver(_ Event) error { return nil }
