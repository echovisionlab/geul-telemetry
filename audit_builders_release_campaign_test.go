package telemetry

import "testing"

func TestReleaseTrackAudioAndCampaignLifecycleBuilders(t *testing.T) {
	m := AuditMetadata{AuditID: "00000000-0000-4000-8000-000000000001", OccurredAt: testOccurredAt, RecordActor: RecordActor{Kind: ActorKindMember, MemberID: "member-1"}}
	if _, err := NewReleaseTrackAudioAuditRecord(m, "release-1", "track-1", "file-1", AuditCollectionOperationAdded); err != nil {
		t.Fatal(err)
	}
	if _, err := NewReleaseTrackAudioAuditRecord(m, "release-1", "", "file-1", AuditCollectionOperationAdded); err == nil {
		t.Fatal("track audio without track accepted")
	}
	if _, err := NewCampaignStatusLifecycleAuditRecord(m, "campaign-1", AuditStateDraft, AuditStatePublished); err != nil {
		t.Fatal(err)
	}
	if _, err := NewCampaignScheduleLifecycleAuditRecord(m, "campaign-1", AuditStateDraft, AuditStateScheduled, testOccurredAt); err != nil {
		t.Fatal(err)
	}
	if _, err := NewCampaignDeliveryRunLifecycleAuditRecord(m, "campaign-1", AuditStateScheduled, AuditStateSending, "run-1"); err != nil {
		t.Fatal(err)
	}
}
