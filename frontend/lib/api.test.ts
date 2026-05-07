import { afterEach, describe, expect, it, vi } from "vitest";
import { RepoCompassApiError, createScan, getSession } from "./api";

afterEach(() => {
  vi.restoreAllMocks();
});

function mockFetch(body: unknown, ok = true, status = 200) {
  vi.stubGlobal(
    "fetch",
    vi.fn().mockResolvedValue({
      ok,
      status,
      json: vi.fn().mockResolvedValue(body),
    }),
  );
}

describe("api client", () => {
  it("unwraps success envelopes", async () => {
    mockFetch({
      data: {
        scan_id: "scan_1",
        repository_id: "repo_1",
        snapshot_id: "snap_1",
        status: "completed",
        analyzers_processed: 4,
        finding_count: 1,
        assessment_score: 82,
      },
      meta: { request_id: "req_1" },
      error: null,
    });

    await expect(createScan({ source_type: "local", path: "./repo" })).resolves.toMatchObject({
      scan_id: "scan_1",
      repository_id: "repo_1",
    });
  });

  it("maps error envelopes to typed errors", async () => {
    mockFetch(
      {
        data: null,
        meta: { request_id: "req_1" },
        error: { code: "unauthorized", message: "session is required" },
      },
      false,
      401,
    );

    await expect(getSession()).rejects.toMatchObject({
      code: "unauthorized",
      status: 401,
      message: "session is required",
    } satisfies Partial<RepoCompassApiError>);
  });
});
