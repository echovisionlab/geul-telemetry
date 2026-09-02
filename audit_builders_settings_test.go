package telemetry

import "testing"

func TestMapThemeAuditBuilders(t *testing.T) {
	metadata := AuditMetadata{AuditID: "00000000-0000-4000-8000-000000000001", OccurredAt: testOccurredAt, RecordActor: RecordActor{Kind: ActorKindMember, MemberID: "member-1"}}
	builders := []func() (AuditRecord, error){
		func() (AuditRecord, error) { return NewMapThemeCreatedAuditRecord(metadata, "theme-1") },
		func() (AuditRecord, error) { return NewMapThemeDeletedAuditRecord(metadata, "theme-1") },
		func() (AuditRecord, error) { return NewMapThemeContentUpdatedAuditRecord(metadata, "theme-1") },
	}
	for _, build := range builders {
		if _, err := build(); err != nil {
			t.Fatal(err)
		}
	}

	record, err := NewMapThemeCreatedAuditRecord(metadata, "theme-1")
	if err != nil {
		t.Fatal(err)
	}
	record.PolicyType = AuditPolicyTypeTerms
	if err := record.Validate(); err == nil {
		t.Fatal("map theme root accepted a typed attribute")
	}
	content, err := NewMapThemeContentUpdatedAuditRecord(metadata, "theme-1")
	if err != nil || content.ChangedFields[0] != "content" {
		t.Fatal("map theme content update was not emitted")
	}
}

func TestLegalPolicyAuditBuildersAndExactLifecycle(t *testing.T) {
	metadata := AuditMetadata{AuditID: "00000000-0000-4000-8000-000000000001", OccurredAt: testOccurredAt, RecordActor: RecordActor{Kind: ActorKindMember, MemberID: "member-1"}}
	effectiveAt := testOccurredAt
	builders := []func() (AuditRecord, error){
		func() (AuditRecord, error) {
			return NewLegalPolicyCreatedAuditRecord(metadata, "policy-1", AuditPolicyTypeTerms, 1)
		},
		func() (AuditRecord, error) {
			return NewLegalPolicyDeletedAuditRecord(metadata, "policy-1", AuditPolicyTypePrivacy, 2)
		},
		func() (AuditRecord, error) {
			return NewLegalPolicyLifecycleAuditRecord(metadata, "policy-1", AuditPolicyTypeTerms, 1, []string{"effective_at", "status"}, AuditStateDraft, AuditStateScheduled, &effectiveAt)
		},
		func() (AuditRecord, error) {
			return NewLegalPolicyLifecycleAuditRecord(metadata, "policy-1", AuditPolicyTypeTerms, 1, []string{"status"}, AuditStateScheduled, AuditStateDraft, nil)
		},
		func() (AuditRecord, error) {
			return NewLegalPolicyLifecycleAuditRecord(metadata, "policy-1", AuditPolicyTypeTerms, 1, []string{"status"}, AuditStateDraft, AuditStateActive, nil)
		},
		func() (AuditRecord, error) {
			return NewLegalPolicyLifecycleAuditRecord(metadata, "policy-1", AuditPolicyTypeTerms, 1, []string{"status"}, AuditStateScheduled, AuditStateActive, nil)
		},
		func() (AuditRecord, error) {
			return NewLegalPolicyLifecycleAuditRecord(metadata, "policy-1", AuditPolicyTypeTerms, 1, []string{"status"}, AuditStateActive, AuditStateArchived, nil)
		},
		func() (AuditRecord, error) {
			return NewLegalPolicyShareLinkAuditRecord(metadata, "policy-1", AuditPolicyTypeTerms, 1, AuditItemOperationCreated, "link-1")
		},
		func() (AuditRecord, error) {
			return NewLegalPolicyShareLinkAuditRecord(metadata, "policy-1", AuditPolicyTypeTerms, 1, AuditItemOperationDeleted, "link-1")
		},
	}
	for _, build := range builders {
		if _, err := build(); err != nil {
			t.Fatal(err)
		}
	}

	invalid := []func(AuditRecord) AuditRecord{
		func(record AuditRecord) AuditRecord { record.PolicyType = "unknown"; return record },
		func(record AuditRecord) AuditRecord {
			version := int64(0)
			record.VersionNumber = &version
			return record
		},
		func(record AuditRecord) AuditRecord { record.AssetID = "extra"; return record },
	}
	created, err := NewLegalPolicyCreatedAuditRecord(metadata, "policy-1", AuditPolicyTypeTerms, 1)
	if err != nil {
		t.Fatal(err)
	}
	for _, mutate := range invalid {
		if err := mutate(created).Validate(); err == nil {
			t.Fatal("invalid legal policy identity accepted")
		}
	}

	lifecycle, err := NewLegalPolicyLifecycleAuditRecord(metadata, "policy-1", AuditPolicyTypeTerms, 1, []string{"effective_at", "status"}, AuditStateDraft, AuditStateScheduled, &effectiveAt)
	if err != nil {
		t.Fatal(err)
	}
	missingEffectiveAt := lifecycle
	missingEffectiveAt.EffectiveAt = nil
	if err := missingEffectiveAt.Validate(); err == nil {
		t.Fatal("lifecycle accepted effective_at without its timestamp")
	}
	extraEffectiveAt := lifecycle
	extraEffectiveAt.ChangedFields = []string{"status"}
	if err := extraEffectiveAt.Validate(); err == nil {
		t.Fatal("lifecycle accepted a timestamp without effective_at")
	}
	invalidTransition := lifecycle
	invalidTransition.PreviousState, invalidTransition.NewState = AuditStateArchived, AuditStateActive
	if err := invalidTransition.Validate(); err == nil {
		t.Fatal("lifecycle accepted an undocumented transition")
	}
}

func TestMemberAccountContentVariantsRejectEveryUnrelatedTypedAttribute(t *testing.T) {
	metadata := AuditMetadata{AuditID: "00000000-0000-4000-8000-000000000001", OccurredAt: testOccurredAt, RecordActor: RecordActor{Kind: ActorKindMember, MemberID: "member-1"}}
	collabMetadata := metadata
	collabMetadata.RecordActor = RecordActor{Kind: ActorKindSystem, Service: string(ServiceEditorCollab)}
	contributor := []string{"1b6bcad2-c90d-49e9-bec7-f9a4ba6b2894"}
	effectiveAt := testOccurredAt
	tests := []struct {
		name   string
		build  func() (AuditRecord, error)
		mutate func(*AuditRecord)
	}{
		{
			name:   "root",
			build:  func() (AuditRecord, error) { return NewPostCreatedAuditRecord(metadata, "post-1") },
			mutate: func(record *AuditRecord) { record.PolicyType = AuditPolicyTypeTerms },
		},
		{
			name: "version",
			build: func() (AuditRecord, error) {
				return NewPageVersionCreatedAuditRecord(collabMetadata, "page-1", "version-1", contributor)
			},
			mutate: func(record *AuditRecord) { record.PolicyType = AuditPolicyTypeTerms },
		},
		{
			name: "account",
			build: func() (AuditRecord, error) {
				return NewAccountEmailLoginAddedAuditRecord(metadata, "member-1", "member@example.test")
			},
			mutate: func(record *AuditRecord) { record.EffectiveAt = &effectiveAt },
		},
		{
			name: "member",
			build: func() (AuditRecord, error) {
				return NewMemberRoleUpdatedAuditRecord(metadata, "member-1", "member", "author")
			},
			mutate: func(record *AuditRecord) { record.ItemID = "link-1" },
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
				t.Fatal("unrelated typed attribute accepted")
			}
		})
	}
}
