import type { AuditRecord } from "../records.ts";

/**
 * Canonical wire representation for set-shaped audit attributes.  In
 * particular, an empty input stays an explicit empty array rather than being
 * omitted; validators decide whether a given catalog variant permits it.
 */
export function canonicalAuditValues(
  values: readonly string[],
): readonly string[] {
  return [...new Set(values)].sort();
}

/** The complete, reviewed Domain Audit JSONB attribute namespace. */
export const AUDIT_RECORD_ATTRIBUTE_NAMES = [
  "nickname",
  "asset_id",
  "consent_id",
  "changed_fields",
  "collection_operation",
  "asset_slot",
  "previous_state",
  "new_state",
  "scheduled_at",
  "scheduled_time_zone",
  "version_id",
  "version_number",
  "contributor_member_ids",
  "event_name",
  "file_id",
  "file_ids",
  "item_id",
  "item_ids",
  "item_operation",
  "item_scope",
  "parent_id",
  "policy_type",
  "effective_at",
  "post_ids",
  "preferred_locale",
  "locale",
  "previous_locale",
  "new_locale",
  "previous_item_id",
  "previous_item_ids",
  "previous_parent_id",
  "new_parent_id",
  "previous_relationship",
  "new_relationship",
  "previous_series_id",
  "new_series_id",
  "subject_member_id",
  "subject_post_id",
  "tag_ids",
  "tag_name",
  "previous_role",
  "new_role",
  "email",
  "previous_email",
  "new_email",
  "provider",
  "provider_subject",
  "passkey_ids",
  "session_scope",
  "session_ids",
] as const;

export type AuditRecordAttributeName =
  (typeof AUDIT_RECORD_ATTRIBUTE_NAMES)[number];

const auditRecordKeys = new Set<string>([
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
  ...AUDIT_RECORD_ATTRIBUTE_NAMES,
]);

/** Rejects input that cannot be represented by the public Audit wire type. */
export function assertKnownAuditRecordKeys(record: AuditRecord): void {
  for (const key of Object.keys(record)) {
    if (!auditRecordKeys.has(key)) {
      throw new TypeError(`audit record has unknown attribute ${key}`);
    }
  }
}

/** Enforces an action variant's exact permitted attribute shape. */
export function assertOnlyAuditAttributes(
  record: AuditRecord,
  allowed: readonly string[],
): void {
  for (const name of AUDIT_RECORD_ATTRIBUTE_NAMES) {
    if (record[name] !== undefined && !allowed.includes(name)) {
      throw new TypeError(
        `audit action ${record.action} does not allow ${name}`,
      );
    }
  }
}
