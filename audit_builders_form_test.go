package telemetry

import "testing"

func TestFormSubmissionActorsAndBuilders(t *testing.T) {
	base := AuditMetadata{AuditID: "00000000-0000-4000-8000-000000000001", OccurredAt: testOccurredAt}
	anon := base
	anon.RecordActor = RecordActor{Kind: ActorKindAnonymous}
	if _, e := NewFormSubmissionCreatedAuditRecord(anon, "submission-1", "form-1"); e != nil {
		t.Fatal(e)
	}
	member := base
	member.RecordActor = RecordActor{Kind: ActorKindMember, MemberID: "member-1"}
	if _, e := NewFormSubmissionCreatedAuditRecord(member, "submission-1", "form-1"); e != nil {
		t.Fatal(e)
	}
	if _, e := NewFormSubmissionCreatedAuditRecord(member, "submission-1", ""); e == nil {
		t.Fatal("missing parent accepted")
	}
	if _, e := NewFormShareLinkAuditRecord(member, "form-1", "link-1", AuditItemScopeForm, AuditItemOperationCreated); e != nil {
		t.Fatal(e)
	}
	if _, e := NewFormShareLinkAuditRecord(member, "form-1", "link-1", AuditItemScope("bad"), AuditItemOperationCreated); e == nil {
		t.Fatal("bad scope accepted")
	}
}
