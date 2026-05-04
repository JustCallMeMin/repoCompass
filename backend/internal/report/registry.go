package report

import "fmt"

// Registry stores renderers by output format.
type Registry struct {
	renderers map[Format]Renderer
}

// NewRegistry creates a registry and registers the supplied renderers.
func NewRegistry(renderers ...Renderer) (*Registry, error) {
	registry := &Registry{renderers: make(map[Format]Renderer, len(renderers))}
	for _, renderer := range renderers {
		if err := registry.Register(renderer); err != nil {
			return nil, err
		}
	}
	return registry, nil
}

// DefaultRegistry creates a registry with built-in Markdown and JSON renderers.
func DefaultRegistry() (*Registry, error) {
	return NewRegistry(NewMarkdownRenderer(), NewJSONRenderer())
}

// Register adds a renderer to the registry.
func (r *Registry) Register(renderer Renderer) error {
	if r == nil {
		return fmt.Errorf("report renderer registry is nil")
	}
	if renderer == nil {
		return fmt.Errorf("report renderer is nil")
	}
	format := renderer.Format()
	if format == "" {
		return fmt.Errorf("report renderer format is empty")
	}
	if _, exists := r.renderers[format]; exists {
		return fmt.Errorf("duplicate report renderer format %q", format)
	}
	r.renderers[format] = renderer
	return nil
}

// RendererFor returns the renderer registered for format.
func (r *Registry) RendererFor(format Format) (Renderer, error) {
	if r == nil {
		return nil, fmt.Errorf("report renderer registry is nil")
	}
	renderer, ok := r.renderers[format]
	if !ok {
		return nil, fmt.Errorf("unsupported report format %q", format)
	}
	return renderer, nil
}
