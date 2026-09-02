package telemetry

func NewClientCreatedAuditRecord(m AuditMetadata, id string) (AuditRecord, error) {
	return newCatalogAuditRecord(m, AuditClientCreated, id, AuditRecord{})
}
func NewClientDeletedAuditRecord(m AuditMetadata, id string) (AuditRecord, error) {
	return newCatalogAuditRecord(m, AuditClientDeleted, id, AuditRecord{})
}
func NewClientMetadataUpdatedAuditRecord(m AuditMetadata, id string, fields []string) (AuditRecord, error) {
	return newCatalogAuditRecord(m, AuditClientUpdated, id, AuditRecord{ChangedFields: canonicalAuditValues(fields)})
}
func NewClientLogoUpdatedAuditRecord(m AuditMetadata, id string, slot AuditAssetSlot, op AuditCollectionOperation, fileID string) (AuditRecord, error) {
	return newCatalogAuditRecord(m, AuditClientUpdated, id, AuditRecord{ChangedFields: []string{"logo"}, AssetSlot: slot, CollectionOperation: op, FileID: fileID})
}
func NewMapPlaceCreatedAuditRecord(m AuditMetadata, id string) (AuditRecord, error) {
	return newCatalogAuditRecord(m, AuditMapPlaceCreated, id, AuditRecord{})
}
func NewMapPlaceDeletedAuditRecord(m AuditMetadata, id string) (AuditRecord, error) {
	return newCatalogAuditRecord(m, AuditMapPlaceDeleted, id, AuditRecord{})
}
func NewMapPlaceMetadataUpdatedAuditRecord(m AuditMetadata, id string, fields []string) (AuditRecord, error) {
	return newCatalogAuditRecord(m, AuditMapPlaceUpdated, id, AuditRecord{ChangedFields: canonicalAuditValues(fields)})
}
func NewMapPlaceImageUpdatedAuditRecord(m AuditMetadata, id string, op AuditCollectionOperation, fileID string) (AuditRecord, error) {
	return newCatalogAuditRecord(m, AuditMapPlaceUpdated, id, AuditRecord{ChangedFields: []string{"image"}, CollectionOperation: op, FileID: fileID})
}
func NewAudienceSegmentCreatedAuditRecord(m AuditMetadata, id string) (AuditRecord, error) {
	return newCatalogAuditRecord(m, AuditAudienceSegmentCreated, id, AuditRecord{})
}
func NewAudienceSegmentConfigUpdatedAuditRecord(m AuditMetadata, id string, fields []string) (AuditRecord, error) {
	return newCatalogAuditRecord(m, AuditAudienceSegmentUpdated, id, AuditRecord{ChangedFields: canonicalAuditValues(fields)})
}
func NewAudienceSegmentLifecycleUpdatedAuditRecord(m AuditMetadata, id string, previous, next AuditState) (AuditRecord, error) {
	return newCatalogAuditRecord(m, AuditAudienceSegmentUpdated, id, AuditRecord{ChangedFields: []string{"status"}, PreviousState: previous, NewState: next})
}
func NewMenuCreatedAuditRecord(m AuditMetadata, id string) (AuditRecord, error) {
	return newCatalogAuditRecord(m, AuditMenuCreated, id, AuditRecord{})
}
func NewMenuDeletedAuditRecord(m AuditMetadata, id string) (AuditRecord, error) {
	return newCatalogAuditRecord(m, AuditMenuDeleted, id, AuditRecord{})
}
func NewMenuSourceUpdatedAuditRecord(m AuditMetadata, id string, fields []string) (AuditRecord, error) {
	return newCatalogAuditRecord(m, AuditMenuUpdated, id, AuditRecord{ChangedFields: canonicalAuditValues(fields)})
}
