package telemetry

import "testing"

func TestContentActorPolicy(t *testing.T) {
	contributors := []string{"1b6bcad2-c90d-49e9-bec7-f9a4ba6b2894"}
	collab := AuditMetadata{
		AuditID:     "00000000-0000-4000-8000-000000000001",
		OccurredAt:  testOccurredAt,
		RecordActor: RecordActor{Kind: ActorKindSystem, Service: string(ServiceEditorCollab)},
	}
	backend := collab
	backend.RecordActor.Service = string(ServiceBackend)
	member := AuditMetadata{
		AuditID:     collab.AuditID,
		OccurredAt:  collab.OccurredAt,
		RecordActor: RecordActor{Kind: ActorKindMember, MemberID: "member-1"},
	}

	tests := []struct {
		name       string
		version    func(AuditMetadata) (AuditRecord, error)
		memberOnly func(AuditMetadata) (AuditRecord, error)
	}{
		{
			name: "post",
			version: func(m AuditMetadata) (AuditRecord, error) {
				return NewPostVersionCreatedAuditRecord(m, "post-1", "version-1", contributors)
			},
			memberOnly: func(m AuditMetadata) (AuditRecord, error) {
				return NewPostConfigurationAuditRecord(m, "post-1", []string{"slug"})
			},
		},
		{
			name: "page",
			version: func(m AuditMetadata) (AuditRecord, error) {
				return NewPageVersionCreatedAuditRecord(m, "page-1", "version-1", contributors)
			},
			memberOnly: func(m AuditMetadata) (AuditRecord, error) {
				return NewPageConfigurationAuditRecord(m, "page-1", []string{"slug"})
			},
		},
		{
			name: "work",
			version: func(m AuditMetadata) (AuditRecord, error) {
				return NewWorkVersionCreatedAuditRecord(m, "work-1", "version-1", contributors)
			},
			memberOnly: func(m AuditMetadata) (AuditRecord, error) {
				return NewWorkMetadataAuditRecord(m, "work-1", []string{"slug"})
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if _, err := tt.version(collab); err != nil {
				t.Fatalf("collab version checkpoint rejected: %v", err)
			}
			if _, err := tt.version(backend); err == nil {
				t.Fatal("backend version checkpoint accepted")
			}
			if _, err := tt.version(member); err == nil {
				t.Fatal("member version checkpoint accepted")
			}
			if _, err := tt.memberOnly(collab); err == nil {
				t.Fatal("collab non-version mutation accepted")
			}
		})
	}
}
