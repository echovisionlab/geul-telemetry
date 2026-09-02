package telemetry

import (
	"fmt"
	"slices"
)

const (
	auditFieldFileDownloadAudience = "file_download_audience"
	auditFieldFileDownloadSegments = "file_download_audience_segment_ids"
)

func isRelationDownloadPolicyVariant(record AuditRecord) bool {
	return containsAuditField(record.ChangedFields, auditFieldFileDownloadAudience) ||
		containsAuditField(record.ChangedFields, auditFieldFileDownloadSegments)
}

func validateRelationDownloadPolicy(record AuditRecord) error {
	if err := validateChangedSubset(record.ChangedFields, auditFieldFileDownloadAudience, auditFieldFileDownloadSegments); err != nil {
		return err
	}
	if record.Kind != ActorKindMember {
		return fmt.Errorf("download policy requires member actor")
	}
	if !isAuditIdentifier(record.ItemID) || !isAuditIdentifier(record.FileID) {
		return fmt.Errorf("download policy requires exact relation item and File")
	}
	hasAudience := containsAuditField(record.ChangedFields, auditFieldFileDownloadAudience)
	hasSegments := containsAuditField(record.ChangedFields, auditFieldFileDownloadSegments)
	if hasAudience {
		if !isDownloadAudience(record.PreviousState) || !isDownloadAudience(record.NewState) || record.PreviousState == record.NewState {
			return fmt.Errorf("download policy audience requires a distinct catalog state transition")
		}
	} else if record.PreviousState != "" || record.NewState != "" {
		return fmt.Errorf("download policy states require changed_fields file_download_audience")
	}
	if hasSegments {
		if err := validateRequiredOrderedIdentifiers("previous_item_ids", record.PreviousItemIDs); err != nil {
			return err
		}
		if err := validateRequiredOrderedIdentifiers("item_ids", record.ItemIDs); err != nil {
			return err
		}
		if slices.Equal(*record.PreviousItemIDs, *record.ItemIDs) {
			return fmt.Errorf("download policy segment IDs require a distinct transition")
		}
	} else if record.PreviousItemIDs != nil || record.ItemIDs != nil {
		return fmt.Errorf("download policy segment IDs require changed_fields file_download_audience_segment_ids")
	}
	if hasAuditAttributesExcept(record, "ChangedFields", "ItemID", "FileID", "PreviousState", "NewState", "PreviousItemIDs", "ItemIDs") {
		return fmt.Errorf("download policy has unsupported attributes")
	}
	return nil
}

func isDownloadAudience(state AuditState) bool {
	return state == AuditStateDisabled || state == AuditStatePublic || state == AuditStateAuthenticated || state == AuditStateRestricted
}

func validateRequiredOrderedIdentifiers(name string, values *[]string) error {
	if values == nil {
		return fmt.Errorf("%s is required", name)
	}
	return validateSortedUniqueIdentifiers(name, *values)
}
