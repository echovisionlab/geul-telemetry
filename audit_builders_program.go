package telemetry

func NewProgramEventCreatedAuditRecord(m AuditMetadata, id string) (AuditRecord, error) {
	return newCatalogAuditRecord(m, AuditProgramEventCreated, id, AuditRecord{})
}
func NewProgramEventDeletedAuditRecord(m AuditMetadata, id string) (AuditRecord, error) {
	return newCatalogAuditRecord(m, AuditProgramEventDeleted, id, AuditRecord{})
}
func NewProgramEventMetadataAuditRecord(m AuditMetadata, id string, f []string) (AuditRecord, error) {
	return newCatalogAuditRecord(m, AuditProgramEventUpdated, id, AuditRecord{ChangedFields: canonicalAuditValues(f)})
}
func NewProgramEventPosterAuditRecord(m AuditMetadata, id, file string, op AuditCollectionOperation) (AuditRecord, error) {
	return newCatalogAuditRecord(m, AuditProgramEventUpdated, id, AuditRecord{ChangedFields: []string{"poster"}, FileID: file, CollectionOperation: op})
}
func NewProgramEventChildAuditRecord(m AuditMetadata, id, kind, item string, op AuditItemOperation) (AuditRecord, error) {
	return newCatalogAuditRecord(m, AuditProgramEventUpdated, id, AuditRecord{ChangedFields: []string{kind}, ItemID: item, ItemOperation: op})
}
func NewProgramEventChildOrderAuditRecord(m AuditMetadata, id, kind string, items []string) (AuditRecord, error) {
	v := append([]string(nil), items...)
	return newCatalogAuditRecord(m, AuditProgramEventUpdated, id, AuditRecord{ChangedFields: []string{kind}, ItemIDs: &v})
}
func NewProgramEventLifecycleAuditRecord(m AuditMetadata, id string, p, n AuditState) (AuditRecord, error) {
	return newCatalogAuditRecord(m, AuditProgramEventUpdated, id, AuditRecord{ChangedFields: []string{"status"}, PreviousState: p, NewState: n})
}
func NewProgramEventSeriesCreatedAuditRecord(m AuditMetadata, id string) (AuditRecord, error) {
	return newCatalogAuditRecord(m, AuditProgramEventSeriesCreated, id, AuditRecord{})
}
func NewProgramEventSeriesDeletedAuditRecord(m AuditMetadata, id string) (AuditRecord, error) {
	return newCatalogAuditRecord(m, AuditProgramEventSeriesDeleted, id, AuditRecord{})
}
func NewProgramEventSeriesMetadataAuditRecord(m AuditMetadata, id string, f []string) (AuditRecord, error) {
	return newCatalogAuditRecord(m, AuditProgramEventSeriesUpdated, id, AuditRecord{ChangedFields: canonicalAuditValues(f)})
}
func NewProgramEventSeriesPosterAuditRecord(m AuditMetadata, id, file string, op AuditCollectionOperation) (AuditRecord, error) {
	return newCatalogAuditRecord(m, AuditProgramEventSeriesUpdated, id, AuditRecord{ChangedFields: []string{"poster"}, FileID: file, CollectionOperation: op})
}
func NewProgramEventSeriesLifecycleAuditRecord(m AuditMetadata, id string, p, n AuditState) (AuditRecord, error) {
	return newCatalogAuditRecord(m, AuditProgramEventSeriesUpdated, id, AuditRecord{ChangedFields: []string{"status"}, PreviousState: p, NewState: n})
}
