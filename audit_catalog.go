package telemetry

import (
	"fmt"
	"slices"
)

type auditCatalogEntry struct {
	targetType      string
	singletonTarget string
	allowAnonymous  bool
	systemServices  []ServiceName
	validate        auditRecordValidator
}

type auditRecordValidator func(AuditRecord) error

// auditCatalog is deliberately data, not a switch: the action owns its target
// type and actor boundary, while variant validation is dispatched separately.
// auditCatalog is the single complete action specification. Every action has
// one typed identifier, authoritative target, actor boundary, and validator.
var auditCatalog = map[AuditAction]auditCatalogEntry{
	AuditAccountDeleted: {targetType: "account", systemServices: []ServiceName{ServiceBackend}},
	AuditAccountUpdated: {targetType: "account"},
	AuditArtistCreated:  {targetType: "artist"}, AuditArtistDeleted: {targetType: "artist"}, AuditArtistUpdated: {targetType: "artist"},
	AuditAudienceSegmentCreated: {targetType: "audience_segment"}, AuditAudienceSegmentUpdated: {targetType: "audience_segment"},
	AuditCampaignCreated: {targetType: "campaign"}, AuditCampaignDeleted: {targetType: "campaign"}, AuditCampaignUpdated: {targetType: "campaign", systemServices: []ServiceName{ServiceBackend}},
	AuditCategoryCreated: {targetType: "category"}, AuditCategoryDeleted: {targetType: "category"}, AuditCategoryUpdated: {targetType: "category"},
	AuditClientCreated: {targetType: "client"}, AuditClientDeleted: {targetType: "client"}, AuditClientUpdated: {targetType: "client"},
	AuditEmailEventMappingUpdated: {targetType: "email_event_mapping"},
	AuditEmailLayoutCreated:       {targetType: "email_layout"}, AuditEmailLayoutDeleted: {targetType: "email_layout"}, AuditEmailLayoutUpdated: {targetType: "email_layout"},
	AuditEmailSuppressionUpdated: {targetType: "email_suppression"},
	AuditEmailTemplateCreated:    {targetType: "email_template"}, AuditEmailTemplateDeleted: {targetType: "email_template"}, AuditEmailTemplateUpdated: {targetType: "email_template"},
	AuditFileCreated: {targetType: "file"}, AuditFileDeleted: {targetType: "file", systemServices: []ServiceName{ServiceBackend}}, AuditFileUpdated: {targetType: "file"},
	AuditFileFolderCreated: {targetType: "file_folder"}, AuditFileFolderDeleted: {targetType: "file_folder"}, AuditFileFolderUpdated: {targetType: "file_folder"},
	AuditFormCreated: {targetType: "form"}, AuditFormDeleted: {targetType: "form"}, AuditFormUpdated: {targetType: "form"},
	AuditFormSubmissionCreated: {targetType: "form_submission", allowAnonymous: true}, AuditFormSubmissionDeleted: {targetType: "form_submission"},
	AuditFormatCreated: {targetType: "format"}, AuditFormatDeleted: {targetType: "format"}, AuditFormatUpdated: {targetType: "format"},
	AuditGenreCreated: {targetType: "genre"}, AuditGenreDeleted: {targetType: "genre"}, AuditGenreUpdated: {targetType: "genre"},
	AuditLabelCreated: {targetType: "label"}, AuditLabelDeleted: {targetType: "label"}, AuditLabelUpdated: {targetType: "label"},
	AuditLegalPolicyCreated: {targetType: "legal_policy"}, AuditLegalPolicyDeleted: {targetType: "legal_policy"}, AuditLegalPolicyUpdated: {targetType: "legal_policy", systemServices: []ServiceName{ServiceBackend}},
	AuditMailAdapterCreated: {targetType: "mail_adapter"}, AuditMailAdapterDeleted: {targetType: "mail_adapter"}, AuditMailAdapterUpdated: {targetType: "mail_adapter"},
	AuditMapPlaceCreated: {targetType: "map_place"}, AuditMapPlaceDeleted: {targetType: "map_place"}, AuditMapPlaceUpdated: {targetType: "map_place"},
	AuditMapThemeCreated: {targetType: "map_theme"}, AuditMapThemeDeleted: {targetType: "map_theme"}, AuditMapThemeUpdated: {targetType: "map_theme"},
	AuditMemberUpdated:    {targetType: "member", systemServices: []ServiceName{ServiceBackend}},
	AuditMemberTagCreated: {targetType: "member_tag"}, AuditMemberTagDeleted: {targetType: "member_tag"},
	AuditMenuCreated: {targetType: "menu"}, AuditMenuDeleted: {targetType: "menu"}, AuditMenuUpdated: {targetType: "menu"},
	AuditPageCreated: {targetType: "page"}, AuditPageDeleted: {targetType: "page"}, AuditPageUpdated: {targetType: "page", systemServices: []ServiceName{ServiceEditorCollab}},
	AuditPostCreated: {targetType: "post"}, AuditPostDeleted: {targetType: "post"}, AuditPostUpdated: {targetType: "post", systemServices: []ServiceName{ServiceEditorCollab}},
	AuditPostSeriesCreated: {targetType: "post_series"}, AuditPostSeriesDeleted: {targetType: "post_series"}, AuditPostSeriesUpdated: {targetType: "post_series"},
	AuditProgramEventCreated: {targetType: "program_event"}, AuditProgramEventDeleted: {targetType: "program_event"}, AuditProgramEventUpdated: {targetType: "program_event"},
	AuditProgramEventSeriesCreated: {targetType: "program_event_series"}, AuditProgramEventSeriesDeleted: {targetType: "program_event_series"}, AuditProgramEventSeriesUpdated: {targetType: "program_event_series"},
	AuditProgramEventTypeCreated: {targetType: "program_event_type"}, AuditProgramEventTypeDeleted: {targetType: "program_event_type"}, AuditProgramEventTypeUpdated: {targetType: "program_event_type"},
	AuditReleaseCreated: {targetType: "release"}, AuditReleaseDeleted: {targetType: "release"}, AuditReleaseUpdated: {targetType: "release", systemServices: []ServiceName{ServiceEditorCollab}},
	AuditSiteSettingsUpdated: {targetType: "site_settings", singletonTarget: "1"},
	AuditStyleCreated:        {targetType: "style"}, AuditStyleDeleted: {targetType: "style"}, AuditStyleUpdated: {targetType: "style"},
	AuditTagCreated: {targetType: "tag"}, AuditTagDeleted: {targetType: "tag"}, AuditTagUpdated: {targetType: "tag"},
	AuditTranslationProviderCreated: {targetType: "translation_provider"}, AuditTranslationProviderDeleted: {targetType: "translation_provider"}, AuditTranslationProviderUpdated: {targetType: "translation_provider"},
	AuditTranslationSettingsUpdated: {targetType: "translation_settings", singletonTarget: "1"},
	AuditWorkCreated:                {targetType: "work"}, AuditWorkDeleted: {targetType: "work"}, AuditWorkUpdated: {targetType: "work", systemServices: []ServiceName{ServiceEditorCollab}},
}

func validateAuditCatalogRecord(record AuditRecord) error {
	entry, ok := auditCatalog[record.Action]
	if !ok {
		return fmt.Errorf("unsupported audit action %s", record.Action)
	}
	if record.TargetType != entry.targetType || record.TargetID == "" {
		return fmt.Errorf("audit action %s requires target_type %s and target_id", record.Action, entry.targetType)
	}
	if entry.singletonTarget != "" && record.TargetID != entry.singletonTarget {
		return fmt.Errorf("audit action %s requires singleton target_id %s", record.Action, entry.singletonTarget)
	}
	if record.Kind == ActorKindAnonymous && !entry.allowAnonymous {
		return fmt.Errorf("domain audit actor cannot be anonymous")
	}
	if record.Kind == ActorKindSystem && !slices.Contains(entry.systemServices, ServiceName(record.Service)) {
		return fmt.Errorf("audit action %s cannot use a system actor", record.Action)
	}
	if handled, err := validateLocaleContentAuditVariant(record); handled || err != nil {
		return err
	}
	if handled, err := validateSourceLocaleAuditVariant(record); handled || err != nil {
		return err
	}
	if entry.validate != nil {
		return entry.validate(record)
	}
	return fmt.Errorf("audit action %s has no implemented variant validator", record.Action)
}

func validateExtendedAuditAttributes(record AuditRecord) error {
	if err := validateSortedUnique("changed_fields", record.ChangedFields, true); err != nil {
		return err
	}
	if err := validateOptionalAuditIdentifiers("file_ids", record.FileIDs); err != nil {
		return err
	}
	if err := validateOptionalAuditIdentifiers("item_ids", record.ItemIDs); err != nil {
		return err
	}
	if err := validateOptionalAuditIdentifiers("previous_item_ids", record.PreviousItemIDs); err != nil {
		return err
	}
	if err := validateOptionalAuditIdentifiers("post_ids", record.PostIDs); err != nil {
		return err
	}
	return validateOptionalAuditIdentifiers("tag_ids", record.TagIDs)
}

// A nil pointer means the attribute is absent. A non-nil empty slice is an
// intentional empty ordered set and must remain valid on the JSON wire.
func validateOptionalAuditIdentifiers(name string, values *[]string) error {
	if values == nil {
		return nil
	}
	for _, value := range *values {
		if !isAuditIdentifier(value) {
			return fmt.Errorf("%s contains an invalid identifier", name)
		}
	}
	return nil
}
