import { readFile } from "node:fs/promises";

import { describe, expect, it } from "vitest";

import {
  CANONICAL_JOB_KINDS,
  JOB_FAILURE_REASONS,
  JOB_KIND_MESH_OPTIMIZATION,
  parseJobFailureReason,
  parseJobKind,
} from "./job.ts";

describe("job catalog", () => {
  it("matches the cross-language fixture", async () => {
    const fixture = JSON.parse(
      await readFile(
        new URL("../fixtures/job-catalog.json", import.meta.url),
        "utf8",
      ),
    ) as { job_kind: string; failure_reasons: string[] }[];

    expect(fixture).toEqual(
      CANONICAL_JOB_KINDS.map((jobKind) => ({
        job_kind: jobKind,
        failure_reasons: JOB_FAILURE_REASONS[jobKind],
      })),
    );
    expect(fixture.map(({ job_kind }) => parseJobKind(job_kind))).toEqual(
      CANONICAL_JOB_KINDS,
    );
    for (const { job_kind, failure_reasons } of fixture) {
      for (const reason of failure_reasons) {
        expect(parseJobFailureReason(parseJobKind(job_kind), reason)).toBe(
          reason,
        );
      }
    }
  });

  it("rejects unregistered values", () => {
    expect(() => parseJobKind("projection")).toThrow(
      "unknown canonical job kind",
    );
    expect(() =>
      parseJobFailureReason(JOB_KIND_MESH_OPTIMIZATION, "failed"),
    ).toThrow("unknown failure reason");
    expect(() =>
      parseJobFailureReason("projection" as never, "failed"),
    ).toThrow("unknown canonical job kind");
  });
});
