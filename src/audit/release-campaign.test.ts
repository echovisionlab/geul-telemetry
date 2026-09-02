import { describe, expect, it } from "vitest";

import {
  buildCampaignCreatedAuditRecord,
  buildCampaignDeletedAuditRecord,
  buildCampaignDeliveryRunLifecycleAuditRecord,
  buildCampaignMetadataAuditRecord,
  buildCampaignScheduleLifecycleAuditRecord,
  buildCampaignStatusLifecycleAuditRecord,
  buildCampaignTargetLayoutAuditRecord,
  buildReleaseArtworkAuditRecord,
  buildReleaseCreatedAuditRecord,
  buildReleaseDeletedAuditRecord,
  buildReleaseLifecycleAuditRecord,
  buildReleaseMetadataAuditRecord,
  buildReleaseShareLinkAuditRecord,
  buildReleaseTrackAudioAuditRecord,
  buildReleaseTrackAuditRecord,
  buildReleaseTrackOrderAuditRecord,
  type AuditMetadata,
} from "../audit.ts";
import { validateAuditRecord, type AuditRecord } from "../records.ts";
import { validateReleaseCampaignAuditRecord } from "./release-campaign.ts";

const metadata: AuditMetadata = {
  audit_id: "00000000-0000-4000-8000-000000000001",
  occurred_at: "2026-08-09T03:04:05Z",
  request_id: "018f47a2-8a3d-4e17-9d42-6f12c89b1234",
  actor_kind: "member",
  actor_member_id: "member-1",
};
function releaseRecord(attributes: Partial<AuditRecord> = {}): AuditRecord {
  return {
    ...metadata,
    action: "release.updated",
    target_type: "release",
    target_id: "release-1",
    changed_fields: ["slug"],
    ...attributes,
  } as AuditRecord;
}
function campaignRecord(attributes: Partial<AuditRecord> = {}): AuditRecord {
  return {
    ...metadata,
    action: "campaign.updated",
    target_type: "campaign",
    target_id: "campaign-1",
    changed_fields: ["layout"],
    ...attributes,
  } as AuditRecord;
}

describe("release and campaign audit", () => {
  it("builds every reviewed root and variant", () => {
    const records = [
      buildReleaseCreatedAuditRecord(metadata, "release-1"),
      buildReleaseMetadataAuditRecord(metadata, "release-1", ["type", "slug"]),
      buildReleaseTrackAuditRecord(metadata, "release-1", "track-1", "updated"),
      buildReleaseTrackOrderAuditRecord(metadata, "release-1", []),
      buildReleaseTrackAudioAuditRecord(
        metadata,
        "release-1",
        "track-1",
        "file-1",
        "added",
      ),
      buildReleaseArtworkAuditRecord(
        metadata,
        "release-1",
        "file-1",
        "removed",
      ),
      buildReleaseLifecycleAuditRecord(
        metadata,
        "release-1",
        "draft",
        "published",
      ),
      buildReleaseShareLinkAuditRecord(
        metadata,
        "release-1",
        "link-1",
        "created",
      ),
      buildReleaseDeletedAuditRecord(metadata, "release-1"),
      buildCampaignCreatedAuditRecord(metadata, "campaign-1"),
      buildCampaignTargetLayoutAuditRecord(metadata, "campaign-1", [
        "target_mode",
        "layout",
      ]),
      buildCampaignMetadataAuditRecord(metadata, "campaign-1", ["name"]),
      buildCampaignStatusLifecycleAuditRecord(
        metadata,
        "campaign-1",
        "draft",
        "published",
      ),
      buildCampaignScheduleLifecycleAuditRecord(
        metadata,
        "campaign-1",
        "draft",
        "scheduled",
        "2026-08-10T09:00:00Z",
      ),
      buildCampaignDeliveryRunLifecycleAuditRecord(
        metadata,
        "campaign-1",
        "scheduled",
        "sending",
        "run-1",
      ),
      buildCampaignDeletedAuditRecord(metadata, "campaign-1"),
    ];
    expect(records).toHaveLength(16);
    expect(records[1].changed_fields).toEqual(["slug", "type"]);
    expect(records[3].item_ids).toEqual([]);
    expect(records[10].changed_fields).toEqual(["layout", "target_mode"]);
  });

  it("accepts exact release and campaign variants", () => {
    for (const valid of [
      releaseRecord({
        changed_fields: ["tracks"],
        item_operation: "created",
        item_id: "track-1",
      }),
      releaseRecord({ changed_fields: ["tracks"], item_ids: [] }),
      releaseRecord({
        changed_fields: ["track_audio"],
        collection_operation: "added",
        item_id: "track-1",
        file_id: "file-1",
      }),
      releaseRecord({
        changed_fields: ["artwork"],
        collection_operation: "removed",
        file_id: "file-1",
      }),
      releaseRecord({
        changed_fields: ["status"],
        previous_state: "published",
        new_state: "draft",
      }),
      releaseRecord({
        changed_fields: ["share_links"],
        item_operation: "deleted",
        item_id: "link-1",
      }),
      releaseRecord({ changed_fields: ["artists", "slug"] }),
      campaignRecord({ changed_fields: ["layout", "target_mode"] }),
      campaignRecord({
        changed_fields: ["status"],
        previous_state: "draft",
        new_state: "published",
      }),
      campaignRecord({
        changed_fields: ["schedule"],
        previous_state: "draft",
        new_state: "scheduled",
        scheduled_at: "2026-08-10T09:00:00Z",
      }),
      campaignRecord({
        changed_fields: ["delivery_run"],
        previous_state: "scheduled",
        new_state: "sending",
        item_id: "run-1",
      }),
    ])
      expect(() => validateAuditRecord(valid)).not.toThrow();
  });

  it("rejects no-ops, PII, extra attributes, malformed schedules, and wrong targets", () => {
    for (const invalid of [
      releaseRecord({ target_id: "" }),
      releaseRecord({ target_type: "campaign" }),
      releaseRecord({ changed_fields: ["unreviewed_field"] }),
      releaseRecord({ changed_fields: ["tracks"], item_id: "track-1" }),
      releaseRecord({ changed_fields: ["tracks"], item_ids: [" bad"] }),
      releaseRecord({
        changed_fields: ["track_audio"],
        collection_operation: "added",
        item_id: "track-1",
      }),
      releaseRecord({
        changed_fields: ["track_audio"],
        item_id: "track-1",
        file_id: "file-1",
      }),
      releaseRecord({ changed_fields: ["artwork"], file_id: "file-1" }),
      releaseRecord({
        changed_fields: ["status"],
        previous_state: "draft",
        new_state: "draft",
      }),
      releaseRecord({
        changed_fields: ["share_links"],
        item_operation: "updated",
        item_id: "link-1",
      }),
      releaseRecord({
        changed_fields: ["artwork"],
        collection_operation: "added",
        file_id: "file-1",
        email: "private@example.test",
      }),
      campaignRecord({
        changed_fields: ["schedule"],
        previous_state: "draft",
        new_state: "scheduled",
        scheduled_at: "not-a-time",
      }),
      campaignRecord({
        changed_fields: ["schedule"],
        previous_state: "draft",
        new_state: "scheduled",
        scheduled_at: "2026-08-10T09:00:00Z",
        scheduled_time_zone: "Asia/Seoul",
      }),
      campaignRecord({
        changed_fields: ["status"],
        previous_state: "draft",
        new_state: "draft",
      }),
      campaignRecord({
        changed_fields: ["delivery_run"],
        previous_state: "scheduled",
        new_state: "sending",
      }),
      campaignRecord({
        changed_fields: ["status"],
        previous_state: "draft",
        new_state: "published",
        scheduled_at: "2026-08-10T09:00:00Z",
      }),
      campaignRecord({
        changed_fields: ["status"],
        previous_state: "draft",
        new_state: "published",
        item_id: "run-1",
      }),
    ])
      expect(() => validateAuditRecord(invalid as AuditRecord)).toThrow(
        TypeError,
      );
  });

  it("limits system writes to their exact Release and Campaign variants", () => {
    for (const record of [
      releaseRecord(),
      releaseRecord({
        changed_fields: ["artwork"],
        file_id: "file-1",
        collection_operation: "added",
      }),
    ]) {
      expect(() =>
        validateAuditRecord({
          ...record,
          actor_kind: "system",
          actor_member_id: undefined,
          actor_service: "geul-collab",
        }),
      ).toThrow(TypeError);
    }
    expect(() =>
      validateAuditRecord({
        ...releaseRecord({
          changed_fields: ["track_audio"],
          item_id: "track-1",
          file_id: "file-1",
          collection_operation: "added",
        }),
        actor_kind: "system",
        actor_member_id: undefined,
        actor_service: "geul-backend",
      }),
    ).toThrow(TypeError);
    expect(() =>
      validateReleaseCampaignAuditRecord({
        ...campaignRecord({
          changed_fields: ["status"],
          previous_state: "sending",
          new_state: "sent",
        }),
        actor_kind: "system",
        actor_member_id: undefined,
        actor_service: "geul-backend",
      }),
    ).not.toThrow();
  });

  it("leaves unrelated catalog actions to their owner", () => {
    expect(
      validateReleaseCampaignAuditRecord(
        releaseRecord({ action: "artist.updated", target_type: "artist" }),
      ),
    ).toBe(false);
  });

  it("rejects each bounded system campaign branch", () => {
    expect(() =>
      validateReleaseCampaignAuditRecord({
        ...campaignRecord({
          changed_fields: ["status"],
          previous_state: "sending",
          new_state: "sent",
        }),
        actor_kind: "system",
        actor_member_id: undefined,
        actor_service: "geul-collab",
      }),
    ).toThrow("limited to terminal status");
    expect(() =>
      validateReleaseCampaignAuditRecord({
        ...campaignRecord({
          changed_fields: ["status"],
          previous_state: "sending",
          new_state: "sending",
        }),
        actor_kind: "system",
        actor_member_id: undefined,
        actor_service: "geul-backend",
      }),
    ).toThrow("must be terminal");
  });
});
