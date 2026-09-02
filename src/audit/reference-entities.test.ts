import { describe, expect, it } from "vitest";

import {
  buildAudienceSegmentConfigUpdatedAuditRecord,
  buildAudienceSegmentCreatedAuditRecord,
  buildAudienceSegmentLifecycleUpdatedAuditRecord,
  buildClientCreatedAuditRecord,
  buildClientDeletedAuditRecord,
  buildClientLogoUpdatedAuditRecord,
  buildClientMetadataUpdatedAuditRecord,
  buildMapPlaceCreatedAuditRecord,
  buildMapPlaceDeletedAuditRecord,
  buildMapPlaceImageUpdatedAuditRecord,
  buildMapPlaceMetadataUpdatedAuditRecord,
  buildMenuCreatedAuditRecord,
  buildMenuDeletedAuditRecord,
  buildMenuSourceUpdatedAuditRecord,
  type AuditMetadata,
} from "../audit.ts";
import { validateAuditRecord, type AuditRecord } from "../records.ts";
import { validateReferenceEntityAuditRecord } from "./reference-entities.ts";

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
    action: "client.updated",
    target_type: "client",
    target_id: "client-1",
    changed_fields: ["name"],
    ...attributes,
  } as AuditRecord;
}

describe("reference entity audit", () => {
  it("builds all eleven reviewed actions with semantic variants", () => {
    const records = [
      buildClientCreatedAuditRecord(metadata, "client-1"),
      buildClientMetadataUpdatedAuditRecord(metadata, "client-1", [
        "website",
        "name",
      ]),
      buildClientLogoUpdatedAuditRecord(
        metadata,
        "client-1",
        "light",
        "added",
        "file-1",
      ),
      buildClientDeletedAuditRecord(metadata, "client-1"),
      buildMapPlaceCreatedAuditRecord(metadata, "place-1"),
      buildMapPlaceMetadataUpdatedAuditRecord(metadata, "place-1", [
        "lng",
        "lat",
      ]),
      buildMapPlaceImageUpdatedAuditRecord(
        metadata,
        "place-1",
        "removed",
        "file-1",
      ),
      buildMapPlaceDeletedAuditRecord(metadata, "place-1"),
      buildAudienceSegmentCreatedAuditRecord(metadata, "segment-1"),
      buildAudienceSegmentConfigUpdatedAuditRecord(metadata, "segment-1", [
        "segment_type",
        "name",
      ]),
      buildAudienceSegmentLifecycleUpdatedAuditRecord(
        metadata,
        "segment-1",
        "active",
        "archived",
      ),
      buildMenuCreatedAuditRecord(metadata, "menu-1"),
      buildMenuSourceUpdatedAuditRecord(metadata, "menu-1", ["name", "items"]),
      buildMenuDeletedAuditRecord(metadata, "menu-1"),
    ];
    expect(records.map(({ action }) => action)).toEqual([
      "client.created",
      "client.updated",
      "client.updated",
      "client.deleted",
      "map_place.created",
      "map_place.updated",
      "map_place.updated",
      "map_place.deleted",
      "audience_segment.created",
      "audience_segment.updated",
      "audience_segment.updated",
      "menu.created",
      "menu.updated",
      "menu.deleted",
    ]);
    expect(records[1].changed_fields).toEqual(["name", "website"]);
    expect(records[2]).toMatchObject({
      asset_slot: "light",
      collection_operation: "added",
      file_id: "file-1",
    });
    expect(records[6]).toMatchObject({
      collection_operation: "removed",
      file_id: "file-1",
    });
  });

  it("accepts only reviewed metadata, file bindings, lifecycle, and source fields", () => {
    for (const valid of [
      record({ changed_fields: ["name", "website"] }),
      record({
        action: "client.updated",
        changed_fields: ["logo"],
        asset_slot: "dark",
        collection_operation: "removed",
        file_id: "file-1",
      }),
      record({
        action: "map_place.updated",
        target_type: "map_place",
        target_id: "place-1",
        changed_fields: [
          "address",
          "address_components",
          "google_place_id",
          "lat",
          "lng",
          "name",
        ],
      }),
      record({
        action: "map_place.updated",
        target_type: "map_place",
        target_id: "place-1",
        changed_fields: ["image"],
        collection_operation: "added",
        file_id: "file-1",
      }),
      record({
        action: "audience_segment.updated",
        target_type: "audience_segment",
        target_id: "segment-1",
        changed_fields: [
          "account_roles",
          "created_after",
          "created_before",
          "description",
          "exclude_member_ids",
          "member_tag_ids",
          "name",
          "segment_type",
        ],
      }),
      record({
        action: "audience_segment.updated",
        target_type: "audience_segment",
        target_id: "segment-1",
        changed_fields: ["status"],
        previous_state: "archived",
        new_state: "active",
      }),
      record({
        action: "menu.updated",
        target_type: "menu",
        target_id: "menu-1",
        changed_fields: ["items", "name"],
      }),
    ])
      expect(() => validateAuditRecord(valid)).not.toThrow();
  });

  it("rejects no-ops, malformed bindings, target mismatches, PII, and extras", () => {
    for (const invalid of [
      record({ target_id: "" }),
      record({ target_type: "map_place" }),
      record({ changed_fields: [] }),
      record({ changed_fields: ["website", "name"] }),
      record({ changed_fields: ["unknown"] }),
      record({
        changed_fields: ["logo"],
        asset_slot: "light",
        file_id: "file-1",
      }),
      record({
        changed_fields: ["logo"],
        collection_operation: "added",
        file_id: "file-1",
      }),
      record({
        changed_fields: ["logo"],
        asset_slot: "light",
        collection_operation: "added",
        file_id: " bad",
      }),
      record({
        action: "map_place.updated",
        target_type: "map_place",
        target_id: "place-1",
        changed_fields: ["image"],
        asset_slot: "light",
        collection_operation: "added",
        file_id: "file-1",
      }),
      record({
        action: "audience_segment.updated",
        target_type: "audience_segment",
        target_id: "segment-1",
        changed_fields: ["status"],
        previous_state: "active",
        new_state: "active",
      }),
      record({
        action: "audience_segment.updated",
        target_type: "audience_segment",
        target_id: "segment-1",
        changed_fields: ["status"],
        previous_state: "active",
        new_state: "draft",
      }),
      record({
        action: "menu.updated",
        target_type: "menu",
        target_id: "menu-1",
        changed_fields: ["items"],
        email: "private@example.test",
      }),
      record({ action: "client.created", changed_fields: ["name"] }),
      record({
        action: "map_place.deleted",
        target_type: "map_place",
        target_id: "place-1",
        item_ids: [],
      }),
    ])
      expect(() => validateAuditRecord(invalid as AuditRecord)).toThrow(
        TypeError,
      );
  });

  it("rejects anonymous and system actors and leaves unrelated actions to their owner", () => {
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
      validateReferenceEntityAuditRecord(
        record({
          action: "member.updated",
          target_type: "member",
          target_id: "member-1",
        }),
      ),
    ).toBe(false);
  });
});
