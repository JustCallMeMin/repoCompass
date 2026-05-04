package repository

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"net/url"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"time"

	"github.com/JustCallMeMin/repoCompass/backend/internal/rcerr"
)

// LocalRepositoryProvider resolves repositories from local filesystem paths.
type LocalRepositoryProvider struct {
	now func() time.Time
}

// NewLocalRepositoryProvider creates a provider for local filesystem sources.
func NewLocalRepositoryProvider() LocalRepositoryProvider {
	return LocalRepositoryProvider{
		now: time.Now,
	}
}

// SourceType returns the source type supported by the local provider.
func (p LocalRepositoryProvider) SourceType() SourceType {
	return SourceTypeLocal
}

// Resolve turns a local filesystem path into repository metadata.
func (p LocalRepositoryProvider) Resolve(ctx context.Context, source RepositorySource) (RepositoryResolution, error) {
	if source.Type != "" && source.Type != SourceTypeLocal {
		return RepositoryResolution{}, rcerr.New(rcerr.CodeInvalidSource,
			fmt.Sprintf("local repository provider does not support source type %q", source.Type), nil)
	}
	if strings.TrimSpace(source.Path) == "" {
		return RepositoryResolution{}, rcerr.New(rcerr.CodeInvalidSource,
			"local repository path cannot be empty", nil)
	}

	absolutePath, err := filepath.Abs(source.Path)
	if err != nil {
		return RepositoryResolution{}, rcerr.New(rcerr.CodeRepoResolveFailed,
			"failed to resolve absolute path", err)
	}

	info, err := os.Stat(absolutePath)
	if err != nil {
		return RepositoryResolution{}, rcerr.New(rcerr.CodeInvalidSource,
			fmt.Sprintf("local repository path not found: %s", absolutePath), err)
	}
	if !info.IsDir() {
		return RepositoryResolution{}, rcerr.New(rcerr.CodeInvalidSource,
			fmt.Sprintf("local repository path must be a directory: %s", absolutePath), nil)
	}

	defaultBranch := readDefaultBranch(ctx, absolutePath)
	metadata := map[string]string{
		"local_path":  absolutePath,
		"source_type": string(SourceTypeLocal),
		"resolved_at": p.clock().UTC().Format(time.RFC3339),
	}
	if defaultBranch != "" {
		metadata["default_branch"] = defaultBranch
	}

	repositoryName := filepath.Base(absolutePath)
	return RepositoryResolution{
		Repository: Repository{
			ID:            localRepositoryID(absolutePath),
			Name:          repositoryName,
			FullName:      repositoryName,
			URL:           (&url.URL{Scheme: "file", Path: absolutePath}).String(),
			LocalPath:     absolutePath,
			Provider:      ProviderLocal,
			DefaultBranch: defaultBranch,
			Status:        StatusActive,
		},
		SnapshotMetadata: metadata,
	}, nil
}

func (p LocalRepositoryProvider) clock() time.Time {
	if p.now == nil {
		return time.Now()
	}
	return p.now()
}

func localRepositoryID(absolutePath string) string {
	sum := sha256.Sum256([]byte(absolutePath))
	return "local_" + hex.EncodeToString(sum[:])[:16]
}

func readDefaultBranch(ctx context.Context, absolutePath string) string {
	cmd := exec.CommandContext(ctx, "git", "-C", absolutePath, "branch", "--show-current")
	output, err := cmd.Output()
	if err != nil {
		return ""
	}
	return strings.TrimSpace(string(output))
}
