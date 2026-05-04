package report_test

import (
	"context"
	"testing"

	"github.com/JustCallMeMin/repoCompass/backend/internal/report"
)

func TestRegistryResolvesMarkdownAndJSON(t *testing.T) {
	registry, err := report.DefaultRegistry()
	if err != nil {
		t.Fatalf("expected default registry creation to succeed: %v", err)
	}

	for _, format := range []report.Format{report.FormatMarkdown, report.FormatJSON} {
		renderer, err := registry.RendererFor(format)
		if err != nil {
			t.Fatalf("expected renderer for %q: %v", format, err)
		}
		if renderer.Format() != format {
			t.Fatalf("expected renderer format %q, got %q", format, renderer.Format())
		}
	}
}

func TestRegistryRejectsUnknownFormat(t *testing.T) {
	registry, err := report.DefaultRegistry()
	if err != nil {
		t.Fatalf("expected default registry creation to succeed: %v", err)
	}

	_, err = registry.RendererFor(report.Format("xml"))
	if err == nil {
		t.Fatal("expected unknown report format to fail")
	}
}

func TestRegistryRejectsDuplicateFormat(t *testing.T) {
	_, err := report.NewRegistry(registryFakeRenderer{}, registryFakeRenderer{})
	if err == nil {
		t.Fatal("expected duplicate renderer format to fail")
	}
}

func TestRegistryRejectsNilRenderer(t *testing.T) {
	_, err := report.NewRegistry(nil)
	if err == nil {
		t.Fatal("expected nil renderer to fail")
	}
}

type registryFakeRenderer struct{}

func (registryFakeRenderer) Format() report.Format {
	return report.FormatMarkdown
}

func (registryFakeRenderer) Render(_ context.Context, _ report.RenderRequest) (report.RenderResult, error) {
	return report.RenderResult{Format: report.FormatMarkdown}, nil
}
