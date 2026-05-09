# Security Threat Model

This document covers the organization management and product-hardening API surface.

## Scope

- Organization, membership, and policy management.
- Org-scoped repository reads, scan history, findings, metrics, and insights.
- Product HTTP API, GitHub webhook intake, and PostgreSQL persistence.
- Out of scope: production session auth, billing, multi-region deployment, and public SaaS tenancy.

## Trust Boundaries

```text
Browser / CLI -> HTTP API -> PostgreSQL
                     |
                     +-> GitHub webhook payloads
                     +-> public GitHub clone requests
```

- All HTTP requests are untrusted until decoded, scoped, and authorized.
- PostgreSQL is trusted only behind the API boundary.
- GitHub webhooks require HMAC validation when `GITHUB_WEBHOOK_SECRET` is set.
- GitHub OAuth callbacks require a persisted one-time `state`.

## Key Controls

- RBAC uses `owner`, `admin`, `member`, and `viewer` roles.
- Org reads and writes check membership through `auth.AuthorizationService`.
- Organization repository history and metrics verify repository ownership before returning data.
- Member and policy mutations require `owner` or `admin`.
- API accepts `X-User-Id` only when local-dev header auth is enabled. Product flows use GitHub OAuth sessions.
- `POST /api/v1/scans` has an in-process per-actor rate limit.
- `X-Request-ID` is propagated through request context and returned on every response.
- Audit event types exist for organization, membership, policy, and scan activity.

## STRIDE Review

| Category | Threat | Risk | M6 status |
| --- | --- | --- | --- |
| Spoofing | Caller forges local-dev `X-User-Id`. | Medium | Header auth is dev-only; production-style flow uses GitHub OAuth session cookie. |
| Spoofing | Caller forges `X-Organization-Id`. | Medium | Mitigated by membership and repository-org checks. |
| Tampering | Policy JSON malformed or unexpected. | Medium | JSONB persistence and API decode validation are present; stricter schema can be added later. |
| Repudiation | Admin denies member/policy changes. | Medium | Audit events are persisted in `audit_events`; add review UI later if needed. |
| Information disclosure | Cross-org repository history or metrics leak. | High | Mitigated by repository ownership checks before history/metric reads. |
| Denial of service | Unbounded scan creation. | High | Mitigated by per-actor in-process rate limit. |
| Denial of service | Expensive org dashboard queries. | Medium | Mitigated by M6 performance indexes and bounded limits. |
| Elevation of privilege | Member manages policies or members. | High | Mitigated by `CheckManageOrg` on policy/member writes. |
| Replay | GitHub webhook payload replay. | Medium | HMAC validates authenticity and persisted GitHub delivery IDs reject duplicates. |
| CSRF | Attacker reuses OAuth callback without a valid login attempt. | Medium | Persisted one-time OAuth state is required and consumed on callback. |

## Required M6 Findings

| ID | Severity | Finding | Status |
| --- | --- | --- | --- |
| SEC-001 | Medium | Local-dev actor identity fallback must not be treated as production auth. | Accepted with explicit `X-User-Id` contract and documented production-auth gap. |
| SEC-002 | High | Scan endpoint needs abuse protection. | Mitigated by in-process per-actor rate limiting. |
| SEC-003 | Medium | GitHub webhook replay cannot be fully prevented without durable delivery tracking. | Reduced by persisted delivery IDs and unique delivery constraint; keep HMAC required. |
| SEC-004 | Medium | Cross-org reads must be blocked. | Mitigated by repository-org checks and RBAC. |
| SEC-005 | Low | Local Docker database uses non-TLS connection. | Accepted for local runtime only. |

## Follow-Up

- Replace `X-User-Id` with real session/auth middleware before any public deployment.
- Add an audit review UI and full notification provider suite after M6 if needed.
- Replace in-memory rate limiting with shared limiter only if running multiple API replicas.
