package telemetry

import "fmt"

func init() {
	for _, a := range []AuditAction{AuditPostSeriesCreated, AuditPostSeriesDeleted, AuditProgramEventTypeCreated, AuditProgramEventTypeDeleted} {
		setAuditValidator(a, validateNoAuditAttributes)
	}
	setAuditValidator(AuditPostSeriesUpdated, validatePostSeriesUpdate)
	setAuditValidator(AuditProgramEventTypeUpdated, validateChangedOnly("requires_place", "requires_stream_url", "slug", "sort_order", "status"))
}
func validatePostSeriesUpdate(r AuditRecord) error {
	if len(r.ChangedFields) == 0 {
		return fmt.Errorf("post series requires variant")
	}
	if err := validateSortedUnique("changed_fields", r.ChangedFields, true); err != nil {
		return err
	}
	if len(r.ChangedFields) == 1 {
		switch r.ChangedFields[0] {
		case "status":
			return validateDraftPublished(r)
		case "managers":
			return validateParticipant(r, AuditRelationshipManager)
		case "posts":
			if !isAuditIdentifier(r.SubjectPostID) || (r.PreviousSeriesID == r.NewSeriesID) {
				return fmt.Errorf("post membership requires subject and distinct series")
			}
			if hasAuditAttributesExcept(r, "ChangedFields", "SubjectPostID", "PreviousSeriesID", "NewSeriesID") {
				return fmt.Errorf("post membership extra attributes")
			}
			return nil
		case "post_order":
			if r.PostIDs == nil {
				return fmt.Errorf("post order requires post_ids")
			}
			if err := validateOrderedIdentifiers("post_ids", *r.PostIDs); err != nil {
				return err
			}
			if hasAuditAttributesExcept(r, "ChangedFields", "PostIDs") {
				return fmt.Errorf("post order extra attributes")
			}
			return nil
		case "featured_image":
			return validateFileBinding(r, "featured_image", false)
		}
	}
	return validateChangedOnly("slug", "source_copy")(r)
}
