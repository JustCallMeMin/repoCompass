# Operations Runbook

## Overview

This runbook covers operational procedures for the local RepoCompass product runtime.

---

## 1. Prerequisites

| Tool | Version | Purpose |
|---|---|---|
| Go | Version from `backend/go.mod` | Backend build and test |
| PostgreSQL | ≥ 15 | Primary data store |
| `golang-migrate` CLI | latest | Database migrations |
| Docker / Docker Compose | latest | Local stack |
| Node.js | ≥ 20 | Frontend build |

---

## 2. Environment Variables

| Variable | Required | Description |
|---|---|---|
| `DATABASE_URL` | Yes | PostgreSQL DSN, e.g. `postgres://user:pass@host:5432/repocompass?sslmode=disable` |
| `PORT` | No (default 8080) | HTTP listen port |
| `GITHUB_WEBHOOK_SECRET` | Yes | HMAC secret for GitHub webhook validation |
| `DEV_HEADER_AUTH` | No | Enables local `X-User-Id` and `X-Organization-Id` headers when set to `true`; leave unset for real session auth |
| `GITHUB_OAUTH_CLIENT_ID` | For login | GitHub OAuth application client ID |
| `GITHUB_OAUTH_CLIENT_SECRET` | For login | GitHub OAuth application client secret |
| `GITHUB_OAUTH_REDIRECT_URL` | For login | GitHub OAuth callback URL |
| `APP_ENV` | No | Set to `production` to mark session cookies `Secure` |
| `LOG_LEVEL` | No (default info) | Structured log level: debug, info, warn, error |
| `NEXT_PUBLIC_REPOCOMPASS_USER_ID` | No | Dashboard local-dev actor ID. Only set when `DEV_HEADER_AUTH=true`. |
| `NEXT_PUBLIC_REPOCOMPASS_ORG_ID` | No | Dashboard local-dev organization header. Defaults to Personal org. |

> No repository secrets are required for the CI test pipeline beyond `DATABASE_URL`
> for integration tests (skipped when absent).

Local migrations still seed `mock_user` for deterministic dev/test data. Product
usage should sign in through GitHub OAuth so the stored user comes from GitHub.

GitHub OAuth uses a persisted one-time `state` value. Expired, missing, or reused
state values are rejected.

---

## 3. Starting the Stack Locally

```bash
# 1. Start PostgreSQL
docker compose up -d postgres

# 2. Run migrations (empty schema → latest)
export DATABASE_URL="postgres://postgres:postgres@localhost:55432/repocompass?sslmode=disable"
migrate -path backend/db/migrations -database "$DATABASE_URL" up

# 3. Start the API server
cd backend && go run ./cmd/server

# 4. Start the frontend
cd frontend && npm run dev
```

---

## 4. Running Migrations

### Fresh install (empty schema)
```bash
migrate -path backend/db/migrations -database "$DATABASE_URL" up
```

### Pre-M6 schema upgrade
```bash
# Verify current version
migrate -path backend/db/migrations -database "$DATABASE_URL" version

# Apply through latest M6 hardening migration
migrate -path backend/db/migrations -database "$DATABASE_URL" up
```

### Rollback
```bash
# Roll back one step
migrate -path backend/db/migrations -database "$DATABASE_URL" down 1

# Roll back one migration at a time and verify data after each step
migrate -path backend/db/migrations -database "$DATABASE_URL" down 1
```

---

## 5. Running Tests

```bash
# Unit + integration tests (postgres skipped without DATABASE_URL)
cd backend && go test ./...

# With database integration tests
export DATABASE_URL="postgres://postgres:postgres@localhost:55432/repocompass?sslmode=disable"
cd backend && go test ./...

# Frontend lint
cd frontend && npm run lint

# Frontend build
cd frontend && npm run build
```

---

## 6. API Smoke Checks

```bash
BASE="http://localhost:8080"

# Health
curl "$BASE/healthz"

# Operational metrics
curl "$BASE/api/v1/metrics"

# List organizations as local-dev actor when DEV_HEADER_AUTH=true
curl -H "X-User-Id: mock_user" "$BASE/api/v1/organizations"

# Browser login with real GitHub user
open "$BASE/api/v1/auth/github/login"

# List org notifications
curl -H "X-User-Id: mock_user" \
  "$BASE/api/v1/organizations/00000000-0000-0000-0000-000000000000/notifications"

# Create scan
curl -X POST "$BASE/api/v1/scans" \
  -H "Content-Type: application/json" \
  -H "X-Organization-Id: 00000000-0000-0000-0000-000000000000" \
  -d '{"source_type":"local","path":"/path/to/repo"}'

# M4 smoke script
BASE_URL="$BASE" sh backend/scripts/dev/test-api-m4.sh
```

---

## 7. Hotfix Procedure

1. Create a branch from the failing release tag: `git checkout -b hotfix/vX.Y.Z-fix tags/vX.Y.Z`
2. Apply the fix, add a regression test.
3. Run `cd backend && go test ./...` and frontend lint.
4. Tag: `git tag vX.Y.Z+1` and push.
5. Re-run the release workflow (`.github/workflows/release.yml`).

---

## 8. Rollback Procedure

1. Identify the last good release tag.
2. Redeploy the prior binary artifact.
3. If schema changed: run `migrate down N` to revert migrations.
4. Verify with health endpoint and API smoke checks.
5. File a post-mortem issue.

---

## 9. Known Limitations (M6 MVP)

| Limitation | Mitigation |
|---|---|
| Local-dev `X-User-Id` actor header is not production auth | Replace with session auth before public deployment |
| Scan rate limit is in-process only | Replace with shared limiter only if API runs multiple replicas |
| No TLS enforcement | Run behind a reverse proxy (nginx/Caddy) in staging/prod |
| Email notification provider is a stub | Configure a real provider after M6 if outbound email is required |
| Webhook notification delivery is minimal | Use short timeouts and inspect `notification_deliveries` on failure |

---

*See `docs/security-threat-model.md` for security findings.*
