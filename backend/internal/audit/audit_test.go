package audit_test

import (
	"testing"
	"time"

	"github.com/JustCallMeMin/repoCompass/backend/internal/audit"
)

func TestNoopLoggerNeverErrors(t *testing.T) {
	l := audit.NoopLogger{}
	if err := l.Log(audit.Entry{Type: audit.EventOrgCreated, OccurredAt: time.Now()}); err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
}

func TestStructuredLoggerCallsCallback(t *testing.T) {
	var captured audit.Entry
	l := audit.NewStructuredLogger(func(e audit.Entry) {
		captured = e
	})

	entry := audit.Entry{
		Type:           audit.EventMemberAdded,
		OrganizationID: "org_1",
		ActorID:        "user_1",
		ResourceID:     "user_2",
		ResourceType:   "membership",
		OccurredAt:     time.Now(),
	}
	if err := l.Log(entry); err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if captured.Type != audit.EventMemberAdded {
		t.Errorf("expected EventMemberAdded, got %s", captured.Type)
	}
}

func TestEventTypes(t *testing.T) {
	types := []audit.EventType{
		audit.EventOrgCreated,
		audit.EventMemberAdded,
		audit.EventPolicyCreated,
		audit.EventAccessDenied,
	}
	for _, et := range types {
		if et == "" {
			t.Errorf("event type must not be empty string")
		}
	}
}
