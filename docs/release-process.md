# Release Process

This document outlines how to release a new version of RepoCompass.

## 1. Version Bump Process
Currently, RepoCompass uses a hardcoded version. Before releasing:
1. Update `Version` in `backend/internal/cli/version.go`.
2. Commit and push the version bump:
   ```bash
   git commit -m "chore(release): bump version to vX.Y.Z"
   ```

## 2. Tag Strategy
We follow Semantic Versioning (`vMAJOR.MINOR.PATCH`).
- `MAJOR` version for incompatible API/CLI changes.
- `MINOR` version for backwards-compatible functionality.
- `PATCH` version for backwards-compatible bug fixes.

## 3. Changelog Generation Policy
We rely on GitHub's automated "Generate Release Notes" feature. Ensure Pull Request titles are descriptive, as they will directly form the changelog.

## 4. Triggering a Release
1. Go to the **Actions** tab on GitHub.
2. Select the **Release** workflow.
3. Click **Run workflow**.
4. Enter the version tag (e.g., `v0.1.0`).
5. This triggers the workflow and creates a **Draft Release** on GitHub.

## 5. Artifact Validation
The Release workflow automatically builds the Go binary (`make build`) and runs a smoke test (`./backend/bin/repocompass version`) to ensure the artifact is valid before attaching it to the release.

## 6. Publishing
Review the Draft Release on GitHub, adjust the changelog text if necessary to make it more user-friendly, and click **Publish release**.
