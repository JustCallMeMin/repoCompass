package repository

import (
	"context"
	"fmt"
)

// RepositoryProvider resolves repository sources into repository metadata.
type RepositoryProvider interface {
	SourceType() SourceType
	Resolve(ctx context.Context, source RepositorySource) (Repository, error)
}

// ProviderRegistry stores repository providers by supported source type.
type ProviderRegistry struct {
	providers map[SourceType]RepositoryProvider
}

// NewProviderRegistry creates a registry from the supplied providers.
func NewProviderRegistry(providers ...RepositoryProvider) (*ProviderRegistry, error) {
	registry := &ProviderRegistry{
		providers: make(map[SourceType]RepositoryProvider, len(providers)),
	}

	for _, provider := range providers {
		if provider == nil {
			return nil, fmt.Errorf("repository provider cannot be nil")
		}

		sourceType := provider.SourceType()
		if sourceType == "" {
			return nil, fmt.Errorf("repository provider source type cannot be empty")
		}
		if _, exists := registry.providers[sourceType]; exists {
			return nil, fmt.Errorf("repository provider already registered for source type %q", sourceType)
		}

		registry.providers[sourceType] = provider
	}

	return registry, nil
}

// ProviderFor returns the provider registered for the source type.
func (r *ProviderRegistry) ProviderFor(sourceType SourceType) (RepositoryProvider, error) {
	if r == nil {
		return nil, fmt.Errorf("repository provider registry is nil")
	}

	provider, ok := r.providers[sourceType]
	if !ok {
		return nil, fmt.Errorf("repository provider not found for source type %q", sourceType)
	}

	return provider, nil
}
