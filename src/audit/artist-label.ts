import type {
  AuditAssetSlot,
  AuditCollectionOperation,
  AuditRecord,
  AuditRelationship,
  AuditState,
} from "../records.ts";
import { assertOnlyAuditAttributes } from "./attributes.ts";
import { buildAuditRecord } from "./builder.ts";
import {
  AUDIT_ACTIONS,
  AUDIT_CATALOG,
  requireCatalogTarget,
} from "./catalog.ts";
import type { AuditMetadata } from "./types.ts";

type ArtistLabelTarget = "artist" | "label";
type ArtistLabelAction = Extract<
  (typeof AUDIT_ACTIONS)[number],
  | "artist.created"
  | "artist.updated"
  | "artist.deleted"
  | "label.created"
  | "label.updated"
  | "label.deleted"
>;
type ArtistLabelLifecycleState = Extract<AuditState, "draft" | "published">;
type ShareLinkOperation = "created" | "deleted";

const targetTypes = new Set<ArtistLabelTarget>(["artist", "label"]);
const artistLabelActions = new Set<ArtistLabelAction>(
  AUDIT_ACTIONS.filter((action) =>
    targetTypes.has(AUDIT_CATALOG[action] as ArtistLabelTarget),
  ) as ArtistLabelAction[],
);
function buildArtistLabelAuditRecord(
  metadata: AuditMetadata,
  action: ArtistLabelAction,
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
function requireMemberActor(record: AuditRecord): void {
  if (record.actor_kind !== "member")
    throw new TypeError(
      "artist and label mutation variants require member actor",
    );
}
function requireLifecycle(record: AuditRecord): void {
  if (
    record.changed_fields?.length !== 1 ||
    record.changed_fields[0] !== "status" ||
    !(
      (record.previous_state === "draft" && record.new_state === "published") ||
      (record.previous_state === "published" && record.new_state === "draft")
    )
  )
    throw new TypeError("lifecycle requires draft/published transition");
  requireNoExtra(record, ["changed_fields", "previous_state", "new_state"]);
}
function requireParticipant(record: AuditRecord): void {
  if (
    record.changed_fields?.length !== 1 ||
    record.changed_fields[0] !== "participants" ||
    !["none", "owner", "manager"].includes(
      record.previous_relationship ?? "",
    ) ||
    !["none", "owner", "manager"].includes(record.new_relationship ?? "") ||
    record.previous_relationship === record.new_relationship
  )
    throw new TypeError(
      "participant requires an allowed relationship transition",
    );
  requireIdentifier("subject_member_id", record.subject_member_id);
  requireNoExtra(record, [
    "changed_fields",
    "subject_member_id",
    "previous_relationship",
    "new_relationship",
  ]);
}
function requireShareLink(record: AuditRecord): void {
  if (
    record.changed_fields?.length !== 1 ||
    record.changed_fields[0] !== "share_links" ||
    (record.item_operation !== "created" && record.item_operation !== "deleted")
  )
    throw new TypeError(
      "share link requires created or deleted item operation",
    );
  requireIdentifier("item_id", record.item_id);
  requireNoExtra(record, ["changed_fields", "item_operation", "item_id"]);
}
function requireGallery(record: AuditRecord): void {
  if (
    record.changed_fields?.length !== 1 ||
    record.changed_fields[0] !== "gallery" ||
    record.file_ids === undefined
  )
    throw new TypeError("artist gallery requires file_ids");
  const seen = new Set<string>();
  for (const fileId of record.file_ids) {
    requireIdentifier("file_ids", fileId);
    if (seen.has(fileId))
      throw new TypeError("artist gallery requires unique file_ids");
    seen.add(fileId);
  }
  requireNoExtra(record, ["changed_fields", "file_ids"]);
}
function requireLogo(record: AuditRecord): void {
  if (
    record.changed_fields?.length !== 1 ||
    record.changed_fields[0] !== "logo" ||
    (record.asset_slot !== "light" && record.asset_slot !== "dark") ||
    (record.collection_operation !== "added" &&
      record.collection_operation !== "removed")
  )
    throw new TypeError("label logo requires slot and operation");
  requireIdentifier("asset_id", record.asset_id);
  requireNoExtra(record, [
    "changed_fields",
    "asset_slot",
    "collection_operation",
    "asset_id",
  ]);
}
function validateArtistUpdate(record: AuditRecord): void {
  switch (record.changed_fields?.[0]) {
    case "status":
      requireMemberActor(record);
      return requireLifecycle(record);
    case "gallery":
      requireMemberActor(record);
      return requireGallery(record);
    case "participants":
      requireMemberActor(record);
      return requireParticipant(record);
    case "share_links":
      requireMemberActor(record);
      return requireShareLink(record);
    default:
      throw new TypeError("artist update rejects variant");
  }
}
function validateLabelUpdate(record: AuditRecord): void {
  switch (record.changed_fields?.[0]) {
    case "status":
      requireMemberActor(record);
      return requireLifecycle(record);
    case "participants":
      requireMemberActor(record);
      return requireParticipant(record);
    case "logo":
      requireMemberActor(record);
      return requireLogo(record);
    case "share_links":
      requireMemberActor(record);
      return requireShareLink(record);
    default:
      throw new TypeError("label update rejects variant");
  }
}

/** Returns true only for reviewed Artist and Label catalog actions. */
export function validateArtistLabelAuditRecord(record: AuditRecord): boolean {
  if (!artistLabelActions.has(record.action as ArtistLabelAction)) return false;
  requireCatalogTarget(record.action, record.target_type, record.target_id);
  if (!record.action.endsWith(".updated")) {
    requireNoExtra(record, []);
    return true;
  }
  if (record.action === "artist.updated") validateArtistUpdate(record);
  else validateLabelUpdate(record);
  return true;
}

export function buildArtistCreatedAuditRecord(
  metadata: AuditMetadata,
  id: string,
): AuditRecord {
  return buildArtistLabelAuditRecord(metadata, "artist.created", id);
}
export function buildArtistDeletedAuditRecord(
  metadata: AuditMetadata,
  id: string,
): AuditRecord {
  return buildArtistLabelAuditRecord(metadata, "artist.deleted", id);
}
export function buildArtistLifecycleAuditRecord(
  metadata: AuditMetadata,
  id: string,
  previousState: ArtistLabelLifecycleState,
  newState: ArtistLabelLifecycleState,
): AuditRecord {
  return buildArtistLabelAuditRecord(metadata, "artist.updated", id, {
    changed_fields: ["status"],
    previous_state: previousState,
    new_state: newState,
  });
}
export function buildArtistGalleryAuditRecord(
  metadata: AuditMetadata,
  id: string,
  fileIds: readonly string[],
): AuditRecord {
  return buildArtistLabelAuditRecord(metadata, "artist.updated", id, {
    changed_fields: ["gallery"],
    file_ids: [...fileIds],
  });
}
export function buildArtistParticipantAuditRecord(
  metadata: AuditMetadata,
  id: string,
  subjectMemberId: string,
  previousRelationship: AuditRelationship,
  newRelationship: AuditRelationship,
): AuditRecord {
  return buildArtistLabelAuditRecord(metadata, "artist.updated", id, {
    changed_fields: ["participants"],
    subject_member_id: subjectMemberId,
    previous_relationship: previousRelationship,
    new_relationship: newRelationship,
  });
}
export function buildArtistShareLinkAuditRecord(
  metadata: AuditMetadata,
  id: string,
  itemId: string,
  itemOperation: ShareLinkOperation,
): AuditRecord {
  return buildArtistLabelAuditRecord(metadata, "artist.updated", id, {
    changed_fields: ["share_links"],
    item_id: itemId,
    item_operation: itemOperation,
  });
}
export function buildLabelCreatedAuditRecord(
  metadata: AuditMetadata,
  id: string,
): AuditRecord {
  return buildArtistLabelAuditRecord(metadata, "label.created", id);
}
export function buildLabelDeletedAuditRecord(
  metadata: AuditMetadata,
  id: string,
): AuditRecord {
  return buildArtistLabelAuditRecord(metadata, "label.deleted", id);
}
export function buildLabelLifecycleAuditRecord(
  metadata: AuditMetadata,
  id: string,
  previousState: ArtistLabelLifecycleState,
  newState: ArtistLabelLifecycleState,
): AuditRecord {
  return buildArtistLabelAuditRecord(metadata, "label.updated", id, {
    changed_fields: ["status"],
    previous_state: previousState,
    new_state: newState,
  });
}
export function buildLabelParticipantAuditRecord(
  metadata: AuditMetadata,
  id: string,
  subjectMemberId: string,
  previousRelationship: AuditRelationship,
  newRelationship: AuditRelationship,
): AuditRecord {
  return buildArtistLabelAuditRecord(metadata, "label.updated", id, {
    changed_fields: ["participants"],
    subject_member_id: subjectMemberId,
    previous_relationship: previousRelationship,
    new_relationship: newRelationship,
  });
}
export function buildLabelLogoAuditRecord(
  metadata: AuditMetadata,
  id: string,
  assetSlot: AuditAssetSlot,
  collectionOperation: AuditCollectionOperation,
  assetId: string,
): AuditRecord {
  return buildArtistLabelAuditRecord(metadata, "label.updated", id, {
    changed_fields: ["logo"],
    asset_slot: assetSlot,
    collection_operation: collectionOperation,
    asset_id: assetId,
  });
}
export function buildLabelShareLinkAuditRecord(
  metadata: AuditMetadata,
  id: string,
  itemId: string,
  itemOperation: ShareLinkOperation,
): AuditRecord {
  return buildArtistLabelAuditRecord(metadata, "label.updated", id, {
    changed_fields: ["share_links"],
    item_id: itemId,
    item_operation: itemOperation,
  });
}
