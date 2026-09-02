package telemetry

func NewMemberProfileUpdatedAuditRecord(m AuditMetadata, id string, fields []string, nickname string) (AuditRecord, error) {
	return newCatalogAuditRecord(m, AuditMemberUpdated, id, AuditRecord{ChangedFields: canonicalAuditValues(fields), Nickname: nickname})
}
func NewMemberAvatarUpdatedAuditRecord(m AuditMetadata, id string, op AuditCollectionOperation, assetID string) (AuditRecord, error) {
	return newCatalogAuditRecord(m, AuditMemberUpdated, id, AuditRecord{ChangedFields: []string{"avatar"}, CollectionOperation: op, AssetID: assetID})
}
func NewMemberPreferencesUpdatedAuditRecord(m AuditMetadata, id string, fields []string, locale, consentID string) (AuditRecord, error) {
	return newCatalogAuditRecord(m, AuditMemberUpdated, id, AuditRecord{ChangedFields: canonicalAuditValues(fields), PreferredLocale: locale, ConsentID: consentID})
}
func NewMemberTagsUpdatedAuditRecord(m AuditMetadata, id string, tagIDs []string) (AuditRecord, error) {
	tags := canonicalAuditValues(tagIDs)
	return newCatalogAuditRecord(m, AuditMemberUpdated, id, AuditRecord{ChangedFields: []string{"tags"}, TagIDs: &tags})
}
func NewMemberTagCreatedAuditRecord(m AuditMetadata, id, name string) (AuditRecord, error) {
	return newCatalogAuditRecord(m, AuditMemberTagCreated, id, AuditRecord{TagName: name})
}
func NewMemberTagDeletedAuditRecord(m AuditMetadata, id, name string) (AuditRecord, error) {
	return newCatalogAuditRecord(m, AuditMemberTagDeleted, id, AuditRecord{TagName: name})
}
func NewAccountNewsletterSubscriptionUpdatedAuditRecord(m AuditMetadata, id string, previous, next AuditState) (AuditRecord, error) {
	return newCatalogAuditRecord(m, AuditAccountUpdated, id, AuditRecord{ChangedFields: []string{"newsletter_subscription"}, PreviousState: previous, NewState: next})
}
