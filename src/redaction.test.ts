import { readFileSync } from "node:fs";

import { describe, expect, it } from "vitest";

import {
  isForbiddenKey,
  normalizeKey,
  normalizeLogAttributes,
  stableErrorType,
} from "./redaction.ts";

describe("redaction", () => {
  it("matches the shared Go and TypeScript key contract", () => {
    const fixture = JSON.parse(
      readFileSync(
        new URL("../fixtures/log-redaction.json", import.meta.url),
        "utf8",
      ),
    ) as {
      forbidden: readonly { key: string; normalized: string }[];
      allowed: readonly { key: string; normalized: string }[];
    };
    for (const entry of fixture.forbidden) {
      expect(normalizeKey(entry.key)).toBe(entry.normalized);
      expect(isForbiddenKey(entry.key)).toBe(true);
    }
    for (const entry of fixture.allowed) {
      expect(normalizeKey(entry.key)).toBe(entry.normalized);
      expect(isForbiddenKey(entry.key)).toBe(false);
    }
  });

  it("normalizes structured field names", () => {
    expect(normalizeKey(" HTTPStatus-Code ")).toBe("httpstatus_code");
    expect(normalizeKey("actorMemberID")).toBe("actor_member_id");
  });

  it("rejects exact, prefixed, and suffixed sensitive keys", () => {
    expect(isForbiddenKey("email")).toBe(true);
    expect(isForbiddenKey("access_token")).toBe(true);
    expect(isForbiddenKey("secret_value")).toBe(true);
    expect(isForbiddenKey("identity_id")).toBe(true);
    expect(isForbiddenKey("session_id")).toBe(true);
    expect(isForbiddenKey("flow_id")).toBe(true);
    expect(isForbiddenKey("error_reason")).toBe(true);
    expect(isForbiddenKey("member_id")).toBe(true);
    expect(isForbiddenKey("display_name")).toBe(true);
    expect(isForbiddenKey("member_name")).toBe(true);
    expect(isForbiddenKey("requester_nickname")).toBe(true);
    expect(isForbiddenKey("user_id")).toBe(true);
    expect(isForbiddenKey("component")).toBe(false);
  });

  it("reduces errors to a stable type", () => {
    expect(stableErrorType(new TypeError("secret message"))).toBe("type_error");
    expect(stableErrorType("raw error")).toBe("reported_error");
  });

  it("uses the stable Error.name when a production bundle minifies the constructor", () => {
    class t extends Error {
      override name = "ConnectError";
    }

    expect(stableErrorType(new t("private provider detail"))).toBe(
      "connect_error",
    );
  });

  it("falls back to a custom constructor name when Error.name is generic", () => {
    class ProviderFailure extends Error {}

    expect(
      stableErrorType(new ProviderFailure("private provider detail")),
    ).toBe("provider_failure");

    const anonymousError = new Error("private provider detail");
    Object.defineProperty(anonymousError, "constructor", {
      value: { name: "" },
    });
    expect(stableErrorType(anonymousError)).toBe("reported_error");
  });

  it("normalizes primitive diagnostics and drops sensitive or untyped data", () => {
    const attributes = normalizeLogAttributes({
      commandId: "command-1",
      recipient: "person@example.com",
      displayName: "Private Display Name",
      requesterNickname: "Private Nickname",
      error: new TypeError("raw provider secret"),
      providerError: "raw provider secret",
      translation: {
        api: { key: "api secret" },
        provider: {
          document: { id: "document secret" },
          response: { body: "response secret" },
        },
      },
      providerType: "deepl",
      modelCatalogValue: "catalog-model",
      operation: "download",
      httpStatusClass: "5xx",
      failureReason: "provider_unavailable",
      exceptionType: "type_error",
      retries: 2,
    });
    expect(attributes).toEqual({
      command_id: "command-1",
      error_type: "type_error",
      provider_type: "deepl",
      model_catalog_value: "catalog-model",
      operation: "download",
      http_status_class: "5xx",
      failure_reason: "provider_unavailable",
      exception_type: "type_error",
      retries: 2,
    });
    expect(JSON.stringify(attributes)).not.toContain("secret");
    expect(normalizeLogAttributes()).toEqual({});
  });
});
