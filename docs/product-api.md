# Product API And GitHub Integration

## Purpose

RepoCompass exposes the Milestone 4 Product API for persisted scans,
repository history, GitHub webhook intake, and session-aware access. The
OpenAPI contract lives in [openapi-m4.yaml](openapi-m4.yaml).

## Running The Server

Start PostgreSQL, apply migrations, and run the API server:

```bash
make db-up
make migrate-up
DEV_HEADER_AUTH=true make server
```

Configuration:

- `DATABASE_URL`: required PostgreSQL connection string.
- `PORT`: optional HTTP port, defaults to `8080`.
- `DEV_HEADER_AUTH`: set to `true` to allow local `X-User-Id` and
  `X-Organization-Id` headers.
- `GITHUB_WEBHOOK_SECRET`: required outside dev mode for GitHub webhook HMAC
  validation. The API rejects GitHub webhook requests when this value is missing.
- `GITHUB_OAUTH_CLIENT_ID`, `GITHUB_OAUTH_CLIENT_SECRET`,
  `GITHUB_OAUTH_REDIRECT_URL`: GitHub OAuth session configuration.

## Response Envelope

All M4 API responses use this envelope:

```json
{"data": {}, "meta": {"request_id": "req_..."}, "error": null}
```

Errors set `data` to `null` and use stable `error.code` values.

## Routes

- `GET /api/v1/health`: returns server health.
- `GET /api/v1/repositories`: lists repositories visible to the current actor.
- `GET /api/v1/repositories/{repository_id}`: returns repository detail.
- `POST /api/v1/repositories/{repository_id}/scans`: runs a persisted scan for
  an existing repository.
- `POST /api/v1/scans`: backward-compatible direct scan trigger.
- `GET /api/v1/repositories/{repository_id}/scans`: lists persisted scan
  history.
- `GET /api/v1/scans/{scan_id}`: returns scan detail.
- `GET /api/v1/scans/{scan_id}/findings`: lists findings with evidence and
  recommendations.
- `GET /api/v1/scans/{scan_id}/assessment`: returns persisted assessment.
- `GET /api/v1/scans/{scan_id}/reports`: returns report metadata.
- `GET /api/v1/repositories/{repository_id}/metrics`: lists metric trend data.
- `POST /api/v1/integrations/github/webhook`: validates GitHub webhook payloads,
  persists the event, and queues a scan job.
- `GET /api/v1/auth/github/login`: persists one-time OAuth state and redirects
  to GitHub OAuth; `?format=json` returns the authorization URL.
- `GET /api/v1/auth/github/callback`: creates a RepoCompass session.
- `GET /api/v1/auth/session`: returns the current actor and organization.
- `POST /api/v1/auth/logout`: revokes the current session.

## GitHub Webhooks

Webhook requests must include `X-GitHub-Event`, `X-GitHub-Delivery`, and
`X-Hub-Signature-256`. Duplicate delivery IDs are rejected by a unique database
constraint. `push` events create a `github_webhook_events` row and a queued
`scan_jobs` row for the repository.
