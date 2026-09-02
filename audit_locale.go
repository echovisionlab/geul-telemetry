package telemetry

import "regexp"

var auditLocaleCodePattern = regexp.MustCompile(`^[A-Za-z0-9]{1,8}(-[A-Za-z0-9]{1,8})*$`)

func isCanonicalAuditLocale(value string) bool {
	return len(value) <= 64 && auditLocaleCodePattern.MatchString(value)
}
