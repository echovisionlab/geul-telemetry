import { describe, expect, it } from "vitest";

import {
  buildPostSeriesSourceLocaleAuditRecord,
  buildPrivacySourceLocaleAuditRecord,
  type AuditMetadata,
} from "../audit.ts";
import { validateSourceLocaleAuditRecord } from "./source-locale.ts";
import { validateAuditRecord } from "../records.ts";

const member: AuditMetadata = {
  audit_id: "00000000-0000-4000-8000-000000000001",
  occurred_at: "2026-08-09T03:04:05Z",
  actor_kind: "member",
  actor_member_id: "member-1",
};

describe("source locale audit", () => {
  it("requires a member actor and distinct bounded locales", () => {
    const record = buildPostSeriesSourceLocaleAuditRecord(
      member,
      "series-1",
      "en",
      "zh-CN",
    );
    expect(() => validateAuditRecord(record)).not.toThrow();
    const legalRecord = buildPrivacySourceLocaleAuditRecord(
      member,
      "privacy-1",
      2,
      "en",
      "ko",
    );
    expect(() =>
      validateAuditRecord({
        ...legalRecord,
        actor_kind: "system",
        actor_member_id: undefined,
        actor_service: "geul-backend",
      }),
    ).toThrow("source_locale requires a member actor");
    expect(() => validateAuditRecord({ ...record, new_locale: "en" })).toThrow(
      "distinct bounded",
    );
    expect(() =>
      validateAuditRecord({ ...record, previous_locale: "ko_KR" }),
    ).toThrow("distinct bounded");
    expect(() =>
      validateAuditRecord({ ...record, new_locale: "a".repeat(65) }),
    ).toThrow("distinct bounded");
    expect(() =>
      validateAuditRecord({ ...record, previous_locale: undefined }),
    ).toThrow("distinct bounded");
    expect(
      validateSourceLocaleAuditRecord({
        ...record,
        changed_fields: undefined,
      }),
    ).toBe(false);
    expect(() =>
      validateSourceLocaleAuditRecord({
        ...record,
        action: "map_theme.updated",
      }),
    ).toThrow("does not support source_locale");
  });

  it("keeps legal policy identity on the locale transition", () => {
    const record = buildPrivacySourceLocaleAuditRecord(
      member,
      "privacy-1",
      2,
      "en",
      "ko",
    );
    expect(() => validateAuditRecord(record)).not.toThrow();
    expect(() =>
      validateAuditRecord({ ...record, version_number: undefined }),
    ).toThrow("legal policy requires policy_type and version_number");
  });
});
