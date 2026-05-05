package example

import (
	"context"
	"testing"
)

func TestExampleProvider_FetchRepository(t *testing.T) {
	p := ExampleProvider{}
	repo, err := p.FetchRepository(context.Background(), "https://example.com/repo")
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if repo.ID != "example-repo-123" {
		t.Errorf("expected ID 'example-repo-123', got %q", repo.ID)
	}
	if repo.Name != "example-repository" {
		t.Errorf("expected Name 'example-repository', got %q", repo.Name)
	}
}
