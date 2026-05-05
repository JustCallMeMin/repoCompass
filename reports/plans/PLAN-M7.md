# M7 Completion Plan: Open-Source Maturity

## Summary
Finish all `T7-001` through `T7-040` on one long-lived branch: `milestone/7-open-source-maturity`. Current `milestone/7-ci-hardening` only covers part of M7 release/CI maturity, mainly `T7-032` partial and `T7-033` partial. Treat it as prerequisite work, not full M7.

Branch default:
- If PR #16 is merged: sync `main`, create `milestone/7-open-source-maturity` from `main`.
- If PR #16 is not merged: base `milestone/7-open-source-maturity` from `milestone/7-ci-hardening`, then final PR will include CI-hardening plus remaining M7.
- Use one Lore commit per task or tightly coupled task pair.
- Open one final M7 PR after `T7-040` passes.

## Implementation Waves

### Wave 1: Contributor Entry Path (`T7-001` to `T7-009`)
- Define OSS readiness goals and success criteria in docs.
- Run a real clone/setup audit and capture friction points.
- Add Start Here guide, OSS-first README rewrite, local setup guide, architecture walkthrough, codebase map, testing guide, and contributor checklist.
- Keep README concise; move long explanations into `docs/`.
- Ensure all commands match real `Makefile`, Docker, backend, frontend, migration, and test workflows.

### Wave 2: Extension Docs And Examples (`T7-010` to `T7-016`)
- Document analyzer, provider, and renderer extension contracts.
- Add minimal runnable examples for analyzer/provider/renderer.
- Add fixture-backed tests or golden tests where examples produce output.
- Build docs navigation around: getting started, concepts, architecture, extensions, testing, operations, releases.
- Do not invent a plugin marketplace; examples stay internal and copyable.

### Wave 3: Testing Maturity (`T7-017` to `T7-021`)
- Expand small deterministic fixtures for Go, Node.js, and Python ecosystems only where current analyzers can meaningfully inspect them.
- Add report golden tests for important JSON/Markdown outputs.
- Document golden update policy.
- Add compatibility matrix mapping analyzers to ecosystems and fixtures.
- Keep fixtures lightweight; no real dependency installs inside fixtures.

### Wave 4: Community Workflow (`T7-022` to `T7-028`)
- Add GitHub issue templates: bug report, feature request, analyzer request.
- Add PR template using the contributor checklist.
- Document label taxonomy: type, area, priority, difficulty, `good first issue`, `help wanted`.
- Add good-first-issue guide and seed issue backlog under docs/planning instead of creating fake GitHub issues directly.
- Add contribution workflow and maintainer triage guide.

### Wave 5: Release Maturity (`T7-029` to `T7-033`)
- Add versioning policy, changelog policy, `CHANGELOG.md`, and release checklist.
- Harden release path beyond current CI:
  - keep `.github/workflows/ci.yml`
  - add release workflow only if repository can validate artifacts without secrets
  - build backend CLI/server artifacts
  - validate artifact smoke: `help`, `scan` fixture path where supported
- Document required secrets if any, but avoid requiring secrets for normal CI.
- Do not introduce deployment automation beyond local/product runtime unless required by release validation.

### Wave 6: Demo Readiness (`T7-034` to `T7-037`)
- Add public demo fixture set: one good repo and one intentionally problematic repo.
- Add public GitHub repository candidate list with license/rate-limit notes.
- Add 5-10 minute demo script covering CLI/API/dashboard path.
- Add screenshots only if they can be generated from current product UI and kept accurate.

### Wave 7: Community Health And Final Review (`T7-038` to `T7-040`)
- Add community health files: `LICENSE`, `CODE_OF_CONDUCT.md`, `SECURITY.md`, `SUPPORT.md`.
- Add minimal governance notes: maintainer responsibilities, review/merge expectations, release owner expectations.
- Run OSS readiness review:
  - clone-from-scratch flow
  - quickstart
  - backend tests/vet
  - frontend lint/build/audit
  - Docker runtime smoke
  - extension examples
  - release checklist dry-run
  - good-first-issue docs review
- Fix any failures before final PR.

## Public Interfaces / Files
- New/updated docs: Start Here, setup, architecture walkthrough, codebase map, testing guide, extension guides, contribution workflow, release docs, demo docs, governance docs.
- GitHub workflow surface: issue templates, PR template, CI/release workflow updates.
- Test assets: expanded fixtures, golden files, compatibility matrix.
- Community health: license, security, support, conduct.
- No billing, auth, SaaS launch, Redis, plugin marketplace, or enterprise scope.

## Test Plan
- Backend:
  - `cd backend && go test ./...`
  - `cd backend && go vet ./...`
- Frontend:
  - `cd frontend && npm ci`
  - `cd frontend && npm run lint`
  - `cd frontend && npm audit --audit-level=moderate`
  - `cd frontend && npm run build`
- Docker:
  - `docker compose config`
  - `make docker-build`
  - `make docker-up`
  - health check API and frontend
  - run local fixture scan
  - run public GitHub scan if network is available
  - `make docker-down`
- Docs/repo hygiene:
  - `git diff --check`
  - verify links manually or with a lightweight link check if available
  - verify issue/PR templates render as Markdown
  - verify no secrets in fixtures/docs/examples

## Assumptions
- M7 should finish as one milestone PR, not six smaller PRs, unless review gets too large.
- PR #16 CI-hardening remains useful baseline but does not complete M7.
- Example modules should be runnable and tested, but not exposed as a formal plugin system.
- Public demo must be deterministic and not depend on paid services or private secrets.
