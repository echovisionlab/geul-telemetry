import { readFileSync } from "node:fs";

import { describe, expect, it } from "vitest";

import {
  actorForRecord,
  CANONICAL_SERVICE_NAMES,
  instrumentationName,
  parseServiceName,
  SERVICE_BACKEND,
} from "./actor.ts";

describe("actorForRecord", () => {
  it("projects only safe record identity fields", () => {
    expect(actorForRecord({ kind: "anonymous" })).toEqual({
      actor_kind: "anonymous",
    });
    expect(
      actorForRecord({
        kind: "member",
        sessionId: "session-1",
        identityId: "identity-1",
        memberId: "member-1",
      }),
    ).toEqual({ actor_kind: "member", actor_member_id: "member-1" });
    expect(
      actorForRecord({ kind: "system", serviceName: SERVICE_BACKEND }),
    ).toEqual({
      actor_kind: "system",
      actor_service: "geul-backend",
    });
  });

  it("rejects incomplete actor identities", () => {
    expect(() =>
      actorForRecord({
        kind: "member",
        sessionId: "session-1",
        identityId: "identity-1",
        memberId: "",
      }),
    ).toThrow("member actor requires memberId");
    expect(() =>
      actorForRecord({
        kind: "system",
        serviceName: "api" as typeof SERVICE_BACKEND,
      }),
    ).toThrow("unknown canonical service name");
  });

  it("matches the cross-language canonical service fixture", () => {
    const fixture = JSON.parse(
      readFileSync(
        new URL("../fixtures/service-identities.json", import.meta.url),
        "utf8",
      ),
    ) as string[];

    expect(CANONICAL_SERVICE_NAMES).toEqual(fixture);
    expect(fixture.map(parseServiceName)).toEqual(fixture);
    expect(instrumentationName(SERVICE_BACKEND, "http")).toBe(
      "geul-backend/http",
    );
    expect(() => parseServiceName("geul-kratos")).toThrow(
      "unknown canonical service name",
    );
  });
});
