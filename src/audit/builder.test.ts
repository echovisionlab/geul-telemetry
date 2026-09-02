import { describe, expect, it } from "vitest";

import { buildAuditRecord } from "./builder.ts";
import type { AuditMetadata } from "./types.ts";

const metadata: AuditMetadata = {
  audit_id: "00000000-0000-4000-8000-000000000001",
  occurred_at: "2026-08-09T03:04:05Z",
  request_id: "018f47a2-8a3d-4e17-9d42-6f12c89b1234",
  actor_kind: "member",
  actor_member_id: "member-1",
};

describe("internal audit builder", () => {
  it("keeps metadata and correlation outside semantic attributes", () => {
    const record = buildAuditRecord(
      metadata,
      {
        action: "site_settings.updated",
        target_type: "site_settings",
        target_id: "1",
      },
      {
        changed_fields: ["site_title"],
        // Semantic attributes cannot override the request envelope.
        // @ts-expect-error AuditMetadata keys are intentionally excluded.
        audit_id: "00000000-0000-4000-8000-000000000002",
      },
    );
    expect(record.audit_id).toBe(metadata.audit_id);
  });
});
