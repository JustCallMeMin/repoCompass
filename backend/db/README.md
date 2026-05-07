# Database Directory Guide

## Purpose

This directory is the database source of truth for RepoCompass backend development. It stores versioned PostgreSQL migrations and development seed workflow assets used by the Dockerized database environment.

## Usage

Start Dockerized PostgreSQL, set `DATABASE_URL`, then apply migrations:

```powershell
make db-up
$env:DATABASE_URL="postgres://postgres:postgres@localhost:55432/repocompass?sslmode=disable"
make migrate-up
make db-status
```

Reset and seed the development database:

```powershell
make db-reset
make db-seed
```

Rollback all migrations:

```powershell
make migrate-down -- -all
```

## Parameters

- `DATABASE_URL`: PostgreSQL connection string used by migration scripts, seed scripts, persisted scans, and history queries.
- `backend/db/migrations`: ordered SQL migration files consumed by `golang-migrate`.
- `backend/scripts/dev`: Dockerized database, migration, seed, and PostgreSQL test scripts.

## Migrations

RepoCompass uses SQL migrations run by `golang-migrate`. PostgreSQL runs in Docker for dev/test; `golang-migrate` is only the migration runner.

Current migration phases:

- `000001_create_schema_bootstrap_checks`: verifies migration apply/down plumbing.
- `000002_create_core_scan_tables`: creates repositories, snapshots, and scans.
- `000003_create_analysis_persistence_tables`: creates rule, analyzer result, finding, evidence, recommendation, assessment, metric, and report metadata tables.
- `000004_harden_persistence_history_indexes`: adds M3 history/query-path indexes and scan status constraints.

Migration conventions:

- Use monotonically increasing numeric prefixes.
- Always add matching `.up.sql` and `.down.sql` files.
- Keep schema changes separate from seed data.
- Add indexes with the query path they protect in mind.
- Do not edit released migrations; add a new migration instead.

## Seed Data

`make db-seed` runs persisted scans against repository fixtures to create development history data. The seed data is dev/test-only and must not be treated as production data.

## M3 Schema Decision

Milestone 3 intentionally keeps a pragmatic persistence schema rather than forcing full parity with the broader Physical Database Schema document. The implementation uses text IDs and focused scan-history tables sufficient for CLI/API foundations. Product operation tables such as organizations, users, integrations, and GitHub-specific integration state are deferred to Milestone 4 and later.

## Examples

Run the PostgreSQL integration suite:

```powershell
make db-up
$env:DATABASE_URL="postgres://postgres:postgres@localhost:55432/repocompass?sslmode=disable"
make test-postgres
```

Run a persisted scan and inspect history:

```powershell
cd backend
go run ./cmd/repocompass scan ./testdata/fixtures/local-repositories/good-onboarding-repo --persist
go run ./cmd/repocompass history <repository-id> --format json
go run ./cmd/repocompass findings <scan-id> --format json
```
