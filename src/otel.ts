import {
  context,
  propagation,
  type Context,
  type TextMapGetter,
  type TextMapSetter,
} from "@opentelemetry/api";

import type { Actor } from "./actor.ts";
import {
  activeRequestContext,
  createPropagatedRequestContext,
  MESSAGE_REQUEST_ID_HEADER,
  type RequestContext,
} from "./request-context.ts";
import type { Correlation } from "./records.ts";
import { traceCorrelationFromActiveContext } from "./trace.ts";

export interface ExtractedCorrelation {
  readonly otelContext: Context;
  readonly requestContext?: RequestContext;
}

export function correlationFromActiveContext(): Correlation {
  const requestContext = activeRequestContext();
  return {
    ...(requestContext === undefined
      ? {}
      : { request_id: requestContext.requestId }),
    ...traceCorrelationFromActiveContext(),
  };
}

export function injectCorrelation<T>(
  carrier: T,
  setter: TextMapSetter<T>,
  requestContext = activeRequestContext(),
): void {
  propagation.inject(context.active(), carrier, setter);
  if (requestContext !== undefined) {
    setter.set(carrier, MESSAGE_REQUEST_ID_HEADER, requestContext.requestId);
  }
}

export function extractCorrelation<T>(
  carrier: T,
  getter: TextMapGetter<T>,
  actor: Actor,
): ExtractedCorrelation {
  const otelContext = propagation.extract(context.active(), carrier, getter);
  const requestId = getter.get(carrier, MESSAGE_REQUEST_ID_HEADER);
  const value = Array.isArray(requestId) ? requestId[0] : requestId;
  if (value === undefined) {
    return { otelContext };
  }
  try {
    return {
      otelContext,
      requestContext: createPropagatedRequestContext(value, actor),
    };
  } catch {
    return { otelContext };
  }
}
