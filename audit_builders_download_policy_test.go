package telemetry

import (
	"encoding/json"
	"strings"
	"testing"
)

func TestRelationDownloadPolicyBuildersUseOwningDomainTargets(t *testing.T) {
	metadata := AuditMetadata{AuditID: "00000000-0000-4000-8000-000000000001", OccurredAt: testOccurredAt, RecordActor: RecordActor{Kind: ActorKindMember, MemberID: "member-1"}}
	builders := []struct {
		name       string
		action     AuditAction
		targetType string
		targetID   string
		itemID     string
		build      func() (AuditRecord, error)
	}{
		{"post File Block", AuditPostUpdated, "post", "post-1", "block-1", func() (AuditRecord, error) {
			return NewPostFileBlockDownloadPolicyAuditRecord(metadata, "post-1", "block-1", "file-1", AuditStateDisabled, AuditStatePublic, nil, nil)
		}},
		{"page File Block", AuditPageUpdated, "page", "page-1", "block-2", func() (AuditRecord, error) {
			return NewPageFileBlockDownloadPolicyAuditRecord(metadata, "page-1", "block-2", "file-2", AuditStatePublic, AuditStateAuthenticated, nil, nil)
		}},
		{"work File Block", AuditWorkUpdated, "work", "work-1", "block-3", func() (AuditRecord, error) {
			return NewWorkFileBlockDownloadPolicyAuditRecord(metadata, "work-1", "block-3", "file-3", AuditStateAuthenticated, AuditStateRestricted, nil, nil)
		}},
		{"program event File Block", AuditProgramEventUpdated, "program_event", "event-1", "block-4", func() (AuditRecord, error) {
			return NewProgramEventFileBlockDownloadPolicyAuditRecord(metadata, "event-1", "block-4", "file-4", AuditStateRestricted, AuditStateDisabled, nil, nil)
		}},
		{"release Track", AuditReleaseUpdated, "release", "release-1", "track-1", func() (AuditRecord, error) {
			return NewReleaseTrackDownloadPolicyAuditRecord(metadata, "release-1", "track-1", "file-5", AuditStateDisabled, AuditStatePublic, nil, nil)
		}},
	}
	for _, tc := range builders {
		t.Run(tc.name, func(t *testing.T) {
			record, err := tc.build()
			if err != nil {
				t.Fatal(err)
			}
			if record.Action != tc.action || record.TargetType != tc.targetType || record.TargetID != tc.targetID || record.ItemID != tc.itemID || record.FileID == "" {
				t.Fatalf("wrong owning relation target: %#v", record)
			}
		})
	}
}

func TestRelationDownloadPolicyBuilderEmitsExactBeforeAfter(t *testing.T) {
	metadata := AuditMetadata{AuditID: "00000000-0000-4000-8000-000000000001", OccurredAt: testOccurredAt, RecordActor: RecordActor{Kind: ActorKindMember, MemberID: "member-1"}}
	record, err := NewPostFileBlockDownloadPolicyAuditRecord(metadata, "post-1", "block-1", "file-1", AuditStatePublic, AuditStateRestricted, []string{"segment-2", "segment-1"}, []string{})
	if err != nil {
		t.Fatal(err)
	}
	wire, err := json.Marshal(record)
	if err != nil {
		t.Fatal(err)
	}
	want := []string{
		`"changed_fields":["file_download_audience","file_download_audience_segment_ids"]`,
		`"item_id":"block-1"`, `"file_id":"file-1"`,
		`"previous_state":"public"`, `"new_state":"restricted"`,
		`"previous_item_ids":["segment-1","segment-2"]`, `"item_ids":[]`,
	}
	for _, fragment := range want {
		if !strings.Contains(string(wire), fragment) {
			t.Fatalf("wire missing %s: %s", fragment, wire)
		}
	}

	segmentOnly, err := NewReleaseTrackDownloadPolicyAuditRecord(metadata, "release-1", "track-1", "file-2", AuditStateRestricted, AuditStateRestricted, nil, []string{"segment-1"})
	if err != nil {
		t.Fatal(err)
	}
	if len(segmentOnly.ChangedFields) != 1 || segmentOnly.ChangedFields[0] != auditFieldFileDownloadSegments || segmentOnly.PreviousState != "" || segmentOnly.NewState != "" {
		t.Fatalf("wrong segment-only shape: %#v", segmentOnly)
	}
}

func TestRelationDownloadPolicyBuildersRejectInvalidOrNoOpTransitions(t *testing.T) {
	member := AuditMetadata{AuditID: "00000000-0000-4000-8000-000000000001", OccurredAt: testOccurredAt, RecordActor: RecordActor{Kind: ActorKindMember, MemberID: "member-1"}}
	system := member
	system.RecordActor = RecordActor{Kind: ActorKindSystem, Service: "geul-backend"}
	cases := []struct {
		name  string
		build func() (AuditRecord, error)
	}{
		{"no-op", func() (AuditRecord, error) {
			return NewPostFileBlockDownloadPolicyAuditRecord(member, "post-1", "block-1", "file-1", AuditStatePublic, AuditStatePublic, []string{"segment-1"}, []string{"segment-1"})
		}},
		{"invalid audience", func() (AuditRecord, error) {
			return NewPageFileBlockDownloadPolicyAuditRecord(member, "page-1", "block-1", "file-1", AuditStateDraft, AuditStatePublic, nil, nil)
		}},
		{"invalid segment", func() (AuditRecord, error) {
			return NewWorkFileBlockDownloadPolicyAuditRecord(member, "work-1", "block-1", "file-1", AuditStatePublic, AuditStatePublic, []string{" bad"}, []string{" bad"})
		}},
		{"invalid current segment", func() (AuditRecord, error) {
			return NewWorkFileBlockDownloadPolicyAuditRecord(member, "work-1", "block-1", "file-1", AuditStatePublic, AuditStatePublic, nil, []string{" bad"})
		}},
		{"missing block", func() (AuditRecord, error) {
			return NewProgramEventFileBlockDownloadPolicyAuditRecord(member, "event-1", "", "file-1", AuditStateDisabled, AuditStatePublic, nil, nil)
		}},
		{"system actor", func() (AuditRecord, error) {
			return NewReleaseTrackDownloadPolicyAuditRecord(system, "release-1", "track-1", "file-1", AuditStateDisabled, AuditStatePublic, nil, nil)
		}},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			if _, err := tc.build(); err == nil {
				t.Fatal("invalid relation download policy accepted")
			}
		})
	}
}

func TestRelationDownloadPolicyValidatorRejectsMalformedRecords(t *testing.T) {
	previousSegments := []string{"segment-1"}
	segments := []string{"segment-2"}
	valid := AuditRecord{
		RecordActor:   RecordActor{Kind: ActorKindMember, MemberID: "member-1"},
		ChangedFields: []string{auditFieldFileDownloadAudience},
		ItemID:        "block-1",
		FileID:        "file-1",
		PreviousState: AuditStateDisabled,
		NewState:      AuditStatePublic,
	}
	cases := []struct {
		name   string
		mutate func(*AuditRecord)
	}{
		{"unknown field", func(record *AuditRecord) {
			record.ChangedFields = []string{auditFieldFileDownloadAudience, "unknown"}
		}},
		{"system actor", func(record *AuditRecord) {
			record.RecordActor = RecordActor{Kind: ActorKindSystem, Service: "geul-backend"}
		}},
		{"missing item", func(record *AuditRecord) { record.ItemID = "" }},
		{"invalid audience transition", func(record *AuditRecord) {
			record.PreviousState = AuditStateDraft
		}},
		{"states without audience field", func(record *AuditRecord) {
			record.ChangedFields = []string{auditFieldFileDownloadSegments}
			record.PreviousItemIDs = &previousSegments
			record.ItemIDs = &segments
		}},
		{"missing previous segments", func(record *AuditRecord) {
			record.ChangedFields = []string{auditFieldFileDownloadSegments}
			record.PreviousState = ""
			record.NewState = ""
			record.ItemIDs = &segments
		}},
		{"missing current segments", func(record *AuditRecord) {
			record.ChangedFields = []string{auditFieldFileDownloadSegments}
			record.PreviousState = ""
			record.NewState = ""
			record.PreviousItemIDs = &previousSegments
		}},
		{"unchanged segments", func(record *AuditRecord) {
			record.ChangedFields = []string{auditFieldFileDownloadSegments}
			record.PreviousState = ""
			record.NewState = ""
			record.PreviousItemIDs = &previousSegments
			record.ItemIDs = &previousSegments
		}},
		{"segments without segment field", func(record *AuditRecord) {
			record.PreviousItemIDs = &previousSegments
			record.ItemIDs = &segments
		}},
		{"unsupported attribute", func(record *AuditRecord) { record.Locale = "en" }},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			record := valid
			tc.mutate(&record)
			if err := validateRelationDownloadPolicy(record); err == nil {
				t.Fatal("malformed relation download policy accepted")
			}
		})
	}
}
