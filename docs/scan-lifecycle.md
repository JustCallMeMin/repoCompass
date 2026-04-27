# Scan Lifecycle

This document describes the developer-facing lifecycle for a RepoCompass scan.
It is a design guide for the core scan engine, not a claim that the full scan
runner is already implemented.

## Core Flow

The scan engine follows this core flow:

```text
repository -> snapshot -> scan
```

Each step has a narrow responsibility:

- `repository`: resolves an input source into repository identity and basic metadata.
- `snapshot`: captures the repository state that one scan should analyze.
- `scan`: owns execution state, result metadata, and failure information for one analysis run.

The flow should stay explicit because later rules, findings, reports, and
persistence work depend on knowing which repository state produced which scan
result.

## Repository Resolution

Repository resolution turns user input into a known repository source. In local
mode, this starts with a filesystem path and should resolve enough metadata for
later steps to identify the repository consistently.

At minimum, repository resolution is responsible for:

- validating that the source can be scanned
- identifying the repository location
- collecting basic repository metadata when available
- returning a stable object that snapshot creation can consume

Resolution should not run analyzers or produce findings. Its job is to answer:
"what repository are we scanning?"

## Snapshot Creation

A repository snapshot captures the repository state used by a scan. It should be
created after repository resolution and before scan execution.

At minimum, snapshot creation is responsible for:

- recording the source repository identity
- recording branch, commit, or equivalent local metadata when available
- preserving the input state needed to reproduce or explain the scan result
- producing a stable snapshot object for the scan runner

Snapshot creation should not decide scan status beyond reporting whether the
snapshot was created successfully.

## Scan Execution

A scan represents one analysis run over one repository snapshot. The scan owns
its lifecycle state and should expose enough result metadata for CLI output,
later reports, and future persistence.

At minimum, scan execution is responsible for:

- creating a scan record or in-memory scan object
- moving through lifecycle states consistently
- attaching the repository snapshot being analyzed
- returning a structured result, even before real analyzers exist
- recording failure details when the scan cannot complete

## Scan States

The initial lifecycle should use these states:

| State | Meaning |
| --- | --- |
| `created` | A scan object exists, but execution has not been scheduled or started. |
| `queued` | A scan is accepted for execution and waiting to run. |
| `running` | Repository resolution, snapshot creation, or analyzer execution is in progress. |
| `completed` | The scan finished successfully and produced a result. |
| `failed` | The scan stopped because of a recoverable or user-facing error. |
| `cancelled` | The scan was intentionally stopped before completion. |

Early local CLI scans may move directly from `created` to `running` to a terminal
state. The state names should still remain stable so future queueing,
persistence, and background execution can use the same vocabulary.

## Failure and Logging Expectations

Failures should be attached to the scan that encountered them. User-facing
errors should be stable enough for the CLI to print clear messages, while
developer-facing logs should preserve enough context to debug the failing stage.

Future scan lifecycle logging should include fields such as operation name,
scan ID, repository ID, and error ID when those values exist. The logging shape
should follow the project error and logging strategy instead of inventing a
separate convention in the scan runner.

## Reference Documents

- [Architecture Design](https://www.notion.so/349e98673be48113a4d3e03b7c0dc6c8)
- [Entity Flow Overview](https://www.notion.so/349e98673be48106b6f8d0b24bb876ad)
- [Error Handling & Logging Strategy](https://www.notion.so/349e98673be481d1bbcadd43641ea67e)
