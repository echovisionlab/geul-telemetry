import type {
  AuditCollectionOperation,
  AuditItemOperation,
  AuditRecord,
  AuditRelationship,
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

type PostConfigurationField =
  "comments_enabled" | "document_layout" | "map_place_id" | "slug";

const postConfigurationFields = new Set<PostConfigurationField>([
  "comments_enabled",
  "document_layout",
  "map_place_id",
  "slug",
]);

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

export function requirePostUpdate(record: AuditRecord): void {
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

export const buildPostCreatedAuditRecord = (
  m: AuditMetadata,
  id: string,
): AuditRecord => buildRootAuditRecord(m, "post.created", id);
export const buildPostDeletedAuditRecord = (
  m: AuditMetadata,
  id: string,
): AuditRecord => buildRootAuditRecord(m, "post.deleted", id);
