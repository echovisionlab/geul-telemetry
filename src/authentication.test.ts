import { describe, expect, it } from "vitest";

import {
  authenticationMethodFromKratos,
  buildAuthenticationBlockedRecord,
  buildAuthenticationFailedRecord,
  buildAuthenticationSucceededRecord,
  buildAuthorizationDeniedRecord,
  buildCampaignRecipientsAccessedRecord,
  buildFormSubmissionAccessedRecord,
  buildFormSubmissionsAccessedRecord,
  buildMemberAccessedRecord,
  buildMemberCollectionAccessedRecord,
  type SecurityAccessMetadata,
} from "./authentication.ts";

const memberMetadata: SecurityAccessMetadata = {
  access_id: "7a7a8fd4-1f69-4e9a-9dc2-2378926ff351",
  occurred_at: "2026-08-09T03:04:05Z",
  request_id: "018f47a2-8a3d-4e17-9d42-6f12c89b1234",
  actor_kind: "member",
  actor_member_id: "member-1",
  source_ip: "192.0.2.4",
};

describe("security access builders", () => {
  it("builds only validated action shapes", () => {
    expect(
      buildAuthenticationSucceededRecord(memberMetadata, {
        flow_kind: "login",
        authentication_method: "passkey",
        principal_state: "active",
      }).action,
    ).toBe("authentication.succeeded");

    const anonymousMetadata: SecurityAccessMetadata = {
      access_id: "0b6bcad2-c90d-49e9-bec7-f9a4ba6b2894",
      occurred_at: memberMetadata.occurred_at,
      request_id: memberMetadata.request_id,
      actor_kind: "anonymous",
      source_ip: memberMetadata.source_ip,
    };
    expect(
      buildAuthenticationFailedRecord(
        anonymousMetadata,
        {
          flow_kind: "login",
          authentication_method: "oidc",
          provider: "google",
        },
        "provider_denied",
      ).action,
    ).toBe("authentication.failed");
    expect(
      buildAuthenticationBlockedRecord(anonymousMetadata, {}, "rate_limited")
        .flow_kind,
    ).toBeUndefined();
    expect(
      buildAuthorizationDeniedRecord(
        anonymousMetadata,
        "/geul.api.v1.PostService/UpdatePost",
        "permission_denied",
      ).action,
    ).toBe("authorization.denied");
    expect(
      buildMemberAccessedRecord(
        memberMetadata,
        "2a7a8fd4-1f69-4e9a-9dc2-2378926ff351",
      ).action,
    ).toBe("personal_data.accessed");
    for (const record of [
      buildMemberCollectionAccessedRecord(memberMetadata),
      buildCampaignRecipientsAccessedRecord(
        memberMetadata,
        "3a7a8fd4-1f69-4e9a-9dc2-2378926ff351",
      ),
      buildFormSubmissionsAccessedRecord(
        memberMetadata,
        "4a7a8fd4-1f69-4e9a-9dc2-2378926ff351",
      ),
      buildFormSubmissionAccessedRecord(
        memberMetadata,
        "5a7a8fd4-1f69-4e9a-9dc2-2378926ff351",
      ),
    ]) {
      expect(record.action).toBe("personal_data.accessed");
    }

    expect(() =>
      buildAuthenticationSucceededRecord(anonymousMetadata, {
        flow_kind: "login",
        authentication_method: "passkey",
        principal_state: "active",
      }),
    ).toThrow("member actor");
  });

  it("maps only pinned Kratos authentication methods", () => {
    expect(authenticationMethodFromKratos("code")).toBe("email_code");
    expect(authenticationMethodFromKratos("oidc")).toBe("oidc");
    expect(authenticationMethodFromKratos("passkey")).toBe("passkey");
    expect(() => authenticationMethodFromKratos("password")).toThrow(
      "unsupported authentication method",
    );
  });
});
