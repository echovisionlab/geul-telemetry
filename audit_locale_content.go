package telemetry

import "fmt"

// Locale-content mutations are reviewed variants of the owning aggregate's
// updated action. Source and translated locales use the same semantic record;
// the action set is explicit so unrelated domains cannot acquire a locale
// content surface by naming convention.
var localeContentAuditActions = map[AuditAction]bool{
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

func validateLocaleContentAuditVariant(record AuditRecord) (bool, error) {
	if len(record.ChangedFields) != 1 || record.ChangedFields[0] != "locale_content" {
		return false, nil
	}
	if !localeContentAuditActions[record.Action] {
		return true, fmt.Errorf("audit action %s does not support locale_content", record.Action)
	}
	if record.Kind != ActorKindMember {
		return true, fmt.Errorf("locale_content requires a member actor")
	}
	if !isCanonicalAuditLocale(record.Locale) {
		return true, fmt.Errorf("locale_content requires a bounded locale")
	}
	if record.ItemOperation != AuditItemOperationCreated &&
		record.ItemOperation != AuditItemOperationUpdated &&
		record.ItemOperation != AuditItemOperationDeleted {
		return true, fmt.Errorf("locale_content requires created, updated, or deleted item operation")
	}
	allowed := []string{"ChangedFields", "Locale", "ItemOperation"}
	if record.Action == AuditLegalPolicyUpdated {
		if !validPolicyIdentity(record) {
			return true, fmt.Errorf("legal policy requires policy_type and version_number")
		}
		allowed = append(allowed, "PolicyType", "VersionNumber")
	}
	if hasAuditAttributesExcept(record, allowed...) {
		return true, fmt.Errorf("locale_content has unsupported attributes")
	}
	return true, nil
}

func newLocaleContentAuditRecord(
	metadata AuditMetadata,
	action AuditAction,
	targetID string,
	locale string,
	operation AuditItemOperation,
) (AuditRecord, error) {
	return newCatalogAuditRecord(metadata, action, targetID, AuditRecord{
		ChangedFields: []string{"locale_content"},
		Locale:        locale,
		ItemOperation: operation,
	})
}

func NewPostLocaleContentAuditRecord(m AuditMetadata, id, locale string, operation AuditItemOperation) (AuditRecord, error) {
	return newLocaleContentAuditRecord(m, AuditPostUpdated, id, locale, operation)
}

func NewPageLocaleContentAuditRecord(m AuditMetadata, id, locale string, operation AuditItemOperation) (AuditRecord, error) {
	return newLocaleContentAuditRecord(m, AuditPageUpdated, id, locale, operation)
}

func NewWorkLocaleContentAuditRecord(m AuditMetadata, id, locale string, operation AuditItemOperation) (AuditRecord, error) {
	return newLocaleContentAuditRecord(m, AuditWorkUpdated, id, locale, operation)
}

func NewPostSeriesLocaleContentAuditRecord(m AuditMetadata, id, locale string, operation AuditItemOperation) (AuditRecord, error) {
	return newLocaleContentAuditRecord(m, AuditPostSeriesUpdated, id, locale, operation)
}

func NewProgramEventLocaleContentAuditRecord(m AuditMetadata, id, locale string, operation AuditItemOperation) (AuditRecord, error) {
	return newLocaleContentAuditRecord(m, AuditProgramEventUpdated, id, locale, operation)
}

func NewReleaseLocaleContentAuditRecord(m AuditMetadata, id, locale string, operation AuditItemOperation) (AuditRecord, error) {
	return newLocaleContentAuditRecord(m, AuditReleaseUpdated, id, locale, operation)
}

func NewArtistLocaleContentAuditRecord(m AuditMetadata, id, locale string, operation AuditItemOperation) (AuditRecord, error) {
	return newLocaleContentAuditRecord(m, AuditArtistUpdated, id, locale, operation)
}

func NewLabelLocaleContentAuditRecord(m AuditMetadata, id, locale string, operation AuditItemOperation) (AuditRecord, error) {
	return newLocaleContentAuditRecord(m, AuditLabelUpdated, id, locale, operation)
}

func NewMenuLocaleContentAuditRecord(m AuditMetadata, id, locale string, operation AuditItemOperation) (AuditRecord, error) {
	return newLocaleContentAuditRecord(m, AuditMenuUpdated, id, locale, operation)
}

func NewCampaignLocaleContentAuditRecord(m AuditMetadata, id, locale string, operation AuditItemOperation) (AuditRecord, error) {
	return newLocaleContentAuditRecord(m, AuditCampaignUpdated, id, locale, operation)
}

func NewFormLocaleContentAuditRecord(m AuditMetadata, id, locale string, operation AuditItemOperation) (AuditRecord, error) {
	return newLocaleContentAuditRecord(m, AuditFormUpdated, id, locale, operation)
}

func NewEmailTemplateLocaleContentAuditRecord(m AuditMetadata, id, locale string, operation AuditItemOperation) (AuditRecord, error) {
	return newLocaleContentAuditRecord(m, AuditEmailTemplateUpdated, id, locale, operation)
}

func NewEmailLayoutLocaleContentAuditRecord(m AuditMetadata, id, locale string, operation AuditItemOperation) (AuditRecord, error) {
	return newLocaleContentAuditRecord(m, AuditEmailLayoutUpdated, id, locale, operation)
}

func newLegalPolicyLocaleContentAuditRecord(
	m AuditMetadata,
	id string,
	policyType AuditPolicyType,
	versionNumber int64,
	locale string,
	operation AuditItemOperation,
) (AuditRecord, error) {
	return newLegalPolicyIdentityAuditRecord(m, AuditLegalPolicyUpdated, id, policyType, versionNumber, AuditRecord{
		ChangedFields: []string{"locale_content"},
		Locale:        locale,
		ItemOperation: operation,
	})
}

func NewPrivacyLocaleContentAuditRecord(m AuditMetadata, id string, versionNumber int64, locale string, operation AuditItemOperation) (AuditRecord, error) {
	return newLegalPolicyLocaleContentAuditRecord(m, id, AuditPolicyTypePrivacy, versionNumber, locale, operation)
}

func NewTermsLocaleContentAuditRecord(m AuditMetadata, id string, versionNumber int64, locale string, operation AuditItemOperation) (AuditRecord, error) {
	return newLegalPolicyLocaleContentAuditRecord(m, id, AuditPolicyTypeTerms, versionNumber, locale, operation)
}
