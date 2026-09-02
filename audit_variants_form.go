package telemetry

import "fmt"

func init() {
	for _, a := range []AuditAction{AuditFormCreated, AuditFormDeleted, AuditFormSubmissionDeleted} {
		setAuditValidator(a, validateNoAuditAttributes)
	}
	setAuditValidator(AuditFormSubmissionCreated, validateSubmissionCreated)
	setAuditValidator(AuditFormUpdated, validateFormUpdate)
}
func validateSubmissionCreated(r AuditRecord) error {
	if r.ParentID == "" {
		return fmt.Errorf("submission requires parent form")
	}
	if hasAuditAttributesExcept(r, "ParentID") {
		return fmt.Errorf("submission PII attributes forbidden")
	}
	return nil
}
func validateFormUpdate(r AuditRecord) error {
	if len(r.ChangedFields) == 1 {
		switch r.ChangedFields[0] {
		case "status":
			return validateDraftPublished(r)
		case "featured_image":
			return validateFileBinding(r, "featured_image", false)
		case "share_links":
			if (r.ItemOperation != AuditItemOperationCreated && r.ItemOperation != AuditItemOperationDeleted) || (r.ItemScope != AuditItemScopeForm && r.ItemScope != AuditItemScopeDashboard) || !isAuditIdentifier(r.ItemID) {
				return fmt.Errorf("form share link requires scope operation item")
			}
			if hasAuditAttributesExcept(r, "ChangedFields", "ItemOperation", "ItemScope", "ItemID") {
				return fmt.Errorf("share link extras")
			}
			return nil
		}
	}
	return validateChangedOnly("access_period", "auth_required", "direct_public", "duplicate_policy", "limit", "password", "required_role", "slug")(r)
}
