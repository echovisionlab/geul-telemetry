package telemetry

import "testing"

func TestReferenceDataAuditBuilders(t *testing.T) {
	m := AuditMetadata{AuditID: "00000000-0000-4000-8000-000000000001", OccurredAt: testOccurredAt, RecordActor: RecordActor{Kind: ActorKindMember, MemberID: "member-1"}}
	builders := []func() (AuditRecord, error){
		func() (AuditRecord, error) { return NewCategoryCreatedAuditRecord(m, "category-1") }, func() (AuditRecord, error) {
			return NewCategoryMetadataUpdatedAuditRecord(m, "category-1", []string{"slug", "name"})
		}, func() (AuditRecord, error) { return NewCategoryDeletedAuditRecord(m, "category-1") },
		func() (AuditRecord, error) { return NewTagCreatedAuditRecord(m, "tag-1") }, func() (AuditRecord, error) { return NewTagMetadataUpdatedAuditRecord(m, "tag-1", []string{"slug"}) }, func() (AuditRecord, error) { return NewTagDeletedAuditRecord(m, "tag-1") },
		func() (AuditRecord, error) { return NewGenreCreatedAuditRecord(m, "genre-1") }, func() (AuditRecord, error) {
			return NewGenreMetadataUpdatedAuditRecord(m, "genre-1", []string{"description"})
		}, func() (AuditRecord, error) { return NewGenreDeletedAuditRecord(m, "genre-1") },
		func() (AuditRecord, error) { return NewStyleCreatedAuditRecord(m, "style-1") }, func() (AuditRecord, error) { return NewStyleMetadataUpdatedAuditRecord(m, "style-1", []string{"name"}) }, func() (AuditRecord, error) { return NewStyleDeletedAuditRecord(m, "style-1") },
		func() (AuditRecord, error) { return NewFormatCreatedAuditRecord(m, "format-1") }, func() (AuditRecord, error) {
			return NewFormatMetadataUpdatedAuditRecord(m, "format-1", []string{"slug", "name"})
		}, func() (AuditRecord, error) { return NewFormatDeletedAuditRecord(m, "format-1") },
	}
	for _, build := range builders {
		if _, err := build(); err != nil {
			t.Fatal(err)
		}
	}
	if _, err := NewTagMetadataUpdatedAuditRecord(m, "tag-1", []string{"description"}); err == nil {
		t.Fatal("tag accepted undocumented field")
	}
	if _, err := NewFormatMetadataUpdatedAuditRecord(m, "", []string{"name"}); err == nil {
		t.Fatal("empty target accepted")
	}
}
