package telemetry

import (
	"encoding/json"
	"os"
	"testing"
	"time"
)

func TestGoSemanticBuildersMatchEveryVariantManifestCase(t *testing.T) {
	contents, err := os.ReadFile("fixtures/domain-audit-wire-parity.json")
	if err != nil {
		t.Fatal(err)
	}
	var manifest []auditVariantManifestEntry
	if err := json.Unmarshal(contents, &manifest); err != nil {
		t.Fatal(err)
	}
	m := AuditMetadata{AuditID: "00000000-0000-4000-8000-000000000001", OccurredAt: testOccurredAt, RecordActor: RecordActor{Kind: ActorKindMember, MemberID: "member-1"}}
	system := AuditMetadata{AuditID: m.AuditID, OccurredAt: m.OccurredAt, RecordActor: RecordActor{Kind: ActorKindSystem, Service: "geul-backend"}}
	collab := AuditMetadata{AuditID: m.AuditID, OccurredAt: m.OccurredAt, RecordActor: RecordActor{Kind: ActorKindSystem, Service: string(ServiceEditorCollab)}}
	anonymous := AuditMetadata{AuditID: m.AuditID, OccurredAt: m.OccurredAt, RecordActor: RecordActor{Kind: ActorKindAnonymous}}
	contributors := []string{"1b6bcad2-c90d-49e9-bec7-f9a4ba6b2894"}
	scheduledAt := time.Date(2026, 9, 1, 0, 0, 0, 0, time.UTC)
	previousSegments := []string{"segment-1"}
	segments := []string{"segment-2"}
	builders := map[string]func() (AuditRecord, error){
		"post participant author to collaborator": func() (AuditRecord, error) {
			return NewPostParticipantAuditRecord(m, "post-1", "member-2", AuditRelationshipAuthor, AuditRelationshipCollaborator)
		},
		"post participant added as author": func() (AuditRecord, error) {
			return NewPostParticipantAuditRecord(m, "post-1", "member-2", AuditRelationshipNone, AuditRelationshipAuthor)
		},
		"post participant added as collaborator": func() (AuditRecord, error) {
			return NewPostParticipantAuditRecord(m, "post-1", "member-2", AuditRelationshipNone, AuditRelationshipCollaborator)
		},
		"post File Block download policy": func() (AuditRecord, error) {
			return NewPostFileBlockDownloadPolicyAuditRecord(m, "post-1", "block-1", "file-1", AuditStateDisabled, AuditStateRestricted, previousSegments, segments)
		},
		"page File Block download policy": func() (AuditRecord, error) {
			return NewPageFileBlockDownloadPolicyAuditRecord(m, "page-1", "block-2", "file-2", AuditStatePublic, AuditStateAuthenticated, nil, nil)
		},
		"map theme content": func() (AuditRecord, error) {
			return NewMapThemeContentUpdatedAuditRecord(m, "theme-1")
		},
		"post source locale":          func() (AuditRecord, error) { return NewPostSourceLocaleAuditRecord(m, "post-1", "en", "zh-CN") },
		"page source locale":          func() (AuditRecord, error) { return NewPageSourceLocaleAuditRecord(m, "page-1", "en", "ko") },
		"work source locale":          func() (AuditRecord, error) { return NewWorkSourceLocaleAuditRecord(m, "work-1", "en", "ko") },
		"post series source locale":   func() (AuditRecord, error) { return NewPostSeriesSourceLocaleAuditRecord(m, "series-1", "en", "ko") },
		"program event source locale": func() (AuditRecord, error) { return NewProgramEventSourceLocaleAuditRecord(m, "event-1", "en", "ko") },
		"release source locale":       func() (AuditRecord, error) { return NewReleaseSourceLocaleAuditRecord(m, "release-1", "en", "ko") },
		"artist source locale":        func() (AuditRecord, error) { return NewArtistSourceLocaleAuditRecord(m, "artist-1", "en", "ko") },
		"label source locale":         func() (AuditRecord, error) { return NewLabelSourceLocaleAuditRecord(m, "label-1", "en", "ko") },
		"menu source locale":          func() (AuditRecord, error) { return NewMenuSourceLocaleAuditRecord(m, "menu-1", "en", "ko") },
		"campaign source locale":      func() (AuditRecord, error) { return NewCampaignSourceLocaleAuditRecord(m, "campaign-1", "en", "ko") },
		"form source locale":          func() (AuditRecord, error) { return NewFormSourceLocaleAuditRecord(m, "form-1", "en", "ko") },
		"email template source locale": func() (AuditRecord, error) {
			return NewEmailTemplateSourceLocaleAuditRecord(m, "template-1", "en", "ko")
		},
		"email layout source locale": func() (AuditRecord, error) { return NewEmailLayoutSourceLocaleAuditRecord(m, "layout-1", "en", "ko") },
		"privacy source locale":      func() (AuditRecord, error) { return NewPrivacySourceLocaleAuditRecord(m, "privacy-1", 2, "en", "ko") },
		"terms source locale":        func() (AuditRecord, error) { return NewTermsSourceLocaleAuditRecord(m, "terms-1", 1, "en", "ko") },
		"site settings fields":       func() (AuditRecord, error) { return NewSiteSettingsUpdatedAuditRecord(m, []string{"site_title"}) },
		"post version": func() (AuditRecord, error) {
			return NewPostVersionCreatedAuditRecord(collab, "post-1", "version-1", contributors)
		},
		"post configuration": func() (AuditRecord, error) { return NewPostConfigurationAuditRecord(m, "post-1", []string{"slug"}) },
		"post lifecycle status": func() (AuditRecord, error) {
			return NewPostStatusLifecycleAuditRecord(m, "post-1", AuditStateDraft, AuditStatePublished)
		},
		"post lifecycle schedule": func() (AuditRecord, error) {
			return NewPostScheduleLifecycleAuditRecord(m, "post-1", AuditStateDraft, AuditStateScheduled, scheduledAt, "Asia/Seoul")
		},
		"post featured image": func() (AuditRecord, error) {
			return NewPostFeaturedImageAuditRecord(m, "post-1", "asset-1", AuditCollectionOperationAdded)
		},
		"post share link": func() (AuditRecord, error) {
			return NewPostShareLinkAuditRecord(m, "post-1", "link-1", AuditItemOperationCreated)
		},
		"post comment": func() (AuditRecord, error) {
			return NewPostCommentAuditRecord(m, "post-1", "comment-1", AuditItemOperationUpdated)
		},
		"post version restore": func() (AuditRecord, error) { return NewPostVersionRestoreAuditRecord(m, "post-1", "version-1") },
		"page version": func() (AuditRecord, error) {
			return NewPageVersionCreatedAuditRecord(collab, "page-1", "version-1", contributors)
		},
		"page configuration": func() (AuditRecord, error) { return NewPageConfigurationAuditRecord(m, "page-1", []string{"slug"}) },
		"page lifecycle": func() (AuditRecord, error) {
			return NewPageLifecycleAuditRecord(m, "page-1", AuditStateDraft, AuditStatePublished)
		},
		"page featured image": func() (AuditRecord, error) {
			return NewPageFeaturedImageAuditRecord(m, "page-1", "asset-1", AuditCollectionOperationAdded)
		},
		"page share link": func() (AuditRecord, error) {
			return NewPageShareLinkAuditRecord(m, "page-1", "link-1", AuditItemOperationCreated)
		},
		"page version restore": func() (AuditRecord, error) { return NewPageVersionRestoreAuditRecord(m, "page-1", "version-1") },
		"work version": func() (AuditRecord, error) {
			return NewWorkVersionCreatedAuditRecord(collab, "work-1", "version-1", contributors)
		},
		"work metadata": func() (AuditRecord, error) { return NewWorkMetadataAuditRecord(m, "work-1", []string{"slug"}) },
		"work lifecycle": func() (AuditRecord, error) {
			return NewWorkLifecycleAuditRecord(m, "work-1", AuditStateDraft, AuditStatePublished)
		},
		"work featured image": func() (AuditRecord, error) {
			return NewWorkFeaturedImageAuditRecord(m, "work-1", "asset-1", AuditCollectionOperationAdded)
		},
		"work credit": func() (AuditRecord, error) {
			return NewWorkCreditAuditRecord(m, "work-1", "credit-1", AuditItemOperationUpdated)
		},
		"work share link": func() (AuditRecord, error) {
			return NewWorkShareLinkAuditRecord(m, "work-1", "link-1", AuditItemOperationCreated)
		},
		"work version restore": func() (AuditRecord, error) { return NewWorkVersionRestoreAuditRecord(m, "work-1", "version-1") },
		"legal lifecycle": func() (AuditRecord, error) {
			return NewLegalPolicyLifecycleAuditRecord(m, "policy-1", AuditPolicyTypeTerms, 1, []string{"status"}, AuditStateDraft, AuditStateScheduled, nil)
		},
		"legal lifecycle effective schedule": func() (AuditRecord, error) {
			return NewLegalPolicyLifecycleAuditRecord(m, "policy-1", AuditPolicyTypeTerms, 1, []string{"status", "effective_at"}, AuditStateDraft, AuditStateScheduled, &scheduledAt)
		},
		"legal share link": func() (AuditRecord, error) {
			return NewLegalPolicyShareLinkAuditRecord(m, "policy-1", AuditPolicyTypeTerms, 1, AuditItemOperationCreated, "link-1")
		},
		"file rename": func() (AuditRecord, error) { return NewFileRenamedAuditRecord(m, "file-1") },
		"file move":   func() (AuditRecord, error) { return NewFileMovedAuditRecord(m, "file-1", "folder-1", "") },
		"file move between folders": func() (AuditRecord, error) {
			return NewFileMovedAuditRecord(m, "file-1", "folder-1", "folder-2")
		},
		"work File Block download policy": func() (AuditRecord, error) {
			return NewWorkFileBlockDownloadPolicyAuditRecord(m, "work-1", "block-3", "file-3", AuditStateRestricted, AuditStateRestricted, nil, []string{"segment-1"})
		},
		"program event File Block download policy": func() (AuditRecord, error) {
			return NewProgramEventFileBlockDownloadPolicyAuditRecord(m, "event-1", "block-4", "file-4", AuditStateRestricted, AuditStatePublic, []string{"segment-1"}, nil)
		},
		"release Track download policy": func() (AuditRecord, error) {
			return NewReleaseTrackDownloadPolicyAuditRecord(m, "release-1", "track-1", "file-5", AuditStateDisabled, AuditStatePublic, nil, nil)
		},
		"folder rename": func() (AuditRecord, error) { return NewFileFolderRenamedAuditRecord(m, "folder-1") },
		"folder move":   func() (AuditRecord, error) { return NewFileFolderMovedAuditRecord(m, "folder-1", "", "folder-2") },
		"folder move between parents": func() (AuditRecord, error) {
			return NewFileFolderMovedAuditRecord(m, "folder-1", "folder-2", "folder-3")
		},
		"category metadata": func() (AuditRecord, error) {
			return NewCategoryMetadataUpdatedAuditRecord(m, "category-1", []string{"name"})
		},
		"tag metadata":   func() (AuditRecord, error) { return NewTagMetadataUpdatedAuditRecord(m, "tag-1", []string{"name"}) },
		"genre metadata": func() (AuditRecord, error) { return NewGenreMetadataUpdatedAuditRecord(m, "genre-1", []string{"name"}) },
		"style metadata": func() (AuditRecord, error) { return NewStyleMetadataUpdatedAuditRecord(m, "style-1", []string{"name"}) },
		"format metadata": func() (AuditRecord, error) {
			return NewFormatMetadataUpdatedAuditRecord(m, "format-1", []string{"name"})
		},
		"client metadata": func() (AuditRecord, error) {
			return NewClientMetadataUpdatedAuditRecord(m, "client-1", []string{"name"})
		},
		"client logo": func() (AuditRecord, error) {
			return NewClientLogoUpdatedAuditRecord(m, "client-1", AuditAssetSlotLight, AuditCollectionOperationAdded, "file-1")
		},
		"map place metadata": func() (AuditRecord, error) {
			return NewMapPlaceMetadataUpdatedAuditRecord(m, "place-1", []string{"name"})
		},
		"map place image": func() (AuditRecord, error) {
			return NewMapPlaceImageUpdatedAuditRecord(m, "place-1", AuditCollectionOperationAdded, "file-1")
		},
		"audience config": func() (AuditRecord, error) {
			return NewAudienceSegmentConfigUpdatedAuditRecord(m, "audience-1", []string{"name"})
		},
		"audience lifecycle": func() (AuditRecord, error) {
			return NewAudienceSegmentLifecycleUpdatedAuditRecord(m, "audience-1", AuditStateActive, AuditStateArchived)
		},
		"menu source": func() (AuditRecord, error) { return NewMenuSourceUpdatedAuditRecord(m, "menu-1", []string{"name"}) },
		"mail adapter config": func() (AuditRecord, error) {
			return NewMailAdapterConfigUpdatedAuditRecord(m, "adapter-1", []string{"name"})
		},
		"translation settings": func() (AuditRecord, error) {
			return NewTranslationSettingsUpdatedAuditRecord(m, []string{"default_locale"})
		},
		"translation provider config": func() (AuditRecord, error) {
			return NewTranslationProviderConfigUpdatedAuditRecord(m, "provider-1", []string{"name"})
		},
		"email suppression release": func() (AuditRecord, error) {
			return NewEmailSuppressionReleasedAuditRecord(m, "suppression-1")
		},
		"artist lifecycle": func() (AuditRecord, error) {
			return NewArtistLifecycleAuditRecord(m, "artist-1", AuditStateDraft, AuditStatePublished)
		},
		"artist gallery": func() (AuditRecord, error) {
			return NewArtistGalleryAuditRecord(m, "artist-1", []string{"file-1", "file-2"})
		},
		"artist participant": func() (AuditRecord, error) {
			return NewArtistParticipantAuditRecord(m, "artist-1", "member-2", AuditRelationshipNone, AuditRelationshipOwner)
		},
		"artist share link": func() (AuditRecord, error) {
			return NewArtistShareLinkAuditRecord(m, "artist-1", "link-1", AuditItemOperationCreated)
		},
		"label lifecycle": func() (AuditRecord, error) {
			return NewLabelLifecycleAuditRecord(m, "label-1", AuditStateDraft, AuditStatePublished)
		},
		"label participant": func() (AuditRecord, error) {
			return NewLabelParticipantAuditRecord(m, "label-1", "member-2", AuditRelationshipNone, AuditRelationshipManager)
		},
		"label logo": func() (AuditRecord, error) {
			return NewLabelLogoAuditRecord(m, "label-1", AuditAssetSlotLight, AuditCollectionOperationAdded, "asset-1")
		},
		"label share link": func() (AuditRecord, error) {
			return NewLabelShareLinkAuditRecord(m, "label-1", "link-1", AuditItemOperationCreated)
		},
		"post series metadata": func() (AuditRecord, error) {
			return NewPostSeriesSourceMetadataAuditRecord(m, "series-1", []string{"slug"})
		},
		"post series lifecycle": func() (AuditRecord, error) {
			return NewPostSeriesLifecycleAuditRecord(m, "series-1", AuditStateDraft, AuditStatePublished)
		},
		"post series manager": func() (AuditRecord, error) {
			return NewPostSeriesManagerAuditRecord(m, "series-1", "member-2", AuditRelationshipNone, AuditRelationshipManager)
		},
		"post series membership": func() (AuditRecord, error) {
			return NewPostSeriesMembershipAuditRecord(m, "series-1", "post-1", "", "series-1")
		},
		"post series membership move": func() (AuditRecord, error) {
			return NewPostSeriesMembershipAuditRecord(m, "series-1", "post-1", "series-2", "series-1")
		},
		"post series membership clear": func() (AuditRecord, error) {
			return NewPostSeriesMembershipAuditRecord(m, "series-1", "post-1", "series-1", "")
		},
		"post series order": func() (AuditRecord, error) {
			return NewPostSeriesOrderAuditRecord(m, "series-1", []string{"post-2", "post-1"})
		},
		"post series featured image": func() (AuditRecord, error) {
			return NewPostSeriesFeaturedImageAuditRecord(m, "series-1", AuditCollectionOperationAdded, "file-1")
		},
		"program event type config": func() (AuditRecord, error) {
			return NewProgramEventTypeConfigUpdatedAuditRecord(m, "type-1", []string{"slug"})
		},
		"program event metadata": func() (AuditRecord, error) { return NewProgramEventMetadataAuditRecord(m, "event-1", []string{"slug"}) },
		"program event poster": func() (AuditRecord, error) {
			return NewProgramEventPosterAuditRecord(m, "event-1", "file-1", AuditCollectionOperationAdded)
		},
		"program event media": func() (AuditRecord, error) {
			return NewProgramEventChildAuditRecord(m, "event-1", "media", "media-1", AuditItemOperationCreated)
		},
		"program event child order": func() (AuditRecord, error) {
			return NewProgramEventChildOrderAuditRecord(m, "event-1", "credits", []string{"credit-2", "credit-1"})
		},
		"program event lifecycle": func() (AuditRecord, error) {
			return NewProgramEventLifecycleAuditRecord(m, "event-1", AuditStateDraft, AuditStatePublished)
		},
		"program event series metadata": func() (AuditRecord, error) {
			return NewProgramEventSeriesMetadataAuditRecord(m, "event-series-1", []string{"title"})
		},
		"program event series poster": func() (AuditRecord, error) {
			return NewProgramEventSeriesPosterAuditRecord(m, "event-series-1", "file-1", AuditCollectionOperationAdded)
		},
		"program event series lifecycle": func() (AuditRecord, error) {
			return NewProgramEventSeriesLifecycleAuditRecord(m, "event-series-1", AuditStateDraft, AuditStatePublished)
		},
		"release metadata": func() (AuditRecord, error) { return NewReleaseMetadataAuditRecord(m, "release-1", []string{"slug"}) },
		"release track": func() (AuditRecord, error) {
			return NewReleaseTrackAuditRecord(m, "release-1", "track-1", AuditItemOperationCreated)
		},
		"release track order": func() (AuditRecord, error) {
			return NewReleaseTrackOrderAuditRecord(m, "release-1", []string{"track-2", "track-1"})
		},
		"release track audio": func() (AuditRecord, error) {
			return NewReleaseTrackAudioAuditRecord(m, "release-1", "track-1", "file-1", AuditCollectionOperationAdded)
		},
		"release artwork": func() (AuditRecord, error) {
			return NewReleaseArtworkAuditRecord(m, "release-1", "file-1", AuditCollectionOperationAdded)
		},
		"release lifecycle": func() (AuditRecord, error) {
			return NewReleaseLifecycleAuditRecord(m, "release-1", AuditStateDraft, AuditStatePublished)
		},
		"release share link": func() (AuditRecord, error) {
			return NewReleaseShareLinkAuditRecord(m, "release-1", "link-1", AuditItemOperationCreated)
		},
		"campaign target layout": func() (AuditRecord, error) {
			return NewCampaignTargetLayoutAuditRecord(m, "campaign-1", []string{"layout"})
		},
		"campaign lifecycle": func() (AuditRecord, error) {
			return NewCampaignStatusLifecycleAuditRecord(m, "campaign-1", AuditStateDraft, AuditStateScheduled)
		},
		"campaign schedule": func() (AuditRecord, error) {
			return NewCampaignScheduleLifecycleAuditRecord(m, "campaign-1", AuditStateDraft, AuditStateScheduled, scheduledAt)
		},
		"campaign delivery run lifecycle": func() (AuditRecord, error) {
			return NewCampaignDeliveryRunLifecycleAuditRecord(m, "campaign-1", AuditStateScheduled, AuditStateSending, "run-1")
		},
		"template metadata": func() (AuditRecord, error) {
			return NewEmailTemplateMetadataAuditRecord(m, "template-1", []string{"name"})
		},
		"template layout": func() (AuditRecord, error) {
			return NewEmailTemplateLayoutRelationAuditRecord(m, "template-1", "", "layout-1")
		},
		"template layout replacement": func() (AuditRecord, error) {
			return NewEmailTemplateLayoutRelationAuditRecord(m, "template-1", "layout-2", "layout-1")
		},
		"template layout cleared": func() (AuditRecord, error) {
			return NewEmailTemplateLayoutRelationAuditRecord(m, "template-1", "layout-1", "")
		},
		"layout metadata": func() (AuditRecord, error) { return NewEmailLayoutMetadataAuditRecord(m, "layout-1", []string{"name"}) },
		"email event mapping": func() (AuditRecord, error) {
			return NewEmailEventMappingTemplateAuditRecord(m, "welcome", "", "template-1")
		},
		"email event mapping replacement": func() (AuditRecord, error) {
			return NewEmailEventMappingTemplateAuditRecord(m, "welcome", "template-2", "template-1")
		},
		"email event mapping cleared": func() (AuditRecord, error) {
			return NewEmailEventMappingTemplateAuditRecord(m, "welcome", "template-1", "")
		},
		"form settings": func() (AuditRecord, error) { return NewFormSettingsAuditRecord(m, "form-1", []string{"slug"}) },
		"form lifecycle": func() (AuditRecord, error) {
			return NewFormLifecycleAuditRecord(m, "form-1", AuditStateDraft, AuditStatePublished)
		},
		"form featured image": func() (AuditRecord, error) {
			return NewFormFeaturedImageAuditRecord(m, "form-1", "file-1", AuditCollectionOperationAdded)
		},
		"form share link": func() (AuditRecord, error) {
			return NewFormShareLinkAuditRecord(m, "form-1", "link-1", AuditItemScopeForm, AuditItemOperationCreated)
		},
		"anonymous form submission": func() (AuditRecord, error) {
			return NewFormSubmissionCreatedAuditRecord(anonymous, "submission-1", "form-1")
		},
		"member onboarding": func() (AuditRecord, error) { return NewMemberOnboardingCompletedAuditRecord(m, "member-1", "Member") },
		"member profile": func() (AuditRecord, error) {
			return NewMemberProfileUpdatedAuditRecord(m, "member-1", []string{"nickname"}, "Member")
		},
		"member profile fields without nickname": func() (AuditRecord, error) {
			return NewMemberProfileUpdatedAuditRecord(m, "member-1", []string{"website", "bio"}, "")
		},
		"member avatar": func() (AuditRecord, error) {
			return NewMemberAvatarUpdatedAuditRecord(m, "member-1", AuditCollectionOperationAdded, "asset-1")
		},
		"member preference": func() (AuditRecord, error) {
			return NewMemberPreferencesUpdatedAuditRecord(m, "member-1", []string{"preferred_locale"}, "ko", "")
		},
		"member cookie consent preference": func() (AuditRecord, error) {
			return NewMemberPreferencesUpdatedAuditRecord(m, "member-1", []string{"cookie_consent"}, "", "consent-1")
		},
		"member tags": func() (AuditRecord, error) {
			return NewMemberTagsUpdatedAuditRecord(m, "member-1", []string{"tag-2", "tag-1"})
		},
		"member role":   func() (AuditRecord, error) { return NewMemberRoleUpdatedAuditRecord(m, "member-1", "user", "author") },
		"member status": func() (AuditRecord, error) { return NewMemberBannedAuditRecord(m, "member-1") },
		"member unban system": func() (AuditRecord, error) {
			return NewMemberUnbannedAuditRecord(system, "member-1")
		},
		"account canonical email": func() (AuditRecord, error) {
			return NewAccountCanonicalEmailUpdatedAuditRecord(m, "member-1", "old@example.test", "new@example.test")
		},
		"account email login": func() (AuditRecord, error) {
			return NewAccountEmailLoginAddedAuditRecord(m, "member-1", "new@example.test")
		},
		"account social login": func() (AuditRecord, error) {
			return NewAccountSocialLoginAddedAuditRecord(m, "member-1", "google", "google-subject")
		},
		"account passkey": func() (AuditRecord, error) {
			return NewAccountPasskeyAddedAuditRecord(m, "member-1", []string{"passkey-1"})
		},
		"account sessions": func() (AuditRecord, error) {
			return NewAccountSessionRevokedAuditRecord(m, "member-1", AccountSessionScopeOthers, []string{"018f47a2-8a3d-4e17-9d42-6f12c89b5678", "018f47a2-8a3d-4e17-9d42-6f12c89b1234"})
		},
		"account newsletter": func() (AuditRecord, error) {
			return NewAccountNewsletterSubscriptionUpdatedAuditRecord(m, "member-1", AuditStateSubscribed, AuditStateUnsubscribed)
		},
		"account deletion lifecycle": func() (AuditRecord, error) {
			return NewAccountDeletionRequestedAuditRecord(m, "member-1", AuditStateNone)
		},
		"account deletion scheduled": func() (AuditRecord, error) {
			return NewAccountDeletionScheduledAuditRecord(m, "member-1", AuditStateConfirmationPending)
		},
		"account deletion cancelled": func() (AuditRecord, error) {
			return NewAccountDeletionCancelledAuditRecord(m, "member-1")
		},
		"account deletion recovered": func() (AuditRecord, error) {
			return NewAccountDeletionRecoveredAuditRecord(m, "member-1")
		},
		"account deleted system": func() (AuditRecord, error) { return NewAccountDeletedAuditRecord(system, "member-1") },
	}
	if len(builders) != len(manifest) {
		t.Fatalf("builder invocations = %d, manifest cases = %d", len(builders), len(manifest))
	}
	seen := map[string]bool{}
	for _, expected := range manifest {
		build, ok := builders[expected.Case]
		if !ok {
			t.Fatalf("%s has no Go semantic builder invocation", expected.Case)
		}
		if seen[expected.Case] {
			t.Fatalf("%s invoked more than once", expected.Case)
		}
		seen[expected.Case] = true
		record, err := build()
		if err != nil {
			t.Fatalf("%s: %v", expected.Case, err)
		}
		if err := record.Validate(); err != nil {
			t.Fatalf("%s invalid: %v", expected.Case, err)
		}
		expectedActor := RecordActor{Kind: expected.ActorKind}
		switch expected.ActorKind {
		case ActorKindMember:
			expectedActor.MemberID = "member-1"
		case ActorKindSystem:
			expectedActor.Service = expected.ActorService
		}
		if record.AuditID != m.AuditID || !record.OccurredAt.Equal(m.OccurredAt) || record.Correlation != (Correlation{}) || record.RecordActor != expectedActor {
			t.Fatalf("%s envelope mismatch: %#v", expected.Case, record)
		}
		wire, _ := json.Marshal(record)
		var actual map[string]any
		_ = json.Unmarshal(wire, &actual)
		for _, key := range []string{"audit_id", "occurred_at", "action", "target_type", "target_id", "request_id", "trace_id", "span_id", "actor_kind", "actor_member_id", "actor_service"} {
			delete(actual, key)
		}
		if record.Action != expected.Action || record.TargetType != expected.TargetType || record.TargetID != expected.TargetID || !jsonValuesEqual(actual, expected.Attributes) {
			t.Fatalf("%s wire mismatch: %#v", expected.Case, actual)
		}
	}
}
