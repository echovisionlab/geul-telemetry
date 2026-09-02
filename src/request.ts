import type { RecordActor } from "./actor.ts";
import {
  validateRequestRecord,
  type Correlation,
  type RequestRecord,
} from "./records.ts";
import type { RequestResult } from "./request-result.ts";

export {
  classifyHTTPResult,
  REQUEST_REASONS,
  type RequestReason,
  type RequestResult,
} from "./request-result.ts";

export type RequestMetadata = Correlation &
  RecordActor & {
    readonly occurred_at: string;
  };

export function buildHTTPRequestRecord(
  metadata: RequestMetadata,
  httpMethod: string,
  httpRoute: string,
  result: RequestResult,
): RequestRecord {
  return buildRequestRecord(metadata, result, {
    http_method: httpMethod,
    http_route: httpRoute,
  });
}

export function buildRPCRequestRecord(
  metadata: RequestMetadata,
  httpMethod: string,
  rpcService: string,
  rpcMethod: string,
  result: RequestResult,
): RequestRecord {
  return buildRequestRecord(metadata, result, {
    http_method: httpMethod,
    rpc_service: rpcService,
    rpc_method: rpcMethod,
  });
}

function buildRequestRecord(
  metadata: RequestMetadata,
  result: RequestResult,
  boundary: Partial<RequestRecord>,
): RequestRecord {
  const record = {
    ...metadata,
    ...boundary,
    ...result,
    event: "request.completed",
  } as RequestRecord;
  validateRequestRecord(record);
  return record;
}
