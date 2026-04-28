# Report Format

This document defines the planned RepoCompass report formats for Milestone 2.
It describes the human-readable Markdown report and the machine-readable JSON
report. The concrete renderer interfaces will be added in later backend tasks.

## Purpose

RepoCompass reports must make scan results useful to both humans and automation.

The Markdown format is for maintainers and contributors who want to read the
result in a terminal, pull request, or GitHub-rendered document.

The JSON format is for automation, API consumers, and future dashboard views.
Its field names and enum values should remain stable once renderer work begins.

## Usage

Milestone 2 targets these commands:

```bash
repocompass scan <local-path> --format markdown
repocompass scan <local-path> --format json
```

The `markdown` format should be the default human-readable output unless a
later CLI task chooses a different default.

## Markdown Report Structure

Markdown reports should use this section order:

```markdown
# RepoCompass Report

## Repository Summary

## Scan Summary

## Assessment

## Analyzer Results

## Findings

## Recommendations

## Metadata
```

### Repository Summary

Purpose: identify the scanned repository.

Expected content:

- repository name
- repository provider
- local path or URL when safe to display
- default branch when available

### Scan Summary

Purpose: identify the scan execution.

Expected content:

- scan ID
- snapshot ID
- scan status
- started and completed timestamps when available
- analyzer count
- finding count

### Assessment

Purpose: show the overall onboarding health in a compact form.

Expected content:

- overall score
- score label when available
- score breakdown by category
- severity counts

### Analyzer Results

Purpose: show which analyzers ran and whether they succeeded.

Expected content:

- analyzer ID
- analyzer name
- analyzer version
- analyzer status
- duration when available
- finding count produced by the analyzer

### Findings

Purpose: explain each detected issue.

Findings should be grouped by severity first, then category when useful.
Each finding should include:

- title
- severity
- category
- rule ID
- analyzer ID
- message
- evidence

### Recommendations

Purpose: list concrete next actions.

Recommendations may be shown under each finding and also summarized in a
dedicated section for quick scanning.

Each recommendation should include:

- title
- action
- rationale
- related finding ID when available

### Metadata

Purpose: expose report generation details without distracting from findings.

Expected content:

- report format
- schema version
- generated timestamp
- RepoCompass version when available

## Markdown Example

```markdown
# RepoCompass Report

## Repository Summary

| Field | Value |
| --- | --- |
| Name | example-service |
| Provider | local |
| Default Branch | main |

## Scan Summary

| Field | Value |
| --- | --- |
| Scan ID | scan_123 |
| Snapshot ID | snap_456 |
| Status | completed |
| Analyzers | 3 |
| Findings | 1 |

## Assessment

Overall score: **82/100**

| Category | Score |
| --- | --- |
| documentation | 70 |
| ci | 100 |

## Analyzer Results

| Analyzer | Status | Findings |
| --- | --- | --- |
| readme | success | 1 |
| ci_workflow | success | 0 |

## Findings

### High: README file is missing

- Rule: `readme.required`
- Analyzer: `readme`
- Category: `documentation`
- Message: The repository does not contain a README.md file at its root.

Evidence:

- `file_missing`: `README.md` was not found at the repository root.

Recommendation:

- Add a root README with project purpose, setup steps, and test commands.
```

## JSON Report Structure

JSON reports should use a stable top-level object:

```json
{
  "schema_version": "0.1.0",
  "repository": {},
  "scan": {},
  "assessment": {},
  "analyzers": [],
  "findings": [],
  "recommendations": [],
  "metadata": {}
}
```

### Stable Top-Level Fields

| Field | Purpose |
| --- | --- |
| `schema_version` | Version of the JSON report contract. |
| `repository` | Repository identity and source metadata. |
| `scan` | Scan identity, snapshot ID, lifecycle status, and timestamps. |
| `assessment` | Overall score and score breakdown. |
| `analyzers` | Analyzer execution results. |
| `findings` | Flat list of findings. |
| `recommendations` | Flat list of recommendations linked to findings. |
| `metadata` | Report generation metadata. |

### Stable Field Requirements

The following fields should be treated as stable for future API and dashboard
work:

- IDs: `repository.id`, `scan.id`, `scan.snapshot_id`, `analyzers[].id`,
  `findings[].id`, `findings[].rule_id`, `findings[].analyzer_id`
- enums: `scan.status`, `analyzers[].status`, `findings[].severity`,
  `findings[].category`, `findings[].evidence[].type`
- timestamps: `scan.started_at`, `scan.completed_at`,
  `metadata.generated_at`
- versioning: `schema_version`

## JSON Example

```json
{
  "schema_version": "0.1.0",
  "repository": {
    "id": "local_abc123",
    "name": "example-service",
    "provider": "local",
    "default_branch": "main"
  },
  "scan": {
    "id": "scan_123",
    "snapshot_id": "snap_456",
    "status": "completed",
    "started_at": "2026-04-28T10:00:00Z",
    "completed_at": "2026-04-28T10:00:02Z"
  },
  "assessment": {
    "overall_score": 82,
    "severity_counts": {
      "critical": 0,
      "high": 1,
      "medium": 0,
      "low": 0,
      "info": 0
    },
    "category_scores": {
      "documentation": 70,
      "ci": 100
    }
  },
  "analyzers": [
    {
      "id": "readme",
      "name": "README Analyzer",
      "version": "0.1.0",
      "status": "success",
      "findings_count": 1
    }
  ],
  "findings": [
    {
      "id": "finding_001",
      "rule_id": "readme.required",
      "analyzer_id": "readme",
      "severity": "high",
      "title": "README file is missing",
      "message": "The repository does not contain a README.md file at its root.",
      "category": "documentation",
      "status": "open",
      "evidence": [
        {
          "type": "file_missing",
          "path": "README.md",
          "message": "README.md was not found at the repository root."
        }
      ]
    }
  ],
  "recommendations": [
    {
      "finding_id": "finding_001",
      "title": "Add a root README",
      "action": "Create README.md with project purpose, setup steps, and test commands.",
      "rationale": "New contributors need one stable entry point before changing code."
    }
  ],
  "metadata": {
    "format": "json",
    "generated_at": "2026-04-28T10:00:02Z",
    "repocompass_version": "unknown"
  }
}
```

## Compatibility Rules

Report renderers should follow these compatibility rules once implemented:

- Additive fields are allowed.
- Existing field names should not be renamed without a schema version change.
- Enum values should remain lowercase strings.
- JSON should not include absolute local paths unless the user explicitly asks
  for them.
- Markdown can be optimized for readability, but it should preserve the same
  core data as JSON.
