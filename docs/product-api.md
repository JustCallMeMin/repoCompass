# Product API And GitHub Integration

## Purpose

RepoCompass exposes a small HTTP API for persisted scan workflows and read-only
history views. The API builds on the CLI scan engine, PostgreSQL persistence,
and report/history models already used by local workflows.

## Running The Server

Start PostgreSQL, apply migrations, and run the API server:

```bash
make db-up
make migrate-up
make server
```

Configuration:

- `DATABASE_URL`: required PostgreSQL connection string.
- `PORT`: optional HTTP port, defaults to `8080`.
- `GITHUB_WEBHOOK_SECRET`: optional secret for GitHub webhook HMAC validation.

The Docker Compose database listens on host port `55432`, matching
`backend/.env.example`.

## Routes

- `GET /healthz`: returns server health.
- `POST /api/v1/scans`: runs a persisted scan.
- `GET /api/v1/repositories/{repository_id}/scans`: lists persisted scan history.
- `GET /api/v1/scans/{scan_id}/findings`: lists persisted findings for one scan.
- `GET /api/v1/repositories/{repository_id}/metrics`: lists metric trend data. Defaults to `assessment.overall_score`.
- `POST /api/v1/integrations/github/webhook`: accepts basic GitHub webhook events.

## Scan Requests

Local repository scan:

```json
{
  "source_type": "local",
  "path": "./testdata/fixtures/local-repositories/good-onboarding-repo"
}
```

Public GitHub repository scan:

```json
{
  "source_type": "github",
  "url": "https://github.com/owner/repo"
}
```

GitHub scans clone the public repository into a temporary local checkout, run
the existing local scan flow, then remove the checkout after the request.

## Webhooks

The webhook endpoint validates `X-Hub-Signature-256` when
`GITHUB_WEBHOOK_SECRET` is configured. The first supported events are:

- `ping`: returns `ok`
- `push`: accepts the event and records the repository identity from the payload

Webhook handling is intentionally minimal in this milestone. It validates and
parses events but does not yet provide a full background queue.
