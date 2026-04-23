# Repository Structure

This document describes the initial backend-oriented directory structure of the repository.

## Overview

The repository is organized as a multi-area project, with the backend implementation isolated under `backend/`.

## Top-Level Directories

- `backend/`: Go backend application code, database assets, scripts, and test fixtures.
- `frontend/`: Reserved for frontend application code.
- `deployments/`: Reserved for deployment-related manifests and environment setup.
- `docs/`: Project documentation written in English.

## Backend Directory Layout

- `backend/cmd/`: Application entrypoints and executable bootstrap code.
- `backend/cmd/server/`: Main server binary startup logic.
- `backend/internal/`: Private application packages that should not be imported from outside the module.
- `backend/internal/app/`: Application wiring and runtime composition.
- `backend/internal/assessment/`: Assessment models and assessment coordination building blocks.
- `backend/internal/cli/`: CLI command wiring for the `repocompass` executable.
- `backend/internal/config/`: Configuration loading and configuration models.
- `backend/internal/integration/`: External integration boundaries.
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

## Notes

- Documentation in this project should be written in English.
- New backend packages should be added under the existing `cmd`, `internal`, `db`, `testdata`, and `scripts` structure unless there is a clear reason to introduce a new top-level backend directory.
