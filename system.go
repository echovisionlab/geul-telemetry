package telemetry

import (
	"fmt"
	"time"
)

// SystemMetadata carries correlation and the fact time shared by System
// records. A zero OccurredAt is accepted by EmitSystem, but builders require an
// explicit time so they remain deterministic in tests and queue boundaries.
type SystemMetadata struct {
	OccurredAt time.Time
	Correlation
}

type SystemFailure struct {
	ErrorCode string
	Reason    string
}

type QueueFailureReason string

const (
	QueueFailureEnqueueFailed          QueueFailureReason = "enqueue_failed"
	QueueFailureHandlerFailed          QueueFailureReason = "handler_failed"
	QueueFailureCompletionFailed       QueueFailureReason = "completion_failed"
	QueueFailureShutdown               QueueFailureReason = "shutdown"
	QueueFailureVisibilityUpdateFailed QueueFailureReason = "visibility_update_failed"
	QueueFailureArchiveFailed          QueueFailureReason = "archive_failed"
)

type QueuePublishContext struct {
	Queue      string
	MessageID  string
	CommandID  string
	DurationMS int64
}

type QueueDeliveryContext struct {
	Queue      string
	MessageID  string
	CommandID  string
	RetryCount int
	DurationMS int64
}

type QueueHandoffContext struct {
	Queue      string
	MessageID  string
	CommandID  string
	RetryCount int
}

type JobContext struct {
	JobKind JobKind
	JobID   string
}

type JobFailure struct {
	Reason JobFailureReason
}

// TranslationJobTerminalContext is the bounded public context shared by each
// authoritative Translation job terminal transition. It deliberately omits
// provider diagnostics and content; correlated traces own those details.
type TranslationJobTerminalContext struct {
	JobID        string
	EntityType   TranslationEntityType
	TargetLocale string
	DurationMS   int64
}

type CollaborationCheckpointFailureReason string

const (
	CollaborationCheckpointFailureSharedDocumentUnavailable    CollaborationCheckpointFailureReason = "shared_document_unavailable"
	CollaborationCheckpointFailureSourceDocumentUnavailable    CollaborationCheckpointFailureReason = "source_document_unavailable"
	CollaborationCheckpointFailureLocaleOwnershipChanged       CollaborationCheckpointFailureReason = "locale_ownership_changed"
	CollaborationCheckpointFailureSharedStructureChanged       CollaborationCheckpointFailureReason = "shared_structure_changed"
	CollaborationCheckpointFailureSharedMaterializationChanged CollaborationCheckpointFailureReason = "shared_materialization_changed"
	CollaborationCheckpointFailureDocumentRevisionChanged      CollaborationCheckpointFailureReason = "document_revision_changed"
	CollaborationCheckpointFailureTargetRevisionChanged        CollaborationCheckpointFailureReason = "target_revision_changed"
	CollaborationCheckpointFailurePersistFailed                CollaborationCheckpointFailureReason = "persist_failed"
)

type CollaborationCheckpointContext struct {
	EntityType string
	EntityID   string
	RetryCount int
}

type ClientRenderComponent string

const (
	ClientRenderComponentGeneral ClientRenderComponent = "general"
	ClientRenderComponentAdmin   ClientRenderComponent = "admin"
	ClientRenderComponentGlobal  ClientRenderComponent = "global"
)

func NewServiceReadyRecord(metadata SystemMetadata, component string) (SystemRecord, error) {
	return newSystemRecord(metadata, EventServiceReady, SystemRecord{Component: component})
}

func NewServiceDegradedRecord(metadata SystemMetadata, component string, failure SystemFailure) (SystemRecord, error) {
	return newSystemFailureRecord(metadata, EventServiceDegraded, SystemRecord{Component: component}, failure)
}

func NewServiceStoppingRecord(metadata SystemMetadata, component string) (SystemRecord, error) {
	return newSystemRecord(metadata, EventServiceStopping, SystemRecord{Component: component})
}

func NewServiceFailedRecord(metadata SystemMetadata, component string, failure SystemFailure) (SystemRecord, error) {
	return newSystemFailureRecord(metadata, EventServiceFailed, SystemRecord{Component: component}, failure)
}

func NewDependencyDegradedRecord(metadata SystemMetadata, dependency, operation string, failure SystemFailure) (SystemRecord, error) {
	return newSystemFailureRecord(metadata, EventDependencyDegraded, SystemRecord{Dependency: dependency, Operation: operation}, failure)
}

func NewDependencyRecoveredRecord(metadata SystemMetadata, dependency, operation string) (SystemRecord, error) {
	return newSystemRecord(metadata, EventDependencyRecovered, SystemRecord{Dependency: dependency, Operation: operation})
}

func NewQueuePublishSucceededRecord(metadata SystemMetadata, queue QueuePublishContext) (SystemRecord, error) {
	return newSystemRecord(metadata, EventQueuePublishSucceeded, queuePublishRecord(queue))
}

func NewQueuePublishFailedRecord(metadata SystemMetadata, queue QueuePublishContext, reason QueueFailureReason) (SystemRecord, error) {
	return newSystemFailureRecord(metadata, EventQueuePublishFailed, queuePublishRecord(queue), SystemFailure{Reason: string(reason)})
}

func NewQueueDeliverySucceededRecord(metadata SystemMetadata, queue QueueDeliveryContext) (SystemRecord, error) {
	return newSystemRecord(metadata, EventQueueDeliverySucceeded, queueDeliveryRecord(queue))
}

func NewQueueDeliveryFailedRecord(metadata SystemMetadata, queue QueueDeliveryContext, reason QueueFailureReason) (SystemRecord, error) {
	return newSystemFailureRecord(metadata, EventQueueDeliveryFailed, queueDeliveryRecord(queue), SystemFailure{Reason: string(reason)})
}

func NewQueueDeliveryRequeuedRecord(metadata SystemMetadata, queue QueueDeliveryContext, reason QueueFailureReason) (SystemRecord, error) {
	record := queueDeliveryRecord(queue)
	record.Reason = string(reason)
	return newSystemRecord(metadata, EventQueueDeliveryRequeued, record)
}

func NewQueueRetryAcceptedRecord(metadata SystemMetadata, queue QueueHandoffContext) (SystemRecord, error) {
	return newSystemRecord(metadata, EventQueueRetryAccepted, queueHandoffRecord(queue))
}

func NewQueueRetryFailedRecord(metadata SystemMetadata, queue QueueHandoffContext, reason QueueFailureReason) (SystemRecord, error) {
	return newSystemFailureRecord(metadata, EventQueueRetryFailed, queueHandoffRecord(queue), SystemFailure{Reason: string(reason)})
}

func NewQueueDLQAcceptedRecord(metadata SystemMetadata, queue QueueHandoffContext) (SystemRecord, error) {
	return newSystemRecord(metadata, EventQueueDLQAccepted, queueHandoffRecord(queue))
}

func NewQueueDLQFailedRecord(metadata SystemMetadata, queue QueueHandoffContext, reason QueueFailureReason) (SystemRecord, error) {
	return newSystemFailureRecord(metadata, EventQueueDLQFailed, queueHandoffRecord(queue), SystemFailure{Reason: string(reason)})
}

func NewJobStartedRecord(metadata SystemMetadata, job JobContext) (SystemRecord, error) {
	return newSystemRecord(metadata, EventJobStarted, SystemRecord{JobKind: job.JobKind.String(), JobID: job.JobID})
}

func NewCollaborationCheckpointFailedRecord(metadata SystemMetadata, checkpoint CollaborationCheckpointContext, reason CollaborationCheckpointFailureReason) (SystemRecord, error) {
	retryCount := checkpoint.RetryCount
	return newSystemRecord(metadata, EventCollaborationCheckpointFailed, SystemRecord{
		Domain: "collaboration", EntityType: checkpoint.EntityType, EntityID: checkpoint.EntityID,
		RetryCount: &retryCount, Reason: string(reason),
	})
}

func NewClientRenderFailedRecord(metadata SystemMetadata, component ClientRenderComponent) (SystemRecord, error) {
	return newSystemRecord(metadata, EventClientRenderFailed, SystemRecord{
		Domain: "client", Component: string(component), Reason: "react_error_boundary",
	})
}

func NewJobSucceededRecord(metadata SystemMetadata, job JobContext, durationMS int64) (SystemRecord, error) {
	return newSystemRecord(metadata, EventJobSucceeded, SystemRecord{JobKind: job.JobKind.String(), JobID: job.JobID, DurationMS: int64Pointer(durationMS)})
}

func NewJobFailedRecord(metadata SystemMetadata, job JobContext, durationMS int64, failure JobFailure) (SystemRecord, error) {
	return newSystemFailureRecord(metadata, EventJobFailed, SystemRecord{JobKind: job.JobKind.String(), JobID: job.JobID, DurationMS: int64Pointer(durationMS)}, SystemFailure{
		Reason: string(failure.Reason),
	})
}

// NewTranslationJobTerminalRecord builds the exact terminal Translation event.
// The outcome is bounded independently from the single event name, and only a
// failed outcome may carry a catalog error classification.
func NewTranslationJobTerminalRecord(
	metadata SystemMetadata,
	job TranslationJobTerminalContext,
	outcome TranslationJobTerminalOutcome,
	errorClassification TranslationFailureReason,
) (SystemRecord, error) {
	record := SystemRecord{
		Domain: "translation", JobID: job.JobID, EntityType: string(job.EntityType),
		TargetLocale: job.TargetLocale, DurationMS: int64Pointer(job.DurationMS), Outcome: string(outcome),
	}
	if outcome == TranslationJobTerminalOutcomeFailed {
		record.ErrorClassification = string(errorClassification)
	} else if errorClassification != "" {
		return SystemRecord{}, fmt.Errorf("translation terminal outcome %q cannot contain an error classification", outcome)
	}
	return newSystemRecord(metadata, EventTranslationJobTerminal, record)
}

func NewDomainAuditAppendFailedRecord(metadata SystemMetadata, action AuditAction, reason AuditAppendFailureReason) (SystemRecord, error) {
	return newAuditAppendFailedRecord(metadata, AuditRecordClassDomain, string(action), reason)
}

func NewSecurityAccessAppendFailedRecord(metadata SystemMetadata, action SecurityAction, reason AuditAppendFailureReason) (SystemRecord, error) {
	return newAuditAppendFailedRecord(metadata, AuditRecordClassSecurity, string(action), reason)
}

func newAuditAppendFailedRecord(metadata SystemMetadata, recordClass AuditRecordClass, action string, reason AuditAppendFailureReason) (SystemRecord, error) {
	return newSystemRecord(metadata, EventAuditAppendFailed, SystemRecord{
		Domain: "audit", RecordClass: recordClass, Action: action, Reason: string(reason),
	})
}

func NewTelemetryPipelineDegradedRecord(metadata SystemMetadata, component string, failure SystemFailure) (SystemRecord, error) {
	return newSystemFailureRecord(metadata, EventTelemetryPipelineDegraded, SystemRecord{Component: component}, failure)
}

func NewTelemetryPipelineRecoveredRecord(metadata SystemMetadata, component string) (SystemRecord, error) {
	return newSystemRecord(metadata, EventTelemetryPipelineRecovered, SystemRecord{Component: component})
}

func queuePublishRecord(queue QueuePublishContext) SystemRecord {
	return SystemRecord{Domain: "queue", Queue: queue.Queue, MessageID: queue.MessageID, CommandID: queue.CommandID, DurationMS: int64Pointer(queue.DurationMS)}
}

func queueDeliveryRecord(queue QueueDeliveryContext) SystemRecord {
	return SystemRecord{Domain: "queue", Queue: queue.Queue, MessageID: queue.MessageID, CommandID: queue.CommandID, RetryCount: intPointer(queue.RetryCount), DurationMS: int64Pointer(queue.DurationMS)}
}

func queueHandoffRecord(queue QueueHandoffContext) SystemRecord {
	return SystemRecord{Domain: "queue", Queue: queue.Queue, MessageID: queue.MessageID, CommandID: queue.CommandID, RetryCount: intPointer(queue.RetryCount)}
}

func newSystemFailureRecord(metadata SystemMetadata, event SystemEvent, attributes SystemRecord, failure SystemFailure) (SystemRecord, error) {
	attributes.ErrorCode, attributes.Reason = failure.ErrorCode, failure.Reason
	return newSystemRecord(metadata, event, attributes)
}

func newSystemRecord(metadata SystemMetadata, event SystemEvent, attributes SystemRecord) (SystemRecord, error) {
	attributes.Event = event
	if attributes.Outcome == "" {
		attributes.Outcome = systemEventOutcomes[event]
	}
	attributes.OccurredAt = metadata.OccurredAt
	attributes.Correlation = metadata.Correlation
	if err := attributes.Validate(); err != nil {
		return SystemRecord{}, err
	}
	return attributes, nil
}

func intPointer(value int) *int       { return &value }
func int64Pointer(value int64) *int64 { return &value }
