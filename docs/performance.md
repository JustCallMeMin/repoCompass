# Performance

This document records query hot paths and indexes for organization management.

## Query Hot Paths

| Surface | Query shape | Risk | M6 mitigation |
| --- | --- | --- | --- |
| Organization switcher | memberships by user, then organizations by id | Low | Membership user/org indexes. |
| Org repositories page | repositories by `organization_id` | Medium | Repository org/status index. |
| Org scan history | scan history by repository and completion time | Medium | Existing repository scan query plus org ownership check. |
| Org metrics page | metric trend by repository and metric key | Medium | Metric repository/key/time index. |
| Org insights | aggregate repositories/scans by organization | Medium | Organization, repository, and scan status indexes. |

## Added Indexes

Migration `backend/db/migrations/000005_add_m6_performance_indexes.up.sql` adds indexes for:

- organization membership lookup by user and organization
- repositories by organization and status
- scans by repository and status/completion time
- findings by scan and severity/status
- metrics by repository, key, and capture time
- policies by organization

## Baseline EXPLAIN Checklist

Run these after `make docker-up` and `make migrate-up` against a populated local database:

```sql
EXPLAIN ANALYZE
SELECT *
FROM repositories
WHERE organization_id = 'org_personal_default'
ORDER BY updated_at DESC
LIMIT 20;

EXPLAIN ANALYZE
SELECT *
FROM scan_runs
WHERE repository_id = 'repo_id'
ORDER BY completed_at DESC
LIMIT 20;

EXPLAIN ANALYZE
SELECT *
FROM metric_points
WHERE repository_id = 'repo_id'
  AND metric_key = 'assessment.overall_score'
ORDER BY captured_at DESC
LIMIT 50;
```

Expected M6 result: each query should use an index scan or bounded sort on the indexed relation. Sequential scans are acceptable only on tiny local fixture datasets.

## Caching Decision

No cache was added in M6. The indexed query paths are small, deterministic, and local-first. Adding cache before production traffic would add invalidation risk without proven benefit.
