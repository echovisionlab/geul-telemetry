package telemetry

import (
	"fmt"
)

// Source-locale changes are one reviewed updated variant per owning aggregate.
// The mapping is deliberately explicit: source entity names are not inferred
// from an action string and privacy/terms retain their legal-policy identity.
var sourceLocaleAuditActions = map[AuditAction]bool{
	AuditPostUpdated:          true,
	AuditPageUpdated:          true,
	AuditWorkUpdated:          true,
	AuditPostSeriesUpdated:    true,
	AuditProgramEventUpdated:  true,
	AuditReleaseUpdated:       true,
	AuditArtistUpdated:        true,
	AuditLabelUpdated:         true,
	AuditMenuUpdated:          true,
	AuditCampaignUpdated:      true,
	AuditFormUpdated:          true,
	AuditEmailTemplateUpdated: true,
	AuditEmailLayoutUpdated:   true,
	AuditLegalPolicyUpdated:   true,
}

func validateSourceLocaleAuditVariant(record AuditRecord) (bool, error) {
	if len(record.ChangedFields) != 1 || record.ChangedFields[0] != "source_locale" {
		return false, nil
	}
	if !sourceLocaleAuditActions[record.Action] {
		return true, fmt.Errorf("audit action %s does not support source_locale", record.Action)
	}
	if record.Kind != ActorKindMember {
		return true, fmt.Errorf("source_locale requires a member actor")
	}
	if !isCanonicalAuditLocale(record.PreviousLocale) || !isCanonicalAuditLocale(record.NewLocale) || record.PreviousLocale == record.NewLocale {
		return true, fmt.Errorf("source_locale requires distinct bounded previous_locale and new_locale")
	}
	allowed := []string{"ChangedFields", "PreviousLocale", "NewLocale"}
	if record.Action == AuditLegalPolicyUpdated {
		if !validPolicyIdentity(record) {
			return true, fmt.Errorf("legal policy requires policy_type and version_number")
		}
		allowed = append(allowed, "PolicyType", "VersionNumber")
	}
	if hasAuditAttributesExcept(record, allowed...) {
		return true, fmt.Errorf("source_locale has unsupported attributes")
	}
	return true, nil
}

func newSourceLocaleAuditRecord(metadata AuditMetadata, action AuditAction, targetID, previousLocale, newLocale string) (AuditRecord, error) {
	return newCatalogAuditRecord(metadata, action, targetID, AuditRecord{
		ChangedFields:  []string{"source_locale"},
		PreviousLocale: previousLocale,
		NewLocale:      newLocale,
	})
}

func NewPostSourceLocaleAuditRecord(m AuditMetadata, id, previousLocale, newLocale string) (AuditRecord, error) {
	return newSourceLocaleAuditRecord(m, AuditPostUpdated, id, previousLocale, newLocale)
}
func NewPageSourceLocaleAuditRecord(m AuditMetadata, id, previousLocale, newLocale string) (AuditRecord, error) {
	return newSourceLocaleAuditRecord(m, AuditPageUpdated, id, previousLocale, newLocale)
}
func NewWorkSourceLocaleAuditRecord(m AuditMetadata, id, previousLocale, newLocale string) (AuditRecord, error) {
	return newSourceLocaleAuditRecord(m, AuditWorkUpdated, id, previousLocale, newLocale)
}
func NewPostSeriesSourceLocaleAuditRecord(m AuditMetadata, id, previousLocale, newLocale string) (AuditRecord, error) {
	return newSourceLocaleAuditRecord(m, AuditPostSeriesUpdated, id, previousLocale, newLocale)
}
func NewProgramEventSourceLocaleAuditRecord(m AuditMetadata, id, previousLocale, newLocale string) (AuditRecord, error) {
	return newSourceLocaleAuditRecord(m, AuditProgramEventUpdated, id, previousLocale, newLocale)
}
func NewReleaseSourceLocaleAuditRecord(m AuditMetadata, id, previousLocale, newLocale string) (AuditRecord, error) {
	return newSourceLocaleAuditRecord(m, AuditReleaseUpdated, id, previousLocale, newLocale)
}
func NewArtistSourceLocaleAuditRecord(m AuditMetadata, id, previousLocale, newLocale string) (AuditRecord, error) {
	return newSourceLocaleAuditRecord(m, AuditArtistUpdated, id, previousLocale, newLocale)
}
func NewLabelSourceLocaleAuditRecord(m AuditMetadata, id, previousLocale, newLocale string) (AuditRecord, error) {
	return newSourceLocaleAuditRecord(m, AuditLabelUpdated, id, previousLocale, newLocale)
}
func NewMenuSourceLocaleAuditRecord(m AuditMetadata, id, previousLocale, newLocale string) (AuditRecord, error) {
	return newSourceLocaleAuditRecord(m, AuditMenuUpdated, id, previousLocale, newLocale)
}
func NewCampaignSourceLocaleAuditRecord(m AuditMetadata, id, previousLocale, newLocale string) (AuditRecord, error) {
	return newSourceLocaleAuditRecord(m, AuditCampaignUpdated, id, previousLocale, newLocale)
}
func NewFormSourceLocaleAuditRecord(m AuditMetadata, id, previousLocale, newLocale string) (AuditRecord, error) {
	return newSourceLocaleAuditRecord(m, AuditFormUpdated, id, previousLocale, newLocale)
}
func NewEmailTemplateSourceLocaleAuditRecord(m AuditMetadata, id, previousLocale, newLocale string) (AuditRecord, error) {
	return newSourceLocaleAuditRecord(m, AuditEmailTemplateUpdated, id, previousLocale, newLocale)
}
func NewEmailLayoutSourceLocaleAuditRecord(m AuditMetadata, id, previousLocale, newLocale string) (AuditRecord, error) {
	return newSourceLocaleAuditRecord(m, AuditEmailLayoutUpdated, id, previousLocale, newLocale)
}

func newLegalPolicySourceLocaleAuditRecord(m AuditMetadata, id string, policyType AuditPolicyType, versionNumber int64, previousLocale, newLocale string) (AuditRecord, error) {
	return newLegalPolicyIdentityAuditRecord(m, AuditLegalPolicyUpdated, id, policyType, versionNumber, AuditRecord{
		ChangedFields:  []string{"source_locale"},
		PreviousLocale: previousLocale,
		NewLocale:      newLocale,
	})
}
func NewPrivacySourceLocaleAuditRecord(m AuditMetadata, id string, versionNumber int64, previousLocale, newLocale string) (AuditRecord, error) {
	return newLegalPolicySourceLocaleAuditRecord(m, id, AuditPolicyTypePrivacy, versionNumber, previousLocale, newLocale)
}
func NewTermsSourceLocaleAuditRecord(m AuditMetadata, id string, versionNumber int64, previousLocale, newLocale string) (AuditRecord, error) {
	return newLegalPolicySourceLocaleAuditRecord(m, id, AuditPolicyTypeTerms, versionNumber, previousLocale, newLocale)
}
