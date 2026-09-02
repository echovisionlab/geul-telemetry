package telemetry

import "testing"

func TestWorkAuditBuildersAndVariants(t *testing.T) {
	m := AuditMetadata{AuditID: "00000000-0000-4000-8000-000000000001", OccurredAt: testOccurredAt, RecordActor: RecordActor{Kind: ActorKindMember, MemberID: "member-1"}}
	collab := m
	collab.RecordActor = RecordActor{Kind: ActorKindSystem, Service: string(ServiceEditorCollab)}
	contributors := []string{"1b6bcad2-c90d-49e9-bec7-f9a4ba6b2894", "7a7a8fd4-1f69-4e9a-9dc2-2378926ff351"}
	builders := []func() (AuditRecord, error){
		func() (AuditRecord, error) { return NewWorkCreatedAuditRecord(m, "work-1") },
		func() (AuditRecord, error) {
			return NewWorkVersionCreatedAuditRecord(collab, "work-1", "version-1", contributors)
		},
		func() (AuditRecord, error) {
			return NewWorkMetadataAuditRecord(m, "work-1", []string{"year", "slug", "year"})
		},
		func() (AuditRecord, error) {
			return NewWorkLifecycleAuditRecord(m, "work-1", AuditStateDraft, AuditStatePublished)
		},
		func() (AuditRecord, error) {
			return NewWorkLifecycleAuditRecord(m, "work-1", AuditStatePublished, AuditStateDraft)
		},
		func() (AuditRecord, error) {
			return NewWorkLifecycleAuditRecord(m, "work-1", AuditStatePublished, AuditStateArchived)
		},
		func() (AuditRecord, error) {
			return NewWorkLifecycleAuditRecord(m, "work-1", AuditStateArchived, AuditStatePublished)
		},
		func() (AuditRecord, error) {
			return NewWorkFeaturedImageAuditRecord(m, "work-1", "asset-1", AuditCollectionOperationAdded)
		},
		func() (AuditRecord, error) {
			return NewWorkFeaturedImageAuditRecord(m, "work-1", "asset-1", AuditCollectionOperationRemoved)
		},
		func() (AuditRecord, error) {
			return NewWorkCreditAuditRecord(m, "work-1", "credit-1", AuditItemOperationCreated)
		},
		func() (AuditRecord, error) {
			return NewWorkCreditAuditRecord(m, "work-1", "group-1", AuditItemOperationUpdated)
		},
		func() (AuditRecord, error) {
			return NewWorkCreditAuditRecord(m, "work-1", "credit-1", AuditItemOperationDeleted)
		},
		func() (AuditRecord, error) {
			return NewWorkShareLinkAuditRecord(m, "work-1", "link-1", AuditItemOperationCreated)
		},
		func() (AuditRecord, error) {
			return NewWorkShareLinkAuditRecord(m, "work-1", "link-1", AuditItemOperationDeleted)
		},
		func() (AuditRecord, error) { return NewWorkVersionRestoreAuditRecord(m, "work-1", "version-1") },
		func() (AuditRecord, error) { return NewWorkDeletedAuditRecord(m, "work-1") },
	}
	for _, build := range builders {
		if _, err := build(); err != nil {
			t.Fatal(err)
		}
	}
	if _, err := NewWorkMetadataAuditRecord(m, "work-1", []string{"title"}); err == nil {
		t.Fatal("unknown metadata field accepted")
	}
	if _, err := NewWorkLifecycleAuditRecord(m, "work-1", AuditStateDraft, AuditStateDraft); err == nil {
		t.Fatal("lifecycle no-op accepted")
	}
	if _, err := NewWorkCreditAuditRecord(m, "work-1", "credit-1", "reordered"); err == nil {
		t.Fatal("unsupported credit operation accepted")
	}
	if _, err := NewWorkShareLinkAuditRecord(m, "work-1", "link-1", AuditItemOperationUpdated); err == nil {
		t.Fatal("share link update accepted")
	}
}

func TestWorkAuditRejectsExtraAttributesAndActors(t *testing.T) {
	m := AuditMetadata{AuditID: "00000000-0000-4000-8000-000000000001", OccurredAt: testOccurredAt, RecordActor: RecordActor{Kind: ActorKindMember, MemberID: "member-1"}}
	collab := m
	collab.RecordActor = RecordActor{Kind: ActorKindSystem, Service: string(ServiceEditorCollab)}
	contributors := []string{"1b6bcad2-c90d-49e9-bec7-f9a4ba6b2894"}
	record, err := NewWorkCreditAuditRecord(m, "work-1", "credit-1", AuditItemOperationCreated)
	if err != nil {
		t.Fatal(err)
	}
	record.NewRole = "director"
	if err := record.Validate(); err == nil {
		t.Fatal("credit free-text accepted")
	}

	featured, err := NewWorkFeaturedImageAuditRecord(m, "work-1", "asset-1", AuditCollectionOperationAdded)
	if err != nil {
		t.Fatal(err)
	}
	featured.Email = "person@example.test"
	if err := featured.Validate(); err == nil {
		t.Fatal("featured image PII accepted")
	}

	root, err := NewWorkCreatedAuditRecord(m, "work-1")
	if err != nil {
		t.Fatal(err)
	}
	root.RecordActor = RecordActor{Kind: ActorKindSystem, Service: string(ServiceBackend)}
	if err := root.Validate(); err == nil {
		t.Fatal("system work root accepted")
	}

	updated, err := NewWorkVersionCreatedAuditRecord(collab, "work-1", "version-1", contributors)
	if err != nil {
		t.Fatal(err)
	}
	updated.RecordActor = RecordActor{Kind: ActorKindSystem, Service: string(ServiceEditorCollab)}
	if err := updated.Validate(); err != nil {
		t.Fatalf("collab work version rejected: %v", err)
	}
	updated.RecordActor = RecordActor{Kind: ActorKindSystem, Service: string(ServiceBackend)}
	if err := updated.Validate(); err == nil {
		t.Fatal("backend work version accepted")
	}
	updated, err = NewWorkVersionRestoreAuditRecord(m, "work-1", "version-1")
	if err != nil {
		t.Fatal(err)
	}
	updated.RecordActor = RecordActor{Kind: ActorKindSystem, Service: string(ServiceEditorCollab)}
	if err := updated.Validate(); err == nil {
		t.Fatal("system work non-version update accepted")
	}
	updated.RecordActor = RecordActor{Kind: ActorKindAnonymous}
	if err := updated.Validate(); err == nil {
		t.Fatal("anonymous work update accepted")
	}
}

func TestWorkAuditRejectsInvalidVariantShapes(t *testing.T) {
	m := AuditMetadata{AuditID: "00000000-0000-4000-8000-000000000001", OccurredAt: testOccurredAt, RecordActor: RecordActor{Kind: ActorKindMember, MemberID: "member-1"}}
	collab := m
	collab.RecordActor = RecordActor{Kind: ActorKindSystem, Service: string(ServiceEditorCollab)}
	contributors := []string{"1b6bcad2-c90d-49e9-bec7-f9a4ba6b2894"}
	tests := []struct {
		name   string
		build  func() (AuditRecord, error)
		mutate func(*AuditRecord)
	}{
		{
			name: "version missing contributor",
			build: func() (AuditRecord, error) {
				return NewWorkVersionCreatedAuditRecord(collab, "work-1", "version-1", contributors)
			},
			mutate: func(r *AuditRecord) { r.ContributorMemberIDs = nil },
		},
		{
			name: "lifecycle unknown state",
			build: func() (AuditRecord, error) {
				return NewWorkLifecycleAuditRecord(m, "work-1", AuditStateDraft, AuditStatePublished)
			},
			mutate: func(r *AuditRecord) { r.NewState = AuditStateActive },
		},
		{
			name: "featured image invalid operation",
			build: func() (AuditRecord, error) {
				return NewWorkFeaturedImageAuditRecord(m, "work-1", "asset-1", AuditCollectionOperationAdded)
			},
			mutate: func(r *AuditRecord) { r.CollectionOperation = "replaced" },
		},
		{
			name: "credit missing child ID",
			build: func() (AuditRecord, error) {
				return NewWorkCreditAuditRecord(m, "work-1", "credit-1", AuditItemOperationCreated)
			},
			mutate: func(r *AuditRecord) { r.ItemID = "" },
		},
		{
			name: "share link extra policy",
			build: func() (AuditRecord, error) {
				return NewWorkShareLinkAuditRecord(m, "work-1", "link-1", AuditItemOperationCreated)
			},
			mutate: func(r *AuditRecord) { r.PolicyType = AuditPolicyTypeTerms },
		},
		{
			name:   "restore missing version",
			build:  func() (AuditRecord, error) { return NewWorkVersionRestoreAuditRecord(m, "work-1", "version-1") },
			mutate: func(r *AuditRecord) { r.VersionID = "" },
		},
		{
			name:   "metadata lifecycle mix",
			build:  func() (AuditRecord, error) { return NewWorkMetadataAuditRecord(m, "work-1", []string{"slug", "year"}) },
			mutate: func(r *AuditRecord) { r.PreviousState, r.NewState = AuditStateDraft, AuditStatePublished },
		},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			record, err := test.build()
			if err != nil {
				t.Fatal(err)
			}
			test.mutate(&record)
			if err := record.Validate(); err == nil {
				t.Fatal("invalid Work audit record accepted")
			}
		})
	}
}
