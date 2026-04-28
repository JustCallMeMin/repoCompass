# Contributor Guide

This guide explains how contributors should approach the repository in its current bootstrap phase.

## Purpose

RepoCompass is currently focused on foundation work: repository structure, internal documentation, and basic development conventions. Contributors should treat the repository as an early-stage project and keep changes aligned with the current layout.

## Read First

Before making changes, review these documents:

- [README.md](../README.md): Project overview and basic local setup
- [docs/structure.md](./structure.md): Repository and backend structure reference
- [docs/finding-taxonomy.md](./finding-taxonomy.md): Finding, evidence, and recommendation taxonomy

## Repository Layout

### Top-Level Directories

- `backend/`: Backend application code, database assets, scripts, and test fixtures
- `frontend/`: Frontend application area
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
- Prefer small, reviewable changes that match the current bootstrap scope.

## Formatting, Linting, and Testing

The repository is still in the bootstrap phase, but the standard local workflow scripts now exist under `backend/scripts/dev/`.

- Formatting: `make fmt`
- Linting/static checks: `make vet`
- Testing: `make test`

At the current stage:

- the backend Go module is already in place
- the repository does not yet provide CI workflow definitions
- the repository now provides Make targets such as `make fmt`, `make vet`, and `make test`

For local database workflow, copy `backend/.env.example` to `backend/.env` and use:

- `make db-up`
- `make migrate-status`
- `make migrate-up`
- `make migrate-down`

The underlying scripts under `backend/scripts/dev/` still remain available and continue to auto-load `backend/.env` when it exists.

## Contribution Conventions

- Do not add new top-level directories unless there is a clear project-level reason.
- Prefer extending the existing backend layout before introducing new structural patterns.
- Update repository-facing docs when the project structure or setup instructions change.
- Keep bootstrap-phase documentation honest: do not document tools or commands as working if they are not yet implemented.
