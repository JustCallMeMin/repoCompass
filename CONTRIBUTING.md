# Contributing to RepoCompass

RepoCompass welcomes small, reviewable contributions that improve repository
analysis, onboarding, dashboard workflows, documentation, and release quality.

## Purpose

This guide is the root entrypoint for contributors. It points to the current
setup, testing, extension, and pull request workflow used by the repository.

## Quick Setup

Docker is the fastest path for new contributors:

```bash
git clone https://github.com/JustCallMeMin/repoCompass.git
cd repoCompass
make docker-up
```

Then open:

- Dashboard: `http://localhost:3000`
- API health: `http://localhost:8080/api/v1/health`

On Windows, run Make targets from Git Bash or WSL. If POSIX shell tools are not
available, prefer Docker Compose commands directly.

## Local Development

Read these docs before changing code:

- [Start Here](docs/start-here.md)
- [Local Setup](docs/local-setup.md)
- [Repository Structure](docs/structure.md)
- [Testing Guide](docs/testing-guide.md)
- [Contributor Checklist](docs/contributor-checklist.md)

Use the standard checks before opening a pull request:

```bash
make fmt
make vet
make test
make frontend-build
```

For database-backed work, start Dockerized PostgreSQL and run:

```bash
make db-up
make migrate-up
make test-postgres
```

## Extension Points

- Add analyzers with [Analyzer Contract](docs/analyzer-contract.md).
- Add repository providers with [Provider Contract](docs/provider-contract.md).
- Add report renderers with [Renderer Contract](docs/renderer-contract.md).

Every new extension must include deterministic tests and fixture data when
behavior depends on repository contents.

## Issues and Pull Requests

- Use issue templates under `.github/ISSUE_TEMPLATE/`.
- Use labels from [Label Taxonomy](docs/label-taxonomy.md).
- Maintainers triage with [Triage Guide](docs/triage-guide.md).
- Fill out `.github/pull_request_template.md` before review.

Security issues must be reported through [SECURITY.md](SECURITY.md), not public
GitHub issues.
