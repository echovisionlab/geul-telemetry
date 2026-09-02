package telemetry

import "testing"

func TestReferenceEntityAuditBuilders(t *testing.T) {
	m := AuditMetadata{AuditID: "00000000-0000-4000-8000-000000000001", OccurredAt: testOccurredAt, RecordActor: RecordActor{Kind: ActorKindMember, MemberID: "member-1"}}
	builders := []func() (AuditRecord, error){
		func() (AuditRecord, error) { return NewClientCreatedAuditRecord(m, "client-1") }, func() (AuditRecord, error) {
			return NewClientLogoUpdatedAuditRecord(m, "client-1", AuditAssetSlotLight, AuditCollectionOperationAdded, "file-1")
		}, func() (AuditRecord, error) { return NewClientDeletedAuditRecord(m, "client-1") },
		func() (AuditRecord, error) { return NewMapPlaceCreatedAuditRecord(m, "place-1") }, func() (AuditRecord, error) {
			return NewMapPlaceImageUpdatedAuditRecord(m, "place-1", AuditCollectionOperationRemoved, "file-1")
		}, func() (AuditRecord, error) { return NewMapPlaceDeletedAuditRecord(m, "place-1") },
		func() (AuditRecord, error) { return NewAudienceSegmentCreatedAuditRecord(m, "segment-1") }, func() (AuditRecord, error) {
			return NewAudienceSegmentLifecycleUpdatedAuditRecord(m, "segment-1", AuditStateActive, AuditStateArchived)
		},
		func() (AuditRecord, error) { return NewMenuCreatedAuditRecord(m, "menu-1") }, func() (AuditRecord, error) { return NewMenuSourceUpdatedAuditRecord(m, "menu-1", []string{"items"}) }, func() (AuditRecord, error) { return NewMenuDeletedAuditRecord(m, "menu-1") },
	}
	for _, build := range builders {
		if _, err := build(); err != nil {
			t.Fatal(err)
		}
	}
	if _, err := NewClientLogoUpdatedAuditRecord(m, "client-1", AuditAssetSlot("bad"), AuditCollectionOperationAdded, "file-1"); err == nil {
		t.Fatal("bad client logo slot accepted")
	}
	if _, err := NewAudienceSegmentLifecycleUpdatedAuditRecord(m, "segment-1", AuditStateActive, AuditStateActive); err == nil {
		t.Fatal("audience no-op accepted")
	}
}
