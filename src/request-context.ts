import { AsyncLocalStorage } from "node:async_hooks";
import { isIP } from "node:net";

import type { Actor } from "./actor.ts";
import {
  createRequestId,
  isRequestId,
  MESSAGE_REQUEST_ID_HEADER,
  REQUEST_ID_HEADER,
} from "./request-id.ts";

export { isRequestId, MESSAGE_REQUEST_ID_HEADER, REQUEST_ID_HEADER };

export interface RequestContext {
  readonly requestId: string;
  readonly actor: Actor;
  readonly requestedAt: Date;
  readonly sourceIp?: string;
}

const requestScope = new AsyncLocalStorage<RequestContext>();
export function isCanonicalSourceIp(value: string): boolean {
  const version = isIP(value);
  if (version === 4) {
    return true;
  }
  if (version !== 6 || value !== value.toLowerCase()) {
    return false;
  }
  const hostname = new URL(`http://[${value}]/`).hostname;
  return hostname.slice(1, -1) === value;
}

export function createPublicRequestContext(sourceIp?: string): RequestContext {
  if (sourceIp !== undefined && !isCanonicalSourceIp(sourceIp)) {
    throw new TypeError("sourceIp must be a canonical IPv4 or IPv6 address");
  }
  const requestContext: RequestContext = {
    requestId: createRequestId(),
    actor: { kind: "anonymous" },
    requestedAt: new Date(),
    ...(sourceIp === undefined ? {} : { sourceIp }),
  };
  return Object.freeze(requestContext);
}

export function createPropagatedRequestContext(
  requestId: string,
  actor: Actor,
): RequestContext {
  if (!isRequestId(requestId)) {
    throw new TypeError("requestId must be a canonical UUIDv4");
  }
  return Object.freeze({ requestId, actor, requestedAt: new Date() });
}

export function withActor(
  requestContext: RequestContext,
  actor: Actor,
): RequestContext {
  return Object.freeze({ ...requestContext, actor });
}

export function runWithRequestContext<T>(
  requestContext: RequestContext,
  callback: () => T,
): T {
  return requestScope.run(requestContext, callback);
}

export function activeRequestContext(): RequestContext | undefined {
  return requestScope.getStore();
}
