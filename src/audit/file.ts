import type { AuditRecord } from "../records.ts";
import { assertOnlyAuditAttributes } from "./attributes.ts";
import { buildAuditRecord } from "./builder.ts";
import { AUDIT_CATALOG, requireCatalogTarget } from "./catalog.ts";
import type { AuditMetadata } from "./types.ts";

type FileAction =
  | "file.created"
  | "file.updated"
  | "file.deleted"
  | "file_folder.created"
  | "file_folder.updated"
  | "file_folder.deleted";
const fileUpdateFields = new Set(["file_name", "folder_id"]);
const folderUpdateFields = new Set(["name", "parent_id"]);
function buildFileAuditRecord(
  metadata: AuditMetadata,
  action: FileAction,
  targetId: string,
  attributes: Parameters<typeof buildAuditRecord>[2] = {},
): AuditRecord {
  // The reviewed catalog, not string construction, owns this target mapping.
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
      (value, index) =>
        !allowed.has(value) || (index > 0 && fields[index - 1] >= value),
    )
  )
    throw new TypeError("invalid or non-canonical changed_fields");
  return fields;
}
function requireNoExtra(record: AuditRecord, allowed: readonly string[]): void {
  assertOnlyAuditAttributes(record, allowed);
}
function requireMove(
  record: AuditRecord,
  field: "folder_id" | "parent_id",
  changedFields: readonly string[],
): void {
  const hasMove = changedFields.includes(field);
  if (hasMove && record.previous_parent_id === record.new_parent_id)
    throw new TypeError(`${field} requires a distinct parent transition`);
  if (
    !hasMove &&
    (record.previous_parent_id !== undefined ||
      record.new_parent_id !== undefined)
  )
    throw new TypeError(`parent IDs require changed_fields ${field}`);
}

/** Returns true only for the reviewed File and File Folder catalog actions. */
export function validateFileAuditRecord(record: AuditRecord): boolean {
  if (
    record.action !== "file.created" &&
    record.action !== "file.updated" &&
    record.action !== "file.deleted" &&
    record.action !== "file_folder.created" &&
    record.action !== "file_folder.updated" &&
    record.action !== "file_folder.deleted"
  )
    return false;
  requireCatalogTarget(record.action, record.target_type, record.target_id);
  if (
    record.action === "file.created" ||
    record.action === "file.deleted" ||
    record.action === "file_folder.created" ||
    record.action === "file_folder.deleted"
  ) {
    requireNoExtra(record, []);
    return true;
  }
  if (record.action === "file_folder.updated") {
    const changedFields = requireFields(record, folderUpdateFields);
    requireMove(record, "parent_id", changedFields);
    requireNoExtra(record, [
      "changed_fields",
      "previous_parent_id",
      "new_parent_id",
    ]);
    return true;
  }
  const changedFields = requireFields(record, fileUpdateFields);
  requireMove(record, "folder_id", changedFields);
  requireNoExtra(record, [
    "changed_fields",
    "previous_parent_id",
    "new_parent_id",
  ]);
  return true;
}

export function buildFileCreatedAuditRecord(
  metadata: AuditMetadata,
  fileId: string,
): AuditRecord {
  return buildFileAuditRecord(metadata, "file.created", fileId);
}
export function buildFileDeletedAuditRecord(
  metadata: AuditMetadata,
  fileId: string,
): AuditRecord {
  return buildFileAuditRecord(metadata, "file.deleted", fileId);
}
export function buildFileRenamedAuditRecord(
  metadata: AuditMetadata,
  fileId: string,
): AuditRecord {
  return buildFileAuditRecord(metadata, "file.updated", fileId, {
    changed_fields: ["file_name"],
  });
}
export function buildFileMovedAuditRecord(
  metadata: AuditMetadata,
  fileId: string,
  previousParentId: string,
  newParentId: string,
): AuditRecord {
  return buildFileAuditRecord(metadata, "file.updated", fileId, {
    changed_fields: ["folder_id"],
    previous_parent_id: previousParentId === "" ? undefined : previousParentId,
    new_parent_id: newParentId === "" ? undefined : newParentId,
  });
}
export function buildFileFolderCreatedAuditRecord(
  metadata: AuditMetadata,
  folderId: string,
): AuditRecord {
  return buildFileAuditRecord(metadata, "file_folder.created", folderId);
}
export function buildFileFolderDeletedAuditRecord(
  metadata: AuditMetadata,
  folderId: string,
): AuditRecord {
  return buildFileAuditRecord(metadata, "file_folder.deleted", folderId);
}
export function buildFileFolderRenamedAuditRecord(
  metadata: AuditMetadata,
  folderId: string,
): AuditRecord {
  return buildFileAuditRecord(metadata, "file_folder.updated", folderId, {
    changed_fields: ["name"],
  });
}
export function buildFileFolderMovedAuditRecord(
  metadata: AuditMetadata,
  folderId: string,
  previousParentId: string,
  newParentId: string,
): AuditRecord {
  return buildFileAuditRecord(metadata, "file_folder.updated", folderId, {
    changed_fields: ["parent_id"],
    previous_parent_id: previousParentId === "" ? undefined : previousParentId,
    new_parent_id: newParentId === "" ? undefined : newParentId,
  });
}
