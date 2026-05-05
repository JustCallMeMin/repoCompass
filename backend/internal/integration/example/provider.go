package example

import (
	"context"

	"github.com/JustCallMeMin/repoCompass/backend/internal/repository"
)

// ExampleProvider is a minimal example of a repository provider.
type ExampleProvider struct{}

// FetchRepository returns a mocked repository representation.
// A real provider would make an HTTP request to GitHub, GitLab, etc.
func (ExampleProvider) FetchRepository(ctx context.Context, url string) (repository.Repository, error) {
	return repository.Repository{
		ID:            "example-repo-123",
		Name:          "example-repository",
		DefaultBranch: "main",
	}, nil
}
