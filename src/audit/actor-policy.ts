import type { AuditRecord } from "../records.ts";

const SYSTEM_AUDIT_SERVICES: Readonly<Record<string, readonly string[]>> = {
  "account.deleted": ["geul-backend"],
  "file.deleted": ["geul-backend"],
  "post.updated": ["geul-collab"],
  "page.updated": ["geul-collab"],
  "work.updated": ["geul-collab"],
  "release.updated": ["geul-collab"],
  "campaign.updated": ["geul-backend"],
  "legal_policy.updated": ["geul-backend"],
  "member.updated": ["geul-backend"],
};

const ANONYMOUS_AUDIT_ACTIONS = new Set(["form_submission.created"]);

export function auditActionAllowsAnonymous(action: string): boolean {
  return ANONYMOUS_AUDIT_ACTIONS.has(action);
}

/** Explicit catalog policy for the few mutations a system actor may append. */
export function systemActorMayAppendAudit(record: AuditRecord): boolean {
  if (
    !SYSTEM_AUDIT_SERVICES[record.action]?.includes(
      String(record.actor_service),
    )
  )
    return false;
  if (record.action === "release.updated")
    return (
      record.actor_service === "geul-collab" &&
      record.changed_fields?.length === 1 &&
      ["track_audio"].includes(record.changed_fields[0] ?? "")
    );
  if (record.action === "campaign.updated") {
    return (
      record.actor_service === "geul-backend" &&
      record.changed_fields?.length === 1 &&
      record.changed_fields[0] === "status" &&
      (record.new_state === "sent" || record.new_state === "failed")
    );
  }
  if (
    record.action === "post.updated" ||
    record.action === "page.updated" ||
    record.action === "work.updated"
  ) {
    return (
      record.actor_service === "geul-collab" &&
      record.changed_fields?.length === 1 &&
      record.changed_fields[0] === "version"
    );
  }
  if (record.action !== "member.updated") return true;
  return (
    (record.changed_fields?.length === 1 &&
      record.changed_fields[0] === "role") ||
    (record.changed_fields?.length === 1 &&
      record.changed_fields[0] === "status" &&
      record.previous_state === "banned" &&
      record.new_state === "active")
  );
}
