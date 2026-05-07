package org

import (
	"encoding/json"
	"time"
)

// Role represents a user's role in an organization
type Role string

const (
	RoleOwner  Role = "owner"
	RoleAdmin  Role = "admin"
	RoleMember Role = "member"
	RoleViewer Role = "viewer"
)

// Organization represents a multi-tenant isolation boundary
type Organization struct {
	ID        string    `json:"id"`
	Name      string    `json:"name"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

// Membership maps a user to an organization with a specific role
type Membership struct {
	OrganizationID string    `json:"organization_id"`
	UserID         string    `json:"user_id"`
	Role           Role      `json:"role"`
	CreatedAt      time.Time `json:"created_at"`
	UpdatedAt      time.Time `json:"updated_at"`
}

// Configuration represents organization-level settings
type Configuration struct {
	OrganizationID string          `json:"organization_id"`
	Settings       json.RawMessage `json:"settings"`
	UpdatedAt      time.Time       `json:"updated_at"`
}

// Policy represents a rule or constraint applied to repositories in an organization
type Policy struct {
	ID             string          `json:"id"`
	OrganizationID string          `json:"organization_id"`
	Name           string          `json:"name"`
	Rules          json.RawMessage `json:"rules"`
	CreatedAt      time.Time       `json:"created_at"`
	UpdatedAt      time.Time       `json:"updated_at"`
}

const DefaultPersonalOrgID = "00000000-0000-0000-0000-000000000000"
