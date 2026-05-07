export const apiBaseUrl =
  process.env.NEXT_PUBLIC_REPOCOMPASS_API_URL?.replace(/\/$/, "") ?? "http://localhost:8080";

export const currentUserId = process.env.NEXT_PUBLIC_REPOCOMPASS_USER_ID ?? "mock_user";
export const currentOrganizationId =
  process.env.NEXT_PUBLIC_REPOCOMPASS_ORG_ID ?? "00000000-0000-0000-0000-000000000000";

export type ApiError = {
  code: string;
  message: string;
};

export type ApiEnvelope<T> = {
  data: T;
  meta: { request_id?: string };
  error: ApiError | null;
};

export type ScanRequest =
  | { source_type: "local"; path: string }
  | { source_type: "github"; url: string };

export type ScanResponse = {
  scan_id: string;
  repository_id: string;
  snapshot_id: string;
  status: string;
  analyzers_processed: number;
  finding_count: number;
  assessment_score: number;
};

export type Repository = {
  id: string;
  name: string;
  owner_name?: string;
  full_name: string;
  organization_id: string;
  status: string;
  provider: string;
  url: string;
  local_path?: string;
  default_branch?: string;
};

export type ScanSummary = {
  scan_id: string;
  repository_id: string;
  snapshot_id: string;
  status: string;
  started_at?: string;
  completed_at?: string;
  score: number;
  label: string;
  finding_count: number;
};

export type ScanDetail = {
  id: string;
  snapshot_id: string;
  status: string;
  start_time?: string;
  end_time?: string;
  error_details?: string;
};

export type FindingDetail = {
  id: string;
  scan_id: string;
  rule_id: string;
  analyzer_id: string;
  severity: string;
  title: string;
  message: string;
  category: string;
  status: string;
  evidence: Array<{ type: string; message: string; path?: string; value?: string }>;
  recommendations: Recommendation[];
};

export type Recommendation = {
  title: string;
  action: string;
  rationale: string;
  priority: string;
  finding_id?: string;
  severity?: string;
  category?: string;
};

export type Assessment = {
  overall_score: number;
  label: string;
  finding_count: number;
  severity_counts?: Record<string, number>;
  category_scores?: Record<string, number>;
  category_breakdown?: Record<string, unknown>;
};

export type MetricPoint = {
  scan_id: string;
  metric_key: string;
  value: number;
  captured_at: string;
  completed_at?: string;
  repository_id: string;
};

export type ReportMetadata = {
  id: number;
  scan_id: string;
  format: string;
  content_type: string;
  metadata?: Record<string, unknown>;
  created_at: string;
};

export type SessionInfo = {
  user_id: string;
  organization_id: string;
};

export type Organization = {
  id: string;
  name: string;
  created_at: string;
  updated_at: string;
};

export type Membership = {
  organization_id: string;
  user_id: string;
  role: "owner" | "admin" | "member" | "viewer";
  created_at: string;
  updated_at: string;
};

export type Policy = {
  id: string;
  organization_id: string;
  name: string;
  rules: Record<string, unknown>;
  created_at: string;
  updated_at: string;
};

export type OrgInsights = {
  organization_id: string;
  average_score: number;
  total_repositories: number;
  total_scans: number;
};

export class RepoCompassApiError extends Error {
  constructor(
    message: string,
    public readonly code: string,
    public readonly status: number,
  ) {
    super(message);
  }
}

async function request<T>(path: string, init?: RequestInit): Promise<T> {
  const response = await fetch(`${apiBaseUrl}${path}`, {
    ...init,
    headers: {
      "Content-Type": "application/json",
      "X-User-Id": currentUserId,
      "X-Organization-Id": currentOrganizationId,
      ...(init?.headers ?? {}),
    },
  });
  const body = await response.json().catch(() => null);
  const error = body?.error as ApiError | null | undefined;
  if (!response.ok || error) {
    throw new RepoCompassApiError(
      error?.message ?? `Request failed with ${response.status}`,
      error?.code ?? "request_failed",
      response.status,
    );
  }
  if (body && typeof body === "object" && "data" in body) {
    return (body as ApiEnvelope<T>).data;
  }
  return body as T;
}

export function canManageOrganization(role?: Membership["role"]) {
  return role === "owner" || role === "admin";
}

export function createScan(input: ScanRequest) {
  return request<ScanResponse>("/api/v1/scans", { method: "POST", body: JSON.stringify(input) });
}

export function listRepositories() {
  return request<Repository[]>("/api/v1/repositories");
}

export function getRepository(repositoryId: string) {
  return request<Repository>(`/api/v1/repositories/${encodeURIComponent(repositoryId)}`);
}

export function createRepositoryScan(repositoryId: string) {
  return request<ScanResponse>(`/api/v1/repositories/${encodeURIComponent(repositoryId)}/scans`, {
    method: "POST",
  });
}

export function listRepositoryScans(repositoryId: string) {
  return request<ScanSummary[]>(`/api/v1/repositories/${encodeURIComponent(repositoryId)}/scans`);
}

export function getScan(scanId: string) {
  return request<ScanDetail>(`/api/v1/scans/${encodeURIComponent(scanId)}`);
}

export function listScanFindings(scanId: string) {
  return request<FindingDetail[]>(`/api/v1/scans/${encodeURIComponent(scanId)}/findings`);
}

export function getScanAssessment(scanId: string) {
  return request<Assessment>(`/api/v1/scans/${encodeURIComponent(scanId)}/assessment`);
}

export function listScanReports(scanId: string) {
  return request<ReportMetadata[]>(`/api/v1/scans/${encodeURIComponent(scanId)}/reports`);
}

export function listRepositoryMetrics(repositoryId: string) {
  return request<MetricPoint[]>(`/api/v1/repositories/${encodeURIComponent(repositoryId)}/metrics`);
}

export function getSession() {
  return request<SessionInfo>("/api/v1/auth/session");
}

export function listOrganizations() {
  return request<Organization[]>("/api/v1/organizations");
}

export function getOrganization(orgId: string) {
  return request<Organization>(`/api/v1/organizations/${encodeURIComponent(orgId)}`);
}

export function listMembers(orgId: string) {
  return request<Membership[]>(`/api/v1/organizations/${encodeURIComponent(orgId)}/members`);
}

export function addMember(orgId: string, userId: string, role: Membership["role"]) {
  return request<{ status: string }>(`/api/v1/organizations/${encodeURIComponent(orgId)}/members`, {
    method: "POST",
    body: JSON.stringify({ user_id: userId, role }),
  });
}

export function listOrgRepositories(orgId: string) {
  return request<Repository[]>(`/api/v1/organizations/${encodeURIComponent(orgId)}/repositories`);
}

export function listPolicies(orgId: string) {
  return request<Policy[]>(`/api/v1/organizations/${encodeURIComponent(orgId)}/policies`);
}

export function savePolicy(orgId: string, policy: Omit<Policy, "created_at" | "updated_at" | "organization_id">) {
  return request<{ status: string }>(`/api/v1/organizations/${encodeURIComponent(orgId)}/policies`, {
    method: "POST",
    body: JSON.stringify(policy),
  });
}

export function getOrgInsights(orgId: string) {
  return request<OrgInsights>(`/api/v1/organizations/${encodeURIComponent(orgId)}/insights`);
}
