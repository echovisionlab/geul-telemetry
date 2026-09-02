import type {
  AuditAction,
  AuditItemOperation,
  AuditRecord,
} from "../records.ts";
import { assertOnlyAuditAttributes } from "./attributes.ts";
import { buildAuditRecord } from "./builder.ts";
import { AUDIT_CATALOG, requireCatalogTarget } from "./catalog.ts";
import { isCanonicalAuditLocale } from "./locale-code.ts";
import type { AuditMetadata } from "./types.ts";

type LocaleContentAction = Extract<
  AuditAction,
  | "post.updated"
  | "page.updated"
  | "work.updated"
  | "post_series.updated"
  | "program_event.updated"
  | "release.updated"
  | "artist.updated"
  | "label.updated"
  | "menu.updated"
  | "campaign.updated"
  | "form.updated"
  | "email_template.updated"
  | "email_layout.updated"
  | "legal_policy.updated"
>;

const localeContentActions = new Set<LocaleContentAction>([
  "post.updated",
  "page.updated",
  "work.updated",
  "post_series.updated",
  "program_event.updated",
  "release.updated",
  "artist.updated",
  "label.updated",
  "menu.updated",
  "campaign.updated",
  "form.updated",
  "email_template.updated",
  "email_layout.updated",
  "legal_policy.updated",
]);

/** Recognizes exactly one locale-document create, update, or delete. */
export function validateLocaleContentAuditRecord(record: AuditRecord): boolean {
  if (
    record.changed_fields?.length !== 1 ||
    record.changed_fields[0] !== "locale_content"
  )
    return false;
  if (!localeContentActions.has(record.action as LocaleContentAction))
    throw new TypeError(
      `audit action ${record.action} does not support locale_content`,
    );
  requireCatalogTarget(record.action, record.target_type, record.target_id);
  if (record.actor_kind !== "member")
    throw new TypeError("locale_content requires a member actor");
  if (!isCanonicalAuditLocale(record.locale))
    throw new TypeError("locale_content requires a bounded locale");
  if (
    record.item_operation !== "created" &&
    record.item_operation !== "updated" &&
    record.item_operation !== "deleted"
  )
    throw new TypeError(
      "locale_content requires created, updated, or deleted item operation",
    );
  const allowed = ["changed_fields", "locale", "item_operation"];
  if (record.action === "legal_policy.updated") {
    if (
      (record.policy_type !== "terms" && record.policy_type !== "privacy") ||
      record.version_number === undefined ||
      !Number.isSafeInteger(record.version_number) ||
      record.version_number < 1
    )
      throw new TypeError(
        "legal policy requires policy_type and version_number",
      );
    allowed.push("policy_type", "version_number");
  }
  assertOnlyAuditAttributes(record, allowed);
  return true;
}

function buildLocaleContentAuditRecord(
  metadata: AuditMetadata,
  action: Exclude<LocaleContentAction, "legal_policy.updated">,
  id: string,
  locale: string,
  operation: AuditItemOperation,
): AuditRecord {
  return buildAuditRecord(
    metadata,
    { action, target_type: AUDIT_CATALOG[action], target_id: id },
    {
      changed_fields: ["locale_content"],
      locale,
      item_operation: operation,
    },
  );
}

export function buildPostLocaleContentAuditRecord(
  m: AuditMetadata,
  id: string,
  locale: string,
  operation: AuditItemOperation,
): AuditRecord {
  return buildLocaleContentAuditRecord(
    m,
    "post.updated",
    id,
    locale,
    operation,
  );
}

export function buildPageLocaleContentAuditRecord(
  m: AuditMetadata,
  id: string,
  locale: string,
  operation: AuditItemOperation,
): AuditRecord {
  return buildLocaleContentAuditRecord(
    m,
    "page.updated",
    id,
    locale,
    operation,
  );
}

export function buildWorkLocaleContentAuditRecord(
  m: AuditMetadata,
  id: string,
  locale: string,
  operation: AuditItemOperation,
): AuditRecord {
  return buildLocaleContentAuditRecord(
    m,
    "work.updated",
    id,
    locale,
    operation,
  );
}

export function buildPostSeriesLocaleContentAuditRecord(
  m: AuditMetadata,
  id: string,
  locale: string,
  operation: AuditItemOperation,
): AuditRecord {
  return buildLocaleContentAuditRecord(
    m,
    "post_series.updated",
    id,
    locale,
    operation,
  );
}

export function buildProgramEventLocaleContentAuditRecord(
  m: AuditMetadata,
  id: string,
  locale: string,
  operation: AuditItemOperation,
): AuditRecord {
  return buildLocaleContentAuditRecord(
    m,
    "program_event.updated",
    id,
    locale,
    operation,
  );
}

export function buildReleaseLocaleContentAuditRecord(
  m: AuditMetadata,
  id: string,
  locale: string,
  operation: AuditItemOperation,
): AuditRecord {
  return buildLocaleContentAuditRecord(
    m,
    "release.updated",
    id,
    locale,
    operation,
  );
}

export function buildArtistLocaleContentAuditRecord(
  m: AuditMetadata,
  id: string,
  locale: string,
  operation: AuditItemOperation,
): AuditRecord {
  return buildLocaleContentAuditRecord(
    m,
    "artist.updated",
    id,
    locale,
    operation,
  );
}

export function buildLabelLocaleContentAuditRecord(
  m: AuditMetadata,
  id: string,
  locale: string,
  operation: AuditItemOperation,
): AuditRecord {
  return buildLocaleContentAuditRecord(
    m,
    "label.updated",
    id,
    locale,
    operation,
  );
}

export function buildMenuLocaleContentAuditRecord(
  m: AuditMetadata,
  id: string,
  locale: string,
  operation: AuditItemOperation,
): AuditRecord {
  return buildLocaleContentAuditRecord(
    m,
    "menu.updated",
    id,
    locale,
    operation,
  );
}

export function buildCampaignLocaleContentAuditRecord(
  m: AuditMetadata,
  id: string,
  locale: string,
  operation: AuditItemOperation,
): AuditRecord {
  return buildLocaleContentAuditRecord(
    m,
    "campaign.updated",
    id,
    locale,
    operation,
  );
}

export function buildFormLocaleContentAuditRecord(
  m: AuditMetadata,
  id: string,
  locale: string,
  operation: AuditItemOperation,
): AuditRecord {
  return buildLocaleContentAuditRecord(
    m,
    "form.updated",
    id,
    locale,
    operation,
  );
}

export function buildEmailTemplateLocaleContentAuditRecord(
  m: AuditMetadata,
  id: string,
  locale: string,
  operation: AuditItemOperation,
): AuditRecord {
  return buildLocaleContentAuditRecord(
    m,
    "email_template.updated",
    id,
    locale,
    operation,
  );
}

export function buildEmailLayoutLocaleContentAuditRecord(
  m: AuditMetadata,
  id: string,
  locale: string,
  operation: AuditItemOperation,
): AuditRecord {
  return buildLocaleContentAuditRecord(
    m,
    "email_layout.updated",
    id,
    locale,
    operation,
  );
}

function buildLegalPolicyLocaleContentAuditRecord(
  metadata: AuditMetadata,
  id: string,
  policyType: "terms" | "privacy",
  versionNumber: number,
  locale: string,
  operation: AuditItemOperation,
): AuditRecord {
  return buildAuditRecord(
    metadata,
    {
      action: "legal_policy.updated",
      target_type: AUDIT_CATALOG["legal_policy.updated"],
      target_id: id,
    },
    {
      policy_type: policyType,
      version_number: versionNumber,
      changed_fields: ["locale_content"],
      locale,
      item_operation: operation,
    },
  );
}

export function buildPrivacyLocaleContentAuditRecord(
  m: AuditMetadata,
  id: string,
  versionNumber: number,
  locale: string,
  operation: AuditItemOperation,
): AuditRecord {
  return buildLegalPolicyLocaleContentAuditRecord(
    m,
    id,
    "privacy",
    versionNumber,
    locale,
    operation,
  );
}

export function buildTermsLocaleContentAuditRecord(
  m: AuditMetadata,
  id: string,
  versionNumber: number,
  locale: string,
  operation: AuditItemOperation,
): AuditRecord {
  return buildLegalPolicyLocaleContentAuditRecord(
    m,
    id,
    "terms",
    versionNumber,
    locale,
    operation,
  );
}
