package postgres

import (
	"context"
	"database/sql"
	"errors"
	"fmt"

	"github.com/JustCallMeMin/repoCompass/backend/internal/org"
	"github.com/JustCallMeMin/repoCompass/backend/internal/repository"
)

// SaveOrganization creates or updates an organization
func (s *Store) SaveOrganization(ctx context.Context, o org.Organization) error {
	query := `
		INSERT INTO organizations (id, name, created_at, updated_at)
		VALUES ($1, $2, $3, $4)
		ON CONFLICT (id) DO UPDATE SET
			name = EXCLUDED.name,
			updated_at = EXCLUDED.updated_at
	`
	_, err := s.db.ExecContext(ctx, query, o.ID, o.Name, o.CreatedAt, o.UpdatedAt)
	if err != nil {
		return fmt.Errorf("postgres: save organization: %w", err)
	}
	return nil
}

// GetOrganization returns an organization by ID
func (s *Store) GetOrganization(ctx context.Context, id string) (org.Organization, error) {
	query := `
		SELECT id, name, created_at, updated_at
		FROM organizations
		WHERE id = $1
	`
	var o org.Organization
	err := s.db.QueryRowContext(ctx, query, id).Scan(
		&o.ID, &o.Name, &o.CreatedAt, &o.UpdatedAt,
	)
	if errors.Is(err, sql.ErrNoRows) {
		return org.Organization{}, fmt.Errorf("postgres: organization not found: %s", id)
	}
	if err != nil {
		return org.Organization{}, fmt.Errorf("postgres: get organization: %w", err)
	}
	return o, nil
}

// ListOrganizations returns all organizations
func (s *Store) ListOrganizations(ctx context.Context) ([]org.Organization, error) {
	query := `
		SELECT id, name, created_at, updated_at
		FROM organizations
		ORDER BY name ASC
	`
	rows, err := s.db.QueryContext(ctx, query)
	if err != nil {
		return nil, fmt.Errorf("postgres: list organizations: %w", err)
	}
	defer rows.Close()

	var orgs []org.Organization
	for rows.Next() {
		var o org.Organization
		if err := rows.Scan(&o.ID, &o.Name, &o.CreatedAt, &o.UpdatedAt); err != nil {
			return nil, fmt.Errorf("postgres: scan organization: %w", err)
		}
		orgs = append(orgs, o)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("postgres: rows error: %w", err)
	}
	return orgs, nil
}

// SaveMembership creates or updates a user's membership in an organization
func (s *Store) SaveMembership(ctx context.Context, m org.Membership) error {
	query := `
		INSERT INTO organization_memberships (organization_id, user_id, role, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5)
		ON CONFLICT (organization_id, user_id) DO UPDATE SET
			role = EXCLUDED.role,
			updated_at = EXCLUDED.updated_at
	`
	_, err := s.db.ExecContext(ctx, query, m.OrganizationID, m.UserID, string(m.Role), m.CreatedAt, m.UpdatedAt)
	if err != nil {
		return fmt.Errorf("postgres: save membership: %w", err)
	}
	return nil
}

// DeleteMembership removes a user's membership from an organization.
func (s *Store) DeleteMembership(ctx context.Context, orgID, userID string) error {
	query := `
		DELETE FROM organization_memberships
		WHERE organization_id = $1 AND user_id = $2
	`
	result, err := s.db.ExecContext(ctx, query, orgID, userID)
	if err != nil {
		return fmt.Errorf("postgres: delete membership: %w", err)
	}
	rows, _ := result.RowsAffected()
	if rows == 0 {
		return fmt.Errorf("postgres: membership not found for user %s in org %s", userID, orgID)
	}
	return nil
}

// GetMembershipsByUser returns all memberships for a given user
func (s *Store) GetMembershipsByUser(ctx context.Context, userID string) ([]org.Membership, error) {
	query := `
		SELECT organization_id, user_id, role, created_at, updated_at
		FROM organization_memberships
		WHERE user_id = $1
	`
	rows, err := s.db.QueryContext(ctx, query, userID)
	if err != nil {
		return nil, fmt.Errorf("postgres: get memberships by user: %w", err)
	}
	defer rows.Close()

	var memberships []org.Membership
	for rows.Next() {
		var m org.Membership
		var roleStr string
		if err := rows.Scan(&m.OrganizationID, &m.UserID, &roleStr, &m.CreatedAt, &m.UpdatedAt); err != nil {
			return nil, fmt.Errorf("postgres: scan membership: %w", err)
		}
		m.Role = org.Role(roleStr)
		memberships = append(memberships, m)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("postgres: rows error: %w", err)
	}
	return memberships, nil
}

// GetMembershipsByOrg returns all memberships for a given organization
func (s *Store) GetMembershipsByOrg(ctx context.Context, orgID string) ([]org.Membership, error) {
	query := `
		SELECT organization_id, user_id, role, created_at, updated_at
		FROM organization_memberships
		WHERE organization_id = $1
	`
	rows, err := s.db.QueryContext(ctx, query, orgID)
	if err != nil {
		return nil, fmt.Errorf("postgres: get memberships by org: %w", err)
	}
	defer rows.Close()

	var memberships []org.Membership
	for rows.Next() {
		var m org.Membership
		var roleStr string
		if err := rows.Scan(&m.OrganizationID, &m.UserID, &roleStr, &m.CreatedAt, &m.UpdatedAt); err != nil {
			return nil, fmt.Errorf("postgres: scan membership: %w", err)
		}
		m.Role = org.Role(roleStr)
		memberships = append(memberships, m)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("postgres: rows error: %w", err)
	}
	return memberships, nil
}

// SaveConfiguration creates or updates an organization configuration
func (s *Store) SaveConfiguration(ctx context.Context, c org.Configuration) error {
	query := `
		INSERT INTO organization_configurations (organization_id, settings, updated_at)
		VALUES ($1, $2, $3)
		ON CONFLICT (organization_id) DO UPDATE SET
			settings = EXCLUDED.settings,
			updated_at = EXCLUDED.updated_at
	`
	_, err := s.db.ExecContext(ctx, query, c.OrganizationID, c.Settings, c.UpdatedAt)
	if err != nil {
		return fmt.Errorf("postgres: save configuration: %w", err)
	}
	return nil
}

// GetConfiguration returns the configuration for an organization
func (s *Store) GetConfiguration(ctx context.Context, orgID string) (org.Configuration, error) {
	query := `
		SELECT organization_id, settings, updated_at
		FROM organization_configurations
		WHERE organization_id = $1
	`
	var c org.Configuration
	err := s.db.QueryRowContext(ctx, query, orgID).Scan(
		&c.OrganizationID, &c.Settings, &c.UpdatedAt,
	)
	if errors.Is(err, sql.ErrNoRows) {
		return org.Configuration{}, fmt.Errorf("postgres: configuration not found for org: %s", orgID)
	}
	if err != nil {
		return org.Configuration{}, fmt.Errorf("postgres: get configuration: %w", err)
	}
	return c, nil
}

// SavePolicy creates or updates an organization policy
func (s *Store) SavePolicy(ctx context.Context, p org.Policy) error {
	query := `
		INSERT INTO policies (id, organization_id, name, status, version, rules, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8)
		ON CONFLICT (id) DO UPDATE SET
			name = EXCLUDED.name,
			status = EXCLUDED.status,
			version = policies.version + 1,
			rules = EXCLUDED.rules,
			updated_at = EXCLUDED.updated_at
	`
	if p.Status == "" {
		p.Status = "active"
	}
	if p.Version <= 0 {
		p.Version = 1
	}
	_, err := s.db.ExecContext(ctx, query, p.ID, p.OrganizationID, p.Name, p.Status, p.Version, p.Rules, p.CreatedAt, p.UpdatedAt)
	if err != nil {
		return fmt.Errorf("postgres: save policy: %w", err)
	}
	return nil
}

// GetPolicy returns a policy by ID
func (s *Store) GetPolicy(ctx context.Context, id string) (org.Policy, error) {
	query := `
		SELECT id, organization_id, name, status, version, rules, created_at, updated_at
		FROM policies
		WHERE id = $1
	`
	var p org.Policy
	err := s.db.QueryRowContext(ctx, query, id).Scan(
		&p.ID, &p.OrganizationID, &p.Name, &p.Status, &p.Version, &p.Rules, &p.CreatedAt, &p.UpdatedAt,
	)
	if errors.Is(err, sql.ErrNoRows) {
		return org.Policy{}, fmt.Errorf("postgres: policy not found: %s", id)
	}
	if err != nil {
		return org.Policy{}, fmt.Errorf("postgres: get policy: %w", err)
	}
	return p, nil
}

// ListPoliciesByOrg returns all policies for a given organization
func (s *Store) ListPoliciesByOrg(ctx context.Context, orgID string) ([]org.Policy, error) {
	query := `
		SELECT id, organization_id, name, status, version, rules, created_at, updated_at
		FROM policies
		WHERE organization_id = $1
		ORDER BY name ASC, version DESC
	`
	rows, err := s.db.QueryContext(ctx, query, orgID)
	if err != nil {
		return nil, fmt.Errorf("postgres: list policies: %w", err)
	}
	defer rows.Close()

	var policies []org.Policy
	for rows.Next() {
		var p org.Policy
		if err := rows.Scan(&p.ID, &p.OrganizationID, &p.Name, &p.Status, &p.Version, &p.Rules, &p.CreatedAt, &p.UpdatedAt); err != nil {
			return nil, fmt.Errorf("postgres: scan policy: %w", err)
		}
		policies = append(policies, p)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("postgres: rows error: %w", err)
	}
	return policies, nil
}

// GetRepository returns a repository by ID
func (s *Store) GetRepository(ctx context.Context, id string) (repository.Repository, error) {
	query := `
		SELECT id, name, owner_name, full_name, url, local_path, provider, default_branch, primary_ecosystem, is_monorepo, status, organization_id
		FROM repositories
		WHERE id = $1
	`
	var r repository.Repository
	var providerStr, statusStr string
	err := s.db.QueryRowContext(ctx, query, id).Scan(
		&r.ID, &r.Name, &r.OwnerName, &r.FullName, &r.URL, &r.LocalPath, &providerStr, &r.DefaultBranch, &r.PrimaryEcosystem, &r.IsMonorepo, &statusStr, &r.OrganizationID,
	)
	if errors.Is(err, sql.ErrNoRows) {
		return repository.Repository{}, fmt.Errorf("postgres: repository not found: %s", id)
	}
	if err != nil {
		return repository.Repository{}, fmt.Errorf("postgres: get repository: %w", err)
	}
	r.Provider = repository.Provider(providerStr)
	r.Status = repository.Status(statusStr)
	return r, nil
}

// ListRepositoriesByOrg returns all repositories for a given organization
func (s *Store) ListRepositoriesByOrg(ctx context.Context, orgID string) ([]repository.Repository, error) {
	query := `
		SELECT id, name, owner_name, full_name, url, local_path, provider, default_branch, primary_ecosystem, is_monorepo, status, organization_id
		FROM repositories
		WHERE organization_id = $1
		ORDER BY name ASC
	`
	rows, err := s.db.QueryContext(ctx, query, orgID)
	if err != nil {
		return nil, fmt.Errorf("postgres: list repositories by org: %w", err)
	}
	defer rows.Close()

	var repos []repository.Repository
	for rows.Next() {
		var r repository.Repository
		var providerStr, statusStr string
		if err := rows.Scan(&r.ID, &r.Name, &r.OwnerName, &r.FullName, &r.URL, &r.LocalPath, &providerStr, &r.DefaultBranch, &r.PrimaryEcosystem, &r.IsMonorepo, &statusStr, &r.OrganizationID); err != nil {
			return nil, fmt.Errorf("postgres: scan repository: %w", err)
		}
		r.Provider = repository.Provider(providerStr)
		r.Status = repository.Status(statusStr)
		repos = append(repos, r)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("postgres: rows error: %w", err)
	}
	return repos, nil
}
