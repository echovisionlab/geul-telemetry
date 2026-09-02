package telemetry

import "fmt"

func init() {
	for _, a := range []AuditAction{AuditReleaseCreated, AuditReleaseDeleted, AuditCampaignCreated, AuditCampaignDeleted} {
		setAuditValidator(a, validateNoAuditAttributes)
	}
	setAuditValidator(AuditReleaseUpdated, validateReleaseUpdate)
	setAuditValidator(AuditCampaignUpdated, validateCampaignUpdate)
}
func validateReleaseUpdate(r AuditRecord) error {
	if len(r.ChangedFields) == 0 {
		return fmt.Errorf("release requires a variant")
	}
	if isRelationDownloadPolicyVariant(r) {
		return validateRelationDownloadPolicy(r)
	}
	if r.Kind == ActorKindSystem && (len(r.ChangedFields) != 1 || r.ChangedFields[0] != "track_audio") {
		return fmt.Errorf("release system actor is limited to track audio")
	}
	if len(r.ChangedFields) != 1 {
		return validateChangedOnly("artists", "categories", "credits", "date", "formats", "genres", "labels", "links", "slug", "styles", "type")(r)
	}
	switch r.ChangedFields[0] {
	case "tracks":
		return validateProgramChild(r)
	case "track_audio":
		if !isAuditIdentifier(r.ItemID) || !isAuditIdentifier(r.FileID) || (r.CollectionOperation != AuditCollectionOperationAdded && r.CollectionOperation != AuditCollectionOperationRemoved) {
			return fmt.Errorf("track audio requires track, file, and binding operation")
		}
		if hasAuditAttributesExcept(r, "ChangedFields", "ItemID", "FileID", "CollectionOperation") {
			return fmt.Errorf("track audio has unsupported attributes")
		}
		if r.Kind == ActorKindSystem && r.Service != string(ServiceEditorCollab) {
			return fmt.Errorf("release track audio system actor must be geul-collab")
		}
		return nil
	case "artwork":
		return validateFileBinding(r, "artwork", false)
	case "status":
		return validateDraftPublished(r)
	case "share_links":
		return validateShareLink(r)
	}
	return validateChangedOnly("artists", "categories", "credits", "date", "formats", "genres", "labels", "links", "slug", "styles", "type")(r)
}
func validateCampaignUpdate(r AuditRecord) error {
	if err := validateChangedSubset(r.ChangedFields, "layout", "name", "recipient_scope", "segment", "target_mode", "delivery_run", "schedule", "status"); err != nil {
		return err
	}
	if r.Kind == ActorKindSystem {
		if r.Service != string(ServiceBackend) || len(r.ChangedFields) != 1 || r.ChangedFields[0] != "status" {
			return fmt.Errorf("campaign system actor is limited to terminal status")
		}
		if r.NewState != AuditStateSent && r.NewState != AuditStateFailed {
			return fmt.Errorf("campaign system status must be terminal")
		}
	}
	life := containsAuditField(r.ChangedFields, "delivery_run") || containsAuditField(r.ChangedFields, "schedule") || containsAuditField(r.ChangedFields, "status")
	if !life {
		return validateNoAuditAttributesExceptChanged(r)
	}
	if r.PreviousState == r.NewState || r.PreviousState == "" || r.NewState == "" {
		return fmt.Errorf("campaign lifecycle requires transition")
	}
	if containsAuditField(r.ChangedFields, "schedule") && r.ScheduledAt == nil {
		return fmt.Errorf("campaign schedule requires scheduled_at")
	}
	if !containsAuditField(r.ChangedFields, "schedule") && (r.ScheduledAt != nil || r.ScheduledTimeZone != "") {
		return fmt.Errorf("campaign schedule attributes require changed_fields schedule")
	}
	if r.ScheduledTimeZone != "" {
		return fmt.Errorf("campaign schedule does not store a time zone")
	}
	if containsAuditField(r.ChangedFields, "delivery_run") && !isAuditIdentifier(r.ItemID) {
		return fmt.Errorf("campaign delivery run requires item_id")
	}
	if hasAuditAttributesExcept(r, "ChangedFields", "PreviousState", "NewState", "ScheduledAt", "ItemID") {
		return fmt.Errorf("campaign lifecycle extras")
	}
	return nil
}
