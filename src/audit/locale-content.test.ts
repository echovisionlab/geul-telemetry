import { readFileSync } from "node:fs";
import { describe, expect, it } from "vitest";
import type {
  AuditAction,
  AuditItemOperation,
  AuditRecord,
} from "../records.ts";
import { validateAuditRecord } from "../records.ts";
import {
  buildArtistLocaleContentAuditRecord,
  buildCampaignLocaleContentAuditRecord,
  buildEmailLayoutLocaleContentAuditRecord,
  buildEmailTemplateLocaleContentAuditRecord,
  buildFormLocaleContentAuditRecord,
  buildLabelLocaleContentAuditRecord,
  buildMenuLocaleContentAuditRecord,
  buildPageLocaleContentAuditRecord,
  buildPostSeriesLocaleContentAuditRecord,
  buildPostLocaleContentAuditRecord,
  buildPrivacyLocaleContentAuditRecord,
  buildProgramEventLocaleContentAuditRecord,
  buildReleaseLocaleContentAuditRecord,
  buildTermsLocaleContentAuditRecord,
  buildWorkLocaleContentAuditRecord,
  validateLocaleContentAuditRecord,
} from "./locale-content.ts";
import type { AuditMetadata } from "./types.ts";

interface LocaleContentFixture {
  readonly case: string;
  readonly action: AuditAction;
  readonly target_type: string;
  readonly target_id: string;
  readonly locale: string;
  readonly item_operation: AuditItemOperation;
  readonly policy_type?: "terms" | "privacy";
  readonly version_number?: number;
}

const metadata: AuditMetadata = {
  audit_id: "00000000-0000-4000-8000-000000000001",
  occurred_at: "2026-08-09T03:04:05Z",
  actor_kind: "member",
  actor_member_id: "member-1",
};

describe("locale content audit", () => {
  it("keeps source and translated locales for every enabled domain on one exact wire variant", () => {
    for (const fixture of loadFixtures()) {
      const record = buildFixtureRecord(fixture);
      validateAuditRecord(record);
      expect(record, fixture.case).toMatchObject({
        action: fixture.action,
        target_type: fixture.target_type,
        target_id: fixture.target_id,
        actor_kind: "member",
        changed_fields: ["locale_content"],
        locale: fixture.locale,
        item_operation: fixture.item_operation,
        ...(fixture.policy_type === undefined
          ? {}
          : {
              policy_type: fixture.policy_type,
              version_number: fixture.version_number,
            }),
      });
    }
  });

  it("rejects unreviewed actions, actors, locales, operations, and attributes", () => {
    const valid = buildPostLocaleContentAuditRecord(
      metadata,
      "post-1",
      "ko",
      "updated",
    );
    const invalid: readonly AuditRecord[] = [
      { ...valid, action: "category.updated", target_type: "category" },
      {
        ...valid,
        actor_kind: "system",
        actor_member_id: undefined,
        actor_service: "geul-backend",
      },
      { ...valid, locale: "ko_KR" },
      { ...valid, item_operation: "replace" as AuditItemOperation },
      { ...valid, asset_id: "unexpected" },
      {
        ...valid,
        action: "legal_policy.updated",
        target_type: "legal_policy",
      },
    ];
    for (const record of invalid) {
      expect(() => validateLocaleContentAuditRecord(record)).toThrow();
    }
  });
});

function loadFixtures(): readonly LocaleContentFixture[] {
  const parsed: unknown = JSON.parse(
    readFileSync(
      new URL("../../fixtures/locale-content-audit.json", import.meta.url),
      "utf8",
    ),
  );
  if (!Array.isArray(parsed)) throw new TypeError("fixture must be an array");
  return parsed.map(parseFixture);
}

function parseFixture(value: unknown): LocaleContentFixture {
  if (value === null || typeof value !== "object")
    throw new TypeError("fixture entry must be an object");
  const entry = value as Record<string, unknown>;
  const caseName = requiredString(entry.case, "case");
  const action = requiredString(entry.action, "action") as AuditAction;
  const targetType = requiredString(entry.target_type, "target_type");
  const targetID = requiredString(entry.target_id, "target_id");
  const locale = requiredString(entry.locale, "locale");
  const itemOperation = parseItemOperation(entry.item_operation);
  const policyType = entry.policy_type;
  if (
    policyType !== undefined &&
    policyType !== "terms" &&
    policyType !== "privacy"
  )
    throw new TypeError("policy_type is invalid");
  const versionNumber = entry.version_number;
  if (
    versionNumber !== undefined &&
    (!Number.isSafeInteger(versionNumber) || Number(versionNumber) < 1)
  )
    throw new TypeError("version_number is invalid");
  return {
    case: caseName,
    action,
    target_type: targetType,
    target_id: targetID,
    locale,
    item_operation: itemOperation,
    ...(policyType === undefined ? {} : { policy_type: policyType }),
    ...(versionNumber === undefined
      ? {}
      : { version_number: Number(versionNumber) }),
  };
}

function requiredString(value: unknown, name: string): string {
  if (typeof value !== "string" || value === "")
    throw new TypeError(`${name} is required`);
  return value;
}

function parseItemOperation(value: unknown): AuditItemOperation {
  if (value === "created" || value === "updated" || value === "deleted")
    return value;
  throw new TypeError("item_operation is invalid");
}

function buildFixtureRecord(fixture: LocaleContentFixture): AuditRecord {
  const args = [
    metadata,
    fixture.target_id,
    fixture.locale,
    fixture.item_operation,
  ] as const;
  switch (fixture.action) {
    case "post.updated":
      return buildPostLocaleContentAuditRecord(...args);
    case "page.updated":
      return buildPageLocaleContentAuditRecord(...args);
    case "work.updated":
      return buildWorkLocaleContentAuditRecord(...args);
    case "post_series.updated":
      return buildPostSeriesLocaleContentAuditRecord(...args);
    case "program_event.updated":
      return buildProgramEventLocaleContentAuditRecord(...args);
    case "release.updated":
      return buildReleaseLocaleContentAuditRecord(...args);
    case "artist.updated":
      return buildArtistLocaleContentAuditRecord(...args);
    case "label.updated":
      return buildLabelLocaleContentAuditRecord(...args);
    case "menu.updated":
      return buildMenuLocaleContentAuditRecord(...args);
    case "campaign.updated":
      return buildCampaignLocaleContentAuditRecord(...args);
    case "form.updated":
      return buildFormLocaleContentAuditRecord(...args);
    case "email_template.updated":
      return buildEmailTemplateLocaleContentAuditRecord(...args);
    case "email_layout.updated":
      return buildEmailLayoutLocaleContentAuditRecord(...args);
    case "legal_policy.updated":
      if (fixture.version_number === undefined)
        throw new TypeError("legal fixture requires version_number");
      if (fixture.policy_type === "privacy")
        return buildPrivacyLocaleContentAuditRecord(
          metadata,
          fixture.target_id,
          fixture.version_number,
          fixture.locale,
          fixture.item_operation,
        );
      if (fixture.policy_type === "terms")
        return buildTermsLocaleContentAuditRecord(
          metadata,
          fixture.target_id,
          fixture.version_number,
          fixture.locale,
          fixture.item_operation,
        );
      throw new TypeError("legal fixture requires policy_type");
    default:
      throw new TypeError(`unsupported fixture action ${fixture.action}`);
  }
}
