package telemetry

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/google/uuid"
	"go.opentelemetry.io/otel/trace"
)

type RequestOutcome string

const (
	RequestOutcomeSucceeded RequestOutcome = "succeeded"
	RequestOutcomeBlocked   RequestOutcome = "blocked"
	RequestOutcomeFailed    RequestOutcome = "failed"
)

type Correlation struct {
	RequestID string `json:"request_id,omitempty"`
	TraceID   string `json:"trace_id,omitempty"`
	SpanID    string `json:"span_id,omitempty"`
}

func CorrelationFromContext(ctx context.Context) Correlation {
	correlation := Correlation{RequestID: RequestIDFromContext(ctx)}
	spanContext := trace.SpanContextFromContext(ctx)
	if spanContext.IsValid() {
		correlation.TraceID = spanContext.TraceID().String()
		correlation.SpanID = spanContext.SpanID().String()
	}
	return correlation
}

type RequestRecord struct {
	Event      string    `json:"event"`
	OccurredAt time.Time `json:"occurred_at"`
	Correlation
	RecordActor
	HTTPMethod string         `json:"http_method,omitempty"`
	HTTPRoute  string         `json:"http_route,omitempty"`
	RPCService string         `json:"rpc_service,omitempty"`
	RPCMethod  string         `json:"rpc_method,omitempty"`
	StatusCode int            `json:"status_code"`
	DurationMS int64          `json:"duration_ms"`
	Outcome    RequestOutcome `json:"outcome"`
	ErrorCode  string         `json:"error_code,omitempty"`
	Reason     string         `json:"reason,omitempty"`
}

func (record RequestRecord) Validate() error {
	if record.Event != "request.completed" {
		return fmt.Errorf("request event must be request.completed")
	}
	if err := validateRecordTime(record.OccurredAt); err != nil {
		return err
	}
	if err := ValidateRequestID(record.RequestID); err != nil {
		return err
	}
	if err := validateTraceCorrelation(record.TraceID, record.SpanID); err != nil {
		return err
	}
	if err := record.RecordActor.Validate(); err != nil {
		return err
	}
	hasHTTP := record.HTTPMethod != "" && record.HTTPRoute != "" && record.RPCService == "" && record.RPCMethod == ""
	hasRPC := record.RPCService != "" && record.RPCMethod != "" && record.HTTPRoute == ""
	if !hasHTTP && !hasRPC {
		return fmt.Errorf("request record requires exactly one HTTP route or RPC service and method")
	}
	if record.StatusCode < 100 || record.StatusCode > 599 || record.DurationMS < 0 {
		return fmt.Errorf("request result requires status_code and non-negative duration_ms")
	}
	switch record.Outcome {
	case RequestOutcomeSucceeded, RequestOutcomeBlocked, RequestOutcomeFailed:
	default:
		return fmt.Errorf("invalid request outcome %q", record.Outcome)
	}
	if record.Outcome == RequestOutcomeSucceeded && (record.ErrorCode != "" || record.Reason != "") {
		return fmt.Errorf("successful request cannot contain error_code or reason")
	}
	if record.Outcome != RequestOutcomeSucceeded && record.ErrorCode == "" && record.Reason == "" {
		return fmt.Errorf("blocked or failed request requires error_code or reason")
	}
	if hasHTTP {
		expected, _ := ClassifyHTTPResult(record.StatusCode, record.DurationMS)
		if record.Outcome != expected.Outcome || record.Reason != string(expected.Reason) {
			return fmt.Errorf("HTTP request outcome and reason must match the shared status classifier")
		}
	}
	return nil
}

type SystemEvent string

type AuditRecordClass string
type AuditAppendFailureReason string
type TranslationJobTerminalOutcome string
type TranslationEntityType string
type TranslationFailureReason string

const (
	AuditRecordClassDomain   AuditRecordClass = "domain_audit"
	AuditRecordClassSecurity AuditRecordClass = "security_access"

	AuditAppendFailureRecordInvalid         AuditAppendFailureReason = "record_invalid"
	AuditAppendFailureTransactionMissing    AuditAppendFailureReason = "transaction_missing"
	AuditAppendFailurePersistenceFailed     AuditAppendFailureReason = "persistence_failed"
	AuditAppendFailureDatabaseMissing       AuditAppendFailureReason = "database_missing"
	AuditAppendFailureRequestContextMissing AuditAppendFailureReason = "request_context_missing"
	AuditAppendFailureActorInvalid          AuditAppendFailureReason = "actor_invalid"
	AuditAppendFailureRecordBuildFailed     AuditAppendFailureReason = "record_build_failed"

	TranslationJobTerminalOutcomeApplied   TranslationJobTerminalOutcome = "applied"
	TranslationJobTerminalOutcomeFailed    TranslationJobTerminalOutcome = "failed"
	TranslationJobTerminalOutcomeCancelled TranslationJobTerminalOutcome = "cancelled"

	TranslationEntityPost          TranslationEntityType = "post"
	TranslationEntityPage          TranslationEntityType = "page"
	TranslationEntityWork          TranslationEntityType = "work"
	TranslationEntityProgramEvent  TranslationEntityType = "program_event"
	TranslationEntityRelease       TranslationEntityType = "release"
	TranslationEntityArtist        TranslationEntityType = "artist"
	TranslationEntityLabel         TranslationEntityType = "label"
	TranslationEntityMenu          TranslationEntityType = "menu"
	TranslationEntityEmailTemplate TranslationEntityType = "email_template"
	TranslationEntityEmailLayout   TranslationEntityType = "email_layout"
	TranslationEntityCampaign      TranslationEntityType = "campaign"
	TranslationEntityForm          TranslationEntityType = "form"
	TranslationEntityPrivacy       TranslationEntityType = "privacy"
	TranslationEntityTerms         TranslationEntityType = "terms"
	TranslationEntitySeries        TranslationEntityType = "series"

	TranslationFailureProviderConfiguration   TranslationFailureReason = "provider_configuration"
	TranslationFailureProviderAuthentication  TranslationFailureReason = "provider_authentication"
	TranslationFailureProviderRateLimited     TranslationFailureReason = "provider_rate_limited"
	TranslationFailureProviderUnavailable     TranslationFailureReason = "provider_unavailable"
	TranslationFailureProviderRejected        TranslationFailureReason = "provider_rejected"
	TranslationFailureProviderResponseInvalid TranslationFailureReason = "provider_response_invalid"
	TranslationFailureTargetApplyFailed       TranslationFailureReason = "target_apply_failed"
	TranslationFailureOGHandoffFailed         TranslationFailureReason = "og_handoff_failed"
	TranslationFailureInternal                TranslationFailureReason = "internal"
)

var auditAppendFailureReasons = map[AuditAppendFailureReason]struct{}{
	AuditAppendFailureRecordInvalid: {}, AuditAppendFailureTransactionMissing: {},
	AuditAppendFailurePersistenceFailed: {}, AuditAppendFailureDatabaseMissing: {},
	AuditAppendFailureRequestContextMissing: {}, AuditAppendFailureActorInvalid: {},
	AuditAppendFailureRecordBuildFailed: {},
}

var translationEntityTypes = map[TranslationEntityType]struct{}{
	TranslationEntityPost: {}, TranslationEntityPage: {}, TranslationEntityWork: {}, TranslationEntityProgramEvent: {},
	TranslationEntityRelease: {}, TranslationEntityArtist: {}, TranslationEntityLabel: {}, TranslationEntityMenu: {},
	TranslationEntityEmailTemplate: {}, TranslationEntityEmailLayout: {}, TranslationEntityCampaign: {}, TranslationEntityForm: {},
	TranslationEntityPrivacy: {}, TranslationEntityTerms: {}, TranslationEntitySeries: {},
}

var translationFailureReasons = map[TranslationFailureReason]struct{}{
	TranslationFailureProviderConfiguration: {}, TranslationFailureProviderAuthentication: {},
	TranslationFailureProviderRateLimited: {}, TranslationFailureProviderUnavailable: {},
	TranslationFailureProviderRejected: {}, TranslationFailureProviderResponseInvalid: {},
	TranslationFailureTargetApplyFailed: {},
	TranslationFailureOGHandoffFailed:   {}, TranslationFailureInternal: {},
}

var translationTerminalOutcomes = map[TranslationJobTerminalOutcome]struct{}{
	TranslationJobTerminalOutcomeApplied: {}, TranslationJobTerminalOutcomeFailed: {},
	TranslationJobTerminalOutcomeCancelled: {},
}

const (
	EventServiceReady                  SystemEvent = "service.ready"
	EventServiceDegraded               SystemEvent = "service.degraded"
	EventServiceStopping               SystemEvent = "service.stopping"
	EventServiceFailed                 SystemEvent = "service.failed"
	EventDependencyDegraded            SystemEvent = "dependency.degraded"
	EventDependencyRecovered           SystemEvent = "dependency.recovered"
	EventQueuePublishSucceeded         SystemEvent = "queue.publish.succeeded"
	EventQueuePublishFailed            SystemEvent = "queue.publish.failed"
	EventQueueDeliverySucceeded        SystemEvent = "queue.delivery.succeeded"
	EventQueueDeliveryFailed           SystemEvent = "queue.delivery.failed"
	EventQueueDeliveryRequeued         SystemEvent = "queue.delivery.requeued"
	EventQueueRetryAccepted            SystemEvent = "queue.retry.accepted"
	EventQueueRetryFailed              SystemEvent = "queue.retry.failed"
	EventQueueDLQAccepted              SystemEvent = "queue.dlq.accepted"
	EventQueueDLQFailed                SystemEvent = "queue.dlq.failed"
	EventJobStarted                    SystemEvent = "job.started"
	EventJobSucceeded                  SystemEvent = "job.succeeded"
	EventJobFailed                     SystemEvent = "job.failed"
	EventAuditAppendFailed             SystemEvent = "audit.append.failed"
	EventCollaborationCheckpointFailed SystemEvent = "collaboration.checkpoint.failed"
	EventClientRenderFailed            SystemEvent = "client.render.failed"
	EventTelemetryPipelineDegraded     SystemEvent = "telemetry.pipeline.degraded"
	EventTelemetryPipelineRecovered    SystemEvent = "telemetry.pipeline.recovered"
	EventTranslationJobTerminal        SystemEvent = "translation.job.terminal"
)

var systemEvents = map[SystemEvent]struct{}{
	EventServiceReady: {}, EventServiceDegraded: {}, EventServiceStopping: {}, EventServiceFailed: {},
	EventDependencyDegraded: {}, EventDependencyRecovered: {},
	EventQueuePublishSucceeded: {}, EventQueuePublishFailed: {},
	EventQueueDeliverySucceeded: {}, EventQueueDeliveryFailed: {}, EventQueueDeliveryRequeued: {},
	EventQueueRetryAccepted: {}, EventQueueRetryFailed: {}, EventQueueDLQAccepted: {}, EventQueueDLQFailed: {},
	EventJobStarted: {}, EventJobSucceeded: {}, EventJobFailed: {}, EventAuditAppendFailed: {},
	EventCollaborationCheckpointFailed: {}, EventClientRenderFailed: {},
	EventTelemetryPipelineDegraded: {}, EventTelemetryPipelineRecovered: {},
	EventTranslationJobTerminal: {},
}

var systemEventOutcomes = map[SystemEvent]string{
	EventServiceReady: "ready", EventServiceDegraded: "degraded", EventServiceStopping: "stopping", EventServiceFailed: "failed",
	EventDependencyDegraded: "degraded", EventDependencyRecovered: "recovered",
	EventQueuePublishSucceeded: "succeeded", EventQueuePublishFailed: "failed",
	EventQueueDeliverySucceeded: "succeeded", EventQueueDeliveryFailed: "failed", EventQueueDeliveryRequeued: "requeued",
	EventQueueRetryAccepted: "accepted", EventQueueRetryFailed: "failed", EventQueueDLQAccepted: "accepted", EventQueueDLQFailed: "failed",
	EventJobStarted: "started", EventJobSucceeded: "succeeded", EventJobFailed: "failed", EventAuditAppendFailed: "failed",
	EventCollaborationCheckpointFailed: "failed", EventClientRenderFailed: "failed",
	EventTelemetryPipelineDegraded: "degraded", EventTelemetryPipelineRecovered: "recovered",
}

type SystemRecord struct {
	Event      SystemEvent `json:"event"`
	OccurredAt time.Time   `json:"occurred_at"`
	Correlation
	Component           string           `json:"component,omitempty"`
	Dependency          string           `json:"dependency,omitempty"`
	Operation           string           `json:"operation,omitempty"`
	Domain              string           `json:"domain,omitempty"`
	Queue               string           `json:"queue,omitempty"`
	MessageID           string           `json:"message_id,omitempty"`
	CommandID           string           `json:"command_id,omitempty"`
	RetryCount          *int             `json:"retry_count,omitempty"`
	DurationMS          *int64           `json:"duration_ms,omitempty"`
	JobKind             string           `json:"job_kind,omitempty"`
	JobID               string           `json:"job_id,omitempty"`
	EntityType          string           `json:"entity_type,omitempty"`
	EntityID            string           `json:"entity_id,omitempty"`
	TargetLocale        string           `json:"target_locale,omitempty"`
	RecordClass         AuditRecordClass `json:"record_class,omitempty"`
	Action              string           `json:"action,omitempty"`
	Outcome             string           `json:"outcome,omitempty"`
	ErrorCode           string           `json:"error_code,omitempty"`
	Reason              string           `json:"reason,omitempty"`
	ErrorClassification string           `json:"error_classification,omitempty"`
}

func (record SystemRecord) Validate() error {
	if _, ok := systemEvents[record.Event]; !ok {
		return fmt.Errorf("unknown system event %q", record.Event)
	}
	if err := validateRecordTime(record.OccurredAt); err != nil {
		return err
	}
	if record.RequestID != "" {
		if err := ValidateRequestID(record.RequestID); err != nil {
			return err
		}
	}
	if err := validateTraceCorrelation(record.TraceID, record.SpanID); err != nil {
		return err
	}
	if record.Event == EventTranslationJobTerminal {
		if _, ok := translationTerminalOutcomes[TranslationJobTerminalOutcome(record.Outcome)]; !ok {
			return fmt.Errorf("translation.job.terminal requires a catalog outcome")
		}
	} else if expected, ok := systemEventOutcomes[record.Event]; !ok || record.Outcome != expected {
		return fmt.Errorf("system event %s requires outcome %s", record.Event, expected)
	}
	if record.Event != EventTranslationJobTerminal && record.ErrorClassification != "" {
		return fmt.Errorf("system event %s does not allow error_classification", record.Event)
	}
	for name, value := range map[string]string{"component": record.Component, "dependency": record.Dependency, "operation": record.Operation, "domain": record.Domain, "job_kind": record.JobKind, "entity_type": record.EntityType, "record_class": string(record.RecordClass), "outcome": record.Outcome, "error_code": record.ErrorCode, "reason": record.Reason, "error_classification": record.ErrorClassification} {
		if value != "" && !isBoundedCode(value) {
			return fmt.Errorf("system field %s must be a bounded code", name)
		}
	}
	if len(record.EntityID) > 128 {
		return fmt.Errorf("entity_id must be a bounded identifier")
	}
	if record.TargetLocale != "" && !isCanonicalAuditLocale(record.TargetLocale) {
		return fmt.Errorf("target_locale must be a bounded locale code")
	}
	if record.RetryCount != nil && *record.RetryCount < 0 {
		return fmt.Errorf("retry_count cannot be negative")
	}
	if record.DurationMS != nil && *record.DurationMS < 0 {
		return fmt.Errorf("duration_ms cannot be negative")
	}
	if err := record.validateRequiredFields(); err != nil {
		return err
	}
	return nil
}

func (record SystemRecord) validateRequiredFields() error {
	require := func(fields ...string) error {
		for _, field := range fields {
			if field == "" {
				return fmt.Errorf("system event %s is missing a required field", record.Event)
			}
		}
		return nil
	}
	requireFailure := func() error {
		if record.ErrorCode == "" && record.Reason == "" {
			return fmt.Errorf("system event %s requires error_code or reason", record.Event)
		}
		return nil
	}
	switch record.Event {
	case EventServiceReady, EventServiceStopping:
		return require(record.Component)
	case EventServiceDegraded, EventServiceFailed:
		if err := require(record.Component); err != nil {
			return err
		}
		return requireFailure()
	case EventDependencyRecovered:
		return require(record.Dependency, record.Operation)
	case EventDependencyDegraded:
		if err := require(record.Dependency, record.Operation); err != nil {
			return err
		}
		return requireFailure()
	case EventQueuePublishSucceeded:
		if err := require(record.Queue, record.MessageID, record.CommandID); err != nil {
			return err
		}
		if record.DurationMS == nil {
			return fmt.Errorf("system event %s requires duration_ms", record.Event)
		}
		return record.requireNoFailure()
	case EventQueuePublishFailed:
		if err := require(record.Queue, record.MessageID, record.CommandID); err != nil {
			return err
		}
		if record.DurationMS == nil {
			return fmt.Errorf("system event %s requires duration_ms", record.Event)
		}
		if err := requireFailure(); err != nil {
			return err
		}
		return record.requireQueueReason(QueueFailureEnqueueFailed)
	case EventQueueDeliverySucceeded:
		if err := require(record.Queue, record.MessageID, record.CommandID); err != nil {
			return err
		}
		if record.RetryCount == nil || record.DurationMS == nil {
			return fmt.Errorf("system event %s requires retry_count and duration_ms", record.Event)
		}
		return record.requireNoFailure()
	case EventQueueDeliveryFailed:
		if err := require(record.Queue, record.MessageID, record.CommandID); err != nil {
			return err
		}
		if record.RetryCount == nil || record.DurationMS == nil {
			return fmt.Errorf("system event %s requires retry_count and duration_ms", record.Event)
		}
		if err := requireFailure(); err != nil {
			return err
		}
		return record.requireQueueReason(QueueFailureHandlerFailed, QueueFailureCompletionFailed)
	case EventQueueDeliveryRequeued:
		if err := require(record.Queue, record.MessageID, record.CommandID); err != nil {
			return err
		}
		if record.RetryCount == nil || record.DurationMS == nil {
			return fmt.Errorf("system event %s requires retry_count and duration_ms", record.Event)
		}
		return record.requireQueueReason(QueueFailureShutdown, QueueFailureHandlerFailed)
	case EventQueueRetryAccepted, EventQueueDLQAccepted:
		if err := require(record.Queue, record.MessageID, record.CommandID); err != nil {
			return err
		}
		if record.RetryCount == nil {
			return fmt.Errorf("system event %s requires retry_count", record.Event)
		}
		return record.requireNoFailure()
	case EventQueueRetryFailed, EventQueueDLQFailed:
		if err := require(record.Queue, record.MessageID, record.CommandID); err != nil {
			return err
		}
		if record.RetryCount == nil {
			return fmt.Errorf("system event %s requires retry_count", record.Event)
		}
		if err := requireFailure(); err != nil {
			return err
		}
		if record.Event == EventQueueRetryFailed {
			return record.requireQueueReason(QueueFailureVisibilityUpdateFailed)
		}
		return record.requireQueueReason(QueueFailureArchiveFailed)
	case EventJobStarted:
		if err := require(record.JobKind, record.JobID); err != nil {
			return err
		}
		_, err := ParseJobKind(record.JobKind)
		return err
	case EventJobSucceeded:
		if err := require(record.JobKind, record.JobID); err != nil {
			return err
		}
		if _, err := ParseJobKind(record.JobKind); err != nil {
			return err
		}
		if record.DurationMS == nil {
			return fmt.Errorf("system event %s requires duration_ms", record.Event)
		}
	case EventJobFailed:
		if err := require(record.JobKind, record.JobID); err != nil {
			return err
		}
		jobKind, err := ParseJobKind(record.JobKind)
		if err != nil {
			return err
		}
		if record.DurationMS == nil {
			return fmt.Errorf("system event %s requires duration_ms", record.Event)
		}
		if record.ErrorCode != "" || record.Reason == "" {
			return fmt.Errorf("system event %s requires a catalog job failure reason", record.Event)
		}
		_, err = ParseJobFailureReason(jobKind, record.Reason)
		return err
	case EventAuditAppendFailed:
		if err := require(string(record.RecordClass), record.Action); err != nil {
			return err
		}
		if !isKnownAppendAction(record.RecordClass, record.Action) {
			return fmt.Errorf("system event %s requires a catalog action for record_class", record.Event)
		}
		if record.ErrorCode != "" {
			return fmt.Errorf("system event %s requires a catalog audit append reason", record.Event)
		}
		if _, ok := auditAppendFailureReasons[AuditAppendFailureReason(record.Reason)]; !ok {
			return fmt.Errorf("system event %s requires a catalog audit append reason", record.Event)
		}
		return nil
	case EventCollaborationCheckpointFailed:
		if err := require(record.Domain, record.EntityType, record.EntityID); err != nil {
			return err
		}
		if record.Domain != "collaboration" || record.RetryCount == nil || record.ErrorCode != "" {
			return fmt.Errorf("collaboration.checkpoint.failed requires a terminal checkpoint context and catalog reason")
		}
		if _, ok := map[string]struct{}{
			"post": {}, "page": {}, "work": {}, "release": {}, "label": {},
			"artist": {}, "campaign": {}, "program_event": {}, "map_theme": {},
			"terms_history": {}, "privacy_history": {}, "form": {}, "email_layout": {},
		}[record.EntityType]; !ok {
			return fmt.Errorf("collaboration.checkpoint.failed requires a catalog entity type")
		}
		switch record.Reason {
		case "locale_ownership_changed", "shared_structure_changed", "shared_materialization_changed", "document_revision_changed", "target_revision_changed":
			if *record.RetryCount != 1 {
				return fmt.Errorf("collaboration.checkpoint.failed conflict requires retry_count 1")
			}
			return nil
		case "shared_document_unavailable", "source_document_unavailable", "persist_failed":
			return nil
		default:
			return fmt.Errorf("collaboration.checkpoint.failed requires a terminal checkpoint context and catalog reason")
		}
	case EventClientRenderFailed:
		if err := require(record.Domain, record.Component); err != nil {
			return err
		}
		if record.Domain != "client" || record.ErrorCode != "" || record.Reason != "react_error_boundary" {
			return fmt.Errorf("client.render.failed requires a catalog component and reason")
		}
		switch record.Component {
		case "general", "admin", "global":
			return nil
		default:
			return fmt.Errorf("client.render.failed requires a catalog component and reason")
		}
	case EventTranslationJobTerminal:
		if err := require(record.Domain, record.JobID, record.EntityType, record.TargetLocale); err != nil {
			return err
		}
		if err := record.requireTranslationTerminalOnlyFields(); err != nil {
			return err
		}
		if record.Domain != "translation" || !isCanonicalUUID(record.JobID) ||
			!isCanonicalAuditLocale(record.TargetLocale) {
			return fmt.Errorf("translation terminal event requires a canonical job and locale context")
		}
		if _, ok := translationEntityTypes[TranslationEntityType(record.EntityType)]; !ok {
			return fmt.Errorf("translation terminal event requires a catalog entity type")
		}
		if record.DurationMS == nil {
			return fmt.Errorf("system event %s requires duration_ms", record.Event)
		}
		if record.Outcome != string(TranslationJobTerminalOutcomeFailed) {
			if record.ErrorClassification != "" {
				return fmt.Errorf("non-failed translation.job.terminal cannot contain error_classification")
			}
			return record.requireNoFailure()
		}
		if _, ok := translationFailureReasons[TranslationFailureReason(record.ErrorClassification)]; !ok {
			return fmt.Errorf("failed translation.job.terminal requires a catalog error_classification")
		}
		return nil
	case EventTelemetryPipelineRecovered:
		return require(record.Component)
	case EventTelemetryPipelineDegraded:
		if err := require(record.Component); err != nil {
			return err
		}
		return requireFailure()
	}
	return nil
}

func (record SystemRecord) requireNoFailure() error {
	if record.ErrorCode != "" || record.Reason != "" {
		return fmt.Errorf("system event %s does not allow failure fields", record.Event)
	}
	return nil
}

func (record SystemRecord) requireTranslationTerminalOnlyFields() error {
	if record.Component != "" || record.Dependency != "" || record.Operation != "" ||
		record.Queue != "" || record.MessageID != "" || record.CommandID != "" ||
		record.RetryCount != nil || record.JobKind != "" || record.EntityID != "" ||
		record.RecordClass != "" || record.Action != "" || record.ErrorCode != "" || record.Reason != "" {
		return fmt.Errorf("translation terminal event contains unsupported fields")
	}
	return nil
}

func (record SystemRecord) requireQueueReason(allowed ...QueueFailureReason) error {
	if record.ErrorCode != "" {
		return fmt.Errorf("system event %s requires a catalog queue reason", record.Event)
	}
	for _, reason := range allowed {
		if record.Reason == string(reason) {
			return nil
		}
	}
	return fmt.Errorf("system event %s requires a catalog queue reason", record.Event)
}

func isKnownAppendAction(recordClass AuditRecordClass, action string) bool {
	switch recordClass {
	case AuditRecordClassDomain:
		_, ok := auditCatalog[AuditAction(action)]
		return ok
	case AuditRecordClassSecurity:
		switch SecurityAction(action) {
		case SecurityAuthenticationSucceeded, SecurityAuthenticationFailed, SecurityAuthenticationBlocked,
			SecurityAuthorizationDenied, SecurityPersonalDataAccessed:
			return true
		}
	}
	return false
}

/*
The domain-audit contract lives in audit_types.go and its focused catalog
validators. Generic record helpers continue below.
*/

func validateSortedUnique(name string, values []string, boundedCodes bool) error {
	for index, value := range values {
		if value == "" || (boundedCodes && !isBoundedCode(value)) {
			return fmt.Errorf("%s contains an invalid value", name)
		}
		if index > 0 && values[index-1] >= value {
			return fmt.Errorf("%s must be sorted and unique", name)
		}
	}
	return nil
}

func validateSortedUniqueUUIDv4(name string, values []string) error {
	for index, value := range values {
		parsed, err := uuid.Parse(value)
		if err != nil || parsed.String() != value || parsed.Version() != 4 || parsed.Variant() != uuid.RFC4122 {
			return fmt.Errorf("%s contains a non-canonical UUIDv4", name)
		}
		if index > 0 && values[index-1] >= value {
			return fmt.Errorf("%s must be sorted and unique", name)
		}
	}
	return nil
}

func validateSortedUniqueIdentifiers(name string, values []string) error {
	for index, value := range values {
		if !isAuditIdentifier(value) {
			return fmt.Errorf("%s contains an invalid identifier", name)
		}
		if index > 0 && values[index-1] >= value {
			return fmt.Errorf("%s must be sorted and unique", name)
		}
	}
	return nil
}

func isAuditIdentifier(value string) bool {
	return value != "" && len(value) <= 255 && strings.TrimSpace(value) == value
}

func isAuditEmail(value string) bool {
	return len(value) <= 320 && strings.TrimSpace(value) == value &&
		!strings.ContainsAny(value, " \t\r\n") && strings.Count(value, "@") == 1
}

type SecurityAction string
type AuthenticationFlowKind string
type AuthenticationMethod string
type AuthenticationPrincipalState string
type AuthenticationFailureReason string
type AuthenticationBlockReason string
type AuthorizationDenialReason string
type PersonalDataAccessKind string

const (
	SecurityAuthenticationSucceeded SecurityAction = "authentication.succeeded"
	SecurityAuthenticationFailed    SecurityAction = "authentication.failed"
	SecurityAuthenticationBlocked   SecurityAction = "authentication.blocked"
	SecurityAuthorizationDenied     SecurityAction = "authorization.denied"
	SecurityPersonalDataAccessed    SecurityAction = "personal_data.accessed"
)

const (
	AuthenticationFlowLogin            AuthenticationFlowKind = "login"
	AuthenticationFlowRegistration     AuthenticationFlowKind = "registration"
	AuthenticationFlowReauthentication AuthenticationFlowKind = "reauthentication"

	AuthenticationMethodEmailCode AuthenticationMethod = "email_code"
	AuthenticationMethodOIDC      AuthenticationMethod = "oidc"
	AuthenticationMethodPasskey   AuthenticationMethod = "passkey"

	AuthenticationPrincipalOnboardingOnly AuthenticationPrincipalState = "onboarding_only"
	AuthenticationPrincipalActive         AuthenticationPrincipalState = "active"

	AuthenticationFailureProofRejected     AuthenticationFailureReason = "proof_rejected"
	AuthenticationFailureAccountBlocked    AuthenticationFailureReason = "account_blocked"
	AuthenticationFailureProviderDenied    AuthenticationFailureReason = "provider_denied"
	AuthenticationFailureProviderFailed    AuthenticationFailureReason = "provider_failed"
	AuthenticationFailureMemberLinkInvalid AuthenticationFailureReason = "member_link_invalid"
	AuthenticationFailureInternalError     AuthenticationFailureReason = "internal_error"

	AuthenticationBlockFlowInvalid          AuthenticationBlockReason = "flow_invalid"
	AuthenticationBlockRequestInvalid       AuthenticationBlockReason = "request_invalid"
	AuthenticationBlockIntegrityCheckFailed AuthenticationBlockReason = "integrity_check_failed"
	AuthenticationBlockRateLimited          AuthenticationBlockReason = "rate_limited"
	AuthenticationBlockServiceUnavailable   AuthenticationBlockReason = "service_unavailable"

	AuthorizationDeniedAuthenticationRequired AuthorizationDenialReason = "authentication_required"
	AuthorizationDeniedPermissionDenied       AuthorizationDenialReason = "permission_denied"
	AuthorizationProcedureInvokePermission                              = "procedure:invoke"

	PersonalDataAccessRead PersonalDataAccessKind = "read"
)

var authenticationFailureReasons = map[AuthenticationFailureReason]struct{}{
	AuthenticationFailureProofRejected: {}, AuthenticationFailureAccountBlocked: {},
	AuthenticationFailureProviderDenied: {}, AuthenticationFailureProviderFailed: {},
	AuthenticationFailureMemberLinkInvalid: {}, AuthenticationFailureInternalError: {},
}

var authenticationBlockReasons = map[AuthenticationBlockReason]struct{}{
	AuthenticationBlockFlowInvalid: {}, AuthenticationBlockRequestInvalid: {},
	AuthenticationBlockIntegrityCheckFailed: {}, AuthenticationBlockRateLimited: {},
	AuthenticationBlockServiceUnavailable: {},
}

var authorizationDenialReasons = map[AuthorizationDenialReason]struct{}{
	AuthorizationDeniedAuthenticationRequired: {},
	AuthorizationDeniedPermissionDenied:       {},
}

type SecurityAccessRecord struct {
	AccessID   string         `json:"access_id"`
	OccurredAt time.Time      `json:"occurred_at"`
	Action     SecurityAction `json:"action"`
	Correlation
	RecordActor
	SourceIP             string                       `json:"source_ip"`
	FlowKind             AuthenticationFlowKind       `json:"flow_kind,omitempty"`
	AuthenticationMethod AuthenticationMethod         `json:"authentication_method,omitempty"`
	PrincipalState       AuthenticationPrincipalState `json:"principal_state,omitempty"`
	Provider             string                       `json:"provider,omitempty"`
	Reason               string                       `json:"reason,omitempty"`
	AttemptedAction      string                       `json:"attempted_action,omitempty"`
	Permission           string                       `json:"permission,omitempty"`
	SubjectType          string                       `json:"subject_type,omitempty"`
	SubjectID            string                       `json:"subject_id,omitempty"`
	AccessKind           PersonalDataAccessKind       `json:"access_kind,omitempty"`
	DataCategory         string                       `json:"data_category,omitempty"`
}

func (record SecurityAccessRecord) Validate() error {
	parsedAccessID, err := uuid.Parse(record.AccessID)
	if err != nil || parsedAccessID.String() != record.AccessID || parsedAccessID.Version() != 4 || parsedAccessID.Variant() != uuid.RFC4122 {
		return fmt.Errorf("access_id must be a canonical UUIDv4")
	}
	if err := validateRecordTime(record.OccurredAt); err != nil {
		return err
	}
	if err := ValidateRequestID(record.RequestID); err != nil {
		return err
	}
	if err := validateTraceCorrelation(record.TraceID, record.SpanID); err != nil {
		return err
	}
	if err := record.RecordActor.Validate(); err != nil {
		return err
	}
	if record.SourceIP == "" || validateSourceIP(record.SourceIP) != nil {
		return ErrInvalidSourceIP
	}
	switch record.Action {
	case SecurityAuthenticationSucceeded:
		if err := record.requireOnlyAuthenticationAttributes(); err != nil {
			return err
		}
		if record.Kind != ActorKindMember {
			return fmt.Errorf("successful authentication requires a member actor")
		}
		if !validAuthenticationFlow(record.FlowKind) || !validAuthenticationMethod(record.AuthenticationMethod) {
			return fmt.Errorf("successful authentication requires flow_kind and authentication_method")
		}
		if !validAuthenticationPrincipalState(record.PrincipalState) {
			return fmt.Errorf("successful authentication requires principal_state")
		}
		if record.Reason != "" {
			return fmt.Errorf("successful authentication cannot contain reason")
		}
		return record.validateAuthenticationProvider()
	case SecurityAuthenticationFailed:
		if err := record.requireOnlyAuthenticationAttributes(); err != nil {
			return err
		}
		if record.PrincipalState != "" {
			return fmt.Errorf("authentication failure cannot contain principal_state")
		}
		if record.Kind == ActorKindSystem {
			return fmt.Errorf("authentication failure cannot use a system actor")
		}
		if !validAuthenticationFlow(record.FlowKind) || !validAuthenticationMethod(record.AuthenticationMethod) {
			return fmt.Errorf("authentication failure requires flow_kind and authentication_method")
		}
		if record.Kind == ActorKindMember && record.FlowKind != AuthenticationFlowReauthentication {
			return fmt.Errorf("only reauthentication failure or block can use a member actor")
		}
		if _, ok := authenticationFailureReasons[AuthenticationFailureReason(record.Reason)]; !ok {
			return fmt.Errorf("invalid authentication failure reason %q", record.Reason)
		}
		return record.validateAuthenticationProvider()
	case SecurityAuthenticationBlocked:
		if err := record.requireOnlyAuthenticationAttributes(); err != nil {
			return err
		}
		if record.PrincipalState != "" {
			return fmt.Errorf("authentication block cannot contain principal_state")
		}
		if record.Kind == ActorKindSystem {
			return fmt.Errorf("authentication block cannot use a system actor")
		}
		if record.FlowKind != "" && !validAuthenticationFlow(record.FlowKind) {
			return fmt.Errorf("invalid authentication flow_kind %q", record.FlowKind)
		}
		if record.AuthenticationMethod != "" && !validAuthenticationMethod(record.AuthenticationMethod) {
			return fmt.Errorf("invalid authentication_method %q", record.AuthenticationMethod)
		}
		if record.Kind == ActorKindMember && record.FlowKind != AuthenticationFlowReauthentication {
			return fmt.Errorf("only reauthentication failure or block can use a member actor")
		}
		if _, ok := authenticationBlockReasons[AuthenticationBlockReason(record.Reason)]; !ok {
			return fmt.Errorf("invalid authentication block reason %q", record.Reason)
		}
		return record.validateAuthenticationProvider()
	case SecurityAuthorizationDenied:
		if err := record.requireOnlyAuthorizationAttributes(); err != nil {
			return err
		}
		if record.Kind == ActorKindSystem {
			return fmt.Errorf("authorization denial cannot use a system actor")
		}
		if !validAuthorizationScope(record.AttemptedAction, record.Permission) {
			return fmt.Errorf("authorization denial requires a cataloged attempted action and permission")
		}
		if _, ok := authorizationDenialReasons[AuthorizationDenialReason(record.Reason)]; !ok {
			return fmt.Errorf("invalid authorization denial reason %q", record.Reason)
		}
	case SecurityPersonalDataAccessed:
		if err := record.requireOnlyPersonalDataAttributes(); err != nil {
			return err
		}
		if record.Kind != ActorKindMember || record.AccessKind != PersonalDataAccessRead || !validPersonalDataSubject(record.SubjectType, record.SubjectID, record.DataCategory) {
			return fmt.Errorf("personal data access requires a cataloged member read scope")
		}
	default:
		return fmt.Errorf("unknown security action %q", record.Action)
	}
	return nil
}

func validAuthorizationScope(attemptedAction, permission string) bool {
	return isConnectProcedure(attemptedAction) && permission == AuthorizationProcedureInvokePermission
}

func (record SecurityAccessRecord) validateAuthenticationProvider() error {
	if record.Provider == "" {
		return nil
	}
	if record.AuthenticationMethod != AuthenticationMethodOIDC || !isBoundedCode(record.Provider) {
		return fmt.Errorf("provider is only valid as a bounded code for oidc authentication")
	}
	return nil
}

func (record SecurityAccessRecord) requireOnlyAuthenticationAttributes() error {
	if record.AttemptedAction != "" || record.Permission != "" ||
		record.SubjectType != "" || record.SubjectID != "" || record.AccessKind != "" || record.DataCategory != "" {
		return fmt.Errorf("authentication access contains attributes for another security action")
	}
	return nil
}

func (record SecurityAccessRecord) requireOnlyAuthorizationAttributes() error {
	if record.FlowKind != "" || record.AuthenticationMethod != "" || record.PrincipalState != "" || record.Provider != "" ||
		record.SubjectType != "" || record.SubjectID != "" || record.AccessKind != "" || record.DataCategory != "" {
		return fmt.Errorf("authorization denial contains attributes for another security action")
	}
	return nil
}

func (record SecurityAccessRecord) requireOnlyPersonalDataAttributes() error {
	if record.FlowKind != "" || record.AuthenticationMethod != "" || record.PrincipalState != "" || record.Provider != "" || record.Reason != "" ||
		record.AttemptedAction != "" || record.Permission != "" {
		return fmt.Errorf("personal data access contains attributes for another security action")
	}
	return nil
}

func validAuthenticationFlow(value AuthenticationFlowKind) bool {
	return value == AuthenticationFlowLogin || value == AuthenticationFlowRegistration || value == AuthenticationFlowReauthentication
}

func validAuthenticationMethod(value AuthenticationMethod) bool {
	return value == AuthenticationMethodEmailCode || value == AuthenticationMethodOIDC || value == AuthenticationMethodPasskey
}

func validAuthenticationPrincipalState(value AuthenticationPrincipalState) bool {
	return value == AuthenticationPrincipalOnboardingOnly || value == AuthenticationPrincipalActive
}

func validPersonalDataSubject(subjectType, subjectID, dataCategory string) bool {
	if subjectType == "member_collection" {
		return subjectID == "1" && dataCategory == "member_administration"
	}
	if !isCanonicalUUID(subjectID) {
		return false
	}
	switch subjectType {
	case "member":
		return dataCategory == "member_administration"
	case "campaign":
		return dataCategory == "campaign_recipients"
	case "form":
		return dataCategory == "form_submissions"
	case "form_submission":
		return dataCategory == "form_submission"
	default:
		return false
	}
}

func isCanonicalUUID(value string) bool {
	parsed, err := uuid.Parse(value)
	if err != nil || parsed.String() != value || parsed.Variant() != uuid.RFC4122 {
		return false
	}
	switch parsed.Version() {
	case 1, 2, 3, 4, 5, 6, 7, 8:
		return true
	default:
		return false
	}
}

func isConnectProcedure(value string) bool {
	if len(value) < 4 || len(value) > 128 || !strings.HasPrefix(value, "/") || strings.Count(value, "/") != 2 {
		return false
	}
	service, method, ok := strings.Cut(strings.TrimPrefix(value, "/"), "/")
	return ok && isConnectIdentifier(service, true) && isConnectIdentifier(method, false)
}

func isConnectIdentifier(value string, allowDot bool) bool {
	if value == "" {
		return false
	}
	for _, character := range value {
		if (character >= 'a' && character <= 'z') || (character >= 'A' && character <= 'Z') ||
			(character >= '0' && character <= '9') || character == '_' || (allowDot && character == '.') {
			continue
		}
		return false
	}
	return true
}

func validateRecordTime(value time.Time) error {
	if value.IsZero() {
		return fmt.Errorf("occurred_at is required")
	}
	_, offset := value.Zone()
	if offset != 0 {
		return fmt.Errorf("occurred_at must be UTC")
	}
	return nil
}

func validateTraceCorrelation(traceID, spanID string) error {
	if traceID == "" && spanID == "" {
		return nil
	}
	if traceID == "" || spanID == "" {
		return fmt.Errorf("trace_id and span_id must be provided together")
	}
	parsedTraceID, traceErr := trace.TraceIDFromHex(traceID)
	parsedSpanID, spanErr := trace.SpanIDFromHex(spanID)
	if traceErr != nil || spanErr != nil || !parsedTraceID.IsValid() || !parsedSpanID.IsValid() {
		return fmt.Errorf("invalid trace correlation")
	}
	return nil
}

func isDottedName(value string) bool {
	parts := strings.Split(value, ".")
	if len(parts) < 2 {
		return false
	}
	for _, part := range parts {
		if part == "" || strings.ToLower(part) != part {
			return false
		}
		for _, character := range part {
			if (character < 'a' || character > 'z') && (character < '0' || character > '9') && character != '_' {
				return false
			}
		}
	}
	return true
}

func isBoundedCode(value string) bool {
	if value == "" || len(value) > 64 {
		return false
	}
	for _, character := range value {
		if (character < 'a' || character > 'z') && (character < '0' || character > '9') && character != '_' {
			return false
		}
	}
	return true
}
