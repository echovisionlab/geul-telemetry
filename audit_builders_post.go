package telemetry

import "time"

func NewPostConfigurationAuditRecord(m AuditMetadata, id string, f []string) (AuditRecord, error) {
	return newCatalogAuditRecord(m, AuditPostUpdated, id, AuditRecord{ChangedFields: canonicalAuditValues(f)})
}
func NewPostStatusLifecycleAuditRecord(m AuditMetadata, id string, p, n AuditState) (AuditRecord, error) {
	return newCatalogAuditRecord(m, AuditPostUpdated, id, AuditRecord{ChangedFields: []string{"status"}, PreviousState: p, NewState: n})
}
func NewPostScheduleLifecycleAuditRecord(m AuditMetadata, id string, p, n AuditState, at time.Time, tz string) (AuditRecord, error) {
	return newCatalogAuditRecord(m, AuditPostUpdated, id, AuditRecord{ChangedFields: []string{"schedule"}, PreviousState: p, NewState: n, ScheduledAt: &at, ScheduledTimeZone: tz})
}
func NewPostFeaturedImageAuditRecord(m AuditMetadata, id, asset string, op AuditCollectionOperation) (AuditRecord, error) {
	return newCatalogAuditRecord(m, AuditPostUpdated, id, AuditRecord{ChangedFields: []string{"featured_image"}, AssetID: asset, CollectionOperation: op})
}
func NewPostShareLinkAuditRecord(m AuditMetadata, id, item string, op AuditItemOperation) (AuditRecord, error) {
	return newCatalogAuditRecord(m, AuditPostUpdated, id, AuditRecord{ChangedFields: []string{"share_links"}, ItemID: item, ItemOperation: op})
}
func NewPostVersionRestoreAuditRecord(m AuditMetadata, id, version string) (AuditRecord, error) {
	return newCatalogAuditRecord(m, AuditPostUpdated, id, AuditRecord{ChangedFields: []string{"version_restore"}, VersionID: version})
}
func NewPostParticipantAuditRecord(m AuditMetadata, id, memberID string, previous, next AuditRelationship) (AuditRecord, error) {
	fields, err := postParticipantChangedFields(previous, next)
	if err != nil {
		return AuditRecord{}, err
	}
	return newCatalogAuditRecord(m, AuditPostUpdated, id, AuditRecord{ChangedFields: fields, SubjectMemberID: memberID, PreviousRelationship: previous, NewRelationship: next})
}
func NewPostCommentAuditRecord(m AuditMetadata, id, itemID string, op AuditItemOperation) (AuditRecord, error) {
	return newCatalogAuditRecord(m, AuditPostUpdated, id, AuditRecord{ChangedFields: []string{"comments"}, ItemID: itemID, ItemOperation: op})
}
