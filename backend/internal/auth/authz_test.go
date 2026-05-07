package auth_test

import (
	"context"
	"testing"

	"github.com/JustCallMeMin/repoCompass/backend/internal/auth"
	"github.com/JustCallMeMin/repoCompass/backend/internal/org"
)

type mockMembershipProvider struct {
	memberships []org.Membership
}

func (m *mockMembershipProvider) GetMembershipsByUser(ctx context.Context, userID string) ([]org.Membership, error) {
	var userMems []org.Membership
	for _, mem := range m.memberships {
		if mem.UserID == userID {
			userMems = append(userMems, mem)
		}
	}
	return userMems, nil
}

func TestAuthorizationService_Roles(t *testing.T) {
	orgID := "org1"
	provider := &mockMembershipProvider{
		memberships: []org.Membership{
			{OrganizationID: orgID, UserID: "user_owner", Role: org.RoleOwner},
			{OrganizationID: orgID, UserID: "user_admin", Role: org.RoleAdmin},
			{OrganizationID: orgID, UserID: "user_member", Role: org.RoleMember},
			{OrganizationID: orgID, UserID: "user_viewer", Role: org.RoleViewer},
		},
	}

	authSvc := auth.NewAuthorizationService(provider)
	ctx := context.Background()

	tests := []struct {
		userID      string
		canManage   bool
		canAccess   bool
		canEdit     bool
	}{
		{"user_owner", true, true, true},
		{"user_admin", true, true, true},
		{"user_member", false, true, true},
		{"user_viewer", false, true, false},
		{"user_none", false, false, false},
	}

	for _, tc := range tests {
		t.Run(tc.userID, func(t *testing.T) {
			errManage := authSvc.CheckManageOrg(ctx, tc.userID, orgID)
			if (errManage == nil) != tc.canManage {
				t.Errorf("CheckManageOrg expected %v, got error: %v", tc.canManage, errManage)
			}

			errAccess := authSvc.CheckAccessRepo(ctx, tc.userID, orgID)
			if (errAccess == nil) != tc.canAccess {
				t.Errorf("CheckAccessRepo expected %v, got error: %v", tc.canAccess, errAccess)
			}

			errEdit := authSvc.CheckEditRepo(ctx, tc.userID, orgID)
			if (errEdit == nil) != tc.canEdit {
				t.Errorf("CheckEditRepo expected %v, got error: %v", tc.canEdit, errEdit)
			}
		})
	}
}

func TestStaticRoleChecks(t *testing.T) {
	if !auth.CanManageOrg(org.RoleOwner) || !auth.CanManageOrg(org.RoleAdmin) {
		t.Errorf("Owner and Admin should be able to manage org")
	}
	if auth.CanManageOrg(org.RoleMember) || auth.CanManageOrg(org.RoleViewer) {
		t.Errorf("Member and Viewer should not be able to manage org")
	}

	if !auth.CanAccessRepo(org.RoleOwner) || !auth.CanAccessRepo(org.RoleViewer) {
		t.Errorf("All roles should access repos")
	}

	if !auth.CanEditRepo(org.RoleMember) || auth.CanEditRepo(org.RoleViewer) {
		t.Errorf("Member should edit repo, Viewer should not")
	}
}
