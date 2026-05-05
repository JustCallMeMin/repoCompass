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
