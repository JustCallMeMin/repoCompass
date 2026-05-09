# Analyzer Contract

This document defines the contributor-facing contract for RepoCompass analyzers.
The concrete Go interface is implemented in `backend/internal/analyzer/analyzer.go`.

## Purpose

An analyzer is a deterministic component that inspects a resolved repository
snapshot and produces structured scan results.

Analyzers turn repository facts into:

- analyzer metadata, such as ID, name, and version
- findings, which describe verifiable repository issues
- evidence, which explains why each finding exists
- recommendations, which describe actionable next steps

Analyzers must not mutate the repository, scan state, configuration, or
filesystem input. They only read from the scan context and return structured
output.

## Usage

A scan runner will call analyzers after repository resolution and snapshot
creation.

Expected flow:

```text
repository -> snapshot -> scan -> analyzers -> findings -> report
```

Each analyzer should:

1. Read the analyzer input.
2. Check a small, well-defined repository concern.
3. Return an analyzer result.
4. Include evidence for every finding.
5. Return stable metadata, even when no finding is produced.

Analyzer execution must be deterministic. For the same input repository state
and configuration, the analyzer must return the same result. Analyzer decisions
must come from repository facts, configuration, and rules, not generated guesses.

## Input

The analyzer input is a scan context containing:

| Field | Purpose |
| --- | --- |
| `Repository` | Resolved repository identity and local path metadata. |
| `Snapshot` | Repository snapshot metadata captured for this scan. |
| `EffectiveConfiguration` | Final configuration after defaults, file config, and CLI overrides are resolved. |
| `RuleSet` | The active rules that decide which checks should run. |

The input should be read-only from the analyzer perspective.

## Output

The analyzer output is an analyzer result containing:

| Field | Purpose |
| --- | --- |
| `AnalyzerID` | Stable machine-readable analyzer ID, such as `readme`. |
| `Name` | Human-readable analyzer name. |
| `Version` | Analyzer contract or implementation version. |
| `Status` | Execution result, such as `success`, `skipped`, or `failed`. |
| `Findings` | Structured findings produced by the analyzer. |
| `Metadata` | Optional analyzer-specific summary values. |
| `Error` | Optional analyzer error information when the analyzer fails. |

A successful analyzer may return zero findings. That means the analyzer ran and
did not find an issue.

## Findings, Evidence, And Recommendations

A finding is a verifiable issue discovered by an analyzer.

Each finding should include:

- stable rule ID
- analyzer ID
- severity
- title
- message
- category
- evidence
- recommendation when the finding is actionable

Evidence is a concrete fact that explains why a finding exists. Examples
include `file_missing`, `file_exists`, `pattern_match`, and `metadata`.

A recommendation is an action a maintainer can take to resolve or reduce the
finding. Recommendations should be specific enough that a contributor can act
without guessing.

## Error Handling

Analyzer errors should be isolated when possible.

Expected behavior:

- A single analyzer failure should produce a failed analyzer result.
- Other analyzers should still be allowed to run when the scan runner can safely continue.
- Fatal scan-level errors should remain scan runner errors, not analyzer results.
- Analyzer errors should include enough metadata for logs and reports to identify the failing analyzer.

## Minimal Example

The project includes a runnable minimal analyzer in
`backend/internal/analyzers/example/analyzer.go`. Test it with:

```bash
cd backend
go test ./internal/analyzers/example
```

For a production analyzer with findings and evidence, inspect the README,
CONTRIBUTING, CI, and scripts analyzers under `backend/internal/analyzers/`.

## Contributor Checklist

When adding a new analyzer, verify that it:

- has a stable analyzer ID
- documents the repository concern it checks
- reads only from analyzer input
- does not mutate repository files
- returns deterministic results
- includes evidence for every finding
- includes actionable recommendations when a finding should be fixed
- has fixture tests for pass and fail scenarios
