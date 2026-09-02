import type { AuditRecord } from "../records.ts";
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

type EmailAuthoringAction = Extract<
  (typeof AUDIT_ACTIONS)[number],
  | "email_template.created"
  | "email_template.updated"
  | "email_template.deleted"
  | "email_layout.created"
  | "email_layout.updated"
  | "email_layout.deleted"
  | "email_event_mapping.updated"
>;
export type EmailTemplateMetadataField =
  "active" | "description" | "key" | "name";
export type EmailLayoutMetadataField = "key" | "name";

const actions = new Set<EmailAuthoringAction>([
  "email_template.created",
  "email_template.updated",
  "email_template.deleted",
  "email_layout.created",
  "email_layout.updated",
  "email_layout.deleted",
  "email_event_mapping.updated",
]);
const templateMetadataFields = new Set<EmailTemplateMetadataField>([
  "active",
  "description",
  "key",
  "name",
]);
const layoutMetadataFields = new Set<EmailLayoutMetadataField>(["key", "name"]);
function buildEmailAuthoringAuditRecord(
  metadata: AuditMetadata,
  action: EmailAuthoringAction,
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
function requireRelation(record: AuditRecord, includeEventName = false): void {
  const previous = record.previous_item_id ?? "";
  const next = record.item_id ?? "";
  if (previous === next) throw new TypeError("relation no-op");
  if (previous !== "") requireIdentifier("previous_item_id", previous);
  if (next !== "") requireIdentifier("item_id", next);
  requireNoExtra(record, [
    "changed_fields",
    "previous_item_id",
    "item_id",
    ...(includeEventName ? ["event_name"] : []),
  ]);
}
function isEmailEventKey(value: string | undefined): boolean {
  return value !== undefined && /^[a-z0-9_]{1,64}$/.test(value);
}
function validateTemplateUpdate(record: AuditRecord): void {
  if (
    record.changed_fields?.length === 1 &&
    record.changed_fields[0] === "layout"
  )
    return requireRelation(record);
  requireMemberActor(record);
  requireFields(record, templateMetadataFields);
  requireNoExtra(record, ["changed_fields"]);
}
function validateLayoutUpdate(record: AuditRecord): void {
  requireMemberActor(record);
  requireFields(record, layoutMetadataFields);
  requireNoExtra(record, ["changed_fields"]);
}
function validateMappingUpdate(record: AuditRecord): void {
  requireMemberActor(record);
  if (
    record.changed_fields?.length !== 1 ||
    record.changed_fields[0] !== "template" ||
    !isEmailEventKey(record.event_name)
  )
    throw new TypeError("event mapping requires event_name and template");
  requireRelation(record, true);
}
function requireMemberActor(record: AuditRecord): void {
  if (record.actor_kind !== "member")
    throw new TypeError("email authoring mutation requires member actor");
}

/** Returns true only for the seven reviewed Email Authoring audit actions. */
export function validateEmailAuthoringAuditRecord(
  record: AuditRecord,
): boolean {
  if (!actions.has(record.action as EmailAuthoringAction)) return false;
  requireCatalogTarget(record.action, record.target_type, record.target_id);
  if (!record.action.endsWith(".updated")) {
    requireMemberActor(record);
    requireNoExtra(record, []);
    return true;
  }
  if (record.action === "email_template.updated")
    validateTemplateUpdate(record);
  else if (record.action === "email_layout.updated")
    validateLayoutUpdate(record);
  else validateMappingUpdate(record);
  return true;
}
export function buildEmailTemplateCreatedAuditRecord(
  metadata: AuditMetadata,
  id: string,
): AuditRecord {
  return buildEmailAuthoringAuditRecord(metadata, "email_template.created", id);
}
export function buildEmailTemplateDeletedAuditRecord(
  metadata: AuditMetadata,
  id: string,
): AuditRecord {
  return buildEmailAuthoringAuditRecord(metadata, "email_template.deleted", id);
}
export function buildEmailTemplateMetadataAuditRecord(
  metadata: AuditMetadata,
  id: string,
  fields: readonly EmailTemplateMetadataField[],
): AuditRecord {
  return buildEmailAuthoringAuditRecord(
    metadata,
    "email_template.updated",
    id,
    { changed_fields: canonicalAuditValues(fields) },
  );
}
export function buildEmailTemplateLayoutRelationAuditRecord(
  metadata: AuditMetadata,
  id: string,
  previousLayoutId: string,
  layoutId: string,
): AuditRecord {
  return buildEmailAuthoringAuditRecord(
    metadata,
    "email_template.updated",
    id,
    {
      changed_fields: ["layout"],
      previous_item_id: previousLayoutId === "" ? undefined : previousLayoutId,
      item_id: layoutId === "" ? undefined : layoutId,
    },
  );
}
export function buildEmailLayoutCreatedAuditRecord(
  metadata: AuditMetadata,
  id: string,
): AuditRecord {
  return buildEmailAuthoringAuditRecord(metadata, "email_layout.created", id);
}
export function buildEmailLayoutDeletedAuditRecord(
  metadata: AuditMetadata,
  id: string,
): AuditRecord {
  return buildEmailAuthoringAuditRecord(metadata, "email_layout.deleted", id);
}
export function buildEmailLayoutMetadataAuditRecord(
  metadata: AuditMetadata,
  id: string,
  fields: readonly EmailLayoutMetadataField[],
): AuditRecord {
  return buildEmailAuthoringAuditRecord(metadata, "email_layout.updated", id, {
    changed_fields: canonicalAuditValues(fields),
  });
}
export function buildEmailEventMappingTemplateAuditRecord(
  metadata: AuditMetadata,
  eventName: string,
  previousTemplateId: string,
  templateId: string,
): AuditRecord {
  return buildEmailAuthoringAuditRecord(
    metadata,
    "email_event_mapping.updated",
    eventName,
    {
      changed_fields: ["template"],
      event_name: eventName,
      previous_item_id:
        previousTemplateId === "" ? undefined : previousTemplateId,
      item_id: templateId === "" ? undefined : templateId,
    },
  );
}
