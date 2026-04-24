// Package snapshot contains snapshot creation and snapshot lifecycle primitives.
package snapshot

import "time"

// SourceType identifies how a repository snapshot was captured.
type SourceType string

const (
	// SourceTypeLocal represents a snapshot captured from a local filesystem repository.
	SourceTypeLocal SourceType = "local"
)

// RepositorySnapshot captures the repository state used by one or more scans.
type RepositorySnapshot struct {
	ID               string
	RepositoryID     string
	SourceType       SourceType
	BranchName       string
	CommitSHA        string
	TreeReference    string
	CapturedAt       time.Time
	SnapshotMetadata map[string]string
}
