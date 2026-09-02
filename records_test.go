package telemetry

import (
	"context"
	"encoding/json"
	"os"
	"reflect"
	"strings"
	"testing"
	"time"

	"go.opentelemetry.io/otel/trace"
)

var testOccurredAt = time.Date(2026, 8, 9, 3, 4, 5, 0, time.UTC)

func validRequestRecord() RequestRecord {
	return RequestRecord{
		Event:       "request.completed",
		OccurredAt:  testOccurredAt,
		Correlation: Correlation{RequestID: testRequestID},
		RecordActor: RecordActor{Kind: ActorKindAnonymous},
		HTTPMethod:  "GET",
		HTTPRoute:   "/posts/{post_id}",
		StatusCode:  200,
		DurationMS:  4,
		Outcome:     RequestOutcomeSucceeded,
	}
}

func TestRequestRecordValidate(t *testing.T) {
	t.Parallel()
	record := validRequestRecord()
	if err := record.Validate(); err != nil {
		t.Fatal(err)
	}
	rpc := record
	rpc.HTTPMethod, rpc.HTTPRoute = "POST", ""
	rpc.RPCService, rpc.RPCMethod = "geul.v1.PostService", "UpdatePost"
	if err := rpc.Validate(); err != nil {
		t.Fatal(err)
	}

	tests := []struct {
		name   string
		mutate func(*RequestRecord)
	}{
		{name: "event", mutate: func(value *RequestRecord) { value.Event = "request.started" }},
		{name: "time", mutate: func(value *RequestRecord) { value.OccurredAt = time.Time{} }},
		{name: "request ID", mutate: func(value *RequestRecord) { value.RequestID = "bad" }},
		{name: "trace", mutate: func(value *RequestRecord) { value.TraceID = "abc" }},
		{name: "actor", mutate: func(value *RequestRecord) { value.RecordActor = RecordActor{Kind: "bad"} }},
		{name: "boundary", mutate: func(value *RequestRecord) { value.HTTPRoute = "" }},
		{name: "status low", mutate: func(value *RequestRecord) { value.StatusCode = 99 }},
		{name: "status high", mutate: func(value *RequestRecord) { value.StatusCode = 600 }},
		{name: "duration", mutate: func(value *RequestRecord) { value.DurationMS = -1 }},
		{name: "outcome", mutate: func(value *RequestRecord) { value.Outcome = "unknown" }},
		{name: "success error", mutate: func(value *RequestRecord) { value.ErrorCode = "internal" }},
		{name: "blocked without detail", mutate: func(value *RequestRecord) { value.Outcome = RequestOutcomeBlocked }},
		{name: "wrong HTTP outcome", mutate: func(value *RequestRecord) {
			value.StatusCode, value.Outcome, value.Reason = 403, RequestOutcomeFailed, string(RequestReasonClientError)
		}},
		{name: "unknown HTTP reason", mutate: func(value *RequestRecord) {
			value.StatusCode, value.Outcome, value.Reason = 500, RequestOutcomeFailed, "unexpected"
		}},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			invalid := record
			test.mutate(&invalid)
			if err := invalid.Validate(); err == nil {
				t.Fatalf("Validate() accepted %#v", invalid)
			}
		})
	}
}

func TestSystemRecordValidate(t *testing.T) {
	t.Parallel()
	zero, negative := 0, -1
	duration := int64(0)
	for event := range systemEvents {
		record := SystemRecord{
			Event: event, OccurredAt: testOccurredAt, Component: "api", Dependency: "postgres", Operation: "connect",
			Queue: "post", MessageID: "message-1", CommandID: "command-1",
			RetryCount: &zero, DurationMS: &duration, JobKind: "mesh_optimization", JobID: "job-1",
			RecordClass: "domain_audit", Action: "post.updated", Outcome: systemEventOutcomes[event], ErrorCode: "test_error",
		}
		switch event {
		case EventQueuePublishSucceeded, EventQueueDeliverySucceeded, EventQueueRetryAccepted, EventQueueDLQAccepted:
			record.ErrorCode = ""
		case EventQueuePublishFailed:
			record.ErrorCode, record.Reason = "", string(QueueFailureEnqueueFailed)
		case EventQueueDeliveryFailed:
			record.ErrorCode, record.Reason = "", string(QueueFailureHandlerFailed)
		case EventQueueDeliveryRequeued:
			record.ErrorCode, record.Reason = "", string(QueueFailureShutdown)
		case EventQueueRetryFailed:
			record.ErrorCode, record.Reason = "", string(QueueFailureVisibilityUpdateFailed)
		case EventQueueDLQFailed:
			record.ErrorCode, record.Reason = "", string(QueueFailureArchiveFailed)
		case EventJobFailed:
			record.ErrorCode, record.Reason = "", string(JobFailureInternal)
		case EventAuditAppendFailed:
			record.ErrorCode, record.Reason = "", string(AuditAppendFailurePersistenceFailed)
		case EventCollaborationCheckpointFailed:
			record.ErrorCode, record.Domain, record.EntityType, record.EntityID, record.Reason = "", "collaboration", "post", "post-1", string(CollaborationCheckpointFailurePersistFailed)
		case EventClientRenderFailed:
			record.ErrorCode, record.Domain, record.Component, record.Reason = "", "client", "general", "react_error_boundary"
		case EventTranslationJobTerminal:
			record = SystemRecord{
				Event: event, OccurredAt: testOccurredAt, Domain: "translation", JobID: testRequestID,
				EntityType: "post", TargetLocale: "en", DurationMS: &duration, Outcome: string(TranslationJobTerminalOutcomeFailed),
			}
			record.ErrorClassification = string(TranslationFailureInternal)
		}
		if err := record.Validate(); err != nil {
			t.Fatalf("Validate(%s) unexpected error = %v", event, err)
		}
	}
	tests := []SystemRecord{
		{Event: "unknown", OccurredAt: testOccurredAt},
		{Event: EventServiceReady},
		{Event: EventServiceReady, Outcome: "ready", OccurredAt: testOccurredAt, Correlation: Correlation{RequestID: "bad"}},
		{Event: EventServiceReady, Outcome: "ready", OccurredAt: testOccurredAt, Correlation: Correlation{TraceID: "bad"}},
		{Event: EventServiceReady, Outcome: "ready", OccurredAt: testOccurredAt, EntityID: strings.Repeat("x", 129)},
		{Event: EventServiceReady, Outcome: "ready", OccurredAt: testOccurredAt, RetryCount: &negative},
		{Event: EventServiceReady, Outcome: "ready", OccurredAt: testOccurredAt, DurationMS: func() *int64 { value := int64(-1); return &value }()},
		{Event: EventServiceReady, Outcome: "ready", OccurredAt: testOccurredAt},
		{Event: EventServiceFailed, Outcome: "failed", OccurredAt: testOccurredAt, Component: "api"},
		{Event: EventQueuePublishSucceeded, Outcome: "succeeded", OccurredAt: testOccurredAt, Queue: "post", MessageID: "message-1", CommandID: "command-1"},
		{Event: EventQueuePublishSucceeded, Outcome: "succeeded", OccurredAt: testOccurredAt, DurationMS: &duration},
		{Event: EventQueueDeliverySucceeded, Outcome: "succeeded", OccurredAt: testOccurredAt, Queue: "post", MessageID: "message-1", CommandID: "command-1"},
		{Event: EventQueueRetryAccepted, Outcome: "accepted", OccurredAt: testOccurredAt, Queue: "post", MessageID: "message-1", CommandID: "command-1"},
		{Event: EventJobStarted, Outcome: "started", OccurredAt: testOccurredAt, JobID: "job-1"},
		{Event: EventJobSucceeded, Outcome: "succeeded", OccurredAt: testOccurredAt, JobKind: "projection", JobID: "job-1", DurationMS: &duration},
		{Event: EventJobSucceeded, Outcome: "succeeded", OccurredAt: testOccurredAt, JobKind: "mesh_optimization", JobID: "job-1"},
		{Event: EventJobFailed, Outcome: "failed", OccurredAt: testOccurredAt, JobKind: "projection", JobID: "job-1", DurationMS: &duration, Reason: "internal"},
		{Event: EventJobFailed, Outcome: "failed", OccurredAt: testOccurredAt, JobKind: "mesh_optimization", JobID: "job-1", DurationMS: &duration},
		{Event: EventCollaborationCheckpointFailed, Outcome: "failed", OccurredAt: testOccurredAt, Domain: "client", EntityType: "post", EntityID: "post-1", RetryCount: &zero, Reason: "persist_failed"},
		{Event: EventCollaborationCheckpointFailed, Outcome: "failed", OccurredAt: testOccurredAt, Domain: "collaboration", EntityType: "post", EntityID: "post-1", RetryCount: &zero, Reason: "unknown"},
		{Event: EventClientRenderFailed, Outcome: "failed", OccurredAt: testOccurredAt, Reason: "react_error_boundary"},
		{Event: EventClientRenderFailed, Outcome: "failed", OccurredAt: testOccurredAt, Domain: "server", Component: "general", Reason: "react_error_boundary"},
		{Event: EventTranslationJobTerminal, Outcome: "applied", OccurredAt: testOccurredAt, Domain: "translation", JobID: testRequestID, EntityType: "post", TargetLocale: "en"},
		{Event: EventTranslationJobTerminal, Outcome: "failed", OccurredAt: testOccurredAt, Domain: "translation", JobID: testRequestID, EntityType: "post", TargetLocale: "en", DurationMS: &duration},
		{Event: EventTranslationJobTerminal, Outcome: "failed", OccurredAt: testOccurredAt, Domain: "translation", JobID: testRequestID, EntityType: "post", TargetLocale: "en", DurationMS: &duration, ErrorClassification: "unknown"},
		{Event: EventTranslationJobTerminal, Outcome: "applied", OccurredAt: testOccurredAt, Domain: "translation", JobID: testRequestID, EntityType: "post", TargetLocale: "en", DurationMS: &duration, ErrorClassification: "internal"},
		{Event: EventTranslationJobTerminal, Outcome: "cancelled", OccurredAt: testOccurredAt, Domain: "translation", JobID: testRequestID, EntityType: "post", TargetLocale: "en", DurationMS: &duration, Reason: "internal"},
		{Event: EventServiceFailed, Outcome: "failed", OccurredAt: testOccurredAt, Component: "api", Reason: "startup_failed", ErrorClassification: "internal"},
	}
	full := SystemRecord{
		OccurredAt: testOccurredAt, Component: "api", Dependency: "postgres", Operation: "connect",
		Queue: "post", MessageID: "message-1", CommandID: "command-1",
		RetryCount: &zero, DurationMS: &duration, JobKind: "mesh_optimization", JobID: "job-1",
		RecordClass: "domain_audit", Action: "post.updated", ErrorCode: "test_error",
	}
	for _, mutation := range []func(*SystemRecord){
		func(value *SystemRecord) {
			value.Event = EventServiceDegraded
			value.Outcome = "degraded"
			value.Component = ""
		},
		func(value *SystemRecord) {
			value.Event = EventDependencyDegraded
			value.Outcome = "degraded"
			value.Dependency = ""
		},
		func(value *SystemRecord) {
			value.Event = EventQueuePublishFailed
			value.Outcome = "failed"
			value.DurationMS = nil
		},
		func(value *SystemRecord) {
			value.Event = EventQueueDeliverySucceeded
			value.Outcome = "succeeded"
			value.Queue = ""
		},
		func(value *SystemRecord) {
			value.Event = EventQueueDeliveryFailed
			value.Outcome = "failed"
			value.Queue = ""
		},
		func(value *SystemRecord) {
			value.Event = EventQueueDeliveryFailed
			value.Outcome = "failed"
			value.RetryCount = nil
		},
		func(value *SystemRecord) {
			value.Event = EventQueueRetryAccepted
			value.Outcome = "accepted"
			value.Queue = ""
		},
		func(value *SystemRecord) {
			value.Event = EventQueueRetryFailed
			value.Outcome = "failed"
			value.Queue = ""
		},
		func(value *SystemRecord) {
			value.Event = EventQueueRetryFailed
			value.Outcome = "failed"
			value.RetryCount = nil
		},
		func(value *SystemRecord) {
			value.Event = EventJobSucceeded
			value.Outcome = "succeeded"
			value.JobKind = ""
		},
		func(value *SystemRecord) { value.Event = EventJobFailed; value.Outcome = "failed"; value.JobKind = "" },
		func(value *SystemRecord) {
			value.Event = EventJobFailed
			value.Outcome = "failed"
			value.DurationMS = nil
		},
		func(value *SystemRecord) {
			value.Event = EventAuditAppendFailed
			value.Outcome = "failed"
			value.RecordClass = ""
		},
		func(value *SystemRecord) {
			value.Event = EventTelemetryPipelineDegraded
			value.Outcome = "degraded"
			value.Component = ""
		},
	} {
		invalid := full
		mutation(&invalid)
		tests = append(tests, invalid)
	}
	for _, invalid := range tests {
		if err := invalid.Validate(); err == nil {
			t.Fatalf("Validate() accepted %#v", invalid)
		}
	}
}

func TestAuditRecordValidate(t *testing.T) {
	t.Parallel()
	memberMetadata := AuditMetadata{AuditID: testRequestID, OccurredAt: testOccurredAt, RecordActor: RecordActor{Kind: ActorKindMember, MemberID: "member-1"}}
	systemMetadata := memberMetadata
	systemMetadata.RecordActor = RecordActor{Kind: ActorKindSystem, Service: "geul-backend"}
	collabMetadata := memberMetadata
	collabMetadata.RecordActor = RecordActor{Kind: ActorKindSystem, Service: string(ServiceEditorCollab)}
	contributors := []string{"1b6bcad2-c90d-49e9-bec7-f9a4ba6b2894", "7a7a8fd4-1f69-4e9a-9dc2-2378926ff351"}
	must := func(record AuditRecord, err error) AuditRecord {
		if err != nil {
			t.Fatal(err)
		}
		return record
	}
	valid := []AuditRecord{
		must(NewSiteSettingsUpdatedAuditRecord(memberMetadata, []string{"primary_color", "site_title"})),
		must(NewMemberOnboardingCompletedAuditRecord(memberMetadata, "member-1", "Member")),
		must(NewMemberRoleUpdatedAuditRecord(memberMetadata, "member-1", "user", "author")),
		must(NewMemberRoleUpdatedAuditRecord(systemMetadata, "member-1", "user", "admin")),
		must(NewMemberBannedAuditRecord(memberMetadata, "member-1")),
		must(NewMemberUnbannedAuditRecord(systemMetadata, "member-1")),
		must(NewPostVersionCreatedAuditRecord(collabMetadata, "post-1", "version-1", contributors)),
		must(NewPageVersionCreatedAuditRecord(collabMetadata, "page-1", "version-1", contributors)),
		must(NewWorkVersionCreatedAuditRecord(collabMetadata, "work-1", "version-1", contributors)),
		must(NewPostCreatedAuditRecord(memberMetadata, "post-1")),
		must(NewPageCreatedAuditRecord(memberMetadata, "page-1")),
		must(NewWorkCreatedAuditRecord(memberMetadata, "work-1")),
		must(NewPostDeletedAuditRecord(memberMetadata, "post-1")),
		must(NewPageDeletedAuditRecord(memberMetadata, "page-1")),
		must(NewWorkDeletedAuditRecord(memberMetadata, "work-1")),
		must(NewAccountCanonicalEmailUpdatedAuditRecord(memberMetadata, "member-1", "old@example.test", "new@example.test")),
		must(NewAccountEmailLoginAddedAuditRecord(memberMetadata, "member-1", "added@example.test")),
		must(NewAccountEmailLoginRemovedAuditRecord(memberMetadata, "member-1", "removed@example.test")),
		must(NewAccountSocialLoginAddedAuditRecord(memberMetadata, "member-1", "google", "subject-1")),
		must(NewAccountSocialLoginRemovedAuditRecord(memberMetadata, "member-1", "github", "subject-2")),
		must(NewAccountPasskeyAddedAuditRecord(memberMetadata, "member-1", []string{"passkey-1"})),
		must(NewAccountPasskeyRemovedAuditRecord(memberMetadata, "member-1", []string{"passkey-2"})),
		must(NewAccountSessionRevokedAuditRecord(memberMetadata, "member-1", AccountSessionScopeOne, []string{testRequestID})),
		must(NewAccountDeletionRequestedAuditRecord(memberMetadata, "member-1", AuditStateNone)),
		must(NewAccountDeletionRequestedAuditRecord(memberMetadata, "member-1", AuditStateCancelled)),
		must(NewAccountDeletionRequestedAuditRecord(memberMetadata, "member-1", AuditStateRecovered)),
		must(NewAccountDeletionScheduledAuditRecord(memberMetadata, "member-1", AuditStateConfirmationPending)),
		must(NewAccountDeletionScheduledAuditRecord(memberMetadata, "member-1", AuditStateNone)),
		must(NewAccountDeletionScheduledAuditRecord(memberMetadata, "member-1", AuditStateCancelled)),
		must(NewAccountDeletionScheduledAuditRecord(memberMetadata, "member-1", AuditStateRecovered)),
		must(NewAccountDeletionCancelledAuditRecord(memberMetadata, "member-1")),
		must(NewAccountDeletionRecoveredAuditRecord(memberMetadata, "member-1")),
		must(NewAccountDeletedAuditRecord(systemMetadata, "member-1")),
	}
	for _, record := range valid {
		if err := record.Validate(); err != nil {
			t.Fatalf("Validate() unexpected error = %v for %#v", err, record)
		}
	}

	base := valid[0]
	invalid := []AuditRecord{
		{},
		func() AuditRecord { value := base; value.AuditID = "audit-1"; return value }(),
		func() AuditRecord {
			value := base
			value.AuditID = "018f47a2-8a3d-4e17-7d42-6f12c89b1234"
			return value
		}(),
		func() AuditRecord { value := base; value.OccurredAt = time.Time{}; return value }(),
		func() AuditRecord { value := base; value.RecordActor = RecordActor{Kind: "bad"}; return value }(),
		func() AuditRecord {
			value := base
			value.RecordActor = RecordActor{Kind: ActorKindAnonymous}
			return value
		}(),
		func() AuditRecord { value := base; value.RequestID = "bad"; return value }(),
		func() AuditRecord { value := base; value.TraceID = "bad"; return value }(),
		func() AuditRecord { value := base; value.Action = "member.role.updated"; return value }(),
		func() AuditRecord { value := base; value.TargetType = "page"; return value }(),
		func() AuditRecord { value := base; value.TargetID = "2"; return value }(),
		func() AuditRecord { value := base; value.ChangedFields = nil; return value }(),
		func() AuditRecord {
			value := base
			value.ChangedFields = []string{"site_title", "primary_color"}
			return value
		}(),
		func() AuditRecord { value := base; value.ChangedFields = []string{"bad-value"}; return value }(),
		func() AuditRecord { value := base; value.VersionID = "version-1"; return value }(),
		func() AuditRecord { value := valid[1]; value.Nickname = " "; return value }(),
		func() AuditRecord {
			value := valid[1]
			value.ChangedFields = []string{"nickname", "wrong"}
			return value
		}(),
		func() AuditRecord {
			value := valid[1]
			value.ChangedFields = []string{"onboarded", "nickname"}
			return value
		}(),
		func() AuditRecord { value := valid[1]; value.TargetType = "account"; return value }(),
		func() AuditRecord { value := valid[2]; value.NewRole = "user"; return value }(),
		func() AuditRecord { value := valid[2]; value.ChangedFields = []string{"status"}; return value }(),
		func() AuditRecord { value := valid[2]; value.ChangedFields = []string{"bad-value"}; return value }(),
		func() AuditRecord { value := valid[4]; value.NewState = AuditStateActive; return value }(),
		func() AuditRecord { value := valid[4]; value.NewState = AuditStateRecovered; return value }(),
		func() AuditRecord { value := valid[4]; value.ChangedFields = []string{"unknown"}; return value }(),
		func() AuditRecord { value := valid[4]; value.RecordActor = systemMetadata.RecordActor; return value }(),
		func() AuditRecord { value := valid[6]; value.ChangedFields = nil; return value }(),
		func() AuditRecord { value := valid[6]; value.ContributorMemberIDs = nil; return value }(),
		func() AuditRecord { value := valid[6]; value.ContributorMemberIDs = []string{"member-1"}; return value }(),
		func() AuditRecord {
			value := valid[6]
			value.ContributorMemberIDs = []string{contributors[1], contributors[0]}
			return value
		}(),
		func() AuditRecord { value := valid[9]; value.RecordActor = systemMetadata.RecordActor; return value }(),
		func() AuditRecord { value := valid[6]; value.TargetType = "page"; return value }(),
		func() AuditRecord { value := valid[10]; value.TargetType = "post"; return value }(),
		func() AuditRecord { value := valid[11]; value.TargetType = "post"; return value }(),
		func() AuditRecord { value := valid[15]; value.PreviousEmail = "invalid"; return value }(),
		func() AuditRecord { value := valid[15]; value.NewEmail = value.PreviousEmail; return value }(),
		func() AuditRecord { value := valid[15]; value.TargetType = "member"; return value }(),
		func() AuditRecord { value := valid[15]; value.ChangedFields = []string{"bad-value"}; return value }(),
		func() AuditRecord { value := valid[16]; value.Email = "invalid"; return value }(),
		func() AuditRecord { value := valid[16]; value.CollectionOperation = "replaced"; return value }(),
		func() AuditRecord { value := valid[16]; value.ChangedFields = []string{"bad-value"}; return value }(),
		func() AuditRecord { value := valid[18]; value.Provider = "Bad Provider"; return value }(),
		func() AuditRecord { value := valid[18]; value.ProviderSubject = ""; return value }(),
		func() AuditRecord { value := valid[18]; value.CollectionOperation = "replaced"; return value }(),
		func() AuditRecord { value := valid[18]; value.ChangedFields = []string{"bad-value"}; return value }(),
		func() AuditRecord { value := valid[20]; value.PasskeyIDs = nil; return value }(),
		func() AuditRecord { value := valid[20]; value.PasskeyIDs = []string{" bad-identifier"}; return value }(),
		func() AuditRecord {
			value := valid[20]
			value.PasskeyIDs = []string{"passkey-1", "passkey-1"}
			return value
		}(),
		func() AuditRecord { value := valid[20]; value.CollectionOperation = "replaced"; return value }(),
		func() AuditRecord { value := valid[20]; value.ChangedFields = []string{"bad-value"}; return value }(),
		func() AuditRecord {
			value := valid[22]
			value.CollectionOperation = AuditCollectionOperationAdded
			return value
		}(),
		func() AuditRecord { value := valid[22]; value.SessionScope = "all"; return value }(),
		func() AuditRecord { value := valid[22]; value.SessionIDs = nil; return value }(),
		func() AuditRecord { value := valid[22]; value.SessionIDs = []string{"invalid"}; return value }(),
		func() AuditRecord { value := valid[22]; value.ChangedFields = []string{"bad-value"}; return value }(),
		func() AuditRecord {
			value := valid[22]
			value.SessionScope = AccountSessionScopeOthers
			value.SessionIDs = nil
			return value
		}(),
		func() AuditRecord { value := valid[23]; value.NewState = AuditStateBanned; return value }(),
		func() AuditRecord { value := valid[23]; value.PreviousState = ""; return value }(),
		func() AuditRecord { value := valid[23]; value.ChangedFields = []string{"unknown"}; return value }(),
		func() AuditRecord { value := valid[23]; value.ChangedFields = []string{"bad-value"}; return value }(),
		func() AuditRecord { value := valid[23]; value.ChangedFields = nil; return value }(),
		func() AuditRecord { value := valid[32]; value.TargetType = "member"; return value }(),
		func() AuditRecord { value := valid[32]; value.ChangedFields = []string{"deletion_state"}; return value }(),
	}
	for _, record := range invalid {
		if err := record.Validate(); err == nil {
			t.Fatalf("Validate() accepted %#v", record)
		}
	}
}
func TestSecurityAccessRecordValidate(t *testing.T) {
	t.Parallel()
	base := SecurityAccessRecord{
		AccessID: "7a7a8fd4-1f69-4e9a-9dc2-2378926ff351", OccurredAt: testOccurredAt, Action: SecurityAuthenticationSucceeded,
		Correlation: Correlation{RequestID: testRequestID}, RecordActor: RecordActor{Kind: ActorKindMember, MemberID: "member-1"},
		SourceIP: "192.0.2.4", FlowKind: AuthenticationFlowLogin, AuthenticationMethod: AuthenticationMethodPasskey,
		PrincipalState: AuthenticationPrincipalActive,
	}
	valid := []SecurityAccessRecord{
		base,
		func() SecurityAccessRecord {
			value := base
			value.Action = SecurityAuthenticationFailed
			value.RecordActor = RecordActor{Kind: ActorKindAnonymous}
			value.PrincipalState = ""
			value.Reason = string(AuthenticationFailureProofRejected)
			return value
		}(),
		func() SecurityAccessRecord {
			value := base
			value.Action = SecurityAuthenticationBlocked
			value.RecordActor = RecordActor{Kind: ActorKindAnonymous}
			value.FlowKind = ""
			value.AuthenticationMethod = ""
			value.PrincipalState = ""
			value.Reason = string(AuthenticationBlockRateLimited)
			return value
		}(),
		func() SecurityAccessRecord {
			value := base
			value.Action = SecurityAuthorizationDenied
			value.FlowKind = ""
			value.AuthenticationMethod = ""
			value.PrincipalState = ""
			value.AttemptedAction = "/geul.api.v1.PostService/UpdatePost"
			value.Permission = AuthorizationProcedureInvokePermission
			value.Reason = string(AuthorizationDeniedPermissionDenied)
			return value
		}(),
		func() SecurityAccessRecord {
			value := base
			value.Action = SecurityPersonalDataAccessed
			value.RecordActor = RecordActor{Kind: ActorKindMember, MemberID: "member-1"}
			value.FlowKind = ""
			value.AuthenticationMethod = ""
			value.PrincipalState = ""
			value.SubjectType = "member"
			value.SubjectID = "2a7a8fd4-1f69-4e9a-9dc2-2378926ff351"
			value.AccessKind = "read"
			value.DataCategory = "member_administration"
			return value
		}(),
	}
	for _, record := range valid {
		if err := record.Validate(); err != nil {
			t.Fatalf("Validate() unexpected error = %v for %#v", err, record)
		}
	}
	invalid := []SecurityAccessRecord{
		{},
		func() SecurityAccessRecord { value := base; value.OccurredAt = time.Time{}; return value }(),
		func() SecurityAccessRecord { value := base; value.RequestID = "bad"; return value }(),
		func() SecurityAccessRecord { value := base; value.TraceID = "bad"; return value }(),
		func() SecurityAccessRecord { value := base; value.RecordActor = RecordActor{Kind: "bad"}; return value }(),
		func() SecurityAccessRecord { value := base; value.SourceIP = "bad"; return value }(),
		func() SecurityAccessRecord { value := base; value.AccessID = "access-1"; return value }(),
		func() SecurityAccessRecord {
			value := base
			value.AccessID = "018f47a2-8a3d-4e17-7d42-6f12c89b1234"
			return value
		}(),
		func() SecurityAccessRecord { value := base; value.FlowKind = ""; return value }(),
		func() SecurityAccessRecord {
			value := base
			value.RecordActor = RecordActor{Kind: ActorKindAnonymous}
			return value
		}(),
		func() SecurityAccessRecord { value := base; value.Provider = "google"; return value }(),
		func() SecurityAccessRecord { value := base; value.Reason = "unexpected"; return value }(),
		func() SecurityAccessRecord {
			value := base
			value.Action = SecurityAuthenticationFailed
			value.Reason = string(AuthenticationFailureProofRejected)
			return value
		}(),
		func() SecurityAccessRecord {
			value := valid[1]
			value.RecordActor = RecordActor{Kind: ActorKindSystem, Service: ServiceBackend.String()}
			return value
		}(),
		func() SecurityAccessRecord { value := valid[1]; value.FlowKind = "invalid"; return value }(),
		func() SecurityAccessRecord {
			value := valid[1]
			value.RecordActor = RecordActor{Kind: ActorKindMember, MemberID: "member-1"}
			return value
		}(),
		func() SecurityAccessRecord { value := valid[1]; value.Reason = "unknown"; return value }(),
		func() SecurityAccessRecord {
			value := base
			value.Action = SecurityAuthenticationFailed
			value.RecordActor = RecordActor{Kind: ActorKindAnonymous}
			value.Reason = "unknown"
			return value
		}(),
		func() SecurityAccessRecord {
			value := base
			value.AttemptedAction = "post.update"
			return value
		}(),
		func() SecurityAccessRecord {
			value := base
			value.Action = SecurityAuthenticationFailed
			value.RecordActor = RecordActor{Kind: ActorKindAnonymous}
			value.AttemptedAction = "post.update"
			value.Reason = string(AuthenticationFailureProofRejected)
			return value
		}(),
		func() SecurityAccessRecord {
			value := base
			value.Action = SecurityAuthenticationFailed
			value.RecordActor = RecordActor{Kind: ActorKindSystem, Service: ServiceBackend.String()}
			value.Reason = string(AuthenticationFailureProofRejected)
			return value
		}(),
		func() SecurityAccessRecord {
			value := base
			value.Action = SecurityAuthenticationFailed
			value.RecordActor = RecordActor{Kind: ActorKindAnonymous}
			value.AuthenticationMethod = ""
			value.Reason = string(AuthenticationFailureProofRejected)
			return value
		}(),
		func() SecurityAccessRecord {
			value := base
			value.Action = SecurityAuthenticationBlocked
			value.RecordActor = RecordActor{Kind: ActorKindAnonymous}
			value.Reason = "unknown"
			return value
		}(),
		func() SecurityAccessRecord {
			value := valid[2]
			value.RecordActor = RecordActor{Kind: ActorKindSystem, Service: ServiceBackend.String()}
			return value
		}(),
		func() SecurityAccessRecord { value := valid[2]; value.FlowKind = "invalid"; return value }(),
		func() SecurityAccessRecord { value := valid[2]; value.AuthenticationMethod = "invalid"; return value }(),
		func() SecurityAccessRecord {
			value := valid[2]
			value.RecordActor = RecordActor{Kind: ActorKindMember, MemberID: "member-1"}
			return value
		}(),
		func() SecurityAccessRecord { value := valid[2]; value.Reason = "unknown"; return value }(),
		func() SecurityAccessRecord {
			value := base
			value.Action = SecurityAuthenticationBlocked
			value.RecordActor = RecordActor{Kind: ActorKindAnonymous}
			value.Provider = "google"
			value.Reason = string(AuthenticationBlockRateLimited)
			return value
		}(),
		func() SecurityAccessRecord {
			value := base
			value.Action = SecurityAuthenticationBlocked
			value.RecordActor = RecordActor{Kind: ActorKindAnonymous}
			value.AttemptedAction = "post.update"
			value.Reason = string(AuthenticationBlockRateLimited)
			return value
		}(),
		func() SecurityAccessRecord {
			value := base
			value.Action = SecurityAuthenticationBlocked
			value.RecordActor = RecordActor{Kind: ActorKindSystem, Service: ServiceBackend.String()}
			value.Reason = string(AuthenticationBlockRateLimited)
			return value
		}(),
		func() SecurityAccessRecord {
			value := base
			value.Action = SecurityAuthenticationBlocked
			value.RecordActor = RecordActor{Kind: ActorKindAnonymous}
			value.FlowKind = "invalid"
			value.Reason = string(AuthenticationBlockRateLimited)
			return value
		}(),
		func() SecurityAccessRecord {
			value := base
			value.Action = SecurityAuthenticationBlocked
			value.RecordActor = RecordActor{Kind: ActorKindAnonymous}
			value.AuthenticationMethod = "invalid"
			value.Reason = string(AuthenticationBlockRateLimited)
			return value
		}(),
		func() SecurityAccessRecord {
			value := base
			value.Action = SecurityAuthenticationBlocked
			value.Reason = string(AuthenticationBlockRateLimited)
			return value
		}(),
		func() SecurityAccessRecord { value := base; value.Action = SecurityAuthorizationDenied; return value }(),
		func() SecurityAccessRecord {
			value := base
			value.Action = SecurityAuthorizationDenied
			value.FlowKind = ""
			value.AuthenticationMethod = ""
			value.RecordActor = RecordActor{Kind: ActorKindSystem, Service: ServiceBackend.String()}
			value.AttemptedAction = "/geul.api.v1.PostService/UpdatePost"
			value.Permission = AuthorizationProcedureInvokePermission
			value.Reason = string(AuthorizationDeniedPermissionDenied)
			return value
		}(),
		func() SecurityAccessRecord {
			value := valid[3]
			value.RecordActor = RecordActor{Kind: ActorKindSystem, Service: ServiceBackend.String()}
			return value
		}(),
		func() SecurityAccessRecord { value := valid[3]; value.AttemptedAction = "bad"; return value }(),
		func() SecurityAccessRecord { value := valid[3]; value.Reason = "unknown"; return value }(),
		func() SecurityAccessRecord {
			value := base
			value.Action = SecurityAuthorizationDenied
			value.FlowKind = ""
			value.AuthenticationMethod = ""
			value.AttemptedAction = "/geul.api.v1.PostService/UpdatePost"
			value.Permission = ""
			value.Reason = string(AuthorizationDeniedPermissionDenied)
			return value
		}(),
		func() SecurityAccessRecord {
			value := base
			value.Action = SecurityAuthorizationDenied
			value.FlowKind = ""
			value.AuthenticationMethod = ""
			value.AttemptedAction = "/geul.api.v1.PostService/UpdatePost"
			value.Permission = "Post:Write"
			value.Reason = string(AuthorizationDeniedPermissionDenied)
			return value
		}(),
		func() SecurityAccessRecord {
			value := base
			value.Action = SecurityAuthorizationDenied
			value.FlowKind = ""
			value.AuthenticationMethod = ""
			value.AttemptedAction = "geul.api.v1.PostService/UpdatePost"
			value.Permission = AuthorizationProcedureInvokePermission
			value.Reason = string(AuthorizationDeniedPermissionDenied)
			return value
		}(),
		func() SecurityAccessRecord {
			value := base
			value.Action = SecurityAuthorizationDenied
			value.FlowKind = ""
			value.AuthenticationMethod = ""
			value.AttemptedAction = "/geul.api.v1.PostService/UpdatePost"
			value.Permission = AuthorizationProcedureInvokePermission
			value.Reason = "policy_denied"
			return value
		}(),
		func() SecurityAccessRecord {
			value := base
			value.Action = SecurityAuthorizationDenied
			value.AttemptedAction = "/geul.api.v1.PostService/UpdatePost"
			value.Permission = AuthorizationProcedureInvokePermission
			value.Reason = "INVALID"
			return value
		}(),
		func() SecurityAccessRecord { value := base; value.Action = SecurityPersonalDataAccessed; return value }(),
		func() SecurityAccessRecord {
			value := base
			value.Action = SecurityPersonalDataAccessed
			value.FlowKind = ""
			value.AuthenticationMethod = ""
			value.SubjectType = "INVALID"
			value.SubjectID = "2a7a8fd4-1f69-4e9a-9dc2-2378926ff351"
			value.AccessKind = PersonalDataAccessRead
			value.DataCategory = "member_administration"
			return value
		}(),
		func() SecurityAccessRecord { value := valid[4]; value.SubjectType = "unknown"; return value }(),
		func() SecurityAccessRecord { value := base; value.Action = "other"; return value }(),
	}
	for _, record := range invalid {
		if err := record.Validate(); err == nil {
			t.Fatalf("Validate() accepted %#v", record)
		}
	}
}

func TestSecurityAccessGoldenFixture(t *testing.T) {
	t.Parallel()
	fixture, err := os.ReadFile("fixtures/security-access-records.json")
	if err != nil {
		t.Fatal(err)
	}
	var records []SecurityAccessRecord
	if err := json.Unmarshal(fixture, &records); err != nil {
		t.Fatal(err)
	}
	if len(records) != 5 {
		t.Fatalf("fixture record count = %d", len(records))
	}
	for _, record := range records {
		if err := record.Validate(); err != nil {
			t.Fatalf("fixture record %#v: %v", record, err)
		}
	}
}

func TestAuditGoldenFixture(t *testing.T) {
	t.Parallel()
	fixture, err := os.ReadFile("fixtures/audit-records.json")
	if err != nil {
		t.Fatal(err)
	}
	var records []AuditRecord
	if err := json.Unmarshal(fixture, &records); err != nil {
		t.Fatal(err)
	}
	if len(records) != 27 {
		t.Fatalf("fixture record count = %d", len(records))
	}
	for _, record := range records {
		if err := record.Validate(); err != nil {
			t.Fatalf("fixture record %#v: %v", record, err)
		}
	}
}

func TestRecordHelpersAndGoldenFixture(t *testing.T) {
	t.Parallel()
	traceID, _ := trace.TraceIDFromHex("4bf92f3577b34da6a3ce929d0e0e4736")
	spanID, _ := trace.SpanIDFromHex("00f067aa0ba902b7")
	spanContext := trace.NewSpanContext(trace.SpanContextConfig{TraceID: traceID, SpanID: spanID})
	requestContext, _ := NewPropagatedRequestContext(testRequestID, AnonymousActor{})
	ctx := trace.ContextWithSpanContext(WithRequestContext(context.Background(), requestContext), spanContext)
	if got := CorrelationFromContext(ctx); got.RequestID != testRequestID || got.TraceID != traceID.String() || got.SpanID != spanID.String() {
		t.Fatalf("CorrelationFromContext() = %#v", got)
	}
	if got := CorrelationFromContext(context.Background()); got != (Correlation{}) {
		t.Fatalf("empty correlation = %#v", got)
	}

	record := RequestRecord{
		Event: "request.completed", OccurredAt: testOccurredAt,
		Correlation: Correlation{RequestID: testRequestID, TraceID: traceID.String(), SpanID: spanID.String()},
		RecordActor: RecordActor{Kind: ActorKindMember, MemberID: "member_123"},
		HTTPMethod:  "PATCH", HTTPRoute: "/posts/{post_id}", StatusCode: 200, DurationMS: 42, Outcome: RequestOutcomeSucceeded,
	}
	encoded, err := json.Marshal(record)
	if err != nil {
		t.Fatal(err)
	}
	fixture, err := os.ReadFile("fixtures/request-record.json")
	if err != nil {
		t.Fatal(err)
	}
	var got, want any
	if err := json.Unmarshal(encoded, &got); err != nil {
		t.Fatal(err)
	}
	if err := json.Unmarshal(fixture, &want); err != nil {
		t.Fatal(err)
	}
	if !reflect.DeepEqual(got, want) {
		t.Fatalf("wire record = %s, want %s", encoded, fixture)
	}
}

func TestRecordValidationUtilities(t *testing.T) {
	t.Parallel()
	if err := validateRecordTime(time.Date(2026, 1, 1, 0, 0, 0, 0, time.FixedZone("west", -3600))); err == nil {
		t.Fatal("non-UTC time accepted")
	}
	if err := validateTraceCorrelation("", "abc"); err == nil {
		t.Fatal("partial trace correlation accepted")
	}
	if err := validateTraceCorrelation("00000000000000000000000000000000", "0000000000000000"); err == nil {
		t.Fatal("zero trace correlation accepted")
	}
	for _, value := range []string{"post", "Post.created", "post.", "post.cre-ated"} {
		if isDottedName(value) {
			t.Fatalf("isDottedName(%q) = true", value)
		}
	}
	if !isDottedName("post.version_created") {
		t.Fatal("valid dotted name rejected")
	}
	if isBoundedCode("") || isBoundedCode("aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa") || isBoundedCode("INVALID") || !isBoundedCode("policy_denied") {
		t.Fatal("bounded code validation returned an unexpected result")
	}
}
