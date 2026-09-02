package telemetry

func NewFileCreatedAuditRecord(metadata AuditMetadata, fileID string) (AuditRecord, error) {
	return newCatalogAuditRecord(metadata, AuditFileCreated, fileID, AuditRecord{})
}
func NewFileDeletedAuditRecord(metadata AuditMetadata, fileID string) (AuditRecord, error) {
	return newCatalogAuditRecord(metadata, AuditFileDeleted, fileID, AuditRecord{})
}
func NewFileRenamedAuditRecord(metadata AuditMetadata, fileID string) (AuditRecord, error) {
	return newCatalogAuditRecord(metadata, AuditFileUpdated, fileID, AuditRecord{ChangedFields: []string{"file_name"}})
}
func NewFileMovedAuditRecord(metadata AuditMetadata, fileID, previousParentID, newParentID string) (AuditRecord, error) {
	return newCatalogAuditRecord(metadata, AuditFileUpdated, fileID, AuditRecord{ChangedFields: []string{"folder_id"}, PreviousParentID: previousParentID, NewParentID: newParentID})
}
func NewFileFolderCreatedAuditRecord(metadata AuditMetadata, folderID string) (AuditRecord, error) {
	return newCatalogAuditRecord(metadata, AuditFileFolderCreated, folderID, AuditRecord{})
}
func NewFileFolderDeletedAuditRecord(metadata AuditMetadata, folderID string) (AuditRecord, error) {
	return newCatalogAuditRecord(metadata, AuditFileFolderDeleted, folderID, AuditRecord{})
}
func NewFileFolderRenamedAuditRecord(metadata AuditMetadata, folderID string) (AuditRecord, error) {
	return newCatalogAuditRecord(metadata, AuditFileFolderUpdated, folderID, AuditRecord{ChangedFields: []string{"name"}})
}
func NewFileFolderMovedAuditRecord(metadata AuditMetadata, folderID, previousParentID, newParentID string) (AuditRecord, error) {
	return newCatalogAuditRecord(metadata, AuditFileFolderUpdated, folderID, AuditRecord{ChangedFields: []string{"parent_id"}, PreviousParentID: previousParentID, NewParentID: newParentID})
}
