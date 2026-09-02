import type { AuditRecord } from "../records.ts";
import {
  assertOnlyAuditAttributes,
  canonicalAuditValues,
} from "./attributes.ts";
import { buildAuditRecord } from "./builder.ts";
import {
  AUDIT_ACTIONS,
  AUDIT_CATALOG,
  requireCatalogTarget,
} from "./catalog.ts";
import type { AuditMetadata } from "./types.ts";

type IntegrationTarget =
  | "mail_adapter"
  | "email_suppression"
  | "translation_settings"
  | "translation_provider";
type IntegrationAction = Extract<
  (typeof AUDIT_ACTIONS)[number],
  | "mail_adapter.created"
  | "mail_adapter.updated"
  | "mail_adapter.deleted"
  | "email_suppression.updated"
  | "translation_settings.updated"
  | "translation_provider.created"
  | "translation_provider.updated"
  | "translation_provider.deleted"
>;
export type AdapterConfigField =
  "active" | "config" | "name" | "priority" | "type";
export type TranslationSettingsField = "default_locale" | "protected_terms";

export const TRANSLATION_SETTINGS_FIELDS = [
  "default_locale",
  "protected_terms",
] as const satisfies readonly TranslationSettingsField[];

const targetTypes = new Set<IntegrationTarget>([
  "mail_adapter",
  "email_suppression",
  "translation_settings",
  "translation_provider",
]);
// Module ownership derives from the immutable catalog, not a duplicate map.
const integrationActions = new Set<IntegrationAction>(
  AUDIT_ACTIONS.filter((action) =>
    targetTypes.has(AUDIT_CATALOG[action] as IntegrationTarget),
  ) as IntegrationAction[],
);
const adapterConfigFields = new Set<AdapterConfigField>([
  "active",
  "config",
  "name",
  "priority",
  "type",
]);
const translationSettingsFields = new Set<TranslationSettingsField>(
  TRANSLATION_SETTINGS_FIELDS,
);
function buildIntegrationAuditRecord(
  metadata: AuditMetadata,
  action: IntegrationAction,
  targetId: string,
  attributes: Parameters<typeof buildAuditRecord>[2] = {},
): AuditRecord {
  return buildAuditRecord(
    metadata,
    { action, target_type: AUDIT_CATALOG[action], target_id: targetId },
    attributes,
  );
}
function requireFields(
  record: AuditRecord,
  allowed: ReadonlySet<string>,
): void {
  const fields = record.changed_fields;
  if (
    !fields?.length ||
    fields.some(
      (field, index) =>
        !allowed.has(field) || (index > 0 && fields[index - 1] >= field),
    )
  )
    throw new TypeError("invalid or non-canonical changed_fields");
}
function requireNoExtra(record: AuditRecord, allowed: readonly string[]): void {
  assertOnlyAuditAttributes(record, allowed);
}
/** Returns true only for the reviewed Platform Integrations catalog actions. */
export function validateIntegrationAuditRecord(record: AuditRecord): boolean {
  if (!integrationActions.has(record.action as IntegrationAction)) return false;
  requireCatalogTarget(record.action, record.target_type, record.target_id);
  switch (record.action) {
    case "mail_adapter.created":
    case "mail_adapter.deleted":
    case "translation_provider.created":
    case "translation_provider.deleted":
      requireNoExtra(record, []);
      return true;
    case "mail_adapter.updated":
    case "translation_provider.updated":
      requireFields(record, adapterConfigFields);
      requireNoExtra(record, ["changed_fields"]);
      return true;
    case "email_suppression.updated":
      if (
        record.changed_fields?.length !== 1 ||
        record.changed_fields[0] !== "status" ||
        record.previous_state !== "active" ||
        record.new_state !== "released"
      )
        throw new TypeError(
          "email suppression requires active to released status",
        );
      requireNoExtra(record, ["changed_fields", "previous_state", "new_state"]);
      return true;
    case "translation_settings.updated":
      if (record.target_id !== "1")
        throw new TypeError("translation settings target must be 1");
      requireFields(record, translationSettingsFields);
      requireNoExtra(record, ["changed_fields"]);
      return true;
  }
  /* v8 ignore next -- IntegrationAction is exhaustive above. */
  throw new TypeError(`unsupported integration audit action ${record.action}`);
}

export function buildMailAdapterCreatedAuditRecord(
  metadata: AuditMetadata,
  id: string,
): AuditRecord {
  return buildIntegrationAuditRecord(metadata, "mail_adapter.created", id);
}
export function buildMailAdapterConfigUpdatedAuditRecord(
  metadata: AuditMetadata,
  id: string,
  fields: readonly AdapterConfigField[],
): AuditRecord {
  return buildIntegrationAuditRecord(metadata, "mail_adapter.updated", id, {
    changed_fields: canonicalAuditValues(fields),
  });
}
export function buildMailAdapterDeletedAuditRecord(
  metadata: AuditMetadata,
  id: string,
): AuditRecord {
  return buildIntegrationAuditRecord(metadata, "mail_adapter.deleted", id);
}
export function buildEmailSuppressionReleasedAuditRecord(
  metadata: AuditMetadata,
  id: string,
): AuditRecord {
  return buildIntegrationAuditRecord(
    metadata,
    "email_suppression.updated",
    id,
    {
      changed_fields: ["status"],
      previous_state: "active",
      new_state: "released",
    },
  );
}
export function buildTranslationSettingsUpdatedAuditRecord(
  metadata: AuditMetadata,
  fields: readonly TranslationSettingsField[],
): AuditRecord {
  return buildIntegrationAuditRecord(
    metadata,
    "translation_settings.updated",
    "1",
    { changed_fields: canonicalAuditValues(fields) },
  );
}
export function buildTranslationProviderCreatedAuditRecord(
  metadata: AuditMetadata,
  id: string,
): AuditRecord {
  return buildIntegrationAuditRecord(
    metadata,
    "translation_provider.created",
    id,
  );
}
export function buildTranslationProviderConfigUpdatedAuditRecord(
  metadata: AuditMetadata,
  id: string,
  fields: readonly AdapterConfigField[],
): AuditRecord {
  return buildIntegrationAuditRecord(
    metadata,
    "translation_provider.updated",
    id,
    { changed_fields: canonicalAuditValues(fields) },
  );
}
export function buildTranslationProviderDeletedAuditRecord(
  metadata: AuditMetadata,
  id: string,
): AuditRecord {
  return buildIntegrationAuditRecord(
    metadata,
    "translation_provider.deleted",
    id,
  );
}
