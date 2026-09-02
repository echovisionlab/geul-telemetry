import type { AuditRecord, AuditState } from "../records.ts";
import {
  assertOnlyAuditAttributes,
  canonicalAuditValues,
} from "./attributes.ts";
import { buildAuditRecord } from "./builder.ts";
import { AUDIT_CATALOG } from "./catalog.ts";
import type { AuditMetadata } from "./types.ts";

export type RelationDownloadPolicyChangedField =
  "file_download_audience" | "file_download_audience_segment_ids";
export type FileDownloadAudience =
  "disabled" | "public" | "authenticated" | "restricted";

const fields = new Set<RelationDownloadPolicyChangedField>([
  "file_download_audience",
  "file_download_audience_segment_ids",
]);
const audiences = new Set<FileDownloadAudience>([
  "disabled",
  "public",
  "authenticated",
  "restricted",
]);

function requireIdentifier(name: string, value: string | undefined): void {
  if (!value || value.length > 255 || value.trim() !== value)
    throw new TypeError(`${name} requires an identifier`);
}

function requireIdentifiers(
  name: string,
  values: readonly string[] | undefined,
): void {
  if (
    values === undefined ||
    values.some(
      (value, index) =>
        value === "" ||
        value.length > 255 ||
        value.trim() !== value ||
        (index > 0 && values[index - 1] >= value),
    )
  )
    throw new TypeError(`${name} requires sorted, unique identifiers`);
}

export function isRelationDownloadPolicyAuditRecord(
  record: AuditRecord,
): boolean {
  return Boolean(
    record.changed_fields?.some((field) =>
      fields.has(field as RelationDownloadPolicyChangedField),
    ),
  );
}

export function validateRelationDownloadPolicyAuditRecord(
  record: AuditRecord,
): void {
  const changedFields = record.changed_fields;
  if (
    !changedFields?.length ||
    changedFields.some(
      (field, index) =>
        !fields.has(field as RelationDownloadPolicyChangedField) ||
        (index > 0 && changedFields[index - 1] >= field),
    )
  )
    throw new TypeError("invalid download policy changed_fields");
  if (record.actor_kind !== "member")
    throw new TypeError("download policy requires member actor");
  requireIdentifier("item_id", record.item_id);
  requireIdentifier("file_id", record.file_id);
  const hasAudience = changedFields.includes("file_download_audience");
  const hasSegments = changedFields.includes(
    "file_download_audience_segment_ids",
  );
  if (hasAudience) {
    if (
      !audiences.has(record.previous_state as FileDownloadAudience) ||
      !audiences.has(record.new_state as FileDownloadAudience) ||
      record.previous_state === record.new_state
    )
      throw new TypeError("download policy audience requires a transition");
  } else if (
    record.previous_state !== undefined ||
    record.new_state !== undefined
  ) {
    throw new TypeError("download policy states require audience field");
  }
  if (hasSegments) {
    requireIdentifiers("previous_item_ids", record.previous_item_ids);
    requireIdentifiers("item_ids", record.item_ids);
    if (
      JSON.stringify(record.previous_item_ids) ===
      JSON.stringify(record.item_ids)
    )
      throw new TypeError("download policy segments require a transition");
  } else if (
    record.previous_item_ids !== undefined ||
    record.item_ids !== undefined
  ) {
    throw new TypeError("download policy segment IDs require segment field");
  }
  assertOnlyAuditAttributes(record, [
    "changed_fields",
    "item_id",
    "file_id",
    "previous_state",
    "new_state",
    "previous_item_ids",
    "item_ids",
  ]);
}

type RelationDownloadPolicyAction =
  | "post.updated"
  | "page.updated"
  | "work.updated"
  | "program_event.updated"
  | "release.updated";

function buildRelationDownloadPolicyAuditRecord(
  metadata: AuditMetadata,
  action: RelationDownloadPolicyAction,
  targetId: string,
  itemId: string,
  fileId: string,
  previousAudience: FileDownloadAudience,
  newAudience: FileDownloadAudience,
  previousSegmentIds: readonly string[],
  segmentIds: readonly string[],
): AuditRecord {
  const previousSegments = canonicalAuditValues(previousSegmentIds);
  const newSegments = canonicalAuditValues(segmentIds);
  if (!audiences.has(previousAudience) || !audiences.has(newAudience))
    throw new TypeError(
      "download policy requires valid before and after audiences",
    );
  requireIdentifiers("previous_segment_ids", previousSegments);
  requireIdentifiers("segment_ids", newSegments);
  const changedFields: RelationDownloadPolicyChangedField[] = [];
  if (previousAudience !== newAudience)
    changedFields.push("file_download_audience");
  if (JSON.stringify(previousSegments) !== JSON.stringify(newSegments))
    changedFields.push("file_download_audience_segment_ids");
  return buildAuditRecord(
    metadata,
    {
      action,
      target_type: AUDIT_CATALOG[action],
      target_id: targetId,
    },
    {
      changed_fields: canonicalAuditValues(changedFields),
      item_id: itemId,
      file_id: fileId,
      previous_state:
        previousAudience === newAudience
          ? undefined
          : (previousAudience as AuditState),
      new_state:
        previousAudience === newAudience
          ? undefined
          : (newAudience as AuditState),
      previous_item_ids:
        JSON.stringify(previousSegments) === JSON.stringify(newSegments)
          ? undefined
          : previousSegments,
      item_ids:
        JSON.stringify(previousSegments) === JSON.stringify(newSegments)
          ? undefined
          : newSegments,
    },
  );
}

export function buildPostFileBlockDownloadPolicyAuditRecord(
  metadata: AuditMetadata,
  postId: string,
  blockId: string,
  fileId: string,
  previousAudience: FileDownloadAudience,
  newAudience: FileDownloadAudience,
  previousSegmentIds: readonly string[],
  segmentIds: readonly string[],
): AuditRecord {
  return buildRelationDownloadPolicyAuditRecord(
    metadata,
    "post.updated",
    postId,
    blockId,
    fileId,
    previousAudience,
    newAudience,
    previousSegmentIds,
    segmentIds,
  );
}

export function buildPageFileBlockDownloadPolicyAuditRecord(
  metadata: AuditMetadata,
  pageId: string,
  blockId: string,
  fileId: string,
  previousAudience: FileDownloadAudience,
  newAudience: FileDownloadAudience,
  previousSegmentIds: readonly string[],
  segmentIds: readonly string[],
): AuditRecord {
  return buildRelationDownloadPolicyAuditRecord(
    metadata,
    "page.updated",
    pageId,
    blockId,
    fileId,
    previousAudience,
    newAudience,
    previousSegmentIds,
    segmentIds,
  );
}

export function buildWorkFileBlockDownloadPolicyAuditRecord(
  metadata: AuditMetadata,
  workId: string,
  blockId: string,
  fileId: string,
  previousAudience: FileDownloadAudience,
  newAudience: FileDownloadAudience,
  previousSegmentIds: readonly string[],
  segmentIds: readonly string[],
): AuditRecord {
  return buildRelationDownloadPolicyAuditRecord(
    metadata,
    "work.updated",
    workId,
    blockId,
    fileId,
    previousAudience,
    newAudience,
    previousSegmentIds,
    segmentIds,
  );
}

export function buildProgramEventFileBlockDownloadPolicyAuditRecord(
  metadata: AuditMetadata,
  eventId: string,
  blockId: string,
  fileId: string,
  previousAudience: FileDownloadAudience,
  newAudience: FileDownloadAudience,
  previousSegmentIds: readonly string[],
  segmentIds: readonly string[],
): AuditRecord {
  return buildRelationDownloadPolicyAuditRecord(
    metadata,
    "program_event.updated",
    eventId,
    blockId,
    fileId,
    previousAudience,
    newAudience,
    previousSegmentIds,
    segmentIds,
  );
}

export function buildReleaseTrackDownloadPolicyAuditRecord(
  metadata: AuditMetadata,
  releaseId: string,
  trackId: string,
  fileId: string,
  previousAudience: FileDownloadAudience,
  newAudience: FileDownloadAudience,
  previousSegmentIds: readonly string[],
  segmentIds: readonly string[],
): AuditRecord {
  return buildRelationDownloadPolicyAuditRecord(
    metadata,
    "release.updated",
    releaseId,
    trackId,
    fileId,
    previousAudience,
    newAudience,
    previousSegmentIds,
    segmentIds,
  );
}
