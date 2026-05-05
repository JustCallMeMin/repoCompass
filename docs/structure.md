# Repository Structure

This document describes the initial backend-oriented directory structure of the repository.

## Overview

The repository is organized as a multi-area project, with the backend implementation isolated under `backend/`.

## Top-Level Directories

- `backend/`: Go backend application code, database assets, scripts, and test fixtures.
- `frontend/`: Next.js dashboard product surface.
- `deployments/`: Reserved for deployment-related manifests and environment setup.
- `docs/`: Project documentation written in English.
- `docs/analyzer-contract.md`: Contributor-facing contract for future analyzer implementations.
- `docs/finding-taxonomy.md`: Contributor-facing taxonomy for findings, evidence, and recommendations.
- `docs/report-format.md`: Planned Markdown and JSON report format contract.
- `docs/scan-lifecycle.md`: Developer-facing scan lifecycle guide for the core scan engine.
- `docs/persistence-history.md`: PostgreSQL persistence and scan history workflow.
- `docs/product-api.md`: Product API and GitHub integration reference.
- `docs/docker-runtime.md`: Local Docker product runtime guide.
- `docker-compose.yml`: Local Docker Compose stack for PostgreSQL, API, and dashboard.
- `.github/workflows/ci.yml`: CI workflow for backend, frontend, PostgreSQL integration, and Docker runtime smoke checks.

## Backend Directory Layout

- `backend/cmd/`: Application entrypoints and executable bootstrap code.
- `backend/cmd/server/`: Main server binary startup logic.
- `backend/Dockerfile`: Docker image definition for the API server runtime.
- `backend/internal/`: Private application packages that should not be imported from outside the module.
- `backend/internal/api/`: HTTP API handlers and product API wiring.
- `backend/internal/app/`: Application wiring and runtime composition.
- `backend/internal/assessment/`: Assessment models and assessment coordination building blocks.
- `backend/internal/cli/`: CLI command wiring for the `repocompass` executable.
- `backend/internal/config/`: Configuration loading and configuration models.
- `backend/internal/integration/`: External integration boundaries.
- `backend/internal/history/`: Read models for persisted scan history and finding details.
- `backend/internal/repository/`: Data access layer and persistence-facing code.
- `backend/internal/report/`: Report generation building blocks and output shaping.
- `backend/internal/rules/`: Rule evaluation and rule definitions.
- `backend/internal/scan/`: Scan orchestration primitives.
- `backend/internal/snapshot/`: Snapshot lifecycle and snapshot creation primitives.
- `backend/internal/storage/`: Storage abstractions and persistence helpers.
- `backend/internal/service/`: Business logic and service orchestration.
- `backend/db/`: Database-related assets.
- `backend/db/README.md`: Usage guide for organizing migrations and seed data during the bootstrap phase.
- `backend/db/migrations/`: Schema migration files.
- `backend/db/seeds/`: Seed data and bootstrap database scripts.
- `backend/testdata/`: Shared test fixtures and sample inputs.
- `backend/testdata/README.md`: Fixture naming and usage conventions for backend tests.
- `backend/testdata/fixtures/`: Concrete fixture files used by tests.
- `backend/scripts/`: Developer and automation scripts.
- `backend/scripts/dev/`: Local development helper scripts.
- `backend/scripts/docker/`: Docker runtime entrypoint scripts.

## Frontend Directory Layout

- `frontend/app/`: Next.js App Router pages for the dashboard, repository history, and scan findings.
- `frontend/components/`: Shared dashboard UI components.
- `frontend/lib/`: API client helpers for the backend product API.
- `frontend/Dockerfile`: Docker image definition for the dashboard runtime.
- `frontend/.env.example`: Local dashboard environment template.

## Notes

- Documentation in this project should be written in English.
- New backend packages should be added under the existing `cmd`, `internal`, `db`, `testdata`, and `scripts` structure unless there is a clear reason to introduce a new top-level backend directory.
