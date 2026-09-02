import {
  trace,
  type TextMapGetter,
  type TextMapSetter,
} from "@opentelemetry/api";
import { describe, expect, it, vi } from "vitest";

import { SERVICE_TRANSCODER } from "./actor.ts";
import {
  correlationFromActiveContext,
  extractCorrelation,
  injectCorrelation,
} from "./otel.ts";
import {
  createPropagatedRequestContext,
  runWithRequestContext,
} from "./request-context.ts";

const requestId = "018f47a2-8a3d-4e17-9d42-6f12c89b1234";
const setter: TextMapSetter<Map<string, string>> = {
  set(carrier, key, value) {
    carrier.set(key, value);
  },
};
const getter: TextMapGetter<Map<string, string>> = {
  get(carrier, key) {
    return carrier.get(key);
  },
  keys(carrier) {
    return [...carrier.keys()];
  },
};

describe("OpenTelemetry correlation", () => {
  it("reads active request and span identifiers", () => {
    const requestContext = createPropagatedRequestContext(requestId, {
      kind: "anonymous",
    });
    const spanContext = {
      traceId: "4bf92f3577b34da6a3ce929d0e0e4736",
      spanId: "00f067aa0ba902b7",
      traceFlags: 1,
    };
    const spanContextSpy = vi
      .spyOn(trace, "getSpanContext")
      .mockReturnValue(spanContext);
    runWithRequestContext(requestContext, () => {
      expect(correlationFromActiveContext()).toEqual({
        request_id: requestId,
        trace_id: spanContext.traceId,
        span_id: spanContext.spanId,
      });
    });
    spanContextSpy.mockRestore();
    expect(correlationFromActiveContext()).toEqual({});
  });

  it("injects and extracts a trusted request ID", () => {
    const requestContext = createPropagatedRequestContext(requestId, {
      kind: "anonymous",
    });
    const carrier = new Map<string, string>();
    injectCorrelation(carrier, setter, requestContext);
    expect(carrier.get("x-request-id")).toBe(requestId);
    expect(
      extractCorrelation(carrier, getter, {
        kind: "system",
        serviceName: SERVICE_TRANSCODER,
      }).requestContext,
    ).toMatchObject({
      requestId,
      actor: { kind: "system", serviceName: SERVICE_TRANSCODER },
    });
  });

  it("ignores missing, invalid, and repeated request ID values", () => {
    const missing = extractCorrelation(new Map(), getter, {
      kind: "anonymous",
    });
    expect(missing.requestContext).toBeUndefined();
    expect(
      extractCorrelation(new Map([["x-request-id", "bad"]]), getter, {
        kind: "anonymous",
      }).requestContext,
    ).toBeUndefined();

    const arrayGetter: TextMapGetter<Map<string, string>> = {
      ...getter,
      get: () => [requestId, "ignored"],
    };
    expect(
      extractCorrelation(new Map(), arrayGetter, { kind: "anonymous" })
        .requestContext?.requestId,
    ).toBe(requestId);

    const carrier = new Map<string, string>();
    injectCorrelation(carrier, setter, undefined);
    expect(carrier.size).toBe(0);
  });
});
