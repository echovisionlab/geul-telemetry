package telemetry

// NewWorkMetadataAuditRecord records the exact non-editor Work fields changed
// by one authoritative mutation.
func NewWorkMetadataAuditRecord(m AuditMetadata, id string, fields []string) (AuditRecord, error) {
	return newCatalogAuditRecord(m, AuditWorkUpdated, id, AuditRecord{ChangedFields: canonicalAuditValues(fields)})
}

func NewWorkLifecycleAuditRecord(m AuditMetadata, id string, previous, next AuditState) (AuditRecord, error) {
	return newCatalogAuditRecord(m, AuditWorkUpdated, id, AuditRecord{ChangedFields: []string{"status"}, PreviousState: previous, NewState: next})
}

func NewWorkFeaturedImageAuditRecord(m AuditMetadata, id, assetID string, operation AuditCollectionOperation) (AuditRecord, error) {
	return newCatalogAuditRecord(m, AuditWorkUpdated, id, AuditRecord{ChangedFields: []string{"featured_image"}, AssetID: assetID, CollectionOperation: operation})
}

// NewWorkCreditAuditRecord records one Credit or Credit group child transition.
// Credit names and roles deliberately remain outside the audit payload.
func NewWorkCreditAuditRecord(m AuditMetadata, id, creditID string, operation AuditItemOperation) (AuditRecord, error) {
	return newCatalogAuditRecord(m, AuditWorkUpdated, id, AuditRecord{ChangedFields: []string{"credits"}, ItemID: creditID, ItemOperation: operation})
}

func NewWorkShareLinkAuditRecord(m AuditMetadata, id, linkID string, operation AuditItemOperation) (AuditRecord, error) {
	return newCatalogAuditRecord(m, AuditWorkUpdated, id, AuditRecord{ChangedFields: []string{"share_links"}, ItemID: linkID, ItemOperation: operation})
}

func NewWorkVersionRestoreAuditRecord(m AuditMetadata, id, versionID string) (AuditRecord, error) {
	return newCatalogAuditRecord(m, AuditWorkUpdated, id, AuditRecord{ChangedFields: []string{"version_restore"}, VersionID: versionID})
}
