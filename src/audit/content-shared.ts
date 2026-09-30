import type { AuditRecord, AuditState } from "../records.ts";
import {
  canonicalAuditValues,
  assertOnlyAuditAttributes,
  type AuditRecordAttributeName,
} from "./attributes.ts";
import { buildAuditRecord } from "./builder.ts";
import { AUDIT_CATALOG } from "./catalog.ts";
import type { AuditMetadata } from "./types.ts";

export type ContentTarget = "post" | "page" | "work";
type ContentAction =
  | "post.created"
  | "post.updated"
  | "post.deleted"
  | "page.created"
  | "page.updated"
  | "page.deleted"
  | "work.created"
  | "work.updated"
  | "work.deleted";

export function buildContentAuditRecord(
  metadata: AuditMetadata,
  action: ContentAction,
  targetId: string,
  attributes: Parameters<typeof buildAuditRecord>[2] = {},
): AuditRecord {
  return buildAuditRecord(
    metadata,
    { action, target_type: AUDIT_CATALOG[action], target_id: targetId },
    attributes,
  );
}

export function requireFields(
  record: AuditRecord,
  allowed: ReadonlySet<string>,
): readonly string[] {
  const fields = record.changed_fields;
  if (
    !fields?.length ||
    fields.some(
      (field, i) => !allowed.has(field) || (i > 0 && fields[i - 1] >= field),
    )
  ) {
    throw new TypeError("invalid or non-canonical changed_fields");
  }
  return fields;
}

export function requireIdentifier(
  name: string,
  value: string | undefined,
): void {
  if (!value || value.length > 255 || value.trim() !== value) {
    throw new TypeError(`${name} requires an identifier`);
  }
}

export function requireOnly(
  record: AuditRecord,
  allowed: readonly AuditRecordAttributeName[],
): void {
  assertOnlyAuditAttributes(record, allowed);
}

export function requireVersion(record: AuditRecord): void {
  if (!record.version_id || !record.contributor_member_ids?.length) {
    throw new TypeError(
      "version requires version_id and contributor_member_ids",
    );
  }
  if (
    record.contributor_member_ids.some(
      (id, i) =>
        !/^[0-9a-f]{8}-[0-9a-f]{4}-4[0-9a-f]{3}-[89ab][0-9a-f]{3}-[0-9a-f]{12}$/.test(
          id,
        ) ||
        (i > 0 && record.contributor_member_ids![i - 1] >= id),
    )
  ) {
    throw new TypeError("version requires sorted unique UUIDv4 contributors");
  }
  requireOnly(record, [
    "changed_fields",
    "version_id",
    "contributor_member_ids",
  ]);
}

export function requireContentVersionActor(record: AuditRecord): void {
  const version =
    record.changed_fields?.length === 1 &&
    record.changed_fields[0] === "version";
  if (version) {
    if (
      record.actor_kind !== "system" ||
      record.actor_service !== "geul-collab"
    )
      throw new TypeError("content version requires geul-collab system actor");
    return;
  }
  if (record.actor_kind === "system") {
    throw new TypeError(
      "content system actor is limited to geul-collab version checkpoints",
    );
  }
}

export function requireStateTransition(
  record: AuditRecord,
  allowed: readonly AuditState[],
): void {
  if (
    !record.previous_state ||
    !record.new_state ||
    record.previous_state === record.new_state ||
    !allowed.includes(record.previous_state) ||
    !allowed.includes(record.new_state)
  ) {
    throw new TypeError("lifecycle requires an allowed state transition");
  }
}

export function requireShareLink(record: AuditRecord): void {
  if (
    record.item_operation !== "created" &&
    record.item_operation !== "deleted"
  ) {
    throw new TypeError(
      "share link requires created or deleted item operation",
    );
  }
  requireIdentifier("item_id", record.item_id);
  requireOnly(record, ["changed_fields", "item_operation", "item_id"]);
}

export function requireFeaturedImage(
  record: AuditRecord,
  attribute: "asset_id" | "file_id",
): void {
  if (
    record.collection_operation !== "added" &&
    record.collection_operation !== "removed"
  ) {
    throw new TypeError("featured image requires a binding operation");
  }
  requireIdentifier(attribute, record[attribute]);
  requireOnly(record, ["changed_fields", "collection_operation", attribute]);
}

export function requireVersionRestore(record: AuditRecord): void {
  requireIdentifier("version_id", record.version_id);
  requireOnly(record, ["changed_fields", "version_id"]);
}

export function buildVersionCreatedAuditRecord(
  metadata: AuditMetadata,
  action: Extract<
    ContentAction,
    "post.updated" | "page.updated" | "work.updated"
  >,
  targetId: string,
  versionId: string,
  contributorMemberIds: readonly string[],
): AuditRecord {
  return buildContentAuditRecord(metadata, action, targetId, {
    changed_fields: ["version"],
    version_id: versionId,
    contributor_member_ids: canonicalAuditValues(contributorMemberIds),
  });
}

export function buildRootAuditRecord(
  m: AuditMetadata,
  action: Extract<
    ContentAction,
    `${ContentTarget}.created` | `${ContentTarget}.deleted`
  >,
  id: string,
): AuditRecord {
  return buildContentAuditRecord(m, action, id);
}
