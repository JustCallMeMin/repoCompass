# Implementation Plan: T1-004 & T1-005 (RepositoryProvider & LocalRepositoryProvider)

## 📌 User Request (VERBATIM)
> Không bắt buộc .git ở giai đoạn hiện tại.
> Bằng chứng code hiện tại: LocalRepositoryProvider.Resolve chỉ kiểm tra:
> source.Path không rỗng
> filepath.Abs(...) resolve được
> os.Stat(...) thành công
> path là directory
> Sau đó nó tạo Repository với ProviderLocal, StatusActive, file://.... Git chỉ dùng best-effort qua:
> git -C <path> branch --show-current
> Nếu lệnh fail, default_branch rỗng, nhưng resolve vẫn thành công.
> Test cũng xác nhận fixture basic-go-repo không có .git, chỉ có go.mod và README.md, nhưng TestLocalRepositoryProviderResolvesValidPath vẫn expect success.
> Kết luận: hiện tại “Repository” nghĩa là local directory scan target, không nhất thiết Git repository

## 🎯 Acceptance Criteria (Derived from User Request)
| ID | Criterion | Verification Method |
|----|-----------|---------------------|
| AC1 | `RepositoryProvider` interface defined | Inspect `backend/internal/repository/provider.go` |
| AC2 | `LocalRepositoryProvider` resolves any valid local directory | Run `go test` on `local_provider_test.go` |
| AC3 | `.git` is not strictly required, branch resolution is best-effort | Verify tests pass without `.git` requirement |

## 📋 Context Summary
**Architecture**: The backend scan engine requires a `Repository` domain model as the input for snapshot creation. `RepositoryProvider` abstracts the resolution process.
**Patterns**: Interface-based design, `ProviderRegistry` for provider selection.
**Constraints**: Must return explicit errors for missing paths or files.

## Overview
This plan outlines the formal verification of the already-implemented `RepositoryProvider` (T1-004) and `LocalRepositoryProvider` (T1-005) to ensure they meet the defined criteria and are ready for the next milestone tasks (Snapshot flow). Since the code is already present, the implementer will only need to run tests and confirm.

## Prerequisites
- [x] Go workspace initialized
- [x] Error handling strategy documented

## Phase 1: Verification of T1-004 & T1-005
### Tasks
- [ ] Task 1.1: Verify `RepositoryProvider` interface and models
  - Agent: `backend-engineer`
  - File(s): `backend/internal/repository/provider.go`, `backend/internal/repository/repository.go`
  - Acceptance: Interfaces and structures are correctly defined.
  - Verification: Code inspection.

- [ ] Task 1.2: Verify `LocalRepositoryProvider` implementation
  - Agent: `backend-engineer`
  - File(s): `backend/internal/repository/local_provider.go`
  - Acceptance: Resolves valid directories, handles errors, best-effort Git branch resolution.
  - Verification: Code inspection.

- [ ] Task 1.3: Run Test Suite for Repository Module
  - Agent: `tester`
  - File(s): `backend/internal/repository/local_provider_test.go`
  - Acceptance: All tests in the repository module pass.
  - Verification: `make test` or `go test ./internal/repository/...` completes successfully.

### Exit Criteria
- [ ] Code meets all functional requirements (FR1-FR5) from Brainstorm phase.
- [ ] Tests pass successfully, proving that a `.git` directory is not mandatory.

## Risks
| Risk | Impact | Mitigation | Rollback |
|------|--------|------------|----------|
| Test Failures | Medium | Debug and fix the failing test cases | Revert changes to test/code if any |

## Rollback Strategy
N/A (Code is already implemented and in the verification stage).

## Implementation Notes
The user has confirmed that the code is already implemented and satisfies the requirements. The implementer (`backend-engineer`/`tester`) should simply verify the existing code structure and run the tests to formally close out T1-004 and T1-005 in the workflow.
