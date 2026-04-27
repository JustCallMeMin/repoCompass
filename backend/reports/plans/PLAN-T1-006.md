# Implementation Plan: T1-006 & T1-007 (Snapshot Creation Flow & SnapshotStore)

## 📌 User Request (VERBATIM)
> Xác nhận: hiện tại không có requirement bắt buộc concurrent scans trong brainstorm/docs. Giai đoạn này sequential local CLI scan là đủ theo tài liệu.
> Nhưng implementation hiện tại đã dùng sync.RWMutex trong backend/internal/snapshot/store.go, nên InMemorySnapshotStore đã thread-safe ở mức Save/Get map access. Nên PLAN nên ghi:
> Requirement tối thiểu: sequential local scans.
> Implementation choice: keep sync.RWMutex as low-cost guard for future concurrent scans.
> Không cần mở rộng scope sang scan scheduler/concurrent runner ở T1-006/T1-007.
> Không cần thêm yêu cầu “support many concurrent scans” làm acceptance chính, trừ khi muốn test race guard.

## 🎯 Acceptance Criteria (Derived from User Request)
| ID | Criterion | Verification Method |
|----|-----------|---------------------|
| AC1 | `RepositorySnapshot` model is correctly defined | Inspect `backend/internal/snapshot/snapshot.go` |
| AC2 | `Creator` handles snapshot creation with stable IDs | Inspect `creator.go` and run tests |
| AC3 | `SnapshotStore` interface defined and `InMemorySnapshotStore` implemented | Inspect `store.go` |
| AC4 | `InMemorySnapshotStore` uses `sync.RWMutex` to guard map access (best practice) | Code inspection of `store.go` |
| AC5 | No concurrent scan runner required for T1-006/T1-007 | Implementation review |

## 📋 Context Summary
**Architecture**: The core scan engine flow is `repository -> snapshot -> scan`. Snapshot captures the repository state and produces a stable object. The `SnapshotStore` persists it.
**Patterns**: Interface-based data store (`SnapshotStore`), separated domain model (`RepositorySnapshot`).
**Constraints**: Keep the scope narrow to local sequential runs. Use of `sync.RWMutex` is permitted as an implementation choice but concurrent stress testing is out of scope.

## Overview
This plan focuses on verifying the existing implementation for `T1-006` (Snapshot creation) and `T1-007` (SnapshotStore). Since the modules are already present in `backend/internal/snapshot/`, the goal is to formally validate that they fulfill the accepted criteria and pass the test suite.

## Prerequisites
- [x] `RepositoryProvider` verification completed (T1-004 & T1-005).

## Phase 1: Verification of T1-006 & T1-007
### Tasks
- [ ] Task 1.1: Verify `RepositorySnapshot` and `Creator` implementation
  - Agent: `backend-engineer`
  - File(s): `backend/internal/snapshot/snapshot.go`, `backend/internal/snapshot/creator.go`
  - Acceptance: Snapshot ID logic, metadata copy, and branch/commit info capture are correct.
  - Verification: Code inspection.

- [ ] Task 1.2: Verify `SnapshotStore` interface and `InMemorySnapshotStore`
  - Agent: `backend-engineer`
  - File(s): `backend/internal/snapshot/store.go`
  - Acceptance: Interface methods (Save/Get) exist. Map accesses are guarded by `sync.RWMutex`.
  - Verification: Code inspection.

- [ ] Task 1.3: Execute Test Suite for Snapshot Module
  - Agent: `tester`
  - File(s): `backend/internal/snapshot/creator_test.go`, `backend/internal/snapshot/store_test.go`
  - Acceptance: All tests pass.
  - Verification: Run `go test ./internal/snapshot/...` successfully.

### Exit Criteria
- [ ] Implementation aligns with sequential local CLI scan expectations.
- [ ] Tests execute without failures, proving valid ID generation and state storage.

## Risks
| Risk | Impact | Mitigation | Rollback |
|------|--------|------------|----------|
| Test Failures | Medium | Debug and fix the failing test cases | Revert test/code changes |

## Rollback Strategy
N/A (Code is already implemented; verification only).

## Implementation Notes
The user has confirmed that the code is currently in a verified state and the concurrency requirement is strictly limited to map protection using `sync.RWMutex` for future-proofing. The implementer (`backend-engineer`/`tester`) should simply run the tests and verify the code against the ACs.
