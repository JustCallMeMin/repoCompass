// Package rules contains rule evaluation building blocks for repository assessment.
package rules

import (
	"fmt"

	"github.com/JustCallMeMin/repoCompass/backend/internal/config"
)

// Severity describes the user impact level of a rule finding.
type Severity string

const (
	SeverityLow    Severity = "low"
	SeverityMedium Severity = "medium"
	SeverityHigh   Severity = "high"
)

// Category groups rules by repository quality area.
type Category string

const (
	CategoryDocumentation   Category = "documentation"
	CategoryWorkflow        Category = "workflow"
	CategoryCI              Category = "ci"
	CategoryMaintainability Category = "maintainability"
)

// Rule defines one deterministic repository check.
type Rule struct {
	ID               string
	Title            string
	Description      string
	Severity         Severity
	Category         Category
	EnabledByDefault bool
}

// Validate checks whether a rule has the required stable fields.
func (r Rule) Validate() error {
	if r.ID == "" {
		return fmt.Errorf("rule id is empty")
	}
	if r.Title == "" {
		return fmt.Errorf("rule title is empty")
	}
	if r.Description == "" {
		return fmt.Errorf("rule description is empty")
	}
	if !r.Severity.Valid() {
		return fmt.Errorf("rule severity %q is invalid", r.Severity)
	}
	if !r.Category.Valid() {
		return fmt.Errorf("rule category %q is invalid", r.Category)
	}
	return nil
}

// RuleSet groups versioned rules for analyzer selection and reporting.
type RuleSet struct {
	ID      string
	Name    string
	Version string
	Rules   []Rule
}

// Validate checks whether a ruleset has required fields and unique rule IDs.
func (r RuleSet) Validate() error {
	if r.ID == "" {
		return fmt.Errorf("ruleset id is empty")
	}
	if r.Name == "" {
		return fmt.Errorf("ruleset name is empty")
	}
	if r.Version == "" {
		return fmt.Errorf("ruleset version is empty")
	}

	seen := make(map[string]struct{}, len(r.Rules))
	for _, rule := range r.Rules {
		if err := rule.Validate(); err != nil {
			return err
		}
		if _, exists := seen[rule.ID]; exists {
			return fmt.Errorf("duplicate rule id %q", rule.ID)
		}
		seen[rule.ID] = struct{}{}
	}
	return nil
}

// Valid reports whether severity is supported by the core taxonomy.
func (s Severity) Valid() bool {
	switch s {
	case SeverityLow, SeverityMedium, SeverityHigh:
		return true
	default:
		return false
	}
}

// Valid reports whether category is supported by the core taxonomy.
func (c Category) Valid() bool {
	switch c {
	case CategoryDocumentation, CategoryWorkflow, CategoryCI, CategoryMaintainability:
		return true
	default:
		return false
	}
}

// DefaultRuleSet returns the initial built-in ruleset for Milestone 2 analyzers.
func DefaultRuleSet() RuleSet {
	return RuleSet{
		ID:      "builtin.default",
		Name:    "Built-in Default RuleSet",
		Version: "0.1.0",
		Rules: []Rule{
			{
				ID:               "readme.exists",
				Title:            "README file exists",
				Description:      "Repository should include a root README file for contributor orientation.",
				Severity:         SeverityHigh,
				Category:         CategoryDocumentation,
				EnabledByDefault: true,
			},
			{
				ID:               "contributing.exists",
				Title:            "CONTRIBUTING file exists",
				Description:      "Repository should include contribution guidance for new contributors.",
				Severity:         SeverityMedium,
				Category:         CategoryDocumentation,
				EnabledByDefault: true,
			},
			{
				ID:               "contributing.setup_instructions",
				Title:            "CONTRIBUTING setup instructions exist",
				Description:      "Contribution guidance should explain how to set up the repository locally.",
				Severity:         SeverityMedium,
				Category:         CategoryDocumentation,
				EnabledByDefault: true,
			},
			{
				ID:               "contributing.test_instructions",
				Title:            "CONTRIBUTING test instructions exist",
				Description:      "Contribution guidance should explain how to run tests before submitting changes.",
				Severity:         SeverityMedium,
				Category:         CategoryDocumentation,
				EnabledByDefault: true,
			},
			{
				ID:               "ci.workflow.exists",
				Title:            "CI workflow exists",
				Description:      "Repository should include a CI workflow so contributors can verify changes.",
				Severity:         SeverityMedium,
				Category:         CategoryCI,
				EnabledByDefault: true,
			},
			{
				ID:               "ci.test_job.exists",
				Title:            "CI test job exists",
				Description:      "CI workflows should include an explicit test command or job signal.",
				Severity:         SeverityMedium,
				Category:         CategoryCI,
				EnabledByDefault: true,
			},
			{
				ID:               "scripts.dev.exists",
				Title:            "Developer scripts exist",
				Description:      "Repository should expose basic development commands through scripts or a Makefile.",
				Severity:         SeverityLow,
				Category:         CategoryWorkflow,
				EnabledByDefault: true,
			},
			{
				ID:               "scripts.test_command.exists",
				Title:            "Test command exists",
				Description:      "Repository should expose a deterministic command for running tests.",
				Severity:         SeverityMedium,
				Category:         CategoryWorkflow,
				EnabledByDefault: true,
			},
			{
				ID:               "scripts.fmt_command.exists",
				Title:            "Format command exists",
				Description:      "Repository should expose a deterministic command for formatting code.",
				Severity:         SeverityLow,
				Category:         CategoryWorkflow,
				EnabledByDefault: true,
			},
			{
				ID:               "scripts.lint_command.exists",
				Title:            "Lint command exists",
				Description:      "Repository should expose a deterministic command for linting code.",
				Severity:         SeverityLow,
				Category:         CategoryWorkflow,
				EnabledByDefault: true,
			},
		},
	}
}

// ResolveRuleSet applies effective configuration to a base ruleset.
func ResolveRuleSet(base RuleSet, effectiveConfig config.EffectiveConfiguration) (RuleSet, error) {
	if err := base.Validate(); err != nil {
		return RuleSet{}, err
	}

	known := make(map[string]Rule, len(base.Rules))
	active := make(map[string]bool, len(base.Rules))
	for _, rule := range base.Rules {
		known[rule.ID] = rule
		active[rule.ID] = rule.EnabledByDefault
	}

	for _, id := range effectiveConfig.DisabledRules {
		if _, exists := known[id]; !exists {
			return RuleSet{}, fmt.Errorf("unknown disabled rule id %q", id)
		}
		active[id] = false
	}
	for _, id := range effectiveConfig.EnabledRules {
		if _, exists := known[id]; !exists {
			return RuleSet{}, fmt.Errorf("unknown enabled rule id %q", id)
		}
		active[id] = true
	}

	resolved := RuleSet{
		ID:      base.ID,
		Name:    base.Name,
		Version: base.Version,
		Rules:   make([]Rule, 0, len(base.Rules)),
	}
	for _, rule := range base.Rules {
		if active[rule.ID] {
			resolved.Rules = append(resolved.Rules, rule)
		}
	}
	return resolved, nil
}
