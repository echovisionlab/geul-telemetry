import { describe, expect, it } from "vitest";

import {
  buildLegalPolicyCreatedAuditRecord,
  buildLegalPolicyDeletedAuditRecord,
  buildLegalPolicyLifecycleAuditRecord,
  buildLegalPolicyShareLinkAuditRecord,
  buildMapThemeCreatedAuditRecord,
  buildMapThemeDeletedAuditRecord,
  buildMapThemeContentUpdatedAuditRecord,
  buildSiteSettingsUpdatedAuditRecord,
  type AuditMetadata,
} from "../audit.ts";
import { validateAuditRecord, type AuditRecord } from "../records.ts";
import { validateSettingsAuditRecord } from "./settings.ts";

const metadata: AuditMetadata = {
  audit_id: "00000000-0000-4000-8000-000000000001",
  occurred_at: "2026-08-09T03:04:05Z",
  request_id: "018f47a2-8a3d-4e17-9d42-6f12c89b1234",
  actor_kind: "member",
  actor_member_id: "member-1",
};

function record(attributes: Partial<AuditRecord>): AuditRecord {
  return {
    ...metadata,
    action: "site_settings.updated",
    target_type: "site_settings",
    target_id: "1",
    changed_fields: ["site_title"],
    ...attributes,
  } as AuditRecord;
}

describe("settings audit", () => {
  it("builds each site and map theme variant", () => {
    expect(
      buildSiteSettingsUpdatedAuditRecord(metadata, [
        "site_title",
        "primary_color",
        "site_title",
      ]).changed_fields,
    ).toEqual(["primary_color", "site_title"]);
    expect(buildMapThemeCreatedAuditRecord(metadata, "theme-1").action).toBe(
      "map_theme.created",
    );
    expect(buildMapThemeDeletedAuditRecord(metadata, "theme-1").action).toBe(
      "map_theme.deleted",
    );
    expect(
      buildMapThemeContentUpdatedAuditRecord(metadata, "theme-1")
        .changed_fields,
    ).toEqual(["content"]);
  });

  it("enforces site settings target, canonical changed fields, and attributes", () => {
    for (const invalid of [
      record({ target_id: "2" }),
      record({ target_type: "map_theme" }),
      record({ changed_fields: [] }),
      record({ changed_fields: ["site_title", "primary_color"] }),
      record({ changed_fields: ["unknown"] }),
      record({ changed_fields: ["site_title"], email: "extra@example.test" }),
    ]) {
      expect(() => validateAuditRecord(invalid as AuditRecord)).toThrow(
        TypeError,
      );
    }
  });

  it("enforces map theme target, content mutation, and no-extra boundaries", () => {
    const content = {
      ...metadata,
      action: "map_theme.updated",
      target_type: "map_theme",
      target_id: "theme-1",
      changed_fields: ["content"],
    } as AuditRecord;
    expect(() => validateAuditRecord(content)).not.toThrow();
    for (const valid of [
      record({
        action: "map_theme.created",
        target_type: "map_theme",
        target_id: "theme-1",
        changed_fields: undefined,
      }),
      record({
        action: "map_theme.deleted",
        target_type: "map_theme",
        target_id: "theme-1",
        changed_fields: undefined,
      }),
    ]) {
      expect(() => validateAuditRecord(valid)).not.toThrow();
    }
    for (const invalid of [
      { ...content, target_id: "" },
      { ...content, changed_fields: [] },
      { ...content, changed_fields: ["content", "status"] },
      { ...content, version_id: "extra" },
      { ...content, email: "extra@example.test" },
      {
        ...content,
        actor_kind: "system",
        actor_member_id: undefined,
        actor_service: "geul-collab",
      },
      record({
        action: "map_theme.created",
        target_type: "map_theme",
        target_id: "theme-1",
        changed_fields: ["status"],
      }),
    ]) {
      expect(() => validateAuditRecord(invalid as AuditRecord)).toThrow(
        TypeError,
      );
    }
    expect(() =>
      validateSettingsAuditRecord({
        ...content,
        actor_kind: "system",
        actor_member_id: undefined,
        actor_service: "geul-collab",
      }),
    ).toThrow("map theme content mutation requires member actor");
  });

  it("builds legal policy identity, lifecycle, and share-link variants", () => {
    expect(
      buildLegalPolicyCreatedAuditRecord(metadata, "policy-1", "terms", 1),
    ).toMatchObject({ action: "legal_policy.created", policy_type: "terms" });
    expect(
      buildLegalPolicyDeletedAuditRecord(metadata, "policy-1", "privacy", 2),
    ).toMatchObject({ action: "legal_policy.deleted", policy_type: "privacy" });
    expect(
      buildLegalPolicyShareLinkAuditRecord(
        metadata,
        "policy-1",
        "terms",
        1,
        "created",
        "share-1",
      ),
    ).toMatchObject({ changed_fields: ["share_links"] });
    for (const [previous, next] of [
      ["draft", "scheduled"],
      ["scheduled", "draft"],
      ["draft", "active"],
      ["scheduled", "active"],
      ["active", "archived"],
    ] as const) {
      expect(() =>
        buildLegalPolicyLifecycleAuditRecord(
          metadata,
          "policy-1",
          "privacy",
          2,
          ["effective_at", "status"],
          previous,
          next,
          "2026-09-01T00:00:00Z",
        ),
      ).not.toThrow();
    }
  });

  it("rejects invalid legal policy identity, share-link, and lifecycle states", () => {
    const base = record({
      action: "legal_policy.updated",
      target_type: "legal_policy",
      target_id: "policy-1",
      changed_fields: ["status"],
      policy_type: "terms",
      version_number: 1,
      previous_state: "draft",
      new_state: "scheduled",
    });
    const lifecycle = {
      ...base,
      changed_fields: ["effective_at", "status"],
      previous_state: "draft" as const,
      new_state: "scheduled" as const,
      effective_at: "2026-09-01T00:00:00Z",
    };
    for (const invalid of [
      { ...base, policy_type: "other" },
      { ...base, version_number: 0 },
      { ...base, version_number: 1.5 },
      { ...base, version_number: Number.MAX_SAFE_INTEGER + 1 },
      { ...base, version_number: "1" as unknown as number },
      { ...base, target_type: "map_theme" },
      { ...base, item_id: "extra" },
      {
        ...base,
        changed_fields: ["share_links"],
        item_operation: undefined,
        item_id: "share-1",
      },
      {
        ...base,
        changed_fields: ["share_links"],
        item_operation: "updated",
        item_id: "share-1",
      },
      {
        ...base,
        changed_fields: ["share_links"],
        item_operation: "created",
        item_id: "",
      },
      { ...lifecycle, effective_at: undefined },
      { ...lifecycle, previous_state: "draft", new_state: "draft" },
      { ...lifecycle, previous_state: "archived", new_state: "active" },
      { ...lifecycle, changed_fields: [] },
      { ...lifecycle, changed_fields: ["status", "unknown"] },
      { ...lifecycle, email: "extra@example.test" },
    ]) {
      expect(() => validateAuditRecord(invalid as AuditRecord)).toThrow(
        TypeError,
      );
    }
    expect(() =>
      validateAuditRecord({
        ...base,
        changed_fields: ["share_links"],
        previous_state: undefined,
        new_state: undefined,
        item_operation: "deleted",
        item_id: "share-1",
      }),
    ).not.toThrow();
  });

  it("enforces legal policy actors by reviewed update variant", () => {
    const lifecycle = record({
      action: "legal_policy.updated",
      target_type: "legal_policy",
      target_id: "policy-1",
      changed_fields: ["status"],
      policy_type: "terms",
      version_number: 1,
      previous_state: "draft" as const,
      new_state: "scheduled" as const,
      actor_kind: "member" as const,
      actor_member_id: "member-1",
      actor_service: undefined,
    });
    const dueLifecycle = {
      ...lifecycle,
      actor_kind: "system" as const,
      actor_member_id: undefined,
      actor_service: "geul-backend",
    };
    const shareLink = {
      ...lifecycle,
      changed_fields: ["share_links"],
      previous_state: undefined,
      new_state: undefined,
      item_operation: "created" as const,
      item_id: "share-1",
    };
    expect(() => validateAuditRecord(lifecycle)).not.toThrow();
    expect(() => validateAuditRecord(dueLifecycle)).not.toThrow();
    expect(() => validateAuditRecord(shareLink)).not.toThrow();
    expect(() =>
      validateSettingsAuditRecord({
        ...lifecycle,
        actor_kind: "system",
        actor_member_id: undefined,
        actor_service: "geul-collab",
      }),
    ).toThrow(
      "legal policy lifecycle requires member or geul-backend system actor",
    );
    for (const invalid of [
      {
        ...lifecycle,
        actor_kind: "system" as const,
        actor_member_id: undefined,
        actor_service: "geul-collab",
      },
      {
        ...shareLink,
        actor_kind: "system" as const,
        actor_member_id: undefined,
        actor_service: "geul-backend",
      },
    ]) {
      expect(() => validateAuditRecord(invalid)).toThrow(TypeError);
    }
  });

  it("returns false for non-settings actions and applies system actor policy", () => {
    expect(
      validateSettingsAuditRecord(
        record({
          action: "member.updated",
          target_type: "member",
          target_id: "member-1",
        }),
      ),
    ).toBe(false);
    expect(() =>
      validateAuditRecord({
        ...record({}),
        actor_kind: "system",
        actor_member_id: undefined,
        actor_service: "geul-backend",
      }),
    ).toThrow("cannot use system actor");
  });

  it("preserves generic changed-field canonicalization after module dispatch", () => {
    const member = record({
      action: "member.updated",
      target_type: "member",
      target_id: "member-1",
      changed_fields: ["nickname", ""],
      nickname: "Member",
    });
    expect(() => validateAuditRecord(member)).toThrow("invalid value");
    expect(() =>
      validateAuditRecord({
        ...member,
        changed_fields: ["nickname", "nickname"],
      }),
    ).toThrow("sorted and unique");
  });
});
