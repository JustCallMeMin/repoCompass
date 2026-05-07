export const apiBaseUrl =
  process.env.NEXT_PUBLIC_REPOCOMPASS_API_URL?.replace(/\/$/, "") ?? "http://localhost:8080";

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
  recommendations: Array<{ title: string; action: string; rationale: string; priority: string }>;
};

export type MetricPoint = {
  scan_id: string;
  metric_key: string;
  value: number;
  captured_at: string;
  completed_at?: string;
  repository_id: string;
};

async function request<T>(path: string, init?: RequestInit): Promise<T> {
  const response = await fetch(`${apiBaseUrl}${path}`, {
    ...init,
    headers: {
      "Content-Type": "application/json",
      "X-User-Id": currentUserId,
      ...(init?.headers ?? {}),
    },
  });
  if (!response.ok) {
    const body = await response.json().catch(() => null);
    const message = body?.error?.message ?? `Request failed with ${response.status}`;
    throw new Error(message);
  }
  return (await response.json()) as T;
}

export function createScan(input: ScanRequest) {
  return request<ScanResponse>("/api/v1/scans", {
    method: "POST",
    body: JSON.stringify(input),
  });
}

export function listRepositoryScans(repositoryId: string) {
  return request<ScanSummary[]>(`/api/v1/repositories/${encodeURIComponent(repositoryId)}/scans`);
}

export function listScanFindings(scanId: string) {
  return request<FindingDetail[]>(`/api/v1/scans/${encodeURIComponent(scanId)}/findings`);
}

export function listRepositoryMetrics(repositoryId: string) {
  return request<MetricPoint[]>(`/api/v1/repositories/${encodeURIComponent(repositoryId)}/metrics`);
}

// --- Organization types ---

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

export type Repository = {
  id: string;
  name: string;
  full_name: string;
  organization_id: string;
  status: string;
  provider: string;
  url: string;
  default_branch?: string;
};

export const currentUserId = process.env.NEXT_PUBLIC_REPOCOMPASS_USER_ID ?? "mock_user";

export function canManageOrganization(role?: Membership["role"]) {
  return role === "owner" || role === "admin";
}

// --- Org API functions ---

export function listOrganizations() {
  return request<{ data: Organization[] }>("/api/v1/organizations");
}

export function getOrganization(orgId: string) {
  return request<{ data: Organization }>(`/api/v1/organizations/${encodeURIComponent(orgId)}`);
}

export function listMembers(orgId: string) {
  return request<{ data: Membership[] }>(`/api/v1/organizations/${encodeURIComponent(orgId)}/members`);
}

export function addMember(orgId: string, userId: string, role: Membership["role"]) {
  return request<{ status: string }>(`/api/v1/organizations/${encodeURIComponent(orgId)}/members`, {
    method: "POST",
    body: JSON.stringify({ user_id: userId, role }),
  });
}

export function listOrgRepositories(orgId: string) {
  return request<{ data: Repository[] }>(`/api/v1/organizations/${encodeURIComponent(orgId)}/repositories`);
}

export function listPolicies(orgId: string) {
  return request<{ data: Policy[] }>(`/api/v1/organizations/${encodeURIComponent(orgId)}/policies`);
}

export function savePolicy(orgId: string, policy: Omit<Policy, "created_at" | "updated_at" | "organization_id">) {
  return request<{ status: string }>(`/api/v1/organizations/${encodeURIComponent(orgId)}/policies`, {
    method: "POST",
    body: JSON.stringify(policy),
  });
}

export function getOrgInsights(orgId: string) {
  return request<{ data: OrgInsights }>(`/api/v1/organizations/${encodeURIComponent(orgId)}/insights`);
}
