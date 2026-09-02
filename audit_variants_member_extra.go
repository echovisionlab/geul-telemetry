package telemetry

import (
	"fmt"
	"strings"
)

func init() {
	setAuditValidator(AuditMemberTagCreated, validateMemberTagRecord)
	setAuditValidator(AuditMemberTagDeleted, validateMemberTagRecord)
}

func validateMemberProfileOrPreference(record AuditRecord, requireOnly func(...string) error) error {
	profile := map[string]bool{"bio": true, "nickname": true, "social_links": true, "website": true}
	preference := map[string]bool{"cookie_consent": true, "preferred_locale": true}
	if len(record.ChangedFields) == 0 {
		return fmt.Errorf("member update requires changed_fields")
	}
	if err := validateSortedUnique("changed_fields", record.ChangedFields, true); err != nil {
		return err
	}
	allProfile, allPreference := true, true
	for _, field := range record.ChangedFields {
		allProfile = allProfile && profile[field]
		allPreference = allPreference && preference[field]
	}
	if allProfile {
		if containsAuditField(record.ChangedFields, "nickname") && (record.Nickname == "" || record.Nickname != trimAuditText(record.Nickname)) {
			return fmt.Errorf("member nickname requires a trimmed value")
		}
		if !containsAuditField(record.ChangedFields, "nickname") && record.Nickname != "" {
			return fmt.Errorf("member nickname requires changed_fields nickname")
		}
		return requireOnly("changed_fields", "nickname")
	}
	if allPreference {
		if containsAuditField(record.ChangedFields, "preferred_locale") && !isBoundedCode(record.PreferredLocale) {
			return fmt.Errorf("member preferred_locale requires a bounded locale")
		}
		if !containsAuditField(record.ChangedFields, "preferred_locale") && record.PreferredLocale != "" {
			return fmt.Errorf("member preferred_locale requires changed_fields preferred_locale")
		}
		if containsAuditField(record.ChangedFields, "cookie_consent") && !isAuditIdentifier(record.ConsentID) {
			return fmt.Errorf("member cookie_consent requires consent_id")
		}
		if !containsAuditField(record.ChangedFields, "cookie_consent") && record.ConsentID != "" {
			return fmt.Errorf("member consent_id requires changed_fields cookie_consent")
		}
		return requireOnly("changed_fields", "preferred_locale", "consent_id")
	}
	return fmt.Errorf("member.updated rejects changed_fields variant")
}

func trimAuditText(value string) string { return strings.TrimSpace(value) }
func validateMemberTagRecord(record AuditRecord) error {
	if record.TagName == "" || record.TagName != trimAuditText(record.TagName) {
		return fmt.Errorf("member tag requires trimmed tag_name")
	}
	if hasAuditAttributesExcept(record, "TagName") {
		return fmt.Errorf("member tag has unsupported attributes")
	}
	return nil
}
