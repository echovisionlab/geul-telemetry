import type {
  AuditAssetSlot,
  AuditCollectionOperation,
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
import type { AuditMetadata } from "./types.ts";

type ReferenceEntityTarget =
  "client" | "map_place" | "audience_segment" | "menu";
type ReferenceEntityAction = Extract<
  (typeof AUDIT_ACTIONS)[number],
  | "client.created"
  | "client.updated"
  | "client.deleted"
  | "map_place.created"
  | "map_place.updated"
  | "map_place.deleted"
  | "audience_segment.created"
  | "audience_segment.updated"
  | "menu.created"
  | "menu.updated"
  | "menu.deleted"
>;
export type ClientMetadataField = "name" | "website";
export type MapPlaceMetadataField =
  "address" | "address_components" | "google_place_id" | "lat" | "lng" | "name";
export type AudienceSegmentConfigField =
  | "account_roles"
  | "created_after"
  | "created_before"
  | "description"
  | "exclude_member_ids"
  | "member_tag_ids"
  | "name"
  | "segment_type";
export type MenuSourceField = "items" | "name";
export type AudienceSegmentLifecycleState = Extract<
  AuditState,
  "active" | "archived"
>;

const targetTypes = new Set<ReferenceEntityTarget>([
  "client",
  "map_place",
  "audience_segment",
  "menu",
]);
// Module ownership follows the immutable catalog mapping, not a duplicate action-target table.
const referenceEntityActions = new Set<ReferenceEntityAction>(
  AUDIT_ACTIONS.filter((action) =>
    targetTypes.has(AUDIT_CATALOG[action] as ReferenceEntityTarget),
  ) as ReferenceEntityAction[],
);
const clientMetadataFields = new Set<ClientMetadataField>(["name", "website"]);
const mapPlaceMetadataFields = new Set<MapPlaceMetadataField>([
  "address",
  "address_components",
  "google_place_id",
  "lat",
  "lng",
  "name",
]);
const audienceSegmentConfigFields = new Set<AudienceSegmentConfigField>([
  "account_roles",
  "created_after",
  "created_before",
  "description",
  "exclude_member_ids",
  "member_tag_ids",
  "name",
  "segment_type",
]);
const menuSourceFields = new Set<MenuSourceField>(["items", "name"]);
function buildReferenceEntityAuditRecord(
  metadata: AuditMetadata,
  action: ReferenceEntityAction,
  targetId: string,
  attributes: Parameters<typeof buildAuditRecord>[2] = {},
): AuditRecord {
  return buildAuditRecord(
    metadata,
    { action, target_type: AUDIT_CATALOG[action], target_id: targetId },
    attributes,
  );
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
function requireIdentifier(name: string, value: string | undefined): void {
  if (!value || value.length > 255 || value.trim() !== value)
    throw new TypeError(`${name} requires an identifier`);
}
function requireNoExtra(record: AuditRecord, allowed: readonly string[]): void {
  assertOnlyAuditAttributes(record, allowed);
}
function requireFileBinding(record: AuditRecord, requiresSlot: boolean): void {
  if (
    record.collection_operation !== "added" &&
    record.collection_operation !== "removed"
  )
    throw new TypeError("file binding requires collection_operation");
  requireIdentifier("file_id", record.file_id);
  if (
    requiresSlot &&
    record.asset_slot !== "light" &&
    record.asset_slot !== "dark"
  )
    throw new TypeError("client logo requires asset_slot");
  requireNoExtra(
    record,
    requiresSlot
      ? ["changed_fields", "collection_operation", "asset_slot", "file_id"]
      : ["changed_fields", "collection_operation", "file_id"],
  );
}

/** Returns true only for Client, Map Place, Audience Segment, and Menu actions. */
export function validateReferenceEntityAuditRecord(
  record: AuditRecord,
): boolean {
  if (!referenceEntityActions.has(record.action as ReferenceEntityAction))
    return false;
  requireCatalogTarget(record.action, record.target_type, record.target_id);
  if (!record.action.endsWith(".updated")) {
    requireNoExtra(record, []);
    return true;
  }
  switch (record.action) {
    case "client.updated":
      if (
        record.changed_fields?.length === 1 &&
        record.changed_fields[0] === "logo"
      ) {
        requireFileBinding(record, true);
        break;
      }
      requireFields(record, clientMetadataFields);
      requireNoExtra(record, ["changed_fields"]);
      break;
    case "map_place.updated":
      if (
        record.changed_fields?.length === 1 &&
        record.changed_fields[0] === "image"
      ) {
        requireFileBinding(record, false);
        break;
      }
      requireFields(record, mapPlaceMetadataFields);
      requireNoExtra(record, ["changed_fields"]);
      break;
    case "audience_segment.updated":
      if (
        record.changed_fields?.length === 1 &&
        record.changed_fields[0] === "status"
      ) {
        if (!(
          (record.previous_state === "active" &&
            record.new_state === "archived") ||
          (record.previous_state === "archived" &&
            record.new_state === "active")
        ))
          throw new TypeError(
            "audience status requires active/archived transition",
          );
        requireNoExtra(record, [
          "changed_fields",
          "previous_state",
          "new_state",
        ]);
        break;
      }
      requireFields(record, audienceSegmentConfigFields);
      requireNoExtra(record, ["changed_fields"]);
      break;
    case "menu.updated":
      requireFields(record, menuSourceFields);
      requireNoExtra(record, ["changed_fields"]);
      break;
  }
  return true;
}

export function buildClientCreatedAuditRecord(
  metadata: AuditMetadata,
  id: string,
): AuditRecord {
  return buildReferenceEntityAuditRecord(metadata, "client.created", id);
}
export function buildClientMetadataUpdatedAuditRecord(
  metadata: AuditMetadata,
  id: string,
  fields: readonly ClientMetadataField[],
): AuditRecord {
  return buildReferenceEntityAuditRecord(metadata, "client.updated", id, {
    changed_fields: canonicalAuditValues(fields),
  });
}
export function buildClientLogoUpdatedAuditRecord(
  metadata: AuditMetadata,
  id: string,
  slot: AuditAssetSlot,
  operation: AuditCollectionOperation,
  fileId: string,
): AuditRecord {
  return buildReferenceEntityAuditRecord(metadata, "client.updated", id, {
    changed_fields: ["logo"],
    asset_slot: slot,
    collection_operation: operation,
    file_id: fileId,
  });
}
export function buildClientDeletedAuditRecord(
  metadata: AuditMetadata,
  id: string,
): AuditRecord {
  return buildReferenceEntityAuditRecord(metadata, "client.deleted", id);
}
export function buildMapPlaceCreatedAuditRecord(
  metadata: AuditMetadata,
  id: string,
): AuditRecord {
  return buildReferenceEntityAuditRecord(metadata, "map_place.created", id);
}
export function buildMapPlaceMetadataUpdatedAuditRecord(
  metadata: AuditMetadata,
  id: string,
  fields: readonly MapPlaceMetadataField[],
): AuditRecord {
  return buildReferenceEntityAuditRecord(metadata, "map_place.updated", id, {
    changed_fields: canonicalAuditValues(fields),
  });
}
export function buildMapPlaceImageUpdatedAuditRecord(
  metadata: AuditMetadata,
  id: string,
  operation: AuditCollectionOperation,
  fileId: string,
): AuditRecord {
  return buildReferenceEntityAuditRecord(metadata, "map_place.updated", id, {
    changed_fields: ["image"],
    collection_operation: operation,
    file_id: fileId,
  });
}
export function buildMapPlaceDeletedAuditRecord(
  metadata: AuditMetadata,
  id: string,
): AuditRecord {
  return buildReferenceEntityAuditRecord(metadata, "map_place.deleted", id);
}
export function buildAudienceSegmentCreatedAuditRecord(
  metadata: AuditMetadata,
  id: string,
): AuditRecord {
  return buildReferenceEntityAuditRecord(
    metadata,
    "audience_segment.created",
    id,
  );
}
export function buildAudienceSegmentConfigUpdatedAuditRecord(
  metadata: AuditMetadata,
  id: string,
  fields: readonly AudienceSegmentConfigField[],
): AuditRecord {
  return buildReferenceEntityAuditRecord(
    metadata,
    "audience_segment.updated",
    id,
    { changed_fields: canonicalAuditValues(fields) },
  );
}
export function buildAudienceSegmentLifecycleUpdatedAuditRecord(
  metadata: AuditMetadata,
  id: string,
  previousState: AudienceSegmentLifecycleState,
  newState: AudienceSegmentLifecycleState,
): AuditRecord {
  return buildReferenceEntityAuditRecord(
    metadata,
    "audience_segment.updated",
    id,
    {
      changed_fields: ["status"],
      previous_state: previousState,
      new_state: newState,
    },
  );
}
export function buildMenuCreatedAuditRecord(
  metadata: AuditMetadata,
  id: string,
): AuditRecord {
  return buildReferenceEntityAuditRecord(metadata, "menu.created", id);
}
export function buildMenuSourceUpdatedAuditRecord(
  metadata: AuditMetadata,
  id: string,
  fields: readonly MenuSourceField[],
): AuditRecord {
  return buildReferenceEntityAuditRecord(metadata, "menu.updated", id, {
    changed_fields: canonicalAuditValues(fields),
  });
}
export function buildMenuDeletedAuditRecord(
  metadata: AuditMetadata,
  id: string,
): AuditRecord {
  return buildReferenceEntityAuditRecord(metadata, "menu.deleted", id);
}
