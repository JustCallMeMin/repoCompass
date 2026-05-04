package analyzer_test

import (
	"context"
	"testing"

	"github.com/JustCallMeMin/repoCompass/backend/internal/analyzer"
	"github.com/JustCallMeMin/repoCompass/backend/internal/config"
)

func TestAnalyzerRegistryRegistersAndResolvesInOrder(t *testing.T) {
	registry, err := analyzer.NewAnalyzerRegistry(
		namedAnalyzer{id: "first", name: "First"},
		namedAnalyzer{id: "second", name: "Second"},
	)
	if err != nil {
		t.Fatalf("expected registry creation to succeed: %v", err)
	}

	resolved, err := registry.Resolve(config.EffectiveConfiguration{EnableDefaultAnalyzers: true})
	if err != nil {
		t.Fatalf("expected resolve to succeed: %v", err)
	}

	if len(resolved) != 2 {
		t.Fatalf("expected 2 analyzers, got %d", len(resolved))
	}
	if resolved[0].Metadata().ID != "first" {
		t.Fatalf("expected first analyzer ID %q, got %q", "first", resolved[0].Metadata().ID)
	}
	if resolved[1].Metadata().ID != "second" {
		t.Fatalf("expected second analyzer ID %q, got %q", "second", resolved[1].Metadata().ID)
	}
}

func TestAnalyzerRegistryRejectsDuplicateID(t *testing.T) {
	_, err := analyzer.NewAnalyzerRegistry(
		namedAnalyzer{id: "duplicate", name: "Duplicate A"},
		namedAnalyzer{id: "duplicate", name: "Duplicate B"},
	)
	if err == nil {
		t.Fatal("expected duplicate analyzer ID to fail")
	}
}

func TestAnalyzerRegistryRejectsEmptyID(t *testing.T) {
	_, err := analyzer.NewAnalyzerRegistry(namedAnalyzer{name: "No ID"})
	if err == nil {
		t.Fatal("expected empty analyzer ID to fail")
	}
}

func TestAnalyzerRegistryRejectsNilAnalyzer(t *testing.T) {
	_, err := analyzer.NewAnalyzerRegistry(nil)
	if err == nil {
		t.Fatal("expected nil analyzer to fail")
	}
}

func TestAnalyzerRegistryResolveDisabledDefaultAnalyzers(t *testing.T) {
	registry, err := analyzer.NewAnalyzerRegistry(namedAnalyzer{id: "first", name: "First"})
	if err != nil {
		t.Fatalf("expected registry creation to succeed: %v", err)
	}

	resolved, err := registry.Resolve(config.EffectiveConfiguration{EnableDefaultAnalyzers: false})
	if err != nil {
		t.Fatalf("expected resolve to succeed: %v", err)
	}
	if len(resolved) != 0 {
		t.Fatalf("expected no analyzers when defaults disabled, got %d", len(resolved))
	}
}

func TestAnalyzerRegistryResolveDisabledAnalyzer(t *testing.T) {
	registry, err := analyzer.NewAnalyzerRegistry(
		namedAnalyzer{id: "first", name: "First"},
		namedAnalyzer{id: "second", name: "Second"},
	)
	if err != nil {
		t.Fatalf("expected registry creation to succeed: %v", err)
	}

	resolved, err := registry.Resolve(config.EffectiveConfiguration{
		EnableDefaultAnalyzers: true,
		DisabledAnalyzers:      []string{"first"},
	})
	if err != nil {
		t.Fatalf("expected resolve to succeed: %v", err)
	}
	if len(resolved) != 1 {
		t.Fatalf("expected 1 analyzer, got %d", len(resolved))
	}
	if resolved[0].Metadata().ID != "second" {
		t.Fatalf("expected second analyzer to remain, got %q", resolved[0].Metadata().ID)
	}
}

func TestAnalyzerRegistryResolveEnabledAnalyzerWhenDefaultsDisabled(t *testing.T) {
	registry, err := analyzer.NewAnalyzerRegistry(
		namedAnalyzer{id: "first", name: "First"},
		namedAnalyzer{id: "second", name: "Second"},
	)
	if err != nil {
		t.Fatalf("expected registry creation to succeed: %v", err)
	}

	resolved, err := registry.Resolve(config.EffectiveConfiguration{
		EnableDefaultAnalyzers: false,
		EnabledAnalyzers:       []string{"second"},
	})
	if err != nil {
		t.Fatalf("expected resolve to succeed: %v", err)
	}
	if len(resolved) != 1 {
		t.Fatalf("expected 1 analyzer, got %d", len(resolved))
	}
	if resolved[0].Metadata().ID != "second" {
		t.Fatalf("expected second analyzer, got %q", resolved[0].Metadata().ID)
	}
}

func TestAnalyzerRegistryResolveRejectsUnknownAnalyzerID(t *testing.T) {
	registry, err := analyzer.NewAnalyzerRegistry(namedAnalyzer{id: "first", name: "First"})
	if err != nil {
		t.Fatalf("expected registry creation to succeed: %v", err)
	}

	_, err = registry.Resolve(config.EffectiveConfiguration{
		EnableDefaultAnalyzers: true,
		DisabledAnalyzers:      []string{"missing"},
	})
	if err == nil {
		t.Fatal("expected unknown disabled analyzer ID to fail")
	}

	_, err = registry.Resolve(config.EffectiveConfiguration{
		EnableDefaultAnalyzers: false,
		EnabledAnalyzers:       []string{"missing"},
	})
	if err == nil {
		t.Fatal("expected unknown enabled analyzer ID to fail")
	}
}

type namedAnalyzer struct {
	id   string
	name string
}

func (a namedAnalyzer) Metadata() analyzer.AnalyzerMetadata {
	return analyzer.AnalyzerMetadata{
		ID:      a.id,
		Name:    a.name,
		Version: "0.1.0",
	}
}

func (a namedAnalyzer) Analyze(_ context.Context, _ analyzer.Input) (analyzer.AnalyzerResult, error) {
	return analyzer.AnalyzerResult{
		AnalyzerID: a.id,
		Name:       a.name,
		Version:    "0.1.0",
		Status:     analyzer.AnalyzerStatusSuccess,
	}, nil
}
