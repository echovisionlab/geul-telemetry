import type {
  AuditCollectionOperation,
  AuditItemOperation,
  AuditRecord,
  AuditState,
} from "../records.ts";
import {
  assertOnlyAuditAttributes,
  canonicalAuditValues,
} from "./attributes.ts";
import { buildAuditRecord } from "./builder.ts";
import {
  AUDIT_ACTIONS,
  AUDIT_CATALOG,
  requireCatalogTarget,
} from "./catalog.ts";
import {
  isRelationDownloadPolicyAuditRecord,
  validateRelationDownloadPolicyAuditRecord,
} from "./relation-download-policy.ts";
import type { AuditMetadata } from "./types.ts";

type ReleaseCampaignAction = Extract<
  (typeof AUDIT_ACTIONS)[number],
  | "release.created"
  | "release.updated"
  | "release.deleted"
  | "campaign.created"
  | "campaign.updated"
  | "campaign.deleted"
>;
export type ReleaseMetadataField =
  | "artists"
  | "categories"
  | "credits"
  | "date"
  | "formats"
  | "genres"
  | "labels"
  | "links"
  | "slug"
  | "styles"
  | "type";
export type CampaignTargetLayoutField =
  "layout" | "recipient_scope" | "segment" | "target_mode";
export type CampaignMetadataField = CampaignTargetLayoutField | "name";
type ReleaseLifecycleState = Extract<AuditState, "draft" | "published">;

const actions = new Set<ReleaseCampaignAction>([
  "release.created",
  "release.updated",
  "release.deleted",
  "campaign.created",
  "campaign.updated",
  "campaign.deleted",
]);
const releaseMetadataFields = new Set<ReleaseMetadataField>([
  "artists",
  "categories",
  "credits",
  "date",
  "formats",
  "genres",
  "labels",
  "links",
  "slug",
  "styles",
  "type",
]);
const campaignFields = new Set<
  CampaignMetadataField | "delivery_run" | "schedule" | "status"
>([
  "layout",
  "name",
  "recipient_scope",
  "segment",
  "target_mode",
  "delivery_run",
  "schedule",
  "status",
]);
function buildReleaseCampaignAuditRecord(
  metadata: AuditMetadata,
  action: ReleaseCampaignAction,
  targetId: string,
  attributes: Parameters<typeof buildAuditRecord>[2] = {},
): AuditRecord {
  return buildAuditRecord(
    metadata,
    { action, target_type: AUDIT_CATALOG[action], target_id: targetId },
    attributes,
  );
}
function requireNoExtra(record: AuditRecord, allowed: readonly string[]): void {
  assertOnlyAuditAttributes(record, allowed);
}
function requireIdentifier(name: string, value: string | undefined): void {
  if (!value || value.length > 255 || value.trim() !== value)
    throw new TypeError(`${name} requires an identifier`);
}
function requireFields(
  record: AuditRecord,
  allowed: ReadonlySet<string>,
): void {
  const fields = record.changed_fields;
  if (
    !fields?.length ||
    fields.some(
      (field, i) => !allowed.has(field) || (i > 0 && fields[i - 1] >= field),
    )
  )
    throw new TypeError("invalid or non-canonical changed_fields");
}
function requireChild(record: AuditRecord): void {
  if (record.item_ids !== undefined) {
    for (const itemId of record.item_ids) requireIdentifier("item_ids", itemId);
    return requireNoExtra(record, ["changed_fields", "item_ids"]);
  }
  if (
    !record.item_id ||
    !["created", "updated", "deleted"].includes(record.item_operation ?? "")
  )
    throw new TypeError("child requires operation and item");
  requireIdentifier("item_id", record.item_id);
  requireNoExtra(record, ["changed_fields", "item_operation", "item_id"]);
}
function requireFileBinding(record: AuditRecord, field: string): void {
  if (
    record.collection_operation !== "added" &&
    record.collection_operation !== "removed"
  )
    throw new TypeError(`${field} requires collection operation`);
  requireIdentifier("file_id", record.file_id);
  requireNoExtra(record, ["changed_fields", "collection_operation", "file_id"]);
}
function requireReleaseLifecycle(record: AuditRecord): void {
  if (!(
    (record.previous_state === "draft" && record.new_state === "published") ||
    (record.previous_state === "published" && record.new_state === "draft")
  ))
    throw new TypeError(
      "release lifecycle requires draft/published transition",
    );
  requireNoExtra(record, ["changed_fields", "previous_state", "new_state"]);
}
function requireShareLink(record: AuditRecord): void {
  if (
    record.item_operation !== "created" &&
    record.item_operation !== "deleted"
  )
    throw new TypeError(
      "share link requires created or deleted item operation",
    );
  requireIdentifier("item_id", record.item_id);
  requireNoExtra(record, ["changed_fields", "item_operation", "item_id"]);
}
function validateReleaseUpdate(record: AuditRecord): void {
  if (isRelationDownloadPolicyAuditRecord(record))
    return validateRelationDownloadPolicyAuditRecord(record);
  if (record.changed_fields?.length !== 1) {
    requireFields(record, releaseMetadataFields);
    return requireNoExtra(record, ["changed_fields"]);
  }
  switch (record.changed_fields[0]) {
    case "tracks":
      return requireChild(record);
    case "track_audio":
      if (
        record.collection_operation !== "added" &&
        record.collection_operation !== "removed"
      )
        throw new TypeError("track audio requires binding operation");
      requireIdentifier("item_id", record.item_id);
      requireIdentifier("file_id", record.file_id);
      return requireNoExtra(record, [
        "changed_fields",
        "collection_operation",
        "item_id",
        "file_id",
      ]);
    case "artwork":
      return requireFileBinding(record, "artwork");
    case "status":
      return requireReleaseLifecycle(record);
    case "share_links":
      return requireShareLink(record);
    default:
      requireFields(record, releaseMetadataFields);
      return requireNoExtra(record, ["changed_fields"]);
  }
}
function validateCampaignUpdate(record: AuditRecord): void {
  requireFields(record, campaignFields);
  if (record.actor_kind === "system") {
    if (
      record.actor_service !== "geul-backend" ||
      record.changed_fields?.length !== 1 ||
      record.changed_fields[0] !== "status"
    )
      throw new TypeError(
        "campaign system actor is limited to terminal status",
      );
    if (record.new_state !== "sent" && record.new_state !== "failed")
      throw new TypeError("campaign system status must be terminal");
  }
  const lifecycle = record.changed_fields!.some((field) =>
    ["delivery_run", "schedule", "status"].includes(field),
  );
  if (!lifecycle) return requireNoExtra(record, ["changed_fields"]);
  if (
    record.previous_state === record.new_state ||
    !record.previous_state ||
    !record.new_state
  )
    throw new TypeError("campaign lifecycle requires transition");
  const scheduled = record.changed_fields!.includes("schedule");
  if (scheduled) {
    if (!record.scheduled_at || Number.isNaN(Date.parse(record.scheduled_at)))
      throw new TypeError("campaign schedule requires scheduled_at");
    if (record.scheduled_time_zone !== undefined)
      throw new TypeError("campaign schedule does not store a time zone");
  } else if (
    record.scheduled_at !== undefined ||
    record.scheduled_time_zone !== undefined
  ) {
    throw new TypeError(
      "campaign schedule attributes require changed_fields schedule",
    );
  }
  if (record.changed_fields!.includes("delivery_run"))
    requireIdentifier("item_id", record.item_id);
  else if (record.item_id !== undefined)
    throw new TypeError(
      "campaign item_id requires changed_fields delivery_run",
    );
  requireNoExtra(record, [
    "changed_fields",
    "previous_state",
    "new_state",
    "scheduled_at",
    "item_id",
  ]);
}

/** Returns true only for Release and Campaign catalog actions. */
export function validateReleaseCampaignAuditRecord(
  record: AuditRecord,
): boolean {
  if (!actions.has(record.action as ReleaseCampaignAction)) return false;
  requireCatalogTarget(record.action, record.target_type, record.target_id);
  if (!record.action.endsWith(".updated")) {
    requireNoExtra(record, []);
    return true;
  }
  if (record.action === "release.updated") validateReleaseUpdate(record);
  else validateCampaignUpdate(record);
  return true;
}

export function buildReleaseCreatedAuditRecord(
  metadata: AuditMetadata,
  id: string,
): AuditRecord {
  return buildReleaseCampaignAuditRecord(metadata, "release.created", id);
}
export function buildReleaseDeletedAuditRecord(
  metadata: AuditMetadata,
  id: string,
): AuditRecord {
  return buildReleaseCampaignAuditRecord(metadata, "release.deleted", id);
}
export function buildReleaseMetadataAuditRecord(
  metadata: AuditMetadata,
  id: string,
  fields: readonly ReleaseMetadataField[],
): AuditRecord {
  return buildReleaseCampaignAuditRecord(metadata, "release.updated", id, {
    changed_fields: canonicalAuditValues(fields),
  });
}
export function buildReleaseTrackAuditRecord(
  metadata: AuditMetadata,
  id: string,
  trackId: string,
  itemOperation: AuditItemOperation,
): AuditRecord {
  return buildReleaseCampaignAuditRecord(metadata, "release.updated", id, {
    changed_fields: ["tracks"],
    item_id: trackId,
    item_operation: itemOperation,
  });
}
export function buildReleaseTrackOrderAuditRecord(
  metadata: AuditMetadata,
  id: string,
  trackIds: readonly string[],
): AuditRecord {
  return buildReleaseCampaignAuditRecord(metadata, "release.updated", id, {
    changed_fields: ["tracks"],
    item_ids: [...trackIds],
  });
}
export function buildReleaseTrackAudioAuditRecord(
  metadata: AuditMetadata,
  id: string,
  trackId: string,
  fileId: string,
  collectionOperation: AuditCollectionOperation,
): AuditRecord {
  return buildReleaseCampaignAuditRecord(metadata, "release.updated", id, {
    changed_fields: ["track_audio"],
    item_id: trackId,
    file_id: fileId,
    collection_operation: collectionOperation,
  });
}
export function buildReleaseArtworkAuditRecord(
  metadata: AuditMetadata,
  id: string,
  fileId: string,
  collectionOperation: AuditCollectionOperation,
): AuditRecord {
  return buildReleaseCampaignAuditRecord(metadata, "release.updated", id, {
    changed_fields: ["artwork"],
    file_id: fileId,
    collection_operation: collectionOperation,
  });
}
export function buildReleaseLifecycleAuditRecord(
  metadata: AuditMetadata,
  id: string,
  previousState: ReleaseLifecycleState,
  newState: ReleaseLifecycleState,
): AuditRecord {
  return buildReleaseCampaignAuditRecord(metadata, "release.updated", id, {
    changed_fields: ["status"],
    previous_state: previousState,
    new_state: newState,
  });
}
export function buildReleaseShareLinkAuditRecord(
  metadata: AuditMetadata,
  id: string,
  shareLinkId: string,
  itemOperation: Extract<AuditItemOperation, "created" | "deleted">,
): AuditRecord {
  return buildReleaseCampaignAuditRecord(metadata, "release.updated", id, {
    changed_fields: ["share_links"],
    item_id: shareLinkId,
    item_operation: itemOperation,
  });
}
export function buildCampaignCreatedAuditRecord(
  metadata: AuditMetadata,
  id: string,
): AuditRecord {
  return buildReleaseCampaignAuditRecord(metadata, "campaign.created", id);
}
export function buildCampaignDeletedAuditRecord(
  metadata: AuditMetadata,
  id: string,
): AuditRecord {
  return buildReleaseCampaignAuditRecord(metadata, "campaign.deleted", id);
}
export function buildCampaignTargetLayoutAuditRecord(
  metadata: AuditMetadata,
  id: string,
  fields: readonly CampaignTargetLayoutField[],
): AuditRecord {
  return buildReleaseCampaignAuditRecord(metadata, "campaign.updated", id, {
    changed_fields: canonicalAuditValues(fields),
  });
}
export function buildCampaignMetadataAuditRecord(
  metadata: AuditMetadata,
  id: string,
  fields: readonly CampaignMetadataField[],
): AuditRecord {
  return buildReleaseCampaignAuditRecord(metadata, "campaign.updated", id, {
    changed_fields: canonicalAuditValues(fields),
  });
}
export function buildCampaignStatusLifecycleAuditRecord(
  metadata: AuditMetadata,
  id: string,
  previousState: AuditState,
  newState: AuditState,
): AuditRecord {
  return buildReleaseCampaignAuditRecord(metadata, "campaign.updated", id, {
    changed_fields: ["status"],
    previous_state: previousState,
    new_state: newState,
  });
}
export function buildCampaignScheduleLifecycleAuditRecord(
  metadata: AuditMetadata,
  id: string,
  previousState: AuditState,
  newState: AuditState,
  scheduledAt: string,
): AuditRecord {
  return buildReleaseCampaignAuditRecord(metadata, "campaign.updated", id, {
    changed_fields: ["schedule"],
    previous_state: previousState,
    new_state: newState,
    scheduled_at: scheduledAt,
  });
}
export function buildCampaignDeliveryRunLifecycleAuditRecord(
  metadata: AuditMetadata,
  id: string,
  previousState: AuditState,
  newState: AuditState,
  runId: string,
): AuditRecord {
  return buildReleaseCampaignAuditRecord(metadata, "campaign.updated", id, {
    changed_fields: ["delivery_run"],
    previous_state: previousState,
    new_state: newState,
    item_id: runId,
  });
}
