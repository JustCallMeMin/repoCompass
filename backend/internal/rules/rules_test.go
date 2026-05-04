package rules_test

import (
	"testing"

	"github.com/JustCallMeMin/repoCompass/backend/internal/config"
	"github.com/JustCallMeMin/repoCompass/backend/internal/rules"
)

func TestRuleValidateAcceptsValidRule(t *testing.T) {
	rule := validRule("readme.exists")

	if err := rule.Validate(); err != nil {
		t.Fatalf("expected valid rule, got error: %v", err)
	}
}

func TestRuleValidateRejectsMissingRequiredFields(t *testing.T) {
	tests := []struct {
		name string
		rule rules.Rule
	}{
		{
			name: "missing id",
			rule: rules.Rule{
				Title:       "README exists",
				Description: "Repository should include README.",
				Severity:    rules.SeverityHigh,
				Category:    rules.CategoryDocumentation,
			},
		},
		{
			name: "missing title",
			rule: rules.Rule{
				ID:          "readme.exists",
				Description: "Repository should include README.",
				Severity:    rules.SeverityHigh,
				Category:    rules.CategoryDocumentation,
			},
		},
		{
			name: "missing description",
			rule: rules.Rule{
				ID:       "readme.exists",
				Title:    "README exists",
				Severity: rules.SeverityHigh,
				Category: rules.CategoryDocumentation,
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if err := tt.rule.Validate(); err == nil {
				t.Fatal("expected validation error")
			}
		})
	}
}

func TestRuleValidateRejectsInvalidSeverityAndCategory(t *testing.T) {
	tests := []struct {
		name string
		rule rules.Rule
	}{
		{
			name: "invalid severity",
			rule: rules.Rule{
				ID:          "readme.exists",
				Title:       "README exists",
				Description: "Repository should include README.",
				Severity:    rules.Severity("critical"),
				Category:    rules.CategoryDocumentation,
			},
		},
		{
			name: "invalid category",
			rule: rules.Rule{
				ID:          "readme.exists",
				Title:       "README exists",
				Description: "Repository should include README.",
				Severity:    rules.SeverityHigh,
				Category:    rules.Category("security"),
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if err := tt.rule.Validate(); err == nil {
				t.Fatal("expected validation error")
			}
		})
	}
}

func TestRuleSetValidateAcceptsValidRuleSet(t *testing.T) {
	ruleSet := rules.RuleSet{
		ID:      "builtin.default",
		Name:    "Built-in Default RuleSet",
		Version: "0.1.0",
		Rules: []rules.Rule{
			validRule("readme.exists"),
			validRule("contributing.exists"),
		},
	}

	if err := ruleSet.Validate(); err != nil {
		t.Fatalf("expected valid ruleset, got error: %v", err)
	}
}

func TestRuleSetValidateRejectsMissingRequiredFields(t *testing.T) {
	tests := []struct {
		name    string
		ruleSet rules.RuleSet
	}{
		{
			name: "missing id",
			ruleSet: rules.RuleSet{
				Name:    "Built-in Default RuleSet",
				Version: "0.1.0",
			},
		},
		{
			name: "missing name",
			ruleSet: rules.RuleSet{
				ID:      "builtin.default",
				Version: "0.1.0",
			},
		},
		{
			name: "missing version",
			ruleSet: rules.RuleSet{
				ID:   "builtin.default",
				Name: "Built-in Default RuleSet",
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if err := tt.ruleSet.Validate(); err == nil {
				t.Fatal("expected validation error")
			}
		})
	}
}

func TestRuleSetValidateRejectsDuplicateRuleID(t *testing.T) {
	ruleSet := rules.RuleSet{
		ID:      "builtin.default",
		Name:    "Built-in Default RuleSet",
		Version: "0.1.0",
		Rules: []rules.Rule{
			validRule("readme.exists"),
			validRule("readme.exists"),
		},
	}

	if err := ruleSet.Validate(); err == nil {
		t.Fatal("expected duplicate rule ID validation error")
	}
}

func TestDefaultRuleSetIsStableAndValid(t *testing.T) {
	ruleSet := rules.DefaultRuleSet()

	if ruleSet.ID != "builtin.default" {
		t.Fatalf("unexpected default ruleset ID: got %q", ruleSet.ID)
	}
	if ruleSet.Version != "0.1.0" {
		t.Fatalf("unexpected default ruleset version: got %q", ruleSet.Version)
	}
	if err := ruleSet.Validate(); err != nil {
		t.Fatalf("expected default ruleset to validate: %v", err)
	}
}

func TestDefaultRuleSetContainsExpectedRules(t *testing.T) {
	ruleSet := rules.DefaultRuleSet()
	expectedIDs := []string{
		"readme.exists",
		"contributing.exists",
		"contributing.setup_instructions",
		"contributing.test_instructions",
		"ci.workflow.exists",
		"ci.test_job.exists",
		"scripts.dev.exists",
		"scripts.test_command.exists",
		"scripts.fmt_command.exists",
		"scripts.lint_command.exists",
	}

	byID := make(map[string]rules.Rule, len(ruleSet.Rules))
	for _, rule := range ruleSet.Rules {
		byID[rule.ID] = rule
	}

	for _, id := range expectedIDs {
		rule, ok := byID[id]
		if !ok {
			t.Fatalf("expected default rule %q", id)
		}
		if !rule.EnabledByDefault {
			t.Fatalf("expected default rule %q to be enabled", id)
		}
	}
}

func TestResolveRuleSetUsesEnabledByDefaultRules(t *testing.T) {
	base := rules.RuleSet{
		ID:      "custom",
		Name:    "Custom",
		Version: "0.1.0",
		Rules: []rules.Rule{
			validRule("enabled"),
			disabledByDefaultRule("disabled"),
		},
	}

	resolved, err := rules.ResolveRuleSet(base, config.EffectiveConfiguration{})
	if err != nil {
		t.Fatalf("expected ruleset resolution to succeed: %v", err)
	}
	if len(resolved.Rules) != 1 {
		t.Fatalf("expected 1 active rule, got %d", len(resolved.Rules))
	}
	if resolved.Rules[0].ID != "enabled" {
		t.Fatalf("expected enabled rule, got %q", resolved.Rules[0].ID)
	}
}

func TestResolveRuleSetAppliesDisabledAndEnabledRules(t *testing.T) {
	base := rules.RuleSet{
		ID:      "custom",
		Name:    "Custom",
		Version: "0.1.0",
		Rules: []rules.Rule{
			validRule("first"),
			validRule("second"),
			disabledByDefaultRule("third"),
		},
	}

	resolved, err := rules.ResolveRuleSet(base, config.EffectiveConfiguration{
		DisabledRules: []string{"first"},
		EnabledRules:  []string{"third"},
	})
	if err != nil {
		t.Fatalf("expected ruleset resolution to succeed: %v", err)
	}

	gotIDs := ruleIDs(resolved.Rules)
	wantIDs := []string{"second", "third"}
	if len(gotIDs) != len(wantIDs) {
		t.Fatalf("expected active rule IDs %v, got %v", wantIDs, gotIDs)
	}
	for i := range wantIDs {
		if gotIDs[i] != wantIDs[i] {
			t.Fatalf("expected active rule IDs %v, got %v", wantIDs, gotIDs)
		}
	}
}

func TestResolveRuleSetRejectsUnknownRuleID(t *testing.T) {
	base := rules.RuleSet{
		ID:      "custom",
		Name:    "Custom",
		Version: "0.1.0",
		Rules:   []rules.Rule{validRule("known")},
	}

	_, err := rules.ResolveRuleSet(base, config.EffectiveConfiguration{DisabledRules: []string{"missing"}})
	if err == nil {
		t.Fatal("expected unknown disabled rule ID to fail")
	}

	_, err = rules.ResolveRuleSet(base, config.EffectiveConfiguration{EnabledRules: []string{"missing"}})
	if err == nil {
		t.Fatal("expected unknown enabled rule ID to fail")
	}
}

func validRule(id string) rules.Rule {
	return rules.Rule{
		ID:               id,
		Title:            "README exists",
		Description:      "Repository should include README.",
		Severity:         rules.SeverityHigh,
		Category:         rules.CategoryDocumentation,
		EnabledByDefault: true,
	}
}

func disabledByDefaultRule(id string) rules.Rule {
	rule := validRule(id)
	rule.EnabledByDefault = false
	return rule
}

func ruleIDs(values []rules.Rule) []string {
	ids := make([]string, 0, len(values))
	for _, rule := range values {
		ids = append(ids, rule.ID)
	}
	return ids
}
