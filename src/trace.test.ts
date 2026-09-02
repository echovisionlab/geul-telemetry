import { trace } from "@opentelemetry/api";
import { afterEach, describe, expect, it, vi } from "vitest";

import { traceCorrelationFromActiveContext } from "./trace.ts";

describe("traceCorrelationFromActiveContext", () => {
  afterEach(() => {
    vi.restoreAllMocks();
  });

  it("returns valid active trace and span identifiers", () => {
    vi.spyOn(trace, "getSpanContext").mockReturnValue({
      traceId: "4bf92f3577b34da6a3ce929d0e0e4736",
      spanId: "00f067aa0ba902b7",
      traceFlags: 1,
    });

    expect(traceCorrelationFromActiveContext()).toEqual({
      trace_id: "4bf92f3577b34da6a3ce929d0e0e4736",
      span_id: "00f067aa0ba902b7",
    });
  });

  it("drops missing and invalid span contexts", () => {
    vi.spyOn(trace, "getSpanContext").mockReturnValue(undefined);
    expect(traceCorrelationFromActiveContext()).toEqual({});

    vi.mocked(trace.getSpanContext).mockReturnValue({
      traceId: "invalid",
      spanId: "invalid",
      traceFlags: 0,
    });
    expect(traceCorrelationFromActiveContext()).toEqual({});
  });
});
