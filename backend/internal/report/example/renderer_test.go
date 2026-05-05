package example

import (
	"context"
	"strings"
	"testing"

	"github.com/JustCallMeMin/repoCompass/backend/internal/report"
	"github.com/JustCallMeMin/repoCompass/backend/internal/repository"
)

func TestExampleRenderer_Render(t *testing.T) {
	r := ExampleRenderer{}
	if r.Format() != "plaintext" {
		t.Errorf("expected format plaintext, got %q", r.Format())
	}

	req := report.RenderRequest{
		Repository: repository.Repository{Name: "test-repo"},
	}
	res, err := r.Render(context.Background(), req)
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if res.Format != "plaintext" {
		t.Errorf("expected format plaintext, got %q", res.Format)
	}
	output := string(res.Content)
	if !strings.Contains(output, "test-repo") {
		t.Errorf("expected output to contain 'test-repo', got %q", output)
	}
}
