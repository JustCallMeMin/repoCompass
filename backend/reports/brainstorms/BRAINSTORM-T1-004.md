## Requirements Discovery: T1-004 (RepositoryProvider Interface) & T1-005 (LocalRepositoryProvider)

### Initial Request
Implementing the foundational repository management system for the repoCompass backend. Specifically:
- **T1-004**: Create the `RepositoryProvider` interface in the backend layer.
- **T1-005**: Implement `LocalRepositoryProvider` for filesystem support.

### Clarifying Questions
1. Q: What exactly should the `RepositoryProvider` do according to the scan engine?
   A: It resolves an input source (e.g., local path) into a known repository identity and metadata, validating that the source can be scanned.
2. Q: What does the returned object look like?
   A: A stable `Repository` model containing location, name, and validation status, which can be consumed by the snapshot creation phase.
3. Q: Should it perform any analysis or generate findings?
   A: No. Its sole job is to answer "what repository are we scanning?"
4. Q: How do we handle errors?
   A: Errors (like path not found, permissions denied, or invalid repository structure) should be returned early during resolution.

### Problem Statement
The scan engine needs a standardized way to identify, validate, and extract metadata from diverse repository sources (starting with local filesystems) before snapshot creation. We need an abstraction (`RepositoryProvider`) and a concrete implementation (`LocalRepositoryProvider`) to handle this.

### Stakeholders
| Role   | Needs   | Priority |
| ------ | ------- | -------- |
| Scan Engine (Core) | Needs a uniform `Repository` model regardless of source | High |
| Developer/User | Needs clear errors if the input path is invalid | High |

### Requirements
#### Functional
| ID  | Requirement | Priority       |
| --- | ----------- | -------------- |
| FR1 | Define `Repository` domain model (ID, Path, Name, basic metadata). | Must |
| FR2 | Define `RepositoryProvider` interface with a `Resolve(ctx, source string) (*Repository, error)` method. | Must |
| FR3 | Implement `LocalRepositoryProvider` that validates local paths and checks readability. | Must |
| FR4 | The `LocalRepositoryProvider` should extract basic metadata (e.g., folder name). | Must |
| FR5 | Return appropriate errors (e.g., "path not found", "not a directory") matching the project's error strategy. | Must |

### Success Criteria
1. `RepositoryProvider` interface is defined in `backend/internal/repository/`.
2. `LocalRepositoryProvider` is implemented and successfully resolves valid local directories.
3. `LocalRepositoryProvider` returns explicit errors for invalid paths (non-existent, files instead of dirs).
4. A stable `Repository` object is returned, ready for the upcoming `SnapshotStore`.

### Open Questions
- Should `LocalRepositoryProvider` ensure it's a Git repository by checking for `.git`, or any valid directory is acceptable for scanning?
