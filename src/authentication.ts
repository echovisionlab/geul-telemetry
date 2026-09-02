import type { RecordActor } from "./actor.ts";
import {
  validateSecurityAccessRecord,
  type AuthenticationBlockReason,
  type AuthenticationFailureReason,
  type AuthenticationFlowKind,
  type AuthenticationMethod,
  type AuthenticationPrincipalState,
  type AuthorizationDenialReason,
  type Correlation,
  type SecurityAccessRecord,
} from "./records.ts";

export type SecurityAccessMetadata = Correlation &
  RecordActor & {
    readonly access_id: string;
    readonly occurred_at: string;
    readonly source_ip: string;
  };

export interface AuthenticationContext {
  readonly flow_kind?: AuthenticationFlowKind;
  readonly authentication_method?: AuthenticationMethod;
  readonly principal_state?: AuthenticationPrincipalState;
  readonly provider?: string;
}

export function buildAuthenticationSucceededRecord(
  metadata: SecurityAccessMetadata,
  authentication: Required<
    Pick<
      AuthenticationContext,
      "flow_kind" | "authentication_method" | "principal_state"
    >
  > &
    Pick<AuthenticationContext, "provider">,
): SecurityAccessRecord {
  return buildSecurityAccessRecord(metadata, {
    action: "authentication.succeeded",
    ...authentication,
  });
}

export function buildAuthenticationFailedRecord(
  metadata: SecurityAccessMetadata,
  authentication: Required<
    Pick<AuthenticationContext, "flow_kind" | "authentication_method">
  > &
    Pick<AuthenticationContext, "provider">,
  reason: AuthenticationFailureReason,
): SecurityAccessRecord {
  return buildSecurityAccessRecord(metadata, {
    action: "authentication.failed",
    ...authentication,
    reason,
  });
}

export function buildAuthenticationBlockedRecord(
  metadata: SecurityAccessMetadata,
  authentication: AuthenticationContext,
  reason: AuthenticationBlockReason,
): SecurityAccessRecord {
  return buildSecurityAccessRecord(metadata, {
    action: "authentication.blocked",
    ...authentication,
    reason,
  });
}

export function buildAuthorizationDeniedRecord(
  metadata: SecurityAccessMetadata,
  procedure: string,
  reason: AuthorizationDenialReason,
): SecurityAccessRecord {
  return buildSecurityAccessRecord(metadata, {
    action: "authorization.denied",
    attempted_action: procedure,
    permission: "procedure:invoke",
    reason,
  });
}

export function buildMemberCollectionAccessedRecord(
  metadata: SecurityAccessMetadata,
): SecurityAccessRecord {
  return buildPersonalDataAccessedRecord(
    metadata,
    "member_collection",
    "1",
    "member_administration",
  );
}

export function buildMemberAccessedRecord(
  metadata: SecurityAccessMetadata,
  memberId: string,
): SecurityAccessRecord {
  return buildPersonalDataAccessedRecord(
    metadata,
    "member",
    memberId,
    "member_administration",
  );
}

export function buildCampaignRecipientsAccessedRecord(
  metadata: SecurityAccessMetadata,
  campaignId: string,
): SecurityAccessRecord {
  return buildPersonalDataAccessedRecord(
    metadata,
    "campaign",
    campaignId,
    "campaign_recipients",
  );
}

export function buildFormSubmissionsAccessedRecord(
  metadata: SecurityAccessMetadata,
  formId: string,
): SecurityAccessRecord {
  return buildPersonalDataAccessedRecord(
    metadata,
    "form",
    formId,
    "form_submissions",
  );
}

export function buildFormSubmissionAccessedRecord(
  metadata: SecurityAccessMetadata,
  submissionId: string,
): SecurityAccessRecord {
  return buildPersonalDataAccessedRecord(
    metadata,
    "form_submission",
    submissionId,
    "form_submission",
  );
}

function buildPersonalDataAccessedRecord(
  metadata: SecurityAccessMetadata,
  subjectType: string,
  subjectId: string,
  dataCategory: string,
): SecurityAccessRecord {
  return buildSecurityAccessRecord(metadata, {
    action: "personal_data.accessed",
    subject_type: subjectType,
    subject_id: subjectId,
    access_kind: "read",
    data_category: dataCategory,
  });
}

export function authenticationMethodFromKratos(
  value: string,
): AuthenticationMethod {
  switch (value) {
    case "code":
      return "email_code";
    case "oidc":
      return "oidc";
    case "passkey":
      return "passkey";
    default:
      throw new TypeError("unsupported authentication method");
  }
}

function buildSecurityAccessRecord(
  metadata: SecurityAccessMetadata,
  attributes: Omit<
    SecurityAccessRecord,
    keyof SecurityAccessMetadata | keyof Correlation | keyof RecordActor
  >,
): SecurityAccessRecord {
  const record = { ...metadata, ...attributes } as SecurityAccessRecord;
  validateSecurityAccessRecord(record);
  return record;
}
