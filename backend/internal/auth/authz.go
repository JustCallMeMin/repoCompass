package auth

import (
	"context"
	"fmt"

	"github.com/JustCallMeMin/repoCompass/backend/internal/org"
)

// MembershipProvider abstracts the storage layer for getting memberships.
type MembershipProvider interface {
	GetMembershipsByUser(ctx context.Context, userID string) ([]org.Membership, error)
}

// CanManageOrg returns true if the user is an Owner or Admin of the organization.
func CanManageOrg(role org.Role) bool {
	return role == org.RoleOwner || role == org.RoleAdmin
}

// CanAccessRepo returns true if the user has any valid role in the organization.
func CanAccessRepo(role org.Role) bool {
	return role == org.RoleOwner || role == org.RoleAdmin || role == org.RoleMember || role == org.RoleViewer
}

// CanEditRepo returns true if the user is an Owner, Admin, or Member of the organization.
func CanEditRepo(role org.Role) bool {
	return role == org.RoleOwner || role == org.RoleAdmin || role == org.RoleMember
}

// AuthorizationService handles organization-level authorization.
type AuthorizationService struct {
	provider MembershipProvider
}

// NewAuthorizationService creates a new AuthorizationService.
func NewAuthorizationService(p MembershipProvider) *AuthorizationService {
	return &AuthorizationService{provider: p}
}

// GetUserRole returns the role of the user in the specified organization.
func (s *AuthorizationService) GetUserRole(ctx context.Context, userID, orgID string) (org.Role, error) {
	memberships, err := s.provider.GetMembershipsByUser(ctx, userID)
	if err != nil {
		return "", fmt.Errorf("auth: get user role: %w", err)
	}

	for _, m := range memberships {
		if m.OrganizationID == orgID {
			return m.Role, nil
		}
	}

	return "", fmt.Errorf("auth: user %s is not a member of org %s", userID, orgID)
}

// CheckManageOrg checks if a user can manage the organization.
func (s *AuthorizationService) CheckManageOrg(ctx context.Context, userID, orgID string) error {
	role, err := s.GetUserRole(ctx, userID, orgID)
	if err != nil {
		return err
	}
	if !CanManageOrg(role) {
		return fmt.Errorf("auth: user %s cannot manage org %s (requires owner/admin)", userID, orgID)
	}
	return nil
}

// CheckAccessRepo checks if a user can access repositories in the organization.
func (s *AuthorizationService) CheckAccessRepo(ctx context.Context, userID, orgID string) error {
	role, err := s.GetUserRole(ctx, userID, orgID)
	if err != nil {
		return err
	}
	if !CanAccessRepo(role) {
		return fmt.Errorf("auth: user %s cannot access repos in org %s", userID, orgID)
	}
	return nil
}

// CheckEditRepo checks if a user can edit/scan repositories in the organization.
func (s *AuthorizationService) CheckEditRepo(ctx context.Context, userID, orgID string) error {
	role, err := s.GetUserRole(ctx, userID, orgID)
	if err != nil {
		return err
	}
	if !CanEditRepo(role) {
		return fmt.Errorf("auth: user %s cannot edit repos in org %s (requires owner/admin/member)", userID, orgID)
	}
	return nil
}
