import { describe, expect, it } from "vitest";

import {
  buildPageFileBlockDownloadPolicyAuditRecord,
  buildPostFileBlockDownloadPolicyAuditRecord,
  buildProgramEventFileBlockDownloadPolicyAuditRecord,
  buildReleaseTrackDownloadPolicyAuditRecord,
  buildWorkFileBlockDownloadPolicyAuditRecord,
  type AuditMetadata,
} from "../audit.ts";
import { validateAuditRecord, type AuditRecord } from "../records.ts";
import { validateRelationDownloadPolicyAuditRecord } from "./relation-download-policy.ts";

const metadata: AuditMetadata = {
  audit_id: "00000000-0000-4000-8000-000000000001",
  occurred_at: "2026-08-09T03:04:05Z",
  actor_kind: "member",
  actor_member_id: "member-1",
};

describe("relation download policy audit", () => {
  it("targets each owning domain and exact File Block or Track relation", () => {
    const records = [
      buildPostFileBlockDownloadPolicyAuditRecord(
        metadata,
        "post-1",
        "block-1",
        "file-1",
        "disabled",
        "public",
        [],
        [],
      ),
      buildPageFileBlockDownloadPolicyAuditRecord(
        metadata,
        "page-1",
        "block-2",
        "file-2",
        "public",
        "authenticated",
        [],
        [],
      ),
      buildWorkFileBlockDownloadPolicyAuditRecord(
        metadata,
        "work-1",
        "block-3",
        "file-3",
        "authenticated",
        "restricted",
        [],
        [],
      ),
      buildProgramEventFileBlockDownloadPolicyAuditRecord(
        metadata,
        "event-1",
        "block-4",
        "file-4",
        "restricted",
        "disabled",
        [],
        [],
      ),
      buildReleaseTrackDownloadPolicyAuditRecord(
        metadata,
        "release-1",
        "track-1",
        "file-5",
        "disabled",
        "public",
        [],
        [],
      ),
    ];
    expect(
      records.map(({ action, target_type, target_id, item_id, file_id }) => ({
        action,
        target_type,
        target_id,
        item_id,
        file_id,
      })),
    ).toEqual([
      {
        action: "post.updated",
        target_type: "post",
        target_id: "post-1",
        item_id: "block-1",
        file_id: "file-1",
      },
      {
        action: "page.updated",
        target_type: "page",
        target_id: "page-1",
        item_id: "block-2",
        file_id: "file-2",
      },
      {
        action: "work.updated",
        target_type: "work",
        target_id: "work-1",
        item_id: "block-3",
        file_id: "file-3",
      },
      {
        action: "program_event.updated",
        target_type: "program_event",
        target_id: "event-1",
        item_id: "block-4",
        file_id: "file-4",
      },
      {
        action: "release.updated",
        target_type: "release",
        target_id: "release-1",
        item_id: "track-1",
        file_id: "file-5",
      },
    ]);
  });

  it("emits only changed before and after policy attributes", () => {
    const combined = buildPostFileBlockDownloadPolicyAuditRecord(
      metadata,
      "post-1",
      "block-1",
      "file-1",
      "public",
      "restricted",
      ["segment-2", "segment-1"],
      [],
    );
    expect(combined).toMatchObject({
      changed_fields: [
        "file_download_audience",
        "file_download_audience_segment_ids",
      ],
      item_id: "block-1",
      file_id: "file-1",
      previous_state: "public",
      new_state: "restricted",
      previous_item_ids: ["segment-1", "segment-2"],
      item_ids: [],
    });
    const segmentOnly = buildReleaseTrackDownloadPolicyAuditRecord(
      metadata,
      "release-1",
      "track-1",
      "file-2",
      "restricted",
      "restricted",
      [],
      ["segment-1"],
    );
    expect(segmentOnly.changed_fields).toEqual([
      "file_download_audience_segment_ids",
    ]);
    expect(segmentOnly.previous_state).toBeUndefined();
    expect(segmentOnly.new_state).toBeUndefined();
  });

  it("rejects malformed, no-op, and non-member transitions", () => {
    expect(() =>
      buildPostFileBlockDownloadPolicyAuditRecord(
        metadata,
        "post-1",
        "block-1",
        "file-1",
        "public",
        "public",
        ["segment-1"],
        ["segment-1"],
      ),
    ).toThrow(TypeError);
    expect(() =>
      buildPageFileBlockDownloadPolicyAuditRecord(
        metadata,
        "page-1",
        "block-1",
        "file-1",
        "draft" as never,
        "public",
        [],
        [],
      ),
    ).toThrow(TypeError);
    expect(() =>
      buildWorkFileBlockDownloadPolicyAuditRecord(
        metadata,
        "work-1",
        "block-1",
        "file-1",
        "public",
        "public",
        [" bad"],
        [" bad"],
      ),
    ).toThrow(TypeError);
    const valid = buildProgramEventFileBlockDownloadPolicyAuditRecord(
      metadata,
      "event-1",
      "block-1",
      "file-1",
      "disabled",
      "public",
      [],
      [],
    );
    for (const invalid of [
      {
        ...valid,
        actor_kind: "system",
        actor_member_id: undefined,
        actor_service: "geul-backend",
      },
      { ...valid, item_id: "" },
      { ...valid, file_id: "" },
      { ...valid, previous_state: "draft" },
      { ...valid, new_state: "public", previous_state: "public" },
      {
        ...valid,
        changed_fields: ["file_download_audience_segment_ids"],
        previous_state: undefined,
        new_state: undefined,
        previous_item_ids: [],
        item_ids: [],
      },
      { ...valid, email: "unsupported@example.test" },
    ])
      expect(() => validateAuditRecord(invalid as AuditRecord)).toThrow(
        TypeError,
      );
  });

  it("validates every relation policy attribute boundary directly", () => {
    const valid = buildPostFileBlockDownloadPolicyAuditRecord(
      metadata,
      "post-1",
      "block-1",
      "file-1",
      "disabled",
      "restricted",
      ["segment-1"],
      ["segment-2"],
    );
    const invalid = [
      { ...valid, changed_fields: [] },
      { ...valid, changed_fields: ["unknown"] },
      {
        ...valid,
        changed_fields: [
          "file_download_audience_segment_ids",
          "file_download_audience",
        ],
      },
      {
        ...valid,
        actor_kind: "system",
        actor_member_id: undefined,
        actor_service: "geul-backend",
      },
      { ...valid, item_id: " bad" },
      { ...valid, file_id: "x".repeat(256) },
      {
        ...valid,
        changed_fields: ["file_download_audience_segment_ids"],
        previous_state: "disabled",
        new_state: "restricted",
      },
      {
        ...valid,
        changed_fields: ["file_download_audience"],
        previous_item_ids: [],
        item_ids: ["segment-2"],
      },
      { ...valid, previous_item_ids: undefined },
      { ...valid, item_ids: ["segment-2", "segment-2"] },
      { ...valid, item_ids: ["segment-2", " bad"] },
      { ...valid, item_ids: ["segment-2", "x".repeat(256)] },
    ];
    for (const record of invalid)
      expect(() =>
        validateRelationDownloadPolicyAuditRecord(record as AuditRecord),
      ).toThrow(TypeError);
  });
});
