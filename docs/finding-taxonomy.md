# Finding, Evidence, And Recommendation Taxonomy

This document defines the RepoCompass taxonomy for findings, evidence, and
recommendations. It is contributor-facing guidance for analyzers and report
renderers.

## Purpose

RepoCompass scan results must be explainable and actionable.

This taxonomy defines:

- a `Finding`: a verifiable repository issue
- `Evidence`: the concrete fact that proves or explains a finding
- a `Recommendation`: the action a maintainer or contributor can take

Analyzers should use this taxonomy so reports stay consistent across README,
CONTRIBUTING, CI, scripts, and future checks.

## Usage

Analyzers should create a finding only when a rule can be evaluated from
deterministic repository facts.

Valid facts include:

- a file exists
- a file is missing
- a simple pattern is present or absent
- repository metadata has a known value
- scan configuration enables a rule

Analyzers must not create findings from guesses, generated opinions, or AI
judgment.

## Finding

A finding describes one issue that RepoCompass detected.

### Required Fields

| Field | Purpose |
| --- | --- |
| `ID` | Stable finding ID generated from scan, rule, analyzer, and target when available. |
| `RuleID` | Stable rule ID that produced the finding. |
| `AnalyzerID` | Stable analyzer ID that produced the finding. |
| `Severity` | Impact level using the severity taxonomy below. |
| `Title` | Short human-readable issue title. |
| `Message` | Clear explanation of what is wrong. |
| `Category` | Product area such as `documentation`, `workflow`, `ci`, or `maintainability`. |
| `Status` | Finding lifecycle status, initially `open`. |
| `Evidence` | One or more facts supporting the finding. |

### Severity Levels

| Severity | Meaning | Example |
| --- | --- | --- |
| `critical` | The repository is very hard to use or assess without fixing this. | Required onboarding entry point is missing and no alternative exists. |
| `high` | New contributors are likely blocked or misled. | No README exists. |
| `medium` | Contributors can proceed, but with avoidable friction. | README exists but lacks setup instructions. |
| `low` | Minor improvement with limited immediate impact. | Optional metadata or polish is missing. |
| `info` | Useful observation that does not require action. | Repository has CI workflow detected. |

Severity should reflect user impact, not implementation difficulty.

### Finding Creation Rules

Create a finding when all of these are true:

1. A rule is enabled.
2. The analyzer can evaluate the rule deterministically.
3. The result indicates a concrete issue.
4. Evidence can explain why the issue exists.

Do not create a finding when:

- the rule is disabled
- required input is unavailable and the analyzer should skip
- the analyzer cannot distinguish failure from absence
- the message would be vague or unverifiable

## Evidence

Evidence records the concrete fact behind a finding.

### Evidence Types

| Type | Meaning | Required Context |
| --- | --- | --- |
| `file_exists` | A specific file or directory exists. | `Path` |
| `file_missing` | A specific file or directory is missing. | `Path` |
| `pattern_match` | A simple text pattern was found. | `Path`, pattern summary |
| `pattern_missing` | A simple text pattern was not found. | `Path`, pattern summary |
| `metadata` | Repository or scan metadata supports the finding. | key and value summary |

### Required Fields

| Field | Purpose |
| --- | --- |
| `Type` | One of the evidence types above. |
| `Message` | Short explanation of the fact. |
| `Path` | Relative repository path when evidence relates to a file. |
| `Value` | Optional short value or pattern summary. |

Evidence should use repository-relative paths. Avoid absolute local paths in
reports because they are noisy and can expose user-specific machine details.

## Recommendation

A recommendation describes the next action for a finding.

### Required Fields

| Field | Purpose |
| --- | --- |
| `Title` | Short action-oriented summary. |
| `Action` | Specific change the maintainer should make. |
| `Rationale` | Why the action improves onboarding or maintainability. |
| `FindingID` | Link back to the finding when IDs are available. |

### Actionable Recommendation Rules

A recommendation is actionable when:

- it names the file, section, command, or config to change
- it explains the expected outcome
- it can be done without guessing the analyzer's intent
- it does not require AI-generated content to be useful

Avoid recommendations such as "improve documentation" without concrete next
steps. Prefer "Add a README.md section named `Local development` with setup,
test, and run commands."

## Example

Missing README finding:

```json
{
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
  ],
  "recommendation": {
    "title": "Add a root README",
    "action": "Create README.md with project purpose, setup steps, test commands, and contribution entry points.",
    "rationale": "New contributors need one stable entry point before changing code."
  }
}
```

## Contributor Checklist

Before adding a new finding type or analyzer output, confirm that:

- severity follows the user-impact scale
- every finding has at least one evidence item
- evidence is based on deterministic facts
- file evidence uses repository-relative paths
- recommendation text names a concrete action
- messages are clear enough for a maintainer to act on
