import type {
  AuditCollectionOperation,
  AuditItemOperation,
  AuditRecord,
  AuditState,
} from "../records.ts";
import { canonicalAuditValues } from "./attributes.ts";
import {
  isRelationDownloadPolicyAuditRecord,
  validateRelationDownloadPolicyAuditRecord,
} from "./relation-download-policy.ts";
import {
  buildContentAuditRecord,
  buildRootAuditRecord,
  buildVersionCreatedAuditRecord,
  requireFeaturedImage,
  requireFields,
  requireIdentifier,
  requireOnly,
  requireShareLink,
  requireStateTransition,
  requireVersion,
  requireVersionRestore,
} from "./content-shared.ts";
import type { AuditMetadata } from "./types.ts";

type WorkMetadataField =
  | "clients"
  | "featured"
  | "is_present"
  | "map_place_id"
  | "metadata"
  | "month"
  | "slug"
  | "type"
  | "until_month"
  | "until_year"
  | "year";

const workMetadataFields = new Set<WorkMetadataField>([
  "clients",
  "featured",
  "is_present",
  "map_place_id",
  "metadata",
  "month",
  "slug",
  "type",
  "until_month",
  "until_year",
  "year",
]);

export function requireWorkUpdate(record: AuditRecord): void {
  if (isRelationDownloadPolicyAuditRecord(record))
    return validateRelationDownloadPolicyAuditRecord(record);
  if (record.changed_fields?.length === 1) {
    switch (record.changed_fields[0]) {
      case "version":
        return requireVersion(record);
      case "featured_image":
        return requireFeaturedImage(record, "asset_id");
      case "share_links":
        return requireShareLink(record);
      case "version_restore":
        return requireVersionRestore(record);
      case "status":
        requireStateTransition(record, ["draft", "published", "archived"]);
        return requireOnly(record, [
          "changed_fields",
          "previous_state",
          "new_state",
        ]);
      case "credits":
        if (
          !record.item_operation ||
          !["created", "updated", "deleted"].includes(record.item_operation)
        ) {
          throw new TypeError("work credit requires an item operation");
        }
        requireIdentifier("item_id", record.item_id);
        return requireOnly(record, [
          "changed_fields",
          "item_operation",
          "item_id",
        ]);
    }
  }
  requireFields(record, workMetadataFields);
  requireOnly(record, ["changed_fields"]);
}

export function buildWorkVersionCreatedAuditRecord(
  m: AuditMetadata,
  id: string,
  version: string,
  contributors: readonly string[],
): AuditRecord {
  return buildVersionCreatedAuditRecord(
    m,
    "work.updated",
    id,
    version,
    contributors,
  );
}

export function buildWorkMetadataAuditRecord(
  m: AuditMetadata,
  id: string,
  fields: readonly WorkMetadataField[],
): AuditRecord {
  return buildContentAuditRecord(m, "work.updated", id, {
    changed_fields: canonicalAuditValues(fields),
  });
}
export function buildWorkLifecycleAuditRecord(
  m: AuditMetadata,
  id: string,
  previous: Extract<AuditState, "draft" | "published" | "archived">,
  next: Extract<AuditState, "draft" | "published" | "archived">,
): AuditRecord {
  return buildContentAuditRecord(m, "work.updated", id, {
    changed_fields: ["status"],
    previous_state: previous,
    new_state: next,
  });
}
export function buildWorkFeaturedImageAuditRecord(
  m: AuditMetadata,
  id: string,
  assetId: string,
  operation: AuditCollectionOperation,
): AuditRecord {
  return buildContentAuditRecord(m, "work.updated", id, {
    changed_fields: ["featured_image"],
    asset_id: assetId,
    collection_operation: operation,
  });
}
export function buildWorkCreditAuditRecord(
  m: AuditMetadata,
  id: string,
  itemId: string,
  operation: AuditItemOperation,
): AuditRecord {
  return buildContentAuditRecord(m, "work.updated", id, {
    changed_fields: ["credits"],
    item_id: itemId,
    item_operation: operation,
  });
}
export function buildWorkShareLinkAuditRecord(
  m: AuditMetadata,
  id: string,
  itemId: string,
  operation: Extract<AuditItemOperation, "created" | "deleted">,
): AuditRecord {
  return buildContentAuditRecord(m, "work.updated", id, {
    changed_fields: ["share_links"],
    item_id: itemId,
    item_operation: operation,
  });
}
export function buildWorkVersionRestoreAuditRecord(
  m: AuditMetadata,
  id: string,
  versionId: string,
): AuditRecord {
  return buildContentAuditRecord(m, "work.updated", id, {
    changed_fields: ["version_restore"],
    version_id: versionId,
  });
}

export const buildWorkCreatedAuditRecord = (
  m: AuditMetadata,
  id: string,
): AuditRecord => buildRootAuditRecord(m, "work.created", id);
export const buildWorkDeletedAuditRecord = (
  m: AuditMetadata,
  id: string,
): AuditRecord => buildRootAuditRecord(m, "work.deleted", id);
