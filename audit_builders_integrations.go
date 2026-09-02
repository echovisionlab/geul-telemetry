package telemetry

func NewMailAdapterCreatedAuditRecord(m AuditMetadata, id string) (AuditRecord, error) {
	return newCatalogAuditRecord(m, AuditMailAdapterCreated, id, AuditRecord{})
}
func NewMailAdapterDeletedAuditRecord(m AuditMetadata, id string) (AuditRecord, error) {
	return newCatalogAuditRecord(m, AuditMailAdapterDeleted, id, AuditRecord{})
}
func NewMailAdapterConfigUpdatedAuditRecord(m AuditMetadata, id string, fields []string) (AuditRecord, error) {
	return newIntegrationChangedAuditRecord(m, AuditMailAdapterUpdated, id, fields)
}
func NewEmailSuppressionReleasedAuditRecord(m AuditMetadata, id string) (AuditRecord, error) {
	return newCatalogAuditRecord(m, AuditEmailSuppressionUpdated, id, AuditRecord{ChangedFields: []string{"status"}, PreviousState: AuditStateActive, NewState: AuditStateReleased})
}
func NewTranslationSettingsUpdatedAuditRecord(m AuditMetadata, fields []string) (AuditRecord, error) {
	return newIntegrationChangedAuditRecord(m, AuditTranslationSettingsUpdated, "1", fields)
}
func NewTranslationProviderCreatedAuditRecord(m AuditMetadata, id string) (AuditRecord, error) {
	return newCatalogAuditRecord(m, AuditTranslationProviderCreated, id, AuditRecord{})
}
func NewTranslationProviderDeletedAuditRecord(m AuditMetadata, id string) (AuditRecord, error) {
	return newCatalogAuditRecord(m, AuditTranslationProviderDeleted, id, AuditRecord{})
}
func NewTranslationProviderConfigUpdatedAuditRecord(m AuditMetadata, id string, fields []string) (AuditRecord, error) {
	return newIntegrationChangedAuditRecord(m, AuditTranslationProviderUpdated, id, fields)
}
func newIntegrationChangedAuditRecord(m AuditMetadata, action AuditAction, id string, fields []string) (AuditRecord, error) {
	return newCatalogAuditRecord(m, action, id, AuditRecord{ChangedFields: canonicalAuditValues(fields)})
}
