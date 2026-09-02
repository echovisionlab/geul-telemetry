package telemetry

import (
	"fmt"
	"slices"
)

func validatePostConfigurationOrLifecycle(r AuditRecord) error {
	if err := validateContentVersionActor(r); err != nil {
		return err
	}
	if containsAuditField(r.ChangedFields, "authors") || containsAuditField(r.ChangedFields, "collaborators") {
		return validatePostParticipant(r)
	}
	if isRelationDownloadPolicyVariant(r) {
		return validateRelationDownloadPolicy(r)
	}
	if len(r.ChangedFields) == 1 && r.ChangedFields[0] == "featured_image" {
		if (r.CollectionOperation != AuditCollectionOperationAdded && r.CollectionOperation != AuditCollectionOperationRemoved) || !isAuditIdentifier(r.AssetID) {
			return fmt.Errorf("post featured image requires operation and asset")
		}
		if hasAuditAttributesExcept(r, "ChangedFields", "CollectionOperation", "AssetID") {
			return fmt.Errorf("post featured image extras")
		}
		return nil
	}
	if len(r.ChangedFields) == 1 && r.ChangedFields[0] == "share_links" {
		return validateShareLink(r)
	}
	if len(r.ChangedFields) == 1 && r.ChangedFields[0] == "version_restore" {
		if !isAuditIdentifier(r.VersionID) || hasAuditAttributesExcept(r, "ChangedFields", "VersionID") {
			return fmt.Errorf("post version restore requires exact version")
		}
		return nil
	}
	if len(r.ChangedFields) == 1 && r.ChangedFields[0] == "comments" {
		if r.Kind != ActorKindMember || !isAuditIdentifier(r.ItemID) || (r.ItemOperation != AuditItemOperationCreated && r.ItemOperation != AuditItemOperationUpdated && r.ItemOperation != AuditItemOperationDeleted) {
			return fmt.Errorf("post comment requires member actor, item operation, and item_id")
		}
		if hasAuditAttributesExcept(r, "ChangedFields", "ItemOperation", "ItemID") {
			return fmt.Errorf("post comment has unsupported attributes")
		}
		return nil
	}
	if err := validateChangedSubset(r.ChangedFields, "comments_enabled", "document_layout", "map_place_id", "slug", "schedule", "status"); err != nil {
		return err
	}
	life := containsAuditField(r.ChangedFields, "schedule") || containsAuditField(r.ChangedFields, "status")
	if !life {
		return validateNoAuditAttributesExceptChanged(r)
	}
	if r.PreviousState == r.NewState || r.PreviousState == "" || r.NewState == "" {
		return fmt.Errorf("post lifecycle requires transition")
	}
	validLifecycleState := func(state AuditState) bool {
		return state == AuditStateDraft || state == AuditStateScheduled || state == AuditStatePublished || state == AuditStateArchived
	}
	if !validLifecycleState(r.PreviousState) || !validLifecycleState(r.NewState) {
		return fmt.Errorf("post lifecycle rejects state")
	}
	if containsAuditField(r.ChangedFields, "schedule") && (r.ScheduledAt == nil || r.ScheduledTimeZone == "") {
		return fmt.Errorf("post schedule requires time and timezone")
	}
	if !containsAuditField(r.ChangedFields, "schedule") && (r.ScheduledAt != nil || r.ScheduledTimeZone != "") {
		return fmt.Errorf("post schedule attributes require schedule")
	}
	if hasAuditAttributesExcept(r, "ChangedFields", "PreviousState", "NewState", "ScheduledAt", "ScheduledTimeZone") {
		return fmt.Errorf("post lifecycle extra attributes")
	}
	return nil
}

// The collab boundary, not geul-backend, creates Version checkpoints. A
// System actor can therefore append only that exact immutable checkpoint;
// every reviewed Post mutation remains attributable to a Member.
func validateContentVersionActor(r AuditRecord) error {
	version := len(r.ChangedFields) == 1 && r.ChangedFields[0] == "version"
	if version {
		if r.Kind != ActorKindSystem || r.Service != string(ServiceEditorCollab) {
			return fmt.Errorf("content version requires geul-collab system actor")
		}
		return nil
	}
	if r.Kind == ActorKindSystem {
		return fmt.Errorf("content system actor is limited to geul-collab version checkpoints")
	}
	return nil
}

func postParticipantChangedFields(previous, next AuditRelationship) ([]string, error) {
	valid := func(value AuditRelationship) bool {
		return value == AuditRelationshipNone || value == AuditRelationshipAuthor || value == AuditRelationshipCollaborator
	}
	if !valid(previous) || !valid(next) || previous == next {
		return nil, fmt.Errorf("post participant requires a distinct author/collaborator transition")
	}
	hasAuthor := previous == AuditRelationshipAuthor || next == AuditRelationshipAuthor
	hasCollaborator := previous == AuditRelationshipCollaborator || next == AuditRelationshipCollaborator
	if hasAuthor && hasCollaborator {
		return []string{"authors", "collaborators"}, nil
	}
	if hasAuthor {
		return []string{"authors"}, nil
	}
	return []string{"collaborators"}, nil
}

func validatePostParticipant(r AuditRecord) error {
	expected, err := postParticipantChangedFields(r.PreviousRelationship, r.NewRelationship)
	if err != nil {
		return err
	}
	if !slices.Equal(r.ChangedFields, expected) {
		return fmt.Errorf("post participant changed_fields must match relationship transition")
	}
	return validateParticipant(r, AuditRelationshipAuthor, AuditRelationshipCollaborator)
}
