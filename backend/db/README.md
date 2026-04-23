# Database Directory Guide

This directory contains database-related assets for the RepoCompass backend.

The database workflow is still in the bootstrap phase. The folder structure exists now so contributors can organize future migration and seed files consistently before the execution tooling is introduced.

## Directory Overview

- `migrations/`: Versioned schema changes for the database
- `seeds/`: Seed data for bootstrap, development, or test scenarios

## Migrations

The `migrations/` directory is reserved for ordered schema changes such as:

- creating or altering tables
- adding indexes or constraints
- introducing rollback-safe structural changes

Migration tooling and execution workflow will be formalized in `T0-008`. Until then, contributors should treat this directory as the canonical location for future schema migration files, but should not document any migration command as available yet.

## Seeds

The `seeds/` directory is reserved for seed data assets such as:

- initial reference data
- development bootstrap data
- test-supporting seed datasets

The exact execution workflow for seeds is not defined yet. For now, this directory exists to establish the correct location and separation of concerns for future work.

## Conventions

- Keep migration filenames ordered and descriptive.
- Keep seed filenames descriptive and scoped to their purpose.
- Separate schema changes from data seeding.
- Do not add usage commands to documentation until migration and seed tooling is implemented.

## Local Migration Workflow

Migration tooling is now available through the scripts in `backend/scripts/dev/`.

Expected setup:

- a running local PostgreSQL instance
- `DATABASE_URL` exported in the shell or stored in `backend/.env`

Example environment variable:

```bash
export DATABASE_URL='postgres://user:pass@localhost:5432/repocompass?sslmode=disable'
```

You can also store local values in `backend/.env` by copying `backend/.env.example`.

## Local PostgreSQL via Docker Compose

Contributors can start a local PostgreSQL instance with one command:

```bash
./backend/scripts/dev/db-up.sh
```

This starts a PostgreSQL service with these defaults:

- database: `repocompass`
- user: `postgres`
- password: `postgres`
- port: `5432`

To stop the local database:

```bash
./backend/scripts/dev/db-down.sh
```

Available scripts:

- `backend/scripts/dev/migrate-up.sh`: apply all pending up migrations
- `backend/scripts/dev/migrate-down.sh`: apply down migrations
- `backend/scripts/dev/migrate-status.sh`: show migration version status
- `backend/scripts/dev/db-up.sh`: start local PostgreSQL with Docker Compose
- `backend/scripts/dev/db-down.sh`: stop the local Docker Compose PostgreSQL service

Equivalent Make aliases:

- `make migrate-up`
- `make migrate-down`
- `make migrate-status`
- `make db-up`
- `make db-down`

Notes:

- these scripts use `golang-migrate`
- the scripts target `backend/db/migrations`
- the scripts auto-load `backend/.env` when present
- local PostgreSQL provisioning is available through Docker Compose in this task
- `migrate-status.sh` maps to the upstream `version` command because `golang-migrate` does not expose a literal `status` subcommand

## Current Status

- `migrations/` now includes a bootstrap validation migration
- `seeds/` exists and is intentionally empty except for placeholder tracking
- migration tooling is available through `backend/scripts/dev/`
- seed execution tooling is not implemented in this task

## Bootstrap Validation Migration

The first migration in this repository is intentionally minimal and exists only to validate the migration toolchain and rollback workflow.

Current bootstrap migration:

- `000001_create_schema_bootstrap_checks.up.sql`
- `000001_create_schema_bootstrap_checks.down.sql`

This migration creates and removes a small table named `schema_bootstrap_checks`. It is infrastructure-focused and should not be treated as the first real RepoCompass domain schema.
