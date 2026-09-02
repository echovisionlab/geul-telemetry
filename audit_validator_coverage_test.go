package telemetry

import (
	"testing"
	"time"
)

func TestAuditValidatorsAcceptEveryReviewedVariant(t *testing.T) {
	member := RecordActor{Kind: ActorKindMember, MemberID: "member-1"}
	collab := RecordActor{Kind: ActorKindSystem, Service: string(ServiceEditorCollab)}
	backend := RecordActor{Kind: ActorKindSystem, Service: string(ServiceBackend)}
	ids := []string{"file-1", "file-2"}
	at := time.Date(2026, time.August, 1, 12, 0, 0, 0, time.UTC)
	version := int64(1)
	cases := []struct {
		name string
		call func() error
	}{
		{"artist status", func() error {
			return validateArtistUpdate(AuditRecord{RecordActor: member, ChangedFields: []string{"status"}, PreviousState: AuditStateDraft, NewState: AuditStatePublished})
		}},
		{"artist gallery", func() error {
			return validateArtistUpdate(AuditRecord{RecordActor: member, ChangedFields: []string{"gallery"}, FileIDs: &ids})
		}},
		{"artist participant", func() error {
			return validateArtistUpdate(AuditRecord{RecordActor: member, ChangedFields: []string{"participants"}, SubjectMemberID: "member-2", PreviousRelationship: AuditRelationshipNone, NewRelationship: AuditRelationshipOwner})
		}},
		{"artist share link", func() error {
			return validateArtistUpdate(AuditRecord{RecordActor: member, ChangedFields: []string{"share_links"}, ItemID: "link-1", ItemOperation: AuditItemOperationCreated})
		}},
		{"artist member actor rejection", func() error {
			return expectValidatorError(validateArtistUpdate(AuditRecord{ChangedFields: []string{"status"}, PreviousState: AuditStateDraft, NewState: AuditStatePublished}))
		}},
		{"artist gallery requires file ids", func() error {
			return expectValidatorError(validateArtistUpdate(AuditRecord{RecordActor: member, ChangedFields: []string{"gallery"}}))
		}},
		{"artist gallery rejects invalid ids", func() error {
			bad := []string{" "}
			return expectValidatorError(validateArtistUpdate(AuditRecord{RecordActor: member, ChangedFields: []string{"gallery"}, FileIDs: &bad}))
		}},
		{"artist gallery rejects extras", func() error {
			return expectValidatorError(validateArtistUpdate(AuditRecord{RecordActor: member, ChangedFields: []string{"gallery"}, FileIDs: &ids, AssetID: "extra-1"}))
		}},
		{"artist participant member rejection", func() error {
			return expectValidatorError(validateArtistUpdate(AuditRecord{ChangedFields: []string{"participants"}}))
		}},
		{"artist share link member rejection", func() error {
			return expectValidatorError(validateArtistUpdate(AuditRecord{ChangedFields: []string{"share_links"}}))
		}},
		{"label status", func() error {
			return validateLabelUpdate(AuditRecord{RecordActor: member, ChangedFields: []string{"status"}, PreviousState: AuditStatePublished, NewState: AuditStateDraft})
		}},
		{"label participant", func() error {
			return validateLabelUpdate(AuditRecord{RecordActor: member, ChangedFields: []string{"participants"}, SubjectMemberID: "member-2", PreviousRelationship: AuditRelationshipNone, NewRelationship: AuditRelationshipManager})
		}},
		{"label logo", func() error {
			return validateLabelUpdate(AuditRecord{RecordActor: member, ChangedFields: []string{"logo"}, AssetSlot: AuditAssetSlotLight, AssetID: "file-1", CollectionOperation: AuditCollectionOperationAdded})
		}},
		{"label share link", func() error {
			return validateLabelUpdate(AuditRecord{RecordActor: member, ChangedFields: []string{"share_links"}, ItemID: "link-1", ItemOperation: AuditItemOperationDeleted})
		}},
		{"label member actor rejection", func() error {
			return expectValidatorError(validateLabelUpdate(AuditRecord{ChangedFields: []string{"logo"}, AssetSlot: AuditAssetSlotLight, AssetID: "file-1", CollectionOperation: AuditCollectionOperationAdded}))
		}},
		{"label status member rejection", func() error {
			return expectValidatorError(validateLabelUpdate(AuditRecord{ChangedFields: []string{"status"}}))
		}},
		{"label participant member rejection", func() error {
			return expectValidatorError(validateLabelUpdate(AuditRecord{ChangedFields: []string{"participants"}}))
		}},
		{"label logo requires slot", func() error {
			return expectValidatorError(validateLabelUpdate(AuditRecord{RecordActor: member, ChangedFields: []string{"logo"}}))
		}},
		{"label logo requires binding", func() error {
			return expectValidatorError(validateLabelUpdate(AuditRecord{RecordActor: member, ChangedFields: []string{"logo"}, AssetSlot: AuditAssetSlotLight}))
		}},
		{"label logo rejects extras", func() error {
			return expectValidatorError(validateLabelUpdate(AuditRecord{RecordActor: member, ChangedFields: []string{"logo"}, AssetSlot: AuditAssetSlotLight, AssetID: "file-1", CollectionOperation: AuditCollectionOperationAdded, FileID: "extra-1"}))
		}},
		{"label share link member rejection", func() error {
			return expectValidatorError(validateLabelUpdate(AuditRecord{ChangedFields: []string{"share_links"}}))
		}},
		{"ordered identifiers", func() error { return validateOrderedIdentifiers("file_ids", ids) }},
		{"ordered identifiers duplicate", func() error {
			return expectValidatorError(validateOrderedIdentifiers("file_ids", []string{"file-1", "file-1"}))
		}},

		{"email template layout", func() error {
			return validateEmailTemplateUpdate(AuditRecord{RecordActor: member, ChangedFields: []string{"layout"}, PreviousItemID: "layout-1", ItemID: "layout-2"})
		}},
		{"email template fields", func() error {
			return validateEmailTemplateUpdate(AuditRecord{RecordActor: member, ChangedFields: []string{"name"}})
		}},
		{"email layout fields", func() error {
			return validateEmailLayoutUpdate(AuditRecord{RecordActor: member, ChangedFields: []string{"name"}})
		}},
		{"email mapping", func() error {
			return validateEmailMappingUpdate(AuditRecord{RecordActor: member, ChangedFields: []string{"template"}, EventName: "welcome", PreviousItemID: "template-1", ItemID: "template-2"})
		}},
		{"email member only", func() error { return validateEmailAuthoringMemberOnly(AuditRecord{RecordActor: member}) }},
		{"email member actor rejection", func() error { return expectValidatorError(validateEmailAuthoringMemberActor(collabRecord())) }},
		{"email template rejects collab configuration", func() error {
			return expectValidatorError(validateEmailTemplateUpdate(AuditRecord{RecordActor: collab, ChangedFields: []string{"name"}}))
		}},
		{"email layout rejects collab configuration", func() error {
			return expectValidatorError(validateEmailLayoutUpdate(AuditRecord{RecordActor: collab, ChangedFields: []string{"name"}}))
		}},
		{"email mapping rejects extra", func() error {
			return expectValidatorError(validateEmailMappingUpdate(AuditRecord{RecordActor: member, ChangedFields: []string{"template"}, EventName: "welcome", ItemID: "template-1", AssetID: "extra-1"}))
		}},
		{"email member only rejects system", func() error { return expectValidatorError(validateEmailAuthoringMemberOnly(collabRecord())) }},

		{"form status", func() error {
			return validateFormUpdate(AuditRecord{ChangedFields: []string{"status"}, PreviousState: AuditStateDraft, NewState: AuditStatePublished})
		}},
		{"form featured image", func() error {
			return validateFormUpdate(AuditRecord{ChangedFields: []string{"featured_image"}, FileID: "file-1", CollectionOperation: AuditCollectionOperationAdded})
		}},
		{"form share links", func() error {
			return validateFormUpdate(AuditRecord{ChangedFields: []string{"share_links"}, ItemID: "link-1", ItemScope: AuditItemScopeForm, ItemOperation: AuditItemOperationCreated})
		}},
		{"form settings", func() error { return validateFormUpdate(AuditRecord{ChangedFields: []string{"slug"}}) }},

		{"post participant", func() error {
			return validatePostConfigurationOrLifecycle(AuditRecord{RecordActor: member, ChangedFields: []string{"authors"}, SubjectMemberID: "member-2", PreviousRelationship: AuditRelationshipNone, NewRelationship: AuditRelationshipAuthor})
		}},
		{"post featured image", func() error {
			return validatePostConfigurationOrLifecycle(AuditRecord{RecordActor: member, ChangedFields: []string{"featured_image"}, AssetID: "file-1", CollectionOperation: AuditCollectionOperationAdded})
		}},
		{"post share link", func() error {
			return validatePostConfigurationOrLifecycle(AuditRecord{RecordActor: member, ChangedFields: []string{"share_links"}, ItemID: "link-1", ItemOperation: AuditItemOperationCreated})
		}},
		{"post version restore", func() error {
			return validatePostConfigurationOrLifecycle(AuditRecord{RecordActor: member, ChangedFields: []string{"version_restore"}, VersionID: "version-1"})
		}},
		{"post comment", func() error {
			return validatePostConfigurationOrLifecycle(AuditRecord{RecordActor: member, ChangedFields: []string{"comments"}, ItemID: "comment-1", ItemOperation: AuditItemOperationUpdated})
		}},
		{"post lifecycle", func() error {
			return validatePostConfigurationOrLifecycle(AuditRecord{RecordActor: member, ChangedFields: []string{"status"}, PreviousState: AuditStateDraft, NewState: AuditStatePublished})
		}},
		{"post schedule", func() error {
			return validatePostConfigurationOrLifecycle(AuditRecord{RecordActor: member, ChangedFields: []string{"schedule"}, PreviousState: AuditStateDraft, NewState: AuditStateScheduled, ScheduledAt: &at, ScheduledTimeZone: "UTC"})
		}},
		{"post configuration", func() error {
			return validatePostConfigurationOrLifecycle(AuditRecord{RecordActor: member, ChangedFields: []string{"slug"}})
		}},
		{"post collab version actor", func() error {
			return validateContentVersionActor(AuditRecord{RecordActor: collab, ChangedFields: []string{"version"}})
		}},
		{"post other system actor rejection", func() error {
			return expectValidatorError(validateContentVersionActor(AuditRecord{RecordActor: backend, ChangedFields: []string{"slug"}}))
		}},
		{"post rejects non collab checkpoint", func() error {
			return expectValidatorError(validatePostConfigurationOrLifecycle(AuditRecord{RecordActor: backend, ChangedFields: []string{"version"}}))
		}},
		{"post rejects changed field", func() error {
			return expectValidatorError(validatePostConfigurationOrLifecycle(AuditRecord{RecordActor: member, ChangedFields: []string{"unknown"}}))
		}},

		{"program poster", func() error {
			return validateProgramEventUpdate(AuditRecord{ChangedFields: []string{"poster"}, FileID: "file-1", CollectionOperation: AuditCollectionOperationAdded})
		}},
		{"program media reorder", func() error {
			return validateProgramEventUpdate(AuditRecord{ChangedFields: []string{"media"}, ItemIDs: &ids})
		}},
		{"program credits item", func() error {
			return validateProgramEventUpdate(AuditRecord{ChangedFields: []string{"credits"}, ItemID: "credit-1", ItemOperation: AuditItemOperationDeleted})
		}},
		{"program lifecycle", func() error {
			return validateProgramEventUpdate(AuditRecord{ChangedFields: []string{"status"}, PreviousState: AuditStateDraft, NewState: AuditStatePublished})
		}},
		{"program fields", func() error {
			return validateProgramEventUpdate(AuditRecord{RecordActor: member, ChangedFields: []string{"slug"}})
		}},
		{"program system actor rejection", func() error {
			return expectValidatorError(validateProgramEventUpdate(AuditRecord{RecordActor: backend, ChangedFields: []string{"slug"}}))
		}},

		{"release tracks", func() error {
			return validateReleaseUpdate(AuditRecord{RecordActor: member, ChangedFields: []string{"tracks"}, ItemID: "track-1", ItemOperation: AuditItemOperationCreated})
		}},
		{"release track audio member", func() error {
			return validateReleaseUpdate(AuditRecord{RecordActor: member, ChangedFields: []string{"track_audio"}, ItemID: "track-1", FileID: "file-1", CollectionOperation: AuditCollectionOperationAdded})
		}},
		{"release track audio collab", func() error {
			return validateReleaseUpdate(AuditRecord{RecordActor: collab, ChangedFields: []string{"track_audio"}, ItemID: "track-1", FileID: "file-1", CollectionOperation: AuditCollectionOperationRemoved})
		}},
		{"release artwork", func() error {
			return validateReleaseUpdate(AuditRecord{RecordActor: member, ChangedFields: []string{"artwork"}, FileID: "file-1", CollectionOperation: AuditCollectionOperationAdded})
		}},
		{"release status", func() error {
			return validateReleaseUpdate(AuditRecord{RecordActor: member, ChangedFields: []string{"status"}, PreviousState: AuditStateDraft, NewState: AuditStatePublished})
		}},
		{"release share links", func() error {
			return validateReleaseUpdate(AuditRecord{RecordActor: member, ChangedFields: []string{"share_links"}, ItemID: "link-1", ItemOperation: AuditItemOperationCreated})
		}},
		{"release fields", func() error {
			return validateReleaseUpdate(AuditRecord{RecordActor: member, ChangedFields: []string{"genres", "slug"}})
		}},
		{"release system rejection", func() error {
			return expectValidatorError(validateReleaseUpdate(AuditRecord{RecordActor: backend, ChangedFields: []string{"status"}, PreviousState: AuditStateDraft, NewState: AuditStatePublished}))
		}},
		{"release track actor rejection", func() error {
			return expectValidatorError(validateReleaseUpdate(AuditRecord{RecordActor: backend, ChangedFields: []string{"track_audio"}, ItemID: "track-1", FileID: "file-1", CollectionOperation: AuditCollectionOperationAdded}))
		}},

		{"campaign fields", func() error {
			return validateCampaignUpdate(AuditRecord{RecordActor: member, ChangedFields: []string{"name"}})
		}},
		{"campaign member lifecycle", func() error {
			return validateCampaignUpdate(AuditRecord{RecordActor: member, ChangedFields: []string{"schedule"}, PreviousState: AuditStateDraft, NewState: AuditStateScheduled, ScheduledAt: &at})
		}},
		{"campaign delivery", func() error {
			return validateCampaignUpdate(AuditRecord{RecordActor: member, ChangedFields: []string{"delivery_run"}, PreviousState: AuditStateDraft, NewState: AuditStateSending, ItemID: "run-1"})
		}},
		{"campaign backend terminal", func() error {
			return validateCampaignUpdate(AuditRecord{RecordActor: backend, ChangedFields: []string{"status"}, PreviousState: AuditStateSending, NewState: AuditStateSent})
		}},
		{"campaign backend actor rejection", func() error {
			return expectValidatorError(validateCampaignUpdate(AuditRecord{RecordActor: backend, ChangedFields: []string{"status"}, PreviousState: AuditStateSending, NewState: AuditStateSending}))
		}},
		{"campaign backend nonterminal rejection", func() error {
			return expectValidatorError(validateCampaignUpdate(AuditRecord{RecordActor: backend, ChangedFields: []string{"status"}, PreviousState: AuditStateSending, NewState: AuditStateSending}))
		}},
		{"campaign rejects nonbackend system actor", func() error {
			return expectValidatorError(validateCampaignUpdate(AuditRecord{RecordActor: collab, ChangedFields: []string{"status"}, PreviousState: AuditStateSending, NewState: AuditStateSent}))
		}},
		{"campaign rejects timezone", func() error {
			return expectValidatorError(validateCampaignUpdate(AuditRecord{RecordActor: member, ChangedFields: []string{"schedule"}, PreviousState: AuditStateDraft, NewState: AuditStateScheduled, ScheduledAt: &at, ScheduledTimeZone: "UTC"}))
		}},

		{"legal share link", func() error {
			return validateLegalPolicyUpdate(AuditRecord{RecordActor: member, PolicyType: AuditPolicyTypeTerms, VersionNumber: &version, ChangedFields: []string{"share_links"}, ItemID: "link-1", ItemOperation: AuditItemOperationDeleted})
		}},
		{"legal member lifecycle", func() error {
			return validateLegalPolicyUpdate(AuditRecord{RecordActor: member, PolicyType: AuditPolicyTypePrivacy, VersionNumber: &version, ChangedFields: []string{"effective_at", "status"}, PreviousState: AuditStateDraft, NewState: AuditStateScheduled, EffectiveAt: &at})
		}},
		{"legal backend lifecycle", func() error {
			return validateLegalPolicyUpdate(AuditRecord{RecordActor: backend, PolicyType: AuditPolicyTypeTerms, VersionNumber: &version, ChangedFields: []string{"status"}, PreviousState: AuditStateScheduled, NewState: AuditStateActive})
		}},
		{"legal share actor rejection", func() error {
			return expectValidatorError(validateLegalPolicyUpdate(AuditRecord{RecordActor: collab, PolicyType: AuditPolicyTypeTerms, VersionNumber: &version, ChangedFields: []string{"share_links"}}))
		}},
		{"legal share item rejection", func() error {
			return expectValidatorError(validateLegalPolicyUpdate(AuditRecord{RecordActor: member, PolicyType: AuditPolicyTypeTerms, VersionNumber: &version, ChangedFields: []string{"share_links"}}))
		}},
		{"legal share extras rejection", func() error {
			return expectValidatorError(validateLegalPolicyUpdate(AuditRecord{RecordActor: member, PolicyType: AuditPolicyTypeTerms, VersionNumber: &version, ChangedFields: []string{"share_links"}, ItemID: "link-1", ItemOperation: AuditItemOperationCreated, AssetID: "extra-1"}))
		}},
		{"legal lifecycle actor rejection", func() error {
			return expectValidatorError(validateLegalPolicyUpdate(AuditRecord{PolicyType: AuditPolicyTypeTerms, VersionNumber: &version, ChangedFields: []string{"status"}, PreviousState: AuditStateDraft, NewState: AuditStateScheduled}))
		}},
		{"legal lifecycle changed field rejection", func() error {
			return expectValidatorError(validateLegalPolicyUpdate(AuditRecord{RecordActor: member, PolicyType: AuditPolicyTypeTerms, VersionNumber: &version, ChangedFields: []string{"unknown"}}))
		}},
		{"legal lifecycle bad time rejection", func() error {
			badTime := time.Time{}
			return expectValidatorError(validateLegalPolicyUpdate(AuditRecord{RecordActor: member, PolicyType: AuditPolicyTypeTerms, VersionNumber: &version, ChangedFields: []string{"effective_at"}, PreviousState: AuditStateDraft, NewState: AuditStateScheduled, EffectiveAt: &badTime}))
		}},
		{"legal lifecycle extras rejection", func() error {
			return expectValidatorError(validateLegalPolicyUpdate(AuditRecord{RecordActor: member, PolicyType: AuditPolicyTypeTerms, VersionNumber: &version, ChangedFields: []string{"status"}, PreviousState: AuditStateDraft, NewState: AuditStateScheduled, AssetID: "extra-1"}))
		}},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			if err := tc.call(); err != nil {
				t.Fatal(err)
			}
		})
	}
}

func collabRecord() AuditRecord {
	return AuditRecord{RecordActor: RecordActor{Kind: ActorKindSystem, Service: string(ServiceEditorCollab)}}
}

func expectValidatorError(err error) error {
	if err == nil {
		return errExpectedValidatorFailure{}
	}
	return nil
}

type errExpectedValidatorFailure struct{}

func (errExpectedValidatorFailure) Error() string { return "expected validator failure" }
