package telemetry

func NewArtistCreatedAuditRecord(m AuditMetadata, id string) (AuditRecord, error) {
	return newCatalogAuditRecord(m, AuditArtistCreated, id, AuditRecord{})
}
func NewArtistDeletedAuditRecord(m AuditMetadata, id string) (AuditRecord, error) {
	return newCatalogAuditRecord(m, AuditArtistDeleted, id, AuditRecord{})
}
func NewArtistLifecycleAuditRecord(m AuditMetadata, id string, previous, next AuditState) (AuditRecord, error) {
	return newCatalogAuditRecord(m, AuditArtistUpdated, id, AuditRecord{ChangedFields: []string{"status"}, PreviousState: previous, NewState: next})
}
func NewArtistGalleryAuditRecord(m AuditMetadata, id string, fileIDs []string) (AuditRecord, error) {
	ids := append([]string(nil), fileIDs...)
	return newCatalogAuditRecord(m, AuditArtistUpdated, id, AuditRecord{ChangedFields: []string{"gallery"}, FileIDs: &ids})
}
func NewArtistParticipantAuditRecord(m AuditMetadata, id, member string, previous, next AuditRelationship) (AuditRecord, error) {
	return newCatalogAuditRecord(m, AuditArtistUpdated, id, AuditRecord{ChangedFields: []string{"participants"}, SubjectMemberID: member, PreviousRelationship: previous, NewRelationship: next})
}
func NewArtistShareLinkAuditRecord(m AuditMetadata, id, item string, op AuditItemOperation) (AuditRecord, error) {
	return newCatalogAuditRecord(m, AuditArtistUpdated, id, AuditRecord{ChangedFields: []string{"share_links"}, ItemID: item, ItemOperation: op})
}
func NewLabelCreatedAuditRecord(m AuditMetadata, id string) (AuditRecord, error) {
	return newCatalogAuditRecord(m, AuditLabelCreated, id, AuditRecord{})
}
func NewLabelDeletedAuditRecord(m AuditMetadata, id string) (AuditRecord, error) {
	return newCatalogAuditRecord(m, AuditLabelDeleted, id, AuditRecord{})
}
func NewLabelLifecycleAuditRecord(m AuditMetadata, id string, p, n AuditState) (AuditRecord, error) {
	return newCatalogAuditRecord(m, AuditLabelUpdated, id, AuditRecord{ChangedFields: []string{"status"}, PreviousState: p, NewState: n})
}
func NewLabelParticipantAuditRecord(m AuditMetadata, id, member string, p, n AuditRelationship) (AuditRecord, error) {
	return newCatalogAuditRecord(m, AuditLabelUpdated, id, AuditRecord{ChangedFields: []string{"participants"}, SubjectMemberID: member, PreviousRelationship: p, NewRelationship: n})
}
func NewLabelLogoAuditRecord(m AuditMetadata, id string, slot AuditAssetSlot, op AuditCollectionOperation, asset string) (AuditRecord, error) {
	return newCatalogAuditRecord(m, AuditLabelUpdated, id, AuditRecord{ChangedFields: []string{"logo"}, AssetSlot: slot, CollectionOperation: op, AssetID: asset})
}
func NewLabelShareLinkAuditRecord(m AuditMetadata, id, item string, op AuditItemOperation) (AuditRecord, error) {
	return newCatalogAuditRecord(m, AuditLabelUpdated, id, AuditRecord{ChangedFields: []string{"share_links"}, ItemID: item, ItemOperation: op})
}
