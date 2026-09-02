import { readFile } from "node:fs/promises";

import { describe, expect, it } from "vitest";

import {
  COLLABORATION_CHECKPOINT_ENTITY_TYPES,
  SYSTEM_EVENTS,
  SYSTEM_EVENT_OUTCOMES,
  TRANSLATION_ENTITY_TYPES,
  TRANSLATION_FAILURE_REASONS,
  TRANSLATION_JOB_TERMINAL_OUTCOMES,
  validateSystemRecord,
  type TranslationFailureReason,
  type TranslationJobTerminalOutcome,
  type SystemRecord,
} from "./records.ts";
import {
  buildClientRenderFailedRecord,
  buildCollaborationCheckpointFailedRecord,
  buildDomainAuditAppendFailedRecord,
  buildDependencyDegradedRecord,
  buildDependencyRecoveredRecord,
  buildJobFailedRecord,
  buildJobStartedRecord,
  buildJobSucceededRecord,
  buildQueueDeliveryFailedRecord,
  buildQueueDeliveryRequeuedRecord,
  buildQueueDeliverySucceededRecord,
  buildQueueDLQAcceptedRecord,
  buildQueueDLQFailedRecord,
  buildQueuePublishFailedRecord,
  buildQueuePublishSucceededRecord,
  buildQueueRetryAcceptedRecord,
  buildQueueRetryFailedRecord,
  buildServiceDegradedRecord,
  buildServiceFailedRecord,
  buildServiceReadyRecord,
  buildServiceStoppingRecord,
  buildSecurityAccessAppendFailedRecord,
  buildTelemetryPipelineDegradedRecord,
  buildTelemetryPipelineRecoveredRecord,
  buildTranslationJobTerminalRecord,
  systemLogLevel,
} from "./system.ts";

describe("system catalog builders", () => {
  const metadata = {
    occurred_at: "2026-08-09T03:04:05Z",
    request_id: "018f47a2-8a3d-4e17-9d42-6f12c89b1234",
  } as const;
  const failure = { reason: "dependency_unavailable" } as const;
  const publish = {
    queue: "post",
    message_id: "message-1",
    command_id: "command-1",
    duration_ms: 4,
  } as const;
  const delivery = {
    queue: "post",
    message_id: "message-1",
    command_id: "command-1",
    retry_count: 1,
    duration_ms: 8,
  } as const;
  const handoff = {
    queue: "post",
    message_id: "message-1",
    command_id: "command-1",
    retry_count: 2,
  } as const;
  const job = { job_kind: "mesh_optimization", job_id: "job-1" } as const;
  const jobFailure = { reason: "internal" } as const;
  const translationTerminal = {
    job_id: "018f47a2-8a3d-4e17-9d42-6f12c89b1234",
    entity_type: "post",
    target_locale: "en",
    duration_ms: 10,
  } as const;

  it("matches the cross-language event and outcome fixture", async () => {
    const fixture = JSON.parse(
      await readFile(
        new URL("../fixtures/system-catalog.json", import.meta.url),
        "utf8",
      ),
    ) as { event: (typeof SYSTEM_EVENTS)[number]; outcome: string }[];
    expect(fixture).toEqual(
      SYSTEM_EVENTS.flatMap<{
        event: (typeof SYSTEM_EVENTS)[number];
        outcome: string;
      }>((event) =>
        event === "translation.job.terminal"
          ? TRANSLATION_JOB_TERMINAL_OUTCOMES.map((outcome) => ({
              event,
              outcome,
            }))
          : [{ event, outcome: SYSTEM_EVENT_OUTCOMES[event]! }],
      ),
    );
  });

  it("matches the cross-language collaboration checkpoint entity fixture", async () => {
    const fixture = JSON.parse(
      await readFile(
        new URL(
          "../fixtures/collaboration-checkpoint-entity-types.json",
          import.meta.url,
        ),
        "utf8",
      ),
    ) as string[];
    expect(fixture).toEqual(COLLABORATION_CHECKPOINT_ENTITY_TYPES);
    for (const entity_type of COLLABORATION_CHECKPOINT_ENTITY_TYPES) {
      expect(() =>
        buildCollaborationCheckpointFailedRecord(
          metadata,
          { entity_type, entity_id: "entity-1", retry_count: 4 },
          "persist_failed",
        ),
      ).not.toThrow();
    }
  });

  it("matches the cross-language translation terminal fixture", async () => {
    const fixture = JSON.parse(
      await readFile(
        new URL(
          "../fixtures/translation-job-terminal-catalog.json",
          import.meta.url,
        ),
        "utf8",
      ),
    ) as {
      event: "translation.job.terminal";
      outcomes: TranslationJobTerminalOutcome[];
      entity_types: (typeof TRANSLATION_ENTITY_TYPES)[number][];
      failure_reasons: TranslationFailureReason[];
      valid_job_ids: string[];
      invalid_job_ids: string[];
    };
    expect(fixture.event).toBe("translation.job.terminal");
    expect(fixture.outcomes).toEqual(TRANSLATION_JOB_TERMINAL_OUTCOMES);
    for (const outcome of fixture.outcomes) {
      const record = buildTranslationJobTerminalRecord(
        metadata,
        { ...translationTerminal, target_locale: "zh-CN" },
        outcome,
        outcome === "failed" ? "internal" : undefined,
      );
      expect(record).toMatchObject({ event: fixture.event, outcome });
    }
    expect(fixture.entity_types).toEqual(TRANSLATION_ENTITY_TYPES);
    for (const entity_type of fixture.entity_types) {
      expect(() =>
        buildTranslationJobTerminalRecord(
          metadata,
          { ...translationTerminal, entity_type },
          "failed",
          "internal",
        ),
      ).not.toThrow();
    }
    expect(fixture.failure_reasons).toEqual(TRANSLATION_FAILURE_REASONS);
    for (const reason of fixture.failure_reasons) {
      expect(() =>
        buildTranslationJobTerminalRecord(
          metadata,
          translationTerminal,
          "failed",
          reason,
        ),
      ).not.toThrow();
    }
    for (const job_id of fixture.valid_job_ids) {
      expect(() =>
        buildTranslationJobTerminalRecord(
          metadata,
          { ...translationTerminal, job_id },
          "failed",
          "internal",
        ),
      ).not.toThrow();
    }
    for (const job_id of fixture.invalid_job_ids) {
      expect(() =>
        buildTranslationJobTerminalRecord(
          metadata,
          { ...translationTerminal, job_id },
          "failed",
          "internal",
        ),
      ).toThrow();
    }
  });

  it("provides one typed builder for every registered event", () => {
    const records = [
      buildServiceReadyRecord(metadata, "api"),
      buildServiceDegradedRecord(metadata, "api", failure),
      buildServiceStoppingRecord(metadata, "api"),
      buildServiceFailedRecord(metadata, "api", failure),
      buildDependencyDegradedRecord(
        metadata,
        "postgres",
        "read_queue",
        failure,
      ),
      buildDependencyRecoveredRecord(metadata, "postgres", "read_queue"),
      buildQueuePublishSucceededRecord(metadata, publish),
      buildQueuePublishFailedRecord(metadata, publish, "enqueue_failed"),
      buildQueueDeliverySucceededRecord(metadata, delivery),
      buildQueueDeliveryFailedRecord(metadata, delivery, "handler_failed"),
      buildQueueDeliveryFailedRecord(metadata, delivery, "completion_failed"),
      buildQueueDeliveryRequeuedRecord(metadata, delivery, "shutdown"),
      buildQueueRetryAcceptedRecord(metadata, handoff),
      buildQueueRetryFailedRecord(
        metadata,
        handoff,
        "visibility_update_failed",
      ),
      buildQueueDLQAcceptedRecord(metadata, handoff),
      buildQueueDLQFailedRecord(metadata, handoff, "archive_failed"),
      buildJobStartedRecord(metadata, job),
      buildJobSucceededRecord(metadata, job, 10),
      buildJobFailedRecord(metadata, job, 10, jobFailure),
      buildTranslationJobTerminalRecord(
        metadata,
        translationTerminal,
        "applied",
      ),
      buildTranslationJobTerminalRecord(
        metadata,
        translationTerminal,
        "failed",
        "internal",
      ),
      buildDomainAuditAppendFailedRecord(
        metadata,
        "post.updated",
        "persistence_failed",
      ),
      buildCollaborationCheckpointFailedRecord(
        metadata,
        { entity_type: "post", entity_id: "post-1", retry_count: 4 },
        "persist_failed",
      ),
      buildClientRenderFailedRecord(metadata, "general"),
      buildTelemetryPipelineDegradedRecord(metadata, "otlp_exporter", failure),
      buildTelemetryPipelineRecoveredRecord(metadata, "otlp_exporter"),
    ];
    expect(new Set(records.map((record) => record.event))).toEqual(
      new Set(SYSTEM_EVENTS),
    );
  });

  it("accepts a first-attempt collaboration conflict and rejects retrying it", () => {
    expect(
      buildCollaborationCheckpointFailedRecord(
        metadata,
        { entity_type: "post", entity_id: "post-1", retry_count: 1 },
        "shared_structure_changed",
      ),
    ).toMatchObject({
      reason: "shared_structure_changed",
      retry_count: 1,
    });
    expect(() =>
      buildCollaborationCheckpointFailedRecord(
        metadata,
        { entity_type: "post", entity_id: "post-1", retry_count: 2 },
        "shared_structure_changed",
      ),
    ).toThrow(TypeError);
    expect(
      buildCollaborationCheckpointFailedRecord(
        metadata,
        { entity_type: "post", entity_id: "post-1", retry_count: 1 },
        "target_revision_changed",
      ),
    ).toMatchObject({
      reason: "target_revision_changed",
      retry_count: 1,
    });
  });

  it("rejects incomplete or unbounded data", () => {
    expect(() => buildServiceReadyRecord(metadata, "API Server")).toThrow();
    expect(() =>
      buildQueuePublishFailedRecord(
        metadata,
        { ...publish, queue: "" },
        "enqueue_failed",
      ),
    ).toThrow();
    expect(() =>
      buildTranslationJobTerminalRecord(
        metadata,
        translationTerminal,
        "failed",
      ),
    ).toThrow();
    expect(() =>
      buildTranslationJobTerminalRecord(
        metadata,
        translationTerminal,
        "applied",
        "internal",
      ),
    ).toThrow();
    expect(() =>
      buildTranslationJobTerminalRecord(
        metadata,
        { ...translationTerminal, job_id: "job-1" },
        "failed",
        "internal",
      ),
    ).toThrow();
    expect(() =>
      buildTranslationJobTerminalRecord(
        metadata,
        { ...translationTerminal, target_locale: "bad_locale!" },
        "failed",
        "internal",
      ),
    ).toThrow();
    const failedRecord = buildTranslationJobTerminalRecord(
      metadata,
      translationTerminal,
      "failed",
      "internal",
    );
    expect(() =>
      validateSystemRecord({ ...failedRecord, outcome: "unknown" }),
    ).toThrow("translation.job.terminal requires a catalog outcome");
    expect(() =>
      validateSystemRecord({ ...failedRecord, outcome: "applied" }),
    ).toThrow(
      "non-failed translation.job.terminal cannot contain error_classification",
    );
    expect(() =>
      validateSystemRecord({
        ...buildServiceFailedRecord(metadata, "api", failure),
        error_classification: "internal",
      }),
    ).toThrow("service.failed does not allow error_classification");
    expect(() =>
      validateSystemRecord({
        ...failedRecord,
        entity_id: "provider-document-id",
      }),
    ).toThrow();
    expect(() =>
      validateSystemRecord({
        ...failedRecord,
        error_classification: undefined,
      }),
    ).toThrow(
      "failed translation.job.terminal requires a catalog error_classification",
    );
    expect(() =>
      validateSystemRecord({
        ...failedRecord,
        error_classification: undefined,
        reason: "provider_error",
      }),
    ).toThrow("translation terminal event contains unsupported fields");
    const cancelledRecord = buildTranslationJobTerminalRecord(
      metadata,
      translationTerminal,
      "cancelled",
    );
    expect(cancelledRecord).toMatchObject({
      event: "translation.job.terminal",
      outcome: "cancelled",
    });
    expect(cancelledRecord).not.toHaveProperty("error_classification");
  });

  it("rejects unsupported system fields on translation terminals", () => {
    const record = buildTranslationJobTerminalRecord(
      metadata,
      translationTerminal,
      "failed",
      "internal",
    );
    const unsupportedFields: readonly Partial<SystemRecord>[] = [
      { component: "api" },
      { dependency: "postgres" },
      { operation: "read_queue" },
      { queue: "translation" },
      { message_id: "message-1" },
      { command_id: "command-1" },
      { retry_count: 1 },
      { job_kind: "translation" },
      { entity_id: "provider-document-id" },
      { record_class: "domain_audit" },
      { action: "post.updated" },
      { error_code: "provider_error" },
      { reason: "provider_error" },
    ];
    for (const fields of unsupportedFields) {
      expect(() => validateSystemRecord({ ...record, ...fields })).toThrow();
    }
  });

  it("restricts delivery requeue reasons", () => {
    for (const reason of ["shutdown", "handler_failed"] as const) {
      expect(() =>
        buildQueueDeliveryRequeuedRecord(metadata, delivery, reason),
      ).not.toThrow();
    }
    expect(() =>
      buildQueueDeliveryRequeuedRecord(metadata, delivery, "archive_failed"),
    ).toThrow();
    expect(() =>
      buildQueueRetryFailedRecord(metadata, handoff, "archive_failed"),
    ).toThrow();
  });

  it("restricts terminal collaboration checkpoint and client render catalogs", () => {
    expect(() =>
      buildCollaborationCheckpointFailedRecord(
        metadata,
        { entity_type: "post", entity_id: "post-1", retry_count: 4 },
        "persist_failed",
      ),
    ).not.toThrow();
    for (const entity_type of ["form", "email_layout"] as const) {
      expect(() =>
        buildCollaborationCheckpointFailedRecord(
          metadata,
          { entity_type, entity_id: "entity-1", retry_count: 4 },
          "persist_failed",
        ),
      ).not.toThrow();
    }
    expect(() =>
      buildCollaborationCheckpointFailedRecord(
        metadata,
        { entity_type: "post", entity_id: "", retry_count: 4 },
        "persist_failed",
      ),
    ).toThrow();
    expect(() =>
      buildClientRenderFailedRecord(metadata, "global"),
    ).not.toThrow();
    expect(() =>
      buildClientRenderFailedRecord(metadata, "other" as "global"),
    ).toThrow();
  });

  it("fixes the append record class in dedicated builders", () => {
    expect(
      buildSecurityAccessAppendFailedRecord(
        metadata,
        "authorization.denied",
        "persistence_failed",
      ).record_class,
    ).toBe("security_access");
  });

  it("uses the same severity policy as Go", () => {
    expect(systemLogLevel(buildServiceReadyRecord(metadata, "api"))).toBe(
      "info",
    );
    expect(
      systemLogLevel(buildServiceDegradedRecord(metadata, "api", failure)),
    ).toBe("warn");
    expect(
      systemLogLevel(buildServiceFailedRecord(metadata, "api", failure)),
    ).toBe("error");
    expect(
      systemLogLevel(
        buildTranslationJobTerminalRecord(
          metadata,
          translationTerminal,
          "failed",
          "internal",
        ),
      ),
    ).toBe("error");
  });
});
