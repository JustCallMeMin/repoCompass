// Package audit defines structured audit events for organization actions.
// All mutations to org, membership, and policy resources MUST produce an audit event.
package audit

import "time"

// EventType identifies a category of auditable action.
type EventType string

const (
	// Organization management
	EventOrgCreated EventType = "org.created"
	EventOrgUpdated EventType = "org.updated"

	// Membership management
	EventMemberAdded       EventType = "org.member.added"
	EventMemberRemoved     EventType = "org.member.removed"
	EventMemberRoleChanged EventType = "org.member.role_changed"

	// Policy management
	EventPolicyCreated EventType = "org.policy.created"
	EventPolicyUpdated EventType = "org.policy.updated"
	EventPolicyDeleted EventType = "org.policy.deleted"

	// Access control
	EventAccessDenied EventType = "auth.access_denied"
)

// Entry is a single auditable event record.
type Entry struct {
	ID             string         `json:"id"`
	Type           EventType      `json:"type"`
	OrganizationID string         `json:"organization_id"`
	ActorID        string         `json:"actor_id"`
	ResourceID     string         `json:"resource_id"`
	ResourceType   string         `json:"resource_type"` // "org", "membership", "policy"
	Details        map[string]any `json:"details,omitempty"`
	OccurredAt     time.Time      `json:"occurred_at"`
}

// Logger records audit entries.
type Logger interface {
	Log(entry Entry) error
}

// Store persists audit entries for later review.
type Store interface {
	SaveAuditEntry(entry Entry) error
}

// NoopLogger silently discards every audit entry.
// Use in tests and non-production environments.
type NoopLogger struct{}

func (n NoopLogger) Log(_ Entry) error { return nil }

// StructuredLogger emits audit entries via slog for structured log aggregation.
type StructuredLogger struct {
	logFn func(entry Entry)
}

// NewStructuredLogger creates an audit logger that calls logFn for each entry.
func NewStructuredLogger(logFn func(entry Entry)) *StructuredLogger {
	return &StructuredLogger{logFn: logFn}
}

func (l *StructuredLogger) Log(entry Entry) error {
	l.logFn(entry)
	return nil
}
