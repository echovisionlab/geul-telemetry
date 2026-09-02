import type {
  AuditCollectionOperation,
  AuditRecord,
  AuditRelationship,
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
import type { AuditMetadata } from "./types.ts";

type SeriesTypeAction = Extract<
  (typeof AUDIT_ACTIONS)[number],
  | "post_series.created"
  | "post_series.updated"
  | "post_series.deleted"
  | "program_event_type.created"
  | "program_event_type.updated"
  | "program_event_type.deleted"
>;
type PostSeriesLifecycleState = Extract<AuditState, "draft" | "published">;
export type PostSeriesSourceMetadataField = "slug" | "source_copy";
export type ProgramEventTypeConfigField =
  "requires_place" | "requires_stream_url" | "slug" | "sort_order" | "status";

const targetTypes = new Set(["post_series", "program_event_type"] as const);
const seriesTypeActions = new Set<SeriesTypeAction>(
  AUDIT_ACTIONS.filter((action) =>
    targetTypes.has(
      AUDIT_CATALOG[action] as "post_series" | "program_event_type",
    ),
  ) as SeriesTypeAction[],
);
const postSeriesSourceMetadataFields = new Set<PostSeriesSourceMetadataField>([
  "slug",
  "source_copy",
]);
const programEventTypeConfigFields = new Set<ProgramEventTypeConfigField>([
  "requires_place",
  "requires_stream_url",
  "slug",
  "sort_order",
  "status",
]);
function buildSeriesTypeAuditRecord(
  metadata: AuditMetadata,
  action: SeriesTypeAction,
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
      (field, index) =>
        !allowed.has(field) || (index > 0 && fields[index - 1] >= field),
    )
  )
    throw new TypeError("invalid or non-canonical changed_fields");
}
function requirePostSeriesLifecycle(record: AuditRecord): void {
  if (
    record.changed_fields?.length !== 1 ||
    record.changed_fields[0] !== "status" ||
    !(
      (record.previous_state === "draft" && record.new_state === "published") ||
      (record.previous_state === "published" && record.new_state === "draft")
    )
  )
    throw new TypeError(
      "post series lifecycle requires draft/published transition",
    );
  requireNoExtra(record, ["changed_fields", "previous_state", "new_state"]);
}
function requirePostSeriesManager(record: AuditRecord): void {
  if (
    record.changed_fields?.length !== 1 ||
    record.changed_fields[0] !== "managers" ||
    !["none", "manager"].includes(record.previous_relationship ?? "") ||
    !["none", "manager"].includes(record.new_relationship ?? "") ||
    record.previous_relationship === record.new_relationship
  )
    throw new TypeError(
      "post series manager requires an allowed relationship transition",
    );
  requireIdentifier("subject_member_id", record.subject_member_id);
  requireNoExtra(record, [
    "changed_fields",
    "subject_member_id",
    "previous_relationship",
    "new_relationship",
  ]);
}
function requirePostSeriesMembership(record: AuditRecord): void {
  if (
    record.changed_fields?.length !== 1 ||
    record.changed_fields[0] !== "posts" ||
    record.previous_series_id === record.new_series_id
  )
    throw new TypeError(
      "post series membership requires subject and distinct series",
    );
  requireIdentifier("subject_post_id", record.subject_post_id);
  requireNoExtra(record, [
    "changed_fields",
    "subject_post_id",
    "previous_series_id",
    "new_series_id",
  ]);
}
function requirePostSeriesOrder(record: AuditRecord): void {
  if (
    record.changed_fields?.length !== 1 ||
    record.changed_fields[0] !== "post_order" ||
    record.post_ids === undefined
  )
    throw new TypeError("post series order requires post_ids");
  for (const postId of record.post_ids) requireIdentifier("post_ids", postId);
  requireNoExtra(record, ["changed_fields", "post_ids"]);
}
function requirePostSeriesFeaturedImage(record: AuditRecord): void {
  if (
    record.changed_fields?.length !== 1 ||
    record.changed_fields[0] !== "featured_image" ||
    (record.collection_operation !== "added" &&
      record.collection_operation !== "removed")
  )
    throw new TypeError(
      "post series featured image requires collection operation",
    );
  requireIdentifier("file_id", record.file_id);
  requireNoExtra(record, ["changed_fields", "collection_operation", "file_id"]);
}
function validatePostSeriesUpdate(record: AuditRecord): void {
  switch (record.changed_fields?.[0]) {
    case "status":
      return requirePostSeriesLifecycle(record);
    case "managers":
      return requirePostSeriesManager(record);
    case "posts":
      return requirePostSeriesMembership(record);
    case "post_order":
      return requirePostSeriesOrder(record);
    case "featured_image":
      return requirePostSeriesFeaturedImage(record);
    default:
      requireFields(record, postSeriesSourceMetadataFields);
      return requireNoExtra(record, ["changed_fields"]);
  }
}

/** Returns true only for Post Series and Program Event Type catalog actions. */
export function validateSeriesTypeAuditRecord(record: AuditRecord): boolean {
  if (!seriesTypeActions.has(record.action as SeriesTypeAction)) return false;
  requireCatalogTarget(record.action, record.target_type, record.target_id);
  if (!record.action.endsWith(".updated")) {
    requireNoExtra(record, []);
    return true;
  }
  if (record.action === "post_series.updated") validatePostSeriesUpdate(record);
  else {
    requireFields(record, programEventTypeConfigFields);
    requireNoExtra(record, ["changed_fields"]);
  }
  return true;
}

export function buildPostSeriesCreatedAuditRecord(
  metadata: AuditMetadata,
  id: string,
): AuditRecord {
  return buildSeriesTypeAuditRecord(metadata, "post_series.created", id);
}
export function buildPostSeriesDeletedAuditRecord(
  metadata: AuditMetadata,
  id: string,
): AuditRecord {
  return buildSeriesTypeAuditRecord(metadata, "post_series.deleted", id);
}
export function buildPostSeriesSourceMetadataAuditRecord(
  metadata: AuditMetadata,
  id: string,
  fields: readonly PostSeriesSourceMetadataField[],
): AuditRecord {
  return buildSeriesTypeAuditRecord(metadata, "post_series.updated", id, {
    changed_fields: canonicalAuditValues(fields),
  });
}
export function buildPostSeriesLifecycleAuditRecord(
  metadata: AuditMetadata,
  id: string,
  previousState: PostSeriesLifecycleState,
  newState: PostSeriesLifecycleState,
): AuditRecord {
  return buildSeriesTypeAuditRecord(metadata, "post_series.updated", id, {
    changed_fields: ["status"],
    previous_state: previousState,
    new_state: newState,
  });
}
export function buildPostSeriesManagerAuditRecord(
  metadata: AuditMetadata,
  id: string,
  subjectMemberId: string,
  previousRelationship: Extract<AuditRelationship, "none" | "manager">,
  newRelationship: Extract<AuditRelationship, "none" | "manager">,
): AuditRecord {
  return buildSeriesTypeAuditRecord(metadata, "post_series.updated", id, {
    changed_fields: ["managers"],
    subject_member_id: subjectMemberId,
    previous_relationship: previousRelationship,
    new_relationship: newRelationship,
  });
}
export function buildPostSeriesMembershipAuditRecord(
  metadata: AuditMetadata,
  id: string,
  subjectPostId: string,
  previousSeriesId: string,
  newSeriesId: string,
): AuditRecord {
  return buildSeriesTypeAuditRecord(metadata, "post_series.updated", id, {
    changed_fields: ["posts"],
    subject_post_id: subjectPostId,
    previous_series_id: previousSeriesId === "" ? undefined : previousSeriesId,
    new_series_id: newSeriesId === "" ? undefined : newSeriesId,
  });
}
export function buildPostSeriesOrderAuditRecord(
  metadata: AuditMetadata,
  id: string,
  postIds: readonly string[],
): AuditRecord {
  return buildSeriesTypeAuditRecord(metadata, "post_series.updated", id, {
    changed_fields: ["post_order"],
    post_ids: [...postIds],
  });
}
export function buildPostSeriesFeaturedImageAuditRecord(
  metadata: AuditMetadata,
  id: string,
  collectionOperation: AuditCollectionOperation,
  fileId: string,
): AuditRecord {
  return buildSeriesTypeAuditRecord(metadata, "post_series.updated", id, {
    changed_fields: ["featured_image"],
    collection_operation: collectionOperation,
    file_id: fileId,
  });
}
export function buildProgramEventTypeCreatedAuditRecord(
  metadata: AuditMetadata,
  id: string,
): AuditRecord {
  return buildSeriesTypeAuditRecord(metadata, "program_event_type.created", id);
}
export function buildProgramEventTypeDeletedAuditRecord(
  metadata: AuditMetadata,
  id: string,
): AuditRecord {
  return buildSeriesTypeAuditRecord(metadata, "program_event_type.deleted", id);
}
export function buildProgramEventTypeConfigUpdatedAuditRecord(
  metadata: AuditMetadata,
  id: string,
  fields: readonly ProgramEventTypeConfigField[],
): AuditRecord {
  return buildSeriesTypeAuditRecord(
    metadata,
    "program_event_type.updated",
    id,
    { changed_fields: canonicalAuditValues(fields) },
  );
}
