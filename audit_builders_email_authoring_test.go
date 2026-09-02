package telemetry

import "testing"

func TestEmailAuthoringBuilders(t *testing.T) {
	m := AuditMetadata{AuditID: "00000000-0000-4000-8000-000000000001", OccurredAt: testOccurredAt, RecordActor: RecordActor{Kind: ActorKindMember, MemberID: "member-1"}}
	if _, e := NewEmailTemplateLayoutRelationAuditRecord(m, "template-1", "layout-1", ""); e != nil {
		t.Fatal(e)
	}
	if _, e := NewEmailTemplateLayoutRelationAuditRecord(m, "template-1", "layout-1", "layout-1"); e == nil {
		t.Fatal("no-op relation accepted")
	}
	if _, e := NewEmailEventMappingTemplateAuditRecord(m, "welcome", "", "template-1"); e != nil {
		t.Fatal(e)
	}
	if _, e := NewEmailEventMappingTemplateAuditRecord(m, "bad event", "", "template-1"); e == nil {
		t.Fatal("invalid event accepted")
	}
	if _, e := NewEmailEventMappingTemplateAuditRecord(m, "welcome", "", " invalid-template"); e == nil {
		t.Fatal("invalid template accepted")
	}
	if _, e := NewEmailEventMappingTemplateAuditRecord(m, "welcome", " invalid-template", "template-1"); e == nil {
		t.Fatal("invalid previous template accepted")
	}
}
