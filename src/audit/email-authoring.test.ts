import { describe, expect, it } from "vitest";

import {
  buildEmailEventMappingTemplateAuditRecord,
  buildEmailLayoutCreatedAuditRecord,
  buildEmailLayoutDeletedAuditRecord,
  buildEmailLayoutMetadataAuditRecord,
  buildEmailTemplateCreatedAuditRecord,
  buildEmailTemplateDeletedAuditRecord,
  buildEmailTemplateLayoutRelationAuditRecord,
  buildEmailTemplateMetadataAuditRecord,
  type AuditMetadata,
} from "../audit.ts";
import { validateAuditRecord, type AuditRecord } from "../records.ts";
import { validateEmailAuthoringAuditRecord } from "./email-authoring.ts";

const metadata: AuditMetadata = {
  audit_id: "00000000-0000-4000-8000-000000000001",
  occurred_at: "2026-08-09T03:04:05Z",
  request_id: "018f47a2-8a3d-4e17-9d42-6f12c89b1234",
  actor_kind: "member",
  actor_member_id: "member-1",
};

function templateRecord(attributes: Partial<AuditRecord> = {}): AuditRecord {
  return {
    ...metadata,
    action: "email_template.updated",
    target_type: "email_template",
    target_id: "template-1",
    changed_fields: ["name"],
    ...attributes,
  } as AuditRecord;
}
function layoutRecord(attributes: Partial<AuditRecord> = {}): AuditRecord {
  return {
    ...metadata,
    action: "email_layout.updated",
    target_type: "email_layout",
    target_id: "layout-1",
    changed_fields: ["name"],
    ...attributes,
  } as AuditRecord;
}
function mappingRecord(attributes: Partial<AuditRecord> = {}): AuditRecord {
  return {
    ...metadata,
    action: "email_event_mapping.updated",
    target_type: "email_event_mapping",
    target_id: "welcome",
    changed_fields: ["template"],
    event_name: "welcome",
    previous_item_id: "template-1",
    item_id: "template-2",
    ...attributes,
  } as AuditRecord;
}

describe("email authoring audit", () => {
  it("builds every reviewed root and variant", () => {
    const records = [
      buildEmailTemplateCreatedAuditRecord(metadata, "template-1"),
      buildEmailTemplateMetadataAuditRecord(metadata, "template-1", [
        "name",
        "active",
      ]),
      buildEmailTemplateLayoutRelationAuditRecord(
        metadata,
        "template-1",
        "layout-1",
        "",
      ),
      buildEmailTemplateDeletedAuditRecord(metadata, "template-1"),
      buildEmailLayoutCreatedAuditRecord(metadata, "layout-1"),
      buildEmailLayoutMetadataAuditRecord(metadata, "layout-1", ["key"]),
      buildEmailLayoutDeletedAuditRecord(metadata, "layout-1"),
      buildEmailEventMappingTemplateAuditRecord(
        metadata,
        "welcome",
        "template-1",
        "template-2",
      ),
    ];
    expect(records).toHaveLength(8);
    expect(records[1].changed_fields).toEqual(["active", "name"]);
    expect(records[2].previous_item_id).toBe("layout-1");
    expect(records[2].item_id).toBeUndefined();
  });

  it("accepts only exact reviewed variants", () => {
    for (const record of [
      templateRecord({ changed_fields: ["active", "key"] }),
      templateRecord({
        changed_fields: ["layout"],
        previous_item_id: "layout-1",
        item_id: "",
      }),
      templateRecord({
        changed_fields: ["layout"],
        previous_item_id: "",
        item_id: "layout-2",
      }),
      layoutRecord({ changed_fields: ["key", "name"] }),
      mappingRecord(),
      mappingRecord({ previous_item_id: undefined }),
      mappingRecord({ item_id: undefined }),
    ])
      expect(() => validateAuditRecord(record)).not.toThrow();
  });

  it("rejects no-ops, unreviewed variants, PII, and forbidden system actors", () => {
    for (const record of [
      templateRecord({ target_type: "email_layout" }),
      templateRecord({ changed_fields: ["target_locale"] }),
      templateRecord({ changed_fields: ["preview"] }),
      templateRecord({
        changed_fields: ["layout"],
        previous_item_id: "",
        item_id: "",
      }),
      templateRecord({
        changed_fields: ["layout"],
        previous_item_id: "layout-1",
        item_id: "layout-1",
      }),
      layoutRecord({ changed_fields: ["test"] }),
      mappingRecord({ event_name: "bad event" }),
      mappingRecord({ previous_item_id: `template-${"x".repeat(256)}` }),
      mappingRecord({ previous_item_id: " template-1" }),
      mappingRecord({ email: "private@example.test" }),
      mappingRecord({ previous_item_id: "template-2", item_id: "template-2" }),
      mappingRecord({
        changed_fields: ["template"],
        event_name: "welcome",
        previous_item_id: "template-1",
        item_id: "template-2",
        actor_kind: "system",
        actor_member_id: undefined,
        actor_service: "geul-backend",
      }),
    ])
      expect(() => validateAuditRecord(record)).toThrow();

    expect(() =>
      validateAuditRecord(
        templateRecord({
          actor_kind: "system",
          actor_member_id: undefined,
          actor_service: "geul-backend",
        }),
      ),
    ).toThrow("system actor");
    expect(() =>
      validateAuditRecord(
        templateRecord({
          actor_kind: "system",
          actor_member_id: undefined,
          actor_service: "not-backend",
        }),
      ),
    ).toThrow();
  });

  it("rejects malformed mapping actor branches", () => {
    for (const invalid of [
      {
        ...mappingRecord(),
        actor_kind: "system",
        actor_member_id: undefined,
        actor_service: "geul-collab",
      },
    ])
      expect(() =>
        validateEmailAuthoringAuditRecord(invalid as AuditRecord),
      ).toThrow(TypeError);
  });
});
