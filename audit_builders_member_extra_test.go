package telemetry

import (
	"encoding/json"
	"strings"
	"testing"
)

func TestMemberExtraAuditBuilders(t *testing.T) {
	m := AuditMetadata{AuditID: "00000000-0000-4000-8000-000000000001", OccurredAt: testOccurredAt, RecordActor: RecordActor{Kind: ActorKindMember, MemberID: "member-1"}}
	builders := []func() (AuditRecord, error){
		func() (AuditRecord, error) {
			return NewMemberProfileUpdatedAuditRecord(m, "member-1", []string{"nickname"}, "Name")
		},
		func() (AuditRecord, error) {
			return NewMemberAvatarUpdatedAuditRecord(m, "member-1", AuditCollectionOperationAdded, "asset-1")
		},
		func() (AuditRecord, error) {
			return NewMemberPreferencesUpdatedAuditRecord(m, "member-1", []string{"preferred_locale"}, "ko", "")
		},
		func() (AuditRecord, error) { return NewMemberTagsUpdatedAuditRecord(m, "member-1", []string{}) },
		func() (AuditRecord, error) { return NewMemberTagCreatedAuditRecord(m, "tag-1", "Featured") }, func() (AuditRecord, error) { return NewMemberTagDeletedAuditRecord(m, "tag-1", "Featured") },
		func() (AuditRecord, error) {
			return NewAccountNewsletterSubscriptionUpdatedAuditRecord(m, "member-1", AuditStateSubscribed, AuditStateUnsubscribed)
		},
	}
	for _, build := range builders {
		if _, err := build(); err != nil {
			t.Fatal(err)
		}
	}
	record, err := NewMemberTagsUpdatedAuditRecord(m, "member-1", []string{})
	if err != nil {
		t.Fatal(err)
	}
	wire, _ := json.Marshal(record)
	if !strings.Contains(string(wire), `"tag_ids":[]`) {
		t.Fatalf("empty tag set omitted: %s", wire)
	}
	if _, err := NewAccountNewsletterSubscriptionUpdatedAuditRecord(m, "member-1", AuditStateSubscribed, AuditStateSubscribed); err == nil {
		t.Fatal("newsletter no-op accepted")
	}
}
