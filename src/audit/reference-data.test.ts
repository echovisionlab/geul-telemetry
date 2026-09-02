import { describe, expect, it } from "vitest";

import {
  buildCategoryCreatedAuditRecord,
  buildCategoryDeletedAuditRecord,
  buildCategoryMetadataUpdatedAuditRecord,
  buildFormatCreatedAuditRecord,
  buildFormatDeletedAuditRecord,
  buildFormatMetadataUpdatedAuditRecord,
  buildGenreCreatedAuditRecord,
  buildGenreDeletedAuditRecord,
  buildGenreMetadataUpdatedAuditRecord,
  buildStyleCreatedAuditRecord,
  buildStyleDeletedAuditRecord,
  buildStyleMetadataUpdatedAuditRecord,
  buildTagCreatedAuditRecord,
  buildTagDeletedAuditRecord,
  buildTagMetadataUpdatedAuditRecord,
  type AuditMetadata,
} from "../audit.ts";
import { validateAuditRecord, type AuditRecord } from "../records.ts";
import { validateReferenceDataAuditRecord } from "./reference-data.ts";

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
    action: "category.updated",
    target_type: "category",
    target_id: "category-1",
    changed_fields: ["name"],
    ...attributes,
  } as AuditRecord;
}

describe("reference-data audit", () => {
  it("builds all fifteen explicit lifecycle and metadata variants", () => {
    const records = [
      buildCategoryCreatedAuditRecord(metadata, "category-1"),
      buildCategoryMetadataUpdatedAuditRecord(metadata, "category-1", [
        "slug",
        "name",
      ]),
      buildCategoryDeletedAuditRecord(metadata, "category-1"),
      buildTagCreatedAuditRecord(metadata, "tag-1"),
      buildTagMetadataUpdatedAuditRecord(metadata, "tag-1", ["slug", "name"]),
      buildTagDeletedAuditRecord(metadata, "tag-1"),
      buildGenreCreatedAuditRecord(metadata, "genre-1"),
      buildGenreMetadataUpdatedAuditRecord(metadata, "genre-1", [
        "description",
      ]),
      buildGenreDeletedAuditRecord(metadata, "genre-1"),
      buildStyleCreatedAuditRecord(metadata, "style-1"),
      buildStyleMetadataUpdatedAuditRecord(metadata, "style-1", ["name"]),
      buildStyleDeletedAuditRecord(metadata, "style-1"),
      buildFormatCreatedAuditRecord(metadata, "format-1"),
      buildFormatMetadataUpdatedAuditRecord(metadata, "format-1", [
        "slug",
        "name",
      ]),
      buildFormatDeletedAuditRecord(metadata, "format-1"),
    ];
    expect(records.map(({ action }) => action)).toEqual([
      "category.created",
      "category.updated",
      "category.deleted",
      "tag.created",
      "tag.updated",
      "tag.deleted",
      "genre.created",
      "genre.updated",
      "genre.deleted",
      "style.created",
      "style.updated",
      "style.deleted",
      "format.created",
      "format.updated",
      "format.deleted",
    ]);
    expect(records[1].changed_fields).toEqual(["name", "slug"]);
  });

  it("accepts only documented canonical metadata fields", () => {
    for (const valid of [
      record({ changed_fields: ["description", "name", "slug"] }),
      record({
        action: "tag.updated",
        target_type: "tag",
        target_id: "tag-1",
        changed_fields: ["name", "slug"],
      }),
      record({
        action: "format.updated",
        target_type: "format",
        target_id: "format-1",
        changed_fields: ["name"],
      }),
    ])
      expect(() => validateAuditRecord(valid)).not.toThrow();
  });

  it("rejects no-ops, extra attributes, and mismatched action targets", () => {
    for (const invalid of [
      record({ target_id: "" }),
      record({ target_type: "tag" }),
      record({ changed_fields: [] }),
      record({ changed_fields: ["slug", "name"] }),
      record({ changed_fields: ["unknown"] }),
      record({
        action: "tag.updated",
        target_type: "tag",
        target_id: "tag-1",
        changed_fields: ["description"],
      }),
      record({
        action: "format.updated",
        target_type: "format",
        target_id: "format-1",
        changed_fields: ["name"],
        email: "extra@example.test",
      }),
      record({
        action: "genre.created",
        target_type: "genre",
        target_id: "genre-1",
        changed_fields: ["name"],
      }),
      record({
        action: "style.deleted",
        target_type: "style",
        target_id: "style-1",
        item_ids: [],
      }),
    ])
      expect(() => validateAuditRecord(invalid)).toThrow(TypeError);
  });

  it("rejects anonymous and system actors and does not claim other actions", () => {
    expect(() =>
      validateAuditRecord(record({ actor_kind: "anonymous" } as never)),
    ).toThrow("anonymous");
    expect(() =>
      validateAuditRecord(
        record({
          actor_kind: "system",
          actor_member_id: undefined,
          actor_service: "geul-backend",
        }),
      ),
    ).toThrow("cannot use system actor");
    expect(
      validateReferenceDataAuditRecord(
        record({
          action: "member.updated",
          target_type: "member",
          target_id: "member-1",
        }),
      ),
    ).toBe(false);
  });
});
