package analyzer

import (
	"fmt"

	"github.com/JustCallMeMin/repoCompass/backend/internal/config"
)

// AnalyzerRegistry stores analyzers by stable ID and returns them in registration order.
type AnalyzerRegistry struct {
	analyzersByID map[string]Analyzer
	order         []string
}

// NewAnalyzerRegistry creates a registry and registers the supplied analyzers.
func NewAnalyzerRegistry(analyzers ...Analyzer) (*AnalyzerRegistry, error) {
	registry := &AnalyzerRegistry{
		analyzersByID: make(map[string]Analyzer),
		order:         make([]string, 0, len(analyzers)),
	}
	for _, analyzer := range analyzers {
		if err := registry.Register(analyzer); err != nil {
			return nil, err
		}
	}
	return registry, nil
}

// Register adds one analyzer to the registry.
func (r *AnalyzerRegistry) Register(analyzer Analyzer) error {
	if r == nil {
		return fmt.Errorf("analyzer registry is nil")
	}
	if analyzer == nil {
		return fmt.Errorf("analyzer is nil")
	}
	metadata := analyzer.Metadata()
	if metadata.ID == "" {
		return fmt.Errorf("analyzer id is empty")
	}
	if _, exists := r.analyzersByID[metadata.ID]; exists {
		return fmt.Errorf("duplicate analyzer id %q", metadata.ID)
	}
	r.analyzersByID[metadata.ID] = analyzer
	r.order = append(r.order, metadata.ID)
	return nil
}

// Resolve returns analyzers selected by the effective configuration.
func (r *AnalyzerRegistry) Resolve(effectiveConfig config.EffectiveConfiguration) ([]Analyzer, error) {
	if r == nil {
		return nil, fmt.Errorf("analyzer registry is nil")
	}

	active := make(map[string]bool, len(r.order))
	for _, id := range r.order {
		if _, exists := r.analyzersByID[id]; !exists {
			return nil, fmt.Errorf("registered analyzer %q is missing", id)
		}
		active[id] = effectiveConfig.EnableDefaultAnalyzers
	}

	for _, id := range effectiveConfig.DisabledAnalyzers {
		if _, exists := r.analyzersByID[id]; !exists {
			return nil, fmt.Errorf("unknown disabled analyzer id %q", id)
		}
		active[id] = false
	}
	for _, id := range effectiveConfig.EnabledAnalyzers {
		if _, exists := r.analyzersByID[id]; !exists {
			return nil, fmt.Errorf("unknown enabled analyzer id %q", id)
		}
		active[id] = true
	}

	resolved := make([]Analyzer, 0, len(r.order))
	for _, id := range r.order {
		if active[id] {
			resolved = append(resolved, r.analyzersByID[id])
		}
	}
	return resolved, nil
}
