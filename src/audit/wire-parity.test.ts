import { readFileSync } from "node:fs";
import { fileURLToPath } from "node:url";

import { describe, expect, it } from "vitest";

import {
  buildMapThemeContentUpdatedAuditRecord,
  buildPostFileBlockDownloadPolicyAuditRecord,
  buildPostParticipantAuditRecord,
  type AuditMetadata,
} from "../audit.ts";
import { validateAuditRecord, type AuditRecord } from "../records.ts";

const metadata: AuditMetadata = {
  audit_id: "00000000-0000-4000-8000-000000000001",
  occurred_at: "2026-08-09T03:04:05Z",
  actor_kind: "member",
  actor_member_id: "member-1",
};

type Fixture = {
  readonly case: string;
  readonly variant: string;
  readonly action: string;
  readonly target_type: string;
  readonly target_id: string;
  readonly actor_kind: "anonymous" | "member" | "system";
  readonly actor_service?: string;
  readonly corrected_shape?: boolean;
  readonly attributes: Record<string, unknown>;
};

describe("domain audit wire parity", () => {
  it("matches the shared Go/TypeScript corrected-shape fixture", () => {
    const fixture = JSON.parse(
      readFileSync(
        fileURLToPath(
          new URL(
            "../../fixtures/domain-audit-wire-parity.json",
            import.meta.url,
          ),
        ),
        "utf8",
      ),
    ) as readonly Fixture[];
    const records: readonly AuditRecord[] = [
      buildPostParticipantAuditRecord(
        metadata,
        "post-1",
        "member-2",
        "author",
        "collaborator",
      ),
      buildPostFileBlockDownloadPolicyAuditRecord(
        metadata,
        "post-1",
        "block-1",
        "file-1",
        "disabled",
        "restricted",
        ["segment-1"],
        ["segment-2"],
      ),
      buildMapThemeContentUpdatedAuditRecord(metadata, "theme-1"),
    ];
    const corrected = fixture.filter((entry) => entry.corrected_shape);
    expect(records).toHaveLength(corrected.length);
    for (const [index, expected] of corrected.entries()) {
      const wire = JSON.parse(JSON.stringify(records[index])) as Record<
        string,
        unknown
      >;
      expect(wire, expected.case).toMatchObject({
        action: expected.action,
        target_type: expected.target_type,
        target_id: expected.target_id,
      });
      for (const key of [
        "audit_id",
        "occurred_at",
        "action",
        "target_type",
        "target_id",
        "request_id",
        "trace_id",
        "span_id",
        "actor_kind",
        "actor_member_id",
        "actor_service",
      ])
        delete wire[key];
      expect(wire, expected.case).toEqual(expected.attributes);
    }
  });
});

describe("domain audit variant manifest", () => {
  it("defines exact validated wire attributes for every reviewed variant", () => {
    const fixture = JSON.parse(
      readFileSync(
        fileURLToPath(
          new URL(
            "../../fixtures/domain-audit-wire-parity.json",
            import.meta.url,
          ),
        ),
        "utf8",
      ),
    ) as readonly Fixture[];
    expect(fixture.length).toBeGreaterThan(0);
    for (const entry of fixture) {
      expect(entry.variant, entry.case).not.toBe("");
      const record = {
        audit_id: metadata.audit_id,
        occurred_at: metadata.occurred_at,
        action: entry.action,
        target_type: entry.target_type,
        target_id: entry.target_id,
        actor_kind: entry.actor_kind,
        ...(entry.actor_kind === "member"
          ? { actor_member_id: "member-1" }
          : {}),
        ...(entry.actor_kind === "system"
          ? { actor_service: entry.actor_service ?? "geul-backend" }
          : {}),
        ...entry.attributes,
      } as AuditRecord;
      validateAuditRecord(record);
      const wire = JSON.parse(JSON.stringify(record)) as Record<
        string,
        unknown
      >;
      for (const key of [
        "audit_id",
        "occurred_at",
        "action",
        "target_type",
        "target_id",
        "request_id",
        "trace_id",
        "span_id",
        "actor_kind",
        "actor_member_id",
        "actor_service",
      ])
        delete wire[key];
      expect(wire, entry.case).toEqual(entry.attributes);
    }
  });
});
