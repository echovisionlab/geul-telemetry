package telemetry

import (
	"errors"
	"time"
)

var ErrUnsupportedAuthenticationMethod = errors.New("unsupported authentication method")

// SecurityAccessMetadata contains the fields shared by all durable Security
// Access records. The owning service supplies the durable record ID and fact
// time so builders remain deterministic and transaction-friendly.
type SecurityAccessMetadata struct {
	AccessID   string
	OccurredAt time.Time
	Correlation
	RecordActor
	SourceIP string
}

type AuthenticationContext struct {
	FlowKind             AuthenticationFlowKind
	AuthenticationMethod AuthenticationMethod
	PrincipalState       AuthenticationPrincipalState
	Provider             string
}

func NewAuthenticationSucceededRecord(metadata SecurityAccessMetadata, authentication AuthenticationContext) (SecurityAccessRecord, error) {
	return newValidatedSecurityAccessRecord(metadata, SecurityAccessRecord{
		Action:               SecurityAuthenticationSucceeded,
		FlowKind:             authentication.FlowKind,
		AuthenticationMethod: authentication.AuthenticationMethod,
		PrincipalState:       authentication.PrincipalState,
		Provider:             authentication.Provider,
	})
}

func NewAuthenticationFailedRecord(metadata SecurityAccessMetadata, authentication AuthenticationContext, reason AuthenticationFailureReason) (SecurityAccessRecord, error) {
	return newValidatedSecurityAccessRecord(metadata, SecurityAccessRecord{
		Action:               SecurityAuthenticationFailed,
		FlowKind:             authentication.FlowKind,
		AuthenticationMethod: authentication.AuthenticationMethod,
		Provider:             authentication.Provider,
		Reason:               string(reason),
	})
}

func NewAuthenticationBlockedRecord(metadata SecurityAccessMetadata, authentication AuthenticationContext, reason AuthenticationBlockReason) (SecurityAccessRecord, error) {
	return newValidatedSecurityAccessRecord(metadata, SecurityAccessRecord{
		Action:               SecurityAuthenticationBlocked,
		FlowKind:             authentication.FlowKind,
		AuthenticationMethod: authentication.AuthenticationMethod,
		Provider:             authentication.Provider,
		Reason:               string(reason),
	})
}

func NewAuthorizationDeniedRecord(metadata SecurityAccessMetadata, procedure string, reason AuthorizationDenialReason) (SecurityAccessRecord, error) {
	return newValidatedSecurityAccessRecord(metadata, SecurityAccessRecord{
		Action:          SecurityAuthorizationDenied,
		AttemptedAction: procedure,
		Permission:      AuthorizationProcedureInvokePermission,
		Reason:          string(reason),
	})
}

func NewMemberCollectionAccessedRecord(metadata SecurityAccessMetadata) (SecurityAccessRecord, error) {
	return newPersonalDataAccessedRecord(metadata, "member_collection", "1", "member_administration")
}

func NewMemberAccessedRecord(metadata SecurityAccessMetadata, memberID string) (SecurityAccessRecord, error) {
	return newPersonalDataAccessedRecord(metadata, "member", memberID, "member_administration")
}

func NewCampaignRecipientsAccessedRecord(metadata SecurityAccessMetadata, campaignID string) (SecurityAccessRecord, error) {
	return newPersonalDataAccessedRecord(metadata, "campaign", campaignID, "campaign_recipients")
}

func NewFormSubmissionsAccessedRecord(metadata SecurityAccessMetadata, formID string) (SecurityAccessRecord, error) {
	return newPersonalDataAccessedRecord(metadata, "form", formID, "form_submissions")
}

func NewFormSubmissionAccessedRecord(metadata SecurityAccessMetadata, submissionID string) (SecurityAccessRecord, error) {
	return newPersonalDataAccessedRecord(metadata, "form_submission", submissionID, "form_submission")
}

func newPersonalDataAccessedRecord(metadata SecurityAccessMetadata, subjectType, subjectID, dataCategory string) (SecurityAccessRecord, error) {
	return newValidatedSecurityAccessRecord(metadata, SecurityAccessRecord{
		Action:       SecurityPersonalDataAccessed,
		SubjectType:  subjectType,
		SubjectID:    subjectID,
		AccessKind:   PersonalDataAccessRead,
		DataCategory: dataCategory,
	})
}

func AuthenticationMethodFromKratos(value string) (AuthenticationMethod, error) {
	var method AuthenticationMethod
	switch value {
	case "code":
		method = AuthenticationMethodEmailCode
	case "oidc":
		method = AuthenticationMethodOIDC
	case "passkey":
		method = AuthenticationMethodPasskey
	default:
		return "", ErrUnsupportedAuthenticationMethod
	}
	return method, nil
}

func newValidatedSecurityAccessRecord(metadata SecurityAccessMetadata, attributes SecurityAccessRecord) (SecurityAccessRecord, error) {
	attributes.AccessID = metadata.AccessID
	attributes.OccurredAt = metadata.OccurredAt
	attributes.Correlation = metadata.Correlation
	attributes.RecordActor = metadata.RecordActor
	attributes.SourceIP = metadata.SourceIP
	if err := attributes.Validate(); err != nil {
		return SecurityAccessRecord{}, err
	}
	return attributes, nil
}
