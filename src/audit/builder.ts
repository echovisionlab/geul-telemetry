import { validateAuditRecord, type AuditRecord } from "../records.ts";
import type { AuditMetadata } from "./types.ts";

type AuditTarget = Pick<AuditRecord, "action" | "target_type" | "target_id">;

// A semantic builder may supply only reviewed domain attributes. Metadata and
// correlation are exclusively supplied by its caller, while action/target are
// exclusively supplied by the focused domain helper.
type AuditRecordAttributes = Omit<
  AuditRecord,
  keyof AuditMetadata | keyof AuditTarget
>;

/** Builds only a fully validated, catalogued audit wire record. */
// Internal assembly point used only by focused semantic builder modules.
// It is intentionally not re-exported from the package facade.
export function buildAuditRecord(
  metadata: AuditMetadata,
  target: AuditTarget,
  attributes: AuditRecordAttributes,
): AuditRecord {
  const record = { ...attributes, ...metadata, ...target } as AuditRecord;
  validateAuditRecord(record);
  return record;
}
