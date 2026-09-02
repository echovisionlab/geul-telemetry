package telemetry

import "fmt"

func init() {
	for _, action := range []AuditAction{AuditArtistCreated, AuditArtistDeleted, AuditLabelCreated, AuditLabelDeleted} {
		setAuditValidator(action, validateNoAuditAttributes)
	}
	setAuditValidator(AuditArtistUpdated, validateArtistUpdate)
	setAuditValidator(AuditLabelUpdated, validateLabelUpdate)
}
func validateArtistUpdate(r AuditRecord) error {
	if len(r.ChangedFields) != 1 {
		return fmt.Errorf("artist update requires one variant")
	}
	switch r.ChangedFields[0] {
	case "status":
		if err := validateArtistLabelMemberActor(r); err != nil {
			return err
		}
		return validateDraftPublished(r)
	case "gallery":
		if err := validateArtistLabelMemberActor(r); err != nil {
			return err
		}
		if r.FileIDs == nil {
			return fmt.Errorf("artist gallery requires file_ids")
		}
		if err := validateOrderedIdentifiers("file_ids", *r.FileIDs); err != nil {
			return err
		}
		if hasAuditAttributesExcept(r, "ChangedFields", "FileIDs") {
			return fmt.Errorf("artist gallery has unsupported attributes")
		}
		return nil
	case "participants":
		if err := validateArtistLabelMemberActor(r); err != nil {
			return err
		}
		return validateParticipant(r, AuditRelationshipOwner, AuditRelationshipManager)
	case "share_links":
		if err := validateArtistLabelMemberActor(r); err != nil {
			return err
		}
		return validateShareLink(r)
	}
	return fmt.Errorf("artist update rejects variant")
}
func validateLabelUpdate(r AuditRecord) error {
	if len(r.ChangedFields) != 1 {
		return fmt.Errorf("label update requires one variant")
	}
	switch r.ChangedFields[0] {
	case "status":
		if err := validateArtistLabelMemberActor(r); err != nil {
			return err
		}
		return validateDraftPublished(r)
	case "participants":
		if err := validateArtistLabelMemberActor(r); err != nil {
			return err
		}
		return validateParticipant(r, AuditRelationshipOwner, AuditRelationshipManager)
	case "logo":
		if err := validateArtistLabelMemberActor(r); err != nil {
			return err
		}
		if r.AssetSlot != AuditAssetSlotLight && r.AssetSlot != AuditAssetSlotDark {
			return fmt.Errorf("label logo requires slot")
		}
		if (r.CollectionOperation != AuditCollectionOperationAdded && r.CollectionOperation != AuditCollectionOperationRemoved) || !isAuditIdentifier(r.AssetID) {
			return fmt.Errorf("label logo requires operation and asset")
		}
		if hasAuditAttributesExcept(r, "ChangedFields", "AssetSlot", "CollectionOperation", "AssetID") {
			return fmt.Errorf("label logo has unsupported attributes")
		}
		return nil
	case "share_links":
		if err := validateArtistLabelMemberActor(r); err != nil {
			return err
		}
		return validateShareLink(r)
	}
	return fmt.Errorf("label update rejects variant")
}
func validateArtistLabelMemberActor(r AuditRecord) error {
	if r.Kind != ActorKindMember {
		return fmt.Errorf("artist and label mutation variants require member actor")
	}
	return nil
}
func validateDraftPublished(r AuditRecord) error {
	if !((r.PreviousState == AuditStateDraft && r.NewState == AuditStatePublished) || (r.PreviousState == AuditStatePublished && r.NewState == AuditStateDraft)) {
		return fmt.Errorf("lifecycle requires draft/published transition")
	}
	if hasAuditAttributesExcept(r, "ChangedFields", "PreviousState", "NewState") {
		return fmt.Errorf("lifecycle has unsupported attributes")
	}
	return nil
}
func validateParticipant(r AuditRecord, allowed ...AuditRelationship) error {
	if !isAuditIdentifier(r.SubjectMemberID) || r.PreviousRelationship == r.NewRelationship {
		return fmt.Errorf("participant requires subject and transition")
	}
	valid := func(v AuditRelationship) bool {
		if v == AuditRelationshipNone {
			return true
		}
		for _, a := range allowed {
			if v == a {
				return true
			}
		}
		return false
	}
	if !valid(r.PreviousRelationship) || !valid(r.NewRelationship) {
		return fmt.Errorf("participant relationship rejected")
	}
	if hasAuditAttributesExcept(r, "ChangedFields", "SubjectMemberID", "PreviousRelationship", "NewRelationship") {
		return fmt.Errorf("participant has unsupported attributes")
	}
	return nil
}
func validateShareLink(r AuditRecord) error {
	if (r.ItemOperation != AuditItemOperationCreated && r.ItemOperation != AuditItemOperationDeleted) || !isAuditIdentifier(r.ItemID) {
		return fmt.Errorf("share link requires operation and item")
	}
	if hasAuditAttributesExcept(r, "ChangedFields", "ItemOperation", "ItemID") {
		return fmt.Errorf("share link has unsupported attributes")
	}
	return nil
}
func validateOrderedIdentifiers(name string, values []string) error {
	seen := make(map[string]struct{}, len(values))
	for _, v := range values {
		if !isAuditIdentifier(v) {
			return fmt.Errorf("%s contains invalid identifier", name)
		}
		if _, exists := seen[v]; exists {
			return fmt.Errorf("%s contains duplicate identifier", name)
		}
		seen[v] = struct{}{}
	}
	return nil
}
