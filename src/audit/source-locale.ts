import type { AuditAction, AuditRecord } from "../records.ts";
import { assertOnlyAuditAttributes } from "./attributes.ts";
import { buildAuditRecord } from "./builder.ts";
import { AUDIT_CATALOG, requireCatalogTarget } from "./catalog.ts";
import { isCanonicalAuditLocale } from "./locale-code.ts";
import type { AuditMetadata } from "./types.ts";

type SourceLocaleAction = Extract<
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

const sourceLocaleActions = new Set<SourceLocaleAction>([
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

/** Recognizes exactly the reviewed source-locale variant, never a generic update. */
export function validateSourceLocaleAuditRecord(record: AuditRecord): boolean {
  if (
    record.changed_fields?.length !== 1 ||
    record.changed_fields[0] !== "source_locale"
  )
    return false;
  if (!sourceLocaleActions.has(record.action as SourceLocaleAction))
    throw new TypeError(
      `audit action ${record.action} does not support source_locale`,
    );
  requireCatalogTarget(record.action, record.target_type, record.target_id);
  if (record.actor_kind !== "member")
    throw new TypeError("source_locale requires a member actor");
  if (
    !isCanonicalAuditLocale(record.previous_locale) ||
    !isCanonicalAuditLocale(record.new_locale) ||
    record.previous_locale === record.new_locale
  )
    throw new TypeError(
      "source_locale requires distinct bounded previous_locale and new_locale",
    );
  const allowed = ["changed_fields", "previous_locale", "new_locale"];
  if (record.action === "legal_policy.updated") {
    if (
      (record.policy_type !== "terms" && record.policy_type !== "privacy") ||
      !Number.isSafeInteger(record.version_number) ||
      record.version_number === undefined ||
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

function buildSourceLocaleAuditRecord(
  metadata: AuditMetadata,
  action: SourceLocaleAction,
  id: string,
  previousLocale: string,
  newLocale: string,
): AuditRecord {
  return buildAuditRecord(
    metadata,
    { action, target_type: AUDIT_CATALOG[action], target_id: id },
    {
      changed_fields: ["source_locale"],
      previous_locale: previousLocale,
      new_locale: newLocale,
    },
  );
}

export function buildPostSourceLocaleAuditRecord(
  m: AuditMetadata,
  id: string,
  previousLocale: string,
  newLocale: string,
): AuditRecord {
  return buildSourceLocaleAuditRecord(
    m,
    "post.updated",
    id,
    previousLocale,
    newLocale,
  );
}
export function buildPageSourceLocaleAuditRecord(
  m: AuditMetadata,
  id: string,
  previousLocale: string,
  newLocale: string,
): AuditRecord {
  return buildSourceLocaleAuditRecord(
    m,
    "page.updated",
    id,
    previousLocale,
    newLocale,
  );
}
export function buildWorkSourceLocaleAuditRecord(
  m: AuditMetadata,
  id: string,
  previousLocale: string,
  newLocale: string,
): AuditRecord {
  return buildSourceLocaleAuditRecord(
    m,
    "work.updated",
    id,
    previousLocale,
    newLocale,
  );
}
export function buildPostSeriesSourceLocaleAuditRecord(
  m: AuditMetadata,
  id: string,
  previousLocale: string,
  newLocale: string,
): AuditRecord {
  return buildSourceLocaleAuditRecord(
    m,
    "post_series.updated",
    id,
    previousLocale,
    newLocale,
  );
}
export function buildProgramEventSourceLocaleAuditRecord(
  m: AuditMetadata,
  id: string,
  previousLocale: string,
  newLocale: string,
): AuditRecord {
  return buildSourceLocaleAuditRecord(
    m,
    "program_event.updated",
    id,
    previousLocale,
    newLocale,
  );
}
export function buildReleaseSourceLocaleAuditRecord(
  m: AuditMetadata,
  id: string,
  previousLocale: string,
  newLocale: string,
): AuditRecord {
  return buildSourceLocaleAuditRecord(
    m,
    "release.updated",
    id,
    previousLocale,
    newLocale,
  );
}
export function buildArtistSourceLocaleAuditRecord(
  m: AuditMetadata,
  id: string,
  previousLocale: string,
  newLocale: string,
): AuditRecord {
  return buildSourceLocaleAuditRecord(
    m,
    "artist.updated",
    id,
    previousLocale,
    newLocale,
  );
}
export function buildLabelSourceLocaleAuditRecord(
  m: AuditMetadata,
  id: string,
  previousLocale: string,
  newLocale: string,
): AuditRecord {
  return buildSourceLocaleAuditRecord(
    m,
    "label.updated",
    id,
    previousLocale,
    newLocale,
  );
}
export function buildMenuSourceLocaleAuditRecord(
  m: AuditMetadata,
  id: string,
  previousLocale: string,
  newLocale: string,
): AuditRecord {
  return buildSourceLocaleAuditRecord(
    m,
    "menu.updated",
    id,
    previousLocale,
    newLocale,
  );
}
export function buildCampaignSourceLocaleAuditRecord(
  m: AuditMetadata,
  id: string,
  previousLocale: string,
  newLocale: string,
): AuditRecord {
  return buildSourceLocaleAuditRecord(
    m,
    "campaign.updated",
    id,
    previousLocale,
    newLocale,
  );
}
export function buildFormSourceLocaleAuditRecord(
  m: AuditMetadata,
  id: string,
  previousLocale: string,
  newLocale: string,
): AuditRecord {
  return buildSourceLocaleAuditRecord(
    m,
    "form.updated",
    id,
    previousLocale,
    newLocale,
  );
}
export function buildEmailTemplateSourceLocaleAuditRecord(
  m: AuditMetadata,
  id: string,
  previousLocale: string,
  newLocale: string,
): AuditRecord {
  return buildSourceLocaleAuditRecord(
    m,
    "email_template.updated",
    id,
    previousLocale,
    newLocale,
  );
}
export function buildEmailLayoutSourceLocaleAuditRecord(
  m: AuditMetadata,
  id: string,
  previousLocale: string,
  newLocale: string,
): AuditRecord {
  return buildSourceLocaleAuditRecord(
    m,
    "email_layout.updated",
    id,
    previousLocale,
    newLocale,
  );
}
function buildLegalPolicySourceLocaleAuditRecord(
  metadata: AuditMetadata,
  id: string,
  policyType: "terms" | "privacy",
  versionNumber: number,
  previousLocale: string,
  newLocale: string,
): AuditRecord {
  return buildAuditRecord(
    metadata,
    {
      action: "legal_policy.updated",
      target_type: AUDIT_CATALOG["legal_policy.updated"],
      target_id: id,
    },
    {
      changed_fields: ["source_locale"],
      policy_type: policyType,
      version_number: versionNumber,
      previous_locale: previousLocale,
      new_locale: newLocale,
    },
  );
}
export function buildPrivacySourceLocaleAuditRecord(
  m: AuditMetadata,
  id: string,
  versionNumber: number,
  previousLocale: string,
  newLocale: string,
): AuditRecord {
  return buildLegalPolicySourceLocaleAuditRecord(
    m,
    id,
    "privacy",
    versionNumber,
    previousLocale,
    newLocale,
  );
}
export function buildTermsSourceLocaleAuditRecord(
  m: AuditMetadata,
  id: string,
  versionNumber: number,
  previousLocale: string,
  newLocale: string,
): AuditRecord {
  return buildLegalPolicySourceLocaleAuditRecord(
    m,
    id,
    "terms",
    versionNumber,
    previousLocale,
    newLocale,
  );
}
