package telemetry

import "fmt"

func init() {
	setAuditValidator(AuditPageCreated, validateNoAuditAttributes)
	setAuditValidator(AuditPageUpdated, validatePageUpdate)
	setAuditValidator(AuditPageDeleted, validateNoAuditAttributes)
}

func validatePageUpdate(r AuditRecord) error {
	if err := validateContentVersionActor(r); err != nil {
		return err
	}
	if isRelationDownloadPolicyVariant(r) {
		return validateRelationDownloadPolicy(r)
	}
	if len(r.ChangedFields) == 1 {
		switch r.ChangedFields[0] {
		case "version":
			return validatePageVersion(r)
		case "status":
			return validateDraftPublished(r)
		case "featured_image":
			if (r.CollectionOperation != AuditCollectionOperationAdded && r.CollectionOperation != AuditCollectionOperationRemoved) || !isAuditIdentifier(r.AssetID) {
				return fmt.Errorf("page featured image requires operation and asset")
			}
			if hasAuditAttributesExcept(r, "ChangedFields", "CollectionOperation", "AssetID") {
				return fmt.Errorf("page featured image has unsupported attributes")
			}
			return nil
		case "share_links":
			return validateShareLink(r)
		case "version_restore":
			if !isAuditIdentifier(r.VersionID) || hasAuditAttributesExcept(r, "ChangedFields", "VersionID") {
				return fmt.Errorf("page version restore requires exact version")
			}
			return nil
		}
	}
	return validateChangedOnly("document_layout", "show_title", "slug")(r)
}

func validatePageVersion(r AuditRecord) error {
	if !isAuditIdentifier(r.VersionID) || len(r.ContributorMemberIDs) == 0 {
		return fmt.Errorf("page version requires version and contributors")
	}
	if err := validateSortedUniqueUUIDv4("contributor_member_ids", r.ContributorMemberIDs); err != nil {
		return err
	}
	if hasAuditAttributesExcept(r, "ChangedFields", "VersionID", "ContributorMemberIDs") {
		return fmt.Errorf("page version has unsupported attributes")
	}
	return nil
}
