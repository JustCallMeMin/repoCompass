package postgres_test

import (
	"context"
	"testing"
	"time"

	"github.com/JustCallMeMin/repoCompass/backend/internal/org"
)

func TestPostgresStore_OrganizationCRUD(t *testing.T) {
	store := openDB(t)
	ctx := context.Background()
	orgID := testID("org-crud")

	// 1. Create Organization
	organization := org.Organization{
		ID:        orgID,
		Name:      "Test Org",
		CreatedAt: time.Now().UTC().Truncate(time.Second),
		UpdatedAt: time.Now().UTC().Truncate(time.Second),
	}
	if err := store.SaveOrganization(ctx, organization); err != nil {
		t.Fatalf("SaveOrganization: %v", err)
	}

	// 2. Get Organization
	o, err := store.GetOrganization(ctx, orgID)
	if err != nil {
		t.Fatalf("GetOrganization: %v", err)
	}
	if o.Name != "Test Org" {
		t.Errorf("expected name 'Test Org', got %q", o.Name)
	}

	// 3. Update Organization
	organization.Name = "Updated Org"
	if err := store.SaveOrganization(ctx, organization); err != nil {
		t.Fatalf("SaveOrganization (update): %v", err)
	}

	o, err = store.GetOrganization(ctx, orgID)
	if err != nil {
		t.Fatalf("GetOrganization after update: %v", err)
	}
	if o.Name != "Updated Org" {
		t.Errorf("expected updated name 'Updated Org', got %q", o.Name)
	}
}

func TestPostgresStore_MembershipCRUD(t *testing.T) {
	store := openDB(t)
	ctx := context.Background()
	orgID := testID("org-mem")
	userID := testID("user")

	// Setup org
	if err := store.SaveOrganization(ctx, org.Organization{
		ID:        orgID,
		Name:      "Membership Org",
		CreatedAt: time.Now().UTC().Truncate(time.Second),
		UpdatedAt: time.Now().UTC().Truncate(time.Second),
	}); err != nil {
		t.Fatalf("SaveOrganization: %v", err)
	}

	mem := org.Membership{
		OrganizationID: orgID,
		UserID:         userID,
		Role:           org.RoleAdmin,
		CreatedAt:      time.Now().UTC().Truncate(time.Second),
		UpdatedAt:      time.Now().UTC().Truncate(time.Second),
	}

	// 1. Create Membership
	if err := store.SaveMembership(ctx, mem); err != nil {
		t.Fatalf("SaveMembership: %v", err)
	}

	// 2. Get Membership
	memberships, err := store.GetMembershipsByUser(ctx, userID)
	if err != nil {
		t.Fatalf("GetMembershipsByUser: %v", err)
	}
	if len(memberships) != 1 || memberships[0].Role != org.RoleAdmin {
		t.Errorf("expected 1 membership with admin role, got %v", memberships)
	}

	// 3. Update Membership
	mem.Role = org.RoleOwner
	if err := store.SaveMembership(ctx, mem); err != nil {
		t.Fatalf("SaveMembership (update): %v", err)
	}

	memberships, err = store.GetMembershipsByOrg(ctx, orgID)
	if err != nil {
		t.Fatalf("GetMembershipsByOrg: %v", err)
	}
	if len(memberships) != 1 || memberships[0].Role != org.RoleOwner {
		t.Errorf("expected 1 membership with owner role, got %v", memberships)
	}
}
