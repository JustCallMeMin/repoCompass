# Renderer Contract

A Renderer is responsible for converting the structured `RenderRequest` (which contains Findings, Scan Metadata, and Configuration) into a final output format (like Markdown, JSON, or HTML).

## Core Concepts

Renderers must implement the `report.Renderer` interface defined in `backend/internal/report/report.go`:

```go
type Renderer interface {
	Format() Format
	Render(ctx context.Context, request RenderRequest) (RenderResult, error)
}
```

## Adding a New Renderer

1. Create a new struct that implements the `report.Renderer` interface.
2. Register the renderer with the `report.Registry` (if applicable).
3. Ensure that your output is deterministic.
4. Add a "Golden Test" for your renderer to catch regressions when the output changes.

## Minimal Example

Check out the runnable example Renderer in `backend/internal/report/example/renderer.go`.
