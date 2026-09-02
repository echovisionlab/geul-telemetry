package telemetry

import (
	"slices"
	"testing"
)

func TestPostParticipantAudit(t *testing.T) {
	m := AuditMetadata{AuditID: "00000000-0000-4000-8000-000000000001", OccurredAt: testOccurredAt, RecordActor: RecordActor{Kind: ActorKindMember, MemberID: "member-1"}}
	record, e := NewPostParticipantAuditRecord(m, "post-1", "member-2", AuditRelationshipNone, AuditRelationshipAuthor)
	if e != nil {
		t.Fatal(e)
	}
	if got, want := record.ChangedFields, []string{"authors"}; !slices.Equal(got, want) {
		t.Fatalf("changed fields = %v, want %v", got, want)
	}
	record, e = NewPostParticipantAuditRecord(m, "post-1", "member-2", AuditRelationshipAuthor, AuditRelationshipCollaborator)
	if e != nil {
		t.Fatal(e)
	}
	if got, want := record.ChangedFields, []string{"authors", "collaborators"}; !slices.Equal(got, want) {
		t.Fatalf("changed fields = %v, want %v", got, want)
	}
	if _, e := NewPostParticipantAuditRecord(m, "post-1", "member-2", AuditRelationshipAuthor, AuditRelationshipAuthor); e == nil {
		t.Fatal("no-op accepted")
	}
	record.ChangedFields = []string{"authors"}
	if err := record.Validate(); err == nil {
		t.Fatal("relationship transition with mismatched changed_fields accepted")
	}
	record.ChangedFields = []string{"authors", "collaborators"}
	record.NewRelationship = AuditRelationshipAuthor
	if err := record.Validate(); err == nil {
		t.Fatal("invalid relationship transition accepted")
	}
}

func TestPostLifecycleRejectsOutOfCatalogState(t *testing.T) {
	m := AuditMetadata{AuditID: "00000000-0000-4000-8000-000000000001", OccurredAt: testOccurredAt, RecordActor: RecordActor{Kind: ActorKindMember, MemberID: "member-1"}}
	if _, err := NewPostStatusLifecycleAuditRecord(m, "post-1", AuditStateActive, AuditStatePublished); err == nil {
		t.Fatal("out-of-catalog post lifecycle state accepted")
	}
}
