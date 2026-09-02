package telemetry

func NewFormCreatedAuditRecord(m AuditMetadata, id string) (AuditRecord, error) {
	return newCatalogAuditRecord(m, AuditFormCreated, id, AuditRecord{})
}
func NewFormDeletedAuditRecord(m AuditMetadata, id string) (AuditRecord, error) {
	return newCatalogAuditRecord(m, AuditFormDeleted, id, AuditRecord{})
}
func NewFormSettingsAuditRecord(m AuditMetadata, id string, f []string) (AuditRecord, error) {
	return newCatalogAuditRecord(m, AuditFormUpdated, id, AuditRecord{ChangedFields: canonicalAuditValues(f)})
}
func NewFormLifecycleAuditRecord(m AuditMetadata, id string, p, n AuditState) (AuditRecord, error) {
	return newCatalogAuditRecord(m, AuditFormUpdated, id, AuditRecord{ChangedFields: []string{"status"}, PreviousState: p, NewState: n})
}
func NewFormFeaturedImageAuditRecord(m AuditMetadata, id, file string, op AuditCollectionOperation) (AuditRecord, error) {
	return newCatalogAuditRecord(m, AuditFormUpdated, id, AuditRecord{ChangedFields: []string{"featured_image"}, FileID: file, CollectionOperation: op})
}
func NewFormShareLinkAuditRecord(m AuditMetadata, id, item string, scope AuditItemScope, op AuditItemOperation) (AuditRecord, error) {
	return newCatalogAuditRecord(m, AuditFormUpdated, id, AuditRecord{ChangedFields: []string{"share_links"}, ItemID: item, ItemScope: scope, ItemOperation: op})
}
func NewFormSubmissionCreatedAuditRecord(m AuditMetadata, submissionID, formID string) (AuditRecord, error) {
	return newCatalogAuditRecord(m, AuditFormSubmissionCreated, submissionID, AuditRecord{ParentID: formID})
}
func NewFormSubmissionDeletedAuditRecord(m AuditMetadata, id string) (AuditRecord, error) {
	return newCatalogAuditRecord(m, AuditFormSubmissionDeleted, id, AuditRecord{})
}
