package telemetry

import (
	"fmt"
	"slices"
	"time"
)

type AuditMetadata struct {
	AuditID    string
	OccurredAt time.Time
	Correlation
	RecordActor
}

func NewSiteSettingsUpdatedAuditRecord(metadata AuditMetadata, changedFields []string) (AuditRecord, error) {
	return newCatalogAuditRecord(metadata, AuditSiteSettingsUpdated, "1", AuditRecord{ChangedFields: canonicalAuditValues(changedFields)})
}

func NewMemberOnboardingCompletedAuditRecord(metadata AuditMetadata, memberID, nickname string) (AuditRecord, error) {
	return newCatalogAuditRecord(metadata, AuditMemberUpdated, memberID, AuditRecord{ChangedFields: []string{"nickname", "onboarded"}, Nickname: nickname})
}

func NewMemberRoleUpdatedAuditRecord(metadata AuditMetadata, memberID, previousRole, newRole string) (AuditRecord, error) {
	return newCatalogAuditRecord(metadata, AuditMemberUpdated, memberID, AuditRecord{ChangedFields: []string{"role"}, PreviousRole: previousRole, NewRole: newRole})
}

func NewMemberBannedAuditRecord(metadata AuditMetadata, memberID string) (AuditRecord, error) {
	return newMemberStatusUpdatedAuditRecord(metadata, memberID, AuditStateActive, AuditStateBanned)
}

func NewMemberUnbannedAuditRecord(metadata AuditMetadata, memberID string) (AuditRecord, error) {
	return newMemberStatusUpdatedAuditRecord(metadata, memberID, AuditStateBanned, AuditStateActive)
}

func newMemberStatusUpdatedAuditRecord(metadata AuditMetadata, memberID string, previousState, newState AuditState) (AuditRecord, error) {
	return newCatalogAuditRecord(metadata, AuditMemberUpdated, memberID, AuditRecord{ChangedFields: []string{"status"}, PreviousState: previousState, NewState: newState})
}

func NewPostVersionCreatedAuditRecord(metadata AuditMetadata, postID, versionID string, contributorMemberIDs []string) (AuditRecord, error) {
	return newCatalogAuditRecord(metadata, AuditPostUpdated, postID, AuditRecord{ChangedFields: []string{"version"}, VersionID: versionID, ContributorMemberIDs: canonicalAuditValues(contributorMemberIDs)})
}

func NewPageVersionCreatedAuditRecord(metadata AuditMetadata, pageID, versionID string, contributorMemberIDs []string) (AuditRecord, error) {
	return newCatalogAuditRecord(metadata, AuditPageUpdated, pageID, AuditRecord{
		ChangedFields:        []string{"version"},
		VersionID:            versionID,
		ContributorMemberIDs: canonicalAuditValues(contributorMemberIDs),
	})
}

func NewWorkVersionCreatedAuditRecord(metadata AuditMetadata, workID, versionID string, contributorMemberIDs []string) (AuditRecord, error) {
	return newCatalogAuditRecord(metadata, AuditWorkUpdated, workID, AuditRecord{
		ChangedFields:        []string{"version"},
		VersionID:            versionID,
		ContributorMemberIDs: canonicalAuditValues(contributorMemberIDs),
	})
}

func NewPostCreatedAuditRecord(metadata AuditMetadata, postID string) (AuditRecord, error) {
	return newCatalogAuditRecord(metadata, AuditPostCreated, postID, AuditRecord{})
}

func NewPageCreatedAuditRecord(metadata AuditMetadata, pageID string) (AuditRecord, error) {
	return newCatalogAuditRecord(metadata, AuditPageCreated, pageID, AuditRecord{})
}

func NewWorkCreatedAuditRecord(metadata AuditMetadata, workID string) (AuditRecord, error) {
	return newCatalogAuditRecord(metadata, AuditWorkCreated, workID, AuditRecord{})
}

func NewPostDeletedAuditRecord(metadata AuditMetadata, postID string) (AuditRecord, error) {
	return newCatalogAuditRecord(metadata, AuditPostDeleted, postID, AuditRecord{})
}

func NewPageDeletedAuditRecord(metadata AuditMetadata, pageID string) (AuditRecord, error) {
	return newCatalogAuditRecord(metadata, AuditPageDeleted, pageID, AuditRecord{})
}

func NewWorkDeletedAuditRecord(metadata AuditMetadata, workID string) (AuditRecord, error) {
	return newCatalogAuditRecord(metadata, AuditWorkDeleted, workID, AuditRecord{})
}

func NewAccountCanonicalEmailUpdatedAuditRecord(metadata AuditMetadata, memberID, previousEmail, newEmail string) (AuditRecord, error) {
	return newCatalogAuditRecord(metadata, AuditAccountUpdated, memberID, AuditRecord{ChangedFields: []string{"canonical_email"}, PreviousEmail: previousEmail, NewEmail: newEmail})
}

func NewAccountEmailLoginAddedAuditRecord(metadata AuditMetadata, memberID, email string) (AuditRecord, error) {
	return newAccountEmailLoginUpdatedAuditRecord(metadata, memberID, AuditCollectionOperationAdded, email)
}

func NewAccountEmailLoginRemovedAuditRecord(metadata AuditMetadata, memberID, email string) (AuditRecord, error) {
	return newAccountEmailLoginUpdatedAuditRecord(metadata, memberID, AuditCollectionOperationRemoved, email)
}

func newAccountEmailLoginUpdatedAuditRecord(metadata AuditMetadata, memberID string, operation AuditCollectionOperation, email string) (AuditRecord, error) {
	return newCatalogAuditRecord(metadata, AuditAccountUpdated, memberID, AuditRecord{ChangedFields: []string{"login_emails"}, CollectionOperation: operation, Email: email})
}

func NewAccountSocialLoginAddedAuditRecord(metadata AuditMetadata, memberID, provider, providerSubject string) (AuditRecord, error) {
	return newAccountSocialLoginUpdatedAuditRecord(metadata, memberID, AuditCollectionOperationAdded, provider, providerSubject)
}

func NewAccountSocialLoginRemovedAuditRecord(metadata AuditMetadata, memberID, provider, providerSubject string) (AuditRecord, error) {
	return newAccountSocialLoginUpdatedAuditRecord(metadata, memberID, AuditCollectionOperationRemoved, provider, providerSubject)
}

func newAccountSocialLoginUpdatedAuditRecord(metadata AuditMetadata, memberID string, operation AuditCollectionOperation, provider, providerSubject string) (AuditRecord, error) {
	return newCatalogAuditRecord(metadata, AuditAccountUpdated, memberID, AuditRecord{ChangedFields: []string{"social_logins"}, CollectionOperation: operation, Provider: provider, ProviderSubject: providerSubject})
}

func NewAccountPasskeyAddedAuditRecord(metadata AuditMetadata, memberID string, passkeyIDs []string) (AuditRecord, error) {
	return newAccountPasskeysUpdatedAuditRecord(metadata, memberID, AuditCollectionOperationAdded, passkeyIDs)
}

func NewAccountPasskeyRemovedAuditRecord(metadata AuditMetadata, memberID string, passkeyIDs []string) (AuditRecord, error) {
	return newAccountPasskeysUpdatedAuditRecord(metadata, memberID, AuditCollectionOperationRemoved, passkeyIDs)
}

func newAccountPasskeysUpdatedAuditRecord(metadata AuditMetadata, memberID string, operation AuditCollectionOperation, passkeyIDs []string) (AuditRecord, error) {
	return newCatalogAuditRecord(metadata, AuditAccountUpdated, memberID, AuditRecord{ChangedFields: []string{"passkeys"}, CollectionOperation: operation, PasskeyIDs: canonicalAuditValues(passkeyIDs)})
}

func NewAccountSessionRevokedAuditRecord(metadata AuditMetadata, memberID string, scope AccountSessionScope, sessionIDs []string) (AuditRecord, error) {
	return newCatalogAuditRecord(metadata, AuditAccountUpdated, memberID, AuditRecord{ChangedFields: []string{"sessions"}, CollectionOperation: AuditCollectionOperationRemoved, SessionScope: scope, SessionIDs: canonicalAuditValues(sessionIDs)})
}

func NewAccountDeletionRequestedAuditRecord(metadata AuditMetadata, memberID string, previousState AuditState) (AuditRecord, error) {
	return newAccountDeletionStateUpdatedAuditRecord(metadata, memberID, previousState, AuditStateConfirmationPending)
}

func NewAccountDeletionScheduledAuditRecord(metadata AuditMetadata, memberID string, previousState AuditState) (AuditRecord, error) {
	return newAccountDeletionStateUpdatedAuditRecord(metadata, memberID, previousState, AuditStateScheduled)
}

func NewAccountDeletionCancelledAuditRecord(metadata AuditMetadata, memberID string) (AuditRecord, error) {
	return newAccountDeletionStateUpdatedAuditRecord(metadata, memberID, AuditStateScheduled, AuditStateCancelled)
}

func NewAccountDeletionRecoveredAuditRecord(metadata AuditMetadata, memberID string) (AuditRecord, error) {
	return newAccountDeletionStateUpdatedAuditRecord(metadata, memberID, AuditStateRecoveryConfirmationPending, AuditStateRecovered)
}

func newAccountDeletionStateUpdatedAuditRecord(metadata AuditMetadata, memberID string, previousState, newState AuditState) (AuditRecord, error) {
	return newCatalogAuditRecord(metadata, AuditAccountUpdated, memberID, AuditRecord{ChangedFields: []string{"deletion_state"}, PreviousState: previousState, NewState: newState})
}

func NewAccountDeletedAuditRecord(metadata AuditMetadata, memberID string) (AuditRecord, error) {
	return newCatalogAuditRecord(metadata, AuditAccountDeleted, memberID, AuditRecord{})
}

func canonicalAuditValues(values []string) []string {
	canonical := slices.Clone(values)
	slices.Sort(canonical)
	return slices.Compact(canonical)
}

func newValidatedAuditRecord(metadata AuditMetadata, attributes AuditRecord) (AuditRecord, error) {
	attributes.AuditID = metadata.AuditID
	attributes.OccurredAt = metadata.OccurredAt
	attributes.Correlation = metadata.Correlation
	attributes.RecordActor = metadata.RecordActor
	if err := attributes.Validate(); err != nil {
		return AuditRecord{}, err
	}
	return attributes, nil
}

// newCatalogAuditRecord binds an explicit catalog action to its registered
// target type. Builders never derive targets from an action string.
func newCatalogAuditRecord(metadata AuditMetadata, action AuditAction, targetID string, attributes AuditRecord) (AuditRecord, error) {
	entry, ok := auditCatalog[action]
	if !ok {
		return AuditRecord{}, fmt.Errorf("unsupported audit action %s", action)
	}
	attributes.Action = action
	attributes.TargetType = entry.targetType
	attributes.TargetID = targetID
	return newValidatedAuditRecord(metadata, attributes)
}
