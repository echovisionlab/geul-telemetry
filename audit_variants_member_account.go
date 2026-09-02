package telemetry

import (
	"fmt"
	"strings"
	"unicode/utf8"
)

func init() {
	setAuditValidator(AuditMemberUpdated, validateMemberUpdated)
	setAuditValidator(AuditAccountUpdated, validateAccountUpdated)
	setAuditValidator(AuditAccountDeleted, validateNoAuditAttributes)
	setAuditValidator(AuditPostCreated, validateNoAuditAttributes)
	setAuditValidator(AuditPostUpdated, validatePostUpdated)
	setAuditValidator(AuditPostDeleted, validateNoAuditAttributes)
}

// validateMemberUpdated is a small field-variant dispatcher. Each variant is
// closed by its focused validator; target and actor boundaries remain catalog
// responsibilities in validateAuditCatalogRecord.
func validateMemberUpdated(r AuditRecord) error {
	if r.Kind == ActorKindSystem && !(auditFieldsEqual(r.ChangedFields, "role") ||
		(auditFieldsEqual(r.ChangedFields, "status") && r.PreviousState == AuditStateBanned && r.NewState == AuditStateActive)) {
		return fmt.Errorf("member.updated system actor requires role or unban variant")
	}
	if auditFieldsEqual(r.ChangedFields, "nickname", "onboarded") {
		return validateMemberOnboarded(r)
	}
	if len(r.ChangedFields) == 1 {
		switch r.ChangedFields[0] {
		case "role":
			return validateMemberRole(r)
		case "avatar":
			return validateMemberAvatar(r)
		case "tags":
			return validateMemberTags(r)
		case "status":
			return validateMemberStatus(r)
		}
	}
	return validateMemberProfileOrPreference(r, func(allowed ...string) error {
		return validateOnlyAuditAttributes(r, allowed...)
	})
}

func validateMemberOnboarded(r AuditRecord) error {
	if r.Kind != ActorKindMember || r.MemberID != r.TargetID {
		return fmt.Errorf("member.updated onboarding requires the committed member as actor and target")
	}
	if r.Nickname == "" || strings.TrimSpace(r.Nickname) != r.Nickname || utf8.RuneCountInString(r.Nickname) > 100 {
		return fmt.Errorf("member.updated onboarding requires a trimmed nickname between 1 and 100 characters")
	}
	return validateOnlyAuditAttributes(r, "ChangedFields", "Nickname")
}

func validateMemberRole(r AuditRecord) error {
	if !isBoundedCode(r.PreviousRole) || !isBoundedCode(r.NewRole) || r.PreviousRole == r.NewRole {
		return fmt.Errorf("member.updated role requires distinct previous_role and new_role")
	}
	return validateOnlyAuditAttributes(r, "ChangedFields", "PreviousRole", "NewRole")
}

func validateMemberAvatar(r AuditRecord) error {
	if (r.CollectionOperation != AuditCollectionOperationAdded && r.CollectionOperation != AuditCollectionOperationRemoved) || !isAuditIdentifier(r.AssetID) {
		return fmt.Errorf("member avatar requires collection operation and asset_id")
	}
	return validateOnlyAuditAttributes(r, "ChangedFields", "CollectionOperation", "AssetID")
}

func validateMemberTags(r AuditRecord) error {
	if r.TagIDs == nil {
		return fmt.Errorf("member tags requires tag_ids")
	}
	if err := validateSortedUniqueIdentifiers("tag_ids", *r.TagIDs); err != nil {
		return err
	}
	return validateOnlyAuditAttributes(r, "ChangedFields", "TagIDs")
}

func validateMemberStatus(r AuditRecord) error {
	return validateStateTransition(r, "status", [][2]AuditState{{AuditStateActive, AuditStateBanned}, {AuditStateBanned, AuditStateActive}})
}

func validateAccountUpdated(r AuditRecord) error {
	if len(r.ChangedFields) != 1 {
		return fmt.Errorf("account.updated requires one catalog changed_field")
	}
	switch r.ChangedFields[0] {
	case "canonical_email":
		return validateAccountCanonicalEmail(r)
	case "login_emails":
		return validateAccountLoginEmails(r)
	case "social_logins":
		return validateAccountSocialLogins(r)
	case "passkeys":
		return validateAccountPasskeys(r)
	case "sessions":
		return validateAccountSessions(r)
	case "newsletter_subscription":
		return validateStateTransition(r, "newsletter_subscription", [][2]AuditState{{AuditStateSubscribed, AuditStateUnsubscribed}, {AuditStateUnsubscribed, AuditStateSubscribed}})
	case "deletion_state":
		return validateStateTransition(r, "deletion_state", [][2]AuditState{
			{AuditStateNone, AuditStateConfirmationPending}, {AuditStateCancelled, AuditStateConfirmationPending}, {AuditStateRecovered, AuditStateConfirmationPending},
			{AuditStateConfirmationPending, AuditStateScheduled}, {AuditStateNone, AuditStateScheduled}, {AuditStateCancelled, AuditStateScheduled}, {AuditStateRecovered, AuditStateScheduled},
			{AuditStateScheduled, AuditStateCancelled}, {AuditStateRecoveryConfirmationPending, AuditStateRecovered},
		})
	default:
		return fmt.Errorf("account.updated rejects changed_field %s", r.ChangedFields[0])
	}
}

func validateAccountCanonicalEmail(r AuditRecord) error {
	if !isAuditEmail(r.PreviousEmail) || !isAuditEmail(r.NewEmail) || r.PreviousEmail == r.NewEmail {
		return fmt.Errorf("account.updated canonical_email requires distinct previous_email and new_email")
	}
	return validateOnlyAuditAttributes(r, "ChangedFields", "PreviousEmail", "NewEmail")
}

func validateAccountLoginEmails(r AuditRecord) error {
	if !isAuditEmail(r.Email) || !isAuditCollectionOperation(r.CollectionOperation) {
		return fmt.Errorf("account.updated login_emails requires email and collection_operation")
	}
	return validateOnlyAuditAttributes(r, "ChangedFields", "CollectionOperation", "Email")
}

func validateAccountSocialLogins(r AuditRecord) error {
	if !isBoundedCode(r.Provider) || !isAuditIdentifier(r.ProviderSubject) || !isAuditCollectionOperation(r.CollectionOperation) {
		return fmt.Errorf("account.updated social_logins requires provider, provider_subject, and collection_operation")
	}
	return validateOnlyAuditAttributes(r, "ChangedFields", "CollectionOperation", "Provider", "ProviderSubject")
}

func validateAccountPasskeys(r AuditRecord) error {
	if !isAuditCollectionOperation(r.CollectionOperation) {
		return fmt.Errorf("account.updated passkeys requires collection_operation")
	}
	if err := validateSortedUniqueIdentifiers("passkey_ids", r.PasskeyIDs); err != nil || len(r.PasskeyIDs) == 0 {
		if err != nil {
			return err
		}
		return fmt.Errorf("account.updated passkeys requires passkey_ids")
	}
	return validateOnlyAuditAttributes(r, "ChangedFields", "CollectionOperation", "PasskeyIDs")
}

func validateAccountSessions(r AuditRecord) error {
	if r.CollectionOperation != AuditCollectionOperationRemoved {
		return fmt.Errorf("account.updated sessions requires collection_operation removed")
	}
	if err := validateSortedUniqueUUIDv4("session_ids", r.SessionIDs); err != nil {
		return err
	}
	switch r.SessionScope {
	case AccountSessionScopeCurrent, AccountSessionScopeOne:
		if len(r.SessionIDs) != 1 {
			return fmt.Errorf("account.updated sessions scope %s requires one session_id", r.SessionScope)
		}
	case AccountSessionScopeOthers:
		if len(r.SessionIDs) == 0 {
			return fmt.Errorf("account.updated sessions scope others requires session_ids")
		}
	default:
		return fmt.Errorf("account.updated sessions requires a catalog session_scope")
	}
	return validateOnlyAuditAttributes(r, "ChangedFields", "CollectionOperation", "SessionScope", "SessionIDs")
}

func validatePostUpdated(r AuditRecord) error {
	if err := validateContentVersionActor(r); err != nil {
		return err
	}
	if auditFieldsEqual(r.ChangedFields, "version") {
		return validateContentVersion(r)
	}
	return validatePostConfigurationOrLifecycle(r)
}

func validateContentVersion(r AuditRecord) error {
	if !isAuditIdentifier(r.VersionID) || len(r.ContributorMemberIDs) == 0 {
		return fmt.Errorf("audit action %s requires version_id and contributor_member_ids", r.Action)
	}
	if err := validateSortedUniqueUUIDv4("contributor_member_ids", r.ContributorMemberIDs); err != nil {
		return err
	}
	return validateOnlyAuditAttributes(r, "ChangedFields", "VersionID", "ContributorMemberIDs")
}

func validateStateTransition(r AuditRecord, field string, allowed [][2]AuditState) error {
	if !isBoundedCode(string(r.PreviousState)) || !isBoundedCode(string(r.NewState)) || r.PreviousState == r.NewState {
		return fmt.Errorf("audit action %s field %s requires distinct previous_state and new_state", r.Action, field)
	}
	for _, transition := range allowed {
		if r.PreviousState == transition[0] && r.NewState == transition[1] {
			return validateOnlyAuditAttributes(r, "ChangedFields", "PreviousState", "NewState")
		}
	}
	return fmt.Errorf("audit action %s field %s rejects transition %s to %s", r.Action, field, r.PreviousState, r.NewState)
}

func isAuditCollectionOperation(operation AuditCollectionOperation) bool {
	return operation == AuditCollectionOperationAdded || operation == AuditCollectionOperationRemoved
}

func auditFieldsEqual(actual []string, expected ...string) bool {
	if len(actual) != len(expected) {
		return false
	}
	for i := range expected {
		if actual[i] != expected[i] {
			return false
		}
	}
	return true
}
