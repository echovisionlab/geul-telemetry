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
  requireOnly,
  requireShareLink,
  requireStateTransition,
  requireVersion,
  requireVersionRestore,
} from "./content-shared.ts";
import type { AuditMetadata } from "./types.ts";

type PageConfigurationField = "document_layout" | "show_title" | "slug";

const pageConfigurationFields = new Set<PageConfigurationField>([
  "document_layout",
  "show_title",
  "slug",
]);

export function requirePageUpdate(record: AuditRecord): void {
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
        requireStateTransition(record, ["draft", "published"]);
        return requireOnly(record, [
          "changed_fields",
          "previous_state",
          "new_state",
        ]);
    }
  }
  requireFields(record, pageConfigurationFields);
  requireOnly(record, ["changed_fields"]);
}

export function buildPageVersionCreatedAuditRecord(
  m: AuditMetadata,
  id: string,
  version: string,
  contributors: readonly string[],
): AuditRecord {
  return buildVersionCreatedAuditRecord(
    m,
    "page.updated",
    id,
    version,
    contributors,
  );
}
export function buildPageConfigurationAuditRecord(
  m: AuditMetadata,
  id: string,
  fields: readonly PageConfigurationField[],
): AuditRecord {
  return buildContentAuditRecord(m, "page.updated", id, {
    changed_fields: canonicalAuditValues(fields),
  });
}
export function buildPageLifecycleAuditRecord(
  m: AuditMetadata,
  id: string,
  previous: Extract<AuditState, "draft" | "published">,
  next: Extract<AuditState, "draft" | "published">,
): AuditRecord {
  return buildContentAuditRecord(m, "page.updated", id, {
    changed_fields: ["status"],
    previous_state: previous,
    new_state: next,
  });
}
export function buildPageFeaturedImageAuditRecord(
  m: AuditMetadata,
  id: string,
  assetId: string,
  operation: AuditCollectionOperation,
): AuditRecord {
  return buildContentAuditRecord(m, "page.updated", id, {
    changed_fields: ["featured_image"],
    asset_id: assetId,
    collection_operation: operation,
  });
}
export function buildPageShareLinkAuditRecord(
  m: AuditMetadata,
  id: string,
  itemId: string,
  operation: Extract<AuditItemOperation, "created" | "deleted">,
): AuditRecord {
  return buildContentAuditRecord(m, "page.updated", id, {
    changed_fields: ["share_links"],
    item_id: itemId,
    item_operation: operation,
  });
}
export function buildPageVersionRestoreAuditRecord(
  m: AuditMetadata,
  id: string,
  versionId: string,
): AuditRecord {
  return buildContentAuditRecord(m, "page.updated", id, {
    changed_fields: ["version_restore"],
    version_id: versionId,
  });
}

export const buildPageCreatedAuditRecord = (
  m: AuditMetadata,
  id: string,
): AuditRecord => buildRootAuditRecord(m, "page.created", id);
export const buildPageDeletedAuditRecord = (
  m: AuditMetadata,
  id: string,
): AuditRecord => buildRootAuditRecord(m, "page.deleted", id);
