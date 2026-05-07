# Organization API Documentation

## Overview

The Organization API provides multi-tenant isolation for RepoCompass.
All org-scoped endpoints require an `X-Organization-Id` header.
Omitting it defaults to the Personal organization (`00000000-0000-0000-0000-000000000000`).

**Base URL:** `/api/v1`

---

## Authentication

> **M6 MVP**: The caller is supplied through the local-dev `X-User-Id` header.
> If the header is omitted, the API falls back to `mock_user` for fixture and demo flows.
> Session-based authentication is still future work and must be added before public deployment.
> Local migrations seed `mock_user` as owner of the default Personal organization.

All requests go through RBAC checks:
- `CheckAccessRepo` — required for any read on org resources (role: viewer+)
- `CheckEditRepo` — required for repository scan/write operations (role: member+)
- `CheckManageOrg` — required for member and policy management (role: owner/admin)

---

## Headers

| Header | Required | Description |
|---|---|---|
| `X-User-Id` | No | Local-dev actor ID. Defaults to `mock_user` when omitted. |
| `X-Organization-Id` | No | UUID of the target organization. Defaults to personal org. |
| `X-Request-ID` | No | Client-supplied request correlation ID. Generated server-side if absent. |
| `Content-Type` | For POST/PUT | Must be `application/json` |

---

## Endpoints

### Organizations

#### `GET /api/v1/organizations`

Returns all organizations the current user is a member of.

**Response 200**
```json
{
  "data": [
    { "id": "org_abc", "name": "ACME Corp", "created_at": "...", "updated_at": "..." }
  ]
}
```

---

#### `GET /api/v1/organizations/{organization_id}`

Returns a single organization by ID.

**Response 200**
```json
{ "data": { "id": "org_abc", "name": "ACME Corp", "created_at": "...", "updated_at": "..." } }
```

**Response 404** — Organization not found.

---

### Memberships

#### `GET /api/v1/organizations/{organization_id}/members`

Lists all members of an organization.

**Response 200**
```json
{
  "data": [
    { "organization_id": "org_abc", "user_id": "user_1", "role": "owner", "created_at": "..." }
  ]
}
```

---

#### `POST /api/v1/organizations/{organization_id}/members`

Adds a user to an organization with a role.

**Request body**
```json
{ "user_id": "user_2", "role": "member" }
```
Valid roles: `owner`, `admin`, `member`, `viewer`

**Response 200**
```json
{ "status": "success" }
```

---

### Policies

#### `GET /api/v1/organizations/{organization_id}/policies`

Lists all assessment policies for an organization.

**Response 200**
```json
{
  "data": [
    {
      "id": "pol_1",
      "organization_id": "org_abc",
      "name": "assessment_policy",
      "rules": { "minimum_score": 80, "require_readme": true },
      "created_at": "...",
      "updated_at": "..."
    }
  ]
}
```

---

#### `GET /api/v1/policies/{policy_id}`

Returns a single policy by ID. Caller must be a member of the policy's organization.

---

#### `POST /api/v1/organizations/{organization_id}/policies`

Creates or updates a policy. Requires `admin` or `owner` role.

**Request body**
```json
{
  "id": "pol_1",
  "name": "assessment_policy",
  "rules": { "minimum_score": 80, "require_readme": true }
}
```

**Response 200**
```json
{ "status": "success" }
```

---

### Insights

#### `GET /api/v1/organizations/{organization_id}/insights`

Returns aggregated health statistics for the organization.

**Response 200**
```json
{
  "data": {
    "organization_id": "org_abc",
    "average_score": 82,
    "total_repositories": 15,
    "total_scans": 120
  }
}
```

---

### Operational

#### `GET /api/v1/metrics`

Returns server operational metrics. No authentication required.

**Response 200**
```json
{
  "uptime_seconds": 3600,
  "goroutines": 12,
  "memory": { "alloc_bytes": 4096000, "total_alloc_bytes": 8192000, "sys_bytes": 16384000, "gc_cycles": 5 },
  "go_version": "go1.22.3"
}
```

---

#### `GET /healthz`

Health check endpoint.

**Response 200**
```json
{ "status": "ok" }
```

---

## Error Format

All errors follow a consistent envelope:

```json
{
  "error": {
    "code": "forbidden",
    "message": "auth: user mock_user is not a member of org org_abc"
  }
}
```

| Code | HTTP Status | Description |
|---|---|---|
| `invalid_json` | 400 | Request body could not be parsed |
| `invalid_request` | 400 | Missing required field |
| `invalid_scan_source` | 400 | Scan source type or path invalid |
| `forbidden` | 403 | RBAC check failed |
| `not_found` | 404 | Resource does not exist |
| `not_implemented` | 501 | Feature not yet configured |
| `history_query_failed` | 500 | Internal query error |
| `insights_query_failed` | 500 | Internal insights query error |
| `db_error` | 500 | Database write failed |
