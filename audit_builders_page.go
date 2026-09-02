package telemetry

// NewPageConfigurationAuditRecord records only the Page settings that actually
// changed: document_layout, show_title, and slug.
func NewPageConfigurationAuditRecord(m AuditMetadata, id string, fields []string) (AuditRecord, error) {
	return newCatalogAuditRecord(m, AuditPageUpdated, id, AuditRecord{ChangedFields: canonicalAuditValues(fields)})
}

func NewPageLifecycleAuditRecord(m AuditMetadata, id string, previous, next AuditState) (AuditRecord, error) {
	return newCatalogAuditRecord(m, AuditPageUpdated, id, AuditRecord{ChangedFields: []string{"status"}, PreviousState: previous, NewState: next})
}

func NewPageFeaturedImageAuditRecord(m AuditMetadata, id, assetID string, operation AuditCollectionOperation) (AuditRecord, error) {
	return newCatalogAuditRecord(m, AuditPageUpdated, id, AuditRecord{ChangedFields: []string{"featured_image"}, AssetID: assetID, CollectionOperation: operation})
}

func NewPageShareLinkAuditRecord(m AuditMetadata, id, linkID string, operation AuditItemOperation) (AuditRecord, error) {
	return newCatalogAuditRecord(m, AuditPageUpdated, id, AuditRecord{ChangedFields: []string{"share_links"}, ItemID: linkID, ItemOperation: operation})
}

func NewPageVersionRestoreAuditRecord(m AuditMetadata, id, versionID string) (AuditRecord, error) {
	return newCatalogAuditRecord(m, AuditPageUpdated, id, AuditRecord{ChangedFields: []string{"version_restore"}, VersionID: versionID})
}
