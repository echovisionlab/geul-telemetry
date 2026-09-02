package telemetry

import "fmt"

func init() {
	for _, a := range []AuditAction{AuditEmailTemplateCreated, AuditEmailTemplateDeleted, AuditEmailLayoutCreated, AuditEmailLayoutDeleted} {
		setAuditValidator(a, validateEmailAuthoringMemberOnly)
	}
	setAuditValidator(AuditEmailTemplateUpdated, validateEmailTemplateUpdate)
	setAuditValidator(AuditEmailLayoutUpdated, validateEmailLayoutUpdate)
	setAuditValidator(AuditEmailEventMappingUpdated, validateEmailMappingUpdate)
}
func validateEmailTemplateUpdate(r AuditRecord) error {
	if err := validateEmailAuthoringMemberActor(r); err != nil {
		return err
	}
	if len(r.ChangedFields) == 1 && r.ChangedFields[0] == "layout" {
		return validateEmailRelation(r, false)
	}
	return validateChangedOnly("active", "description", "key", "name")(r)
}
func validateEmailLayoutUpdate(r AuditRecord) error {
	if err := validateEmailAuthoringMemberActor(r); err != nil {
		return err
	}
	return validateChangedOnly("key", "name")(r)
}
func validateEmailMappingUpdate(r AuditRecord) error {
	if err := validateEmailAuthoringMemberActor(r); err != nil {
		return err
	}
	if len(r.ChangedFields) != 1 || r.ChangedFields[0] != "template" || !isBoundedCode(r.EventName) || r.ItemID == r.PreviousItemID {
		return fmt.Errorf("event mapping requires event_name and template")
	}
	if r.ItemID != "" && !isAuditIdentifier(r.ItemID) {
		return fmt.Errorf("event mapping has invalid template")
	}
	if r.PreviousItemID != "" && !isAuditIdentifier(r.PreviousItemID) {
		return fmt.Errorf("event mapping has invalid previous template")
	}
	if hasAuditAttributesExcept(r, "ChangedFields", "EventName", "PreviousItemID", "ItemID") {
		return fmt.Errorf("event mapping extra attributes")
	}
	return nil
}
func validateEmailAuthoringMemberOnly(r AuditRecord) error {
	if err := validateEmailAuthoringMemberActor(r); err != nil {
		return err
	}
	return validateNoAuditAttributes(r)
}
func validateEmailAuthoringMemberActor(r AuditRecord) error {
	if r.Kind != ActorKindMember {
		return fmt.Errorf("email authoring mutation requires member actor")
	}
	return nil
}
func validateEmailRelation(r AuditRecord, _ bool) error {
	if r.ItemID == r.PreviousItemID {
		return fmt.Errorf("relation no-op")
	}
	if r.ItemID != "" && !isAuditIdentifier(r.ItemID) {
		return fmt.Errorf("invalid item")
	}
	if r.PreviousItemID != "" && !isAuditIdentifier(r.PreviousItemID) {
		return fmt.Errorf("invalid previous item")
	}
	if hasAuditAttributesExcept(r, "ChangedFields", "PreviousItemID", "ItemID") {
		return fmt.Errorf("relation extra")
	}
	return nil
}
