package telemetry

import (
	"encoding/json"
	"os"
	"slices"
	"strings"
	"testing"
)

func TestAuditBuildersCoverExactCatalog(t *testing.T) {
	t.Parallel()
	memberMetadata := AuditMetadata{
		AuditID:     "00000000-0000-4000-8000-000000000001",
		OccurredAt:  testOccurredAt,
		Correlation: Correlation{RequestID: testRequestID},
		RecordActor: RecordActor{Kind: ActorKindMember, MemberID: "member-1"},
	}
	systemMetadata := memberMetadata
	systemMetadata.RecordActor = RecordActor{Kind: ActorKindSystem, Service: "geul-backend"}
	collabMetadata := memberMetadata
	collabMetadata.RecordActor = RecordActor{Kind: ActorKindSystem, Service: string(ServiceEditorCollab)}
	contributors := []string{
		"1b6bcad2-c90d-49e9-bec7-f9a4ba6b2894",
		"7a7a8fd4-1f69-4e9a-9dc2-2378926ff351",
	}
	builders := []func() (AuditRecord, error){
		func() (AuditRecord, error) {
			return NewSiteSettingsUpdatedAuditRecord(memberMetadata, []string{"site_title"})
		},
		func() (AuditRecord, error) {
			return NewMemberOnboardingCompletedAuditRecord(memberMetadata, "member-1", "Onboarded Member")
		},
		func() (AuditRecord, error) {
			return NewMemberRoleUpdatedAuditRecord(memberMetadata, "member-2", "user", "author")
		},
		func() (AuditRecord, error) { return NewMemberBannedAuditRecord(memberMetadata, "member-2") },
		func() (AuditRecord, error) { return NewMemberUnbannedAuditRecord(memberMetadata, "member-2") },
		func() (AuditRecord, error) {
			return NewPostVersionCreatedAuditRecord(collabMetadata, "post-1", "version-1", contributors)
		},
		func() (AuditRecord, error) {
			return NewPageVersionCreatedAuditRecord(collabMetadata, "page-1", "version-1", contributors)
		},
		func() (AuditRecord, error) {
			return NewWorkVersionCreatedAuditRecord(collabMetadata, "work-1", "version-1", contributors)
		},
		func() (AuditRecord, error) { return NewPostCreatedAuditRecord(memberMetadata, "post-1") },
		func() (AuditRecord, error) { return NewPageCreatedAuditRecord(memberMetadata, "page-1") },
		func() (AuditRecord, error) { return NewWorkCreatedAuditRecord(memberMetadata, "work-1") },
		func() (AuditRecord, error) { return NewPostDeletedAuditRecord(memberMetadata, "post-1") },
		func() (AuditRecord, error) { return NewPageDeletedAuditRecord(memberMetadata, "page-1") },
		func() (AuditRecord, error) { return NewWorkDeletedAuditRecord(memberMetadata, "work-1") },
		func() (AuditRecord, error) {
			return NewAccountCanonicalEmailUpdatedAuditRecord(memberMetadata, "member-1", "old@example.test", "new@example.test")
		},
		func() (AuditRecord, error) {
			return NewAccountEmailLoginAddedAuditRecord(memberMetadata, "member-1", "added@example.test")
		},
		func() (AuditRecord, error) {
			return NewAccountEmailLoginRemovedAuditRecord(memberMetadata, "member-1", "removed@example.test")
		},
		func() (AuditRecord, error) {
			return NewAccountSocialLoginAddedAuditRecord(memberMetadata, "member-1", "google", "google-subject")
		},
		func() (AuditRecord, error) {
			return NewAccountSocialLoginRemovedAuditRecord(memberMetadata, "member-1", "github", "github-subject")
		},
		func() (AuditRecord, error) {
			return NewAccountPasskeyAddedAuditRecord(memberMetadata, "member-1", []string{"passkey-1"})
		},
		func() (AuditRecord, error) {
			return NewAccountPasskeyRemovedAuditRecord(memberMetadata, "member-1", []string{"passkey-2"})
		},
		func() (AuditRecord, error) {
			return NewAccountSessionRevokedAuditRecord(memberMetadata, "member-1", AccountSessionScopeOne, []string{testRequestID})
		},
		func() (AuditRecord, error) {
			return NewAccountDeletionRequestedAuditRecord(memberMetadata, "member-1", AuditStateNone)
		},
		func() (AuditRecord, error) {
			return NewAccountDeletionScheduledAuditRecord(memberMetadata, "member-1", AuditStateConfirmationPending)
		},
		func() (AuditRecord, error) { return NewAccountDeletionCancelledAuditRecord(memberMetadata, "member-1") },
		func() (AuditRecord, error) { return NewAccountDeletionRecoveredAuditRecord(memberMetadata, "member-1") },
		func() (AuditRecord, error) { return NewAccountDeletedAuditRecord(systemMetadata, "member-1") },
	}
	seen := make(map[AuditAction]struct{}, len(builders))
	for _, build := range builders {
		record, err := build()
		if err != nil {
			t.Fatal(err)
		}
		seen[record.Action] = struct{}{}
	}
	if len(seen) != 13 {
		t.Fatalf("builder action count = %d", len(seen))
	}
	if _, err := NewPostCreatedAuditRecord(memberMetadata, ""); err == nil {
		t.Fatal("empty target accepted")
	}
	if _, err := NewMemberOnboardingCompletedAuditRecord(memberMetadata, "member-1", " "); err == nil {
		t.Fatal("blank onboarding nickname accepted")
	}
}

func TestAuditBuildersCanonicalizeSetAttributes(t *testing.T) {
	t.Parallel()
	metadata := AuditMetadata{
		AuditID:     "00000000-0000-4000-8000-000000000001",
		OccurredAt:  testOccurredAt,
		Correlation: Correlation{RequestID: testRequestID},
		RecordActor: RecordActor{Kind: ActorKindMember, MemberID: "member-1"},
	}
	collabMetadata := metadata
	collabMetadata.RecordActor = RecordActor{Kind: ActorKindSystem, Service: string(ServiceEditorCollab)}
	settings, err := NewSiteSettingsUpdatedAuditRecord(metadata, []string{"site_title", "primary_color", "site_title"})
	if err != nil {
		t.Fatal(err)
	}
	wantFields := []string{"primary_color", "site_title"}
	if !slices.Equal(settings.ChangedFields, wantFields) {
		t.Fatalf("changed fields = %#v, want %#v", settings.ChangedFields, wantFields)
	}

	first := "1b6bcad2-c90d-49e9-bec7-f9a4ba6b2894"
	second := "7a7a8fd4-1f69-4e9a-9dc2-2378926ff351"
	version, err := NewPostVersionCreatedAuditRecord(collabMetadata, "post-1", "version-1", []string{second, first, second})
	if err != nil {
		t.Fatal(err)
	}
	wantContributors := []string{first, second}
	if !slices.Equal(version.ContributorMemberIDs, wantContributors) {
		t.Fatalf("contributors = %#v, want %#v", version.ContributorMemberIDs, wantContributors)
	}
}

func TestAuditCatalogMatchesDomainFixture(t *testing.T) {
	fixture, err := os.ReadFile("fixtures/domain-audit-actions.json")
	if err != nil {
		t.Fatal(err)
	}
	var expected []struct {
		Action     AuditAction `json:"action"`
		TargetType string      `json:"target_type"`
	}
	if err := json.Unmarshal(fixture, &expected); err != nil {
		t.Fatal(err)
	}
	if len(auditCatalog) != len(expected) {
		t.Fatalf("catalog count = %d, fixture count = %d", len(auditCatalog), len(expected))
	}
	for _, want := range expected {
		got, ok := auditCatalog[want.Action]
		if !ok || got.targetType != want.TargetType {
			t.Fatalf("catalog entry for %s = %#v, want target %q", want.Action, got, want.TargetType)
		}
		if got.validate == nil {
			t.Fatalf("catalog entry for %s has no validator", want.Action)
		}
	}
}

func TestAuditAppendCatalogUsesEveryDomainAction(t *testing.T) {
	metadata := SystemMetadata{OccurredAt: testOccurredAt}
	for action := range auditCatalog {
		record, err := NewDomainAuditAppendFailedRecord(metadata, action, AuditAppendFailurePersistenceFailed)
		if err != nil {
			t.Fatalf("append failure rejected catalog action %s: %v", action, err)
		}
		if !isKnownAppendAction(AuditRecordClassDomain, record.Action) {
			t.Fatalf("append lookup omitted catalog action %s", action)
		}
	}
}

func TestAuditRecordPreservesIntentionalEmptyPointerSlice(t *testing.T) {
	t.Parallel()
	empty := []string{}
	previous := []string{"segment-1"}
	record := AuditRecord{
		AuditID:     "00000000-0000-4000-8000-000000000001",
		OccurredAt:  testOccurredAt,
		Action:      "post.updated",
		RecordActor: RecordActor{Kind: ActorKindMember, MemberID: "member-1"},
		TargetType:  "post",
		TargetID:    "post-1",
		ChangedFields: []string{
			"file_download_audience_segment_ids",
		},
		ItemID:          "block-1",
		FileID:          "file-1",
		PreviousItemIDs: &previous,
		ItemIDs:         &empty,
	}
	if err := record.Validate(); err != nil {
		t.Fatal(err)
	}
	wire, err := json.Marshal(record)
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(string(wire), `"item_ids":[]`) {
		t.Fatalf("intentional empty item_ids was omitted: %s", wire)
	}
}

func TestCatalogActionWithoutValidatorIsClosed(t *testing.T) {
	t.Parallel()
	record := AuditRecord{
		AuditID:     "00000000-0000-4000-8000-000000000001",
		OccurredAt:  testOccurredAt,
		Action:      "unknown.created",
		RecordActor: RecordActor{Kind: ActorKindMember, MemberID: "member-1"},
		TargetType:  "unknown",
		TargetID:    "unknown-1",
	}
	if err := record.Validate(); err == nil {
		t.Fatal("unimplemented action without attributes accepted")
	}
	record.ChangedFields = []string{"legacy_field"}
	if err := record.Validate(); err == nil {
		t.Fatal("unimplemented action with attributes accepted")
	}
}
