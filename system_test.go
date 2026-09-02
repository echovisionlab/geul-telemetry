package telemetry

import (
	"bytes"
	"context"
	"encoding/json"
	"log/slog"
	"os"
	"testing"
	"time"
)

func TestSystemCatalogMatchesCrossLanguageFixture(t *testing.T) {
	t.Parallel()
	data, err := os.ReadFile("fixtures/system-catalog.json")
	if err != nil {
		t.Fatal(err)
	}
	var fixture []struct {
		Event   SystemEvent `json:"event"`
		Outcome string      `json:"outcome"`
	}
	if err := json.Unmarshal(data, &fixture); err != nil {
		t.Fatal(err)
	}
	expectedVariants := len(systemEventOutcomes) + len(translationTerminalOutcomes)
	if len(fixture) != expectedVariants {
		t.Fatalf("fixture has %d of %d event outcomes", len(fixture), expectedVariants)
	}
	seen := make(map[string]struct{}, len(fixture))
	for _, entry := range fixture {
		if _, ok := systemEvents[entry.Event]; !ok {
			t.Fatalf("fixture entry does not match Go catalog: %#v", entry)
		}
		if entry.Event == EventTranslationJobTerminal {
			if _, ok := translationTerminalOutcomes[TranslationJobTerminalOutcome(entry.Outcome)]; !ok {
				t.Fatalf("fixture entry does not match Go catalog: %#v", entry)
			}
		} else if systemEventOutcomes[entry.Event] != entry.Outcome {
			t.Fatalf("fixture entry does not match Go catalog: %#v", entry)
		}
		key := string(entry.Event) + "\x00" + entry.Outcome
		if _, duplicate := seen[key]; duplicate {
			t.Fatalf("fixture repeats event outcome: %#v", entry)
		}
		seen[key] = struct{}{}
	}
}

func TestCollaborationCheckpointEntityTypesMatchFixture(t *testing.T) {
	t.Parallel()
	data, err := os.ReadFile("fixtures/collaboration-checkpoint-entity-types.json")
	if err != nil {
		t.Fatal(err)
	}
	var fixture []string
	if err := json.Unmarshal(data, &fixture); err != nil {
		t.Fatal(err)
	}
	want := []string{
		"post", "page", "work", "release", "label", "artist", "campaign",
		"program_event", "map_theme", "terms_history", "privacy_history", "form",
		"email_layout",
	}
	if len(fixture) != len(want) {
		t.Fatalf("fixture has %d entity types, want %d", len(fixture), len(want))
	}
	for index, entityType := range fixture {
		if entityType != want[index] {
			t.Fatalf("fixture entity type %d = %q, want %q", index, entityType, want[index])
		}
		if _, err := NewCollaborationCheckpointFailedRecord(
			SystemMetadata{OccurredAt: testOccurredAt},
			CollaborationCheckpointContext{EntityType: entityType, EntityID: "entity-1", RetryCount: 4},
			CollaborationCheckpointFailurePersistFailed,
		); err != nil {
			t.Fatalf("entity type %q rejected: %v", entityType, err)
		}
	}
}

func TestTranslationJobTerminalCatalogMatchesFixture(t *testing.T) {
	t.Parallel()
	data, err := os.ReadFile("fixtures/translation-job-terminal-catalog.json")
	if err != nil {
		t.Fatal(err)
	}
	var fixture struct {
		Event          SystemEvent                     `json:"event"`
		Outcomes       []TranslationJobTerminalOutcome `json:"outcomes"`
		EntityTypes    []TranslationEntityType         `json:"entity_types"`
		FailureReasons []TranslationFailureReason      `json:"failure_reasons"`
		ValidJobIDs    []string                        `json:"valid_job_ids"`
		InvalidJobIDs  []string                        `json:"invalid_job_ids"`
	}
	if err := json.Unmarshal(data, &fixture); err != nil {
		t.Fatal(err)
	}
	metadata := SystemMetadata{OccurredAt: testOccurredAt}
	context := TranslationJobTerminalContext{
		JobID: "018f47a2-8a3d-4e17-9d42-6f12c89b1234", EntityType: TranslationEntityPost,
		TargetLocale: "zh-CN", DurationMS: 9,
	}
	if fixture.Event != EventTranslationJobTerminal {
		t.Fatalf("fixture event = %q, want %q", fixture.Event, EventTranslationJobTerminal)
	}
	seenOutcomes := make(map[TranslationJobTerminalOutcome]struct{}, len(fixture.Outcomes))
	for _, outcome := range fixture.Outcomes {
		if _, duplicate := seenOutcomes[outcome]; duplicate {
			t.Fatalf("fixture repeats outcome %q", outcome)
		}
		seenOutcomes[outcome] = struct{}{}
		if _, ok := translationTerminalOutcomes[outcome]; !ok {
			t.Fatalf("fixture outcome does not match catalog: %q", outcome)
		}
		reason := TranslationFailureReason("")
		if outcome == TranslationJobTerminalOutcomeFailed {
			reason = TranslationFailureInternal
		}
		record, err := NewTranslationJobTerminalRecord(metadata, context, outcome, reason)
		if err != nil {
			t.Fatalf("outcome %q rejected: %v", outcome, err)
		}
		if record.Event != fixture.Event || record.Outcome != string(outcome) {
			t.Fatalf("outcome %q = (%q, %q), want (%q, %q)", outcome, record.Event, record.Outcome, fixture.Event, outcome)
		}
	}
	if len(seenOutcomes) != len(translationTerminalOutcomes) {
		t.Fatalf("fixture has %d terminal outcomes, want %d", len(seenOutcomes), len(translationTerminalOutcomes))
	}
	assertExactFixtureSet(t, "translation entity types", fixture.EntityTypes, translationEntityTypes)
	assertExactFixtureSet(t, "translation failure reasons", fixture.FailureReasons, translationFailureReasons)
	for _, jobID := range fixture.ValidJobIDs {
		context.JobID = jobID
		if _, err := NewTranslationJobTerminalRecord(metadata, context, TranslationJobTerminalOutcomeFailed, TranslationFailureInternal); err != nil {
			t.Fatalf("valid translation job ID %q rejected: %v", jobID, err)
		}
	}
	for _, jobID := range fixture.InvalidJobIDs {
		context.JobID = jobID
		if _, err := NewTranslationJobTerminalRecord(metadata, context, TranslationJobTerminalOutcomeFailed, TranslationFailureInternal); err == nil {
			t.Fatalf("invalid translation job ID %q accepted", jobID)
		}
	}
	context.JobID = "018f47a2-8a3d-4e17-9d42-6f12c89b1234"
	for _, entityType := range fixture.EntityTypes {
		context.EntityType = entityType
		if _, err := NewTranslationJobTerminalRecord(metadata, context, TranslationJobTerminalOutcomeFailed, TranslationFailureInternal); err != nil {
			t.Fatalf("entity type %q rejected: %v", entityType, err)
		}
	}
	for _, reason := range fixture.FailureReasons {
		if _, err := NewTranslationJobTerminalRecord(metadata, context, TranslationJobTerminalOutcomeFailed, reason); err != nil {
			t.Fatalf("failure reason %q rejected: %v", reason, err)
		}
	}
}

func assertExactFixtureSet[T comparable](t *testing.T, name string, values []T, expected map[T]struct{}) {
	t.Helper()
	actual := make(map[T]struct{}, len(values))
	for _, value := range values {
		if _, expectedValue := expected[value]; !expectedValue {
			t.Fatalf("fixture %s includes unknown value %v", name, value)
		}
		if _, duplicate := actual[value]; duplicate {
			t.Fatalf("fixture %s repeats value %v", name, value)
		}
		actual[value] = struct{}{}
	}
	if len(actual) != len(expected) {
		t.Fatalf("fixture %s has %d values, want %d", name, len(actual), len(expected))
	}
	for value := range expected {
		if _, present := actual[value]; !present {
			t.Fatalf("fixture %s is missing %v", name, value)
		}
	}
}

func TestSystemBuildersCoverCatalog(t *testing.T) {
	t.Parallel()
	metadata := SystemMetadata{
		OccurredAt:  time.Date(2026, 8, 9, 3, 4, 5, 0, time.UTC),
		Correlation: Correlation{RequestID: "018f47a2-8a3d-4e17-9d42-6f12c89b1234"},
	}
	failure := SystemFailure{Reason: "dependency_unavailable"}
	jobFailure := JobFailure{Reason: JobFailureInternal}
	publish := QueuePublishContext{Queue: "post", MessageID: "message-1", CommandID: "command-1", DurationMS: 4}
	delivery := QueueDeliveryContext{Queue: "post", MessageID: "message-1", CommandID: "command-1", RetryCount: 1, DurationMS: 8}
	handoff := QueueHandoffContext{Queue: "post", MessageID: "message-1", CommandID: "command-1", RetryCount: 2}
	job := JobContext{JobKind: JobKindMeshOptimization, JobID: "job-1"}
	translationTerminal := TranslationJobTerminalContext{
		JobID: "018f47a2-8a3d-4e17-9d42-6f12c89b1234", EntityType: TranslationEntityPost,
		TargetLocale: "en", DurationMS: 10,
	}

	builders := []func() (SystemRecord, error){
		func() (SystemRecord, error) { return NewServiceReadyRecord(metadata, "api") },
		func() (SystemRecord, error) { return NewServiceDegradedRecord(metadata, "api", failure) },
		func() (SystemRecord, error) { return NewServiceStoppingRecord(metadata, "api") },
		func() (SystemRecord, error) { return NewServiceFailedRecord(metadata, "api", failure) },
		func() (SystemRecord, error) {
			return NewDependencyDegradedRecord(metadata, "postgres", "read_queue", failure)
		},
		func() (SystemRecord, error) { return NewDependencyRecoveredRecord(metadata, "postgres", "read_queue") },
		func() (SystemRecord, error) { return NewQueuePublishSucceededRecord(metadata, publish) },
		func() (SystemRecord, error) {
			return NewQueuePublishFailedRecord(metadata, publish, QueueFailureEnqueueFailed)
		},
		func() (SystemRecord, error) { return NewQueueDeliverySucceededRecord(metadata, delivery) },
		func() (SystemRecord, error) {
			return NewQueueDeliveryFailedRecord(metadata, delivery, QueueFailureHandlerFailed)
		},
		func() (SystemRecord, error) {
			return NewQueueDeliveryFailedRecord(metadata, delivery, QueueFailureCompletionFailed)
		},
		func() (SystemRecord, error) {
			return NewQueueDeliveryRequeuedRecord(metadata, delivery, QueueFailureShutdown)
		},
		func() (SystemRecord, error) { return NewQueueRetryAcceptedRecord(metadata, handoff) },
		func() (SystemRecord, error) {
			return NewQueueRetryFailedRecord(metadata, handoff, QueueFailureVisibilityUpdateFailed)
		},
		func() (SystemRecord, error) { return NewQueueDLQAcceptedRecord(metadata, handoff) },
		func() (SystemRecord, error) {
			return NewQueueDLQFailedRecord(metadata, handoff, QueueFailureArchiveFailed)
		},
		func() (SystemRecord, error) { return NewJobStartedRecord(metadata, job) },
		func() (SystemRecord, error) { return NewJobSucceededRecord(metadata, job, 10) },
		func() (SystemRecord, error) { return NewJobFailedRecord(metadata, job, 10, jobFailure) },
		func() (SystemRecord, error) {
			return NewTranslationJobTerminalRecord(metadata, translationTerminal, TranslationJobTerminalOutcomeApplied, "")
		},
		func() (SystemRecord, error) {
			return NewDomainAuditAppendFailedRecord(metadata, AuditPostUpdated, AuditAppendFailurePersistenceFailed)
		},
		func() (SystemRecord, error) {
			return NewCollaborationCheckpointFailedRecord(metadata, CollaborationCheckpointContext{EntityType: "post", EntityID: "post-1", RetryCount: 4}, CollaborationCheckpointFailurePersistFailed)
		},
		func() (SystemRecord, error) {
			return NewClientRenderFailedRecord(metadata, ClientRenderComponentGeneral)
		},
		func() (SystemRecord, error) {
			return NewTelemetryPipelineDegradedRecord(metadata, "otlp_exporter", failure)
		},
		func() (SystemRecord, error) { return NewTelemetryPipelineRecoveredRecord(metadata, "otlp_exporter") },
	}
	seen := make(map[SystemEvent]struct{}, len(builders))
	for _, build := range builders {
		record, err := build()
		if err != nil {
			t.Fatal(err)
		}
		seen[record.Event] = struct{}{}
		if record.Event != EventTranslationJobTerminal && record.Outcome != systemEventOutcomes[record.Event] {
			t.Fatalf("outcome %q does not match %s", record.Outcome, record.Event)
		}
	}
	if len(seen) != len(systemEvents) {
		t.Fatalf("builders covered %d of %d events", len(seen), len(systemEvents))
	}
}

func TestTerminalCollaborationCheckpointAndClientRenderCatalogs(t *testing.T) {
	t.Parallel()
	metadata := SystemMetadata{OccurredAt: testOccurredAt}
	if _, err := NewCollaborationCheckpointFailedRecord(metadata, CollaborationCheckpointContext{EntityType: "post", EntityID: "post-1", RetryCount: 4}, CollaborationCheckpointFailurePersistFailed); err != nil {
		t.Fatal(err)
	}
	for _, entityType := range []string{"form", "email_layout"} {
		if _, err := NewCollaborationCheckpointFailedRecord(metadata, CollaborationCheckpointContext{EntityType: entityType, EntityID: "entity-1", RetryCount: 4}, CollaborationCheckpointFailurePersistFailed); err != nil {
			t.Fatalf("%s checkpoint rejected: %v", entityType, err)
		}
	}
	if _, err := NewCollaborationCheckpointFailedRecord(metadata, CollaborationCheckpointContext{EntityType: "post", RetryCount: 4}, CollaborationCheckpointFailurePersistFailed); err == nil {
		t.Fatal("empty checkpoint entity accepted")
	}
	if _, err := NewCollaborationCheckpointFailedRecord(metadata, CollaborationCheckpointContext{EntityType: "post", EntityID: "post-1", RetryCount: 1}, CollaborationCheckpointFailureSharedStructureChanged); err != nil {
		t.Fatal(err)
	}
	if _, err := NewCollaborationCheckpointFailedRecord(metadata, CollaborationCheckpointContext{EntityType: "post", EntityID: "post-1", RetryCount: 2}, CollaborationCheckpointFailureSharedStructureChanged); err == nil {
		t.Fatal("retried collaboration conflict accepted")
	}
	if _, err := NewCollaborationCheckpointFailedRecord(metadata, CollaborationCheckpointContext{EntityType: "post", EntityID: "post-1", RetryCount: 1}, CollaborationCheckpointFailureTargetRevisionChanged); err != nil {
		t.Fatalf("target revision conflict rejected: %v", err)
	}
	if _, err := NewCollaborationCheckpointFailedRecord(metadata, CollaborationCheckpointContext{EntityType: "unknown", EntityID: "post-1", RetryCount: 4}, CollaborationCheckpointFailurePersistFailed); err == nil {
		t.Fatal("unknown checkpoint entity accepted")
	}
	if _, err := NewClientRenderFailedRecord(metadata, ClientRenderComponentGlobal); err != nil {
		t.Fatal(err)
	}
	if _, err := NewClientRenderFailedRecord(metadata, "other"); err == nil {
		t.Fatal("unknown client render component accepted")
	}
}

func TestSystemBuildersRejectUnboundedAndIncompleteValues(t *testing.T) {
	t.Parallel()
	metadata := SystemMetadata{OccurredAt: time.Now().UTC()}
	if _, err := NewServiceReadyRecord(metadata, "API Server"); err == nil {
		t.Fatal("unbounded component accepted")
	}
	if _, err := NewQueuePublishFailedRecord(metadata, QueuePublishContext{}, QueueFailureEnqueueFailed); err == nil {
		t.Fatal("incomplete queue failure accepted")
	}
	if _, err := NewJobStartedRecord(metadata, JobContext{JobKind: "projection", JobID: "job-1"}); err == nil {
		t.Fatal("unknown job kind accepted")
	}
	if _, err := NewJobFailedRecord(metadata, JobContext{JobKind: JobKindMeshOptimization, JobID: "job-1"}, 1, JobFailure{Reason: "failed"}); err == nil {
		t.Fatal("unknown job failure reason accepted")
	}
	terminal := TranslationJobTerminalContext{JobID: "018f47a2-8a3d-4e17-9d42-6f12c89b1234", EntityType: TranslationEntityPost, TargetLocale: "en", DurationMS: 1}
	if _, err := NewTranslationJobTerminalRecord(metadata, terminal, "unknown", ""); err == nil {
		t.Fatal("unknown terminal status accepted")
	}
	if _, err := NewTranslationJobTerminalRecord(metadata, terminal, TranslationJobTerminalOutcomeFailed, "unknown"); err == nil {
		t.Fatal("unknown translation error classification accepted")
	}
	if _, err := NewTranslationJobTerminalRecord(metadata, terminal, TranslationJobTerminalOutcomeApplied, TranslationFailureInternal); err == nil {
		t.Fatal("applied terminal event accepted an error classification")
	}
	if _, err := NewTranslationJobTerminalRecord(metadata, terminal, TranslationJobTerminalOutcomeCancelled, TranslationFailureInternal); err == nil {
		t.Fatal("cancelled terminal event accepted an error classification")
	}
	incompleteTerminal := terminal
	incompleteTerminal.TargetLocale = ""
	if _, err := NewTranslationJobTerminalRecord(metadata, incompleteTerminal, TranslationJobTerminalOutcomeFailed, TranslationFailureInternal); err == nil {
		t.Fatal("translation terminal event accepted a missing target locale")
	}
	unknownEntityTerminal := terminal
	unknownEntityTerminal.EntityType = TranslationEntityType("unknown")
	if _, err := NewTranslationJobTerminalRecord(metadata, unknownEntityTerminal, TranslationJobTerminalOutcomeFailed, TranslationFailureInternal); err == nil {
		t.Fatal("translation terminal event accepted an unknown entity type")
	}
	terminalRecord, err := NewTranslationJobTerminalRecord(metadata, terminal, TranslationJobTerminalOutcomeFailed, TranslationFailureInternal)
	if err != nil {
		t.Fatal(err)
	}
	terminalRecord.EntityID = "provider-document-id"
	if err := terminalRecord.Validate(); err == nil {
		t.Fatal("translation terminal record accepted entity_id")
	}
	terminal.JobID = "job-1"
	if _, err := NewTranslationJobTerminalRecord(metadata, terminal, TranslationJobTerminalOutcomeFailed, TranslationFailureInternal); err == nil {
		t.Fatal("non-canonical translation job ID accepted")
	}
	terminal.JobID, terminal.TargetLocale = "018f47a2-8a3d-4e17-9d42-6f12c89b1234", "bad_locale!"
	if _, err := NewTranslationJobTerminalRecord(metadata, terminal, TranslationJobTerminalOutcomeFailed, TranslationFailureInternal); err == nil {
		t.Fatal("invalid target locale accepted")
	}
}

func TestDeliveryRequeueReasonCatalog(t *testing.T) {
	t.Parallel()
	metadata := SystemMetadata{OccurredAt: testOccurredAt}
	delivery := QueueDeliveryContext{Queue: "post", MessageID: "message-1", CommandID: "command-1", RetryCount: 1, DurationMS: 8}
	for _, reason := range []QueueFailureReason{QueueFailureShutdown, QueueFailureHandlerFailed} {
		if _, err := NewQueueDeliveryRequeuedRecord(metadata, delivery, reason); err != nil {
			t.Fatalf("requeue reason %q rejected: %v", reason, err)
		}
	}
	if _, err := NewQueueDeliveryRequeuedRecord(metadata, delivery, QueueFailureArchiveFailed); err == nil {
		t.Fatal("non-catalog requeue reason accepted")
	}
	handoff := QueueHandoffContext{Queue: "post", MessageID: "message-1", CommandID: "command-1", RetryCount: 1}
	if _, err := NewQueueRetryFailedRecord(metadata, handoff, QueueFailureArchiveFailed); err == nil {
		t.Fatal("archive failure accepted as visibility update failure")
	}
}

func TestTranslationTerminalRejectsUnsupportedSystemFields(t *testing.T) {
	t.Parallel()
	base, err := NewTranslationJobTerminalRecord(
		SystemMetadata{OccurredAt: testOccurredAt},
		TranslationJobTerminalContext{
			JobID: "018f47a2-8a3d-4e17-9d42-6f12c89b1234", EntityType: TranslationEntityPost,
			TargetLocale: "en", DurationMS: 1,
		},
		TranslationJobTerminalOutcomeFailed,
		TranslationFailureInternal,
	)
	if err != nil {
		t.Fatal(err)
	}
	retryCount := 1
	for _, test := range []struct {
		name   string
		mutate func(*SystemRecord)
	}{
		{"component", func(record *SystemRecord) { record.Component = "api" }},
		{"dependency", func(record *SystemRecord) { record.Dependency = "postgres" }},
		{"operation", func(record *SystemRecord) { record.Operation = "read_queue" }},
		{"queue", func(record *SystemRecord) { record.Queue = "translation" }},
		{"message ID", func(record *SystemRecord) { record.MessageID = "message-1" }},
		{"command ID", func(record *SystemRecord) { record.CommandID = "command-1" }},
		{"retry count", func(record *SystemRecord) { record.RetryCount = &retryCount }},
		{"job kind", func(record *SystemRecord) { record.JobKind = "translation" }},
		{"entity ID", func(record *SystemRecord) { record.EntityID = "provider-document-id" }},
		{"record class", func(record *SystemRecord) { record.RecordClass = AuditRecordClassDomain }},
		{"action", func(record *SystemRecord) { record.Action = "post.updated" }},
		{"error code", func(record *SystemRecord) { record.ErrorCode = "provider_error" }},
		{"reason", func(record *SystemRecord) { record.Reason = "provider_error" }},
	} {
		t.Run(test.name, func(t *testing.T) {
			record := base
			test.mutate(&record)
			if err := record.Validate(); err == nil {
				t.Fatalf("translation terminal record accepted %s: %#v", test.name, record)
			}
		})
	}
}

func TestSystemCatalogRejectsAmbiguousRecords(t *testing.T) {
	t.Parallel()
	metadata := SystemMetadata{OccurredAt: testOccurredAt}
	publish := QueuePublishContext{
		Queue: "post", MessageID: "message-1", CommandID: "command-1", DurationMS: 4,
	}
	delivery := QueueDeliveryContext{
		Queue: "post", MessageID: "message-1", CommandID: "command-1", RetryCount: 1, DurationMS: 8,
	}
	handoff := QueueHandoffContext{Queue: "post", MessageID: "message-1", CommandID: "command-1", RetryCount: 2}

	ready, _ := NewServiceReadyRecord(metadata, "api")
	publishFailed, _ := NewQueuePublishFailedRecord(metadata, publish, QueueFailureEnqueueFailed)
	deliverySucceeded, _ := NewQueueDeliverySucceededRecord(metadata, delivery)
	deliveryFailed, _ := NewQueueDeliveryFailedRecord(metadata, delivery, QueueFailureHandlerFailed)
	deliveryRequeued, _ := NewQueueDeliveryRequeuedRecord(metadata, delivery, QueueFailureShutdown)
	retryFailed, _ := NewQueueRetryFailedRecord(metadata, handoff, QueueFailureVisibilityUpdateFailed)
	domainAppendFailed, _ := NewDomainAuditAppendFailedRecord(metadata, AuditPostUpdated, AuditAppendFailurePersistenceFailed)
	securityAppendFailed, err := NewSecurityAccessAppendFailedRecord(metadata, SecurityAuthorizationDenied, AuditAppendFailurePersistenceFailed)
	if err != nil {
		t.Fatal(err)
	}

	tests := []struct {
		name   string
		record SystemRecord
		mutate func(*SystemRecord)
	}{
		{name: "outcome mismatch", record: ready, mutate: func(record *SystemRecord) { record.Outcome = "failed" }},
		{name: "failed queue publish has no queue", record: publishFailed, mutate: func(record *SystemRecord) { record.Queue = "" }},
		{name: "failed publish has no failure", record: publishFailed, mutate: func(record *SystemRecord) { record.Reason = "" }},
		{name: "failed delivery has no failure", record: deliveryFailed, mutate: func(record *SystemRecord) { record.Reason = "" }},
		{name: "requeued delivery has no queue", record: deliveryRequeued, mutate: func(record *SystemRecord) { record.Queue = "" }},
		{name: "requeued delivery has no duration", record: deliveryRequeued, mutate: func(record *SystemRecord) { record.DurationMS = nil }},
		{name: "retry failure has no failure", record: retryFailed, mutate: func(record *SystemRecord) { record.Reason = "" }},
		{name: "successful delivery has failure", record: deliverySucceeded, mutate: func(record *SystemRecord) { record.Reason = "handler_failed" }},
		{name: "queue failure has error code", record: deliveryFailed, mutate: func(record *SystemRecord) { record.ErrorCode = "broker_error" }},
		{name: "queue failure has unknown reason", record: deliveryFailed, mutate: func(record *SystemRecord) { record.Reason = "unknown" }},
		{name: "domain append has unknown action", record: domainAppendFailed, mutate: func(record *SystemRecord) { record.Action = "post.version.created" }},
		{name: "security append has unknown action", record: securityAppendFailed, mutate: func(record *SystemRecord) { record.Action = "session.revoked" }},
		{name: "append has unknown class", record: domainAppendFailed, mutate: func(record *SystemRecord) { record.RecordClass = "other" }},
		{name: "append has error code", record: domainAppendFailed, mutate: func(record *SystemRecord) { record.ErrorCode = "database_error" }},
		{name: "append has unknown reason", record: domainAppendFailed, mutate: func(record *SystemRecord) { record.Reason = "unknown" }},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			record := test.record
			test.mutate(&record)
			if err := record.Validate(); err == nil {
				t.Fatalf("Validate() accepted %#v", record)
			}
		})
	}
}

func TestEmitSystemUsesErrorLevelForFailure(t *testing.T) {
	t.Parallel()
	var output bytes.Buffer
	handler := slog.NewJSONHandler(&output, nil)
	record, err := NewServiceFailedRecord(
		SystemMetadata{OccurredAt: testOccurredAt},
		"api",
		SystemFailure{Reason: "startup_failed"},
	)
	if err != nil {
		t.Fatal(err)
	}
	if err := EmitSystem(context.Background(), handler, record); err != nil {
		t.Fatal(err)
	}
	if !bytes.Contains(output.Bytes(), []byte(`"level":"ERROR"`)) {
		t.Fatalf("system output = %s", output.Bytes())
	}
}

func TestEmitSystemSerializesTranslationDimensions(t *testing.T) {
	t.Parallel()
	var output bytes.Buffer
	handler := slog.NewJSONHandler(&output, nil)
	record, err := NewTranslationJobTerminalRecord(
		SystemMetadata{OccurredAt: testOccurredAt},
		TranslationJobTerminalContext{
			JobID: "018f47a2-8a3d-4e17-9d42-6f12c89b1234", EntityType: TranslationEntityPost,
			TargetLocale: "pt-BR", DurationMS: 17,
		},
		TranslationJobTerminalOutcomeFailed,
		TranslationFailureProviderUnavailable,
	)
	if err != nil {
		t.Fatal(err)
	}
	if err := EmitSystem(context.Background(), handler, record); err != nil {
		t.Fatal(err)
	}
	var emitted map[string]any
	if err := json.Unmarshal(output.Bytes(), &emitted); err != nil {
		t.Fatal(err)
	}
	for key, want := range map[string]string{
		"event": "translation.job.terminal", "domain": "translation", "job_id": "018f47a2-8a3d-4e17-9d42-6f12c89b1234",
		"entity_type": "post", "target_locale": "pt-BR", "outcome": "failed", "error_classification": "provider_unavailable",
	} {
		if emitted[key] != want {
			t.Fatalf("emitted[%q] = %#v, want %q", key, emitted[key], want)
		}
	}
	if _, exists := emitted["entity_id"]; exists {
		t.Fatalf("translation terminal event exposed entity_id: %#v", emitted)
	}
	if _, exists := emitted["reason"]; exists {
		t.Fatalf("translation terminal event exposed generic reason: %#v", emitted)
	}
}

func TestEmitSystemSerializesExistingEntityDimensions(t *testing.T) {
	t.Parallel()
	var output bytes.Buffer
	handler := slog.NewJSONHandler(&output, nil)
	record, err := NewServiceReadyRecord(SystemMetadata{OccurredAt: testOccurredAt}, "api")
	if err != nil {
		t.Fatal(err)
	}
	record.EntityType, record.EntityID, record.TargetLocale = "post", "post-1", "en"
	if err := EmitSystem(context.Background(), handler, record); err != nil {
		t.Fatal(err)
	}
	var emitted map[string]any
	if err := json.Unmarshal(output.Bytes(), &emitted); err != nil {
		t.Fatal(err)
	}
	for key, want := range map[string]string{"entity_type": "post", "entity_id": "post-1", "target_locale": "en"} {
		if emitted[key] != want {
			t.Fatalf("emitted[%q] = %#v, want %q", key, emitted[key], want)
		}
	}
}
