# Brainstorm T1-008: Scan Model & Status

## Objective
Define the `Scan` domain model and its lifecycle statuses in `backend/internal/scan/scan.go`. This task is strictly scoped to the data structures and state definitions representing a single repository scan analysis.

## Context & Constraints
- According to the Milestone 1 plan, T1-008 is focused purely on defining the core `Scan` struct and the possible statuses.
- The states must mirror the project vocabulary defined in `docs/scan-lifecycle.md`: `created`, `queued`, `running`, `completed`, `failed`, `cancelled`.
- No runners, orchestrators, or CLI wiring should be included in this task.
- We need basic state transition rules (e.g., cannot transition from `completed` to `running`) with corresponding unit tests.

## Decisions

### 1. Scan Model
- **`Scan` struct**:
  - `ID`: Unique identifier (string).
  - `SnapshotID`: Reference to the `RepositorySnapshot` being analyzed.
  - `Status`: Current state of the scan (`ScanStatus`).
  - `StartTime`, `EndTime`: Timestamps (`*time.Time` or `time.Time` with zero-value checks) to track duration.
  - `ErrorDetails`: Optional string/error details if the scan fails.

### 2. Scan Status Enum
- `ScanStatus` type (`string`) with constants:
  - `StatusCreated` = "created"
  - `StatusQueued` = "queued"
  - `StatusRunning` = "running"
  - `StatusCompleted` = "completed"
  - `StatusFailed` = "failed"
  - `StatusCancelled` = "cancelled"

### 3. State Transition Rules
- Add a helper method on `Scan` or package-level function (e.g., `TransitionTo(newStatus) error`) to enforce valid state transitions:
  - `created` -> `queued`, `running`, `cancelled`
  - `queued` -> `running`, `cancelled`
  - `running` -> `completed`, `failed`, `cancelled`
  - terminal states (`completed`, `failed`, `cancelled`) cannot transition to any other state.

## Acceptance Criteria
- `Scan` struct and status constants are defined clearly.
- Statuses include: `created`, `queued`, `running`, `completed`, `failed`, `cancelled`.
- Unit tests verify the state transition rules.
