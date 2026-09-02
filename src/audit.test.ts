import { describe, expect, it } from "vitest";
import { readFileSync } from "node:fs";
import { fileURLToPath } from "node:url";

import {
  buildAccountCanonicalEmailUpdatedAuditRecord,
  buildAccountDeletedAuditRecord,
  buildAccountDeletionCancelledAuditRecord,
  buildAccountDeletionRecoveredAuditRecord,
  buildAccountDeletionRequestedAuditRecord,
  buildAccountDeletionScheduledAuditRecord,
  buildAccountEmailLoginAddedAuditRecord,
  buildAccountEmailLoginRemovedAuditRecord,
  buildAccountPasskeyAddedAuditRecord,
  buildAccountPasskeyRemovedAuditRecord,
  buildAccountNewsletterSubscriptionUpdatedAuditRecord,
  buildAccountSessionRevokedAuditRecord,
  buildAccountSocialLoginAddedAuditRecord,
  buildAccountSocialLoginRemovedAuditRecord,
  buildMemberBannedAuditRecord,
  buildMemberAvatarUpdatedAuditRecord,
  buildMemberOnboardingCompletedAuditRecord,
  buildMemberPreferencesUpdatedAuditRecord,
  buildMemberProfileUpdatedAuditRecord,
  buildMemberRoleUpdatedAuditRecord,
  buildMemberTagCreatedAuditRecord,
  buildMemberTagDeletedAuditRecord,
  buildMemberTagsUpdatedAuditRecord,
  buildMemberUnbannedAuditRecord,
  buildPageCreatedAuditRecord,
  buildPageDeletedAuditRecord,
  buildPageVersionCreatedAuditRecord,
  buildPostCreatedAuditRecord,
  buildPostDeletedAuditRecord,
  buildPostVersionCreatedAuditRecord,
  buildSiteSettingsUpdatedAuditRecord,
  buildWorkCreatedAuditRecord,
  buildWorkDeletedAuditRecord,
  buildWorkVersionCreatedAuditRecord,
  type AuditMetadata,
} from "./audit.ts";
import {
  AUDIT_ACTIONS,
  validateAuditRecord,
  type AuditRecord,
} from "./records.ts";
import { canonicalAuditValues } from "./audit/attributes.ts";
import { AUDIT_CATALOG } from "./audit/catalog.ts";

const metadata: AuditMetadata = {
  audit_id: "00000000-0000-4000-8000-000000000001",
  occurred_at: "2026-08-09T03:04:05Z",
  request_id: "018f47a2-8a3d-4e17-9d42-6f12c89b1234",
  actor_kind: "member",
  actor_member_id: "member-1",
};
const contributors = [
  "1b6bcad2-c90d-49e9-bec7-f9a4ba6b2894",
  "7a7a8fd4-1f69-4e9a-9dc2-2378926ff351",
] as const;
const systemMetadata: AuditMetadata = {
  audit_id: metadata.audit_id,
  occurred_at: metadata.occurred_at,
  request_id: metadata.request_id,
  actor_kind: "system",
  actor_service: "geul-backend",
};
const collabMetadata: AuditMetadata = {
  audit_id: metadata.audit_id,
  occurred_at: metadata.occurred_at,
  request_id: metadata.request_id,
  actor_kind: "system",
  actor_service: "geul-collab",
};

describe("audit builders", () => {
  it("covers only the exact approved catalog", () => {
    const records = [
      buildSiteSettingsUpdatedAuditRecord(metadata, ["site_title"]),
      buildMemberOnboardingCompletedAuditRecord(
        metadata,
        "member-1",
        "Onboarded Member",
      ),
      buildMemberRoleUpdatedAuditRecord(metadata, "member-2", "user", "author"),
      buildMemberProfileUpdatedAuditRecord(
        metadata,
        "member-1",
        ["social_links"],
        "",
      ),
      buildMemberAvatarUpdatedAuditRecord(
        metadata,
        "member-1",
        "added",
        "asset-1",
      ),
      buildMemberPreferencesUpdatedAuditRecord(
        metadata,
        "member-1",
        ["preferred_locale"],
        "ko",
        "",
      ),
      buildMemberTagsUpdatedAuditRecord(metadata, "member-1", []),
      buildMemberTagCreatedAuditRecord(metadata, "tag-1", "Featured"),
      buildMemberTagDeletedAuditRecord(metadata, "tag-1", "Featured"),
      buildMemberBannedAuditRecord(metadata, "member-2"),
      buildMemberUnbannedAuditRecord(metadata, "member-2"),
      buildPostVersionCreatedAuditRecord(
        collabMetadata,
        "post-1",
        "version-1",
        contributors,
      ),
      buildPageVersionCreatedAuditRecord(
        collabMetadata,
        "page-1",
        "version-1",
        contributors,
      ),
      buildWorkVersionCreatedAuditRecord(
        collabMetadata,
        "work-1",
        "version-1",
        contributors,
      ),
      buildPostCreatedAuditRecord(metadata, "post-1"),
      buildPageCreatedAuditRecord(metadata, "page-1"),
      buildWorkCreatedAuditRecord(metadata, "work-1"),
      buildPostDeletedAuditRecord(metadata, "post-1"),
      buildPageDeletedAuditRecord(metadata, "page-1"),
      buildWorkDeletedAuditRecord(metadata, "work-1"),
      buildAccountCanonicalEmailUpdatedAuditRecord(
        metadata,
        "member-1",
        "old@example.test",
        "new@example.test",
      ),
      buildAccountEmailLoginAddedAuditRecord(
        metadata,
        "member-1",
        "added@example.test",
      ),
      buildAccountEmailLoginRemovedAuditRecord(
        metadata,
        "member-1",
        "removed@example.test",
      ),
      buildAccountSocialLoginAddedAuditRecord(
        metadata,
        "member-1",
        "google",
        "google-subject",
      ),
      buildAccountSocialLoginRemovedAuditRecord(
        metadata,
        "member-1",
        "github",
        "github-subject",
      ),
      buildAccountPasskeyAddedAuditRecord(metadata, "member-1", ["passkey-1"]),
      buildAccountPasskeyRemovedAuditRecord(metadata, "member-1", [
        "passkey-2",
      ]),
      buildAccountSessionRevokedAuditRecord(metadata, "member-1", "one", [
        "018f47a2-8a3d-4e17-9d42-6f12c89b1234",
      ]),
      buildAccountNewsletterSubscriptionUpdatedAuditRecord(
        metadata,
        "member-1",
        "subscribed",
        "unsubscribed",
      ),
      buildAccountDeletionRequestedAuditRecord(metadata, "member-1", "none"),
      buildAccountDeletionScheduledAuditRecord(
        metadata,
        "member-1",
        "confirmation_pending",
      ),
      buildAccountDeletionCancelledAuditRecord(metadata, "member-1"),
      buildAccountDeletionRecoveredAuditRecord(metadata, "member-1"),
      buildAccountDeletedAuditRecord(systemMetadata, "member-1"),
    ];
    expect(new Set(records.map((record) => record.action))).toEqual(
      new Set([
        "site_settings.updated",
        "member.updated",
        "member_tag.created",
        "member_tag.deleted",
        "post.updated",
        "page.updated",
        "work.updated",
        "post.created",
        "page.created",
        "work.created",
        "post.deleted",
        "page.deleted",
        "work.deleted",
        "account.updated",
        "account.deleted",
      ]),
    );
    expect(() => buildPostCreatedAuditRecord(metadata, "")).toThrow(
      "target_id",
    );
    expect(() =>
      buildMemberOnboardingCompletedAuditRecord(metadata, "member-1", " "),
    ).toThrow("nickname");
    expect(() =>
      buildMemberProfileUpdatedAuditRecord(
        metadata,
        "member-1",
        ["bio"],
        "name",
      ),
    ).toThrow("nickname requires changed_fields");
    expect(() =>
      buildMemberPreferencesUpdatedAuditRecord(
        metadata,
        "member-1",
        ["cookie_consent"],
        "ko",
        "consent-1",
      ),
    ).toThrow("preferred_locale requires changed_fields");
    expect(() =>
      buildAccountNewsletterSubscriptionUpdatedAuditRecord(
        metadata,
        "member-1",
        "subscribed",
        "subscribed",
      ),
    ).toThrow("newsletter_subscription");
    expect(() =>
      buildMemberOnboardingCompletedAuditRecord(
        { ...metadata, actor_member_id: "member-2" },
        "member-1",
        "Name",
      ),
    ).toThrow("actor and target");
    expect(() =>
      buildMemberProfileUpdatedAuditRecord(
        metadata,
        "member-1",
        ["nickname"],
        "",
      ),
    ).toThrow("nickname requires a trimmed");
    expect(() =>
      buildMemberPreferencesUpdatedAuditRecord(
        metadata,
        "member-1",
        ["preferred_locale"],
        "",
        "",
      ),
    ).toThrow("preferred_locale requires");
    expect(() =>
      buildMemberPreferencesUpdatedAuditRecord(
        metadata,
        "member-1",
        ["cookie_consent"],
        "",
        "",
      ),
    ).toThrow("cookie_consent requires");
    expect(() =>
      buildMemberPreferencesUpdatedAuditRecord(
        metadata,
        "member-1",
        ["preferred_locale"],
        "ko",
        "consent-1",
      ),
    ).toThrow("consent_id requires changed_fields");
    expect(() =>
      buildMemberAvatarUpdatedAuditRecord(metadata, "member-1", "added", ""),
    ).toThrow("asset_id");
    expect(() =>
      buildMemberTagCreatedAuditRecord(metadata, "tag-1", " "),
    ).toThrow("tag_name");
    expect(() =>
      validateAuditRecord({
        ...metadata,
        action: "member.updated",
        target_type: "member",
        target_id: "member-1",
        changed_fields: ["tags"],
      } as AuditRecord),
    ).toThrow("tag_ids");
    expect(() =>
      validateAuditRecord({
        ...metadata,
        action: "member.updated",
        target_type: "member",
        target_id: "member-1",
      } as AuditRecord),
    ).toThrow("catalog changed_fields variant");
    expect(() =>
      validateAuditRecord({
        ...metadata,
        action: "member.updated",
        target_type: "member",
        target_id: "member-1",
        changed_fields: ["avatar"],
        collection_operation: "added",
      } as AuditRecord),
    ).toThrow("asset_id");
    expect(() =>
      validateAuditRecord({
        ...metadata,
        action: "member.updated",
        target_type: "member",
        target_id: "member-1",
        changed_fields: ["avatar"],
        collection_operation: "replaced" as never,
        asset_id: "asset-1",
      } as AuditRecord),
    ).toThrow("collection operation");
  });

  it("matches the complete reviewed action-to-target fixture", () => {
    const fixturePath = fileURLToPath(
      new URL("../fixtures/domain-audit-actions.json", import.meta.url),
    );
    const fixture = JSON.parse(readFileSync(fixturePath, "utf8")) as readonly {
      action: string;
      target_type: string;
    }[];
    expect(Object.keys(AUDIT_CATALOG)).toHaveLength(97);
    expect(AUDIT_CATALOG).toEqual(
      Object.fromEntries(
        fixture.map(({ action, target_type }) => [action, target_type]),
      ),
    );
    expect(AUDIT_ACTIONS).toHaveLength(97);
  });

  it("rejects unknown and disallowed runtime attributes", () => {
    const version = {
      ...collabMetadata,
      action: "post.updated",
      target_type: "post",
      target_id: "post-1",
      changed_fields: ["version"],
      version_id: "version-1",
      contributor_member_ids: contributors,
    } as AuditRecord;
    expect(() =>
      validateAuditRecord({ ...version, policy_type: "terms" }),
    ).toThrow("policy_type");
    expect(() =>
      validateAuditRecord({
        ...version,
        submitted_payload: "secret",
      } as AuditRecord),
    ).toThrow("unknown attribute");
  });

  it("rejects missing changed fields and duplicate session identifiers", () => {
    expect(() =>
      validateAuditRecord({
        ...metadata,
        action: "member.updated",
        target_type: "member",
        target_id: "member-1",
      } as AuditRecord),
    ).toThrow("catalog changed_fields variant");
    expect(() =>
      validateAuditRecord({
        ...metadata,
        action: "account.updated",
        target_type: "account",
        target_id: "member-1",
        changed_fields: ["sessions"],
        collection_operation: "removed",
        session_scope: "others",
        session_ids: [
          "018f47a2-8a3d-49e9-9d42-6f12c89b1234",
          "018f47a2-8a3d-49e9-9d42-6f12c89b1234",
        ],
      } as AuditRecord),
    ).toThrow("sorted and unique");
  });

  it("canonicalizes set-shaped attributes", () => {
    expect(canonicalAuditValues([])).toEqual([]);
    expect(
      buildSiteSettingsUpdatedAuditRecord(metadata, [
        "site_title",
        "primary_color",
        "site_title",
      ]).changed_fields,
    ).toEqual(["primary_color", "site_title"]);
    expect(
      buildPostVersionCreatedAuditRecord(
        collabMetadata,
        "post-1",
        "version-1",
        [contributors[1], contributors[0], contributors[1]],
      ).contributor_member_ids,
    ).toEqual(contributors);
    expect(
      buildMemberTagsUpdatedAuditRecord(metadata, "member-1", [
        "tag-b",
        "tag-a",
        "tag-b",
      ]),
    ).toMatchObject({ tag_ids: ["tag-a", "tag-b"] });
    expect(
      buildMemberTagsUpdatedAuditRecord(metadata, "member-1", []),
    ).toMatchObject({
      tag_ids: [],
    });
    expect(
      buildMemberProfileUpdatedAuditRecord(
        metadata,
        "member-1",
        ["nickname"],
        "Name",
      ),
    ).toMatchObject({ nickname: "Name" });
    expect(
      buildMemberPreferencesUpdatedAuditRecord(
        metadata,
        "member-1",
        ["cookie_consent"],
        "",
        "consent-1",
      ),
    ).toMatchObject({ consent_id: "consent-1" });
  });
});
