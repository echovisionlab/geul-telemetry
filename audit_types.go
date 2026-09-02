package telemetry

import (
	"fmt"
	"time"

	"github.com/google/uuid"
)

type AuditAction string
type AccountSessionScope string
type AuditCollectionOperation string
type AuditItemOperation string
type AuditState string
type AuditAssetSlot string
type AuditItemScope string
type AuditPolicyType string
type AuditRelationship string

const (
	AuditAccountDeleted             AuditAction = "account.deleted"
	AuditAccountUpdated             AuditAction = "account.updated"
	AuditArtistCreated              AuditAction = "artist.created"
	AuditArtistDeleted              AuditAction = "artist.deleted"
	AuditArtistUpdated              AuditAction = "artist.updated"
	AuditAudienceSegmentCreated     AuditAction = "audience_segment.created"
	AuditAudienceSegmentUpdated     AuditAction = "audience_segment.updated"
	AuditCampaignCreated            AuditAction = "campaign.created"
	AuditCampaignDeleted            AuditAction = "campaign.deleted"
	AuditCampaignUpdated            AuditAction = "campaign.updated"
	AuditCategoryCreated            AuditAction = "category.created"
	AuditCategoryDeleted            AuditAction = "category.deleted"
	AuditCategoryUpdated            AuditAction = "category.updated"
	AuditClientCreated              AuditAction = "client.created"
	AuditClientDeleted              AuditAction = "client.deleted"
	AuditClientUpdated              AuditAction = "client.updated"
	AuditEmailEventMappingUpdated   AuditAction = "email_event_mapping.updated"
	AuditEmailLayoutCreated         AuditAction = "email_layout.created"
	AuditEmailLayoutDeleted         AuditAction = "email_layout.deleted"
	AuditEmailLayoutUpdated         AuditAction = "email_layout.updated"
	AuditEmailSuppressionUpdated    AuditAction = "email_suppression.updated"
	AuditEmailTemplateCreated       AuditAction = "email_template.created"
	AuditEmailTemplateDeleted       AuditAction = "email_template.deleted"
	AuditEmailTemplateUpdated       AuditAction = "email_template.updated"
	AuditFileCreated                AuditAction = "file.created"
	AuditFileDeleted                AuditAction = "file.deleted"
	AuditFileUpdated                AuditAction = "file.updated"
	AuditFileFolderCreated          AuditAction = "file_folder.created"
	AuditFileFolderDeleted          AuditAction = "file_folder.deleted"
	AuditFileFolderUpdated          AuditAction = "file_folder.updated"
	AuditFormCreated                AuditAction = "form.created"
	AuditFormDeleted                AuditAction = "form.deleted"
	AuditFormUpdated                AuditAction = "form.updated"
	AuditFormSubmissionCreated      AuditAction = "form_submission.created"
	AuditFormSubmissionDeleted      AuditAction = "form_submission.deleted"
	AuditFormatCreated              AuditAction = "format.created"
	AuditFormatDeleted              AuditAction = "format.deleted"
	AuditFormatUpdated              AuditAction = "format.updated"
	AuditGenreCreated               AuditAction = "genre.created"
	AuditGenreDeleted               AuditAction = "genre.deleted"
	AuditGenreUpdated               AuditAction = "genre.updated"
	AuditLabelCreated               AuditAction = "label.created"
	AuditLabelDeleted               AuditAction = "label.deleted"
	AuditLabelUpdated               AuditAction = "label.updated"
	AuditMapThemeCreated            AuditAction = "map_theme.created"
	AuditMapThemeUpdated            AuditAction = "map_theme.updated"
	AuditMapThemeDeleted            AuditAction = "map_theme.deleted"
	AuditLegalPolicyCreated         AuditAction = "legal_policy.created"
	AuditLegalPolicyUpdated         AuditAction = "legal_policy.updated"
	AuditLegalPolicyDeleted         AuditAction = "legal_policy.deleted"
	AuditMailAdapterCreated         AuditAction = "mail_adapter.created"
	AuditMailAdapterDeleted         AuditAction = "mail_adapter.deleted"
	AuditMailAdapterUpdated         AuditAction = "mail_adapter.updated"
	AuditMapPlaceCreated            AuditAction = "map_place.created"
	AuditMapPlaceDeleted            AuditAction = "map_place.deleted"
	AuditMapPlaceUpdated            AuditAction = "map_place.updated"
	AuditMemberUpdated              AuditAction = "member.updated"
	AuditMemberTagCreated           AuditAction = "member_tag.created"
	AuditMemberTagDeleted           AuditAction = "member_tag.deleted"
	AuditMenuCreated                AuditAction = "menu.created"
	AuditMenuDeleted                AuditAction = "menu.deleted"
	AuditMenuUpdated                AuditAction = "menu.updated"
	AuditPostCreated                AuditAction = "post.created"
	AuditPostUpdated                AuditAction = "post.updated"
	AuditPostDeleted                AuditAction = "post.deleted"
	AuditPageCreated                AuditAction = "page.created"
	AuditPageUpdated                AuditAction = "page.updated"
	AuditPageDeleted                AuditAction = "page.deleted"
	AuditWorkCreated                AuditAction = "work.created"
	AuditWorkUpdated                AuditAction = "work.updated"
	AuditWorkDeleted                AuditAction = "work.deleted"
	AuditPostSeriesCreated          AuditAction = "post_series.created"
	AuditPostSeriesUpdated          AuditAction = "post_series.updated"
	AuditPostSeriesDeleted          AuditAction = "post_series.deleted"
	AuditProgramEventTypeCreated    AuditAction = "program_event_type.created"
	AuditProgramEventTypeUpdated    AuditAction = "program_event_type.updated"
	AuditProgramEventTypeDeleted    AuditAction = "program_event_type.deleted"
	AuditProgramEventCreated        AuditAction = "program_event.created"
	AuditProgramEventUpdated        AuditAction = "program_event.updated"
	AuditProgramEventDeleted        AuditAction = "program_event.deleted"
	AuditProgramEventSeriesCreated  AuditAction = "program_event_series.created"
	AuditProgramEventSeriesUpdated  AuditAction = "program_event_series.updated"
	AuditProgramEventSeriesDeleted  AuditAction = "program_event_series.deleted"
	AuditReleaseCreated             AuditAction = "release.created"
	AuditReleaseUpdated             AuditAction = "release.updated"
	AuditReleaseDeleted             AuditAction = "release.deleted"
	AuditSiteSettingsUpdated        AuditAction = "site_settings.updated"
	AuditStyleCreated               AuditAction = "style.created"
	AuditStyleDeleted               AuditAction = "style.deleted"
	AuditStyleUpdated               AuditAction = "style.updated"
	AuditTagCreated                 AuditAction = "tag.created"
	AuditTagDeleted                 AuditAction = "tag.deleted"
	AuditTagUpdated                 AuditAction = "tag.updated"
	AuditTranslationProviderCreated AuditAction = "translation_provider.created"
	AuditTranslationProviderDeleted AuditAction = "translation_provider.deleted"
	AuditTranslationProviderUpdated AuditAction = "translation_provider.updated"
	AuditTranslationSettingsUpdated AuditAction = "translation_settings.updated"

	AccountSessionScopeCurrent AccountSessionScope = "current"
	AccountSessionScopeOne     AccountSessionScope = "one"
	AccountSessionScopeOthers  AccountSessionScope = "others"

	AuditCollectionOperationAdded   AuditCollectionOperation = "added"
	AuditCollectionOperationRemoved AuditCollectionOperation = "removed"

	AuditItemOperationCreated AuditItemOperation = "created"
	AuditItemOperationUpdated AuditItemOperation = "updated"
	AuditItemOperationDeleted AuditItemOperation = "deleted"

	AuditAssetSlotLight AuditAssetSlot = "light"
	AuditAssetSlotDark  AuditAssetSlot = "dark"

	AuditItemScopeForm      AuditItemScope = "form"
	AuditItemScopeDashboard AuditItemScope = "dashboard"

	AuditPolicyTypeTerms   AuditPolicyType = "terms"
	AuditPolicyTypePrivacy AuditPolicyType = "privacy"

	AuditRelationshipNone         AuditRelationship = "none"
	AuditRelationshipAuthor       AuditRelationship = "author"
	AuditRelationshipCollaborator AuditRelationship = "collaborator"
	AuditRelationshipOwner        AuditRelationship = "owner"
	AuditRelationshipManager      AuditRelationship = "manager"

	AuditStateNone                        AuditState = "none"
	AuditStateActive                      AuditState = "active"
	AuditStateBanned                      AuditState = "banned"
	AuditStateConfirmationPending         AuditState = "confirmation_pending"
	AuditStateScheduled                   AuditState = "scheduled"
	AuditStateRecoveryConfirmationPending AuditState = "recovery_confirmation_pending"
	AuditStateCancelled                   AuditState = "cancelled"
	AuditStateRecovered                   AuditState = "recovered"
	AuditStateReleased                    AuditState = "released"
	AuditStateDraft                       AuditState = "draft"
	AuditStatePublished                   AuditState = "published"
	AuditStateArchived                    AuditState = "archived"
	AuditStateSending                     AuditState = "sending"
	AuditStateSent                        AuditState = "sent"
	AuditStateFailed                      AuditState = "failed"
	AuditStateSubscribed                  AuditState = "subscribed"
	AuditStateUnsubscribed                AuditState = "unsubscribed"
	AuditStateDisabled                    AuditState = "disabled"
	AuditStatePublic                      AuditState = "public"
	AuditStateAuthenticated               AuditState = "authenticated"
	AuditStateRestricted                  AuditState = "restricted"
)

type AuditRecord struct {
	AuditID    string      `json:"audit_id"`
	OccurredAt time.Time   `json:"occurred_at"`
	Action     AuditAction `json:"action"`
	Correlation
	RecordActor
	TargetType           string                   `json:"target_type"`
	TargetID             string                   `json:"target_id"`
	AssetID              string                   `json:"asset_id,omitempty"`
	AssetSlot            AuditAssetSlot           `json:"asset_slot,omitempty"`
	ConsentID            string                   `json:"consent_id,omitempty"`
	Nickname             string                   `json:"nickname,omitempty"`
	ChangedFields        []string                 `json:"changed_fields,omitempty"`
	CollectionOperation  AuditCollectionOperation `json:"collection_operation,omitempty"`
	ItemOperation        AuditItemOperation       `json:"item_operation,omitempty"`
	PreviousState        AuditState               `json:"previous_state,omitempty"`
	NewState             AuditState               `json:"new_state,omitempty"`
	EffectiveAt          *time.Time               `json:"effective_at,omitempty"`
	ScheduledAt          *time.Time               `json:"scheduled_at,omitempty"`
	ScheduledTimeZone    string                   `json:"scheduled_time_zone,omitempty"`
	VersionID            string                   `json:"version_id,omitempty"`
	VersionNumber        *int64                   `json:"version_number,omitempty"`
	ContributorMemberIDs []string                 `json:"contributor_member_ids,omitempty"`
	EventName            string                   `json:"event_name,omitempty"`
	FileID               string                   `json:"file_id,omitempty"`
	FileIDs              *[]string                `json:"file_ids,omitempty"`
	ItemID               string                   `json:"item_id,omitempty"`
	ItemIDs              *[]string                `json:"item_ids,omitempty"`
	ItemScope            AuditItemScope           `json:"item_scope,omitempty"`
	ParentID             string                   `json:"parent_id,omitempty"`
	PolicyType           AuditPolicyType          `json:"policy_type,omitempty"`
	PostIDs              *[]string                `json:"post_ids,omitempty"`
	PreferredLocale      string                   `json:"preferred_locale,omitempty"`
	Locale               string                   `json:"locale,omitempty"`
	PreviousLocale       string                   `json:"previous_locale,omitempty"`
	NewLocale            string                   `json:"new_locale,omitempty"`
	PreviousItemID       string                   `json:"previous_item_id,omitempty"`
	PreviousItemIDs      *[]string                `json:"previous_item_ids,omitempty"`
	PreviousParentID     string                   `json:"previous_parent_id,omitempty"`
	NewParentID          string                   `json:"new_parent_id,omitempty"`
	PreviousRelationship AuditRelationship        `json:"previous_relationship,omitempty"`
	NewRelationship      AuditRelationship        `json:"new_relationship,omitempty"`
	PreviousSeriesID     string                   `json:"previous_series_id,omitempty"`
	NewSeriesID          string                   `json:"new_series_id,omitempty"`
	SubjectMemberID      string                   `json:"subject_member_id,omitempty"`
	SubjectPostID        string                   `json:"subject_post_id,omitempty"`
	TagIDs               *[]string                `json:"tag_ids,omitempty"`
	TagName              string                   `json:"tag_name,omitempty"`
	PreviousRole         string                   `json:"previous_role,omitempty"`
	NewRole              string                   `json:"new_role,omitempty"`
	Email                string                   `json:"email,omitempty"`
	PreviousEmail        string                   `json:"previous_email,omitempty"`
	NewEmail             string                   `json:"new_email,omitempty"`
	Provider             string                   `json:"provider,omitempty"`
	ProviderSubject      string                   `json:"provider_subject,omitempty"`
	PasskeyIDs           []string                 `json:"passkey_ids,omitempty"`
	SessionScope         AccountSessionScope      `json:"session_scope,omitempty"`
	SessionIDs           []string                 `json:"session_ids,omitempty"`
}

func (record AuditRecord) Validate() error {
	parsedAuditID, err := uuid.Parse(record.AuditID)
	if err != nil || parsedAuditID.String() != record.AuditID || parsedAuditID.Version() != 4 || parsedAuditID.Variant() != uuid.RFC4122 {
		return fmt.Errorf("audit_id must be a canonical UUIDv4")
	}
	if err := validateRecordTime(record.OccurredAt); err != nil {
		return err
	}
	if err := record.RecordActor.Validate(); err != nil {
		return err
	}
	if record.Kind == ActorKindAnonymous && record.Action != AuditFormSubmissionCreated {
		return fmt.Errorf("domain audit actor cannot be anonymous")
	}
	if record.RequestID != "" {
		if err := ValidateRequestID(record.RequestID); err != nil {
			return err
		}
	}
	if err := validateTraceCorrelation(record.TraceID, record.SpanID); err != nil {
		return err
	}
	return validateAuditCatalogRecord(record)
}
