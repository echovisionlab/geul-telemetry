import { describe, expect, it } from "vitest";

import {
  buildProgramEventChildAuditRecord,
  buildProgramEventChildOrderAuditRecord,
  buildProgramEventCreatedAuditRecord,
  buildProgramEventDeletedAuditRecord,
  buildProgramEventLifecycleAuditRecord,
  buildProgramEventMetadataAuditRecord,
  buildProgramEventPosterAuditRecord,
  buildProgramEventSeriesCreatedAuditRecord,
  buildProgramEventSeriesDeletedAuditRecord,
  buildProgramEventSeriesLifecycleAuditRecord,
  buildProgramEventSeriesMetadataAuditRecord,
  buildProgramEventSeriesPosterAuditRecord,
  type AuditMetadata,
} from "../audit.ts";
import { validateAuditRecord, type AuditRecord } from "../records.ts";
import { validateProgramAuditRecord } from "./program.ts";

const metadata: AuditMetadata = {
  audit_id: "00000000-0000-4000-8000-000000000001",
  occurred_at: "2026-08-09T03:04:05Z",
  request_id: "018f47a2-8a3d-4e17-9d42-6f12c89b1234",
  actor_kind: "member",
  actor_member_id: "member-1",
};
function eventRecord(attributes: Partial<AuditRecord> = {}): AuditRecord {
  return {
    ...metadata,
    action: "program_event.updated",
    target_type: "program_event",
    target_id: "event-1",
    changed_fields: ["slug"],
    ...attributes,
  } as AuditRecord;
}
function seriesRecord(attributes: Partial<AuditRecord> = {}): AuditRecord {
  return {
    ...metadata,
    action: "program_event_series.updated",
    target_type: "program_event_series",
    target_id: "series-1",
    changed_fields: ["slug"],
    ...attributes,
  } as AuditRecord;
}

describe("program event and program event series audit", () => {
  it("builds all reviewed roots and variants, including empty child order", () => {
    const records = [
      buildProgramEventCreatedAuditRecord(metadata, "event-1"),
      buildProgramEventMetadataAuditRecord(metadata, "event-1", [
        "type",
        "slug",
      ]),
      buildProgramEventPosterAuditRecord(
        metadata,
        "event-1",
        "added",
        "file-1",
      ),
      buildProgramEventChildAuditRecord(
        metadata,
        "event-1",
        "media",
        "media-1",
        "created",
      ),
      buildProgramEventChildOrderAuditRecord(
        metadata,
        "event-1",
        "credits",
        [],
      ),
      buildProgramEventLifecycleAuditRecord(
        metadata,
        "event-1",
        "draft",
        "published",
      ),
      buildProgramEventDeletedAuditRecord(metadata, "event-1"),
      buildProgramEventSeriesCreatedAuditRecord(metadata, "series-1"),
      buildProgramEventSeriesMetadataAuditRecord(metadata, "series-1", [
        "title",
        "slug",
      ]),
      buildProgramEventSeriesPosterAuditRecord(
        metadata,
        "series-1",
        "removed",
        "file-1",
      ),
      buildProgramEventSeriesLifecycleAuditRecord(
        metadata,
        "series-1",
        "published",
        "draft",
      ),
      buildProgramEventSeriesDeletedAuditRecord(metadata, "series-1"),
    ];
    expect(records.map(({ action }) => action)).toEqual([
      "program_event.created",
      "program_event.updated",
      "program_event.updated",
      "program_event.updated",
      "program_event.updated",
      "program_event.updated",
      "program_event.deleted",
      "program_event_series.created",
      "program_event_series.updated",
      "program_event_series.updated",
      "program_event_series.updated",
      "program_event_series.deleted",
    ]);
    expect(records[1].changed_fields).toEqual(["slug", "type"]);
    expect(records[4].item_ids).toEqual([]);
  });

  it("accepts exact event and series variants", () => {
    for (const valid of [
      eventRecord({
        changed_fields: ["poster"],
        collection_operation: "removed",
        file_id: "file-1",
      }),
      eventRecord({
        changed_fields: ["media"],
        item_operation: "updated",
        item_id: "media-1",
      }),
      eventRecord({ changed_fields: ["credits"], item_ids: [] }),
      eventRecord({
        changed_fields: ["status"],
        previous_state: "published",
        new_state: "archived",
      }),
      eventRecord({ changed_fields: ["all_day", "starts_at"] }),
      seriesRecord({ changed_fields: ["description", "title"] }),
      seriesRecord({
        changed_fields: ["poster"],
        collection_operation: "added",
        file_id: "file-1",
      }),
      seriesRecord({
        changed_fields: ["status"],
        previous_state: "draft",
        new_state: "published",
      }),
    ])
      expect(() => validateAuditRecord(valid)).not.toThrow();
  });

  it("rejects no-ops, PII, extra fields, wrong targets, invalid child shapes, and disallowed system actors", () => {
    for (const invalid of [
      eventRecord({ target_id: "" }),
      eventRecord({ target_type: "program_event_series" }),
      eventRecord({ changed_fields: ["media"], item_operation: "created" }),
      eventRecord({ changed_fields: ["media"], item_id: "media-1" }),
      eventRecord({ changed_fields: ["credits"], item_ids: [" bad"] }),
      eventRecord({
        changed_fields: ["status"],
        previous_state: "draft",
        new_state: "draft",
      }),
      eventRecord({
        changed_fields: ["poster"],
        collection_operation: "added",
        file_id: "file-1",
        email: "private@example.test",
      }),
      eventRecord({ changed_fields: ["title"] }),
      eventRecord({
        action: "program_event.created",
        changed_fields: ["slug"],
      }),
      seriesRecord({
        changed_fields: ["status"],
        previous_state: "published",
        new_state: "published",
      }),
      seriesRecord({ changed_fields: ["poster"], file_id: "file-1" }),
      seriesRecord({ changed_fields: ["title"], item_id: "extra" }),
      seriesRecord({
        actor_kind: "system",
        actor_member_id: undefined,
        actor_service: "geul-backend",
      }),
    ])
      expect(() => validateAuditRecord(invalid as AuditRecord)).toThrow(
        TypeError,
      );
  });

  it("rejects system writes for program event updates until a reviewed producer exists", () => {
    expect(() =>
      validateAuditRecord(
        eventRecord({
          actor_kind: "system",
          actor_member_id: undefined,
          actor_service: "geul-backend",
        }),
      ),
    ).toThrow("system actor");
    expect(() =>
      validateAuditRecord(
        eventRecord({
          actor_kind: "system",
          actor_member_id: undefined,
          actor_service: "editor-collab",
        }),
      ),
    ).toThrow(TypeError);
  });

  it("rejects system actors for program event updates", () => {
    const systemEvent = eventRecord({
      actor_kind: "system",
      actor_member_id: undefined,
      actor_service: "geul-collab",
    });
    expect(() =>
      validateProgramAuditRecord({ ...systemEvent, changed_fields: [] }),
    ).toThrow("does not allow system actor");
    expect(() => validateProgramAuditRecord(systemEvent)).toThrow(
      "does not allow system actor",
    );
  });

  it("leaves unrelated catalog actions to their owner", () => {
    expect(
      validateProgramAuditRecord(
        eventRecord({ action: "artist.updated", target_type: "artist" }),
      ),
    ).toBe(false);
  });
});
