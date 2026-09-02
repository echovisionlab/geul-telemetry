package telemetry

import (
	"encoding/json"
	"os"
	"testing"
)

type localeContentFixture struct {
	Case          string             `json:"case"`
	Action        AuditAction        `json:"action"`
	TargetType    string             `json:"target_type"`
	TargetID      string             `json:"target_id"`
	Locale        string             `json:"locale"`
	ItemOperation AuditItemOperation `json:"item_operation"`
	PolicyType    AuditPolicyType    `json:"policy_type,omitempty"`
	VersionNumber int64              `json:"version_number,omitempty"`
}

func TestLocaleContentSemanticBuilders(t *testing.T) {
	contents, err := os.ReadFile("fixtures/locale-content-audit.json")
	if err != nil {
		t.Fatal(err)
	}
	var fixtures []localeContentFixture
	if err := json.Unmarshal(contents, &fixtures); err != nil {
		t.Fatal(err)
	}
	metadata := AuditMetadata{
		AuditID:     "00000000-0000-4000-8000-000000000001",
		OccurredAt:  testOccurredAt,
		RecordActor: RecordActor{Kind: ActorKindMember, MemberID: "member-1"},
	}
	for _, fixture := range fixtures {
		fixture := fixture
		t.Run(fixture.Case, func(t *testing.T) {
			record, err := buildLocaleContentFixture(metadata, fixture)
			if err != nil {
				t.Fatal(err)
			}
			if err := record.Validate(); err != nil {
				t.Fatal(err)
			}
			if record.Action != fixture.Action || record.TargetType != fixture.TargetType || record.TargetID != fixture.TargetID ||
				record.Locale != fixture.Locale || record.ItemOperation != fixture.ItemOperation ||
				len(record.ChangedFields) != 1 || record.ChangedFields[0] != "locale_content" ||
				record.PolicyType != fixture.PolicyType {
				t.Fatalf("record = %#v, fixture = %#v", record, fixture)
			}
			if fixture.VersionNumber == 0 {
				if record.VersionNumber != nil {
					t.Fatalf("version number = %v, want absent", record.VersionNumber)
				}
			} else if record.VersionNumber == nil || *record.VersionNumber != fixture.VersionNumber {
				t.Fatalf("version number = %v, want %d", record.VersionNumber, fixture.VersionNumber)
			}
		})
	}
}

func TestLocaleContentVariantRejectsUnreviewedShapes(t *testing.T) {
	version := int64(1)
	base := AuditRecord{
		Action:        AuditPostUpdated,
		RecordActor:   RecordActor{Kind: ActorKindMember, MemberID: "member-1"},
		ChangedFields: []string{"locale_content"},
		Locale:        "ko",
		ItemOperation: AuditItemOperationUpdated,
	}
	tests := []AuditRecord{
		func() AuditRecord { record := base; record.Action = AuditCategoryUpdated; return record }(),
		func() AuditRecord {
			record := base
			record.Kind = ActorKindSystem
			record.Service = string(ServiceBackend)
			return record
		}(),
		func() AuditRecord { record := base; record.Locale = "ko_KR"; return record }(),
		func() AuditRecord { record := base; record.ItemOperation = "replace"; return record }(),
		func() AuditRecord { record := base; record.AssetID = "unexpected"; return record }(),
		func() AuditRecord { record := base; record.Action = AuditLegalPolicyUpdated; return record }(),
		func() AuditRecord {
			record := base
			record.Action = AuditLegalPolicyUpdated
			record.PolicyType = AuditPolicyTypeTerms
			record.VersionNumber = &version
			record.AssetID = "unexpected"
			return record
		}(),
	}
	for index, record := range tests {
		if handled, err := validateLocaleContentAuditVariant(record); !handled || err == nil {
			t.Fatalf("case %d: handled = %v, error = %v", index, handled, err)
		}
	}
}

func buildLocaleContentFixture(metadata AuditMetadata, fixture localeContentFixture) (AuditRecord, error) {
	switch fixture.Action {
	case AuditPostUpdated:
		return NewPostLocaleContentAuditRecord(metadata, fixture.TargetID, fixture.Locale, fixture.ItemOperation)
	case AuditPageUpdated:
		return NewPageLocaleContentAuditRecord(metadata, fixture.TargetID, fixture.Locale, fixture.ItemOperation)
	case AuditWorkUpdated:
		return NewWorkLocaleContentAuditRecord(metadata, fixture.TargetID, fixture.Locale, fixture.ItemOperation)
	case AuditPostSeriesUpdated:
		return NewPostSeriesLocaleContentAuditRecord(metadata, fixture.TargetID, fixture.Locale, fixture.ItemOperation)
	case AuditProgramEventUpdated:
		return NewProgramEventLocaleContentAuditRecord(metadata, fixture.TargetID, fixture.Locale, fixture.ItemOperation)
	case AuditReleaseUpdated:
		return NewReleaseLocaleContentAuditRecord(metadata, fixture.TargetID, fixture.Locale, fixture.ItemOperation)
	case AuditArtistUpdated:
		return NewArtistLocaleContentAuditRecord(metadata, fixture.TargetID, fixture.Locale, fixture.ItemOperation)
	case AuditLabelUpdated:
		return NewLabelLocaleContentAuditRecord(metadata, fixture.TargetID, fixture.Locale, fixture.ItemOperation)
	case AuditMenuUpdated:
		return NewMenuLocaleContentAuditRecord(metadata, fixture.TargetID, fixture.Locale, fixture.ItemOperation)
	case AuditCampaignUpdated:
		return NewCampaignLocaleContentAuditRecord(metadata, fixture.TargetID, fixture.Locale, fixture.ItemOperation)
	case AuditFormUpdated:
		return NewFormLocaleContentAuditRecord(metadata, fixture.TargetID, fixture.Locale, fixture.ItemOperation)
	case AuditEmailTemplateUpdated:
		return NewEmailTemplateLocaleContentAuditRecord(metadata, fixture.TargetID, fixture.Locale, fixture.ItemOperation)
	case AuditEmailLayoutUpdated:
		return NewEmailLayoutLocaleContentAuditRecord(metadata, fixture.TargetID, fixture.Locale, fixture.ItemOperation)
	case AuditLegalPolicyUpdated:
		if fixture.PolicyType == AuditPolicyTypePrivacy {
			return NewPrivacyLocaleContentAuditRecord(metadata, fixture.TargetID, fixture.VersionNumber, fixture.Locale, fixture.ItemOperation)
		}
		return NewTermsLocaleContentAuditRecord(metadata, fixture.TargetID, fixture.VersionNumber, fixture.Locale, fixture.ItemOperation)
	default:
		return AuditRecord{}, nil
	}
}
