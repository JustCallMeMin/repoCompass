# RepoCompass Report

## Repository Summary

| Field | Value |
| --- | --- |
| ID | repo_123 |
| Name | example |
| Provider | local |
| URL | file:///repo/example |
| Default Branch | main |

## Scan Summary

| Field | Value |
| --- | --- |
| Scan ID | scan_123 |
| Snapshot ID | snap_123 |
| Status | completed |
| Started At | 2026-05-01T01:00:00Z |
| Completed At | 2026-05-01T01:00:02Z |
| Analyzers | 1 |
| Findings | 1 |

## Assessment

Overall score: **75/100** (good)

### Severity Counts

| Severity | Count |
| --- | --- |
| high | 1 |

### Category Scores

| Category | Score |
| --- | --- |
| documentation | 75 |

## Analyzer Results

| Analyzer | Name | Version | Status | Findings |
| --- | --- | --- | --- | --- |
| readme | README Analyzer | 0.1.0 | success | 1 |

## Findings

### High: README file is missing

- Rule: `readme.exists`
- Analyzer: `readme`
- Category: `documentation`
- Message: The repository does not contain a README file at its root.

Evidence:

- `file_missing`: README.md was not found at the repository root. (`README.md`)

## Recommendations

### Add a root README

- Finding: `finding_readme`
- Priority: `high`
- Action: Create README.md with project purpose, setup steps, and test commands.
- Rationale: New contributors need one stable entry point before changing code.

## Metadata

| Field | Value |
| --- | --- |
| Format | markdown |
| Schema Version | 0.1.0 |
| Generated At | 2026-05-01T01:02:03Z |
| RepoCompass Version | unknown |
