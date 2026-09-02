package telemetry

import (
	"testing"
	"time"
)

// These cases exercise the public builder surface by domain.  A successful
// build is also a contract check: every builder must emit a record accepted by
// the complete catalog and its action-specific validator.
func TestDomainAuditBuildersEmitValidRecords(t *testing.T) {
	m := AuditMetadata{AuditID: "00000000-0000-4000-8000-000000000001", OccurredAt: testOccurredAt, RecordActor: RecordActor{Kind: ActorKindMember, MemberID: "member-1"}}
	timezone := "Asia/Seoul"
	cases := []struct {
		name  string
		build func() (AuditRecord, error)
	}{
		{"artist created", func() (AuditRecord, error) { return NewArtistCreatedAuditRecord(m, "artist-1") }},
		{"artist deleted", func() (AuditRecord, error) { return NewArtistDeletedAuditRecord(m, "artist-1") }},
		{"artist lifecycle", func() (AuditRecord, error) {
			return NewArtistLifecycleAuditRecord(m, "artist-1", AuditStateDraft, AuditStatePublished)
		}},
		{"artist gallery", func() (AuditRecord, error) { return NewArtistGalleryAuditRecord(m, "artist-1", []string{"file-1"}) }},
		{"artist participant", func() (AuditRecord, error) {
			return NewArtistParticipantAuditRecord(m, "artist-1", "member-2", AuditRelationshipNone, AuditRelationshipOwner)
		}},
		{"artist share link", func() (AuditRecord, error) {
			return NewArtistShareLinkAuditRecord(m, "artist-1", "link-1", AuditItemOperationCreated)
		}},
		{"label created", func() (AuditRecord, error) { return NewLabelCreatedAuditRecord(m, "label-1") }},
		{"label deleted", func() (AuditRecord, error) { return NewLabelDeletedAuditRecord(m, "label-1") }},
		{"label lifecycle", func() (AuditRecord, error) {
			return NewLabelLifecycleAuditRecord(m, "label-1", AuditStateDraft, AuditStatePublished)
		}},
		{"label participant", func() (AuditRecord, error) {
			return NewLabelParticipantAuditRecord(m, "label-1", "member-2", AuditRelationshipNone, AuditRelationshipManager)
		}},
		{"label logo", func() (AuditRecord, error) {
			return NewLabelLogoAuditRecord(m, "label-1", AuditAssetSlotLight, AuditCollectionOperationAdded, "asset-1")
		}},
		{"label share link", func() (AuditRecord, error) {
			return NewLabelShareLinkAuditRecord(m, "label-1", "link-1", AuditItemOperationDeleted)
		}},

		{"template created", func() (AuditRecord, error) { return NewEmailTemplateCreatedAuditRecord(m, "template-1") }},
		{"template deleted", func() (AuditRecord, error) { return NewEmailTemplateDeletedAuditRecord(m, "template-1") }},
		{"template metadata", func() (AuditRecord, error) {
			return NewEmailTemplateMetadataAuditRecord(m, "template-1", []string{"name"})
		}},
		{"layout created", func() (AuditRecord, error) { return NewEmailLayoutCreatedAuditRecord(m, "layout-1") }},
		{"layout deleted", func() (AuditRecord, error) { return NewEmailLayoutDeletedAuditRecord(m, "layout-1") }},
		{"layout metadata", func() (AuditRecord, error) { return NewEmailLayoutMetadataAuditRecord(m, "layout-1", []string{"key"}) }},

		{"form created", func() (AuditRecord, error) { return NewFormCreatedAuditRecord(m, "form-1") }},
		{"form deleted", func() (AuditRecord, error) { return NewFormDeletedAuditRecord(m, "form-1") }},
		{"form settings", func() (AuditRecord, error) { return NewFormSettingsAuditRecord(m, "form-1", []string{"auth_required"}) }},
		{"form lifecycle", func() (AuditRecord, error) {
			return NewFormLifecycleAuditRecord(m, "form-1", AuditStateDraft, AuditStatePublished)
		}},
		{"form featured image", func() (AuditRecord, error) {
			return NewFormFeaturedImageAuditRecord(m, "form-1", "file-1", AuditCollectionOperationAdded)
		}},
		{"form submission deleted", func() (AuditRecord, error) { return NewFormSubmissionDeletedAuditRecord(m, "submission-1") }},

		{"post configuration", func() (AuditRecord, error) { return NewPostConfigurationAuditRecord(m, "post-1", []string{"slug"}) }},
		{"post status", func() (AuditRecord, error) {
			return NewPostStatusLifecycleAuditRecord(m, "post-1", AuditStateDraft, AuditStatePublished)
		}},
		{"post schedule", func() (AuditRecord, error) {
			return NewPostScheduleLifecycleAuditRecord(m, "post-1", AuditStateDraft, AuditStateScheduled, testOccurredAt, timezone)
		}},
		{"post featured image", func() (AuditRecord, error) {
			return NewPostFeaturedImageAuditRecord(m, "post-1", "asset-1", AuditCollectionOperationAdded)
		}},
		{"post share link", func() (AuditRecord, error) {
			return NewPostShareLinkAuditRecord(m, "post-1", "link-1", AuditItemOperationCreated)
		}},
		{"post version restore", func() (AuditRecord, error) { return NewPostVersionRestoreAuditRecord(m, "post-1", "version-1") }},
		{"post comment", func() (AuditRecord, error) {
			return NewPostCommentAuditRecord(m, "post-1", "comment-1", AuditItemOperationUpdated)
		}},

		{"event created", func() (AuditRecord, error) { return NewProgramEventCreatedAuditRecord(m, "event-1") }},
		{"event deleted", func() (AuditRecord, error) { return NewProgramEventDeletedAuditRecord(m, "event-1") }},
		{"event metadata", func() (AuditRecord, error) { return NewProgramEventMetadataAuditRecord(m, "event-1", []string{"slug"}) }},
		{"event poster", func() (AuditRecord, error) {
			return NewProgramEventPosterAuditRecord(m, "event-1", "file-1", AuditCollectionOperationAdded)
		}},
		{"event child", func() (AuditRecord, error) {
			return NewProgramEventChildAuditRecord(m, "event-1", "media", "media-1", AuditItemOperationCreated)
		}},
		{"event child order", func() (AuditRecord, error) {
			return NewProgramEventChildOrderAuditRecord(m, "event-1", "credits", []string{"credit-1"})
		}},
		{"event lifecycle", func() (AuditRecord, error) {
			return NewProgramEventLifecycleAuditRecord(m, "event-1", AuditStateDraft, AuditStatePublished)
		}},
		{"event series created", func() (AuditRecord, error) { return NewProgramEventSeriesCreatedAuditRecord(m, "series-1") }},
		{"event series deleted", func() (AuditRecord, error) { return NewProgramEventSeriesDeletedAuditRecord(m, "series-1") }},
		{"event series metadata", func() (AuditRecord, error) {
			return NewProgramEventSeriesMetadataAuditRecord(m, "series-1", []string{"title"})
		}},
		{"event series poster", func() (AuditRecord, error) {
			return NewProgramEventSeriesPosterAuditRecord(m, "series-1", "file-1", AuditCollectionOperationRemoved)
		}},
		{"event series lifecycle", func() (AuditRecord, error) {
			return NewProgramEventSeriesLifecycleAuditRecord(m, "series-1", AuditStateDraft, AuditStatePublished)
		}},

		{"release created", func() (AuditRecord, error) { return NewReleaseCreatedAuditRecord(m, "release-1") }},
		{"release deleted", func() (AuditRecord, error) { return NewReleaseDeletedAuditRecord(m, "release-1") }},
		{"release metadata", func() (AuditRecord, error) { return NewReleaseMetadataAuditRecord(m, "release-1", []string{"slug"}) }},
		{"release track", func() (AuditRecord, error) {
			return NewReleaseTrackAuditRecord(m, "release-1", "track-1", AuditItemOperationCreated)
		}},
		{"release artwork", func() (AuditRecord, error) {
			return NewReleaseArtworkAuditRecord(m, "release-1", "file-1", AuditCollectionOperationAdded)
		}},
		{"release lifecycle", func() (AuditRecord, error) {
			return NewReleaseLifecycleAuditRecord(m, "release-1", AuditStateDraft, AuditStatePublished)
		}},
		{"release share link", func() (AuditRecord, error) {
			return NewReleaseShareLinkAuditRecord(m, "release-1", "link-1", AuditItemOperationCreated)
		}},
		{"campaign created", func() (AuditRecord, error) { return NewCampaignCreatedAuditRecord(m, "campaign-1") }},
		{"campaign deleted", func() (AuditRecord, error) { return NewCampaignDeletedAuditRecord(m, "campaign-1") }},
		{"campaign layout", func() (AuditRecord, error) {
			return NewCampaignTargetLayoutAuditRecord(m, "campaign-1", []string{"layout"})
		}},
		{"campaign metadata", func() (AuditRecord, error) {
			return NewCampaignMetadataAuditRecord(m, "campaign-1", []string{"name"})
		}},

		{"post series created", func() (AuditRecord, error) { return NewPostSeriesCreatedAuditRecord(m, "series-1") }},
		{"post series deleted", func() (AuditRecord, error) { return NewPostSeriesDeletedAuditRecord(m, "series-1") }},
		{"post series metadata", func() (AuditRecord, error) {
			return NewPostSeriesSourceMetadataAuditRecord(m, "series-1", []string{"slug"})
		}},
		{"post series lifecycle", func() (AuditRecord, error) {
			return NewPostSeriesLifecycleAuditRecord(m, "series-1", AuditStateDraft, AuditStatePublished)
		}},
		{"post series manager", func() (AuditRecord, error) {
			return NewPostSeriesManagerAuditRecord(m, "series-1", "member-1", AuditRelationshipNone, AuditRelationshipManager)
		}},
		{"post series membership", func() (AuditRecord, error) {
			return NewPostSeriesMembershipAuditRecord(m, "series-1", "post-1", "", "other-series")
		}},
		{"post series order", func() (AuditRecord, error) { return NewPostSeriesOrderAuditRecord(m, "series-1", []string{"post-1"}) }},
		{"post series image", func() (AuditRecord, error) {
			return NewPostSeriesFeaturedImageAuditRecord(m, "series-1", AuditCollectionOperationAdded, "file-1")
		}},
		{"event type created", func() (AuditRecord, error) { return NewProgramEventTypeCreatedAuditRecord(m, "type-1") }},
		{"event type deleted", func() (AuditRecord, error) { return NewProgramEventTypeDeletedAuditRecord(m, "type-1") }},
		{"event type config", func() (AuditRecord, error) {
			return NewProgramEventTypeConfigUpdatedAuditRecord(m, "type-1", []string{"slug"})
		}},

		{"client metadata", func() (AuditRecord, error) {
			return NewClientMetadataUpdatedAuditRecord(m, "client-1", []string{"name"})
		}},
		{"place metadata", func() (AuditRecord, error) {
			return NewMapPlaceMetadataUpdatedAuditRecord(m, "place-1", []string{"name"})
		}},
		{"audience config", func() (AuditRecord, error) {
			return NewAudienceSegmentConfigUpdatedAuditRecord(m, "segment-1", []string{"name"})
		}},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			record, err := tc.build()
			if err != nil {
				t.Fatal(err)
			}
			if err := record.Validate(); err != nil {
				t.Fatalf("builder output invalid: %v", err)
			}
		})
	}
}

func TestAuditVariantRejectsInvalidSemanticShapes(t *testing.T) {
	identifier := "item-1"
	orderedID := []string{identifier}
	invalidID := []string{""}
	validTransition := AuditRecord{ChangedFields: []string{"status"}, PreviousState: AuditStateDraft, NewState: AuditStatePublished}
	cases := []struct {
		name     string
		validate func() error
	}{
		{"artist needs one variant", func() error { return validateArtistUpdate(AuditRecord{}) }},
		{"artist rejects variant", func() error { return validateArtistUpdate(AuditRecord{ChangedFields: []string{"unknown"}}) }},
		{"artist gallery needs files", func() error { return validateArtistUpdate(AuditRecord{ChangedFields: []string{"gallery"}}) }},
		{"artist gallery rejects invalid file", func() error {
			return validateArtistUpdate(AuditRecord{ChangedFields: []string{"gallery"}, FileIDs: &invalidID})
		}},
		{"artist gallery rejects duplicate files", func() error {
			duplicateIDs := []string{"file-1", "file-1"}
			return validateArtistUpdate(AuditRecord{ChangedFields: []string{"gallery"}, FileIDs: &duplicateIDs})
		}},
		{"artist gallery rejects extras", func() error {
			return validateArtistUpdate(AuditRecord{ChangedFields: []string{"gallery"}, FileIDs: &orderedID, AssetID: "asset-1"})
		}},
		{"artist lifecycle rejects state", func() error {
			return validateDraftPublished(AuditRecord{ChangedFields: []string{"status"}, PreviousState: AuditStateActive, NewState: AuditStatePublished})
		}},
		{"artist lifecycle rejects extras", func() error { r := validTransition; r.AssetID = "asset-1"; return validateDraftPublished(r) }},
		{"artist participant needs subject", func() error { return validateParticipant(AuditRecord{}, AuditRelationshipOwner) }},
		{"artist participant rejects relationship", func() error {
			return validateParticipant(AuditRecord{SubjectMemberID: identifier, PreviousRelationship: AuditRelationshipNone, NewRelationship: AuditRelationshipAuthor}, AuditRelationshipOwner)
		}},
		{"artist participant rejects extras", func() error {
			return validateParticipant(AuditRecord{SubjectMemberID: identifier, PreviousRelationship: AuditRelationshipNone, NewRelationship: AuditRelationshipOwner, AssetID: "asset-1"}, AuditRelationshipOwner)
		}},
		{"ordered identifiers reject invalid", func() error { return validateOrderedIdentifiers("items", invalidID) }},
		{"label needs one variant", func() error { return validateLabelUpdate(AuditRecord{}) }},
		{"label rejects variant", func() error { return validateLabelUpdate(AuditRecord{ChangedFields: []string{"unknown"}}) }},
		{"label logo needs slot", func() error { return validateLabelUpdate(AuditRecord{ChangedFields: []string{"logo"}}) }},
		{"label logo needs binding operation", func() error {
			return validateLabelUpdate(AuditRecord{ChangedFields: []string{"logo"}, AssetSlot: AuditAssetSlotLight, AssetID: identifier})
		}},
		{"label logo rejects extras", func() error {
			return validateLabelUpdate(AuditRecord{ChangedFields: []string{"logo"}, AssetSlot: AuditAssetSlotLight, CollectionOperation: AuditCollectionOperationAdded, AssetID: identifier, FileID: "file-1"})
		}},
		{"template mapping rejects extras", func() error {
			return validateEmailMappingUpdate(AuditRecord{ChangedFields: []string{"template"}, EventName: "welcome", ItemID: identifier, AssetID: "asset-1"})
		}},
		{"template relation rejects invalid next", func() error { return validateEmailRelation(AuditRecord{ItemID: " invalid"}, false) }},
		{"template relation rejects invalid previous", func() error {
			return validateEmailRelation(AuditRecord{ItemID: identifier, PreviousItemID: " invalid"}, false)
		}},
		{"template relation rejects extras", func() error {
			return validateEmailRelation(AuditRecord{ItemID: identifier, PreviousItemID: "old-1", AssetID: "asset-1"}, false)
		}},
		{"submission rejects PII", func() error {
			return validateSubmissionCreated(AuditRecord{ParentID: "form-1", Email: "person@example.test"})
		}},
		{"form share link rejects extras", func() error {
			return validateFormUpdate(AuditRecord{ChangedFields: []string{"share_links"}, ItemID: identifier, ItemScope: AuditItemScopeForm, ItemOperation: AuditItemOperationCreated, AssetID: "asset-1"})
		}},
		{"no attributes except changed rejects typed values", func() error { return validateNoAuditAttributesExceptChanged(AuditRecord{AssetID: "asset-1"}) }},
		{"suppression requires status", func() error { return validateEmailSuppressionUpdate(AuditRecord{}) }},
		{"suppression requires release transition", func() error { return validateEmailSuppressionUpdate(AuditRecord{ChangedFields: []string{"status"}}) }},
		{"suppression rejects email PII", func() error {
			return validateEmailSuppressionUpdate(AuditRecord{ChangedFields: []string{"status"}, PreviousState: AuditStateActive, NewState: AuditStateReleased, Email: "person@example.test"})
		}},
		{"suppression rejects every other typed extra", func() error {
			return validateEmailSuppressionUpdate(AuditRecord{ChangedFields: []string{"status"}, PreviousState: AuditStateActive, NewState: AuditStateReleased, AssetID: "asset-1"})
		}},
		{"member onboarding requires self actor", func() error {
			return validateMemberOnboarded(AuditRecord{RecordActor: RecordActor{Kind: ActorKindMember, MemberID: "other"}, TargetID: "member-1", Nickname: "Name"})
		}},
		{"member avatar requires asset", func() error { return validateMemberAvatar(AuditRecord{}) }},
		{"member tags require list", func() error { return validateMemberTags(AuditRecord{}) }},
		{"member tags reject unsorted identifiers", func() error { return validateMemberTags(AuditRecord{TagIDs: &invalidID}) }},
		{"member profile needs fields", func() error {
			return validateMemberProfileOrPreference(AuditRecord{}, func(...string) error { return nil })
		}},
		{"member profile requires trimmed nickname", func() error {
			return validateMemberProfileOrPreference(AuditRecord{ChangedFields: []string{"nickname"}, Nickname: " Name "}, func(...string) error { return nil })
		}},
		{"member profile rejects nickname outside fields", func() error {
			return validateMemberProfileOrPreference(AuditRecord{ChangedFields: []string{"bio"}, Nickname: "Name"}, func(...string) error { return nil })
		}},
		{"member preference requires locale", func() error {
			return validateMemberProfileOrPreference(AuditRecord{ChangedFields: []string{"preferred_locale"}}, func(...string) error { return nil })
		}},
		{"member preference rejects consent outside fields", func() error {
			return validateMemberProfileOrPreference(AuditRecord{ChangedFields: []string{"preferred_locale"}, PreferredLocale: "ko", ConsentID: identifier}, func(...string) error { return nil })
		}},
		{"member preference rejects locale outside fields", func() error {
			return validateMemberProfileOrPreference(AuditRecord{ChangedFields: []string{"cookie_consent"}, PreferredLocale: "ko", ConsentID: identifier}, func(...string) error { return nil })
		}},
		{"member preference needs consent ID", func() error {
			return validateMemberProfileOrPreference(AuditRecord{ChangedFields: []string{"cookie_consent"}}, func(...string) error { return nil })
		}},
		{"member rejects mixed variants", func() error {
			return validateMemberProfileOrPreference(AuditRecord{ChangedFields: []string{"bio", "preferred_locale"}}, func(...string) error { return nil })
		}},
		{"member tag requires name", func() error { return validateMemberTagRecord(AuditRecord{}) }},
		{"member tag rejects extras", func() error { return validateMemberTagRecord(AuditRecord{TagName: "tag", AssetID: "asset-1"}) }},
		{"page featured image requires asset", func() error { return validatePageUpdate(AuditRecord{ChangedFields: []string{"featured_image"}}) }},
		{"page featured image rejects extras", func() error {
			return validatePageUpdate(AuditRecord{ChangedFields: []string{"featured_image"}, CollectionOperation: AuditCollectionOperationAdded, AssetID: identifier, FileID: "file-1"})
		}},
		{"page version needs contributors", func() error { return validatePageVersion(AuditRecord{VersionID: identifier}) }},
		{"page version rejects extras", func() error {
			return validatePageVersion(AuditRecord{VersionID: identifier, ContributorMemberIDs: []string{"1b6bcad2-c90d-49e9-bec7-f9a4ba6b2894"}, AssetID: "asset-1"})
		}},
		{"page version rejects invalid contributor", func() error {
			return validatePageVersion(AuditRecord{VersionID: identifier, ContributorMemberIDs: []string{"invalid"}})
		}},
		{"page restore rejects extras", func() error {
			return validatePageUpdate(AuditRecord{ChangedFields: []string{"version_restore"}, VersionID: identifier, AssetID: "asset-1"})
		}},
		{"post image requires asset", func() error {
			return validatePostConfigurationOrLifecycle(AuditRecord{ChangedFields: []string{"featured_image"}})
		}},
		{"post image rejects extras", func() error {
			return validatePostConfigurationOrLifecycle(AuditRecord{ChangedFields: []string{"featured_image"}, CollectionOperation: AuditCollectionOperationAdded, AssetID: identifier, FileID: "file-1"})
		}},
		{"post restore requires version", func() error {
			return validatePostConfigurationOrLifecycle(AuditRecord{ChangedFields: []string{"version_restore"}})
		}},
		{"post comment requires member", func() error {
			return validatePostConfigurationOrLifecycle(AuditRecord{ChangedFields: []string{"comments"}, ItemID: identifier, ItemOperation: AuditItemOperationCreated})
		}},
		{"post comment rejects extras", func() error {
			return validatePostConfigurationOrLifecycle(AuditRecord{ChangedFields: []string{"comments"}, RecordActor: RecordActor{Kind: ActorKindMember}, ItemID: identifier, ItemOperation: AuditItemOperationCreated, AssetID: "asset-1"})
		}},
		{"post lifecycle requires transition", func() error {
			return validatePostConfigurationOrLifecycle(AuditRecord{ChangedFields: []string{"status"}})
		}},
		{"post schedule requires time", func() error {
			return validatePostConfigurationOrLifecycle(AuditRecord{ChangedFields: []string{"schedule"}, PreviousState: AuditStateDraft, NewState: AuditStateScheduled})
		}},
		{"post status rejects schedule time", func() error {
			at := testOccurredAt
			return validatePostConfigurationOrLifecycle(AuditRecord{ChangedFields: []string{"status"}, PreviousState: AuditStateDraft, NewState: AuditStatePublished, ScheduledAt: &at})
		}},
		{"post lifecycle rejects extras", func() error {
			return validatePostConfigurationOrLifecycle(AuditRecord{ChangedFields: []string{"status"}, PreviousState: AuditStateDraft, NewState: AuditStatePublished, AssetID: "asset-1"})
		}},
		{"program needs one variant", func() error { return validateProgramEventUpdate(AuditRecord{}) }},
		{"program child rejects bad ordered ID", func() error { return validateProgramChild(AuditRecord{ItemIDs: &invalidID}) }},
		{"program child reorder rejects extras", func() error { return validateProgramChild(AuditRecord{ItemIDs: &orderedID, AssetID: "asset-1"}) }},
		{"program child requires item", func() error { return validateProgramChild(AuditRecord{}) }},
		{"program child rejects extras", func() error {
			return validateProgramChild(AuditRecord{ItemID: identifier, ItemOperation: AuditItemOperationCreated, AssetID: "asset-1"})
		}},
		{"program lifecycle rejects state", func() error {
			return validateProgramLifecycle(AuditRecord{PreviousState: AuditStateActive, NewState: AuditStatePublished})
		}},
		{"program lifecycle rejects extras", func() error {
			return validateProgramLifecycle(AuditRecord{PreviousState: AuditStateDraft, NewState: AuditStatePublished, AssetID: "asset-1"})
		}},
		{"release needs one variant", func() error { return validateReleaseUpdate(AuditRecord{}) }},
		{"release track audio requires values", func() error { return validateReleaseUpdate(AuditRecord{ChangedFields: []string{"track_audio"}}) }},
		{"release track audio rejects extras", func() error {
			return validateReleaseUpdate(AuditRecord{ChangedFields: []string{"track_audio"}, ItemID: identifier, FileID: "file-1", CollectionOperation: AuditCollectionOperationAdded, AssetID: "asset-1"})
		}},
		{"campaign lifecycle requires transition", func() error { return validateCampaignUpdate(AuditRecord{ChangedFields: []string{"status"}}) }},
		{"campaign schedule requires time", func() error {
			return validateCampaignUpdate(AuditRecord{ChangedFields: []string{"schedule"}, PreviousState: AuditStateDraft, NewState: AuditStateScheduled})
		}},
		{"campaign rejects invalid changed field", func() error { return validateCampaignUpdate(AuditRecord{ChangedFields: []string{"unknown"}}) }},
		{"campaign status rejects schedule values", func() error {
			at := testOccurredAt
			return validateCampaignUpdate(AuditRecord{ChangedFields: []string{"status"}, PreviousState: AuditStateDraft, NewState: AuditStateScheduled, ScheduledAt: &at})
		}},
		{"campaign delivery run needs item", func() error {
			return validateCampaignUpdate(AuditRecord{ChangedFields: []string{"delivery_run"}, PreviousState: AuditStateDraft, NewState: AuditStateScheduled})
		}},
		{"campaign lifecycle rejects extras", func() error {
			return validateCampaignUpdate(AuditRecord{ChangedFields: []string{"status"}, PreviousState: AuditStateDraft, NewState: AuditStateScheduled, AssetID: "asset-1"})
		}},
		{"series needs variant", func() error { return validatePostSeriesUpdate(AuditRecord{}) }},
		{"series rejects unsorted fields", func() error { return validatePostSeriesUpdate(AuditRecord{ChangedFields: []string{"slug", "slug"}}) }},
		{"series membership needs subject", func() error { return validatePostSeriesUpdate(AuditRecord{ChangedFields: []string{"posts"}}) }},
		{"series membership rejects extras", func() error {
			return validatePostSeriesUpdate(AuditRecord{ChangedFields: []string{"posts"}, SubjectPostID: identifier, NewSeriesID: "series-2", AssetID: "asset-1"})
		}},
		{"series order needs posts", func() error { return validatePostSeriesUpdate(AuditRecord{ChangedFields: []string{"post_order"}}) }},
		{"series order rejects extras", func() error {
			return validatePostSeriesUpdate(AuditRecord{ChangedFields: []string{"post_order"}, PostIDs: &orderedID, AssetID: "asset-1"})
		}},
		{"series order rejects invalid post", func() error {
			return validatePostSeriesUpdate(AuditRecord{ChangedFields: []string{"post_order"}, PostIDs: &invalidID})
		}},
		{"work version needs contributors", func() error { return validateWorkVersion(AuditRecord{VersionID: identifier}) }},
		{"work version rejects extras", func() error {
			return validateWorkVersion(AuditRecord{VersionID: identifier, ContributorMemberIDs: []string{"1b6bcad2-c90d-49e9-bec7-f9a4ba6b2894"}, AssetID: "asset-1"})
		}},
		{"work version rejects invalid contributor", func() error {
			return validateWorkVersion(AuditRecord{VersionID: identifier, ContributorMemberIDs: []string{"invalid"}})
		}},
		{"work lifecycle rejects state", func() error {
			return validateWorkLifecycle(AuditRecord{PreviousState: AuditStateActive, NewState: AuditStatePublished})
		}},
		{"work lifecycle rejects extras", func() error {
			return validateWorkLifecycle(AuditRecord{PreviousState: AuditStateDraft, NewState: AuditStatePublished, AssetID: "asset-1"})
		}},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			if err := tc.validate(); err == nil {
				t.Fatal("invalid audit record was accepted")
			}
		})
	}
}

func TestAuditCatalogRejectsIncompleteAndMalformedRecords(t *testing.T) {
	m := AuditMetadata{AuditID: "00000000-0000-4000-8000-000000000001", OccurredAt: testOccurredAt, RecordActor: RecordActor{Kind: ActorKindMember, MemberID: "member-1"}}
	if _, err := newCatalogAuditRecord(m, AuditAction("unknown.created"), "target-1", AuditRecord{}); err == nil {
		t.Fatal("unknown catalog action accepted")
	}
	base := AuditRecord{Action: AuditArtistCreated, TargetType: "artist", TargetID: "artist-1", RecordActor: RecordActor{Kind: ActorKindMember, MemberID: "member-1"}}
	cases := []AuditRecord{
		{Action: AuditArtistCreated, TargetType: "wrong", TargetID: "artist-1", RecordActor: base.RecordActor},
		{Action: AuditSiteSettingsUpdated, TargetType: "site_settings", TargetID: "not-singleton", RecordActor: base.RecordActor},
		{Action: AuditArtistCreated, TargetType: "artist", TargetID: "artist-1", RecordActor: RecordActor{Kind: ActorKindAnonymous}},
		{Action: AuditArtistCreated, TargetType: "artist", TargetID: "artist-1", RecordActor: RecordActor{Kind: ActorKindSystem, Service: "other"}},
	}
	for _, record := range cases {
		if err := validateAuditCatalogRecord(record); err == nil {
			t.Fatal("catalog accepted malformed record")
		}
	}
	entry := auditCatalog[AuditArtistCreated]
	auditCatalog[AuditArtistCreated] = auditCatalogEntry{targetType: "artist"}
	defer func() { auditCatalog[AuditArtistCreated] = entry }()
	if err := validateAuditCatalogRecord(base); err == nil {
		t.Fatal("catalog accepted action without a variant validator")
	}
}

func TestExtendedAuditAttributesValidateEveryOptionalSet(t *testing.T) {
	valid := []string{"item-1"}
	invalid := []string{" invalid"}
	cases := []AuditRecord{
		{ChangedFields: []string{"z", "a"}},
		{ChangedFields: []string{"field"}, FileIDs: &invalid},
		{ChangedFields: []string{"field"}, ItemIDs: &invalid},
		{ChangedFields: []string{"field"}, PreviousItemIDs: &invalid},
		{ChangedFields: []string{"field"}, PostIDs: &invalid},
		{ChangedFields: []string{"field"}, TagIDs: &invalid},
	}
	for _, record := range cases {
		if err := validateExtendedAuditAttributes(record); err == nil {
			t.Fatal("invalid extended audit attributes accepted")
		}
	}
	if err := validateExtendedAuditAttributes(AuditRecord{ChangedFields: []string{"field"}, FileIDs: &valid, ItemIDs: &valid, PreviousItemIDs: &valid, PostIDs: &valid, TagIDs: &valid}); err != nil {
		t.Fatal(err)
	}
}

func TestSettingsReferenceValidatorsRejectInvalidBoundaries(t *testing.T) {
	validIDs := []string{"item-1"}
	version := int64(1)
	cases := []struct {
		name string
		call func() error
	}{
		{"legal identity requires values", func() error { return validateLegalIdentityOnly(AuditRecord{}) }},
		{"legal update requires identity", func() error { return validateLegalPolicyUpdate(AuditRecord{}) }},
		{"map theme content requires member", func() error {
			return auditCatalog[AuditMapThemeUpdated].validate(AuditRecord{RecordActor: RecordActor{Kind: ActorKindSystem}})
		}},
		{"legal identity rejects extras", func() error {
			return validateLegalIdentityOnly(AuditRecord{PolicyType: AuditPolicyTypeTerms, VersionNumber: &version, AssetID: "asset-1"})
		}},
		{"legal share requires item", func() error {
			return validateLegalPolicyUpdate(AuditRecord{PolicyType: AuditPolicyTypeTerms, VersionNumber: &version, ChangedFields: []string{"share_links"}})
		}},
		{"legal share rejects extras", func() error {
			return validateLegalPolicyUpdate(AuditRecord{PolicyType: AuditPolicyTypeTerms, VersionNumber: &version, ChangedFields: []string{"share_links"}, ItemID: "link-1", ItemOperation: AuditItemOperationCreated, AssetID: "asset-1"})
		}},
		{"legal lifecycle needs effective time", func() error {
			return validateLegalPolicyUpdate(AuditRecord{PolicyType: AuditPolicyTypeTerms, VersionNumber: &version, ChangedFields: []string{"effective_at"}, PreviousState: AuditStateDraft, NewState: AuditStateScheduled})
		}},
		{"legal lifecycle requires changed field", func() error {
			return validateLegalPolicyUpdate(AuditRecord{PolicyType: AuditPolicyTypeTerms, VersionNumber: &version})
		}},
		{"legal lifecycle validates effective time", func() error {
			var invalidTime time.Time
			return validateLegalPolicyUpdate(AuditRecord{PolicyType: AuditPolicyTypeTerms, VersionNumber: &version, ChangedFields: []string{"effective_at"}, PreviousState: AuditStateDraft, NewState: AuditStateScheduled, EffectiveAt: &invalidTime})
		}},
		{"legal lifecycle rejects transition", func() error {
			return validateLegalPolicyUpdate(AuditRecord{PolicyType: AuditPolicyTypeTerms, VersionNumber: &version, ChangedFields: []string{"status"}, PreviousState: AuditStateActive, NewState: AuditStateDraft})
		}},
		{"legal lifecycle rejects extras", func() error {
			return validateLegalPolicyUpdate(AuditRecord{PolicyType: AuditPolicyTypeTerms, VersionNumber: &version, ChangedFields: []string{"status"}, PreviousState: AuditStateDraft, NewState: AuditStateScheduled, AssetID: "asset-1"})
		}},
		{"file move needs distinct parents", func() error { return validateFileUpdate(AuditRecord{ChangedFields: []string{"folder_id"}}) }},
		{"file rejects changed field", func() error { return validateFileUpdate(AuditRecord{ChangedFields: []string{"unknown"}}) }},
		{"file parent needs field", func() error {
			return validateFileUpdate(AuditRecord{ChangedFields: []string{"file_name"}, PreviousParentID: "old"})
		}},
		{"file optional segments need field", func() error {
			return validateFileUpdate(AuditRecord{ChangedFields: []string{"file_name"}, ItemIDs: &validIDs})
		}},
		{"file audience states need field", func() error {
			return validateFileUpdate(AuditRecord{ChangedFields: []string{"file_name"}, PreviousState: AuditStatePublic, NewState: AuditStateRestricted})
		}},
		{"file rejects extras", func() error {
			return validateFileUpdate(AuditRecord{ChangedFields: []string{"file_name"}, AssetID: "asset-1"})
		}},
		{"folder parent needs distinction", func() error { return validateFolderUpdate(AuditRecord{ChangedFields: []string{"parent_id"}}) }},
		{"folder rejects changed field", func() error { return validateFolderUpdate(AuditRecord{ChangedFields: []string{"unknown"}}) }},
		{"folder parent needs field", func() error {
			return validateFolderUpdate(AuditRecord{ChangedFields: []string{"name"}, NewParentID: "parent-1"})
		}},
		{"folder rejects extras", func() error {
			return validateFolderUpdate(AuditRecord{ChangedFields: []string{"name"}, AssetID: "asset-1"})
		}},
		{"audience rejects transition", func() error {
			return validateAudienceUpdate(AuditRecord{ChangedFields: []string{"status"}, PreviousState: AuditStateActive, NewState: AuditStatePublished})
		}},
		{"audience rejects extras", func() error {
			return validateAudienceUpdate(AuditRecord{ChangedFields: []string{"status"}, PreviousState: AuditStateActive, NewState: AuditStateArchived, AssetID: "asset-1"})
		}},
		{"file binding needs operation", func() error { return validateFileBinding(AuditRecord{}, "image", false) }},
		{"file binding needs file", func() error {
			return validateFileBinding(AuditRecord{CollectionOperation: AuditCollectionOperationAdded}, "image", false)
		}},
		{"file binding needs slot", func() error {
			return validateFileBinding(AuditRecord{CollectionOperation: AuditCollectionOperationAdded, FileID: "file-1"}, "logo", true)
		}},
		{"file binding rejects extras", func() error {
			return validateFileBinding(AuditRecord{CollectionOperation: AuditCollectionOperationAdded, FileID: "file-1", AssetID: "asset-1"}, "image", false)
		}},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			if err := tc.call(); err == nil {
				t.Fatal("invalid settings/reference record accepted")
			}
		})
	}
}
