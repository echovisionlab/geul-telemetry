import { context, isSpanContextValid, trace } from "@opentelemetry/api";

import type { Correlation } from "./records.ts";

export type TraceCorrelation = Pick<Correlation, "trace_id" | "span_id">;

// Browser-safe correlation primitive. Request IDs remain server RequestContext
// state; browser-imported code can read only the active OpenTelemetry span.
export function traceCorrelationFromActiveContext(): TraceCorrelation {
  const spanContext = trace.getSpanContext(context.active());
  if (spanContext === undefined || !isSpanContextValid(spanContext)) return {};
  return { trace_id: spanContext.traceId, span_id: spanContext.spanId };
}
