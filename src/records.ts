import { parseServiceName, type RecordActor } from "./actor.ts";
import type { CollaborationCheckpointEntityType } from "./system.ts";
import { parseJobFailureReason, parseJobKind } from "./job.ts";
import { isCanonicalSourceIp, isRequestId } from "./request-context.ts";
import { classifyHTTPResult } from "./request-result.ts";
import {
  AUDIT_ACTIONS as auditActions,
  type AuditAction,
} from "./audit/catalog.ts";
import {
  auditActionAllowsAnonymous,
  systemActorMayAppendAudit,
} from "./audit/actor-policy.ts";
import { validateFileAuditRecord } from "./audit/file.ts";
import { validateArtistLabelAuditRecord } from "./audit/artist-label.ts";
import { validateIntegrationAuditRecord } from "./audit/integrations.ts";
import { validateReferenceEntityAuditRecord } from "./audit/reference-entities.ts";
import { validateReferenceDataAuditRecord } from "./audit/reference-data.ts";
import { validateSeriesTypeAuditRecord } from "./audit/series-type.ts";
import { validateProgramAuditRecord } from "./audit/program.ts";
import { validateReleaseCampaignAuditRecord } from "./audit/release-campaign.ts";
import { validateEmailAuthoringAuditRecord } from "./audit/email-authoring.ts";
import { validateSettingsAuditRecord } from "./audit/settings.ts";
import { validateFormAuditRecord } from "./audit/form.ts";
import { validateContentAuditRecord } from "./audit/content.ts";
import { validateMemberAccountAuditRecord } from "./audit/member-account.ts";
import { validateSourceLocaleAuditRecord } from "./audit/source-locale.ts";
import { validateLocaleContentAuditRecord } from "./audit/locale-content.ts";
import { assertKnownAuditRecordKeys } from "./audit/attributes.ts";

export interface Correlation {
  readonly request_id?: string;
  readonly trace_id?: string;
  readonly span_id?: string;
}

export type RequestOutcome = "succeeded" | "blocked" | "failed";

export type RequestRecord = Correlation &
  RecordActor & {
    readonly event: "request.completed";
    readonly occurred_at: string;
    readonly http_method?: string;
    readonly http_route?: string;
    readonly rpc_service?: string;
    readonly rpc_method?: string;
    readonly status_code: number;
    readonly duration_ms: number;
    readonly outcome: RequestOutcome;
    readonly error_code?: string;
    readonly reason?: string;
  };

export const SYSTEM_EVENTS = [
  "service.ready",
  "service.degraded",
  "service.stopping",
  "service.failed",
  "dependency.degraded",
  "dependency.recovered",
  "queue.publish.succeeded",
  "queue.publish.failed",
  "queue.delivery.succeeded",
  "queue.delivery.failed",
  "queue.delivery.requeued",
  "queue.retry.accepted",
  "queue.retry.failed",
  "queue.dlq.accepted",
  "queue.dlq.failed",
  "job.started",
  "job.succeeded",
  "job.failed",
  "audit.append.failed",
  "collaboration.checkpoint.failed",
  "client.render.failed",
  "telemetry.pipeline.degraded",
  "telemetry.pipeline.recovered",
  "translation.job.terminal",
] as const;

export const SYSTEM_EVENT_OUTCOMES: Partial<Record<SystemEvent, string>> = {
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
};

export type SystemEvent = (typeof SYSTEM_EVENTS)[number];

export const TRANSLATION_JOB_TERMINAL_OUTCOMES = [
  "applied",
  "failed",
  "cancelled",
] as const;
export type TranslationJobTerminalOutcome =
  (typeof TRANSLATION_JOB_TERMINAL_OUTCOMES)[number];

export const TRANSLATION_ENTITY_TYPES = [
  "post",
  "page",
  "work",
  "program_event",
  "release",
  "artist",
  "label",
  "menu",
  "email_template",
  "email_layout",
  "campaign",
  "form",
  "privacy",
  "terms",
  "series",
] as const;
export type TranslationEntityType = (typeof TRANSLATION_ENTITY_TYPES)[number];

export const TRANSLATION_FAILURE_REASONS = [
  "provider_configuration",
  "provider_authentication",
  "provider_rate_limited",
  "provider_unavailable",
  "provider_rejected",
  "provider_response_invalid",
  "target_apply_failed",
  "og_handoff_failed",
  "internal",
] as const;
export type TranslationFailureReason =
  (typeof TRANSLATION_FAILURE_REASONS)[number];

export const COLLABORATION_CHECKPOINT_ENTITY_TYPES = [
  "post",
  "page",
  "work",
  "release",
  "label",
  "artist",
  "campaign",
  "program_event",
  "map_theme",
  "terms_history",
  "privacy_history",
  "form",
  "email_layout",
] as const satisfies readonly CollaborationCheckpointEntityType[];

const COLLABORATION_CONFLICT_REASONS = [
  "locale_ownership_changed",
  "shared_structure_changed",
  "shared_materialization_changed",
  "document_revision_changed",
  "target_revision_changed",
];

export interface SystemRecord extends Correlation {
  readonly event: SystemEvent;
  readonly occurred_at: string;
  readonly component?: string;
  readonly dependency?: string;
  readonly operation?: string;
  readonly domain?: string;
  readonly queue?: string;
  readonly message_id?: string;
  readonly command_id?: string;
  readonly retry_count?: number;
  readonly duration_ms?: number;
  readonly job_kind?: string;
  readonly job_id?: string;
  readonly entity_type?: string;
  readonly entity_id?: string;
  readonly target_locale?: string;
  readonly record_class?: AuditRecordClass;
  readonly action?: string;
  readonly outcome?: string;
  readonly error_code?: string;
  readonly reason?: string;
  readonly error_classification?: string;
}

export const AUDIT_ACTIONS = auditActions;
export type { AuditAction } from "./audit/catalog.ts";

export type AuditRecordClass = "domain_audit" | "security_access";
export type AccountSessionScope = "current" | "one" | "others";
export type AuditCollectionOperation = "added" | "removed";
export type AuditItemOperation = "created" | "updated" | "deleted";
export type AuditItemScope = "form" | "dashboard";
export type AuditAssetSlot = "light" | "dark";
export type AuditRelationship =
  "none" | "author" | "collaborator" | "owner" | "manager";
export type AuditState =
  | "none"
  | "active"
  | "banned"
  | "confirmation_pending"
  | "scheduled"
  | "recovery_confirmation_pending"
  | "cancelled"
  | "recovered"
  | "released"
  | "disabled"
  | "public"
  | "authenticated"
  | "restricted"
  | "subscribed"
  | "unsubscribed"
  | "draft"
  | "published"
  | "archived"
  | "sending"
  | "sent"
  | "failed";
export type AuditAppendFailureReason =
  | "record_invalid"
  | "transaction_missing"
  | "persistence_failed"
  | "database_missing"
  | "request_context_missing"
  | "actor_invalid"
  | "record_build_failed";

export interface AuditRecord extends Correlation {
  readonly audit_id: string;
  readonly occurred_at: string;
  readonly action: AuditAction;
  readonly target_type: string;
  readonly target_id: string;
  readonly nickname?: string;
  readonly asset_id?: string;
  readonly consent_id?: string;
  readonly changed_fields?: readonly string[];
  readonly collection_operation?: AuditCollectionOperation;
  readonly asset_slot?: AuditAssetSlot;
  readonly previous_state?: AuditState;
  readonly new_state?: AuditState;
  readonly scheduled_at?: string;
  readonly scheduled_time_zone?: string;
  readonly version_id?: string;
  readonly contributor_member_ids?: readonly string[];
  readonly previous_role?: string;
  readonly new_role?: string;
  readonly email?: string;
  readonly previous_email?: string;
  readonly new_email?: string;
  readonly provider?: string;
  readonly provider_subject?: string;
  readonly passkey_ids?: readonly string[];
  readonly session_scope?: AccountSessionScope;
  readonly session_ids?: readonly string[];
  readonly preferred_locale?: string;
  readonly locale?: string;
  readonly previous_locale?: string;
  readonly new_locale?: string;
  /** Present empty arrays intentionally represent an empty member-tag set. */
  readonly tag_ids?: readonly string[];
  readonly tag_name?: string;
  readonly event_name?: string;
  readonly policy_type?: "terms" | "privacy";
  readonly version_number?: number;
  readonly effective_at?: string;
  readonly item_operation?: AuditItemOperation;
  readonly item_id?: string;
  readonly item_scope?: AuditItemScope;
  /** Parent Form for a submission; submission payload is never audited. */
  readonly parent_id?: string;
  /** Empty denotes a deliberately cleared relation. */
  readonly previous_item_id?: string;
  readonly file_id?: string;
  /** Present arrays preserve an Artist gallery's ordered file binding. */
  readonly file_ids?: readonly string[];
  readonly subject_member_id?: string;
  readonly subject_post_id?: string;
  readonly previous_relationship?: AuditRelationship;
  readonly new_relationship?: AuditRelationship;
  readonly previous_series_id?: string;
  readonly new_series_id?: string;
  /** Present arrays preserve the exact persisted Post Series order, including empty. */
  readonly post_ids?: readonly string[];
  /** File/folder move boundary; empty string denotes the root folder. */
  readonly previous_parent_id?: string;
  /** File/folder move boundary; empty string denotes the root folder. */
  readonly new_parent_id?: string;
  /** Deliberately present empty arrays represent an empty previous audience. */
  readonly previous_item_ids?: readonly string[];
  /** Deliberately present empty arrays represent an empty new audience. */
  readonly item_ids?: readonly string[];
  readonly actor_kind: "anonymous" | "member" | "system";
  readonly actor_member_id?: string;
  readonly actor_service?: string;
}

export type SecurityAction =
  | "authentication.succeeded"
  | "authentication.failed"
  | "authentication.blocked"
  | "authorization.denied"
  | "personal_data.accessed";
export type AuthenticationFlowKind =
  "login" | "registration" | "reauthentication";
export type AuthenticationMethod = "email_code" | "oidc" | "passkey";
export type AuthenticationPrincipalState = "onboarding_only" | "active";
export type AuthenticationFailureReason =
  | "proof_rejected"
  | "account_blocked"
  | "provider_denied"
  | "provider_failed"
  | "member_link_invalid"
  | "internal_error";
export type AuthenticationBlockReason =
  | "flow_invalid"
  | "request_invalid"
  | "integrity_check_failed"
  | "rate_limited"
  | "service_unavailable";
export type AuthorizationDenialReason =
  "authentication_required" | "permission_denied";
export type PersonalDataAccessKind = "read";

export interface SecurityAccessRecord extends Correlation {
  readonly access_id: string;
  readonly occurred_at: string;
  readonly action: SecurityAction;
  readonly actor_kind: "anonymous" | "member" | "system";
  readonly actor_member_id?: string;
  readonly actor_service?: string;
  readonly source_ip: string;
  readonly flow_kind?: AuthenticationFlowKind;
  readonly authentication_method?: AuthenticationMethod;
  readonly principal_state?: AuthenticationPrincipalState;
  readonly provider?: string;
  readonly reason?: string;
  readonly attempted_action?: string;
  readonly permission?: string;
  readonly subject_type?: string;
  readonly subject_id?: string;
  readonly access_kind?: PersonalDataAccessKind;
  readonly data_category?: string;
}

const authenticationFailureReasons: readonly AuthenticationFailureReason[] = [
  "proof_rejected",
  "account_blocked",
  "provider_denied",
  "provider_failed",
  "member_link_invalid",
  "internal_error",
];
const authenticationBlockReasons: readonly AuthenticationBlockReason[] = [
  "flow_invalid",
  "request_invalid",
  "integrity_check_failed",
  "rate_limited",
  "service_unavailable",
];
const authorizationDenialReasons: readonly AuthorizationDenialReason[] = [
  "authentication_required",
  "permission_denied",
];

export interface TelemetrySink {
  write(record: RequestRecord | SystemRecord): void | Promise<void>;
}

export function validateRequestRecord(record: RequestRecord): void {
  validateOccurredAt(record.occurred_at);
  if (record.request_id === undefined || !isRequestId(record.request_id)) {
    throw new TypeError(
      "request record requires a canonical UUIDv4 request_id",
    );
  }
  validateTraceCorrelation(record);
  validateRecordActor(record);
  const hasHttp =
    record.http_method !== undefined &&
    record.http_method !== "" &&
    record.http_route !== undefined &&
    record.http_route !== "" &&
    record.rpc_service === undefined &&
    record.rpc_method === undefined;
  const hasRpc =
    record.rpc_service !== undefined &&
    record.rpc_service !== "" &&
    record.rpc_method !== undefined &&
    record.rpc_method !== "" &&
    record.http_route === undefined;
  if (!hasHttp && !hasRpc) {
    throw new TypeError(
      "request record requires exactly one HTTP route or RPC service and method",
    );
  }
  if (
    !Number.isInteger(record.status_code) ||
    record.status_code < 100 ||
    record.status_code > 599
  ) {
    throw new TypeError("status_code must be an HTTP status code");
  }
  if (!Number.isInteger(record.duration_ms) || record.duration_ms < 0) {
    throw new TypeError("duration_ms must be a non-negative integer");
  }
  if (!["succeeded", "blocked", "failed"].includes(record.outcome)) {
    throw new TypeError(`invalid request outcome: ${record.outcome}`);
  }
  if (
    record.outcome === "succeeded" &&
    (record.error_code !== undefined || record.reason !== undefined)
  ) {
    throw new TypeError(
      "successful request cannot contain error_code or reason",
    );
  }
  if (
    record.outcome !== "succeeded" &&
    record.error_code === undefined &&
    record.reason === undefined
  ) {
    throw new TypeError(
      "blocked or failed request requires error_code or reason",
    );
  }
  if (hasHttp) {
    const expected = classifyHTTPResult(record.status_code, record.duration_ms);
    if (
      record.outcome !== expected.outcome ||
      record.reason !== expected.reason
    ) {
      throw new TypeError(
        "HTTP request outcome and reason must match the shared status classifier",
      );
    }
  }
}

export function validateSystemRecord(record: SystemRecord): void {
  validateOccurredAt(record.occurred_at);
  if (!(SYSTEM_EVENTS as readonly string[]).includes(record.event)) {
    throw new TypeError(`unknown system event: ${record.event}`);
  }
  if (record.request_id !== undefined && !isRequestId(record.request_id)) {
    throw new TypeError("request_id must be a canonical UUIDv4");
  }
  validateTraceCorrelation(record);
  if (record.event === "translation.job.terminal") {
    if (
      !TRANSLATION_JOB_TERMINAL_OUTCOMES.includes(
        record.outcome as TranslationJobTerminalOutcome,
      )
    ) {
      throw new TypeError(
        "translation.job.terminal requires a catalog outcome",
      );
    }
  } else if (record.outcome !== SYSTEM_EVENT_OUTCOMES[record.event]) {
    throw new TypeError(
      `system event ${record.event} requires outcome ${SYSTEM_EVENT_OUTCOMES[record.event]}`,
    );
  }
  if (
    record.event !== "translation.job.terminal" &&
    record.error_classification !== undefined
  ) {
    throw new TypeError(
      `system event ${record.event} does not allow error_classification`,
    );
  }
  for (const [name, value] of Object.entries({
    component: record.component,
    dependency: record.dependency,
    operation: record.operation,
    domain: record.domain,
    job_kind: record.job_kind,
    entity_type: record.entity_type,
    record_class: record.record_class,
    outcome: record.outcome,
    error_code: record.error_code,
    reason: record.reason,
    error_classification: record.error_classification,
  })) {
    if (value !== undefined && !isBoundedCode(value)) {
      throw new TypeError(`system field ${name} must be a bounded code`);
    }
  }
  if (
    record.entity_id !== undefined &&
    (record.entity_id.length === 0 || record.entity_id.length > 128)
  ) {
    throw new TypeError("entity_id must be a non-empty bounded identifier");
  }
  if (
    record.target_locale !== undefined &&
    !isBoundedLocaleCode(record.target_locale)
  ) {
    throw new TypeError("target_locale must be a bounded locale code");
  }
  for (const [name, value] of [
    ["retry_count", record.retry_count],
    ["duration_ms", record.duration_ms],
  ] as const) {
    if (value !== undefined && (!Number.isInteger(value) || value < 0)) {
      throw new TypeError(`${name} must be a non-negative integer`);
    }
  }
  validateSystemRequiredFields(record);
}

export function validateAuditRecord(record: AuditRecord): void {
  assertKnownAuditRecordKeys(record);
  validateOccurredAt(record.occurred_at);
  if (!isRequestId(record.audit_id)) {
    throw new TypeError("audit_id must be a canonical UUIDv4");
  }
  if (record.request_id !== undefined && !isRequestId(record.request_id)) {
    throw new TypeError("request_id must be a canonical UUIDv4");
  }
  validateTraceCorrelation(record);
  validateRecordActor(record as RecordActor);
  if (
    (record as RecordActor).actor_kind === "anonymous" &&
    !auditActionAllowsAnonymous(record.action)
  ) {
    throw new TypeError("domain audit actor cannot be anonymous");
  }
  if (record.actor_kind === "system" && !systemActorMayAppendAudit(record)) {
    throw new TypeError(
      `audit action ${record.action} cannot use system actor`,
    );
  }
  if (validateLocaleContentAuditRecord(record)) return;
  if (validateSourceLocaleAuditRecord(record)) return;
  if (validateFileAuditRecord(record)) return;
  if (validateArtistLabelAuditRecord(record)) return;
  if (validateSettingsAuditRecord(record)) return;
  if (validateIntegrationAuditRecord(record)) return;
  if (validateReferenceDataAuditRecord(record)) return;
  if (validateReferenceEntityAuditRecord(record)) return;
  if (validateSeriesTypeAuditRecord(record)) return;
  if (validateProgramAuditRecord(record)) return;
  if (validateReleaseCampaignAuditRecord(record)) return;
  if (validateEmailAuthoringAuditRecord(record)) return;
  if (validateFormAuditRecord(record)) return;
  if (validateContentAuditRecord(record)) return;
  if (validateMemberAccountAuditRecord(record)) return;
  throw new TypeError(`unsupported audit action ${String(record.action)}`);
}

export function validateSecurityAccessRecord(
  record: SecurityAccessRecord,
): void {
  validateOccurredAt(record.occurred_at);
  if (
    !isRequestId(record.access_id) ||
    record.request_id === undefined ||
    !isRequestId(record.request_id)
  ) {
    throw new TypeError(
      "security access record requires canonical UUIDv4 access_id and request_id",
    );
  }
  validateTraceCorrelation(record);
  validateRecordActor(record as RecordActor);
  if (!isCanonicalSourceIp(record.source_ip)) {
    throw new TypeError("source_ip must be a canonical IPv4 or IPv6 address");
  }
  switch (record.action) {
    case "authentication.succeeded": {
      requireOnlyAuthenticationAttributes(record);
      if (record.actor_kind !== "member") {
        throw new TypeError(
          "successful authentication requires a member actor",
        );
      }
      if (
        !isAuthenticationFlow(record.flow_kind) ||
        !isAuthenticationMethod(record.authentication_method)
      ) {
        throw new TypeError(
          "successful authentication requires flow_kind and authentication_method",
        );
      }
      if (!isAuthenticationPrincipalState(record.principal_state)) {
        throw new TypeError(
          "successful authentication requires principal_state",
        );
      }
      if (record.reason !== undefined) {
        throw new TypeError("successful authentication cannot contain reason");
      }
      validateAuthenticationProvider(record);
      return;
    }
    case "authentication.failed": {
      requireOnlyAuthenticationAttributes(record);
      if (record.principal_state !== undefined) {
        throw new TypeError(
          "authentication failure cannot contain principal_state",
        );
      }
      if (record.actor_kind === "system") {
        throw new TypeError("authentication failure cannot use a system actor");
      }
      if (
        !isAuthenticationFlow(record.flow_kind) ||
        !isAuthenticationMethod(record.authentication_method)
      ) {
        throw new TypeError(
          "authentication failure requires flow_kind and authentication_method",
        );
      }
      if (
        record.actor_kind === "member" &&
        record.flow_kind !== "reauthentication"
      ) {
        throw new TypeError(
          "only reauthentication failure or block can use a member actor",
        );
      }
      if (
        !authenticationFailureReasons.includes(
          record.reason as AuthenticationFailureReason,
        )
      ) {
        throw new TypeError(
          `invalid authentication failure reason: ${record.reason}`,
        );
      }
      validateAuthenticationProvider(record);
      return;
    }
    case "authentication.blocked": {
      requireOnlyAuthenticationAttributes(record);
      if (record.principal_state !== undefined) {
        throw new TypeError(
          "authentication block cannot contain principal_state",
        );
      }
      if (record.actor_kind === "system") {
        throw new TypeError("authentication block cannot use a system actor");
      }
      if (
        record.flow_kind !== undefined &&
        !isAuthenticationFlow(record.flow_kind)
      ) {
        throw new TypeError(`invalid authentication flow_kind`);
      }
      if (
        record.authentication_method !== undefined &&
        !isAuthenticationMethod(record.authentication_method)
      ) {
        throw new TypeError(`invalid authentication_method`);
      }
      if (
        record.actor_kind === "member" &&
        record.flow_kind !== "reauthentication"
      ) {
        throw new TypeError(
          "only reauthentication failure or block can use a member actor",
        );
      }
      if (
        !authenticationBlockReasons.includes(
          record.reason as AuthenticationBlockReason,
        )
      ) {
        throw new TypeError(
          `invalid authentication block reason: ${record.reason}`,
        );
      }
      validateAuthenticationProvider(record);
      return;
    }
    case "authorization.denied": {
      requireOnlyAuthorizationAttributes(record);
      if (
        record.actor_kind === "system" ||
        !isValidAuthorizationScope(
          record.attempted_action ?? "",
          record.permission ?? "",
        )
      ) {
        throw new TypeError(
          "authorization denial requires a cataloged attempted action and permission",
        );
      }
      if (
        !authorizationDenialReasons.includes(
          record.reason as AuthorizationDenialReason,
        )
      ) {
        throw new TypeError(
          `invalid authorization denial reason: ${record.reason}`,
        );
      }
      return;
    }
    case "personal_data.accessed": {
      requireOnlyPersonalDataAttributes(record);
      if (
        record.actor_kind !== "member" ||
        record.access_kind !== "read" ||
        !isValidPersonalDataSubject(
          record.subject_type ?? "",
          record.subject_id ?? "",
          record.data_category ?? "",
        )
      ) {
        throw new TypeError(
          "personal data access requires a cataloged member read scope",
        );
      }
      return;
    }
    default:
      throw new TypeError(`unknown security action: ${String(record.action)}`);
  }
}

function isValidAuthorizationScope(
  attemptedAction: string,
  permission: string,
): boolean {
  return (
    isConnectProcedure(attemptedAction) && permission === "procedure:invoke"
  );
}

function validateAuthenticationProvider(record: SecurityAccessRecord): void {
  if (record.provider === undefined) {
    return;
  }
  if (
    record.authentication_method !== "oidc" ||
    !isBoundedCode(record.provider)
  ) {
    throw new TypeError(
      "provider is only valid as a bounded code for oidc authentication",
    );
  }
}

function requireOnlyAuthenticationAttributes(
  record: SecurityAccessRecord,
): void {
  if (
    record.attempted_action !== undefined ||
    record.permission !== undefined ||
    record.subject_type !== undefined ||
    record.subject_id !== undefined ||
    record.access_kind !== undefined ||
    record.data_category !== undefined
  ) {
    throw new TypeError(
      "authentication access contains attributes for another security action",
    );
  }
}

function requireOnlyAuthorizationAttributes(
  record: SecurityAccessRecord,
): void {
  if (
    record.flow_kind !== undefined ||
    record.authentication_method !== undefined ||
    record.principal_state !== undefined ||
    record.provider !== undefined ||
    record.subject_type !== undefined ||
    record.subject_id !== undefined ||
    record.access_kind !== undefined ||
    record.data_category !== undefined
  ) {
    throw new TypeError(
      "authorization denial contains attributes for another security action",
    );
  }
}

function requireOnlyPersonalDataAttributes(record: SecurityAccessRecord): void {
  if (
    record.flow_kind !== undefined ||
    record.authentication_method !== undefined ||
    record.principal_state !== undefined ||
    record.provider !== undefined ||
    record.reason !== undefined ||
    record.attempted_action !== undefined ||
    record.permission !== undefined
  ) {
    throw new TypeError(
      "personal data access contains attributes for another security action",
    );
  }
}

function isValidPersonalDataSubject(
  subjectType: string,
  subjectId: string,
  dataCategory: string,
): boolean {
  if (subjectType === "member_collection") {
    return subjectId === "1" && dataCategory === "member_administration";
  }
  if (!isCanonicalUuid(subjectId)) {
    return false;
  }
  switch (subjectType) {
    case "member":
      return dataCategory === "member_administration";
    case "campaign":
      return dataCategory === "campaign_recipients";
    case "form":
      return dataCategory === "form_submissions";
    case "form_submission":
      return dataCategory === "form_submission";
    default:
      return false;
  }
}

function isCanonicalUuid(value: string): boolean {
  return /^[0-9a-f]{8}-[0-9a-f]{4}-[1-8][0-9a-f]{3}-[89ab][0-9a-f]{3}-[0-9a-f]{12}$/.test(
    value,
  );
}

function isConnectProcedure(value: string): boolean {
  return (
    value.length >= 4 &&
    value.length <= 128 &&
    /^\/[A-Za-z0-9_.]+\/[A-Za-z0-9_]+$/.test(value)
  );
}

export async function emitRequest(
  sink: TelemetrySink,
  record: RequestRecord,
): Promise<void> {
  validateRequestRecord(record);
  await sink.write(record);
}

export async function emitSystem(
  sink: TelemetrySink,
  record: SystemRecord,
): Promise<void> {
  validateSystemRecord(record);
  await sink.write(record);
}

function validateOccurredAt(value: string): void {
  if (
    !/^\d{4}-\d{2}-\d{2}T\d{2}:\d{2}:\d{2}(?:\.\d+)?Z$/.test(value) ||
    Number.isNaN(Date.parse(value))
  ) {
    throw new TypeError("occurred_at must be a UTC RFC3339 timestamp");
  }
}

function validateTraceCorrelation(record: Correlation): void {
  if (record.trace_id === undefined && record.span_id === undefined) {
    return;
  }
  if (
    record.trace_id === undefined ||
    record.span_id === undefined ||
    !/^[0-9a-f]{32}$/.test(record.trace_id) ||
    /^0+$/.test(record.trace_id) ||
    !/^[0-9a-f]{16}$/.test(record.span_id) ||
    /^0+$/.test(record.span_id)
  ) {
    throw new TypeError(
      "trace_id and span_id must be valid and provided together",
    );
  }
}

function validateRecordActor(
  record: Readonly<{
    actor_kind: string;
    actor_member_id?: string;
    actor_service?: string;
  }>,
): void {
  switch (record.actor_kind) {
    case "anonymous":
      if (
        record.actor_member_id !== undefined ||
        record.actor_service !== undefined
      ) {
        throw new TypeError(
          "anonymous record actor cannot contain member or service identity",
        );
      }
      return;
    case "member":
      if (
        record.actor_member_id === undefined ||
        record.actor_member_id === "" ||
        record.actor_service !== undefined
      ) {
        throw new TypeError(
          "member record actor requires only actor_member_id",
        );
      }
      return;
    case "system":
      if (
        record.actor_service === undefined ||
        record.actor_service === "" ||
        record.actor_member_id !== undefined
      ) {
        throw new TypeError("system record actor requires only actor_service");
      }
      parseServiceName(record.actor_service);
      return;
    default:
      throw new TypeError(`invalid actor kind: ${record.actor_kind}`);
  }
}

function validateSystemRequiredFields(record: SystemRecord): void {
  const requireStrings = (...values: (string | undefined)[]): void => {
    if (values.some((value) => value === undefined || value === "")) {
      throw new TypeError(
        `system event ${record.event} is missing a required field`,
      );
    }
  };
  const requireFailure = (): void => {
    if (record.error_code === undefined && record.reason === undefined) {
      throw new TypeError(
        `system event ${record.event} requires error_code or reason`,
      );
    }
  };
  switch (record.event) {
    case "service.ready":
    case "service.stopping":
      requireStrings(record.component);
      return;
    case "service.degraded":
    case "service.failed":
      requireStrings(record.component);
      requireFailure();
      return;
    case "dependency.recovered":
      requireStrings(record.dependency, record.operation);
      return;
    case "dependency.degraded":
      requireStrings(record.dependency, record.operation);
      requireFailure();
      return;
    case "queue.publish.succeeded":
    case "queue.publish.failed":
      requireStrings(record.queue, record.message_id, record.command_id);
      if (record.duration_ms === undefined) {
        throw new TypeError(
          `system event ${record.event} requires duration_ms`,
        );
      }
      if (record.event === "queue.publish.failed") requireFailure();
      if (record.event === "queue.publish.succeeded") {
        requireNoFailure(record);
      } else {
        requireQueueReason(record, ["enqueue_failed"]);
      }
      return;
    case "queue.delivery.succeeded":
    case "queue.delivery.failed":
    case "queue.delivery.requeued":
      requireStrings(record.queue, record.message_id, record.command_id);
      if (
        record.retry_count === undefined ||
        record.duration_ms === undefined
      ) {
        throw new TypeError(
          `system event ${record.event} requires retry_count and duration_ms`,
        );
      }
      if (record.event === "queue.delivery.succeeded") {
        requireNoFailure(record);
      } else if (record.event === "queue.delivery.requeued") {
        requireQueueReason(record, ["shutdown", "handler_failed"]);
      } else {
        requireFailure();
        requireQueueReason(record, ["handler_failed", "completion_failed"]);
      }
      return;
    case "queue.retry.accepted":
    case "queue.retry.failed":
    case "queue.dlq.accepted":
    case "queue.dlq.failed":
      requireStrings(record.queue, record.message_id, record.command_id);
      if (record.retry_count === undefined) {
        throw new TypeError(
          `system event ${record.event} requires retry_count`,
        );
      }
      if (record.event.endsWith(".failed")) {
        requireFailure();
        requireQueueReason(record, [
          record.event === "queue.retry.failed"
            ? "visibility_update_failed"
            : "archive_failed",
        ]);
      } else {
        requireNoFailure(record);
      }
      return;
    case "job.started":
      requireStrings(record.job_kind, record.job_id);
      parseJobKind(record.job_kind!);
      return;
    case "job.succeeded":
    case "job.failed":
      requireStrings(record.job_kind, record.job_id);
      const jobKind = parseJobKind(record.job_kind!);
      if (record.duration_ms === undefined) {
        throw new TypeError(
          `system event ${record.event} requires duration_ms`,
        );
      }
      if (record.event === "job.failed") {
        if (record.error_code !== undefined || record.reason === undefined) {
          throw new TypeError(
            `system event ${record.event} requires a catalog job failure reason`,
          );
        }
        parseJobFailureReason(jobKind, record.reason);
      }
      return;
    case "audit.append.failed":
      requireStrings(record.record_class, record.action);
      if (!isKnownAppendAction(record.record_class!, record.action!)) {
        throw new TypeError(
          "audit.append.failed requires a catalog action for record_class",
        );
      }
      if (
        record.error_code !== undefined ||
        ![
          "record_invalid",
          "transaction_missing",
          "persistence_failed",
          "database_missing",
          "request_context_missing",
          "actor_invalid",
          "record_build_failed",
        ].includes(record.reason ?? "")
      ) {
        throw new TypeError(
          "audit.append.failed requires a catalog audit append reason",
        );
      }
      return;
    case "collaboration.checkpoint.failed":
      requireStrings(record.domain, record.entity_type, record.entity_id);
      if (
        record.domain !== "collaboration" ||
        record.retry_count === undefined ||
        record.error_code !== undefined ||
        !COLLABORATION_CHECKPOINT_ENTITY_TYPES.some(
          (entityType) => entityType === record.entity_type,
        ) ||
        ![
          "shared_document_unavailable",
          "source_document_unavailable",
          ...COLLABORATION_CONFLICT_REASONS,
          "persist_failed",
        ].includes(record.reason ?? "") ||
        (COLLABORATION_CONFLICT_REASONS.includes(record.reason ?? "") &&
          record.retry_count !== 1)
      ) {
        throw new TypeError(
          "collaboration.checkpoint.failed requires a terminal checkpoint context and catalog reason",
        );
      }
      return;
    case "client.render.failed":
      requireStrings(record.domain, record.component);
      if (
        record.domain !== "client" ||
        !["general", "admin", "global"].includes(record.component ?? "") ||
        record.error_code !== undefined ||
        record.reason !== "react_error_boundary"
      ) {
        throw new TypeError(
          "client.render.failed requires a catalog component and reason",
        );
      }
      return;
    case "translation.job.terminal":
      requireStrings(
        record.domain,
        record.job_id,
        record.entity_type,
        record.target_locale,
      );
      requireTranslationTerminalOnlyFields(record);
      if (
        record.domain !== "translation" ||
        !isCanonicalUuid(record.job_id!) ||
        !TRANSLATION_ENTITY_TYPES.includes(
          record.entity_type as TranslationEntityType,
        ) ||
        !isBoundedLocaleCode(record.target_locale!) ||
        record.duration_ms === undefined
      ) {
        throw new TypeError(
          "translation terminal event requires a canonical job and locale context",
        );
      }
      if (record.outcome !== "failed") {
        if (record.error_classification !== undefined) {
          throw new TypeError(
            "non-failed translation.job.terminal cannot contain error_classification",
          );
        }
        requireNoFailure(record);
        return;
      }
      if (
        !TRANSLATION_FAILURE_REASONS.includes(
          record.error_classification as TranslationFailureReason,
        )
      ) {
        throw new TypeError(
          "failed translation.job.terminal requires a catalog error_classification",
        );
      }
      return;
    case "telemetry.pipeline.recovered":
      requireStrings(record.component);
      return;
    case "telemetry.pipeline.degraded":
      requireStrings(record.component);
      requireFailure();
  }
}

function requireNoFailure(record: SystemRecord): void {
  if (record.error_code !== undefined || record.reason !== undefined) {
    throw new TypeError(
      `system event ${record.event} does not allow failure fields`,
    );
  }
}

function requireTranslationTerminalOnlyFields(record: SystemRecord): void {
  if (
    record.component !== undefined ||
    record.dependency !== undefined ||
    record.operation !== undefined ||
    record.queue !== undefined ||
    record.message_id !== undefined ||
    record.command_id !== undefined ||
    record.retry_count !== undefined ||
    record.job_kind !== undefined ||
    record.entity_id !== undefined ||
    record.record_class !== undefined ||
    record.action !== undefined ||
    record.error_code !== undefined ||
    record.reason !== undefined
  ) {
    throw new TypeError(
      "translation terminal event contains unsupported fields",
    );
  }
}

function requireQueueReason(
  record: SystemRecord,
  allowed: readonly string[],
): void {
  if (record.error_code !== undefined || !allowed.includes(record.reason!)) {
    throw new TypeError(
      `system event ${record.event} requires a catalog queue reason`,
    );
  }
}

function isKnownAppendAction(recordClass: string, action: string): boolean {
  if (recordClass === "domain_audit") {
    return (AUDIT_ACTIONS as readonly string[]).includes(action);
  }
  if (recordClass === "security_access") {
    return [
      "authentication.succeeded",
      "authentication.failed",
      "authentication.blocked",
      "authorization.denied",
      "personal_data.accessed",
    ].includes(action);
  }
  return false;
}

function isBoundedCode(value: string): boolean {
  return /^[a-z0-9_]{1,64}$/.test(value);
}

function isBoundedLocaleCode(value: string): boolean {
  return (
    value.length <= 64 && /^[A-Za-z0-9]{1,8}(-[A-Za-z0-9]{1,8})*$/.test(value)
  );
}

function isAuthenticationFlow(
  value: AuthenticationFlowKind | undefined,
): value is AuthenticationFlowKind {
  return (
    value === "login" ||
    value === "registration" ||
    value === "reauthentication"
  );
}

function isAuthenticationMethod(
  value: AuthenticationMethod | undefined,
): value is AuthenticationMethod {
  return value === "email_code" || value === "oidc" || value === "passkey";
}

function isAuthenticationPrincipalState(
  value: AuthenticationPrincipalState | undefined,
): value is AuthenticationPrincipalState {
  return value === "onboarding_only" || value === "active";
}
