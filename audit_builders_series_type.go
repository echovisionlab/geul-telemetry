package telemetry

func NewPostSeriesCreatedAuditRecord(m AuditMetadata, id string) (AuditRecord, error) {
	return newCatalogAuditRecord(m, AuditPostSeriesCreated, id, AuditRecord{})
}
func NewPostSeriesDeletedAuditRecord(m AuditMetadata, id string) (AuditRecord, error) {
	return newCatalogAuditRecord(m, AuditPostSeriesDeleted, id, AuditRecord{})
}
func NewPostSeriesSourceMetadataAuditRecord(m AuditMetadata, id string, fields []string) (AuditRecord, error) {
	return newCatalogAuditRecord(m, AuditPostSeriesUpdated, id, AuditRecord{ChangedFields: canonicalAuditValues(fields)})
}
func NewPostSeriesLifecycleAuditRecord(m AuditMetadata, id string, p, n AuditState) (AuditRecord, error) {
	return newCatalogAuditRecord(m, AuditPostSeriesUpdated, id, AuditRecord{ChangedFields: []string{"status"}, PreviousState: p, NewState: n})
}
func NewPostSeriesManagerAuditRecord(m AuditMetadata, id, member string, p, n AuditRelationship) (AuditRecord, error) {
	return newCatalogAuditRecord(m, AuditPostSeriesUpdated, id, AuditRecord{ChangedFields: []string{"managers"}, SubjectMemberID: member, PreviousRelationship: p, NewRelationship: n})
}
func NewPostSeriesMembershipAuditRecord(m AuditMetadata, id, post, previous, next string) (AuditRecord, error) {
	return newCatalogAuditRecord(m, AuditPostSeriesUpdated, id, AuditRecord{ChangedFields: []string{"posts"}, SubjectPostID: post, PreviousSeriesID: previous, NewSeriesID: next})
}
func NewPostSeriesOrderAuditRecord(m AuditMetadata, id string, posts []string) (AuditRecord, error) {
	p := append([]string(nil), posts...)
	return newCatalogAuditRecord(m, AuditPostSeriesUpdated, id, AuditRecord{ChangedFields: []string{"post_order"}, PostIDs: &p})
}
func NewPostSeriesFeaturedImageAuditRecord(m AuditMetadata, id string, op AuditCollectionOperation, file string) (AuditRecord, error) {
	return newCatalogAuditRecord(m, AuditPostSeriesUpdated, id, AuditRecord{ChangedFields: []string{"featured_image"}, CollectionOperation: op, FileID: file})
}
func NewProgramEventTypeCreatedAuditRecord(m AuditMetadata, id string) (AuditRecord, error) {
	return newCatalogAuditRecord(m, AuditProgramEventTypeCreated, id, AuditRecord{})
}
func NewProgramEventTypeDeletedAuditRecord(m AuditMetadata, id string) (AuditRecord, error) {
	return newCatalogAuditRecord(m, AuditProgramEventTypeDeleted, id, AuditRecord{})
}
func NewProgramEventTypeConfigUpdatedAuditRecord(m AuditMetadata, id string, fields []string) (AuditRecord, error) {
	return newCatalogAuditRecord(m, AuditProgramEventTypeUpdated, id, AuditRecord{ChangedFields: canonicalAuditValues(fields)})
}
