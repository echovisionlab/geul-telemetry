import { describe, expect, it } from "vitest";

import { SERVICE_TRANSCODER } from "./actor.ts";
import {
  activeRequestContext,
  createPropagatedRequestContext,
  createPublicRequestContext,
  isCanonicalSourceIp,
  isRequestId,
  runWithRequestContext,
  withActor,
} from "./request-context.ts";

const requestId = "018f47a2-8a3d-4e17-9d42-6f12c89b1234";

describe("request context", () => {
  it("creates a fresh anonymous public context", () => {
    const requestContext = createPublicRequestContext("192.0.2.4");
    expect(isRequestId(requestContext.requestId)).toBe(true);
    expect(requestContext.actor).toEqual({ kind: "anonymous" });
    expect(requestContext.sourceIp).toBe("192.0.2.4");
    expect(Object.isFrozen(requestContext)).toBe(true);
    expect(createPublicRequestContext().sourceIp).toBeUndefined();
  });

  it("validates public and propagated values", () => {
    expect(() => createPublicRequestContext("invalid")).toThrow("sourceIp");
    expect(isCanonicalSourceIp("2001:db8::4")).toBe(true);
    expect(isCanonicalSourceIp("2001:0db8::4")).toBe(false);
    expect(isCanonicalSourceIp("2001:DB8::4")).toBe(false);
    expect(isRequestId("BAD")).toBe(false);
    expect(() =>
      createPropagatedRequestContext("BAD", { kind: "anonymous" }),
    ).toThrow("UUIDv4");
    expect(
      createPropagatedRequestContext(requestId, {
        kind: "system",
        serviceName: SERVICE_TRANSCODER,
      }),
    ).toMatchObject({
      requestId,
      actor: { kind: "system", serviceName: SERVICE_TRANSCODER },
    });
  });

  it("keeps actor replacement immutable and scopes context", () => {
    const original = createPropagatedRequestContext(requestId, {
      kind: "anonymous",
    });
    const resolved = withActor(original, {
      kind: "member",
      sessionId: "session-1",
      identityId: "identity-1",
      memberId: "member-1",
    });
    expect(original.actor.kind).toBe("anonymous");
    expect(resolved.actor.kind).toBe("member");
    expect(activeRequestContext()).toBeUndefined();
    const value = runWithRequestContext(resolved, () => activeRequestContext());
    expect(value).toBe(resolved);
  });
});
