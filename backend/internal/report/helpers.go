package report

import (
	"sort"
	"strings"
	"time"

	"github.com/JustCallMeMin/repoCompass/backend/internal/analyzer"
	"github.com/JustCallMeMin/repoCompass/backend/internal/findings"
	"github.com/JustCallMeMin/repoCompass/backend/internal/rules"
)

const schemaVersion = "0.1.0"

func collectFindings(results []analyzer.AnalyzerResult) []findings.Finding {
	var collected []findings.Finding
	for _, result := range results {
		collected = append(collected, result.Findings...)
	}
	return collected
}

func collectRecommendations(values []findings.Finding) []findings.Recommendation {
	var collected []findings.Recommendation
	for _, finding := range values {
		collected = append(collected, finding.Recommendations...)
	}
	return collected
}

func sortedSeverityKeys(values map[rules.Severity]int) []rules.Severity {
	keys := make([]rules.Severity, 0, len(values))
	for key := range values {
		keys = append(keys, key)
	}
	sort.Slice(keys, func(i, j int) bool { return string(keys[i]) < string(keys[j]) })
	return keys
}

func sortedCategoryKeys(values map[rules.Category]int) []rules.Category {
	keys := make([]rules.Category, 0, len(values))
	for key := range values {
		keys = append(keys, key)
	}
	sort.Slice(keys, func(i, j int) bool { return string(keys[i]) < string(keys[j]) })
	return keys
}

func markdownCell(value string) string {
	value = strings.ReplaceAll(value, "\r\n", " ")
	value = strings.ReplaceAll(value, "\n", " ")
	value = strings.ReplaceAll(value, "|", "\\|")
	return value
}

func formatTimePointer(value *time.Time) string {
	if value == nil {
		return ""
	}
	return value.UTC().Format(time.RFC3339)
}
