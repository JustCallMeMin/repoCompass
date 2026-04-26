# Implementation Plan: T1-009 ScanRunner Interface

## 📌 User Request (VERBATIM)
> T1-009: ScanRunner interface. Define the ScanRunner interface in backend/internal/scan/.

## 🎯 Acceptance Criteria (Derived from User Request)
| ID | Criterion | Verification Method |
|----|-----------|---------------------|
| AC1 | `ScanRunner` interface defined in `backend/internal/scan/` | Check file content |
| AC2 | `RunRequest` struct includes `snapshot.RepositorySnapshot` | Check struct definition |
| AC3 | `RunResult` struct includes `Scan` output | Check struct definition |
| AC4 | `Run` method uses `RunRequest` and `RunResult` with error handling | Check method signature |
| AC5 | No circular dependencies between `scan` and `snapshot` packages | `go list` or compilation check |

## 📋 Context Summary
**Architecture**: Following the pattern in `internal/snapshot`, we use explicit request/result structs for service-like operations to ensure a stable contract.
**Patterns**: Input/Output encapsulation via structs.
**Constraints**: `scan` package imports `snapshot`. `snapshot` must NOT import `scan`.

## Overview
This task defines the core contract for scan execution using encapsulated request and result objects as required by the project specifications. This ensures the interface remains stable even if the underlying execution data needs to change.

## Prerequisites
- [ ] `backend/internal/scan/scan.go` defines the `Scan` struct.
- [ ] `backend/internal/snapshot/snapshot.go` defines the `RepositorySnapshot` struct.

## Phase 1: Interface Definition
### Tasks
- [ ] Task 1.1: Create `backend/internal/scan/runner.go`
  - Agent: `backend-engineer`
  - File(s): `c:\Users\hoang\Documents\dev\go\repoCompass\repoCompass\backend\internal\scan\runner.go`
  - Acceptance: 
    - Package is `scan`.
    - `RunRequest` contains `Snapshot snapshot.RepositorySnapshot`.
    - `RunResult` contains `Scan Scan`.
    - `ScanRunner` interface has method `Run(ctx context.Context, request RunRequest) (RunResult, error)`.
  - Verification: Manual inspection of the file.

- [ ] Task 1.2: Verification of compilation and dependencies
  - Agent: `tester`
  - File(s): `c:\Users\hoang\Documents\dev\go\repoCompass\repoCompass\backend\internal\scan\runner.go`
  - Acceptance: Code compiles without errors. No circular dependencies detected.
  - Verification: Run `go build ./internal/scan/...`.

### Exit Criteria
- [ ] `ScanRunner` interface and supporting structs are defined.
- [ ] Contract for input/output/error is explicit.
- [ ] Compilation successful.

## Risks
| Risk | Impact | Mitigation | Rollback |
|------|--------|------------|----------|
| Circular Dependency | High | Ensure `snapshot` never imports `scan`. | Delete `runner.go` |

## Rollback Strategy
1. Delete `backend/internal/scan/runner.go`.

## Implementation Notes
- The signature must be:
```go
type RunRequest struct {
    Snapshot snapshot.RepositorySnapshot
}

type RunResult struct {
    Scan Scan
}

type ScanRunner interface {
    Run(ctx context.Context, request RunRequest) (RunResult, error)
}
```
