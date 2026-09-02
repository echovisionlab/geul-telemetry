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

type ProgramTarget = "program_event" | "program_event_series";
type ProgramAction = Extract<
  (typeof AUDIT_ACTIONS)[number],
  | "program_event.created"
  | "program_event.updated"
  | "program_event.deleted"
  | "program_event_series.created"
  | "program_event_series.updated"
  | "program_event_series.deleted"
>;
export type ProgramEventMetadataField =
  | "all_day"
  | "artists"
  | "clients"
  | "ends_at"
  | "external_url"
  | "labels"
  | "location_mode"
  | "map_place_id"
  | "series"
  | "slug"
  | "starts_at"
  | "stream_url"
  | "ticket_url"
  | "timezone"
  | "type";
export type ProgramEventSeriesMetadataField =
  "description" | "slug" | "summary" | "title";
type ProgramEventLifecycleState = Extract<
  AuditState,
  "draft" | "published" | "archived"
>;
type ProgramEventSeriesLifecycleState = Extract<
  AuditState,
  "draft" | "published"
>;

const targetTypes = new Set<ProgramTarget>([
  "program_event",
  "program_event_series",
]);
const programActions = new Set<ProgramAction>(
  AUDIT_ACTIONS.filter((action) =>
    targetTypes.has(AUDIT_CATALOG[action] as ProgramTarget),
  ) as ProgramAction[],
);
const eventMetadataFields = new Set<ProgramEventMetadataField>([
  "all_day",
  "artists",
  "clients",
  "ends_at",
  "external_url",
  "labels",
  "location_mode",
  "map_place_id",
  "series",
  "slug",
  "starts_at",
  "stream_url",
  "ticket_url",
  "timezone",
  "type",
]);
const seriesMetadataFields = new Set<ProgramEventSeriesMetadataField>([
  "description",
  "slug",
  "summary",
  "title",
]);
function buildProgramAuditRecord(
  metadata: AuditMetadata,
  action: ProgramAction,
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
function requireEventLifecycle(record: AuditRecord): void {
  if (
    !(["draft", "published", "archived"] as const).includes(
      record.previous_state as ProgramEventLifecycleState,
    ) ||
    !(["draft", "published", "archived"] as const).includes(
      record.new_state as ProgramEventLifecycleState,
    ) ||
    record.previous_state === record.new_state
  )
    throw new TypeError(
      "program event lifecycle requires distinct valid states",
    );
  requireNoExtra(record, ["changed_fields", "previous_state", "new_state"]);
}
function requireSeriesLifecycle(record: AuditRecord): void {
  if (
    !(["draft", "published"] as const).includes(
      record.previous_state as ProgramEventSeriesLifecycleState,
    ) ||
    !(["draft", "published"] as const).includes(
      record.new_state as ProgramEventSeriesLifecycleState,
    ) ||
    record.previous_state === record.new_state
  )
    throw new TypeError(
      "program event series lifecycle requires draft/published transition",
    );
  requireNoExtra(record, ["changed_fields", "previous_state", "new_state"]);
}
function requirePoster(record: AuditRecord): void {
  if (
    record.collection_operation !== "added" &&
    record.collection_operation !== "removed"
  )
    throw new TypeError("poster requires collection operation");
  requireIdentifier("file_id", record.file_id);
  requireNoExtra(record, ["changed_fields", "collection_operation", "file_id"]);
}
function validateEventUpdate(record: AuditRecord): void {
  if (isRelationDownloadPolicyAuditRecord(record))
    return validateRelationDownloadPolicyAuditRecord(record);
  if (record.changed_fields?.length !== 1) {
    if (record.actor_kind === "system")
      throw new TypeError("program event does not allow system actor");
    return requireFields(record, eventMetadataFields);
  }
  if (record.actor_kind === "system")
    throw new TypeError("program event does not allow system actor");
  switch (record.changed_fields[0]) {
    case "poster":
      return requirePoster(record);
    case "media":
    case "credits":
      return requireChild(record);
    case "status":
      return requireEventLifecycle(record);
    default:
      return requireFields(record, eventMetadataFields);
  }
}
function validateSeriesUpdate(record: AuditRecord): void {
  if (
    record.changed_fields?.length === 1 &&
    record.changed_fields[0] === "poster"
  )
    return requirePoster(record);
  if (
    record.changed_fields?.length === 1 &&
    record.changed_fields[0] === "status"
  )
    return requireSeriesLifecycle(record);
  requireFields(record, seriesMetadataFields);
  requireNoExtra(record, ["changed_fields"]);
}

/** Returns true only for Program Event and Program Event Series catalog actions. */
export function validateProgramAuditRecord(record: AuditRecord): boolean {
  if (!programActions.has(record.action as ProgramAction)) return false;
  requireCatalogTarget(record.action, record.target_type, record.target_id);
  if (!record.action.endsWith(".updated")) {
    requireNoExtra(record, []);
    return true;
  }
  if (record.action === "program_event.updated") validateEventUpdate(record);
  else validateSeriesUpdate(record);
  return true;
}

export function buildProgramEventCreatedAuditRecord(
  metadata: AuditMetadata,
  id: string,
): AuditRecord {
  return buildProgramAuditRecord(metadata, "program_event.created", id);
}
export function buildProgramEventDeletedAuditRecord(
  metadata: AuditMetadata,
  id: string,
): AuditRecord {
  return buildProgramAuditRecord(metadata, "program_event.deleted", id);
}
export function buildProgramEventMetadataAuditRecord(
  metadata: AuditMetadata,
  id: string,
  fields: readonly ProgramEventMetadataField[],
): AuditRecord {
  return buildProgramAuditRecord(metadata, "program_event.updated", id, {
    changed_fields: canonicalAuditValues(fields),
  });
}
export function buildProgramEventPosterAuditRecord(
  metadata: AuditMetadata,
  id: string,
  collectionOperation: AuditCollectionOperation,
  fileId: string,
): AuditRecord {
  return buildProgramAuditRecord(metadata, "program_event.updated", id, {
    changed_fields: ["poster"],
    collection_operation: collectionOperation,
    file_id: fileId,
  });
}
export function buildProgramEventChildAuditRecord(
  metadata: AuditMetadata,
  id: string,
  kind: "media" | "credits",
  itemId: string,
  itemOperation: AuditItemOperation,
): AuditRecord {
  return buildProgramAuditRecord(metadata, "program_event.updated", id, {
    changed_fields: [kind],
    item_id: itemId,
    item_operation: itemOperation,
  });
}
export function buildProgramEventChildOrderAuditRecord(
  metadata: AuditMetadata,
  id: string,
  kind: "media" | "credits",
  itemIds: readonly string[],
): AuditRecord {
  return buildProgramAuditRecord(metadata, "program_event.updated", id, {
    changed_fields: [kind],
    item_ids: [...itemIds],
  });
}
export function buildProgramEventLifecycleAuditRecord(
  metadata: AuditMetadata,
  id: string,
  previousState: ProgramEventLifecycleState,
  newState: ProgramEventLifecycleState,
): AuditRecord {
  return buildProgramAuditRecord(metadata, "program_event.updated", id, {
    changed_fields: ["status"],
    previous_state: previousState,
    new_state: newState,
  });
}
export function buildProgramEventSeriesCreatedAuditRecord(
  metadata: AuditMetadata,
  id: string,
): AuditRecord {
  return buildProgramAuditRecord(metadata, "program_event_series.created", id);
}
export function buildProgramEventSeriesDeletedAuditRecord(
  metadata: AuditMetadata,
  id: string,
): AuditRecord {
  return buildProgramAuditRecord(metadata, "program_event_series.deleted", id);
}
export function buildProgramEventSeriesMetadataAuditRecord(
  metadata: AuditMetadata,
  id: string,
  fields: readonly ProgramEventSeriesMetadataField[],
): AuditRecord {
  return buildProgramAuditRecord(metadata, "program_event_series.updated", id, {
    changed_fields: canonicalAuditValues(fields),
  });
}
export function buildProgramEventSeriesPosterAuditRecord(
  metadata: AuditMetadata,
  id: string,
  collectionOperation: AuditCollectionOperation,
  fileId: string,
): AuditRecord {
  return buildProgramAuditRecord(metadata, "program_event_series.updated", id, {
    changed_fields: ["poster"],
    collection_operation: collectionOperation,
    file_id: fileId,
  });
}
export function buildProgramEventSeriesLifecycleAuditRecord(
  metadata: AuditMetadata,
  id: string,
  previousState: ProgramEventSeriesLifecycleState,
  newState: ProgramEventSeriesLifecycleState,
): AuditRecord {
  return buildProgramAuditRecord(metadata, "program_event_series.updated", id, {
    changed_fields: ["status"],
    previous_state: previousState,
    new_state: newState,
  });
}
