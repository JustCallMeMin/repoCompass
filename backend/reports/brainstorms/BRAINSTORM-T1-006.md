## Requirements Discovery: T1-006 (Snapshot Creation) & T1-007 (SnapshotStore)

### Initial Request
Thiết kế/implement flow tạo snapshot (T1-006) và interface/in-memory implementation cho SnapshotStore (T1-007).

### Clarifying Questions
1. Q: What is the exact responsibility of Snapshot Creation?
   A: It receives a resolved `Repository` from the `RepositoryProvider` and captures the repository state used by a scan. It assigns a stable ID, records branch/commit info, and preserves input state.
2. Q: Should Snapshot Creation run any code analysis?
   A: No. It simply produces a stable snapshot object for the scan runner and returns whether the snapshot was created successfully.
3. Q: What is the role of SnapshotStore?
   A: To store and retrieve snapshot records. It serves as an abstraction for persistence.
4. Q: What implementation is required for SnapshotStore at this stage?
   A: An in-memory implementation to support local CLI scans without needing a database right away.

### Problem Statement
The scan engine needs to freeze the state of a repository at a given moment in time (a "Snapshot") before running analyses, so that scans are reproducible and tied to specific repository state. We need a `Creator` to build these snapshots and a `SnapshotStore` (with an in-memory implementation) to persist them during the scan lifecycle.

### Stakeholders
| Role   | Needs   | Priority |
| ------ | ------- | -------- |
| Scan Engine (Core) | Needs a stable, immutable snapshot object to run analysis against. | High |
| Persistence Layer | Needs a defined `SnapshotStore` interface to persist scan contexts. | High |

### Requirements
#### Functional
| ID  | Requirement | Priority       |
| --- | ----------- | -------------- |
| FR1 | Define `RepositorySnapshot` domain model. | Must |
| FR2 | Implement `Creator.Create()` to build a snapshot from a `RepositoryResolution`. | Must |
| FR3 | Generate stable snapshot IDs based on repository ID, source type, and captured time. | Must |
| FR4 | Define `SnapshotStore` interface with basic Save and Get operations. | Must |
| FR5 | Implement `InMemorySnapshotStore` for local development and testing. | Must |

### Success Criteria
1. `Creator` can successfully create a `RepositorySnapshot` containing all necessary metadata from a resolved repository.
2. `SnapshotStore` interface is clearly defined.
3. `InMemorySnapshotStore` correctly stores and retrieves snapshots by ID.
4. Tests verify that snapshot creation and storing work as expected without a real database.

### Open Questions
- Is the in-memory store intended to be thread-safe (e.g. using `sync.RWMutex`) to support concurrent local scans?
