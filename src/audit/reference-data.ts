import type { AuditRecord } from "../records.ts";
import {
  assertOnlyAuditAttributes,
  canonicalAuditValues,
} from "./attributes.ts";
import { buildAuditRecord } from "./builder.ts";
import { AUDIT_CATALOG, requireCatalogTarget } from "./catalog.ts";
import type { AuditMetadata } from "./types.ts";

type ReferenceDataTarget = "category" | "tag" | "genre" | "style" | "format";
type ReferenceDataAction =
  | "category.created"
  | "category.updated"
  | "category.deleted"
  | "tag.created"
  | "tag.updated"
  | "tag.deleted"
  | "genre.created"
  | "genre.updated"
  | "genre.deleted"
  | "style.created"
  | "style.updated"
  | "style.deleted"
  | "format.created"
  | "format.updated"
  | "format.deleted";
type ReferenceDataUpdateAction = Extract<
  ReferenceDataAction,
  `${string}.updated`
>;
type DescribedMetadataField = "description" | "name" | "slug";
type NamedMetadataField = "name" | "slug";

const describedTargets = new Set<ReferenceDataTarget>([
  "category",
  "genre",
  "style",
]);
const referenceDataActions = new Set<ReferenceDataAction>([
  "category.created",
  "category.updated",
  "category.deleted",
  "tag.created",
  "tag.updated",
  "tag.deleted",
  "genre.created",
  "genre.updated",
  "genre.deleted",
  "style.created",
  "style.updated",
  "style.deleted",
  "format.created",
  "format.updated",
  "format.deleted",
]);
function buildReferenceDataAuditRecord(
  metadata: AuditMetadata,
  action: ReferenceDataAction,
  targetId: string,
  attributes: Parameters<typeof buildAuditRecord>[2] = {},
): AuditRecord {
  return buildAuditRecord(
    metadata,
    { action, target_type: AUDIT_CATALOG[action], target_id: targetId },
    attributes,
  );
}

function buildReferenceDataMetadataUpdatedAuditRecord(
  metadata: AuditMetadata,
  action: ReferenceDataUpdateAction,
  targetId: string,
  changedFields: readonly string[],
): AuditRecord {
  return buildReferenceDataAuditRecord(metadata, action, targetId, {
    changed_fields: canonicalAuditValues(changedFields),
  });
}

function requireNoExtra(record: AuditRecord, allowed: readonly string[]): void {
  assertOnlyAuditAttributes(record, allowed);
}

/** Returns true only for the reviewed Category, Tag, Genre, Style, and Format actions. */
export function validateReferenceDataAuditRecord(record: AuditRecord): boolean {
  if (!referenceDataActions.has(record.action as ReferenceDataAction))
    return false;
  requireCatalogTarget(record.action, record.target_type, record.target_id);
  if (!record.action.endsWith(".updated")) {
    requireNoExtra(record, []);
    return true;
  }
  const target = AUDIT_CATALOG[record.action] as ReferenceDataTarget;
  const allowed = describedTargets.has(target)
    ? new Set<DescribedMetadataField>(["description", "name", "slug"])
    : new Set<NamedMetadataField>(["name", "slug"]);
  const fields = record.changed_fields;
  if (
    !fields?.length ||
    fields.some(
      (field, index) =>
        !allowed.has(field as never) ||
        (index > 0 && fields[index - 1] >= field),
    )
  )
    throw new TypeError("invalid or non-canonical changed_fields");
  requireNoExtra(record, ["changed_fields"]);
  return true;
}

export function buildCategoryCreatedAuditRecord(
  metadata: AuditMetadata,
  id: string,
): AuditRecord {
  return buildReferenceDataAuditRecord(metadata, "category.created", id);
}
export function buildCategoryMetadataUpdatedAuditRecord(
  metadata: AuditMetadata,
  id: string,
  fields: readonly DescribedMetadataField[],
): AuditRecord {
  return buildReferenceDataMetadataUpdatedAuditRecord(
    metadata,
    "category.updated",
    id,
    fields,
  );
}
export function buildCategoryDeletedAuditRecord(
  metadata: AuditMetadata,
  id: string,
): AuditRecord {
  return buildReferenceDataAuditRecord(metadata, "category.deleted", id);
}
export function buildTagCreatedAuditRecord(
  metadata: AuditMetadata,
  id: string,
): AuditRecord {
  return buildReferenceDataAuditRecord(metadata, "tag.created", id);
}
export function buildTagMetadataUpdatedAuditRecord(
  metadata: AuditMetadata,
  id: string,
  fields: readonly NamedMetadataField[],
): AuditRecord {
  return buildReferenceDataMetadataUpdatedAuditRecord(
    metadata,
    "tag.updated",
    id,
    fields,
  );
}
export function buildTagDeletedAuditRecord(
  metadata: AuditMetadata,
  id: string,
): AuditRecord {
  return buildReferenceDataAuditRecord(metadata, "tag.deleted", id);
}
export function buildGenreCreatedAuditRecord(
  metadata: AuditMetadata,
  id: string,
): AuditRecord {
  return buildReferenceDataAuditRecord(metadata, "genre.created", id);
}
export function buildGenreMetadataUpdatedAuditRecord(
  metadata: AuditMetadata,
  id: string,
  fields: readonly DescribedMetadataField[],
): AuditRecord {
  return buildReferenceDataMetadataUpdatedAuditRecord(
    metadata,
    "genre.updated",
    id,
    fields,
  );
}
export function buildGenreDeletedAuditRecord(
  metadata: AuditMetadata,
  id: string,
): AuditRecord {
  return buildReferenceDataAuditRecord(metadata, "genre.deleted", id);
}
export function buildStyleCreatedAuditRecord(
  metadata: AuditMetadata,
  id: string,
): AuditRecord {
  return buildReferenceDataAuditRecord(metadata, "style.created", id);
}
export function buildStyleMetadataUpdatedAuditRecord(
  metadata: AuditMetadata,
  id: string,
  fields: readonly DescribedMetadataField[],
): AuditRecord {
  return buildReferenceDataMetadataUpdatedAuditRecord(
    metadata,
    "style.updated",
    id,
    fields,
  );
}
export function buildStyleDeletedAuditRecord(
  metadata: AuditMetadata,
  id: string,
): AuditRecord {
  return buildReferenceDataAuditRecord(metadata, "style.deleted", id);
}
export function buildFormatCreatedAuditRecord(
  metadata: AuditMetadata,
  id: string,
): AuditRecord {
  return buildReferenceDataAuditRecord(metadata, "format.created", id);
}
export function buildFormatMetadataUpdatedAuditRecord(
  metadata: AuditMetadata,
  id: string,
  fields: readonly NamedMetadataField[],
): AuditRecord {
  return buildReferenceDataMetadataUpdatedAuditRecord(
    metadata,
    "format.updated",
    id,
    fields,
  );
}
export function buildFormatDeletedAuditRecord(
  metadata: AuditMetadata,
  id: string,
): AuditRecord {
  return buildReferenceDataAuditRecord(metadata, "format.deleted", id);
}
