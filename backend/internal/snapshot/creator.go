package snapshot

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"strings"
	"time"

	"github.com/JustCallMeMin/repoCompass/backend/internal/repository"
)

const (
	metadataDefaultBranch = "default_branch"
	metadataCommitSHA     = "commit_sha"
	metadataTreeReference = "tree_reference"
)

// CreateRequest contains resolved repository data needed to create a snapshot.
type CreateRequest struct {
	Repository       repository.Repository
	SourceType       SourceType
	SnapshotMetadata map[string]string
}

// Creator creates repository snapshots from resolved repository metadata.
type Creator struct {
	now func() time.Time
}

// NewCreator creates a snapshot creator with the default clock.
func NewCreator() Creator {
	return Creator{
		now: time.Now,
	}
}

// Create builds a repository snapshot for later scan execution.
func (c Creator) Create(ctx context.Context, request CreateRequest) (RepositorySnapshot, error) {
	select {
	case <-ctx.Done():
		return RepositorySnapshot{}, ctx.Err()
	default:
	}

	if strings.TrimSpace(request.Repository.ID) == "" {
		return RepositorySnapshot{}, fmt.Errorf("repository ID cannot be empty")
	}
	if request.SourceType == "" {
		return RepositorySnapshot{}, fmt.Errorf("snapshot source type cannot be empty")
	}

	capturedAt := c.clock().UTC()
	metadata := copyMetadata(request.SnapshotMetadata)
	branchName := firstNonEmpty(metadata[metadataDefaultBranch], request.Repository.DefaultBranch)
	commitSHA := metadata[metadataCommitSHA]
	treeReference := metadata[metadataTreeReference]

	return RepositorySnapshot{
		ID:               snapshotID(request.Repository.ID, request.SourceType, branchName, commitSHA, treeReference, capturedAt),
		RepositoryID:     request.Repository.ID,
		SourceType:       request.SourceType,
		BranchName:       branchName,
		CommitSHA:        commitSHA,
		TreeReference:    treeReference,
		CapturedAt:       capturedAt,
		SnapshotMetadata: metadata,
	}, nil
}

func (c Creator) clock() time.Time {
	if c.now == nil {
		return time.Now()
	}
	return c.now()
}

func copyMetadata(metadata map[string]string) map[string]string {
	if metadata == nil {
		return map[string]string{}
	}

	copied := make(map[string]string, len(metadata))
	for key, value := range metadata {
		copied[key] = value
	}
	return copied
}

func firstNonEmpty(values ...string) string {
	for _, value := range values {
		if value != "" {
			return value
		}
	}
	return ""
}

func snapshotID(repositoryID string, sourceType SourceType, branchName, commitSHA, treeReference string, capturedAt time.Time) string {
	parts := []string{
		repositoryID,
		string(sourceType),
		branchName,
		commitSHA,
		treeReference,
		capturedAt.Format(time.RFC3339Nano),
	}
	sum := sha256.Sum256([]byte(strings.Join(parts, "\x00")))
	return "snap_" + hex.EncodeToString(sum[:])[:16]
}
