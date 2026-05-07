package insights_test

import (
	"context"
	"testing"

	"github.com/JustCallMeMin/repoCompass/backend/internal/insights"
)

type mockProvider struct{}

func (m *mockProvider) GetOrganizationInsights(ctx context.Context, orgID string) (insights.OrganizationInsights, error) {
	if orgID == "org_1" {
		return insights.OrganizationInsights{
			OrganizationID:    "org_1",
			AverageScore:      85,
			TotalRepositories: 10,
			TotalScans:        50,
		}, nil
	}
	return insights.OrganizationInsights{}, nil
}

func TestGetOrganizationInsights(t *testing.T) {
	provider := &mockProvider{}
	
	result, err := provider.GetOrganizationInsights(context.Background(), "org_1")
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	
	if result.OrganizationID != "org_1" {
		t.Errorf("expected org_1, got %s", result.OrganizationID)
	}
	if result.AverageScore != 85 {
		t.Errorf("expected 85, got %d", result.AverageScore)
	}
	if result.TotalRepositories != 10 {
		t.Errorf("expected 10, got %d", result.TotalRepositories)
	}
	if result.TotalScans != 50 {
		t.Errorf("expected 50, got %d", result.TotalScans)
	}
}
