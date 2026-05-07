# Milestone 5 Dashboard Product Surface

## Purpose

Milestone 5 turns RepoCompass scan data into a usable dashboard for maintainers.
The dashboard answers: which repositories need attention, which scan is latest,
which findings are important, which recommendations should be handled first,
and whether score is improving or declining.

## Routes

- `/dashboard`: overview, scan trigger, repository metrics.
- `/repositories`: searchable repository inventory.
- `/repositories/[repositoryId]`: repository overview, latest scan, score trend,
  insights, recommendations preview, scan history.
- `/repositories/[repositoryId]/scans`: repository scan history alias.
- `/scans/[scanId]`: scan summary, assessment, reports, findings preview.
- `/scans/[scanId]/findings`: filtered findings and evidence panel.
- `/scans/[scanId]/recommendations`: grouped recommendations.

Organization management pages can exist, but they are not the core M5 flow.

## Local Setup

```bash
make db-up
make migrate-up
make db-seed
DEV_HEADER_AUTH=true make server
cd frontend
npm ci
npm run dev
```

Set `NEXT_PUBLIC_REPOCOMPASS_API_URL=http://localhost:8080` in
`frontend/.env.local`. Local dev uses `NEXT_PUBLIC_REPOCOMPASS_USER_ID=mock_user`.

## Testing

```bash
cd frontend
npm run lint
npm run typecheck
npm test
npm run build
```

With the API and dashboard already running:

```bash
DASHBOARD_URL=http://localhost:3000 make dashboard-smoke
```

Backend gates remain:

```bash
cd backend
go test ./...
go vet ./...
```

## Demo Script

1. Start Dockerized PostgreSQL and apply migrations.
2. Run `make db-seed` to create repositories with scans, findings,
   recommendations, assessment, reports, and metric snapshots.
3. Open `/dashboard`.
4. Navigate `/repositories` -> repository detail -> scan detail.
5. Open findings, expand evidence, then open recommendations.
6. Confirm score trend and assessment are visible.

## Acceptance Checklist

- Dashboard runs locally.
- Session/dev-header protected product surface loads.
- Repository list and detail use API data.
- Scan history/detail use persisted API data.
- Findings and recommendations are not raw JSON.
- Assessment and score trend are visible.
- Loading, empty, error, and unauthorized states render.
- Frontend lint, typecheck, tests, and build pass.
