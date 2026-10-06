import { describe, expect, it } from "vitest";
import { buildPageConfigurationAuditRecord } from "./page.ts";
import type { AuditMetadata } from "./types.ts";
import { validateAuditRecord } from "../records.ts";

const metadata: AuditMetadata = {
  audit_id: "00000000-0000-4000-8000-000000000001",
  occurred_at: "2026-08-09T03:04:05Z",
  actor_kind: "member",
  actor_member_id: "member-1",
};

describe("Page configuration audit", () => {
  it("accepts policy updates and canonicalizes duplicate changed fields", () => {
    const record = buildPageConfigurationAuditRecord(metadata, "page-1", [
      "slug",
      "access_policy",
      "access_policy",
    ]);
    expect(record.changed_fields).toEqual(["access_policy", "slug"]);
    expect(() => validateAuditRecord(record)).not.toThrow();
  });

  it("continues rejecting unknown configuration fields", () => {
    const record = buildPageConfigurationAuditRecord(metadata, "page-1", [
      "access_policy",
    ]);
    expect(() =>
      validateAuditRecord({
        ...record,
        changed_fields: ["access_policy", "unknown"],
      }),
    ).toThrow("changed_fields");
  });
});
