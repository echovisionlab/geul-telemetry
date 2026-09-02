package telemetry

import "time"

func NewReleaseCreatedAuditRecord(m AuditMetadata, id string) (AuditRecord, error) {
	return newCatalogAuditRecord(m, AuditReleaseCreated, id, AuditRecord{})
}
func NewReleaseDeletedAuditRecord(m AuditMetadata, id string) (AuditRecord, error) {
	return newCatalogAuditRecord(m, AuditReleaseDeleted, id, AuditRecord{})
}
func NewReleaseMetadataAuditRecord(m AuditMetadata, id string, f []string) (AuditRecord, error) {
	return newCatalogAuditRecord(m, AuditReleaseUpdated, id, AuditRecord{ChangedFields: canonicalAuditValues(f)})
}
func NewReleaseTrackAuditRecord(m AuditMetadata, id, item string, op AuditItemOperation) (AuditRecord, error) {
	return newCatalogAuditRecord(m, AuditReleaseUpdated, id, AuditRecord{ChangedFields: []string{"tracks"}, ItemID: item, ItemOperation: op})
}
func NewReleaseTrackOrderAuditRecord(m AuditMetadata, id string, trackIDs []string) (AuditRecord, error) {
	items := append([]string(nil), trackIDs...)
	return newCatalogAuditRecord(m, AuditReleaseUpdated, id, AuditRecord{ChangedFields: []string{"tracks"}, ItemIDs: &items})
}
func NewReleaseTrackAudioAuditRecord(m AuditMetadata, id, trackID, fileID string, op AuditCollectionOperation) (AuditRecord, error) {
	return newCatalogAuditRecord(m, AuditReleaseUpdated, id, AuditRecord{ChangedFields: []string{"track_audio"}, ItemID: trackID, FileID: fileID, CollectionOperation: op})
}
func NewReleaseArtworkAuditRecord(m AuditMetadata, id, file string, op AuditCollectionOperation) (AuditRecord, error) {
	return newCatalogAuditRecord(m, AuditReleaseUpdated, id, AuditRecord{ChangedFields: []string{"artwork"}, FileID: file, CollectionOperation: op})
}
func NewReleaseLifecycleAuditRecord(m AuditMetadata, id string, p, n AuditState) (AuditRecord, error) {
	return newCatalogAuditRecord(m, AuditReleaseUpdated, id, AuditRecord{ChangedFields: []string{"status"}, PreviousState: p, NewState: n})
}
func NewReleaseShareLinkAuditRecord(m AuditMetadata, id, item string, op AuditItemOperation) (AuditRecord, error) {
	return newCatalogAuditRecord(m, AuditReleaseUpdated, id, AuditRecord{ChangedFields: []string{"share_links"}, ItemID: item, ItemOperation: op})
}
func NewCampaignCreatedAuditRecord(m AuditMetadata, id string) (AuditRecord, error) {
	return newCatalogAuditRecord(m, AuditCampaignCreated, id, AuditRecord{})
}
func NewCampaignDeletedAuditRecord(m AuditMetadata, id string) (AuditRecord, error) {
	return newCatalogAuditRecord(m, AuditCampaignDeleted, id, AuditRecord{})
}
func NewCampaignTargetLayoutAuditRecord(m AuditMetadata, id string, f []string) (AuditRecord, error) {
	return newCatalogAuditRecord(m, AuditCampaignUpdated, id, AuditRecord{ChangedFields: canonicalAuditValues(f)})
}

// NewCampaignMetadataAuditRecord records one or more Admin-owned, non-lifecycle
// campaign configuration changes. It deliberately excludes source revisions and
// delivery state transitions, which have dedicated evidence shapes.
func NewCampaignMetadataAuditRecord(m AuditMetadata, id string, f []string) (AuditRecord, error) {
	return newCatalogAuditRecord(m, AuditCampaignUpdated, id, AuditRecord{ChangedFields: canonicalAuditValues(f)})
}
func NewCampaignStatusLifecycleAuditRecord(m AuditMetadata, id string, previous, next AuditState) (AuditRecord, error) {
	return newCatalogAuditRecord(m, AuditCampaignUpdated, id, AuditRecord{ChangedFields: []string{"status"}, PreviousState: previous, NewState: next})
}
func NewCampaignScheduleLifecycleAuditRecord(m AuditMetadata, id string, previous, next AuditState, scheduledAt time.Time) (AuditRecord, error) {
	return newCatalogAuditRecord(m, AuditCampaignUpdated, id, AuditRecord{ChangedFields: []string{"schedule"}, PreviousState: previous, NewState: next, ScheduledAt: &scheduledAt})
}
func NewCampaignDeliveryRunLifecycleAuditRecord(m AuditMetadata, id string, previous, next AuditState, runID string) (AuditRecord, error) {
	return newCatalogAuditRecord(m, AuditCampaignUpdated, id, AuditRecord{ChangedFields: []string{"delivery_run"}, PreviousState: previous, NewState: next, ItemID: runID})
}
