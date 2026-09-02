package telemetry

func NewEmailTemplateCreatedAuditRecord(m AuditMetadata, id string) (AuditRecord, error) {
	return newCatalogAuditRecord(m, AuditEmailTemplateCreated, id, AuditRecord{})
}
func NewEmailTemplateDeletedAuditRecord(m AuditMetadata, id string) (AuditRecord, error) {
	return newCatalogAuditRecord(m, AuditEmailTemplateDeleted, id, AuditRecord{})
}
func NewEmailTemplateMetadataAuditRecord(m AuditMetadata, id string, f []string) (AuditRecord, error) {
	return newCatalogAuditRecord(m, AuditEmailTemplateUpdated, id, AuditRecord{ChangedFields: canonicalAuditValues(f)})
}
func NewEmailTemplateLayoutRelationAuditRecord(m AuditMetadata, id, previous, next string) (AuditRecord, error) {
	return newCatalogAuditRecord(m, AuditEmailTemplateUpdated, id, AuditRecord{ChangedFields: []string{"layout"}, PreviousItemID: previous, ItemID: next})
}
func NewEmailLayoutCreatedAuditRecord(m AuditMetadata, id string) (AuditRecord, error) {
	return newCatalogAuditRecord(m, AuditEmailLayoutCreated, id, AuditRecord{})
}
func NewEmailLayoutDeletedAuditRecord(m AuditMetadata, id string) (AuditRecord, error) {
	return newCatalogAuditRecord(m, AuditEmailLayoutDeleted, id, AuditRecord{})
}
func NewEmailLayoutMetadataAuditRecord(m AuditMetadata, id string, f []string) (AuditRecord, error) {
	return newCatalogAuditRecord(m, AuditEmailLayoutUpdated, id, AuditRecord{ChangedFields: canonicalAuditValues(f)})
}
func NewEmailEventMappingTemplateAuditRecord(m AuditMetadata, event, previous, next string) (AuditRecord, error) {
	return newCatalogAuditRecord(m, AuditEmailEventMappingUpdated, event, AuditRecord{ChangedFields: []string{"template"}, EventName: event, PreviousItemID: previous, ItemID: next})
}
