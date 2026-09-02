package telemetry

func NewCategoryCreatedAuditRecord(m AuditMetadata, id string) (AuditRecord, error) {
	return newReferenceLifecycleAuditRecord(m, AuditCategoryCreated, id)
}
func NewCategoryDeletedAuditRecord(m AuditMetadata, id string) (AuditRecord, error) {
	return newReferenceLifecycleAuditRecord(m, AuditCategoryDeleted, id)
}
func NewCategoryMetadataUpdatedAuditRecord(m AuditMetadata, id string, fields []string) (AuditRecord, error) {
	return newReferenceMetadataAuditRecord(m, AuditCategoryUpdated, id, fields)
}
func NewTagCreatedAuditRecord(m AuditMetadata, id string) (AuditRecord, error) {
	return newReferenceLifecycleAuditRecord(m, AuditTagCreated, id)
}
func NewTagDeletedAuditRecord(m AuditMetadata, id string) (AuditRecord, error) {
	return newReferenceLifecycleAuditRecord(m, AuditTagDeleted, id)
}
func NewTagMetadataUpdatedAuditRecord(m AuditMetadata, id string, fields []string) (AuditRecord, error) {
	return newReferenceMetadataAuditRecord(m, AuditTagUpdated, id, fields)
}
func NewGenreCreatedAuditRecord(m AuditMetadata, id string) (AuditRecord, error) {
	return newReferenceLifecycleAuditRecord(m, AuditGenreCreated, id)
}
func NewGenreDeletedAuditRecord(m AuditMetadata, id string) (AuditRecord, error) {
	return newReferenceLifecycleAuditRecord(m, AuditGenreDeleted, id)
}
func NewGenreMetadataUpdatedAuditRecord(m AuditMetadata, id string, fields []string) (AuditRecord, error) {
	return newReferenceMetadataAuditRecord(m, AuditGenreUpdated, id, fields)
}
func NewStyleCreatedAuditRecord(m AuditMetadata, id string) (AuditRecord, error) {
	return newReferenceLifecycleAuditRecord(m, AuditStyleCreated, id)
}
func NewStyleDeletedAuditRecord(m AuditMetadata, id string) (AuditRecord, error) {
	return newReferenceLifecycleAuditRecord(m, AuditStyleDeleted, id)
}
func NewStyleMetadataUpdatedAuditRecord(m AuditMetadata, id string, fields []string) (AuditRecord, error) {
	return newReferenceMetadataAuditRecord(m, AuditStyleUpdated, id, fields)
}
func NewFormatCreatedAuditRecord(m AuditMetadata, id string) (AuditRecord, error) {
	return newReferenceLifecycleAuditRecord(m, AuditFormatCreated, id)
}
func NewFormatDeletedAuditRecord(m AuditMetadata, id string) (AuditRecord, error) {
	return newReferenceLifecycleAuditRecord(m, AuditFormatDeleted, id)
}
func NewFormatMetadataUpdatedAuditRecord(m AuditMetadata, id string, fields []string) (AuditRecord, error) {
	return newReferenceMetadataAuditRecord(m, AuditFormatUpdated, id, fields)
}

func newReferenceLifecycleAuditRecord(m AuditMetadata, action AuditAction, id string) (AuditRecord, error) {
	return newCatalogAuditRecord(m, action, id, AuditRecord{})
}
func newReferenceMetadataAuditRecord(m AuditMetadata, action AuditAction, id string, fields []string) (AuditRecord, error) {
	return newCatalogAuditRecord(m, action, id, AuditRecord{ChangedFields: canonicalAuditValues(fields)})
}
