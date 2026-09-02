import type { AuditRecord } from "../records.ts";
import {
  assertOnlyAuditAttributes,
  canonicalAuditValues,
} from "./attributes.ts";
import { buildAuditRecord } from "./builder.ts";
import { requireCatalogTarget } from "./catalog.ts";
import type { AuditMetadata } from "./types.ts";

function buildSettingsAuditRecord(
  metadata: AuditMetadata,
  action: AuditRecord["action"],
  targetType: string,
  targetId: string,
  attributes: Parameters<typeof buildAuditRecord>[2],
): AuditRecord {
  return buildAuditRecord(
    metadata,
    { action, target_type: targetType, target_id: targetId },
    attributes,
  );
}

const siteFields = new Set([
  "company_address",
  "company_name",
  "default_comments_enabled",
  "default_map_theme_id",
  "favicon_file_id",
  "google_analytics_id",
  "homepage_page_id",
  "legal_email",
  "loader_file_ids",
  "logo_dark_file_id",
  "logo_email_file_id",
  "logo_light_file_id",
  "menu_avatar_dropdown_id",
  "menu_footer_id",
  "menu_header_id",
  "menu_secondary_id",
  "meta_description",
  "og_image_config",
  "primary_color",
  "privacy_email",
  "privacy_og_background_file_id",
  "site_og_background_file_id",
  "site_title",
  "social_links",
  "support_email",
  "tax_id",
  "terms_og_background_file_id",
]);
const lifecycle = new Set([
  "draft:scheduled",
  "scheduled:draft",
  "draft:active",
  "scheduled:active",
  "active:archived",
]);

function requireFields(
  record: AuditRecord,
  allowed: ReadonlySet<string>,
): void {
  const fields = record.changed_fields;
  if (
    !fields?.length ||
    fields.some(
      (value, i) => !allowed.has(value) || (i > 0 && fields[i - 1] >= value),
    )
  )
    throw new TypeError("invalid or non-canonical changed_fields");
}
function requireNoExtra(record: AuditRecord, allowed: readonly string[]): void {
  assertOnlyAuditAttributes(record, allowed);
}
function requirePolicyIdentity(record: AuditRecord): void {
  const versionNumber = record.version_number;
  if (
    (record.policy_type !== "terms" && record.policy_type !== "privacy") ||
    !Number.isSafeInteger(versionNumber) ||
    versionNumber === undefined ||
    versionNumber < 1
  )
    throw new TypeError("legal policy requires policy_type and version_number");
}

/** Returns true only for this module's exact catalog actions. */
export function validateSettingsAuditRecord(record: AuditRecord): boolean {
  if (record.action === "site_settings.updated") {
    requireCatalogTarget(record.action, record.target_type, record.target_id);
    if (record.target_id !== "1")
      throw new TypeError("site_settings target must be 1");
    requireFields(record, siteFields);
    requireNoExtra(record, ["changed_fields"]);
    return true;
  }
  if (
    record.action === "map_theme.created" ||
    record.action === "map_theme.deleted"
  ) {
    requireCatalogTarget(record.action, record.target_type, record.target_id);
    requireNoExtra(record, []);
    return true;
  }
  if (record.action === "map_theme.updated") {
    requireCatalogTarget(record.action, record.target_type, record.target_id);
    if (record.actor_kind !== "member")
      throw new TypeError("map theme content mutation requires member actor");
    requireFields(record, new Set(["content"]));
    requireNoExtra(record, ["changed_fields"]);
    return true;
  }
  if (
    record.action === "legal_policy.created" ||
    record.action === "legal_policy.deleted"
  ) {
    requireCatalogTarget(record.action, record.target_type, record.target_id);
    requirePolicyIdentity(record);
    requireNoExtra(record, ["policy_type", "version_number"]);
    return true;
  }
  if (record.action !== "legal_policy.updated") return false;
  requireCatalogTarget(record.action, record.target_type, record.target_id);
  requirePolicyIdentity(record);
  if (
    record.changed_fields?.length === 1 &&
    record.changed_fields[0] === "share_links"
  ) {
    if (record.actor_kind !== "member")
      throw new TypeError("legal policy share links require member actor");
    if (
      !["created", "deleted"].includes(record.item_operation ?? "") ||
      !record.item_id
    )
      throw new TypeError(
        "legal policy share link requires item operation and id",
      );
    requireNoExtra(record, [
      "changed_fields",
      "policy_type",
      "version_number",
      "item_operation",
      "item_id",
    ]);
    return true;
  }
  if (
    record.actor_kind !== "member" &&
    !(record.actor_kind === "system" && record.actor_service === "geul-backend")
  )
    throw new TypeError(
      "legal policy lifecycle requires member or geul-backend system actor",
    );
  requireFields(record, new Set(["effective_at", "status"]));
  if (
    !record.previous_state ||
    !record.new_state ||
    !lifecycle.has(`${record.previous_state}:${record.new_state}`) ||
    (record.changed_fields?.includes("effective_at") && !record.effective_at)
  )
    throw new TypeError("invalid legal policy lifecycle");
  requireNoExtra(record, [
    "changed_fields",
    "policy_type",
    "version_number",
    "previous_state",
    "new_state",
    "effective_at",
  ]);
  return true;
}

export function buildSiteSettingsUpdatedAuditRecord(
  metadata: AuditMetadata,
  changedFields: readonly string[],
): AuditRecord {
  return buildSettingsAuditRecord(
    metadata,
    "site_settings.updated",
    "site_settings",
    "1",
    {
      changed_fields: canonicalAuditValues(changedFields),
    },
  );
}
export function buildMapThemeCreatedAuditRecord(
  metadata: AuditMetadata,
  themeId: string,
): AuditRecord {
  return buildSettingsAuditRecord(
    metadata,
    "map_theme.created",
    "map_theme",
    themeId,
    {},
  );
}
export function buildMapThemeDeletedAuditRecord(
  metadata: AuditMetadata,
  themeId: string,
): AuditRecord {
  return buildSettingsAuditRecord(
    metadata,
    "map_theme.deleted",
    "map_theme",
    themeId,
    {},
  );
}
export function buildMapThemeContentUpdatedAuditRecord(
  metadata: AuditMetadata,
  themeId: string,
): AuditRecord {
  return buildSettingsAuditRecord(
    metadata,
    "map_theme.updated",
    "map_theme",
    themeId,
    {
      changed_fields: ["content"],
    },
  );
}
export function buildLegalPolicyCreatedAuditRecord(
  metadata: AuditMetadata,
  id: string,
  policyType: "terms" | "privacy",
  versionNumber: number,
): AuditRecord {
  return buildSettingsAuditRecord(
    metadata,
    "legal_policy.created",
    "legal_policy",
    id,
    {
      policy_type: policyType,
      version_number: versionNumber,
    },
  );
}
export function buildLegalPolicyDeletedAuditRecord(
  metadata: AuditMetadata,
  id: string,
  policyType: "terms" | "privacy",
  versionNumber: number,
): AuditRecord {
  return buildSettingsAuditRecord(
    metadata,
    "legal_policy.deleted",
    "legal_policy",
    id,
    {
      policy_type: policyType,
      version_number: versionNumber,
    },
  );
}
export function buildLegalPolicyLifecycleAuditRecord(
  metadata: AuditMetadata,
  id: string,
  policyType: "terms" | "privacy",
  versionNumber: number,
  changedFields: readonly ("effective_at" | "status")[],
  previousState: AuditRecord["previous_state"],
  newState: AuditRecord["new_state"],
  effectiveAt?: string,
): AuditRecord {
  return buildSettingsAuditRecord(
    metadata,
    "legal_policy.updated",
    "legal_policy",
    id,
    {
      changed_fields: canonicalAuditValues(changedFields),
      policy_type: policyType,
      version_number: versionNumber,
      previous_state: previousState,
      new_state: newState,
      effective_at: effectiveAt,
    },
  );
}
export function buildLegalPolicyShareLinkAuditRecord(
  metadata: AuditMetadata,
  id: string,
  policyType: "terms" | "privacy",
  versionNumber: number,
  operation: "created" | "deleted",
  itemId: string,
): AuditRecord {
  return buildSettingsAuditRecord(
    metadata,
    "legal_policy.updated",
    "legal_policy",
    id,
    {
      changed_fields: ["share_links"],
      policy_type: policyType,
      version_number: versionNumber,
      item_operation: operation,
      item_id: itemId,
    },
  );
}
