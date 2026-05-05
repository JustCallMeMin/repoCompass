package example

import (
	"context"

	"github.com/JustCallMeMin/repoCompass/backend/internal/report"
)

// ExampleRenderer is a minimal renderer that outputs plain text.
type ExampleRenderer struct{}

// Format returns the identifier for this renderer's output format.
func (ExampleRenderer) Format() report.Format {
	return report.Format("plaintext")
}

// Render processes the RenderRequest and produces a plaintext RenderResult.
func (ExampleRenderer) Render(ctx context.Context, request report.RenderRequest) (report.RenderResult, error) {
	content := []byte("Scan Complete for Repository: " + request.Repository.Name)

	return report.RenderResult{
		Format:      report.Format("plaintext"),
		Content:     content,
		ContentType: "text/plain",
	}, nil
}
