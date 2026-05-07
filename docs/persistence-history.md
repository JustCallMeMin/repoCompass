# Persistence And History

## Purpose

RepoCompass persistence stores local scan results in PostgreSQL so maintainers can inspect scan history, findings, evidence, recommendations, assessments, and metric trends.

## Usage

Start Dockerized PostgreSQL, apply migrations, run a persisted scan, then query history:

```powershell
make db-up
$env:DATABASE_URL="postgres://postgres:postgres@localhost:55432/repocompass?sslmode=disable"
make migrate-up
cd backend
go run ./cmd/repocompass scan ./testdata/fixtures/local-repositories/good-onboarding-repo --persist
go run ./cmd/repocompass history <repository-id>
go run ./cmd/repocompass history <repository-id> --format json
go run ./cmd/repocompass findings <scan-id> --format json
```

## Parameters

- `DATABASE_URL`: PostgreSQL connection string used by migration scripts, `scan --persist`, `history`, and `findings`.
- `scan --persist`: enables DB persistence. Without this flag, scans remain no-DB local runs.
- `history <repository-id>`: returns newest scans for one persisted repository.
- `findings <scan-id>`: returns persisted finding details for one scan.

## Schema

Core identity tables store repositories, snapshots, and scans. Analysis tables store rule sets, rules, analyzer results, findings, finding evidence, recommendations, assessments, metric snapshots, and report metadata.

## M3 Schema Decision

Milestone 3 uses a pragmatic persistence schema with text IDs and scan-history-focused tables. This intentionally differs from the broader product Physical Database Schema in a few places so the persisted CLI/API foundation can remain small and usable. Product operation tables such as organizations, users, integrations, and GitHub integration state are deferred to later milestones.

## Rollback

Use migration down scripts for rollback:

```powershell
make migrate-down
```

For a full local reset:

```powershell
make db-reset
```
