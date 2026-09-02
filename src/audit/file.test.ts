import { describe, expect, it } from "vitest";

import {
  buildFileCreatedAuditRecord,
  buildFileDeletedAuditRecord,
  buildFileFolderCreatedAuditRecord,
  buildFileFolderDeletedAuditRecord,
  buildFileFolderMovedAuditRecord,
  buildFileFolderRenamedAuditRecord,
  buildFileMovedAuditRecord,
  buildFileRenamedAuditRecord,
  type AuditMetadata,
} from "../audit.ts";
import { validateAuditRecord, type AuditRecord } from "../records.ts";
import { validateFileAuditRecord } from "./file.ts";

const metadata: AuditMetadata = {
  audit_id: "00000000-0000-4000-8000-000000000001",
  occurred_at: "2026-08-09T03:04:05Z",
  request_id: "018f47a2-8a3d-4e17-9d42-6f12c89b1234",
  actor_kind: "member",
  actor_member_id: "member-1",
};

function record(attributes: Partial<AuditRecord>): AuditRecord {
  return {
    ...metadata,
    action: "file.updated",
    target_type: "file",
    target_id: "file-1",
    changed_fields: ["file_name"],
    ...attributes,
  } as AuditRecord;
}

describe("file and folder audit", () => {
  it("builds all lifecycle and semantic update variants", () => {
    const records = [
      buildFileCreatedAuditRecord(metadata, "file-1"),
      buildFileDeletedAuditRecord(metadata, "file-1"),
      buildFileRenamedAuditRecord(metadata, "file-1"),
      buildFileMovedAuditRecord(metadata, "file-1", "folder-1", ""),
      buildFileFolderCreatedAuditRecord(metadata, "folder-1"),
      buildFileFolderDeletedAuditRecord(metadata, "folder-1"),
      buildFileFolderRenamedAuditRecord(metadata, "folder-1"),
      buildFileFolderMovedAuditRecord(metadata, "folder-1", "", "folder-2"),
    ];
    expect(records.map(({ action }) => action)).toEqual([
      "file.created",
      "file.deleted",
      "file.updated",
      "file.updated",
      "file_folder.created",
      "file_folder.deleted",
      "file_folder.updated",
      "file_folder.updated",
    ]);
  });

  it("accepts root moves", () => {
    expect(
      JSON.parse(
        JSON.stringify(
          buildFileMovedAuditRecord(metadata, "file-1", "", "folder-1"),
        ),
      ),
    ).not.toHaveProperty("previous_parent_id");
    expect(
      JSON.parse(
        JSON.stringify(
          buildFileFolderMovedAuditRecord(metadata, "folder-1", "folder-1", ""),
        ),
      ),
    ).not.toHaveProperty("new_parent_id");
    expect(() =>
      validateAuditRecord(
        record({
          changed_fields: ["folder_id"],
          previous_parent_id: "",
          new_parent_id: "folder-1",
        }),
      ),
    ).not.toThrow();
    expect(
      validateFileAuditRecord(
        record({ action: "member.updated", target_type: "member" }),
      ),
    ).toBe(false);
  });

  it("rejects no-ops, malformed fields, target mismatches, and unsupported attributes", () => {
    for (const invalid of [
      record({ target_type: "file_folder" }),
      record({ target_id: "" }),
      record({ changed_fields: [] }),
      record({
        changed_fields: ["folder_id"],
        previous_parent_id: "",
        new_parent_id: "",
      }),
      record({ changed_fields: ["file_name"], previous_parent_id: "folder-1" }),
      record({ changed_fields: ["file_name"], item_ids: [] }),
      record({ changed_fields: ["file_name"], previous_state: "public" }),
      record({ changed_fields: ["file_name", "unknown"] }),
      record({ changed_fields: ["folder_id", "file_name"] }),
      record({ changed_fields: ["file_name"], email: "extra@example.test" }),
      record({ action: "file.created", changed_fields: ["file_name"] }),
      record({
        action: "file_folder.updated",
        target_type: "file_folder",
        target_id: "folder-1",
        changed_fields: ["parent_id"],
        previous_parent_id: "folder-1",
        new_parent_id: "folder-1",
      }),
      record({
        action: "file_folder.updated",
        target_type: "file_folder",
        target_id: "folder-1",
        changed_fields: ["name"],
        item_ids: [],
      }),
    ])
      expect(() => validateAuditRecord(invalid as AuditRecord)).toThrow(
        TypeError,
      );
  });

  it("rejects anonymous actors and permits only the Go-approved system deletion", () => {
    expect(() =>
      validateAuditRecord(record({ actor_kind: "anonymous" } as never)),
    ).toThrow("anonymous");
    expect(() =>
      validateAuditRecord(
        record({
          action: "file.deleted",
          changed_fields: undefined,
          actor_kind: "system",
          actor_member_id: undefined,
          actor_service: "geul-backend",
        }),
      ),
    ).not.toThrow();
    expect(() =>
      validateAuditRecord(
        record({
          actor_kind: "system",
          actor_member_id: undefined,
          actor_service: "geul-backend",
        }),
      ),
    ).toThrow("cannot use system actor");
  });
});
