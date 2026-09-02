package telemetry

import "fmt"

func init() {
	for _, action := range []AuditAction{AuditMailAdapterCreated, AuditMailAdapterDeleted, AuditTranslationProviderCreated, AuditTranslationProviderDeleted} {
		setAuditValidator(action, validateNoAuditAttributes)
	}
	setAuditValidator(AuditMailAdapterUpdated, validateChangedOnly("active", "config", "name", "priority", "type"))
	setAuditValidator(AuditEmailSuppressionUpdated, validateEmailSuppressionUpdate)
	setAuditValidator(AuditTranslationSettingsUpdated, validateChangedOnly("default_locale", "protected_terms"))
	setAuditValidator(AuditTranslationProviderUpdated, validateChangedOnly("active", "config", "name", "priority", "type"))
}

func validateNoAuditAttributesExceptChanged(record AuditRecord) error {
	if hasAuditAttributesExcept(record, "ChangedFields") {
		return fmt.Errorf("audit action %s does not allow typed attributes", record.Action)
	}
	return nil
}

func validateEmailSuppressionUpdate(record AuditRecord) error {
	if len(record.ChangedFields) != 1 || record.ChangedFields[0] != "status" {
		return fmt.Errorf("email suppression requires changed_fields status")
	}
	if record.PreviousState != AuditStateActive || record.NewState != AuditStateReleased {
		return fmt.Errorf("email suppression requires active to released")
	}
	if hasAuditAttributesExcept(record, "ChangedFields", "PreviousState", "NewState") {
		return fmt.Errorf("email suppression has unsupported attributes")
	}
	return nil
}
