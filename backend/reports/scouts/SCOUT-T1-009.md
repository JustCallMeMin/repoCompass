# Scout Report: T1-009 ScanRunner Interface

## Exploration Scope
- Target: `ScanRunner` interface integration and codebase patterns.
- Boundaries: `backend/internal/scan/`, `backend/internal/snapshot/`.

## Patterns Discovered
### Pattern: Contract-first interfaces
- **Location**: `backend/internal/snapshot/store.go`
- **Usage**: `SnapshotStore` defines the core methods. Implementation `InMemorySnapshotStore` is in the same file (though sometimes separated).
- **Must Follow**: Yes.

### Pattern: Service Structs
- **Location**: `backend/internal/snapshot/creator.go`
- **Usage**: `Creator` is a struct because it doesn't currently require multiple implementations.
- **Must Follow**: No, for `ScanRunner` we explicitly want an interface to support different scan engines (local, cloud, mock).

## Integration Points
| Point | File | Function | New Code Location |
| ------ | ------ | -------- | ----------------- |
| Runner Interface | `backend/internal/scan/runner.go` | `ScanRunner` | `backend/internal/scan/runner.go` |
| Snapshot Reference | `backend/internal/snapshot/snapshot.go` | `RepositorySnapshot` | Used as input to `Run` |
| Scan Result | `backend/internal/scan/scan.go` | `Scan` | Returned from `Run` |

## Conventions
- Naming: `ScanRunner` is the standard name for the interface.
- File organization: Interface and its implementation should likely be separated into `runner.go` and specific implementation files (like `local_runner.go` in T1-010).

## Warnings
- ⚠️ Ensure `ScanRunner` does not create circular dependencies with the `snapshot` package. Since `scan` already depends on `snapshot` (via the `Scan` struct referencing `SnapshotID`), we must ensure `snapshot` doesn't depend on `scan`. Currently, `snapshot` is clean of `scan` dependencies.
