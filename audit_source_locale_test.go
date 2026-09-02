package telemetry

import (
	"strings"
	"testing"
)

func TestSourceLocaleAuditRejectsSystemActorAndInvalidTransition(t *testing.T) {
	member := AuditMetadata{AuditID: "00000000-0000-4000-8000-000000000001", OccurredAt: testOccurredAt, RecordActor: RecordActor{Kind: ActorKindMember, MemberID: "member-1"}}
	record, err := NewPostSeriesSourceLocaleAuditRecord(member, "series-1", "en", "zh-CN")
	if err != nil {
		t.Fatal(err)
	}
	privacy, err := NewPrivacySourceLocaleAuditRecord(member, "privacy-1", 2, "en", "ko")
	if err != nil {
		t.Fatal(err)
	}
	invalid := privacy
	invalid.RecordActor = RecordActor{Kind: ActorKindSystem, Service: string(ServiceBackend)}
	if err := invalid.Validate(); err == nil {
		t.Fatal("source locale accepted a system actor")
	}
	invalid = record
	invalid.NewLocale = "en"
	if err := invalid.Validate(); err == nil {
		t.Fatal("source locale accepted a no-op transition")
	}
	invalid = record
	invalid.PreviousLocale = "ko_KR"
	if err := invalid.Validate(); err == nil {
		t.Fatal("source locale accepted a malformed code")
	}
	invalid = record
	invalid.NewLocale = strings.Repeat("a", 65)
	if err := invalid.Validate(); err == nil {
		t.Fatal("source locale accepted an overlong code")
	}
	invalid = record
	invalid.Action = AuditMapThemeUpdated
	if _, err := validateSourceLocaleAuditVariant(invalid); err == nil {
		t.Fatal("source locale accepted an unlisted action")
	}
	invalid = record
	invalid.PreferredLocale = "ko"
	if _, err := validateSourceLocaleAuditVariant(invalid); err == nil {
		t.Fatal("source locale accepted an unrelated attribute")
	}

	if privacy.PolicyType != AuditPolicyTypePrivacy || privacy.VersionNumber == nil || *privacy.VersionNumber != 2 {
		t.Fatal("privacy source locale lost its legal policy identity")
	}
	privacy.VersionNumber = nil
	if err := privacy.Validate(); err == nil {
		t.Fatal("legal source locale accepted missing version identity")
	}
}
