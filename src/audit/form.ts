import type {
  AuditCollectionOperation,
  AuditItemOperation,
  AuditItemScope,
  AuditRecord,
  AuditState,
} from "../records.ts";
import {
  assertOnlyAuditAttributes,
  canonicalAuditValues,
} from "./attributes.ts";
import { buildAuditRecord } from "./builder.ts";
import { AUDIT_CATALOG, requireCatalogTarget } from "./catalog.ts";
import type { AuditMetadata } from "./types.ts";

type FormSettingsField =
  | "access_period"
  | "auth_required"
  | "direct_public"
  | "duplicate_policy"
  | "limit"
  | "password"
  | "required_role"
  | "slug";
type FormLifecycleState = Extract<AuditState, "draft" | "published">;

const formSettingsFields = new Set<FormSettingsField>([
  "access_period",
  "auth_required",
  "direct_public",
  "duplicate_policy",
  "limit",
  "password",
  "required_role",
  "slug",
]);
const formActions = new Set([
  "form.created",
  "form.updated",
  "form.deleted",
  "form_submission.created",
  "form_submission.deleted",
] as const);
function buildFormAuditRecord(
  metadata: AuditMetadata,
  action: typeof formActions extends Set<infer T> ? T : never,
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
function validateFormUpdate(record: AuditRecord): void {
  if (record.changed_fields?.length !== 1) {
    requireFields(record, formSettingsFields);
    return requireNoExtra(record, ["changed_fields"]);
  }
  switch (record.changed_fields[0]) {
    case "status":
      if (
        !(["draft", "published"] as const).includes(
          record.previous_state as FormLifecycleState,
        ) ||
        !(["draft", "published"] as const).includes(
          record.new_state as FormLifecycleState,
        ) ||
        record.previous_state === record.new_state
      )
        throw new TypeError(
          "form status requires a distinct draft/published transition",
        );
      return requireNoExtra(record, [
        "changed_fields",
        "previous_state",
        "new_state",
      ]);
    case "featured_image":
      if (
        (record.collection_operation !== "added" &&
          record.collection_operation !== "removed") ||
        !record.file_id
      )
        throw new TypeError(
          "featured_image requires collection operation and file_id",
        );
      requireIdentifier("file_id", record.file_id);
      return requireNoExtra(record, [
        "changed_fields",
        "collection_operation",
        "file_id",
      ]);
    case "share_links":
      if (
        (record.item_operation !== "created" &&
          record.item_operation !== "deleted") ||
        (record.item_scope !== "form" && record.item_scope !== "dashboard")
      )
        throw new TypeError(
          "form share_links requires form/dashboard scope and created/deleted operation",
        );
      requireIdentifier("item_id", record.item_id);
      return requireNoExtra(record, [
        "changed_fields",
        "item_operation",
        "item_scope",
        "item_id",
      ]);
    default:
      requireFields(record, formSettingsFields);
      return requireNoExtra(record, ["changed_fields"]);
  }
}

/** Returns true only for the five reviewed Form and Submission actions. */
export function validateFormAuditRecord(record: AuditRecord): boolean {
  if (!formActions.has(record.action as never)) return false;
  requireCatalogTarget(record.action, record.target_type, record.target_id);
  if (record.action === "form_submission.created") {
    requireIdentifier("parent_id", record.parent_id);
    requireNoExtra(record, ["parent_id"]);
  } else if (record.action === "form.updated") validateFormUpdate(record);
  else requireNoExtra(record, []);
  return true;
}

export function buildFormCreatedAuditRecord(
  metadata: AuditMetadata,
  formId: string,
): AuditRecord {
  return buildFormAuditRecord(metadata, "form.created", formId);
}
export function buildFormDeletedAuditRecord(
  metadata: AuditMetadata,
  formId: string,
): AuditRecord {
  return buildFormAuditRecord(metadata, "form.deleted", formId);
}
export function buildFormSettingsAuditRecord(
  metadata: AuditMetadata,
  formId: string,
  changedFields: readonly FormSettingsField[],
): AuditRecord {
  return buildFormAuditRecord(metadata, "form.updated", formId, {
    changed_fields: canonicalAuditValues(changedFields),
  });
}
export function buildFormLifecycleAuditRecord(
  metadata: AuditMetadata,
  formId: string,
  previousState: FormLifecycleState,
  newState: FormLifecycleState,
): AuditRecord {
  return buildFormAuditRecord(metadata, "form.updated", formId, {
    changed_fields: ["status"],
    previous_state: previousState,
    new_state: newState,
  });
}
export function buildFormFeaturedImageAuditRecord(
  metadata: AuditMetadata,
  formId: string,
  fileId: string,
  collectionOperation: AuditCollectionOperation,
): AuditRecord {
  return buildFormAuditRecord(metadata, "form.updated", formId, {
    changed_fields: ["featured_image"],
    file_id: fileId,
    collection_operation: collectionOperation,
  });
}
export function buildFormShareLinkAuditRecord(
  metadata: AuditMetadata,
  formId: string,
  itemId: string,
  itemScope: AuditItemScope,
  itemOperation: Extract<AuditItemOperation, "created" | "deleted">,
): AuditRecord {
  return buildFormAuditRecord(metadata, "form.updated", formId, {
    changed_fields: ["share_links"],
    item_id: itemId,
    item_scope: itemScope,
    item_operation: itemOperation,
  });
}
export function buildFormSubmissionCreatedAuditRecord(
  metadata: AuditMetadata,
  submissionId: string,
  formId: string,
): AuditRecord {
  return buildFormAuditRecord(
    metadata,
    "form_submission.created",
    submissionId,
    { parent_id: formId },
  );
}
export function buildFormSubmissionDeletedAuditRecord(
  metadata: AuditMetadata,
  submissionId: string,
): AuditRecord {
  return buildFormAuditRecord(
    metadata,
    "form_submission.deleted",
    submissionId,
  );
}
