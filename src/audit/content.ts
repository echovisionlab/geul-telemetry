import type {
  AuditCollectionOperation,
  AuditItemOperation,
  AuditRecord,
  AuditRelationship,
  AuditState,
} from "../records.ts";
import {
  canonicalAuditValues,
  assertOnlyAuditAttributes,
  type AuditRecordAttributeName,
} from "./attributes.ts";
import { buildAuditRecord } from "./builder.ts";
import { AUDIT_CATALOG, requireCatalogTarget } from "./catalog.ts";
import {
  isRelationDownloadPolicyAuditRecord,
  validateRelationDownloadPolicyAuditRecord,
} from "./relation-download-policy.ts";
import type { AuditMetadata } from "./types.ts";

type ContentTarget = "post" | "page" | "work";
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

type PostConfigurationField =
  "comments_enabled" | "document_layout" | "map_place_id" | "slug";
type PageConfigurationField = "document_layout" | "show_title" | "slug";
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

const postConfigurationFields = new Set<PostConfigurationField>([
  "comments_enabled",
  "document_layout",
  "map_place_id",
  "slug",
]);
const pageConfigurationFields = new Set<PageConfigurationField>([
  "document_layout",
  "show_title",
  "slug",
]);
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

function buildContentAuditRecord(
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

function requireFields(
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

function requireIdentifier(name: string, value: string | undefined): void {
  if (!value || value.length > 255 || value.trim() !== value) {
    throw new TypeError(`${name} requires an identifier`);
  }
}

function requireOnly(
  record: AuditRecord,
  allowed: readonly AuditRecordAttributeName[],
): void {
  assertOnlyAuditAttributes(record, allowed);
}

function requireVersion(record: AuditRecord): void {
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

function requireContentVersionActor(record: AuditRecord): void {
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

function requireStateTransition(
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

function requireShareLink(record: AuditRecord): void {
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

function requireFeaturedImage(
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

function requireVersionRestore(record: AuditRecord): void {
  requireIdentifier("version_id", record.version_id);
  requireOnly(record, ["changed_fields", "version_id"]);
}

function postParticipantFields(
  previous: AuditRelationship,
  next: AuditRelationship,
): readonly ("authors" | "collaborators")[] {
  const valid = new Set<AuditRelationship>(["none", "author", "collaborator"]);
  if (!valid.has(previous) || !valid.has(next) || previous === next) {
    throw new TypeError(
      "post participant requires a distinct author/collaborator transition",
    );
  }
  if (previous === "author" || next === "author") {
    if (previous === "collaborator" || next === "collaborator") {
      return ["authors", "collaborators"];
    }
    return ["authors"];
  }
  return ["collaborators"];
}

function requirePostParticipant(record: AuditRecord): void {
  const previous = record.previous_relationship;
  const next = record.new_relationship;
  if (previous === undefined || next === undefined) {
    throw new TypeError("post participant requires relationships");
  }
  const fields = postParticipantFields(previous, next);
  if (
    record.changed_fields?.length !== fields.length ||
    record.changed_fields.some((field, i) => field !== fields[i])
  ) {
    throw new TypeError(
      "post participant changed_fields must match the relationship transition",
    );
  }
  requireIdentifier("subject_member_id", record.subject_member_id);
  requireOnly(record, [
    "changed_fields",
    "subject_member_id",
    "previous_relationship",
    "new_relationship",
  ]);
}

function requirePostUpdate(record: AuditRecord): void {
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
      case "comments": {
        if (record.actor_kind !== "member") {
          throw new TypeError("post comment requires a member actor");
        }
        if (
          !record.item_operation ||
          !["created", "updated", "deleted"].includes(record.item_operation)
        ) {
          throw new TypeError("post comment requires an item operation");
        }
        requireIdentifier("item_id", record.item_id);
        requireOnly(record, ["changed_fields", "item_operation", "item_id"]);
        return;
      }
    }
  }
  if (
    record.changed_fields?.includes("authors") ||
    record.changed_fields?.includes("collaborators")
  ) {
    return requirePostParticipant(record);
  }
  const fields = requireFields(
    record,
    new Set([...postConfigurationFields, "schedule", "status"]),
  );
  const lifecycle = fields.includes("schedule") || fields.includes("status");
  if (!lifecycle) return requireOnly(record, ["changed_fields"]);
  requireStateTransition(record, [
    "draft",
    "scheduled",
    "published",
    "archived",
  ]);
  if (fields.includes("schedule")) {
    if (
      !record.scheduled_at ||
      Number.isNaN(Date.parse(record.scheduled_at)) ||
      !record.scheduled_time_zone
    ) {
      throw new TypeError(
        "post schedule requires scheduled_at and scheduled_time_zone",
      );
    }
  } else if (
    record.scheduled_at !== undefined ||
    record.scheduled_time_zone !== undefined
  ) {
    throw new TypeError(
      "post schedule attributes require changed_fields schedule",
    );
  }
  requireOnly(record, [
    "changed_fields",
    "previous_state",
    "new_state",
    "scheduled_at",
    "scheduled_time_zone",
  ]);
}

function requirePageUpdate(record: AuditRecord): void {
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

function requireWorkUpdate(record: AuditRecord): void {
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

/** Returns true only for Post, Page, and Work catalog actions. */
export function validateContentAuditRecord(record: AuditRecord): boolean {
  if (
    !(["post", "page", "work"] as const).includes(
      record.target_type as ContentTarget,
    )
  ) {
    return false;
  }
  requireCatalogTarget(record.action, record.target_type, record.target_id);
  if (!record.action.endsWith(".updated")) {
    requireOnly(record, []);
    return true;
  }
  // Version persistence is called exclusively by the authenticated collab
  // boundary. All other Content mutations must retain their Member actor.
  requireContentVersionActor(record);
  switch (record.target_type) {
    case "post":
      requirePostUpdate(record);
      break;
    case "page":
      requirePageUpdate(record);
      break;
    case "work":
      requireWorkUpdate(record);
      break;
  }
  return true;
}

function buildVersionCreatedAuditRecord(
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

export function buildPostVersionCreatedAuditRecord(
  m: AuditMetadata,
  id: string,
  version: string,
  contributors: readonly string[],
): AuditRecord {
  return buildVersionCreatedAuditRecord(
    m,
    "post.updated",
    id,
    version,
    contributors,
  );
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

export function buildPostConfigurationAuditRecord(
  m: AuditMetadata,
  id: string,
  fields: readonly PostConfigurationField[],
): AuditRecord {
  return buildContentAuditRecord(m, "post.updated", id, {
    changed_fields: canonicalAuditValues(fields),
  });
}
export function buildPostLifecycleAuditRecord(
  m: AuditMetadata,
  id: string,
  fields: readonly ("schedule" | "status")[],
  previous: AuditState,
  next: AuditState,
  scheduledAt?: string,
  scheduledTimeZone?: string,
): AuditRecord {
  return buildContentAuditRecord(m, "post.updated", id, {
    changed_fields: canonicalAuditValues(fields),
    previous_state: previous,
    new_state: next,
    scheduled_at: scheduledAt,
    scheduled_time_zone: scheduledTimeZone,
  });
}
export function buildPostFeaturedImageAuditRecord(
  m: AuditMetadata,
  id: string,
  assetId: string,
  operation: AuditCollectionOperation,
): AuditRecord {
  return buildContentAuditRecord(m, "post.updated", id, {
    changed_fields: ["featured_image"],
    asset_id: assetId,
    collection_operation: operation,
  });
}
export function buildPostParticipantAuditRecord(
  m: AuditMetadata,
  id: string,
  memberId: string,
  previous: AuditRelationship,
  next: AuditRelationship,
): AuditRecord {
  return buildContentAuditRecord(m, "post.updated", id, {
    changed_fields: postParticipantFields(previous, next),
    subject_member_id: memberId,
    previous_relationship: previous,
    new_relationship: next,
  });
}
export function buildPostShareLinkAuditRecord(
  m: AuditMetadata,
  id: string,
  itemId: string,
  operation: Extract<AuditItemOperation, "created" | "deleted">,
): AuditRecord {
  return buildContentAuditRecord(m, "post.updated", id, {
    changed_fields: ["share_links"],
    item_id: itemId,
    item_operation: operation,
  });
}
export function buildPostCommentAuditRecord(
  m: AuditMetadata,
  id: string,
  itemId: string,
  operation: AuditItemOperation,
): AuditRecord {
  return buildContentAuditRecord(m, "post.updated", id, {
    changed_fields: ["comments"],
    item_id: itemId,
    item_operation: operation,
  });
}
export function buildPostVersionRestoreAuditRecord(
  m: AuditMetadata,
  id: string,
  versionId: string,
): AuditRecord {
  return buildContentAuditRecord(m, "post.updated", id, {
    changed_fields: ["version_restore"],
    version_id: versionId,
  });
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

function buildRootAuditRecord(
  m: AuditMetadata,
  action: Extract<
    ContentAction,
    `${ContentTarget}.created` | `${ContentTarget}.deleted`
  >,
  id: string,
): AuditRecord {
  return buildContentAuditRecord(m, action, id);
}
export const buildPostCreatedAuditRecord = (
  m: AuditMetadata,
  id: string,
): AuditRecord => buildRootAuditRecord(m, "post.created", id);
export const buildPostDeletedAuditRecord = (
  m: AuditMetadata,
  id: string,
): AuditRecord => buildRootAuditRecord(m, "post.deleted", id);
export const buildPageCreatedAuditRecord = (
  m: AuditMetadata,
  id: string,
): AuditRecord => buildRootAuditRecord(m, "page.created", id);
export const buildPageDeletedAuditRecord = (
  m: AuditMetadata,
  id: string,
): AuditRecord => buildRootAuditRecord(m, "page.deleted", id);
export const buildWorkCreatedAuditRecord = (
  m: AuditMetadata,
  id: string,
): AuditRecord => buildRootAuditRecord(m, "work.created", id);
export const buildWorkDeletedAuditRecord = (
  m: AuditMetadata,
  id: string,
): AuditRecord => buildRootAuditRecord(m, "work.deleted", id);
