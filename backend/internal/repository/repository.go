// Package repository contains repository source and repository identity models.
package repository

// SourceType identifies how RepoCompass received a repository input.
type SourceType string

const (
	// SourceTypeLocal represents a repository available on the local filesystem.
	SourceTypeLocal SourceType = "local"
)

// RepositorySource describes the user-provided input before it is resolved into
// repository metadata.
type RepositorySource struct {
	Type SourceType
	Path string
	URL  string
}

// RepositoryResolution contains repository metadata plus snapshot seed metadata.
type RepositoryResolution struct {
	Repository       Repository
	SnapshotMetadata map[string]string
}

// Provider identifies the platform or source family backing a repository.
type Provider string

const (
	// ProviderLocal represents repositories resolved from local filesystem paths.
	ProviderLocal Provider = "local"
)

// Status describes whether a repository can currently be scanned.
type Status string

const (
	StatusActive   Status = "active"
	StatusArchived Status = "archived"
	StatusDisabled Status = "disabled"
)

// Repository is the resolved repository identity used by snapshots and scans.
type Repository struct {
	ID               string
	Name             string
	OwnerName        string
	FullName         string
	URL              string
	LocalPath        string
	Provider         Provider
	DefaultBranch    string
	PrimaryEcosystem string
	IsMonorepo       bool
	Status           Status
	OrganizationID   string
}
