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
The Release workflow uses `backend/go.mod` as the Go version source of truth.
It builds CLI artifacts for:

- Linux amd64
- Linux arm64
- macOS amd64
- macOS arm64
- Windows amd64

The workflow smoke-tests the Linux artifact with:

```bash
./dist/repocompass-linux-amd64 help
./dist/repocompass-linux-amd64 version
./dist/repocompass-linux-amd64 scan ./backend/testdata/fixtures/local-repositories/good-onboarding-repo
```

It also writes `dist/SHA256SUMS` for artifact verification.

To verify a downloaded artifact:

```bash
sha256sum -c SHA256SUMS
```

## 6. Publishing
Review the Draft Release on GitHub, check the generated notes, confirm the
attached artifacts and `SHA256SUMS`, then click **Publish release**.

## 7. Rollbacks and Hotfixes
If a released version contains a critical flaw:
1. **Rollback**: Instruct users to downgrade to the previous stable version tag. We do not unpublish tags unless there is a severe security exposure.
2. **Hotfix**: Create a branch from the affected release tag (e.g., `git checkout -b hotfix/v0.1.1 v0.1.0`), push the fix, update the version in `version.go`, and trigger a new release.

## 8. Repository Secrets
**No repository secrets are required** for the standard release pipeline. The GitHub Action uses the default `GITHUB_TOKEN` to create the release and upload artifacts.
