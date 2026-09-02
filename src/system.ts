import {
  SYSTEM_EVENT_OUTCOMES,
  validateSystemRecord,
  type AuditAction,
  type AuditAppendFailureReason,
  type Correlation,
  type SecurityAction,
  type SystemEvent,
  type SystemRecord,
  type TranslationEntityType,
  type TranslationFailureReason,
  type TranslationJobTerminalOutcome,
} from "./records.ts";
import type { JobFailureReason, JobKind } from "./job.ts";

export type SystemMetadata = Correlation & {
  readonly occurred_at: string;
};

export interface SystemFailure {
  readonly error_code?: string;
  readonly reason?: string;
}

export type QueueFailureReason =
  | "enqueue_failed"
  | "handler_failed"
  | "completion_failed"
  | "shutdown"
  | "visibility_update_failed"
  | "archive_failed";

export interface QueuePublishContext {
  readonly queue: string;
  readonly message_id: string;
  readonly command_id: string;
  readonly duration_ms: number;
}

export interface QueueDeliveryContext {
  readonly queue: string;
  readonly message_id: string;
  readonly command_id: string;
  readonly retry_count: number;
  readonly duration_ms: number;
}

export interface QueueHandoffContext {
  readonly queue: string;
  readonly message_id: string;
  readonly command_id: string;
  readonly retry_count: number;
}

export interface JobContext {
  readonly job_kind: JobKind;
  readonly job_id: string;
}

export interface JobFailure {
  readonly reason: JobFailureReason;
}

/** Bounded public context for an authoritative Translation terminal transition. */
export interface TranslationJobTerminalContext {
  readonly job_id: string;
  readonly entity_type: TranslationEntityType;
  readonly target_locale: string;
  readonly duration_ms: number;
}

export type CollaborationCheckpointFailureReason =
  | "shared_document_unavailable"
  | "source_document_unavailable"
  | "locale_ownership_changed"
  | "shared_structure_changed"
  | "shared_materialization_changed"
  | "document_revision_changed"
  | "target_revision_changed"
  | "persist_failed";

export type CollaborationCheckpointEntityType =
  | "post"
  | "page"
  | "work"
  | "release"
  | "label"
  | "artist"
  | "campaign"
  | "program_event"
  | "map_theme"
  | "terms_history"
  | "privacy_history"
  | "form"
  | "email_layout";

export interface CollaborationCheckpointContext {
  readonly entity_type: CollaborationCheckpointEntityType;
  readonly entity_id: string;
  readonly retry_count: number;
}

export type ClientRenderComponent = "general" | "admin" | "global";

export type SystemLogLevel = "info" | "warn" | "error";

export function systemLogLevel(record: SystemRecord): SystemLogLevel {
  validateSystemRecord(record);
  if (record.outcome === "failed") return "error";
  if (record.outcome === "degraded" || record.outcome === "requeued") {
    return "warn";
  }
  return "info";
}

export const buildServiceReadyRecord = (
  metadata: SystemMetadata,
  component: string,
) => buildSystemRecord(metadata, "service.ready", { component });
export const buildServiceDegradedRecord = (
  metadata: SystemMetadata,
  component: string,
  failure: SystemFailure,
) => buildSystemRecord(metadata, "service.degraded", { component, ...failure });
export const buildServiceStoppingRecord = (
  metadata: SystemMetadata,
  component: string,
) => buildSystemRecord(metadata, "service.stopping", { component });
export const buildServiceFailedRecord = (
  metadata: SystemMetadata,
  component: string,
  failure: SystemFailure,
) => buildSystemRecord(metadata, "service.failed", { component, ...failure });
export const buildDependencyDegradedRecord = (
  metadata: SystemMetadata,
  dependency: string,
  operation: string,
  failure: SystemFailure,
) =>
  buildSystemRecord(metadata, "dependency.degraded", {
    dependency,
    operation,
    ...failure,
  });
export const buildDependencyRecoveredRecord = (
  metadata: SystemMetadata,
  dependency: string,
  operation: string,
) =>
  buildSystemRecord(metadata, "dependency.recovered", {
    dependency,
    operation,
  });
export const buildQueuePublishSucceededRecord = (
  metadata: SystemMetadata,
  queue: QueuePublishContext,
) =>
  buildSystemRecord(metadata, "queue.publish.succeeded", {
    domain: "queue",
    ...queue,
  });
export const buildQueuePublishFailedRecord = (
  metadata: SystemMetadata,
  queue: QueuePublishContext,
  reason: QueueFailureReason,
) =>
  buildSystemRecord(metadata, "queue.publish.failed", {
    domain: "queue",
    ...queue,
    reason,
  });
export const buildQueueDeliverySucceededRecord = (
  metadata: SystemMetadata,
  queue: QueueDeliveryContext,
) =>
  buildSystemRecord(metadata, "queue.delivery.succeeded", {
    domain: "queue",
    ...queue,
  });
export const buildQueueDeliveryFailedRecord = (
  metadata: SystemMetadata,
  queue: QueueDeliveryContext,
  reason: QueueFailureReason,
) =>
  buildSystemRecord(metadata, "queue.delivery.failed", {
    domain: "queue",
    ...queue,
    reason,
  });
export const buildQueueDeliveryRequeuedRecord = (
  metadata: SystemMetadata,
  queue: QueueDeliveryContext,
  reason: QueueFailureReason,
) =>
  buildSystemRecord(metadata, "queue.delivery.requeued", {
    domain: "queue",
    ...queue,
    reason,
  });
export const buildQueueRetryAcceptedRecord = (
  metadata: SystemMetadata,
  queue: QueueHandoffContext,
) =>
  buildSystemRecord(metadata, "queue.retry.accepted", {
    domain: "queue",
    ...queue,
  });
export const buildQueueRetryFailedRecord = (
  metadata: SystemMetadata,
  queue: QueueHandoffContext,
  reason: QueueFailureReason,
) =>
  buildSystemRecord(metadata, "queue.retry.failed", {
    domain: "queue",
    ...queue,
    reason,
  });
export const buildQueueDLQAcceptedRecord = (
  metadata: SystemMetadata,
  queue: QueueHandoffContext,
) =>
  buildSystemRecord(metadata, "queue.dlq.accepted", {
    domain: "queue",
    ...queue,
  });
export const buildQueueDLQFailedRecord = (
  metadata: SystemMetadata,
  queue: QueueHandoffContext,
  reason: QueueFailureReason,
) =>
  buildSystemRecord(metadata, "queue.dlq.failed", {
    domain: "queue",
    ...queue,
    reason,
  });
export const buildJobStartedRecord = (
  metadata: SystemMetadata,
  job: JobContext,
) => buildSystemRecord(metadata, "job.started", job);
export const buildJobSucceededRecord = (
  metadata: SystemMetadata,
  job: JobContext,
  durationMs: number,
) =>
  buildSystemRecord(metadata, "job.succeeded", {
    ...job,
    duration_ms: durationMs,
  });
export const buildJobFailedRecord = (
  metadata: SystemMetadata,
  job: JobContext,
  durationMs: number,
  failure: JobFailure,
) =>
  buildSystemRecord(metadata, "job.failed", {
    ...job,
    duration_ms: durationMs,
    ...failure,
  });
export const buildTranslationJobTerminalRecord = (
  metadata: SystemMetadata,
  job: TranslationJobTerminalContext,
  outcome: TranslationJobTerminalOutcome,
  errorClassification?: TranslationFailureReason,
) => {
  if (outcome === "failed") {
    if (errorClassification === undefined) {
      throw new TypeError(
        "failed translation.job.terminal requires an error classification",
      );
    }
  } else if (errorClassification !== undefined) {
    throw new TypeError(
      `${outcome} translation terminal events cannot contain an error classification`,
    );
  }
  return buildSystemRecord(metadata, "translation.job.terminal", {
    domain: "translation",
    ...job,
    outcome,
    ...(errorClassification === undefined
      ? {}
      : { error_classification: errorClassification }),
  });
};
export const buildCollaborationCheckpointFailedRecord = (
  metadata: SystemMetadata,
  checkpoint: CollaborationCheckpointContext,
  reason: CollaborationCheckpointFailureReason,
) =>
  buildSystemRecord(metadata, "collaboration.checkpoint.failed", {
    domain: "collaboration",
    ...checkpoint,
    reason,
  });
export const buildClientRenderFailedRecord = (
  metadata: SystemMetadata,
  component: ClientRenderComponent,
) =>
  buildSystemRecord(metadata, "client.render.failed", {
    domain: "client",
    component,
    reason: "react_error_boundary",
  });
export const buildDomainAuditAppendFailedRecord = (
  metadata: SystemMetadata,
  action: AuditAction,
  reason: AuditAppendFailureReason,
) =>
  buildSystemRecord(metadata, "audit.append.failed", {
    domain: "audit",
    record_class: "domain_audit",
    action,
    reason,
  });
export const buildSecurityAccessAppendFailedRecord = (
  metadata: SystemMetadata,
  action: SecurityAction,
  reason: AuditAppendFailureReason,
) =>
  buildSystemRecord(metadata, "audit.append.failed", {
    domain: "audit",
    record_class: "security_access",
    action,
    reason,
  });
export const buildTelemetryPipelineDegradedRecord = (
  metadata: SystemMetadata,
  component: string,
  failure: SystemFailure,
) =>
  buildSystemRecord(metadata, "telemetry.pipeline.degraded", {
    component,
    ...failure,
  });
export const buildTelemetryPipelineRecoveredRecord = (
  metadata: SystemMetadata,
  component: string,
) => buildSystemRecord(metadata, "telemetry.pipeline.recovered", { component });

function buildSystemRecord(
  metadata: SystemMetadata,
  event: SystemEvent,
  attributes: Partial<SystemRecord>,
): SystemRecord {
  const outcome = attributes.outcome ?? SYSTEM_EVENT_OUTCOMES[event]!;
  const record = {
    ...metadata,
    ...attributes,
    event,
    outcome,
  } as SystemRecord;
  validateSystemRecord(record);
  return record;
}
