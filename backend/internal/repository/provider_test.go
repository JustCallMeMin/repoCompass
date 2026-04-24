package repository

import (
	"context"
	"errors"
	"testing"
)

func TestProviderRegistrySelectsProviderBySourceType(t *testing.T) {
	localProvider := fakeProvider{sourceType: SourceTypeLocal}
	registry, err := NewProviderRegistry(localProvider)
	if err != nil {
		t.Fatalf("expected registry creation to succeed: %v", err)
	}

	provider, err := registry.ProviderFor(SourceTypeLocal)
	if err != nil {
		t.Fatalf("expected provider lookup to succeed: %v", err)
	}

	if provider != localProvider {
		t.Fatalf("unexpected provider: got %#v want %#v", provider, localProvider)
	}
}

func TestProviderRegistryRejectsDuplicateSourceType(t *testing.T) {
	_, err := NewProviderRegistry(
		fakeProvider{sourceType: SourceTypeLocal},
		fakeProvider{sourceType: SourceTypeLocal},
	)
	if err == nil {
		t.Fatal("expected duplicate source type to fail")
	}
}

func TestProviderRegistryReturnsErrorForMissingProvider(t *testing.T) {
	registry, err := NewProviderRegistry(fakeProvider{sourceType: SourceTypeLocal})
	if err != nil {
		t.Fatalf("expected registry creation to succeed: %v", err)
	}

	_, err = registry.ProviderFor(SourceType("github"))
	if err == nil {
		t.Fatal("expected missing provider lookup to fail")
	}
}

type fakeProvider struct {
	sourceType SourceType
}

func (p fakeProvider) SourceType() SourceType {
	return p.sourceType
}

func (p fakeProvider) Resolve(context.Context, RepositorySource) (RepositoryResolution, error) {
	return RepositoryResolution{}, errors.New("fake provider should not be called by registry tests")
}
