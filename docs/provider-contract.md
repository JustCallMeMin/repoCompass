# Provider Contract

A Provider (or Integration) is responsible for connecting RepoCompass to an external source of repositories (like GitHub, GitLab, or a local filesystem).

## Core Concepts

A Repository Provider typically implements ways to:
1. Authenticate with the external service.
2. Fetch metadata about a repository.
3. Download or clone the repository to the local filesystem so Snapshots can be created.

## Implementing a Provider

When adding a new Provider:
- Ensure that credentials or tokens are passed securely (usually via context or constructor).
- Abstract the remote API behind interfaces so that the core scan engine is unaware of the provider specifics.
- Write tests using local mock servers or mocked interfaces to prevent flaky network tests.

## Minimal Example

Check out the runnable example Provider in `backend/internal/integration/example/provider.go`.
