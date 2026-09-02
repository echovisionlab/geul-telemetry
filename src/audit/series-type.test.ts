import { describe, expect, it } from "vitest";

import {
  buildPostSeriesCreatedAuditRecord,
  buildPostSeriesDeletedAuditRecord,
  buildPostSeriesFeaturedImageAuditRecord,
  buildPostSeriesLifecycleAuditRecord,
  buildPostSeriesManagerAuditRecord,
  buildPostSeriesMembershipAuditRecord,
  buildPostSeriesOrderAuditRecord,
  buildPostSeriesSourceMetadataAuditRecord,
  buildProgramEventTypeConfigUpdatedAuditRecord,
  buildProgramEventTypeCreatedAuditRecord,
  buildProgramEventTypeDeletedAuditRecord,
  type AuditMetadata,
} from "../audit.ts";
import { validateAuditRecord, type AuditRecord } from "../records.ts";
import { validateSeriesTypeAuditRecord } from "./series-type.ts";

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
    action: "post_series.updated",
    target_type: "post_series",
    target_id: "series-1",
    changed_fields: ["slug"],
    ...attributes,
  } as AuditRecord;
}

describe("post series and program event type audit", () => {
  it("builds the complete reviewed roots and update variants", () => {
    const records = [
      buildPostSeriesCreatedAuditRecord(metadata, "series-1"),
      buildPostSeriesSourceMetadataAuditRecord(metadata, "series-1", [
        "source_copy",
        "slug",
      ]),
      buildPostSeriesLifecycleAuditRecord(
        metadata,
        "series-1",
        "draft",
        "published",
      ),
      buildPostSeriesManagerAuditRecord(
        metadata,
        "series-1",
        "member-2",
        "none",
        "manager",
      ),
      buildPostSeriesMembershipAuditRecord(
        metadata,
        "series-1",
        "post-1",
        "series-0",
        "series-1",
      ),
      buildPostSeriesOrderAuditRecord(metadata, "series-1", [
        "post-2",
        "post-1",
        "post-2",
      ]),
      buildPostSeriesOrderAuditRecord(metadata, "series-1", []),
      buildPostSeriesFeaturedImageAuditRecord(
        metadata,
        "series-1",
        "added",
        "file-1",
      ),
      buildPostSeriesDeletedAuditRecord(metadata, "series-1"),
      buildProgramEventTypeCreatedAuditRecord(metadata, "type-1"),
      buildProgramEventTypeConfigUpdatedAuditRecord(metadata, "type-1", [
        "status",
        "slug",
      ]),
      buildProgramEventTypeDeletedAuditRecord(metadata, "type-1"),
    ];
    expect(records.map(({ action }) => action)).toEqual([
      "post_series.created",
      "post_series.updated",
      "post_series.updated",
      "post_series.updated",
      "post_series.updated",
      "post_series.updated",
      "post_series.updated",
      "post_series.updated",
      "post_series.deleted",
      "program_event_type.created",
      "program_event_type.updated",
      "program_event_type.deleted",
    ]);
    expect(records[1].changed_fields).toEqual(["slug", "source_copy"]);
    expect(records[5].post_ids).toEqual(["post-2", "post-1", "post-2"]);
    expect(records[6].post_ids).toEqual([]);
  });

  it("accepts only reviewed variants, including intentional empty ordered posts", () => {
    for (const valid of [
      record({ changed_fields: ["slug", "source_copy"] }),
      record({
        changed_fields: ["status"],
        previous_state: "published",
        new_state: "draft",
      }),
      record({
        changed_fields: ["managers"],
        subject_member_id: "member-2",
        previous_relationship: "manager",
        new_relationship: "none",
      }),
      record({
        changed_fields: ["posts"],
        subject_post_id: "post-1",
        previous_series_id: "series-1",
        new_series_id: "",
      }),
      record({ changed_fields: ["post_order"], post_ids: [] }),
      record({
        changed_fields: ["featured_image"],
        collection_operation: "removed",
        file_id: "file-1",
      }),
      record({
        action: "program_event_type.updated",
        target_type: "program_event_type",
        target_id: "type-1",
        changed_fields: [
          "requires_place",
          "requires_stream_url",
          "slug",
          "sort_order",
          "status",
        ],
      }),
    ])
      expect(() => validateAuditRecord(valid)).not.toThrow();
  });

  it("rejects no-ops, locale labels, PII, extras, wrong targets, and system actors", () => {
    for (const invalid of [
      record({ target_id: "" }),
      record({ target_type: "program_event_type" }),
      record({ changed_fields: [] }),
      record({ changed_fields: ["source_copy", "slug"] }),
      record({ changed_fields: ["locale_label"] }),
      record({
        changed_fields: ["status"],
        previous_state: "draft",
        new_state: "draft",
      }),
      record({
        changed_fields: ["managers"],
        subject_member_id: "member-2",
        previous_relationship: "owner",
        new_relationship: "manager",
      }),
      record({
        changed_fields: ["managers"],
        subject_member_id: "member-2",
        previous_relationship: "manager",
        new_relationship: "manager",
      }),
      record({
        changed_fields: ["posts"],
        subject_post_id: "post-1",
        previous_series_id: "series-1",
        new_series_id: "series-1",
      }),
      record({ changed_fields: ["post_order"] }),
      record({ changed_fields: ["post_order"], post_ids: [" bad"] }),
      record({ changed_fields: ["featured_image"], file_id: "file-1" }),
      record({
        changed_fields: ["featured_image"],
        collection_operation: "added",
        file_id: "",
        email: "private@example.test",
      }),
      record({ changed_fields: ["slug"], previous_state: "draft" }),
      record({ action: "post_series.created", changed_fields: ["slug"] }),
      record({
        actor_kind: "system",
        actor_member_id: undefined,
        actor_service: "geul-backend",
      }),
      record({
        action: "program_event_type.updated",
        target_type: "program_event_type",
        target_id: "type-1",
        changed_fields: ["locale_label"],
      }),
    ])
      expect(() => validateAuditRecord(invalid as AuditRecord)).toThrow(
        TypeError,
      );
  });

  it("leaves unrelated catalog actions to their owner", () => {
    expect(
      validateSeriesTypeAuditRecord(
        record({ action: "artist.updated", target_type: "artist" }),
      ),
    ).toBe(false);
  });
});
