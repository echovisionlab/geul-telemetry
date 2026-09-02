import { describe, expect, it } from "vitest";

import { systemActorMayAppendAudit } from "./actor-policy.ts";
import type { AuditRecord } from "../records.ts";

function systemRecord(attributes: Partial<AuditRecord>): AuditRecord {
  return {
    audit_id: "00000000-0000-4000-8000-000000000001",
    occurred_at: "2026-08-09T03:04:05Z",
    actor_kind: "system",
    actor_service: "geul-collab",
    action: "release.updated",
    target_type: "release",
    target_id: "release-1",
    changed_fields: ["track_audio"],
    ...attributes,
  } as AuditRecord;
}

describe("system audit actor policy", () => {
  it("allows only the exact bounded Release and Campaign variants", () => {
    expect(systemActorMayAppendAudit(systemRecord({}))).toBe(true);
    expect(
      systemActorMayAppendAudit(systemRecord({ changed_fields: ["version"] })),
    ).toBe(false);
    expect(
      systemActorMayAppendAudit(
        systemRecord({
          action: "release.updated",
          target_type: "release",
          changed_fields: [undefined] as unknown as string[],
        }),
      ),
    ).toBe(false);
    expect(
      systemActorMayAppendAudit(
        systemRecord({
          action: "form.updated",
          target_type: "form",
          changed_fields: ["status"],
        }),
      ),
    ).toBe(false);
    expect(
      systemActorMayAppendAudit(
        systemRecord({
          action: "campaign.updated",
          target_type: "campaign",
          actor_service: "geul-collab",
          changed_fields: ["status"],
        }),
      ),
    ).toBe(false);
    expect(
      systemActorMayAppendAudit(
        systemRecord({
          action: "campaign.updated",
          target_type: "campaign",
          actor_service: "geul-backend",
          changed_fields: ["status"],
          new_state: "sent",
        }),
      ),
    ).toBe(true);
    expect(
      systemActorMayAppendAudit(
        systemRecord({
          action: "campaign.updated",
          target_type: "campaign",
          actor_service: "geul-backend",
          changed_fields: ["status"],
          new_state: "sending",
        }),
      ),
    ).toBe(false);
  });
});
