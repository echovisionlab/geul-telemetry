package telemetry

import (
	"encoding/json"
	"os"
	"testing"
)

type auditWireFixture struct {
	Case           string         `json:"case"`
	Action         AuditAction    `json:"action"`
	TargetType     string         `json:"target_type"`
	TargetID       string         `json:"target_id"`
	Attributes     map[string]any `json:"attributes"`
	CorrectedShape bool           `json:"corrected_shape"`
}

func TestDomainAuditWireParityFixture(t *testing.T) {
	contents, err := os.ReadFile("fixtures/domain-audit-wire-parity.json")
	if err != nil {
		t.Fatal(err)
	}
	var fixtures []auditWireFixture
	if err := json.Unmarshal(contents, &fixtures); err != nil {
		t.Fatal(err)
	}
	metadata := AuditMetadata{AuditID: "00000000-0000-4000-8000-000000000001", OccurredAt: testOccurredAt, RecordActor: RecordActor{Kind: ActorKindMember, MemberID: "member-1"}}
	records := []AuditRecord{}
	if record, err := NewPostParticipantAuditRecord(metadata, "post-1", "member-2", AuditRelationshipAuthor, AuditRelationshipCollaborator); err != nil {
		t.Fatal(err)
	} else {
		records = append(records, record)
	}
	if record, err := NewPostFileBlockDownloadPolicyAuditRecord(metadata, "post-1", "block-1", "file-1", AuditStateDisabled, AuditStateRestricted, []string{"segment-1"}, []string{"segment-2"}); err != nil {
		t.Fatal(err)
	} else {
		records = append(records, record)
	}
	if record, err := NewMapThemeContentUpdatedAuditRecord(metadata, "theme-1"); err != nil {
		t.Fatal(err)
	} else {
		records = append(records, record)
	}
	corrected := make([]auditWireFixture, 0, len(records))
	for _, fixture := range fixtures {
		if fixture.CorrectedShape {
			corrected = append(corrected, fixture)
		}
	}
	if len(corrected) != len(records) {
		t.Fatalf("corrected fixture count = %d, records = %d", len(corrected), len(records))
	}
	for index, fixture := range corrected {
		record := records[index]
		if record.Action != fixture.Action || record.TargetType != fixture.TargetType || record.TargetID != fixture.TargetID {
			t.Fatalf("%s envelope = %s/%s/%s", fixture.Case, record.Action, record.TargetType, record.TargetID)
		}
		wire, err := json.Marshal(record)
		if err != nil {
			t.Fatal(err)
		}
		var decoded map[string]any
		if err := json.Unmarshal(wire, &decoded); err != nil {
			t.Fatal(err)
		}
		for _, key := range []string{
			"audit_id", "occurred_at", "action", "target_type", "target_id",
			"request_id", "trace_id", "span_id",
			"actor_kind", "actor_member_id", "actor_service",
		} {
			delete(decoded, key)
		}
		if !jsonValuesEqual(decoded, fixture.Attributes) {
			t.Fatalf("%s attributes = %#v, want %#v", fixture.Case, decoded, fixture.Attributes)
		}
	}
}

func jsonValuesEqual(left, right any) bool {
	leftWire, _ := json.Marshal(left)
	rightWire, _ := json.Marshal(right)
	return string(leftWire) == string(rightWire)
}
