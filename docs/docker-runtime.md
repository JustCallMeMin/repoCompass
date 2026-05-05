# Docker Runtime

This document describes the local Docker Compose product runtime for RepoCompass.

## Overview

The Docker runtime starts the full local product stack:

- PostgreSQL at `localhost:55432`
- RepoCompass API at `http://localhost:8080`
- RepoCompass dashboard at `http://localhost:3000`

The API container runs database migrations automatically before starting the
server. This behavior is intended for the local product runtime only.

## Commands

Build the images:

```bash
make docker-build
```

Start the stack:

```bash
make docker-up
```

Check running services:

```bash
make docker-ps
```

Follow logs:

```bash
make docker-logs
```

Stop the stack:

```bash
make docker-down
```

## Scanning Repositories

GitHub scans use public repository URLs and do not need host filesystem access.

Local scans from the Dockerized dashboard or API should use paths under
`/workspace`. The Compose stack mounts the repository root read-only at that
path inside the API container.

Example local scan path:

```text
/workspace/backend/testdata/fixtures/local-repositories/good-onboarding-repo
```

Example API request:

```bash
curl -sS -X POST http://localhost:8080/api/v1/scans \
  -H 'Content-Type: application/json' \
  -d '{"source_type":"local","path":"/workspace/backend/testdata/fixtures/local-repositories/good-onboarding-repo"}'
```

## Boundaries

- No Redis service is part of this runtime.
- No authentication layer is part of this runtime.
- No cloud deployment target is defined here.
- Existing `make db-up` remains a PostgreSQL-only helper for local Go and Next.js development.
