import type { AuditRecord, AuditState } from "../records.ts";
import { canonicalAuditValues } from "./attributes.ts";
import { buildAuditRecord } from "./builder.ts";
import type { AuditMetadata } from "./types.ts";

function buildMemberAuditRecord(
  metadata: AuditMetadata,
  memberId: string,
  attributes: Parameters<typeof buildAuditRecord>[2],
): AuditRecord {
  return buildAuditRecord(
    metadata,
    { action: "member.updated", target_type: "member", target_id: memberId },
    attributes,
  );
}

export function buildMemberOnboardingCompletedAuditRecord(
  metadata: AuditMetadata,
  memberId: string,
  nickname: string,
): AuditRecord {
  return buildMemberAuditRecord(metadata, memberId, {
    changed_fields: ["nickname", "onboarded"],
    nickname,
  });
}

export function buildMemberRoleUpdatedAuditRecord(
  metadata: AuditMetadata,
  memberId: string,
  previousRole: string,
  newRole: string,
): AuditRecord {
  return buildMemberAuditRecord(metadata, memberId, {
    changed_fields: ["role"],
    previous_role: previousRole,
    new_role: newRole,
  });
}

export function buildMemberProfileUpdatedAuditRecord(
  metadata: AuditMetadata,
  memberId: string,
  fields: readonly string[],
  nickname: string,
): AuditRecord {
  return buildMemberAuditRecord(metadata, memberId, {
    changed_fields: canonicalAuditValues(fields),
    ...(nickname === "" ? {} : { nickname }),
  });
}

export function buildMemberAvatarUpdatedAuditRecord(
  metadata: AuditMetadata,
  memberId: string,
  operation: "added" | "removed",
  assetId: string,
): AuditRecord {
  return buildMemberAuditRecord(metadata, memberId, {
    changed_fields: ["avatar"],
    collection_operation: operation,
    asset_id: assetId,
  });
}

export function buildMemberPreferencesUpdatedAuditRecord(
  metadata: AuditMetadata,
  memberId: string,
  fields: readonly string[],
  preferredLocale: string,
  consentId: string,
): AuditRecord {
  return buildMemberAuditRecord(metadata, memberId, {
    changed_fields: canonicalAuditValues(fields),
    ...(preferredLocale === "" ? {} : { preferred_locale: preferredLocale }),
    ...(consentId === "" ? {} : { consent_id: consentId }),
  });
}

export function buildMemberTagsUpdatedAuditRecord(
  metadata: AuditMetadata,
  memberId: string,
  tagIds: readonly string[],
): AuditRecord {
  return buildMemberAuditRecord(metadata, memberId, {
    changed_fields: ["tags"],
    tag_ids: canonicalAuditValues(tagIds),
  });
}

function buildMemberTagAuditRecord(
  metadata: AuditMetadata,
  action: "member_tag.created" | "member_tag.deleted",
  tagId: string,
  tagName: string,
): AuditRecord {
  return buildAuditRecord(
    metadata,
    {
      action,
      target_type: "member_tag",
      target_id: tagId,
    },
    {
      tag_name: tagName,
    },
  );
}

export function buildMemberTagCreatedAuditRecord(
  metadata: AuditMetadata,
  tagId: string,
  tagName: string,
): AuditRecord {
  return buildMemberTagAuditRecord(
    metadata,
    "member_tag.created",
    tagId,
    tagName,
  );
}

export function buildMemberTagDeletedAuditRecord(
  metadata: AuditMetadata,
  tagId: string,
  tagName: string,
): AuditRecord {
  return buildMemberTagAuditRecord(
    metadata,
    "member_tag.deleted",
    tagId,
    tagName,
  );
}

function buildMemberStatusUpdatedAuditRecord(
  metadata: AuditMetadata,
  memberId: string,
  previousState: AuditState,
  newState: AuditState,
): AuditRecord {
  return buildMemberAuditRecord(metadata, memberId, {
    changed_fields: ["status"],
    previous_state: previousState,
    new_state: newState,
  });
}

export function buildMemberBannedAuditRecord(
  metadata: AuditMetadata,
  memberId: string,
): AuditRecord {
  return buildMemberStatusUpdatedAuditRecord(
    metadata,
    memberId,
    "active",
    "banned",
  );
}
export function buildMemberUnbannedAuditRecord(
  metadata: AuditMetadata,
  memberId: string,
): AuditRecord {
  return buildMemberStatusUpdatedAuditRecord(
    metadata,
    memberId,
    "banned",
    "active",
  );
}
