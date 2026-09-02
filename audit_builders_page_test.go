package telemetry

import "testing"

func TestPageAuditBuildersAndVariants(t *testing.T) {
	m := AuditMetadata{AuditID: "00000000-0000-4000-8000-000000000001", OccurredAt: testOccurredAt, RecordActor: RecordActor{Kind: ActorKindMember, MemberID: "member-1"}}
	collab := m
	collab.RecordActor = RecordActor{Kind: ActorKindSystem, Service: string(ServiceEditorCollab)}
	contributors := []string{"1b6bcad2-c90d-49e9-bec7-f9a4ba6b2894", "7a7a8fd4-1f69-4e9a-9dc2-2378926ff351"}
	builders := []func() (AuditRecord, error){
		func() (AuditRecord, error) { return NewPageCreatedAuditRecord(m, "page-1") },
		func() (AuditRecord, error) {
			return NewPageVersionCreatedAuditRecord(collab, "page-1", "version-1", contributors)
		},
		func() (AuditRecord, error) {
			return NewPageConfigurationAuditRecord(m, "page-1", []string{"slug", "show_title", "slug"})
		},
		func() (AuditRecord, error) {
			return NewPageLifecycleAuditRecord(m, "page-1", AuditStateDraft, AuditStatePublished)
		},
		func() (AuditRecord, error) {
			return NewPageLifecycleAuditRecord(m, "page-1", AuditStatePublished, AuditStateDraft)
		},
		func() (AuditRecord, error) {
			return NewPageFeaturedImageAuditRecord(m, "page-1", "asset-1", AuditCollectionOperationAdded)
		},
		func() (AuditRecord, error) {
			return NewPageFeaturedImageAuditRecord(m, "page-1", "asset-1", AuditCollectionOperationRemoved)
		},
		func() (AuditRecord, error) {
			return NewPageShareLinkAuditRecord(m, "page-1", "link-1", AuditItemOperationCreated)
		},
		func() (AuditRecord, error) {
			return NewPageShareLinkAuditRecord(m, "page-1", "link-1", AuditItemOperationDeleted)
		},
		func() (AuditRecord, error) { return NewPageVersionRestoreAuditRecord(m, "page-1", "version-1") },
		func() (AuditRecord, error) { return NewPageDeletedAuditRecord(m, "page-1") },
	}
	for _, build := range builders {
		if _, err := build(); err != nil {
			t.Fatal(err)
		}
	}
	if _, err := NewPageConfigurationAuditRecord(m, "page-1", []string{"title"}); err == nil {
		t.Fatal("unknown configuration field accepted")
	}
	if _, err := NewPageLifecycleAuditRecord(m, "page-1", AuditStateDraft, AuditStateDraft); err == nil {
		t.Fatal("lifecycle no-op accepted")
	}
	if _, err := NewPageShareLinkAuditRecord(m, "page-1", "link-1", AuditItemOperationUpdated); err == nil {
		t.Fatal("share link update accepted")
	}
}

func TestPageAuditRejectsExtraAttributesAndActors(t *testing.T) {
	m := AuditMetadata{AuditID: "00000000-0000-4000-8000-000000000001", OccurredAt: testOccurredAt, RecordActor: RecordActor{Kind: ActorKindMember, MemberID: "member-1"}}
	collab := m
	collab.RecordActor = RecordActor{Kind: ActorKindSystem, Service: string(ServiceEditorCollab)}
	contributors := []string{"1b6bcad2-c90d-49e9-bec7-f9a4ba6b2894"}
	record, err := NewPageFeaturedImageAuditRecord(m, "page-1", "asset-1", AuditCollectionOperationAdded)
	if err != nil {
		t.Fatal(err)
	}
	record.ItemID = "link-1"
	if err := record.Validate(); err == nil {
		t.Fatal("featured image extra accepted")
	}

	root, err := NewPageCreatedAuditRecord(m, "page-1")
	if err != nil {
		t.Fatal(err)
	}
	root.RecordActor = RecordActor{Kind: ActorKindSystem, Service: string(ServiceBackend)}
	if err := root.Validate(); err == nil {
		t.Fatal("system page root accepted")
	}
	updated, err := NewPageVersionCreatedAuditRecord(collab, "page-1", "version-1", contributors)
	if err != nil {
		t.Fatal(err)
	}
	updated.RecordActor = RecordActor{Kind: ActorKindSystem, Service: string(ServiceEditorCollab)}
	if err := updated.Validate(); err != nil {
		t.Fatalf("collab page version rejected: %v", err)
	}
	updated.RecordActor = RecordActor{Kind: ActorKindSystem, Service: string(ServiceBackend)}
	if err := updated.Validate(); err == nil {
		t.Fatal("backend page version accepted")
	}
	updated, err = NewPageVersionRestoreAuditRecord(m, "page-1", "version-1")
	if err != nil {
		t.Fatal(err)
	}
	updated.RecordActor = RecordActor{Kind: ActorKindSystem, Service: string(ServiceEditorCollab)}
	if err := updated.Validate(); err == nil {
		t.Fatal("system page non-version update accepted")
	}
}
