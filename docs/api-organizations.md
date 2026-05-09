# Organization API Documentation

## Overview

The Organization API provides multi-tenant isolation for RepoCompass.
All org-scoped endpoints require an `X-Organization-Id` header.
Omitting it defaults to the Personal organization (`00000000-0000-0000-0000-000000000000`).

**Base URL:** `/api/v1`

---

## Authentication

RepoCompass supports session auth from the product API. Local development can enable
`DEV_HEADER_AUTH=true` and pass `X-User-Id` / `X-Organization-Id`. Production-style
flows should use GitHub OAuth and the `repocompass_session` cookie.
OAuth login stores and consumes a one-time `state`; replayed or expired callbacks
are rejected.

All requests go through RBAC checks:
- `CheckAccessRepo` — required for any read on org resources (role: viewer+)
- `CheckEditRepo` — required for repository scan/write operations (role: member+)
- `CheckManageOrg` — required for member and policy management (role: owner/admin)

---

## Headers

| Header | Required | Description |
|---|---|---|
| `X-User-Id` | Dev only | Local-dev actor ID when `DEV_HEADER_AUTH=true`. |
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

#### `POST /api/v1/organizations`

Creates an organization and adds the actor as `owner`.

**Request body**
```json
{ "id": "org_acme", "name": "ACME" }
```

---

#### `GET /api/v1/organizations/{organization_id}`

Returns a single organization by ID.

**Response 200**
```json
{ "data": { "id": "org_abc", "name": "ACME Corp", "created_at": "...", "updated_at": "..." } }
```

**Response 404** — Organization not found.

#### `PUT /api/v1/organizations/{organization_id}`

Updates an organization name. Requires `owner` or `admin`.

**Request body**
```json
{ "name": "ACME Platform" }
```

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
{
  "data": {
    "organization_id": "org_abc",
    "user_id": "user_2",
    "role": "member"
  }
}
```

#### `PUT /api/v1/organizations/{organization_id}/members/{user_id}`

Updates a member role. Requires `owner` or `admin`. The API rejects removing or
demoting the last `owner`.

#### `DELETE /api/v1/organizations/{organization_id}/members/{user_id}`

Removes a member. Requires `owner` or `admin`; last-owner protection applies.

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
      "status": "active",
      "version": 1,
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
  "status": "active",
  "version": 1,
  "rules": { "minimum_score": 80, "require_readme": true }
}
```

**Response 200**
```json
{ "data": { "id": "pol_1", "name": "assessment_policy", "status": "active", "version": 2 } }
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
    "total_scans": 120,
    "high_risk_count": 2,
    "stale_scan_count": 1,
    "insights": [
      {
        "severity": "critical",
        "title": "Repositories below baseline",
        "explanation": "2 repositories have latest scores below 70.",
        "next_action": "Open the lowest ranked repository and address high severity findings first."
      }
    ]
  }
}
```

### Notifications

#### `GET /api/v1/organizations/{organization_id}/notifications`

Lists recent in-app notifications for the actor and org.

#### `POST /api/v1/organizations/{organization_id}/notifications/{notification_id}/read`

Marks one notification as read.

#### `GET /api/v1/organizations/{organization_id}/notification-preferences`

Returns actor notification preferences. Defaults to `in_app`.

#### `PUT /api/v1/organizations/{organization_id}/notification-preferences`

Updates actor preferences. Supported channels: `in_app`, `webhook`, `email`.

---

### Operational

### Auth

#### `GET /api/v1/auth/github/login`

Redirects browsers to GitHub OAuth. Pass `?format=json` to receive the
authorization URL in an API envelope instead.

#### `GET /api/v1/auth/github/callback`

Exchanges the GitHub OAuth code, fetches the real GitHub user from
`https://api.github.com/user`, stores the user/session, and sets the
`repocompass_session` cookie.

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
| `notification_query_failed` | 500 | Notification query failed |
| `notification_preference_write_failed` | 500 | Notification preference write failed |
