import { isRequestId } from "../request-context.ts";
import type { AuditRecord, AuditState } from "../records.ts";
import { assertOnlyAuditAttributes } from "./attributes.ts";
import { requireCatalogTarget } from "./catalog.ts";

function isIdentifier(value: string): boolean {
  return value.length > 0 && value.length <= 255 && value.trim() === value;
}

function isEmail(value: string): boolean {
  return (
    value.length <= 320 &&
    value.trim() === value &&
    !/[\s]/.test(value) &&
    value.split("@").length === 2
  );
}

function isBoundedCode(value: string): boolean {
  return /^[a-z0-9_]{1,64}$/.test(value);
}

function requireSortedUnique(
  name: string,
  values: readonly string[],
  valid: (value: string) => boolean,
): void {
  for (let index = 0; index < values.length; index += 1) {
    const value = values[index];
    if (!valid(value)) throw new TypeError(`${name} contains an invalid value`);
    if (index > 0 && values[index - 1] >= value)
      throw new TypeError(`${name} must be sorted and unique`);
  }
}

function requireOnly(record: AuditRecord, ...allowed: readonly string[]): void {
  assertOnlyAuditAttributes(record, allowed);
}

function requireTransition(
  record: AuditRecord,
  field: string,
  allowed: readonly (readonly [AuditState, AuditState])[],
): void {
  if (
    record.previous_state === undefined ||
    record.new_state === undefined ||
    record.previous_state === record.new_state ||
    !allowed.some(
      ([previous, next]) =>
        previous === record.previous_state && next === record.new_state,
    )
  )
    throw new TypeError(
      `audit action ${record.action} field ${field} rejects transition ${String(record.previous_state)} to ${String(record.new_state)}`,
    );
  requireOnly(record, "changed_fields", "previous_state", "new_state");
}

function validateMember(record: AuditRecord): void {
  requireCatalogTarget(record.action, record.target_type, record.target_id);
  if (
    record.changed_fields?.length === 2 &&
    record.changed_fields[0] === "nickname" &&
    record.changed_fields[1] === "onboarded"
  ) {
    if (
      record.actor_kind !== "member" ||
      record.actor_member_id !== record.target_id
    )
      throw new TypeError(
        "member.updated onboarding requires the committed member as actor and target",
      );
    if (
      record.nickname === undefined ||
      record.nickname === "" ||
      record.nickname.trim() !== record.nickname ||
      [...record.nickname].length > 100
    )
      throw new TypeError(
        "member.updated onboarding requires a trimmed nickname between 1 and 100 characters",
      );
    requireOnly(record, "changed_fields", "nickname");
    return;
  }
  const fields = record.changed_fields;
  if (fields === undefined || fields.length === 0)
    throw new TypeError(
      "member.updated requires a catalog changed_fields variant",
    );
  requireSortedUnique("changed_fields", fields, isBoundedCode);
  const field = fields[0];
  if (field === "role") {
    if (
      !isBoundedCode(record.previous_role ?? "") ||
      !isBoundedCode(record.new_role ?? "") ||
      record.previous_role === record.new_role
    )
      throw new TypeError(
        "member.updated role requires distinct previous_role and new_role",
      );
    requireOnly(record, "changed_fields", "previous_role", "new_role");
    return;
  }
  if (field === "avatar") {
    if (
      (record.collection_operation !== "added" &&
        record.collection_operation !== "removed") ||
      !isIdentifier(record.asset_id ?? "")
    )
      throw new TypeError(
        "member avatar requires collection operation and asset_id",
      );
    requireOnly(record, "changed_fields", "collection_operation", "asset_id");
    return;
  }
  if (field === "tags") {
    if (record.tag_ids === undefined)
      throw new TypeError("member tags requires tag_ids");
    requireSortedUnique("tag_ids", record.tag_ids, isIdentifier);
    requireOnly(record, "changed_fields", "tag_ids");
    return;
  }
  const profile = new Set(["bio", "nickname", "social_links", "website"]);
  if (fields.every((candidate) => profile.has(candidate))) {
    if (fields.includes("nickname")) {
      if (
        record.nickname === undefined ||
        record.nickname === "" ||
        record.nickname.trim() !== record.nickname
      )
        throw new TypeError("member nickname requires a trimmed value");
    } else if (record.nickname !== undefined)
      throw new TypeError("member nickname requires changed_fields nickname");
    requireOnly(record, "changed_fields", "nickname");
    return;
  }
  const preference = new Set(["cookie_consent", "preferred_locale"]);
  if (fields.every((candidate) => preference.has(candidate))) {
    if (fields.includes("preferred_locale")) {
      if (!isBoundedCode(record.preferred_locale ?? ""))
        throw new TypeError(
          "member preferred_locale requires a bounded locale",
        );
    } else if (record.preferred_locale !== undefined)
      throw new TypeError(
        "member preferred_locale requires changed_fields preferred_locale",
      );
    if (fields.includes("cookie_consent")) {
      if (!isIdentifier(record.consent_id ?? ""))
        throw new TypeError("member cookie_consent requires consent_id");
    } else if (record.consent_id !== undefined)
      throw new TypeError(
        "member consent_id requires changed_fields cookie_consent",
      );
    requireOnly(record, "changed_fields", "preferred_locale", "consent_id");
    return;
  }
  if (fields.length !== 1 || field !== "status")
    throw new TypeError(
      "member.updated requires a catalog changed_fields variant",
    );
  requireTransition(record, "status", [
    ["active", "banned"],
    ["banned", "active"],
  ]);
}

function validateAccount(record: AuditRecord): void {
  requireCatalogTarget(record.action, record.target_type, record.target_id);
  if (record.changed_fields?.length !== 1)
    throw new TypeError("account.updated requires one catalog changed_field");
  const field = record.changed_fields[0];
  const requireCollection = (): void => {
    if (
      record.collection_operation !== "added" &&
      record.collection_operation !== "removed"
    )
      throw new TypeError(
        `audit action ${record.action} requires a catalog collection_operation`,
      );
  };
  switch (field) {
    case "canonical_email":
      if (
        !isEmail(record.previous_email ?? "") ||
        !isEmail(record.new_email ?? "") ||
        record.previous_email === record.new_email
      )
        throw new TypeError(
          "account.updated canonical_email requires distinct previous_email and new_email",
        );
      requireOnly(record, "changed_fields", "previous_email", "new_email");
      return;
    case "login_emails":
      if (!isEmail(record.email ?? ""))
        throw new TypeError("account.updated login_emails requires email");
      requireCollection();
      requireOnly(record, "changed_fields", "collection_operation", "email");
      return;
    case "social_logins":
      if (
        !isBoundedCode(record.provider ?? "") ||
        !isIdentifier(record.provider_subject ?? "")
      )
        throw new TypeError(
          "account.updated social_logins requires provider and provider_subject",
        );
      requireCollection();
      requireOnly(
        record,
        "changed_fields",
        "collection_operation",
        "provider",
        "provider_subject",
      );
      return;
    case "passkeys":
      if (record.passkey_ids === undefined || record.passkey_ids.length === 0)
        throw new TypeError("account.updated passkeys requires passkey_ids");
      requireSortedUnique("passkey_ids", record.passkey_ids, isIdentifier);
      requireCollection();
      requireOnly(
        record,
        "changed_fields",
        "collection_operation",
        "passkey_ids",
      );
      return;
    case "sessions":
      if (record.collection_operation !== "removed")
        throw new TypeError(
          "account.updated sessions requires collection_operation removed",
        );
      if (
        record.session_scope !== "current" &&
        record.session_scope !== "one" &&
        record.session_scope !== "others"
      )
        throw new TypeError(
          "account.updated sessions requires a catalog session_scope",
        );
      if (record.session_ids === undefined)
        throw new TypeError("account.updated sessions requires session_ids");
      requireSortedUnique("session_ids", record.session_ids, isRequestId);
      if (
        ((record.session_scope === "current" ||
          record.session_scope === "one") &&
          record.session_ids.length !== 1) ||
        (record.session_scope === "others" && record.session_ids.length === 0)
      )
        throw new TypeError(
          `account.updated sessions scope ${record.session_scope} has invalid session_ids`,
        );
      requireOnly(
        record,
        "changed_fields",
        "collection_operation",
        "session_scope",
        "session_ids",
      );
      return;
    case "newsletter_subscription":
      requireTransition(record, "newsletter_subscription", [
        ["subscribed", "unsubscribed"],
        ["unsubscribed", "subscribed"],
      ]);
      return;
    case "deletion_state":
      requireTransition(record, "deletion_state", [
        ["none", "confirmation_pending"],
        ["cancelled", "confirmation_pending"],
        ["recovered", "confirmation_pending"],
        ["confirmation_pending", "scheduled"],
        ["none", "scheduled"],
        ["cancelled", "scheduled"],
        ["recovered", "scheduled"],
        ["scheduled", "cancelled"],
        ["recovery_confirmation_pending", "recovered"],
      ]);
      return;
    default:
      throw new TypeError(`account.updated rejects changed_field ${field}`);
  }
}

export function validateMemberAccountAuditRecord(record: AuditRecord): boolean {
  switch (record.action) {
    case "member.updated":
      validateMember(record);
      return true;
    case "member_tag.created":
    case "member_tag.deleted":
      requireCatalogTarget(record.action, record.target_type, record.target_id);
      if (
        record.tag_name === undefined ||
        record.tag_name === "" ||
        record.tag_name.trim() !== record.tag_name
      )
        throw new TypeError("member tag requires trimmed tag_name");
      requireOnly(record, "tag_name");
      return true;
    case "account.updated":
      validateAccount(record);
      return true;
    case "account.deleted":
      requireCatalogTarget(record.action, record.target_type, record.target_id);
      requireOnly(record);
      return true;
    default:
      return false;
  }
}
