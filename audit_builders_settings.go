package telemetry

import "time"

func NewMapThemeCreatedAuditRecord(metadata AuditMetadata, themeID string) (AuditRecord, error) {
	return newCatalogAuditRecord(metadata, AuditMapThemeCreated, themeID, AuditRecord{})
}

func NewMapThemeDeletedAuditRecord(metadata AuditMetadata, themeID string) (AuditRecord, error) {
	return newCatalogAuditRecord(metadata, AuditMapThemeDeleted, themeID, AuditRecord{})
}

func NewMapThemeContentUpdatedAuditRecord(metadata AuditMetadata, themeID string) (AuditRecord, error) {
	return newCatalogAuditRecord(metadata, AuditMapThemeUpdated, themeID, AuditRecord{
		ChangedFields: []string{"content"},
	})
}

func NewLegalPolicyCreatedAuditRecord(metadata AuditMetadata, policyID string, policyType AuditPolicyType, versionNumber int64) (AuditRecord, error) {
	return newLegalPolicyIdentityAuditRecord(metadata, AuditLegalPolicyCreated, policyID, policyType, versionNumber, AuditRecord{})
}

func NewLegalPolicyDeletedAuditRecord(metadata AuditMetadata, policyID string, policyType AuditPolicyType, versionNumber int64) (AuditRecord, error) {
	return newLegalPolicyIdentityAuditRecord(metadata, AuditLegalPolicyDeleted, policyID, policyType, versionNumber, AuditRecord{})
}

func NewLegalPolicyLifecycleAuditRecord(metadata AuditMetadata, policyID string, policyType AuditPolicyType, versionNumber int64, changedFields []string, previousState, newState AuditState, effectiveAt *time.Time) (AuditRecord, error) {
	return newLegalPolicyIdentityAuditRecord(metadata, AuditLegalPolicyUpdated, policyID, policyType, versionNumber, AuditRecord{
		ChangedFields: canonicalAuditValues(changedFields),
		PreviousState: previousState,
		NewState:      newState,
		EffectiveAt:   effectiveAt,
	})
}

func NewLegalPolicyShareLinkAuditRecord(metadata AuditMetadata, policyID string, policyType AuditPolicyType, versionNumber int64, operation AuditItemOperation, itemID string) (AuditRecord, error) {
	return newLegalPolicyIdentityAuditRecord(metadata, AuditLegalPolicyUpdated, policyID, policyType, versionNumber, AuditRecord{
		ChangedFields: []string{"share_links"},
		ItemOperation: operation,
		ItemID:        itemID,
	})
}

func newLegalPolicyIdentityAuditRecord(metadata AuditMetadata, action AuditAction, policyID string, policyType AuditPolicyType, versionNumber int64, attributes AuditRecord) (AuditRecord, error) {
	attributes.PolicyType = policyType
	attributes.VersionNumber = &versionNumber
	return newCatalogAuditRecord(metadata, action, policyID, attributes)
}
