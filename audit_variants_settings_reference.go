package telemetry

import (
	"fmt"
	"reflect"
	"strings"
)

func init() {
	registerSettingsReferenceValidators()
}

func registerSettingsReferenceValidators() {
	noAttributes := []AuditAction{
		AuditMapThemeCreated, AuditMapThemeDeleted, AuditFileCreated, AuditFileDeleted, AuditFileFolderCreated, AuditFileFolderDeleted,
		AuditCategoryCreated, AuditCategoryDeleted, AuditTagCreated, AuditTagDeleted, AuditGenreCreated, AuditGenreDeleted,
		AuditStyleCreated, AuditStyleDeleted, AuditFormatCreated, AuditFormatDeleted, AuditClientCreated, AuditClientDeleted,
		AuditMapPlaceCreated, AuditMapPlaceDeleted, AuditAudienceSegmentCreated, AuditMenuCreated, AuditMenuDeleted,
	}
	for _, action := range noAttributes {
		setAuditValidator(action, validateNoAuditAttributes)
	}
	setAuditValidator(AuditSiteSettingsUpdated, validateChangedOnly(siteSettingsFields...))
	setAuditValidator(AuditMapThemeUpdated, func(record AuditRecord) error {
		if record.Kind != ActorKindMember {
			return fmt.Errorf("map theme content mutation requires member actor")
		}
		return validateChangedOnly("content")(record)
	})
	setAuditValidator(AuditLegalPolicyCreated, validateLegalIdentityOnly)
	setAuditValidator(AuditLegalPolicyDeleted, validateLegalIdentityOnly)
	setAuditValidator(AuditLegalPolicyUpdated, validateLegalPolicyUpdate)
	setAuditValidator(AuditFileUpdated, validateFileUpdate)
	setAuditValidator(AuditFileFolderUpdated, validateFolderUpdate)
	setAuditValidator(AuditCategoryUpdated, validateChangedOnly("description", "name", "slug"))
	setAuditValidator(AuditTagUpdated, validateChangedOnly("name", "slug"))
	setAuditValidator(AuditGenreUpdated, validateChangedOnly("description", "name", "slug"))
	setAuditValidator(AuditStyleUpdated, validateChangedOnly("description", "name", "slug"))
	setAuditValidator(AuditFormatUpdated, validateChangedOnly("name", "slug"))
	setAuditValidator(AuditClientUpdated, validateClientUpdate)
	setAuditValidator(AuditMapPlaceUpdated, validateMapPlaceUpdate)
	setAuditValidator(AuditAudienceSegmentUpdated, validateAudienceUpdate)
	setAuditValidator(AuditMenuUpdated, validateChangedOnly("items", "name"))
}

func setAuditValidator(action AuditAction, validator auditRecordValidator) {
	entry := auditCatalog[action]
	entry.validate = validator
	auditCatalog[action] = entry
}

func validateNoAuditAttributes(record AuditRecord) error {
	if hasAuditAttributes(record) {
		return fmt.Errorf("audit action %s does not allow attributes", record.Action)
	}
	return nil
}

func validateChangedOnly(allowed ...string) auditRecordValidator {
	return func(record AuditRecord) error {
		if err := validateChangedSubset(record.ChangedFields, allowed...); err != nil {
			return err
		}
		if hasAuditAttributesExcept(record, "ChangedFields") {
			return fmt.Errorf("audit action %s does not allow typed attributes", record.Action)
		}
		return nil
	}
}

func validateChangedSubset(fields []string, allowed ...string) error {
	if len(fields) == 0 {
		return fmt.Errorf("changed_fields is required")
	}
	if err := validateSortedUnique("changed_fields", fields, true); err != nil {
		return err
	}
	allowedSet := map[string]struct{}{}
	for _, field := range allowed {
		allowedSet[field] = struct{}{}
	}
	for _, field := range fields {
		if _, ok := allowedSet[field]; !ok {
			return fmt.Errorf("changed_fields rejects %s", field)
		}
	}
	return nil
}

func hasAuditAttributes(record AuditRecord) bool { return hasAuditAttributesExcept(record) }
func hasAuditAttributesExcept(record AuditRecord, except ...string) bool {
	present := auditAttributePresence(record)
	excluded := map[string]struct{}{}
	for _, name := range except {
		excluded[name] = struct{}{}
	}
	for name := range present {
		if _, ok := excluded[name]; !ok {
			return true
		}
	}
	return false
}

// auditAttributePresence derives presence from every typed AuditRecord field.
// Structural metadata is excluded; a future typed field is therefore rejected
// by attribute-boundary validators until an explicit variant allows it.
func auditAttributePresence(record AuditRecord) map[string]bool {
	present := map[string]bool{}
	v := reflect.ValueOf(record)
	t := v.Type()
	for i := 0; i < v.NumField(); i++ {
		name := t.Field(i).Name
		switch name {
		case "AuditID", "OccurredAt", "Action", "Correlation", "RecordActor", "TargetType", "TargetID":
			continue
		}
		if !v.Field(i).IsZero() {
			present[name] = true
		}
	}
	return present
}

func validateOnlyAuditAttributes(record AuditRecord, allowed ...string) error {
	allowedSet := make(map[string]struct{}, len(allowed))
	for _, name := range allowed {
		allowedSet[name] = struct{}{}
	}
	t := reflect.TypeOf(AuditRecord{})
	for i := 0; i < t.NumField(); i++ {
		field := t.Field(i)
		jsonName, _, _ := strings.Cut(field.Tag.Get("json"), ",")
		if _, ok := allowedSet[jsonName]; ok {
			allowedSet[field.Name] = struct{}{}
		}
	}
	for name := range auditAttributePresence(record) {
		if _, ok := allowedSet[name]; !ok {
			return fmt.Errorf("audit action %s does not allow %s", record.Action, name)
		}
	}
	return nil
}

func validateLegalIdentityOnly(record AuditRecord) error {
	if !validPolicyIdentity(record) {
		return fmt.Errorf("legal policy requires policy_type and version_number")
	}
	if hasAuditAttributesExcept(record, "PolicyType", "VersionNumber") {
		return fmt.Errorf("legal policy has unsupported attributes")
	}
	return nil
}
func validPolicyIdentity(record AuditRecord) bool {
	return (record.PolicyType == AuditPolicyTypeTerms || record.PolicyType == AuditPolicyTypePrivacy) && record.VersionNumber != nil && *record.VersionNumber > 0
}
func validateLegalPolicyUpdate(record AuditRecord) error {
	if !validPolicyIdentity(record) {
		return fmt.Errorf("legal policy requires policy_type and version_number")
	}
	if len(record.ChangedFields) == 1 && record.ChangedFields[0] == "share_links" {
		if record.Kind != ActorKindMember {
			return fmt.Errorf("legal policy share links require member actor")
		}
		if (record.ItemOperation != AuditItemOperationCreated && record.ItemOperation != AuditItemOperationDeleted) || !isAuditIdentifier(record.ItemID) {
			return fmt.Errorf("legal policy share_links requires item operation and item_id")
		}
		if hasAuditAttributesExcept(record, "ChangedFields", "PolicyType", "VersionNumber", "ItemOperation", "ItemID") {
			return fmt.Errorf("legal policy share_links has unsupported attributes")
		}
		return nil
	}
	if record.Kind != ActorKindMember && (record.Kind != ActorKindSystem || record.Service != string(ServiceBackend)) {
		return fmt.Errorf("legal policy lifecycle requires member or geul-backend system actor")
	}
	if err := validateChangedSubset(record.ChangedFields, "effective_at", "status"); err != nil {
		return err
	}
	hasEffectiveAt := containsAuditField(record.ChangedFields, "effective_at")
	if hasEffectiveAt != (record.EffectiveAt != nil) {
		return fmt.Errorf("legal policy lifecycle requires effective_at iff changed_fields includes it")
	}
	if record.EffectiveAt != nil {
		if err := validateRecordTime(*record.EffectiveAt); err != nil {
			return fmt.Errorf("legal policy lifecycle effective_at: %w", err)
		}
	}
	validTransition := (record.PreviousState == AuditStateDraft && record.NewState == AuditStateScheduled) ||
		(record.PreviousState == AuditStateScheduled && record.NewState == AuditStateDraft) ||
		(record.PreviousState == AuditStateDraft && record.NewState == AuditStateActive) ||
		(record.PreviousState == AuditStateScheduled && record.NewState == AuditStateActive) ||
		(record.PreviousState == AuditStateActive && record.NewState == AuditStateArchived)
	if !validTransition {
		return fmt.Errorf("legal policy lifecycle rejects transition %s to %s", record.PreviousState, record.NewState)
	}
	if hasAuditAttributesExcept(record, "ChangedFields", "PolicyType", "VersionNumber", "PreviousState", "NewState", "EffectiveAt") {
		return fmt.Errorf("legal policy lifecycle has unsupported attributes")
	}
	return nil
}

func validateFileUpdate(record AuditRecord) error {
	if err := validateChangedSubset(record.ChangedFields, "file_name", "folder_id"); err != nil {
		return err
	}
	hasMove := containsAuditField(record.ChangedFields, "folder_id")
	if hasMove && record.PreviousParentID == record.NewParentID {
		return fmt.Errorf("file folder_id requires a distinct parent transition")
	}
	if !hasMove && (record.PreviousParentID != "" || record.NewParentID != "") {
		return fmt.Errorf("file parent IDs require changed_fields folder_id")
	}
	if hasAuditAttributesExcept(record, "ChangedFields", "PreviousParentID", "NewParentID") {
		return fmt.Errorf("file.updated has unsupported attributes")
	}
	return nil
}
func validateFolderUpdate(record AuditRecord) error {
	if err := validateChangedSubset(record.ChangedFields, "name", "parent_id"); err != nil {
		return err
	}
	if containsAuditField(record.ChangedFields, "parent_id") && record.PreviousParentID == record.NewParentID {
		return fmt.Errorf("file folder parent_id requires a distinct parent transition")
	}
	if !containsAuditField(record.ChangedFields, "parent_id") && (record.PreviousParentID != "" || record.NewParentID != "") {
		return fmt.Errorf("file folder parent IDs require changed_fields parent_id")
	}
	if hasAuditAttributesExcept(record, "ChangedFields", "PreviousParentID", "NewParentID") {
		return fmt.Errorf("file_folder.updated has unsupported attributes")
	}
	return nil
}

func containsAuditField(fields []string, want string) bool {
	for _, field := range fields {
		if field == want {
			return true
		}
	}
	return false
}
func validateClientUpdate(record AuditRecord) error {
	if len(record.ChangedFields) == 1 && record.ChangedFields[0] == "logo" {
		return validateFileBinding(record, "logo", true)
	}
	return validateChangedOnly("name", "website")(record)
}
func validateMapPlaceUpdate(record AuditRecord) error {
	if len(record.ChangedFields) == 1 && record.ChangedFields[0] == "image" {
		return validateFileBinding(record, "image", false)
	}
	return validateChangedOnly("address", "address_components", "google_place_id", "lat", "lng", "name")(record)
}
func validateAudienceUpdate(record AuditRecord) error {
	if len(record.ChangedFields) == 1 && record.ChangedFields[0] == "status" {
		if !((record.PreviousState == AuditStateActive && record.NewState == AuditStateArchived) || (record.PreviousState == AuditStateArchived && record.NewState == AuditStateActive)) {
			return fmt.Errorf("audience status requires active/archived transition")
		}
		if hasAuditAttributesExcept(record, "ChangedFields", "PreviousState", "NewState") {
			return fmt.Errorf("audience status has unsupported attributes")
		}
		return nil
	}
	return validateChangedOnly("account_roles", "created_after", "created_before", "description", "exclude_member_ids", "member_tag_ids", "name", "segment_type")(record)
}

func validateFileBinding(record AuditRecord, field string, requiresSlot bool) error {
	if record.CollectionOperation != AuditCollectionOperationAdded && record.CollectionOperation != AuditCollectionOperationRemoved {
		return fmt.Errorf("%s requires collection operation", field)
	}
	if !isAuditIdentifier(record.FileID) {
		return fmt.Errorf("%s requires file_id", field)
	}
	if requiresSlot && record.AssetSlot != AuditAssetSlotLight && record.AssetSlot != AuditAssetSlotDark {
		return fmt.Errorf("%s requires asset_slot", field)
	}
	allowed := []string{"ChangedFields", "CollectionOperation", "FileID"}
	if requiresSlot {
		allowed = append(allowed, "AssetSlot")
	}
	if hasAuditAttributesExcept(record, allowed...) {
		return fmt.Errorf("%s has unsupported attributes", field)
	}
	return nil
}

var siteSettingsFields = []string{
	"company_address", "company_name", "default_comments_enabled", "default_map_theme_id", "favicon_file_id",
	"google_analytics_id", "homepage_page_id", "legal_email", "loader_file_ids", "logo_dark_file_id",
	"logo_email_file_id", "logo_light_file_id", "menu_avatar_dropdown_id", "menu_footer_id", "menu_header_id",
	"menu_secondary_id", "meta_description", "og_image_config", "primary_color", "privacy_email",
	"privacy_og_background_file_id", "site_og_background_file_id", "site_title", "social_links", "support_email",
	"tax_id", "terms_og_background_file_id",
}
