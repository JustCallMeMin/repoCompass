# Contributor Guide

This guide explains how contributors should approach the RepoCompass repository.

## Purpose

RepoCompass has a working backend, CLI, Dockerized PostgreSQL workflow, API surface, and frontend shell. Contributors should keep changes aligned with the current module layout and update docs when workflow or behavior changes.

## Read First

Before making changes, review these documents:

- [README.md](../README.md): Project overview and basic local setup
- [docs/structure.md](./structure.md): Repository and backend structure reference
- [docs/analyzer-contract.md](./analyzer-contract.md): Contributor-facing analyzer contract
- [docs/finding-taxonomy.md](./finding-taxonomy.md): Finding, evidence, and recommendation taxonomy
- [docs/report-format.md](./report-format.md): Planned Markdown and JSON report formats
- [docs/persistence-history.md](./persistence-history.md): PostgreSQL persistence and history workflow

## Repository Layout

### Top-Level Directories

- `backend/`: Backend application code, database assets, scripts, and test fixtures
- `frontend/`: Next.js dashboard product surface
- `deployments/`: Deployment manifests and environment-specific setup
- `docs/`: Repository-facing documentation written in English

### Backend Layout

- `backend/cmd/`: Application entrypoints
- `backend/internal/`: Private application packages
- `backend/db/`: Database assets such as migrations and seeds
- `backend/scripts/`: Developer and automation scripts
- `backend/testdata/`: Shared test fixtures and sample inputs

Current subdirectories under `backend/` include:

- `backend/cmd/server`
- `backend/internal/app`
- `backend/internal/config`
- `backend/internal/repository`
- `backend/internal/service`
- `backend/db/migrations`
- `backend/db/seeds`
- `backend/scripts/dev`
- `backend/testdata/fixtures`

## Contributor Workflow

- Start with `README.md` for project orientation.
- Read `docs/structure.md` before adding or moving files.
- Place new files in the existing area that best matches their responsibility.
- Keep repository-facing documentation in English.
- Prefer small, reviewable changes that match the current milestone scope.

## Formatting, Linting, and Testing

The repository provides local workflow commands and GitHub Actions CI checks.

- Formatting: `make fmt`
- Linting/static checks: `make vet`
- Testing: `make test`
- Frontend build: `make frontend-build`
- Docker runtime build: `make docker-build`

At the current stage:

- the backend Go module is already in place
- the frontend dashboard package is already in place
- CI runs backend tests, backend vet, PostgreSQL integration tests, frontend lint/audit/build, and Docker runtime smoke checks
- the repository provides Make targets such as `make fmt`, `make vet`, `make test`, `make frontend-build`, and `make docker-build`

For local database workflow, copy `backend/.env.example` to `backend/.env` and use:

- `make db-up`
- `make migrate-status`
- `make migrate-up`
- `make migrate-down`
- `make db-reset`
- `make db-seed`
- `make db-status`

The underlying scripts under `backend/scripts/dev/` still remain available and continue to auto-load `backend/.env` when it exists.

Default Dockerized database URL for host-side commands:

```bash
export DATABASE_URL='postgres://postgres:postgres@localhost:55432/repocompass?sslmode=disable'
```

GitHub Actions workflow definition:

- `.github/workflows/ci.yml`

## Contribution Conventions

- Do not add new top-level directories unless there is a clear project-level reason.
- Prefer extending the existing backend layout before introducing new structural patterns.
- Update repository-facing docs when the project structure or setup instructions change.
- Keep documentation honest: do not document tools or commands as working if they are not implemented and tested.
