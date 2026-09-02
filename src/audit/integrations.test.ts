import { readFile } from "node:fs/promises";
import { describe, expect, it } from "vitest";

import {
  buildEmailSuppressionReleasedAuditRecord,
  buildMailAdapterConfigUpdatedAuditRecord,
  buildMailAdapterCreatedAuditRecord,
  buildMailAdapterDeletedAuditRecord,
  buildTranslationProviderConfigUpdatedAuditRecord,
  buildTranslationProviderCreatedAuditRecord,
  buildTranslationProviderDeletedAuditRecord,
  buildTranslationSettingsUpdatedAuditRecord,
  TRANSLATION_SETTINGS_FIELDS,
  type AuditMetadata,
} from "../audit.ts";
import { validateAuditRecord, type AuditRecord } from "../records.ts";
import { AUDIT_RECORD_ATTRIBUTE_NAMES } from "./attributes.ts";
import { validateIntegrationAuditRecord } from "./integrations.ts";

const metadata: AuditMetadata = {
  audit_id: "00000000-0000-4000-8000-000000000001",
  occurred_at: "2026-08-09T03:04:05Z",
  request_id: "018f47a2-8a3d-4e17-9d42-6f12c89b1234",
  actor_kind: "member",
  actor_member_id: "member-1",
};
function record(attributes: Partial<AuditRecord> = {}): AuditRecord {
  return {
    ...metadata,
    action: "mail_adapter.updated",
    target_type: "mail_adapter",
    target_id: "adapter-1",
    changed_fields: ["name"],
    ...attributes,
  } as AuditRecord;
}

describe("Platform Integrations domain audit", () => {
  it("matches the cross-language translation settings field fixture", async () => {
    const fixture = JSON.parse(
      await readFile(
        new URL(
          "../../fixtures/translation-settings-fields.json",
          import.meta.url,
        ),
        "utf8",
      ),
    ) as string[];
    expect(fixture).toEqual(TRANSLATION_SETTINGS_FIELDS);
  });

  it("builds every approved action with semantic variants", () => {
    const records = [
      buildMailAdapterCreatedAuditRecord(metadata, "adapter-1"),
      buildMailAdapterConfigUpdatedAuditRecord(metadata, "adapter-1", [
        "config",
      ]),
      buildMailAdapterDeletedAuditRecord(metadata, "adapter-1"),
      buildEmailSuppressionReleasedAuditRecord(metadata, "suppression-1"),
      buildTranslationSettingsUpdatedAuditRecord(metadata, [
        "default_locale",
        "protected_terms",
      ]),
      buildTranslationProviderCreatedAuditRecord(metadata, "provider-1"),
      buildTranslationProviderConfigUpdatedAuditRecord(metadata, "provider-1", [
        "type",
      ]),
      buildTranslationProviderDeletedAuditRecord(metadata, "provider-1"),
    ];
    expect(records.map(({ action }) => action)).toEqual([
      "mail_adapter.created",
      "mail_adapter.updated",
      "mail_adapter.deleted",
      "email_suppression.updated",
      "translation_settings.updated",
      "translation_provider.created",
      "translation_provider.updated",
      "translation_provider.deleted",
    ]);
    expect(records[3]).toEqual(
      expect.objectContaining({
        changed_fields: ["status"],
        previous_state: "active",
        new_state: "released",
      }),
    );
    expect(records[4].target_id).toBe("1");
  });

  it("accepts only exact integration fields, suppression release, and singleton settings", () => {
    for (const valid of [
      record({
        action: "mail_adapter.updated",
        target_type: "mail_adapter",
        changed_fields: ["active", "config"],
      }),
      record({
        action: "email_suppression.updated",
        target_type: "email_suppression",
        target_id: "suppression-1",
        changed_fields: ["status"],
        previous_state: "active",
        new_state: "released",
      }),
      record({
        action: "translation_settings.updated",
        target_type: "translation_settings",
        target_id: "1",
        changed_fields: ["default_locale", "protected_terms"],
      }),
      record({
        action: "translation_provider.updated",
        target_type: "translation_provider",
        target_id: "provider-1",
        changed_fields: ["priority", "type"],
      }),
    ])
      expect(() => validateAuditRecord(valid)).not.toThrow();
  });

  it("rejects no-ops, generic extras, malformed suppression, and system actors", () => {
    const retiredTranslationSettingsFields = [
      "debounce_seconds",
      "english_fallback_enabled",
      "machine_generated_public_serve",
      "stale_english_enabled",
      "stale_exact_enabled",
    ];
    for (const invalid of [
      record({ changed_fields: [] }),
      record({ changed_fields: ["scopes", "name"] }),
      record({
        action: "email_suppression.updated",
        target_type: "email_suppression",
        target_id: "suppression-1",
        changed_fields: ["status"],
        previous_state: "active",
        new_state: "released",
        email: "member@example.test",
      }),
      record({
        action: "email_suppression.updated",
        target_type: "email_suppression",
        target_id: "suppression-1",
        changed_fields: ["status"],
        previous_state: "released",
        new_state: "active",
        email: "member@example.test",
      }),
      record({
        action: "translation_settings.updated",
        target_type: "translation_settings",
        target_id: "settings-1",
        changed_fields: ["default_locale"],
      }),
      ...retiredTranslationSettingsFields.map((changedField) =>
        record({
          action: "translation_settings.updated",
          target_type: "translation_settings",
          target_id: "1",
          changed_fields: [changedField],
        }),
      ),
      record({
        action: "translation_provider.updated",
        target_type: "translation_provider",
        target_id: "provider-1",
        changed_fields: ["unknown"],
      }),
      record({
        actor_kind: "system",
        actor_member_id: undefined,
        actor_service: "geul-backend",
      }),
    ])
      expect(() => validateAuditRecord(invalid as AuditRecord)).toThrow(
        TypeError,
      );
  });

  it("rejects every attribute outside the suppression release allowlist", () => {
    const suppression = record({
      action: "email_suppression.updated",
      target_type: "email_suppression",
      target_id: "suppression-1",
      changed_fields: ["status"],
      previous_state: "active",
      new_state: "released",
    });
    const allowed = new Set(["changed_fields", "previous_state", "new_state"]);
    for (const attribute of AUDIT_RECORD_ATTRIBUTE_NAMES) {
      if (allowed.has(attribute)) continue;
      expect(() =>
        validateIntegrationAuditRecord({
          ...suppression,
          [attribute]: "extra",
        }),
      ).toThrow(TypeError);
    }
    expect(() =>
      validateAuditRecord({
        ...suppression,
        unexpected_attribute: "extra",
      } as AuditRecord),
    ).toThrow(TypeError);
  });

  it("leaves unrelated actions to their owning validator", () => {
    expect(
      validateIntegrationAuditRecord(
        record({ action: "member.updated", target_type: "member" }),
      ),
    ).toBe(false);
  });
});
