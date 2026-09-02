import { describe, expect, it } from "vitest";

import {
  buildArtistCreatedAuditRecord,
  buildArtistDeletedAuditRecord,
  buildArtistGalleryAuditRecord,
  buildArtistLifecycleAuditRecord,
  buildArtistParticipantAuditRecord,
  buildArtistShareLinkAuditRecord,
  buildLabelCreatedAuditRecord,
  buildLabelDeletedAuditRecord,
  buildLabelLifecycleAuditRecord,
  buildLabelLogoAuditRecord,
  buildLabelParticipantAuditRecord,
  buildLabelShareLinkAuditRecord,
  type AuditMetadata,
} from "../audit.ts";
import { validateAuditRecord, type AuditRecord } from "../records.ts";
import { validateArtistLabelAuditRecord } from "./artist-label.ts";

const metadata: AuditMetadata = {
  audit_id: "00000000-0000-4000-8000-000000000001",
  occurred_at: "2026-08-09T03:04:05Z",
  request_id: "018f47a2-8a3d-4e17-9d42-6f12c89b1234",
  actor_kind: "member",
  actor_member_id: "member-1",
};
const collabMetadata: AuditMetadata = {
  audit_id: metadata.audit_id,
  occurred_at: metadata.occurred_at,
  actor_kind: "system",
  actor_service: "geul-collab",
};

function record(attributes: Partial<AuditRecord> = {}): AuditRecord {
  return {
    ...metadata,
    action: "artist.updated",
    target_type: "artist",
    target_id: "artist-1",
    changed_fields: ["gallery"],
    file_ids: ["file-1"],
    ...attributes,
  } as AuditRecord;
}

describe("artist and label audit", () => {
  it("builds every reviewed lifecycle and update variant", () => {
    const records = [
      buildArtistCreatedAuditRecord(metadata, "artist-1"),
      buildArtistLifecycleAuditRecord(
        metadata,
        "artist-1",
        "draft",
        "published",
      ),
      buildArtistGalleryAuditRecord(metadata, "artist-1", ["file-2", "file-1"]),
      buildArtistParticipantAuditRecord(
        metadata,
        "artist-1",
        "member-2",
        "none",
        "owner",
      ),
      buildArtistShareLinkAuditRecord(
        metadata,
        "artist-1",
        "share-1",
        "created",
      ),
      buildArtistDeletedAuditRecord(metadata, "artist-1"),
      buildLabelCreatedAuditRecord(metadata, "label-1"),
      buildLabelLifecycleAuditRecord(metadata, "label-1", "published", "draft"),
      buildLabelParticipantAuditRecord(
        metadata,
        "label-1",
        "member-2",
        "manager",
        "none",
      ),
      buildLabelLogoAuditRecord(
        metadata,
        "label-1",
        "light",
        "added",
        "file-1",
      ),
      buildLabelLogoAuditRecord(
        metadata,
        "label-1",
        "dark",
        "removed",
        "file-2",
      ),
      buildLabelShareLinkAuditRecord(metadata, "label-1", "share-1", "deleted"),
      buildLabelDeletedAuditRecord(metadata, "label-1"),
    ];

    expect(records.map(({ action }) => action)).toEqual([
      "artist.created",
      "artist.updated",
      "artist.updated",
      "artist.updated",
      "artist.updated",
      "artist.deleted",
      "label.created",
      "label.updated",
      "label.updated",
      "label.updated",
      "label.updated",
      "label.updated",
      "label.deleted",
    ]);
    expect(records[2].file_ids).toEqual(["file-2", "file-1"]);
  });

  it("accepts explicit empty gallery and exact actor authority", () => {
    expect(() => validateAuditRecord(record({ file_ids: [] }))).not.toThrow();
    expect(() =>
      validateAuditRecord(
        record({
          action: "label.updated",
          target_type: "label",
          target_id: "label-1",
          changed_fields: ["status"],
          file_ids: undefined,
          previous_state: "draft",
          new_state: "published",
          actor_kind: "system",
          actor_member_id: undefined,
          actor_service: "geul-collab",
        }),
      ),
    ).toThrow("cannot use system actor");
  });

  it("rejects no-ops, invalid bindings, PII, extras, and wrong targets", () => {
    for (const invalid of [
      record({ target_type: "label" }),
      record({ target_id: "" }),
      record({ changed_fields: [], file_ids: undefined }),
      record({ changed_fields: ["gallery"], file_ids: undefined }),
      record({ changed_fields: ["gallery"], file_ids: [" bad"] }),
      record({ changed_fields: ["gallery"], file_ids: ["file-1", "file-1"] }),
      {
        ...record({
          changed_fields: ["status"],
          file_ids: undefined,
          previous_state: "draft",
          new_state: "published",
        }),
        ...collabMetadata,
      },
      {
        ...record({
          changed_fields: ["gallery"],
          file_ids: ["file-1"],
        }),
        ...collabMetadata,
      },
      {
        ...record({
          action: "label.updated",
          target_type: "label",
          target_id: "label-1",
          changed_fields: ["logo"],
          file_ids: undefined,
          asset_slot: "light",
          collection_operation: "added",
          asset_id: "file-1",
        }),
        ...collabMetadata,
      },
      record({
        changed_fields: ["gallery"],
        file_ids: [],
        email: "private@example.test",
      }),
      record({
        changed_fields: ["status"],
        file_ids: undefined,
        previous_state: "draft",
        new_state: "draft",
      }),
      record({
        changed_fields: ["participants"],
        file_ids: undefined,
        subject_member_id: "member-2",
        previous_relationship: "owner",
        new_relationship: "owner",
      }),
      record({
        changed_fields: ["participants"],
        file_ids: undefined,
        subject_member_id: "member-2",
        previous_relationship: "author" as never,
        new_relationship: "manager",
      }),
      record({
        changed_fields: ["share_links"],
        file_ids: undefined,
        item_operation: "updated",
        item_id: "share-1",
      }),
      record({
        action: "label.updated",
        target_type: "label",
        target_id: "label-1",
        changed_fields: ["logo"],
        file_ids: undefined,
        asset_slot: "light",
        asset_id: "file-1",
      }),
      record({
        action: "label.updated",
        target_type: "label",
        target_id: "label-1",
        changed_fields: ["unreviewed"],
        file_ids: undefined,
      }),
      record({
        action: "label.updated",
        target_type: "label",
        target_id: "label-1",
        changed_fields: ["logo"],
        file_ids: undefined,
        asset_slot: "light",
        collection_operation: "added",
        asset_id: "",
      }),
      record({
        action: "label.updated",
        target_type: "label",
        target_id: "label-1",
        changed_fields: ["logo"],
        file_ids: undefined,
        asset_slot: "light",
        collection_operation: "added",
        asset_id: "file-1",
        item_id: "share-1",
      }),
      record({ action: "artist.created", changed_fields: ["gallery"] }),
    ])
      expect(() =>
        validateArtistLabelAuditRecord(invalid as AuditRecord),
      ).toThrow(TypeError);
  });

  it("leaves unrelated catalog actions to their owner", () => {
    expect(
      validateArtistLabelAuditRecord(
        record({ action: "member.updated", target_type: "member" }),
      ),
    ).toBe(false);
  });
});
