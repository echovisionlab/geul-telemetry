import { readFileSync } from "node:fs";
import { fileURLToPath } from "node:url";

import { describe, expect, it } from "vitest";

import * as audit from "../audit.ts";
import { validateAuditRecord, type AuditRecord } from "../records.ts";
import type { AuditMetadata } from "./types.ts";

type Fixture = {
  readonly case: string;
  readonly variant: string;
  readonly action: string;
  readonly target_type: string;
  readonly target_id: string;
  readonly actor_kind: "anonymous" | "member" | "system";
  readonly actor_service?: string;
  readonly attributes: Record<string, unknown>;
};

const metadata: AuditMetadata = {
  audit_id: "00000000-0000-4000-8000-000000000001",
  occurred_at: "2026-08-09T03:04:05Z",
  actor_kind: "member",
  actor_member_id: "member-1",
};
const systemMetadata: AuditMetadata = {
  audit_id: metadata.audit_id,
  occurred_at: metadata.occurred_at,
  actor_kind: "system",
  actor_service: "geul-backend",
};
const collabMetadata: AuditMetadata = {
  audit_id: metadata.audit_id,
  occurred_at: metadata.occurred_at,
  actor_kind: "system",
  actor_service: "geul-collab",
};
const anonymousMetadata: AuditMetadata = {
  audit_id: metadata.audit_id,
  occurred_at: metadata.occurred_at,
  actor_kind: "anonymous",
};
const contributors = ["1b6bcad2-c90d-49e9-bec7-f9a4ba6b2894"];
const scheduledAt = "2026-09-01T00:00:00Z";

function readManifest(): readonly Fixture[] {
  return JSON.parse(
    readFileSync(
      fileURLToPath(
        new URL(
          "../../fixtures/domain-audit-wire-parity.json",
          import.meta.url,
        ),
      ),
      "utf8",
    ),
  ) as readonly Fixture[];
}

function serialized(record: AuditRecord): Record<string, unknown> {
  return JSON.parse(JSON.stringify(record)) as Record<string, unknown>;
}

const envelopeKeys = [
  "audit_id",
  "occurred_at",
  "action",
  "target_type",
  "target_id",
  "request_id",
  "trace_id",
  "span_id",
  "actor_kind",
  "actor_member_id",
  "actor_service",
] as const;

function exactAttributes(
  wire: Record<string, unknown>,
): Record<string, unknown> {
  const attributes = { ...wire };
  for (const key of envelopeKeys) delete attributes[key];
  return attributes;
}

function expectExactEnvelope(
  wire: Record<string, unknown>,
  expected: Fixture,
): void {
  expect(wire.audit_id, expected.case).toBe(metadata.audit_id);
  expect(wire.occurred_at, expected.case).toBe(metadata.occurred_at);
  expect(wire.action, expected.case).toBe(expected.action);
  expect(wire.target_type, expected.case).toBe(expected.target_type);
  expect(wire.target_id, expected.case).toBe(expected.target_id);
  expect(wire.request_id, expected.case).toBeUndefined();
  expect(wire.trace_id, expected.case).toBeUndefined();
  expect(wire.span_id, expected.case).toBeUndefined();
  expect(wire.actor_kind, expected.case).toBe(expected.actor_kind);
  if (expected.actor_kind === "member") {
    expect(wire.actor_member_id, expected.case).toBe("member-1");
    expect(wire.actor_service, expected.case).toBeUndefined();
  } else if (expected.actor_kind === "system") {
    expect(wire.actor_member_id, expected.case).toBeUndefined();
    expect(wire.actor_service, expected.case).toBe(expected.actor_service);
  } else {
    expect(wire.actor_member_id, expected.case).toBeUndefined();
    expect(wire.actor_service, expected.case).toBeUndefined();
  }
}

describe("TypeScript semantic builder wire parity", () => {
  it("invokes one public semantic builder for every manifest variant", () => {
    const manifest = readManifest();
    const builders: Record<string, () => AuditRecord> = {
      "post participant author to collaborator": () =>
        audit.buildPostParticipantAuditRecord(
          metadata,
          "post-1",
          "member-2",
          "author",
          "collaborator",
        ),
      "post participant added as author": () =>
        audit.buildPostParticipantAuditRecord(
          metadata,
          "post-1",
          "member-2",
          "none",
          "author",
        ),
      "post participant added as collaborator": () =>
        audit.buildPostParticipantAuditRecord(
          metadata,
          "post-1",
          "member-2",
          "none",
          "collaborator",
        ),
      "post File Block download policy": () =>
        audit.buildPostFileBlockDownloadPolicyAuditRecord(
          metadata,
          "post-1",
          "block-1",
          "file-1",
          "disabled",
          "restricted",
          ["segment-1"],
          ["segment-2"],
        ),
      "page File Block download policy": () =>
        audit.buildPageFileBlockDownloadPolicyAuditRecord(
          metadata,
          "page-1",
          "block-2",
          "file-2",
          "public",
          "authenticated",
          [],
          [],
        ),
      "map theme content": () =>
        audit.buildMapThemeContentUpdatedAuditRecord(metadata, "theme-1"),
      "post source locale": () =>
        audit.buildPostSourceLocaleAuditRecord(
          metadata,
          "post-1",
          "en",
          "zh-CN",
        ),
      "page source locale": () =>
        audit.buildPageSourceLocaleAuditRecord(metadata, "page-1", "en", "ko"),
      "work source locale": () =>
        audit.buildWorkSourceLocaleAuditRecord(metadata, "work-1", "en", "ko"),
      "post series source locale": () =>
        audit.buildPostSeriesSourceLocaleAuditRecord(
          metadata,
          "series-1",
          "en",
          "ko",
        ),
      "program event source locale": () =>
        audit.buildProgramEventSourceLocaleAuditRecord(
          metadata,
          "event-1",
          "en",
          "ko",
        ),
      "release source locale": () =>
        audit.buildReleaseSourceLocaleAuditRecord(
          metadata,
          "release-1",
          "en",
          "ko",
        ),
      "artist source locale": () =>
        audit.buildArtistSourceLocaleAuditRecord(
          metadata,
          "artist-1",
          "en",
          "ko",
        ),
      "label source locale": () =>
        audit.buildLabelSourceLocaleAuditRecord(
          metadata,
          "label-1",
          "en",
          "ko",
        ),
      "menu source locale": () =>
        audit.buildMenuSourceLocaleAuditRecord(metadata, "menu-1", "en", "ko"),
      "campaign source locale": () =>
        audit.buildCampaignSourceLocaleAuditRecord(
          metadata,
          "campaign-1",
          "en",
          "ko",
        ),
      "form source locale": () =>
        audit.buildFormSourceLocaleAuditRecord(metadata, "form-1", "en", "ko"),
      "email template source locale": () =>
        audit.buildEmailTemplateSourceLocaleAuditRecord(
          metadata,
          "template-1",
          "en",
          "ko",
        ),
      "email layout source locale": () =>
        audit.buildEmailLayoutSourceLocaleAuditRecord(
          metadata,
          "layout-1",
          "en",
          "ko",
        ),
      "privacy source locale": () =>
        audit.buildPrivacySourceLocaleAuditRecord(
          metadata,
          "privacy-1",
          2,
          "en",
          "ko",
        ),
      "terms source locale": () =>
        audit.buildTermsSourceLocaleAuditRecord(
          metadata,
          "terms-1",
          1,
          "en",
          "ko",
        ),
      "site settings fields": () =>
        audit.buildSiteSettingsUpdatedAuditRecord(metadata, ["site_title"]),
      "post version": () =>
        audit.buildPostVersionCreatedAuditRecord(
          collabMetadata,
          "post-1",
          "version-1",
          contributors,
        ),
      "post configuration": () =>
        audit.buildPostConfigurationAuditRecord(metadata, "post-1", ["slug"]),
      "post lifecycle status": () =>
        audit.buildPostLifecycleAuditRecord(
          metadata,
          "post-1",
          ["status"],
          "draft",
          "published",
        ),
      "post lifecycle schedule": () =>
        audit.buildPostLifecycleAuditRecord(
          metadata,
          "post-1",
          ["schedule"],
          "draft",
          "scheduled",
          scheduledAt,
          "Asia/Seoul",
        ),
      "post featured image": () =>
        audit.buildPostFeaturedImageAuditRecord(
          metadata,
          "post-1",
          "asset-1",
          "added",
        ),
      "post share link": () =>
        audit.buildPostShareLinkAuditRecord(
          metadata,
          "post-1",
          "link-1",
          "created",
        ),
      "post comment": () =>
        audit.buildPostCommentAuditRecord(
          metadata,
          "post-1",
          "comment-1",
          "updated",
        ),
      "post version restore": () =>
        audit.buildPostVersionRestoreAuditRecord(
          metadata,
          "post-1",
          "version-1",
        ),
      "page version": () =>
        audit.buildPageVersionCreatedAuditRecord(
          collabMetadata,
          "page-1",
          "version-1",
          contributors,
        ),
      "page configuration": () =>
        audit.buildPageConfigurationAuditRecord(metadata, "page-1", ["slug"]),
      "page lifecycle": () =>
        audit.buildPageLifecycleAuditRecord(
          metadata,
          "page-1",
          "draft",
          "published",
        ),
      "page featured image": () =>
        audit.buildPageFeaturedImageAuditRecord(
          metadata,
          "page-1",
          "asset-1",
          "added",
        ),
      "page share link": () =>
        audit.buildPageShareLinkAuditRecord(
          metadata,
          "page-1",
          "link-1",
          "created",
        ),
      "page version restore": () =>
        audit.buildPageVersionRestoreAuditRecord(
          metadata,
          "page-1",
          "version-1",
        ),
      "work version": () =>
        audit.buildWorkVersionCreatedAuditRecord(
          collabMetadata,
          "work-1",
          "version-1",
          contributors,
        ),
      "work metadata": () =>
        audit.buildWorkMetadataAuditRecord(metadata, "work-1", ["slug"]),
      "work lifecycle": () =>
        audit.buildWorkLifecycleAuditRecord(
          metadata,
          "work-1",
          "draft",
          "published",
        ),
      "work featured image": () =>
        audit.buildWorkFeaturedImageAuditRecord(
          metadata,
          "work-1",
          "asset-1",
          "added",
        ),
      "work credit": () =>
        audit.buildWorkCreditAuditRecord(
          metadata,
          "work-1",
          "credit-1",
          "updated",
        ),
      "work share link": () =>
        audit.buildWorkShareLinkAuditRecord(
          metadata,
          "work-1",
          "link-1",
          "created",
        ),
      "work version restore": () =>
        audit.buildWorkVersionRestoreAuditRecord(
          metadata,
          "work-1",
          "version-1",
        ),
      "legal lifecycle": () =>
        audit.buildLegalPolicyLifecycleAuditRecord(
          metadata,
          "policy-1",
          "terms",
          1,
          ["status"],
          "draft",
          "scheduled",
        ),
      "legal lifecycle effective schedule": () =>
        audit.buildLegalPolicyLifecycleAuditRecord(
          metadata,
          "policy-1",
          "terms",
          1,
          ["status", "effective_at"],
          "draft",
          "scheduled",
          scheduledAt,
        ),
      "legal share link": () =>
        audit.buildLegalPolicyShareLinkAuditRecord(
          metadata,
          "policy-1",
          "terms",
          1,
          "created",
          "link-1",
        ),
      "file rename": () =>
        audit.buildFileRenamedAuditRecord(metadata, "file-1"),
      "file move": () =>
        audit.buildFileMovedAuditRecord(metadata, "file-1", "folder-1", ""),
      "file move between folders": () =>
        audit.buildFileMovedAuditRecord(
          metadata,
          "file-1",
          "folder-1",
          "folder-2",
        ),
      "work File Block download policy": () =>
        audit.buildWorkFileBlockDownloadPolicyAuditRecord(
          metadata,
          "work-1",
          "block-3",
          "file-3",
          "restricted",
          "restricted",
          [],
          ["segment-1"],
        ),
      "program event File Block download policy": () =>
        audit.buildProgramEventFileBlockDownloadPolicyAuditRecord(
          metadata,
          "event-1",
          "block-4",
          "file-4",
          "restricted",
          "public",
          ["segment-1"],
          [],
        ),
      "release Track download policy": () =>
        audit.buildReleaseTrackDownloadPolicyAuditRecord(
          metadata,
          "release-1",
          "track-1",
          "file-5",
          "disabled",
          "public",
          [],
          [],
        ),
      "folder rename": () =>
        audit.buildFileFolderRenamedAuditRecord(metadata, "folder-1"),
      "folder move": () =>
        audit.buildFileFolderMovedAuditRecord(
          metadata,
          "folder-1",
          "",
          "folder-2",
        ),
      "folder move between parents": () =>
        audit.buildFileFolderMovedAuditRecord(
          metadata,
          "folder-1",
          "folder-2",
          "folder-3",
        ),
      "category metadata": () =>
        audit.buildCategoryMetadataUpdatedAuditRecord(metadata, "category-1", [
          "name",
        ]),
      "tag metadata": () =>
        audit.buildTagMetadataUpdatedAuditRecord(metadata, "tag-1", ["name"]),
      "genre metadata": () =>
        audit.buildGenreMetadataUpdatedAuditRecord(metadata, "genre-1", [
          "name",
        ]),
      "style metadata": () =>
        audit.buildStyleMetadataUpdatedAuditRecord(metadata, "style-1", [
          "name",
        ]),
      "format metadata": () =>
        audit.buildFormatMetadataUpdatedAuditRecord(metadata, "format-1", [
          "name",
        ]),
      "client metadata": () =>
        audit.buildClientMetadataUpdatedAuditRecord(metadata, "client-1", [
          "name",
        ]),
      "client logo": () =>
        audit.buildClientLogoUpdatedAuditRecord(
          metadata,
          "client-1",
          "light",
          "added",
          "file-1",
        ),
      "map place metadata": () =>
        audit.buildMapPlaceMetadataUpdatedAuditRecord(metadata, "place-1", [
          "name",
        ]),
      "map place image": () =>
        audit.buildMapPlaceImageUpdatedAuditRecord(
          metadata,
          "place-1",
          "added",
          "file-1",
        ),
      "audience config": () =>
        audit.buildAudienceSegmentConfigUpdatedAuditRecord(
          metadata,
          "audience-1",
          ["name"],
        ),
      "audience lifecycle": () =>
        audit.buildAudienceSegmentLifecycleUpdatedAuditRecord(
          metadata,
          "audience-1",
          "active",
          "archived",
        ),
      "menu source": () =>
        audit.buildMenuSourceUpdatedAuditRecord(metadata, "menu-1", ["name"]),
      "mail adapter config": () =>
        audit.buildMailAdapterConfigUpdatedAuditRecord(metadata, "adapter-1", [
          "name",
        ]),
      "translation settings": () =>
        audit.buildTranslationSettingsUpdatedAuditRecord(metadata, [
          "default_locale",
        ]),
      "translation provider config": () =>
        audit.buildTranslationProviderConfigUpdatedAuditRecord(
          metadata,
          "provider-1",
          ["name"],
        ),
      "email suppression release": () =>
        audit.buildEmailSuppressionReleasedAuditRecord(
          metadata,
          "suppression-1",
        ),
      "artist lifecycle": () =>
        audit.buildArtistLifecycleAuditRecord(
          metadata,
          "artist-1",
          "draft",
          "published",
        ),
      "artist gallery": () =>
        audit.buildArtistGalleryAuditRecord(metadata, "artist-1", [
          "file-1",
          "file-2",
        ]),
      "artist participant": () =>
        audit.buildArtistParticipantAuditRecord(
          metadata,
          "artist-1",
          "member-2",
          "none",
          "owner",
        ),
      "artist share link": () =>
        audit.buildArtistShareLinkAuditRecord(
          metadata,
          "artist-1",
          "link-1",
          "created",
        ),
      "label lifecycle": () =>
        audit.buildLabelLifecycleAuditRecord(
          metadata,
          "label-1",
          "draft",
          "published",
        ),
      "label participant": () =>
        audit.buildLabelParticipantAuditRecord(
          metadata,
          "label-1",
          "member-2",
          "none",
          "manager",
        ),
      "label logo": () =>
        audit.buildLabelLogoAuditRecord(
          metadata,
          "label-1",
          "light",
          "added",
          "asset-1",
        ),
      "label share link": () =>
        audit.buildLabelShareLinkAuditRecord(
          metadata,
          "label-1",
          "link-1",
          "created",
        ),
      "post series metadata": () =>
        audit.buildPostSeriesSourceMetadataAuditRecord(metadata, "series-1", [
          "slug",
        ]),
      "post series lifecycle": () =>
        audit.buildPostSeriesLifecycleAuditRecord(
          metadata,
          "series-1",
          "draft",
          "published",
        ),
      "post series manager": () =>
        audit.buildPostSeriesManagerAuditRecord(
          metadata,
          "series-1",
          "member-2",
          "none",
          "manager",
        ),
      "post series membership": () =>
        audit.buildPostSeriesMembershipAuditRecord(
          metadata,
          "series-1",
          "post-1",
          "",
          "series-1",
        ),
      "post series membership move": () =>
        audit.buildPostSeriesMembershipAuditRecord(
          metadata,
          "series-1",
          "post-1",
          "series-2",
          "series-1",
        ),
      "post series membership clear": () =>
        audit.buildPostSeriesMembershipAuditRecord(
          metadata,
          "series-1",
          "post-1",
          "series-1",
          "",
        ),
      "post series order": () =>
        audit.buildPostSeriesOrderAuditRecord(metadata, "series-1", [
          "post-2",
          "post-1",
        ]),
      "post series featured image": () =>
        audit.buildPostSeriesFeaturedImageAuditRecord(
          metadata,
          "series-1",
          "added",
          "file-1",
        ),
      "program event type config": () =>
        audit.buildProgramEventTypeConfigUpdatedAuditRecord(
          metadata,
          "type-1",
          ["slug"],
        ),
      "program event metadata": () =>
        audit.buildProgramEventMetadataAuditRecord(metadata, "event-1", [
          "slug",
        ]),
      "program event poster": () =>
        audit.buildProgramEventPosterAuditRecord(
          metadata,
          "event-1",
          "added",
          "file-1",
        ),
      "program event media": () =>
        audit.buildProgramEventChildAuditRecord(
          metadata,
          "event-1",
          "media",
          "media-1",
          "created",
        ),
      "program event child order": () =>
        audit.buildProgramEventChildOrderAuditRecord(
          metadata,
          "event-1",
          "credits",
          ["credit-2", "credit-1"],
        ),
      "program event lifecycle": () =>
        audit.buildProgramEventLifecycleAuditRecord(
          metadata,
          "event-1",
          "draft",
          "published",
        ),
      "program event series metadata": () =>
        audit.buildProgramEventSeriesMetadataAuditRecord(
          metadata,
          "event-series-1",
          ["title"],
        ),
      "program event series poster": () =>
        audit.buildProgramEventSeriesPosterAuditRecord(
          metadata,
          "event-series-1",
          "added",
          "file-1",
        ),
      "program event series lifecycle": () =>
        audit.buildProgramEventSeriesLifecycleAuditRecord(
          metadata,
          "event-series-1",
          "draft",
          "published",
        ),
      "release metadata": () =>
        audit.buildReleaseMetadataAuditRecord(metadata, "release-1", ["slug"]),
      "release track": () =>
        audit.buildReleaseTrackAuditRecord(
          metadata,
          "release-1",
          "track-1",
          "created",
        ),
      "release track order": () =>
        audit.buildReleaseTrackOrderAuditRecord(metadata, "release-1", [
          "track-2",
          "track-1",
        ]),
      "release track audio": () =>
        audit.buildReleaseTrackAudioAuditRecord(
          metadata,
          "release-1",
          "track-1",
          "file-1",
          "added",
        ),
      "release artwork": () =>
        audit.buildReleaseArtworkAuditRecord(
          metadata,
          "release-1",
          "file-1",
          "added",
        ),
      "release lifecycle": () =>
        audit.buildReleaseLifecycleAuditRecord(
          metadata,
          "release-1",
          "draft",
          "published",
        ),
      "release share link": () =>
        audit.buildReleaseShareLinkAuditRecord(
          metadata,
          "release-1",
          "link-1",
          "created",
        ),
      "campaign target layout": () =>
        audit.buildCampaignTargetLayoutAuditRecord(metadata, "campaign-1", [
          "layout",
        ]),
      "campaign lifecycle": () =>
        audit.buildCampaignStatusLifecycleAuditRecord(
          metadata,
          "campaign-1",
          "draft",
          "scheduled",
        ),
      "campaign schedule": () =>
        audit.buildCampaignScheduleLifecycleAuditRecord(
          metadata,
          "campaign-1",
          "draft",
          "scheduled",
          scheduledAt,
        ),
      "campaign delivery run lifecycle": () =>
        audit.buildCampaignDeliveryRunLifecycleAuditRecord(
          metadata,
          "campaign-1",
          "scheduled",
          "sending",
          "run-1",
        ),
      "template metadata": () =>
        audit.buildEmailTemplateMetadataAuditRecord(metadata, "template-1", [
          "name",
        ]),
      "template layout": () =>
        audit.buildEmailTemplateLayoutRelationAuditRecord(
          metadata,
          "template-1",
          "",
          "layout-1",
        ),
      "template layout replacement": () =>
        audit.buildEmailTemplateLayoutRelationAuditRecord(
          metadata,
          "template-1",
          "layout-2",
          "layout-1",
        ),
      "template layout cleared": () =>
        audit.buildEmailTemplateLayoutRelationAuditRecord(
          metadata,
          "template-1",
          "layout-1",
          "",
        ),
      "layout metadata": () =>
        audit.buildEmailLayoutMetadataAuditRecord(metadata, "layout-1", [
          "name",
        ]),
      "email event mapping": () =>
        audit.buildEmailEventMappingTemplateAuditRecord(
          metadata,
          "welcome",
          "",
          "template-1",
        ),
      "email event mapping replacement": () =>
        audit.buildEmailEventMappingTemplateAuditRecord(
          metadata,
          "welcome",
          "template-2",
          "template-1",
        ),
      "email event mapping cleared": () =>
        audit.buildEmailEventMappingTemplateAuditRecord(
          metadata,
          "welcome",
          "template-1",
          "",
        ),
      "form settings": () =>
        audit.buildFormSettingsAuditRecord(metadata, "form-1", ["slug"]),
      "form lifecycle": () =>
        audit.buildFormLifecycleAuditRecord(
          metadata,
          "form-1",
          "draft",
          "published",
        ),
      "form featured image": () =>
        audit.buildFormFeaturedImageAuditRecord(
          metadata,
          "form-1",
          "file-1",
          "added",
        ),
      "form share link": () =>
        audit.buildFormShareLinkAuditRecord(
          metadata,
          "form-1",
          "link-1",
          "form",
          "created",
        ),
      "anonymous form submission": () =>
        audit.buildFormSubmissionCreatedAuditRecord(
          anonymousMetadata,
          "submission-1",
          "form-1",
        ),
      "member onboarding": () =>
        audit.buildMemberOnboardingCompletedAuditRecord(
          metadata,
          "member-1",
          "Member",
        ),
      "member profile": () =>
        audit.buildMemberProfileUpdatedAuditRecord(
          metadata,
          "member-1",
          ["nickname"],
          "Member",
        ),
      "member profile fields without nickname": () =>
        audit.buildMemberProfileUpdatedAuditRecord(
          metadata,
          "member-1",
          ["website", "bio"],
          "",
        ),
      "member avatar": () =>
        audit.buildMemberAvatarUpdatedAuditRecord(
          metadata,
          "member-1",
          "added",
          "asset-1",
        ),
      "member preference": () =>
        audit.buildMemberPreferencesUpdatedAuditRecord(
          metadata,
          "member-1",
          ["preferred_locale"],
          "ko",
          "",
        ),
      "member cookie consent preference": () =>
        audit.buildMemberPreferencesUpdatedAuditRecord(
          metadata,
          "member-1",
          ["cookie_consent"],
          "",
          "consent-1",
        ),
      "member tags": () =>
        audit.buildMemberTagsUpdatedAuditRecord(metadata, "member-1", [
          "tag-2",
          "tag-1",
        ]),
      "member role": () =>
        audit.buildMemberRoleUpdatedAuditRecord(
          metadata,
          "member-1",
          "user",
          "author",
        ),
      "member status": () =>
        audit.buildMemberBannedAuditRecord(metadata, "member-1"),
      "member unban system": () =>
        audit.buildMemberUnbannedAuditRecord(systemMetadata, "member-1"),
      "account canonical email": () =>
        audit.buildAccountCanonicalEmailUpdatedAuditRecord(
          metadata,
          "member-1",
          "old@example.test",
          "new@example.test",
        ),
      "account email login": () =>
        audit.buildAccountEmailLoginAddedAuditRecord(
          metadata,
          "member-1",
          "new@example.test",
        ),
      "account social login": () =>
        audit.buildAccountSocialLoginAddedAuditRecord(
          metadata,
          "member-1",
          "google",
          "google-subject",
        ),
      "account passkey": () =>
        audit.buildAccountPasskeyAddedAuditRecord(metadata, "member-1", [
          "passkey-1",
        ]),
      "account sessions": () =>
        audit.buildAccountSessionRevokedAuditRecord(
          metadata,
          "member-1",
          "others",
          [
            "018f47a2-8a3d-4e17-9d42-6f12c89b5678",
            "018f47a2-8a3d-4e17-9d42-6f12c89b1234",
          ],
        ),
      "account newsletter": () =>
        audit.buildAccountNewsletterSubscriptionUpdatedAuditRecord(
          metadata,
          "member-1",
          "subscribed",
          "unsubscribed",
        ),
      "account deletion lifecycle": () =>
        audit.buildAccountDeletionRequestedAuditRecord(
          metadata,
          "member-1",
          "none",
        ),
      "account deletion scheduled": () =>
        audit.buildAccountDeletionScheduledAuditRecord(
          metadata,
          "member-1",
          "confirmation_pending",
        ),
      "account deletion cancelled": () =>
        audit.buildAccountDeletionCancelledAuditRecord(metadata, "member-1"),
      "account deletion recovered": () =>
        audit.buildAccountDeletionRecoveredAuditRecord(metadata, "member-1"),
      "account deleted system": () =>
        audit.buildAccountDeletedAuditRecord(systemMetadata, "member-1"),
    };

    expect(Object.keys(builders)).toHaveLength(manifest.length);
    expect(new Set(manifest.map((entry) => entry.case)).size).toBe(
      manifest.length,
    );
    const seen = new Set<string>();
    for (const expected of manifest) {
      const build = builders[expected.case];
      expect(
        build,
        `${expected.case} has no TypeScript semantic builder`,
      ).toBeDefined();
      expect(seen.has(expected.case), expected.case).toBe(false);
      seen.add(expected.case);
      const record = build!();
      validateAuditRecord(record);
      const wire = serialized(record);
      expectExactEnvelope(wire, expected);
      expect(exactAttributes(wire), expected.case).toEqual(expected.attributes);
    }
    expect(seen.size).toBe(Object.keys(builders).length);
  });
});
