package telemetry

import "fmt"

func init() {
	registerWorkAuditValidators()
}

// registerWorkAuditValidators binds Work's catalog actions to their reviewed
// variants. It intentionally does not infer behavior from action text.
func registerWorkAuditValidators() {
	setAuditValidator(AuditWorkCreated, validateNoAuditAttributes)
	setAuditValidator(AuditWorkUpdated, validateWorkUpdate)
	setAuditValidator(AuditWorkDeleted, validateNoAuditAttributes)
}

func validateWorkUpdate(r AuditRecord) error {
	if err := validateContentVersionActor(r); err != nil {
		return err
	}
	if isRelationDownloadPolicyVariant(r) {
		return validateRelationDownloadPolicy(r)
	}
	if len(r.ChangedFields) != 1 {
		return validateChangedOnly(workMetadataFields...)(r)
	}
	switch r.ChangedFields[0] {
	case "version":
		return validateWorkVersion(r)
	case "status":
		return validateWorkLifecycle(r)
	case "featured_image":
		return validateWorkFeaturedImage(r)
	case "credits":
		return validateWorkCredit(r)
	case "share_links":
		return validateShareLink(r)
	case "version_restore":
		if !isAuditIdentifier(r.VersionID) || hasAuditAttributesExcept(r, "ChangedFields", "VersionID") {
			return fmt.Errorf("work version restore requires exact version")
		}
		return nil
	default:
		return validateChangedOnly(workMetadataFields...)(r)
	}
}

var workMetadataFields = []string{
	"clients", "featured", "is_present", "map_place_id", "metadata", "month", "slug", "type", "until_month", "until_year", "year",
}

func validateWorkVersion(r AuditRecord) error {
	if !isAuditIdentifier(r.VersionID) || len(r.ContributorMemberIDs) == 0 {
		return fmt.Errorf("work version requires version and contributors")
	}
	if err := validateSortedUniqueUUIDv4("contributor_member_ids", r.ContributorMemberIDs); err != nil {
		return err
	}
	if hasAuditAttributesExcept(r, "ChangedFields", "VersionID", "ContributorMemberIDs") {
		return fmt.Errorf("work version has unsupported attributes")
	}
	return nil
}

func validateWorkLifecycle(r AuditRecord) error {
	valid := map[AuditState]bool{AuditStateDraft: true, AuditStatePublished: true, AuditStateArchived: true}
	if !valid[r.PreviousState] || !valid[r.NewState] || r.PreviousState == r.NewState {
		return fmt.Errorf("work lifecycle requires distinct draft, published, or archived states")
	}
	if hasAuditAttributesExcept(r, "ChangedFields", "PreviousState", "NewState") {
		return fmt.Errorf("work lifecycle has unsupported attributes")
	}
	return nil
}

func validateWorkFeaturedImage(r AuditRecord) error {
	if (r.CollectionOperation != AuditCollectionOperationAdded && r.CollectionOperation != AuditCollectionOperationRemoved) || !isAuditIdentifier(r.AssetID) {
		return fmt.Errorf("work featured image requires operation and asset")
	}
	if hasAuditAttributesExcept(r, "ChangedFields", "CollectionOperation", "AssetID") {
		return fmt.Errorf("work featured image has unsupported attributes")
	}
	return nil
}

func validateWorkCredit(r AuditRecord) error {
	if !isAuditIdentifier(r.ItemID) || (r.ItemOperation != AuditItemOperationCreated && r.ItemOperation != AuditItemOperationUpdated && r.ItemOperation != AuditItemOperationDeleted) {
		return fmt.Errorf("work credit requires operation and exact item")
	}
	if hasAuditAttributesExcept(r, "ChangedFields", "ItemOperation", "ItemID") {
		return fmt.Errorf("work credit has unsupported attributes")
	}
	return nil
}
