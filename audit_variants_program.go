package telemetry

import "fmt"

func init() {
	for _, a := range []AuditAction{AuditProgramEventCreated, AuditProgramEventDeleted, AuditProgramEventSeriesCreated, AuditProgramEventSeriesDeleted} {
		setAuditValidator(a, validateNoAuditAttributes)
	}
	setAuditValidator(AuditProgramEventUpdated, validateProgramEventUpdate)
	setAuditValidator(AuditProgramEventSeriesUpdated, validateProgramEventSeriesUpdate)
}
func validateProgramEventUpdate(r AuditRecord) error {
	if len(r.ChangedFields) == 0 {
		return fmt.Errorf("program event requires a variant")
	}
	if isRelationDownloadPolicyVariant(r) {
		return validateRelationDownloadPolicy(r)
	}
	if len(r.ChangedFields) == 1 {
		switch r.ChangedFields[0] {
		case "poster":
			return validateFileBinding(r, "poster", false)
		case "media", "credits":
			return validateProgramChild(r)
		case "status":
			return validateProgramLifecycle(r)
		}
	}
	if r.Kind == ActorKindSystem {
		return fmt.Errorf("program event does not allow system actor")
	}
	return validateChangedOnly("all_day", "artists", "clients", "ends_at", "external_url", "labels", "location_mode", "map_place_id", "series", "slug", "starts_at", "stream_url", "ticket_url", "timezone", "type")(r)
}
func validateProgramEventSeriesUpdate(r AuditRecord) error {
	if len(r.ChangedFields) == 1 && r.ChangedFields[0] == "poster" {
		return validateFileBinding(r, "poster", false)
	}
	if len(r.ChangedFields) == 1 && r.ChangedFields[0] == "status" {
		return validateDraftPublished(r)
	}
	return validateChangedOnly("description", "slug", "summary", "title")(r)
}
func validateProgramChild(r AuditRecord) error {
	if r.ItemIDs != nil {
		if err := validateOrderedIdentifiers("item_ids", *r.ItemIDs); err != nil {
			return err
		}
		if hasAuditAttributesExcept(r, "ChangedFields", "ItemIDs") {
			return fmt.Errorf("child reorder extras")
		}
		return nil
	}
	if !isAuditIdentifier(r.ItemID) || (r.ItemOperation != AuditItemOperationCreated && r.ItemOperation != AuditItemOperationUpdated && r.ItemOperation != AuditItemOperationDeleted) {
		return fmt.Errorf("child requires operation and item")
	}
	if hasAuditAttributesExcept(r, "ChangedFields", "ItemOperation", "ItemID") {
		return fmt.Errorf("child extra")
	}
	return nil
}
func validateProgramLifecycle(r AuditRecord) error {
	valid := map[AuditState]bool{AuditStateDraft: true, AuditStatePublished: true, AuditStateArchived: true}
	if !valid[r.PreviousState] || !valid[r.NewState] || r.PreviousState == r.NewState {
		return fmt.Errorf("event lifecycle invalid")
	}
	if hasAuditAttributesExcept(r, "ChangedFields", "PreviousState", "NewState") {
		return fmt.Errorf("event lifecycle extras")
	}
	return nil
}
