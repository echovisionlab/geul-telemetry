package telemetry

import (
	"testing"
)

func TestFileAuditBuilders(t *testing.T) {
	metadata := AuditMetadata{AuditID: "00000000-0000-4000-8000-000000000001", OccurredAt: testOccurredAt, RecordActor: RecordActor{Kind: ActorKindMember, MemberID: "member-1"}}
	builders := []func() (AuditRecord, error){
		func() (AuditRecord, error) { return NewFileCreatedAuditRecord(metadata, "file-1") }, func() (AuditRecord, error) { return NewFileDeletedAuditRecord(metadata, "file-1") },
		func() (AuditRecord, error) { return NewFileRenamedAuditRecord(metadata, "file-1") }, func() (AuditRecord, error) { return NewFileMovedAuditRecord(metadata, "file-1", "folder-1", "") },
		func() (AuditRecord, error) { return NewFileFolderCreatedAuditRecord(metadata, "folder-1") }, func() (AuditRecord, error) { return NewFileFolderDeletedAuditRecord(metadata, "folder-1") },
		func() (AuditRecord, error) { return NewFileFolderRenamedAuditRecord(metadata, "folder-1") }, func() (AuditRecord, error) {
			return NewFileFolderMovedAuditRecord(metadata, "folder-1", "", "folder-2")
		},
	}
	for _, build := range builders {
		if _, err := build(); err != nil {
			t.Fatal(err)
		}
	}
}

func TestFileAuditValidatorsRejectWrongVariantShape(t *testing.T) {
	metadata := AuditMetadata{AuditID: "00000000-0000-4000-8000-000000000001", OccurredAt: testOccurredAt, RecordActor: RecordActor{Kind: ActorKindMember, MemberID: "member-1"}}
	if _, err := NewFileMovedAuditRecord(metadata, "file-1", "", ""); err == nil {
		t.Fatal("no-op file move accepted")
	}
	if _, err := NewFileFolderMovedAuditRecord(metadata, "folder-1", "", ""); err == nil {
		t.Fatal("no-op folder move accepted")
	}
}
