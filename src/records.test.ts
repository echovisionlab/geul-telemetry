import { readFile } from "node:fs/promises";

import { describe, expect, it, vi } from "vitest";

import {
  emitRequest,
  emitSystem,
  SYSTEM_EVENTS,
  validateAuditRecord,
  validateRequestRecord,
  validateSecurityAccessRecord,
  validateSystemRecord,
  type AuditRecord,
  type RequestRecord,
  type SecurityAccessRecord,
  type SystemRecord,
} from "./records.ts";

const fixturePath = new URL("../fixtures/request-record.json", import.meta.url);
const securityFixturePath = new URL(
  "../fixtures/security-access-records.json",
  import.meta.url,
);
const auditFixturePath = new URL(
  "../fixtures/audit-records.json",
  import.meta.url,
);

async function requestFixture(): Promise<RequestRecord> {
  return JSON.parse(await readFile(fixturePath, "utf8")) as RequestRecord;
}

describe("records", () => {
  it("matches the cross-language request fixture", async () => {
    const record = await requestFixture();
    expect(() => validateRequestRecord(record)).not.toThrow();
    expect(JSON.parse(JSON.stringify(record))).toEqual(record);
  });

  it("accepts an exact RPC boundary", async () => {
    const record = await requestFixture();
    const rpc: RequestRecord = {
      ...record,
      http_route: undefined,
      rpc_service: "geul.v1.PostService",
      rpc_method: "UpdatePost",
    };
    expect(() => validateRequestRecord(rpc)).not.toThrow();
  });

  it("rejects invalid request records", async () => {
    const record = await requestFixture();
    const invalid: RequestRecord[] = [
      { ...record, occurred_at: "local time" },
      { ...record, request_id: "bad" },
      { ...record, trace_id: "bad", span_id: undefined },
      { ...record, actor_kind: "member", actor_member_id: "" },
      { ...record, http_route: undefined },
      { ...record, status_code: 99 },
      { ...record, status_code: 200.5 },
      { ...record, duration_ms: -1 },
      { ...record, duration_ms: 1.5 },
      { ...record, error_code: "internal" },
      { ...record, outcome: "unknown" as RequestRecord["outcome"] },
      {
        ...record,
        outcome: "blocked",
        error_code: undefined,
        reason: undefined,
      },
      {
        ...record,
        status_code: 403,
        outcome: "failed",
        reason: "client_error",
      },
      {
        ...record,
        status_code: 500,
        outcome: "failed",
        reason: "unexpected",
      },
    ];
    for (const value of invalid) {
      expect(() => validateRequestRecord(value)).toThrow(TypeError);
    }
    expect(() =>
      validateRequestRecord({
        ...record,
        actor_kind: "system",
        actor_service: "",
      }),
    ).toThrow("actor_service");
    expect(() =>
      validateRequestRecord({
        ...record,
        actor_kind: "system",
        actor_member_id: undefined,
        actor_service: "geul-backend",
      } as unknown as RequestRecord),
    ).not.toThrow();
    for (const invalidActor of [
      { ...record, actor_kind: "anonymous", actor_member_id: "member-1" },
      {
        ...record,
        actor_kind: "member",
        actor_member_id: "member-1",
        actor_service: "api",
      },
      {
        ...record,
        actor_kind: "system",
        actor_service: "api",
        actor_member_id: "member-1",
      },
      { ...record, actor_kind: "other" },
    ] as unknown as RequestRecord[]) {
      expect(() => validateRequestRecord(invalidActor)).toThrow(TypeError);
    }
  });

  it("validates system records and numeric transport fields", () => {
    const base = {
      occurred_at: "2026-08-09T03:04:05Z",
      request_id: "018f47a2-8a3d-4e17-9d42-6f12c89b1234",
      component: "api",
      dependency: "postgres",
      operation: "connect",
      queue: "post",
      message_id: "message-1",
      command_id: "command-1",
      retry_count: 0,
      duration_ms: 4,
      job_kind: "mesh_optimization",
      job_id: "job-1",
      entity_type: "post",
      entity_id: "post-1",
      record_class: "domain_audit",
      action: "post.updated",
      error_code: "test_error",
    } as const;
    const queueFailureReasons: Partial<Record<SystemRecord["event"], string>> =
      {
        "queue.publish.failed": "enqueue_failed",
        "queue.delivery.failed": "handler_failed",
        "queue.delivery.requeued": "shutdown",
        "queue.retry.failed": "visibility_update_failed",
        "queue.dlq.failed": "archive_failed",
      };
    for (const event of SYSTEM_EVENTS) {
      const isTranslationTerminal = event === "translation.job.terminal";
      expect(() =>
        validateSystemRecord({
          ...base,
          event,
          error_code:
            event.startsWith("queue.") ||
            event === "job.failed" ||
            event === "audit.append.failed" ||
            event === "collaboration.checkpoint.failed" ||
            event === "client.render.failed" ||
            isTranslationTerminal
              ? undefined
              : base.error_code,
          reason:
            event === "audit.append.failed"
              ? "persistence_failed"
              : event === "collaboration.checkpoint.failed"
                ? "persist_failed"
                : event === "client.render.failed"
                  ? "react_error_boundary"
                  : event === "job.failed"
                    ? "internal"
                    : queueFailureReasons[event],
          error_classification: isTranslationTerminal ? "internal" : undefined,
          component:
            event === "client.render.failed"
              ? "general"
              : isTranslationTerminal
                ? undefined
                : base.component,
          domain:
            event === "collaboration.checkpoint.failed"
              ? "collaboration"
              : event === "client.render.failed"
                ? "client"
                : isTranslationTerminal
                  ? "translation"
                  : undefined,
          job_id: isTranslationTerminal
            ? "018f47a2-8a3d-4e17-9d42-6f12c89b1234"
            : base.job_id,
          entity_id: isTranslationTerminal ? undefined : base.entity_id,
          target_locale: isTranslationTerminal ? "en" : undefined,
          dependency: isTranslationTerminal ? undefined : base.dependency,
          operation: isTranslationTerminal ? undefined : base.operation,
          queue: isTranslationTerminal ? undefined : base.queue,
          message_id: isTranslationTerminal ? undefined : base.message_id,
          command_id: isTranslationTerminal ? undefined : base.command_id,
          retry_count: isTranslationTerminal ? undefined : base.retry_count,
          job_kind: isTranslationTerminal ? undefined : base.job_kind,
          record_class: isTranslationTerminal ? undefined : base.record_class,
          action: isTranslationTerminal ? undefined : base.action,
          outcome: {
            "service.ready": "ready",
            "service.degraded": "degraded",
            "service.stopping": "stopping",
            "service.failed": "failed",
            "dependency.degraded": "degraded",
            "dependency.recovered": "recovered",
            "queue.publish.succeeded": "succeeded",
            "queue.publish.failed": "failed",
            "queue.delivery.succeeded": "succeeded",
            "queue.delivery.failed": "failed",
            "queue.delivery.requeued": "requeued",
            "queue.retry.accepted": "accepted",
            "queue.retry.failed": "failed",
            "queue.dlq.accepted": "accepted",
            "queue.dlq.failed": "failed",
            "job.started": "started",
            "job.succeeded": "succeeded",
            "job.failed": "failed",
            "audit.append.failed": "failed",
            "collaboration.checkpoint.failed": "failed",
            "client.render.failed": "failed",
            "telemetry.pipeline.degraded": "degraded",
            "telemetry.pipeline.recovered": "recovered",
            "translation.job.terminal": "failed",
          }[event],
        }),
      ).not.toThrow();
    }
    expect(() =>
      validateSystemRecord({
        ...base,
        event: "audit.append.failed",
        outcome: "failed",
        record_class: "security_access",
        action: "authentication.failed",
        error_code: undefined,
        reason: "persistence_failed",
      }),
    ).not.toThrow();
    const record: SystemRecord = {
      ...base,
      event: "queue.delivery.succeeded",
      outcome: "succeeded",
    };
    for (const invalid of [
      { ...record, occurred_at: "" },
      { ...record, event: "unknown" },
      { ...record, request_id: "bad" },
      { ...record, trace_id: "bad" },
      { ...record, retry_count: -1 },
      { ...record, duration_ms: 0.5 },
      { ...base, event: "service.ready", component: "" },
      {
        ...base,
        event: "service.failed",
        error_code: undefined,
        reason: undefined,
      },
      {
        ...base,
        event: "queue.publish.succeeded",
        duration_ms: undefined,
      },
      {
        ...base,
        event: "service.failed",
        outcome: "failed",
        error_code: undefined,
        reason: undefined,
      },
      {
        ...base,
        event: "queue.publish.succeeded",
        outcome: "succeeded",
        error_code: undefined,
        duration_ms: undefined,
      },
      {
        ...base,
        event: "queue.publish.succeeded",
        outcome: "succeeded",
        error_code: "unexpected",
      },
      {
        ...base,
        event: "queue.publish.succeeded",
        outcome: "succeeded",
        error_code: undefined,
        reason: "unexpected",
      },
      {
        ...base,
        event: "queue.publish.failed",
        outcome: "failed",
        error_code: "unexpected",
        reason: "unsupported_reason",
      },
      {
        ...base,
        event: "queue.publish.failed",
        outcome: "failed",
        error_code: undefined,
        reason: "unexpected",
      },
      {
        ...base,
        event: "queue.publish.succeeded",
        outcome: "succeeded",
        error_code: undefined,
        queue: undefined,
      },
      {
        ...base,
        event: "queue.publish.succeeded",
        outcome: "succeeded",
        error_code: undefined,
        queue: "",
      },
      {
        ...base,
        event: "queue.delivery.succeeded",
        retry_count: undefined,
      },
      {
        ...base,
        event: "queue.delivery.succeeded",
        outcome: "succeeded",
        error_code: undefined,
        duration_ms: undefined,
      },
      {
        ...base,
        event: "queue.delivery.succeeded",
        outcome: "succeeded",
        error_code: undefined,
        reason: "unexpected",
      },
      {
        ...base,
        event: "queue.retry.accepted",
        retry_count: undefined,
      },
      {
        ...base,
        event: "queue.retry.accepted",
        outcome: "accepted",
        error_code: undefined,
        retry_count: undefined,
      },
      {
        ...base,
        event: "queue.retry.accepted",
        outcome: "accepted",
        error_code: "unexpected",
      },
      {
        ...base,
        event: "queue.retry.failed",
        outcome: "failed",
        error_code: undefined,
        reason: "unexpected",
      },
      { ...base, event: "job.succeeded", duration_ms: undefined },
      {
        ...base,
        event: "job.succeeded",
        outcome: "succeeded",
        duration_ms: undefined,
      },
      {
        ...base,
        event: "job.failed",
        outcome: "failed",
        error_code: undefined,
        reason: undefined,
      },
      {
        ...base,
        event: "audit.append.failed",
        outcome: "failed",
        record_class: "domain_audit",
        action: "unknown.action",
        error_code: undefined,
        reason: "persistence_failed",
      },
      {
        ...base,
        event: "audit.append.failed",
        outcome: "failed",
        record_class: "security_access",
        action: "unknown.action",
        error_code: undefined,
        reason: "persistence_failed",
      },
      {
        ...base,
        event: "audit.append.failed",
        outcome: "failed",
        record_class: "other",
        action: "authentication.failed",
        error_code: undefined,
        reason: "persistence_failed",
      },
      {
        ...base,
        event: "audit.append.failed",
        outcome: "failed",
        record_class: "domain_audit",
        action: "post.updated",
        error_code: "unexpected",
        reason: "persistence_failed",
      },
      {
        ...base,
        event: "audit.append.failed",
        outcome: "failed",
        record_class: "domain_audit",
        action: "post.updated",
        error_code: undefined,
        reason: "unexpected",
      },
      {
        ...base,
        event: "audit.append.failed",
        outcome: "failed",
        record_class: "domain_audit",
        action: "post.updated",
        error_code: undefined,
        reason: undefined,
      },
      {
        ...base,
        event: "collaboration.checkpoint.failed",
        outcome: "failed",
        domain: "collaboration",
        error_code: undefined,
        reason: "persist_failed",
        entity_id: "",
      },
      {
        ...base,
        event: "collaboration.checkpoint.failed",
        outcome: "failed",
        domain: "collaboration",
        error_code: undefined,
        reason: "unknown",
      },
      {
        ...base,
        event: "collaboration.checkpoint.failed",
        outcome: "failed",
        domain: "collaboration",
        entity_type: "unknown",
        error_code: undefined,
        reason: "persist_failed",
      },
      {
        ...base,
        event: "collaboration.checkpoint.failed",
        outcome: "failed",
        domain: "collaboration",
        retry_count: 2,
        error_code: undefined,
        reason: "document_revision_changed",
      },
      {
        ...base,
        event: "client.render.failed",
        outcome: "failed",
        domain: "client",
        component: "other",
        error_code: undefined,
        reason: "react_error_boundary",
      },
    ]) {
      expect(() => validateSystemRecord(invalid as SystemRecord)).toThrow(
        TypeError,
      );
    }
  });

  it("validates audit records", () => {
    const base: AuditRecord = {
      audit_id: "018f47a2-8a3d-4e17-9d42-6f12c89b1234",
      occurred_at: "2026-08-09T03:04:05Z",
      action: "site_settings.updated",
      request_id: "018f47a2-8a3d-4e17-9d42-6f12c89b1234",
      actor_kind: "member",
      actor_member_id: "member-1",
      target_type: "site_settings",
      target_id: "1",
      changed_fields: ["primary_color", "site_title"],
    };
    const contributors = [
      "1b6bcad2-c90d-49e9-bec7-f9a4ba6b2894",
      "7a7a8fd4-1f69-4e9a-9dc2-2378926ff351",
    ] as const;
    const member = {
      actor_kind: "member",
      actor_member_id: "member-1",
    } as const;
    const system = {
      actor_kind: "system",
      actor_member_id: undefined,
      actor_service: "geul-backend",
    } as const;
    const collabSystem = {
      actor_kind: "system",
      actor_member_id: undefined,
      actor_service: "geul-collab",
    } as const;
    const valid = [
      base,
      {
        ...base,
        action: "member.updated",
        target_type: "member",
        target_id: "member-1",
        changed_fields: ["nickname", "onboarded"],
        nickname: "Member",
      },
      {
        ...base,
        ...system,
        action: "member.updated",
        target_type: "member",
        target_id: "member-1",
        changed_fields: ["role"],
        previous_role: "user",
        new_role: "admin",
      },
      {
        ...base,
        action: "member.updated",
        target_type: "member",
        target_id: "member-1",
        changed_fields: ["status"],
        previous_state: "active",
        new_state: "banned",
      },
      {
        ...base,
        ...system,
        action: "member.updated",
        target_type: "member",
        target_id: "member-1",
        changed_fields: ["status"],
        previous_state: "banned",
        new_state: "active",
      },
      ...(["post", "page", "work"] as const).map((domain) => ({
        ...base,
        ...collabSystem,
        action: `${domain}.updated`,
        target_type: domain,
        target_id: `${domain}-1`,
        changed_fields: ["version"],
        version_id: "version-1",
        contributor_member_ids: contributors,
      })),
      ...(["post", "page", "work"] as const).flatMap((domain) =>
        (["created", "deleted"] as const).map((operation) => ({
          ...base,
          ...member,
          action: `${domain}.${operation}`,
          target_type: domain,
          target_id: `${domain}-1`,
          changed_fields: undefined,
        })),
      ),
      {
        ...base,
        action: "account.updated",
        target_type: "account",
        target_id: "member-1",
        changed_fields: ["canonical_email"],
        previous_email: "old@example.test",
        new_email: "new@example.test",
      },
      {
        ...base,
        action: "account.updated",
        target_type: "account",
        target_id: "member-1",
        changed_fields: ["login_emails"],
        collection_operation: "added",
        email: "added@example.test",
      },
      {
        ...base,
        action: "account.updated",
        target_type: "account",
        target_id: "member-1",
        changed_fields: ["social_logins"],
        collection_operation: "removed",
        provider: "google",
        provider_subject: "subject-1",
      },
      {
        ...base,
        action: "account.updated",
        target_type: "account",
        target_id: "member-1",
        changed_fields: ["passkeys"],
        collection_operation: "added",
        passkey_ids: ["passkey-1"],
      },
      {
        ...base,
        action: "account.updated",
        target_type: "account",
        target_id: "member-1",
        changed_fields: ["sessions"],
        collection_operation: "removed",
        session_scope: "one",
        session_ids: [base.audit_id],
      },
      ...(
        [
          ["none", "confirmation_pending"],
          ["cancelled", "confirmation_pending"],
          ["recovered", "confirmation_pending"],
          ["confirmation_pending", "scheduled"],
          ["none", "scheduled"],
          ["cancelled", "scheduled"],
          ["recovered", "scheduled"],
          ["scheduled", "cancelled"],
          ["recovery_confirmation_pending", "recovered"],
        ] as const
      ).map(([previousState, newState]) => ({
        ...base,
        action: "account.updated",
        target_type: "account",
        target_id: "member-1",
        changed_fields: ["deletion_state"],
        previous_state: previousState,
        new_state: newState,
      })),
      {
        ...base,
        ...system,
        action: "account.deleted",
        target_type: "account",
        target_id: "member-1",
        changed_fields: undefined,
      },
    ];
    for (const record of valid) {
      expect(() => validateAuditRecord(record as AuditRecord)).not.toThrow();
    }

    const invalid = [
      { ...base, occurred_at: "" },
      { ...base, audit_id: "" },
      { ...base, action: "member.role.updated" },
      { ...base, target_type: "page" },
      { ...base, target_id: "2" },
      { ...base, request_id: "bad" },
      { ...base, trace_id: "bad" },
      { ...base, actor_member_id: "" },
      { ...base, actor_kind: "anonymous", actor_member_id: undefined },
      { ...base, changed_fields: undefined },
      { ...base, changed_fields: ["site_title", "primary_color"] },
      { ...base, changed_fields: ["bad-value"] },
      { ...base, version_id: "version-1" },
      { ...valid[1], nickname: " " },
      { ...valid[1], changed_fields: ["nickname", "wrong"] },
      { ...valid[2], new_role: "user" },
      { ...valid[2], changed_fields: ["status"] },
      { ...valid[3], new_state: "active" },
      { ...valid[3], new_state: "recovered" },
      { ...valid[3], changed_fields: ["unknown"] },
      { ...valid[3], ...system },
      { ...valid[6], changed_fields: undefined },
      { ...valid[6], contributor_member_ids: undefined },
      { ...valid[6], contributor_member_ids: ["member-1"] },
      {
        ...valid[6],
        contributor_member_ids: [contributors[1], contributors[0]],
      },
      {
        ...base,
        ...system,
        action: "post.created",
        target_type: "post",
        target_id: "post-1",
        changed_fields: undefined,
      },
      { ...valid[14], previous_email: "invalid" },
      { ...valid[14], new_email: "old@example.test" },
      { ...valid[14], target_type: "member" },
      { ...valid[15], email: "invalid" },
      { ...valid[15], email: undefined },
      { ...valid[15], collection_operation: "replaced" },
      { ...valid[16], provider: "Bad Provider" },
      { ...valid[16], provider: undefined },
      { ...valid[16], provider_subject: "" },
      { ...valid[16], provider_subject: undefined },
      { ...valid[17], passkey_ids: [] },
      { ...valid[17], passkey_ids: [" bad-identifier"] },
      { ...valid[17], passkey_ids: ["passkey-1", "passkey-1"] },
      { ...valid[18], collection_operation: "added" },
      { ...valid[18], session_scope: "all" },
      { ...valid[18], session_ids: undefined },
      { ...valid[18], session_ids: ["invalid"] },
      { ...valid[18], session_scope: "others", session_ids: [] },
      {
        ...base,
        action: "account.updated",
        target_type: "account",
        target_id: "member-1",
        changed_fields: undefined,
      },
      {
        ...base,
        action: "account.updated",
        target_type: "account",
        target_id: "member-1",
        changed_fields: ["deletion_state"],
        previous_state: "none",
        new_state: "banned",
      },
      {
        ...base,
        action: "account.updated",
        target_type: "account",
        target_id: "member-1",
        changed_fields: ["unknown"],
      },
      {
        ...valid.at(-1),
        changed_fields: ["deletion_state"],
      },
    ];
    for (const [index, record] of invalid.entries()) {
      expect(
        () => validateAuditRecord(record as AuditRecord),
        `invalid audit fixture ${index}: ${JSON.stringify(record)}`,
      ).toThrow(TypeError);
    }
  });
  it("matches the exact cross-language audit fixture", async () => {
    const records = JSON.parse(
      await readFile(auditFixturePath, "utf8"),
    ) as AuditRecord[];
    expect(records).toHaveLength(27);
    for (const record of records) {
      expect(() => validateAuditRecord(record)).not.toThrow();
    }
  });

  it("validates security access actions and bounded authentication reasons", () => {
    const base: SecurityAccessRecord = {
      access_id: "7a7a8fd4-1f69-4e9a-9dc2-2378926ff351",
      occurred_at: "2026-08-09T03:04:05Z",
      action: "authentication.succeeded",
      request_id: "018f47a2-8a3d-4e17-9d42-6f12c89b1234",
      actor_kind: "member",
      actor_member_id: "member-1",
      source_ip: "192.0.2.4",
      flow_kind: "login",
      authentication_method: "passkey",
      principal_state: "active",
    };
    const valid: SecurityAccessRecord[] = [
      base,
      {
        ...base,
        action: "authentication.succeeded",
        flow_kind: "registration",
        authentication_method: "oidc",
        principal_state: "onboarding_only",
        provider: "google",
      },
      {
        ...base,
        action: "authentication.failed",
        flow_kind: "reauthentication",
        authentication_method: "email_code",
        principal_state: undefined,
        reason: "proof_rejected",
      },
      {
        ...base,
        action: "authentication.failed",
        actor_kind: "anonymous",
        actor_member_id: undefined,
        flow_kind: "login",
        authentication_method: "email_code",
        principal_state: undefined,
        reason: "proof_rejected",
      },
      {
        ...base,
        action: "authentication.blocked",
        actor_kind: "anonymous",
        actor_member_id: undefined,
        flow_kind: undefined,
        authentication_method: undefined,
        principal_state: undefined,
        reason: "rate_limited",
      },
      {
        ...base,
        action: "authorization.denied",
        actor_kind: "anonymous",
        actor_member_id: undefined,
        flow_kind: undefined,
        authentication_method: undefined,
        principal_state: undefined,
        attempted_action: "/geul.api.v1.PostService/UpdatePost",
        permission: "procedure:invoke",
        reason: "permission_denied",
      },
      {
        ...base,
        action: "personal_data.accessed",
        flow_kind: undefined,
        authentication_method: undefined,
        principal_state: undefined,
        subject_type: "member",
        subject_id: "2a7a8fd4-1f69-4e9a-9dc2-2378926ff351",
        access_kind: "read",
        data_category: "member_administration",
      },
      {
        ...base,
        action: "personal_data.accessed",
        flow_kind: undefined,
        authentication_method: undefined,
        principal_state: undefined,
        subject_type: "member_collection",
        subject_id: "1",
        access_kind: "read",
        data_category: "member_administration",
      },
      {
        ...base,
        action: "personal_data.accessed",
        flow_kind: undefined,
        authentication_method: undefined,
        principal_state: undefined,
        subject_type: "campaign",
        subject_id: "2a7a8fd4-1f69-4e9a-9dc2-2378926ff351",
        access_kind: "read",
        data_category: "campaign_recipients",
      },
      {
        ...base,
        action: "personal_data.accessed",
        flow_kind: undefined,
        authentication_method: undefined,
        principal_state: undefined,
        subject_type: "form",
        subject_id: "2a7a8fd4-1f69-4e9a-9dc2-2378926ff351",
        access_kind: "read",
        data_category: "form_submissions",
      },
      {
        ...base,
        action: "personal_data.accessed",
        flow_kind: undefined,
        authentication_method: undefined,
        principal_state: undefined,
        subject_type: "form_submission",
        subject_id: "2a7a8fd4-1f69-4e9a-9dc2-2378926ff351",
        access_kind: "read",
        data_category: "form_submission",
      },
    ];
    for (const record of valid) {
      expect(() => validateSecurityAccessRecord(record)).not.toThrow();
    }

    const invalid: SecurityAccessRecord[] = [
      { ...base, occurred_at: "" },
      { ...base, access_id: "" },
      { ...base, access_id: "access-1" },
      { ...base, request_id: "bad" },
      { ...base, trace_id: "bad" },
      { ...base, actor_member_id: "" },
      { ...base, source_ip: "bad" },
      { ...base, source_ip: "2001:0db8::4" },
      { ...base, source_ip: "2001:DB8::4" },
      { ...base, flow_kind: "bad" as SecurityAccessRecord["flow_kind"] },
      {
        ...base,
        principal_state: "bad" as SecurityAccessRecord["principal_state"],
      },
      { ...base, action: "authentication.failed", reason: "proof_rejected" },
      {
        ...base,
        action: "authentication.failed",
        actor_kind: "system",
        actor_member_id: undefined,
        actor_service: "geul-backend",
        reason: "proof_rejected",
      },
      {
        ...base,
        action: "authentication.failed",
        actor_kind: "anonymous",
        actor_member_id: undefined,
        authentication_method: undefined,
        reason: "proof_rejected",
      },
      { ...base, actor_kind: "anonymous", actor_member_id: undefined },
      { ...base, provider: "google" },
      {
        ...base,
        authentication_method: "oidc",
        provider: "INVALID",
      },
      { ...base, reason: "unexpected" },
      {
        ...base,
        action: "authentication.failed",
        actor_kind: "anonymous",
        actor_member_id: undefined,
        reason: "bad",
      },
      {
        ...valid[2],
        actor_kind: "system",
        actor_member_id: undefined,
        actor_service: "geul-backend",
      },
      { ...valid[2], flow_kind: "bad" as SecurityAccessRecord["flow_kind"] },
      {
        ...valid[2],
        authentication_method:
          "bad" as SecurityAccessRecord["authentication_method"],
      },
      { ...valid[2], flow_kind: "login" },
      { ...valid[2], reason: "bad" },
      {
        ...base,
        action: "authentication.blocked",
        actor_kind: "anonymous",
        actor_member_id: undefined,
        reason: "bad",
      },
      {
        ...valid[4],
        actor_kind: "system",
        actor_member_id: undefined,
        actor_service: "geul-backend",
      },
      {
        ...valid[4],
        flow_kind: "bad" as SecurityAccessRecord["flow_kind"],
      },
      {
        ...valid[4],
        authentication_method:
          "bad" as SecurityAccessRecord["authentication_method"],
      },
      {
        ...valid[4],
        actor_kind: "member",
        actor_member_id: "member-1",
        flow_kind: "login",
      },
      { ...valid[4], reason: "bad" },
      {
        ...base,
        action: "authentication.blocked",
        actor_kind: "system",
        actor_member_id: undefined,
        actor_service: "geul-backend",
        reason: "rate_limited",
      },
      {
        ...base,
        action: "authentication.blocked",
        actor_kind: "anonymous",
        actor_member_id: undefined,
        flow_kind: "bad" as SecurityAccessRecord["flow_kind"],
        reason: "rate_limited",
      },
      {
        ...base,
        action: "authentication.blocked",
        actor_kind: "anonymous",
        actor_member_id: undefined,
        authentication_method:
          "bad" as SecurityAccessRecord["authentication_method"],
        reason: "rate_limited",
      },
      {
        ...base,
        action: "authentication.blocked",
        flow_kind: "login",
        reason: "rate_limited",
      },
      { ...base, action: "authorization.denied" },
      {
        ...base,
        action: "authorization.denied",
        actor_kind: "system",
        actor_member_id: undefined,
        actor_service: "api",
        flow_kind: undefined,
        authentication_method: undefined,
        attempted_action: "/geul.api.v1.PostService/UpdatePost",
        permission: "procedure:invoke",
        reason: "permission_denied",
      },
      {
        ...valid[5],
        actor_kind: "system",
        actor_member_id: undefined,
        actor_service: "api",
      },
      { ...valid[5], attempted_action: undefined },
      { ...valid[5], permission: undefined },
      { ...valid[5], attempted_action: "bad" },
      { ...valid[5], permission: "invalid" },
      { ...valid[5], reason: "bad" },
      {
        ...base,
        action: "authorization.denied",
        flow_kind: undefined,
        authentication_method: undefined,
        attempted_action: "/geul.api.v1.PostService/UpdatePost",
        permission: "procedure:invoke",
        reason: undefined,
      },
      {
        ...base,
        action: "authorization.denied",
        attempted_action: "/geul.api.v1.PostService/UpdatePost",
        permission: "procedure:invoke",
        reason: "INVALID",
      },
      {
        ...base,
        action: "authorization.denied",
        flow_kind: undefined,
        authentication_method: undefined,
        attempted_action: "/geul.api.v1.PostService/UpdatePost",
        permission: "INVALID",
        reason: "permission_denied",
      },
      {
        ...base,
        action: "authorization.denied",
        flow_kind: undefined,
        authentication_method: undefined,
        attempted_action: "geul.api.v1.PostService/UpdatePost",
        permission: "procedure:invoke",
        reason: "permission_denied",
      },
      {
        ...base,
        action: "authorization.denied",
        flow_kind: undefined,
        authentication_method: undefined,
        attempted_action: "/geul.api.v1.PostService/Update/Post",
        permission: "procedure:invoke",
        reason: "permission_denied",
      },
      {
        ...base,
        action: "authorization.denied",
        flow_kind: undefined,
        authentication_method: undefined,
        attempted_action: "/geul.api.v1.PostService/UpdatePost?target=1",
        permission: "procedure:invoke",
        reason: "permission_denied",
      },
      { ...base, action: "personal_data.accessed" },
      {
        ...base,
        action: "personal_data.accessed",
        flow_kind: undefined,
        authentication_method: undefined,
        subject_type: "member",
        subject_id: "2a7a8fd4-1f69-4e9a-9dc2-2378926ff351",
        access_kind: "read",
        data_category: undefined,
      },
      {
        ...base,
        action: "personal_data.accessed",
        flow_kind: undefined,
        authentication_method: undefined,
        subject_type: "INVALID",
        subject_id: "2a7a8fd4-1f69-4e9a-9dc2-2378926ff351",
        access_kind: "read",
        data_category: "member_administration",
      },
      {
        ...valid[6],
        actor_kind: "anonymous",
        actor_member_id: undefined,
      },
      {
        ...valid[6],
        access_kind: "invalid" as SecurityAccessRecord["access_kind"],
      },
      {
        ...valid[6],
        subject_type: undefined,
        subject_id: undefined,
        data_category: undefined,
      },
      { ...valid[7], subject_id: "2" },
      { ...valid[7], data_category: "campaign_recipients" },
      { ...valid[6], subject_id: "member-2" },
      { ...valid[8], data_category: "member_administration" },
      { ...valid[9], data_category: "member_administration" },
      { ...valid[10], data_category: "member_administration" },
      { ...valid[6], subject_type: "unknown" },
      {
        ...base,
        action: "personal_data.accessed",
        flow_kind: undefined,
        authentication_method: undefined,
        subject_type: "member",
        subject_id: "member-2",
        access_kind: "read",
        data_category: "member_administration",
      },
      {
        ...base,
        attempted_action: "post.update",
      },
      { ...base, action: "unknown" as SecurityAccessRecord["action"] },
    ];
    for (const record of invalid) {
      expect(() => validateSecurityAccessRecord(record)).toThrow(TypeError);
    }
  });

  it("matches the cross-language security access fixture", async () => {
    const records = JSON.parse(
      await readFile(securityFixturePath, "utf8"),
    ) as SecurityAccessRecord[];
    expect(records).toHaveLength(5);
    for (const record of records) {
      expect(() => validateSecurityAccessRecord(record)).not.toThrow();
    }
  });

  it("emits only validated typed records", async () => {
    const record = await requestFixture();
    const sink = { write: vi.fn() };
    await emitRequest(sink, record);
    const system: SystemRecord = {
      event: "service.ready",
      occurred_at: "2026-08-09T03:04:05Z",
      component: "api",
      outcome: "ready",
    };
    await emitSystem(sink, system);
    expect(sink.write).toHaveBeenNthCalledWith(1, record);
    expect(sink.write).toHaveBeenNthCalledWith(2, system);
    await expect(
      emitSystem(sink, {
        ...system,
        event: "unknown" as SystemRecord["event"],
      }),
    ).rejects.toThrow();
  });

  it("rejects malformed trace and timestamp variants", async () => {
    const record = await requestFixture();
    for (const correlation of [
      {
        trace_id: "00000000000000000000000000000000",
        span_id: "00f067aa0ba902b7",
      },
      {
        trace_id: "4bf92f3577b34da6a3ce929d0e0e4736",
        span_id: "0000000000000000",
      },
      { trace_id: "4bf92f3577b34da6a3ce929d0e0e4736", span_id: "bad" },
    ]) {
      expect(() =>
        validateRequestRecord({ ...record, ...correlation }),
      ).toThrow("trace_id");
    }
    expect(() =>
      validateRequestRecord({
        ...record,
        occurred_at: "2026-08-09T03:04:05+09:00",
      }),
    ).toThrow("UTC");
  });
});
