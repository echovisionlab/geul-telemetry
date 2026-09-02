package telemetry

import (
	"fmt"
	"slices"
)

func NewPostFileBlockDownloadPolicyAuditRecord(metadata AuditMetadata, postID, blockID, fileID string, previousAudience, newAudience AuditState, previousSegmentIDs, segmentIDs []string) (AuditRecord, error) {
	return newRelationDownloadPolicyAuditRecord(metadata, AuditPostUpdated, postID, blockID, fileID, previousAudience, newAudience, previousSegmentIDs, segmentIDs)
}

func NewPageFileBlockDownloadPolicyAuditRecord(metadata AuditMetadata, pageID, blockID, fileID string, previousAudience, newAudience AuditState, previousSegmentIDs, segmentIDs []string) (AuditRecord, error) {
	return newRelationDownloadPolicyAuditRecord(metadata, AuditPageUpdated, pageID, blockID, fileID, previousAudience, newAudience, previousSegmentIDs, segmentIDs)
}

func NewWorkFileBlockDownloadPolicyAuditRecord(metadata AuditMetadata, workID, blockID, fileID string, previousAudience, newAudience AuditState, previousSegmentIDs, segmentIDs []string) (AuditRecord, error) {
	return newRelationDownloadPolicyAuditRecord(metadata, AuditWorkUpdated, workID, blockID, fileID, previousAudience, newAudience, previousSegmentIDs, segmentIDs)
}

func NewProgramEventFileBlockDownloadPolicyAuditRecord(metadata AuditMetadata, eventID, blockID, fileID string, previousAudience, newAudience AuditState, previousSegmentIDs, segmentIDs []string) (AuditRecord, error) {
	return newRelationDownloadPolicyAuditRecord(metadata, AuditProgramEventUpdated, eventID, blockID, fileID, previousAudience, newAudience, previousSegmentIDs, segmentIDs)
}

func NewReleaseTrackDownloadPolicyAuditRecord(metadata AuditMetadata, releaseID, trackID, fileID string, previousAudience, newAudience AuditState, previousSegmentIDs, segmentIDs []string) (AuditRecord, error) {
	return newRelationDownloadPolicyAuditRecord(metadata, AuditReleaseUpdated, releaseID, trackID, fileID, previousAudience, newAudience, previousSegmentIDs, segmentIDs)
}

func newRelationDownloadPolicyAuditRecord(metadata AuditMetadata, action AuditAction, targetID, itemID, fileID string, previousAudience, newAudience AuditState, previousSegmentIDs, segmentIDs []string) (AuditRecord, error) {
	if !isDownloadAudience(previousAudience) || !isDownloadAudience(newAudience) {
		return AuditRecord{}, fmt.Errorf("download policy requires valid before and after audiences")
	}
	previousSegments := canonicalAuditValues(previousSegmentIDs)
	newSegments := canonicalAuditValues(segmentIDs)
	if previousSegments == nil {
		previousSegments = []string{}
	}
	if newSegments == nil {
		newSegments = []string{}
	}
	if err := validateSortedUniqueIdentifiers("previous_segment_ids", previousSegments); err != nil {
		return AuditRecord{}, err
	}
	if err := validateSortedUniqueIdentifiers("segment_ids", newSegments); err != nil {
		return AuditRecord{}, err
	}
	attributes := AuditRecord{ItemID: itemID, FileID: fileID}
	if previousAudience != newAudience {
		attributes.ChangedFields = append(attributes.ChangedFields, auditFieldFileDownloadAudience)
		attributes.PreviousState = previousAudience
		attributes.NewState = newAudience
	}
	if !slices.Equal(previousSegments, newSegments) {
		attributes.ChangedFields = append(attributes.ChangedFields, auditFieldFileDownloadSegments)
		attributes.PreviousItemIDs = &previousSegments
		attributes.ItemIDs = &newSegments
	}
	attributes.ChangedFields = canonicalAuditValues(attributes.ChangedFields)
	return newCatalogAuditRecord(metadata, action, targetID, attributes)
}
