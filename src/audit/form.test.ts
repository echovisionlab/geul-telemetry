import { describe, expect, it } from "vitest";

import {
  buildFormCreatedAuditRecord,
  buildFormDeletedAuditRecord,
  buildFormFeaturedImageAuditRecord,
  buildFormLifecycleAuditRecord,
  buildFormSettingsAuditRecord,
  buildFormShareLinkAuditRecord,
  buildFormSubmissionCreatedAuditRecord,
  buildFormSubmissionDeletedAuditRecord,
  type AuditMetadata,
} from "../audit.ts";
import { validateAuditRecord, type AuditRecord } from "../records.ts";
import { validateFormAuditRecord } from "./form.ts";

const metadata: AuditMetadata = {
  audit_id: "00000000-0000-4000-8000-000000000001",
  occurred_at: "2026-08-09T03:04:05Z",
  request_id: "018f47a2-8a3d-4e17-9d42-6f12c89b1234",
  actor_kind: "member",
  actor_member_id: "member-1",
};
const anonymousMetadata: AuditMetadata = {
  audit_id: metadata.audit_id,
  occurred_at: metadata.occurred_at,
  request_id: metadata.request_id,
  actor_kind: "anonymous",
};

function formRecord(attributes: Partial<AuditRecord> = {}): AuditRecord {
  return {
    ...metadata,
    action: "form.updated",
    target_type: "form",
    target_id: "form-1",
    changed_fields: ["slug"],
    ...attributes,
  } as AuditRecord;
}

describe("form and form submission audit", () => {
  it("builds every reviewed root and update variant", () => {
    const records = [
      buildFormCreatedAuditRecord(metadata, "form-1"),
      buildFormDeletedAuditRecord(metadata, "form-1"),
      buildFormSettingsAuditRecord(metadata, "form-1", ["slug", "limit"]),
      buildFormLifecycleAuditRecord(metadata, "form-1", "draft", "published"),
      buildFormFeaturedImageAuditRecord(metadata, "form-1", "file-1", "added"),
      buildFormShareLinkAuditRecord(
        metadata,
        "form-1",
        "link-1",
        "form",
        "created",
      ),
      buildFormShareLinkAuditRecord(
        metadata,
        "form-1",
        "link-2",
        "dashboard",
        "deleted",
      ),
      buildFormSubmissionCreatedAuditRecord(
        anonymousMetadata,
        "submission-1",
        "form-1",
      ),
      buildFormSubmissionCreatedAuditRecord(metadata, "submission-2", "form-1"),
      buildFormSubmissionDeletedAuditRecord(metadata, "submission-1"),
    ];
    expect(records[2]).toMatchObject({ changed_fields: ["limit", "slug"] });
    expect(records[7]).toMatchObject({
      actor_kind: "anonymous",
      parent_id: "form-1",
    });
    expect(records.map((record) => record.action)).toEqual([
      "form.created",
      "form.deleted",
      "form.updated",
      "form.updated",
      "form.updated",
      "form.updated",
      "form.updated",
      "form_submission.created",
      "form_submission.created",
      "form_submission.deleted",
    ]);
  });

  it("permits only anonymous submission creation", () => {
    const { actor_member_id: memberID, ...systemRejectedForm } = formRecord();
    expect(memberID).toBe("member-1");
    expect(() =>
      validateAuditRecord({
        ...systemRejectedForm,
        actor_kind: "anonymous",
        actor_member_id: undefined,
      }),
    ).toThrow("anonymous");
    expect(() =>
      validateAuditRecord({
        ...metadata,
        actor_kind: "anonymous",
        actor_member_id: undefined,
        action: "form_submission.created",
        target_type: "form_submission",
        target_id: "submission-1",
        parent_id: "form-1",
      }),
    ).not.toThrow();
    expect(() =>
      validateAuditRecord({
        ...metadata,
        action: "form_submission.deleted",
        target_type: "form_submission",
        target_id: "submission-1",
        actor_kind: "anonymous",
        actor_member_id: undefined,
      }),
    ).toThrow("anonymous");
    expect(() =>
      validateAuditRecord({
        ...systemRejectedForm,
        actor_kind: "system",
        actor_service: "geul-backend",
      }),
    ).toThrow("system actor");
  });

  it("rejects no-ops, unsupported fields, malformed variants, and submission PII", () => {
    const submission = {
      ...metadata,
      action: "form_submission.created",
      target_type: "form_submission",
      target_id: "submission-1",
      parent_id: "form-1",
    } as AuditRecord;
    for (const invalid of [
      formRecord({ target_type: "form_submission" }),
      formRecord({ changed_fields: [] }),
      formRecord({ changed_fields: ["unknown"] }),
      formRecord({ changed_fields: ["slug"], email: "secret@example.test" }),
      formRecord({
        changed_fields: ["status"],
        previous_state: "draft",
        new_state: "draft",
      }),
      formRecord({
        changed_fields: ["status"],
        previous_state: "active",
        new_state: "published",
      }),
      formRecord({ changed_fields: ["featured_image"], file_id: "file-1" }),
      formRecord({
        changed_fields: ["share_links"],
        item_id: "link-1",
        item_scope: "form",
        item_operation: "updated",
      }),
      formRecord({
        changed_fields: ["share_links"],
        item_id: "link-1",
        item_scope: "bad" as never,
        item_operation: "created",
      }),
      { ...submission, parent_id: "" },
      { ...submission, email: "visitor@example.test" },
      { ...submission, item_id: "field-value" },
      { ...submission, target_id: "" },
    ])
      expect(() => validateFormAuditRecord(invalid as AuditRecord)).toThrow(
        TypeError,
      );
    expect(
      validateFormAuditRecord(
        formRecord({ action: "member.updated", target_type: "member" }),
      ),
    ).toBe(false);
  });
});
