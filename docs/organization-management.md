# Organization Management

## 1. Organization Management Scope, User Jobs, and Success Criteria (T6-001)

### Purpose
Transition RepoCompass from a single-tenant (personal) utility to a multi-tenant platform capable of serving teams and organizations, without losing the simplicity of the single-user flow.

### User Jobs
- **As an engineering manager**, I want to group my team's repositories together so I can see aggregate security and health scores across our entire portfolio.
- **As a team lead**, I want to enforce standard policies (e.g., minimum score thresholds) across all our repositories, so I don't have to check them manually.
- **As a developer**, I want to scan my personal repositories without being forced into complex organizational setups.

### Success Criteria
- Organizations can be created, and users can be invited with specific roles.
- Existing single-user repositories are seamlessly migrated to a default "Personal" organization.
- Cross-organization data leakage is impossible at the database and API levels.
- The UI provides an aggregate view of organizational health.

## 2. Information Architecture and Route Map (T6-002)

### Global Navigation
- **Org Switcher**: Dropdown in the header allowing users to switch between their accessible organizations.

### Organization Routes (`/organizations/[id]/*`)
- `/organizations/[id]`: Overview dashboard (aggregate metrics, risk summary).
- `/organizations/[id]/repositories`: List of repositories owned by the organization.
- `/organizations/[id]/policies`: Policy management UI.
- `/organizations/[id]/members`: Member management UI.
- `/organizations/[id]/settings`: General organization settings (renaming, deletion).

### Repository Routes (Preserving M5 Flows)
- `/repositories/[id]`: Maintained for backward compatibility, automatically contextualized to the current active organization via session/header.

## 3. Minimal Role Model (T6-003)

The RBAC system utilizes static, predefined roles to manage access:

| Role | Description | Permissions |
| ---- | ----------- | ----------- |
| **Owner** | Full administrative control over the organization. | Read/Write everything, Manage members, Delete organization. |
| **Admin** | Operational management. | Read/Write repositories, Manage policies, Manage members (except Owners). |
| **Member** | Standard contributor. | Read repositories, Trigger scans. Cannot manage policies or members. |
| **Viewer** | Read-only access. | View repositories and scan results only. Cannot trigger scans or edit anything. |

## 4. Organization Data Model (T6-004)

Row-level multi-tenancy is enforced using the `organization_id` foreign key.

- **`organizations`**: The core tenant entity.
  - `id` (UUID, PK)
  - `name` (String)
- **`organization_memberships`**: Mapping users to organizations with roles.
  - `organization_id` (UUID, FK)
  - `user_id` (String)
  - `role` (Enum)
- **`repositories`**: Modified to belong to an organization.
  - `organization_id` (UUID, FK, NOT NULL)

## 5. Organization Configuration and Policy Model (T6-005)

- **`organization_configurations`**: 1:1 mapping with organizations for global settings.
  - `organization_id` (UUID, FK, PK)
  - `settings` (JSONB)
- **`policies`**: Organization-specific rules evaluated during repository scans.
  - `id` (UUID, PK)
  - `organization_id` (UUID, FK)
  - `name` (String)
  - `rules` (JSONB)
