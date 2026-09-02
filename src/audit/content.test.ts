import { describe, expect, it } from "vitest";

import {
  buildPageConfigurationAuditRecord,
  buildPageFeaturedImageAuditRecord,
  buildPageLifecycleAuditRecord,
  buildPageShareLinkAuditRecord,
  buildPageVersionCreatedAuditRecord,
  buildPageVersionRestoreAuditRecord,
  buildPostCommentAuditRecord,
  buildPostConfigurationAuditRecord,
  buildPostFeaturedImageAuditRecord,
  buildPostLifecycleAuditRecord,
  buildPostParticipantAuditRecord,
  buildPostShareLinkAuditRecord,
  buildPostVersionRestoreAuditRecord,
  buildPostVersionCreatedAuditRecord,
  buildWorkCreditAuditRecord,
  buildWorkFeaturedImageAuditRecord,
  buildWorkLifecycleAuditRecord,
  buildWorkMetadataAuditRecord,
  buildWorkShareLinkAuditRecord,
  buildWorkVersionRestoreAuditRecord,
  buildWorkVersionCreatedAuditRecord,
  type AuditMetadata,
} from "../audit.ts";
import { validateAuditRecord, type AuditRecord } from "../records.ts";
import { validateContentAuditRecord } from "./content.ts";

const metadata: AuditMetadata = {
  audit_id: "00000000-0000-4000-8000-000000000001",
  occurred_at: "2026-08-09T03:04:05Z",
  actor_kind: "member",
  actor_member_id: "member-1",
};

describe("content audit", () => {
  it("builds every non-Version reviewed Post/Page/Work update shape", () => {
    const records = [
      buildPostConfigurationAuditRecord(metadata, "post-1", ["slug"]),
      buildPostConfigurationAuditRecord(metadata, "post-1", [
        "comments_enabled",
        "slug",
      ]),
      buildPostLifecycleAuditRecord(
        metadata,
        "post-1",
        ["schedule"],
        "draft",
        "scheduled",
        "2026-09-01T00:00:00Z",
        "Asia/Seoul",
      ),
      buildPostLifecycleAuditRecord(
        metadata,
        "post-1",
        ["status"],
        "draft",
        "published",
      ),
      buildPostFeaturedImageAuditRecord(metadata, "post-1", "asset-1", "added"),
      buildPostParticipantAuditRecord(
        metadata,
        "post-1",
        "member-2",
        "none",
        "author",
      ),
      buildPostParticipantAuditRecord(
        metadata,
        "post-1",
        "member-2",
        "author",
        "collaborator",
      ),
      buildPostShareLinkAuditRecord(metadata, "post-1", "link-1", "created"),
      buildPostCommentAuditRecord(metadata, "post-1", "comment-1", "updated"),
      buildPostVersionRestoreAuditRecord(metadata, "post-1", "version-1"),
      buildPageConfigurationAuditRecord(metadata, "page-1", ["slug"]),
      buildPageLifecycleAuditRecord(metadata, "page-1", "draft", "published"),
      buildPageFeaturedImageAuditRecord(
        metadata,
        "page-1",
        "asset-1",
        "removed",
      ),
      buildPageShareLinkAuditRecord(metadata, "page-1", "link-1", "deleted"),
      buildPageVersionRestoreAuditRecord(metadata, "page-1", "version-1"),
      buildWorkMetadataAuditRecord(metadata, "work-1", ["slug"]),
      buildWorkLifecycleAuditRecord(metadata, "work-1", "draft", "published"),
      buildWorkFeaturedImageAuditRecord(metadata, "work-1", "asset-1", "added"),
      buildWorkCreditAuditRecord(metadata, "work-1", "credit-1", "updated"),
      buildWorkShareLinkAuditRecord(metadata, "work-1", "link-1", "created"),
      buildWorkVersionRestoreAuditRecord(metadata, "work-1", "version-1"),
    ];
    expect(records).toHaveLength(21);
    expect(records[5].changed_fields).toEqual(["authors"]);
    expect(records[6].changed_fields).toEqual(["authors", "collaborators"]);
  });

  it("rejects a post participant field that does not represent its transition", () => {
    const record = {
      ...metadata,
      action: "post.updated",
      target_type: "post",
      target_id: "post-1",
      changed_fields: ["authors"],
      subject_member_id: "member-2",
      previous_relationship: "author",
      new_relationship: "collaborator",
    } as AuditRecord;
    expect(() => validateAuditRecord(record)).toThrow("changed_fields");
    expect(() =>
      buildPostParticipantAuditRecord(
        metadata,
        "post-1",
        "member-2",
        "author",
        "author",
      ),
    ).toThrow("distinct");
    expect(() =>
      validateAuditRecord({
        ...buildPostConfigurationAuditRecord(metadata, "post-1", ["slug"]),
        changed_fields: ["slug", "comments_enabled"],
      }),
    ).toThrow("changed_fields");
  });

  it("rejects every content variant's malformed required shape", () => {
    const invalid = [
      () => buildPostConfigurationAuditRecord(metadata, "post-1", []),
      () =>
        buildPostVersionCreatedAuditRecord(metadata, "post-1", "version-1", []),
      () =>
        buildPostLifecycleAuditRecord(
          metadata,
          "post-1",
          ["schedule"],
          "draft",
          "scheduled",
        ),
      () =>
        buildPostLifecycleAuditRecord(
          metadata,
          "post-1",
          ["status"],
          "draft",
          "published",
          "2026-09-01T00:00:00Z",
          "Asia/Seoul",
        ),
      () =>
        buildPostLifecycleAuditRecord(
          metadata,
          "post-1",
          ["schedule"],
          "draft",
          "scheduled",
          "invalid-time",
          "Asia/Seoul",
        ),
      () =>
        buildPostLifecycleAuditRecord(
          metadata,
          "post-1",
          ["schedule"],
          "draft",
          "scheduled",
          "2026-09-01T00:00:00Z",
          "",
        ),
      () => buildPostFeaturedImageAuditRecord(metadata, "post-1", "", "added"),
      () =>
        buildPostFeaturedImageAuditRecord(
          metadata,
          "post-1",
          "asset-1",
          "updated" as never,
        ),
      () => buildPostShareLinkAuditRecord(metadata, "post-1", "", "created"),
      () =>
        buildPostShareLinkAuditRecord(
          metadata,
          "post-1",
          "link-1",
          "updated" as never,
        ),
      () => buildPostCommentAuditRecord(metadata, "post-1", "", "updated"),
      () =>
        buildPostCommentAuditRecord(
          metadata,
          "post-1",
          "comment-1",
          "invalid" as never,
        ),
      () => buildPostVersionRestoreAuditRecord(metadata, "post-1", ""),
      () => buildPageConfigurationAuditRecord(metadata, "page-1", []),
      () => buildPageLifecycleAuditRecord(metadata, "page-1", "draft", "draft"),
      () =>
        buildPageFeaturedImageAuditRecord(metadata, "page-1", "", "removed"),
      () => buildPageShareLinkAuditRecord(metadata, "page-1", "", "deleted"),
      () => buildPageVersionRestoreAuditRecord(metadata, "page-1", ""),
      () => buildWorkMetadataAuditRecord(metadata, "work-1", []),
      () => buildWorkLifecycleAuditRecord(metadata, "work-1", "draft", "draft"),
      () => buildWorkFeaturedImageAuditRecord(metadata, "work-1", "", "added"),
      () => buildWorkCreditAuditRecord(metadata, "work-1", "", "updated"),
      () =>
        buildWorkCreditAuditRecord(
          metadata,
          "work-1",
          "credit-1",
          "invalid" as never,
        ),
      () => buildWorkShareLinkAuditRecord(metadata, "work-1", "", "created"),
      () => buildWorkVersionRestoreAuditRecord(metadata, "work-1", ""),
    ];
    for (const build of invalid) expect(build).toThrow(TypeError);
    expect(() =>
      buildPostCommentAuditRecord(
        {
          audit_id: metadata.audit_id,
          occurred_at: metadata.occurred_at,
          actor_kind: "system",
          actor_service: "geul-backend",
        },
        "post-1",
        "comment-1",
        "updated",
      ),
    ).toThrow("system actor");
  });

  it("keeps the single-category post participant transitions exact", () => {
    expect(
      buildPostParticipantAuditRecord(
        metadata,
        "post-1",
        "member-2",
        "none",
        "collaborator",
      ).changed_fields,
    ).toEqual(["collaborators"]);
    expect(() =>
      validateAuditRecord({
        ...metadata,
        action: "post.updated",
        target_type: "post",
        target_id: "post-1",
        changed_fields: ["authors"],
      } as AuditRecord),
    ).toThrow("relationships");
  });

  it("permits only collab Version checkpoints for system content actors", () => {
    const collabMetadata: AuditMetadata = {
      audit_id: metadata.audit_id,
      occurred_at: metadata.occurred_at,
      actor_kind: "system",
      actor_service: "geul-collab",
    };
    const contributors = ["1b6bcad2-c90d-49e9-bec7-f9a4ba6b2894"];
    const cases = [
      {
        version: (m: AuditMetadata) =>
          buildPostVersionCreatedAuditRecord(
            m,
            "post-1",
            "version-1",
            contributors,
          ),
        memberOnly: (m: AuditMetadata) =>
          buildPostConfigurationAuditRecord(m, "post-1", ["slug"]),
      },
      {
        version: (m: AuditMetadata) =>
          buildPageVersionCreatedAuditRecord(
            m,
            "page-1",
            "version-1",
            contributors,
          ),
        memberOnly: (m: AuditMetadata) =>
          buildPageConfigurationAuditRecord(m, "page-1", ["slug"]),
      },
      {
        version: (m: AuditMetadata) =>
          buildWorkVersionCreatedAuditRecord(
            m,
            "work-1",
            "version-1",
            contributors,
          ),
        memberOnly: (m: AuditMetadata) =>
          buildWorkMetadataAuditRecord(m, "work-1", ["slug"]),
      },
    ];
    for (const entry of cases) {
      expect(() => entry.version(collabMetadata)).not.toThrow();
      expect(() => entry.memberOnly(collabMetadata)).toThrow("system actor");
      expect(() => entry.version(metadata)).toThrow("version requires");
      expect(() =>
        entry.version({ ...collabMetadata, actor_service: "geul-backend" }),
      ).toThrow("system actor");
    }
  });

  it("rejects system actors before non-Version content variants", () => {
    const systemPost = {
      audit_id: metadata.audit_id,
      occurred_at: metadata.occurred_at,
      actor_kind: "system",
      actor_service: "geul-collab",
      action: "post.updated",
      target_type: "post",
      target_id: "post-1",
    } as AuditRecord;
    expect(() =>
      validateContentAuditRecord({ ...systemPost, changed_fields: ["slug"] }),
    ).toThrow("limited to geul-collab version checkpoints");
    expect(() =>
      validateContentAuditRecord({ ...systemPost, changed_fields: [] }),
    ).toThrow("limited to geul-collab version checkpoints");
    expect(() =>
      validateContentAuditRecord({
        ...systemPost,
        actor_kind: "anonymous",
        actor_service: undefined,
        changed_fields: ["comments"],
        item_operation: "created",
        item_id: "comment-1",
      }),
    ).toThrow("post comment requires a member actor");
  });
});
