package telemetry

import (
	"errors"
	"testing"
)

func TestSecurityAccessBuilders(t *testing.T) {
	t.Parallel()
	memberMetadata := SecurityAccessMetadata{
		AccessID:    "7a7a8fd4-1f69-4e9a-9dc2-2378926ff351",
		OccurredAt:  testOccurredAt,
		Correlation: Correlation{RequestID: testRequestID},
		RecordActor: RecordActor{Kind: ActorKindMember, MemberID: "member-1"},
		SourceIP:    "192.0.2.4",
	}
	authentication := AuthenticationContext{
		FlowKind:             AuthenticationFlowLogin,
		AuthenticationMethod: AuthenticationMethodPasskey,
		PrincipalState:       AuthenticationPrincipalActive,
	}
	succeeded, err := NewAuthenticationSucceededRecord(memberMetadata, authentication)
	if err != nil || succeeded.Action != SecurityAuthenticationSucceeded {
		t.Fatalf("succeeded = %#v, %v", succeeded, err)
	}

	anonymousMetadata := memberMetadata
	anonymousMetadata.AccessID = "0b6bcad2-c90d-49e9-bec7-f9a4ba6b2894"
	anonymousMetadata.RecordActor = RecordActor{Kind: ActorKindAnonymous}
	failed, err := NewAuthenticationFailedRecord(anonymousMetadata, AuthenticationContext{
		FlowKind: AuthenticationFlowLogin, AuthenticationMethod: AuthenticationMethodOIDC, Provider: "google",
	}, AuthenticationFailureProviderDenied)
	if err != nil || failed.Action != SecurityAuthenticationFailed {
		t.Fatalf("failed = %#v, %v", failed, err)
	}
	blockedAccount, err := NewAuthenticationFailedRecord(anonymousMetadata, AuthenticationContext{
		FlowKind: AuthenticationFlowLogin, AuthenticationMethod: AuthenticationMethodEmailCode,
	}, AuthenticationFailureAccountBlocked)
	if err != nil || blockedAccount.Reason != "account_blocked" {
		t.Fatalf("blocked account failure = %#v, %v", blockedAccount, err)
	}
	blocked, err := NewAuthenticationBlockedRecord(anonymousMetadata, AuthenticationContext{}, AuthenticationBlockRateLimited)
	if err != nil || blocked.FlowKind != "" || blocked.AuthenticationMethod != "" {
		t.Fatalf("blocked = %#v, %v", blocked, err)
	}
	denied, err := NewAuthorizationDeniedRecord(anonymousMetadata, "/geul.api.v1.PostService/UpdatePost", AuthorizationDeniedPermissionDenied)
	if err != nil || denied.Action != SecurityAuthorizationDenied {
		t.Fatalf("denied = %#v, %v", denied, err)
	}
	denied.Permission = "not-a-permission"
	if err := denied.Validate(); err == nil {
		t.Fatal("authorization denial with unknown permission accepted")
	}
	accessed, err := NewMemberAccessedRecord(memberMetadata, "2a7a8fd4-1f69-4e9a-9dc2-2378926ff351")
	if err != nil || accessed.Action != SecurityPersonalDataAccessed {
		t.Fatalf("accessed = %#v, %v", accessed, err)
	}
	for name, build := range map[string]func() (SecurityAccessRecord, error){
		"member collection": func() (SecurityAccessRecord, error) { return NewMemberCollectionAccessedRecord(memberMetadata) },
		"campaign recipients": func() (SecurityAccessRecord, error) {
			return NewCampaignRecipientsAccessedRecord(memberMetadata, "3a7a8fd4-1f69-4e9a-9dc2-2378926ff351")
		},
		"form submissions": func() (SecurityAccessRecord, error) {
			return NewFormSubmissionsAccessedRecord(memberMetadata, "4a7a8fd4-1f69-4e9a-9dc2-2378926ff351")
		},
		"form submission": func() (SecurityAccessRecord, error) {
			return NewFormSubmissionAccessedRecord(memberMetadata, "5a7a8fd4-1f69-4e9a-9dc2-2378926ff351")
		},
	} {
		record, buildErr := build()
		if buildErr != nil || record.Action != SecurityPersonalDataAccessed {
			t.Fatalf("%s = %#v, %v", name, record, buildErr)
		}
	}

	if _, err := NewAuthenticationSucceededRecord(anonymousMetadata, authentication); err == nil {
		t.Fatal("anonymous authentication success accepted")
	}
	missingState := authentication
	missingState.PrincipalState = ""
	if _, err := NewAuthenticationSucceededRecord(memberMetadata, missingState); err == nil {
		t.Fatal("authentication success without principal state accepted")
	}
}

func TestAuthenticationMethodFromKratos(t *testing.T) {
	t.Parallel()
	for input, expected := range map[string]AuthenticationMethod{
		"code": AuthenticationMethodEmailCode, "oidc": AuthenticationMethodOIDC, "passkey": AuthenticationMethodPasskey,
	} {
		method, err := AuthenticationMethodFromKratos(input)
		if err != nil || method != expected {
			t.Fatalf("AuthenticationMethodFromKratos(%q) = %q, %v", input, method, err)
		}
	}
	if _, err := AuthenticationMethodFromKratos("password"); !errors.Is(err, ErrUnsupportedAuthenticationMethod) {
		t.Fatalf("unsupported method error = %v", err)
	}
}

func TestSecurityCatalogRejectsMalformedSubjectsAndProcedures(t *testing.T) {
	t.Parallel()
	metadata := SecurityAccessMetadata{
		AccessID:    "018f47a2-8a3d-4e17-9d42-6f12c89b1234",
		OccurredAt:  testOccurredAt,
		Correlation: Correlation{RequestID: testRequestID},
		RecordActor: RecordActor{Kind: ActorKindMember, MemberID: "member-1"},
		SourceIP:    "192.0.2.1",
	}
	if _, err := NewMemberAccessedRecord(metadata, "not-a-uuid"); err == nil {
		t.Fatal("malformed personal data subject accepted")
	}
	for _, procedure := range []string{"//Method", "/bad-service/Method"} {
		if _, err := NewAuthorizationDeniedRecord(metadata, procedure, AuthorizationDeniedPermissionDenied); err == nil {
			t.Fatalf("malformed Connect procedure %q accepted", procedure)
		}
	}
}
