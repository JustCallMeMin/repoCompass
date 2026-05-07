package notification_test

import (
	"testing"
	"time"

	"github.com/JustCallMeMin/repoCompass/backend/internal/notification"
)

func TestNoopDelivererNeverErrors(t *testing.T) {
	d := notification.NoopDeliverer{}
	ev := notification.Event{
		Type:           notification.EventScanCompleted,
		OrganizationID: "org_1",
		ActorID:        "user_1",
		ResourceID:     "scan_1",
		OccurredAt:     time.Now(),
	}
	if err := d.Deliver(ev); err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
}

func TestEventTypes(t *testing.T) {
	types := []notification.EventType{
		notification.EventOrgCreated,
		notification.EventMemberAdded,
		notification.EventPolicyCreated,
		notification.EventScanCompleted,
		notification.EventScoreDegraded,
	}
	for _, et := range types {
		if et == "" {
			t.Errorf("event type must not be empty string")
		}
	}
}
