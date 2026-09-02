import type { RequestOutcome } from "./records.ts";

export const REQUEST_REASONS = [
  "authentication_required",
  "permission_denied",
  "rate_limited",
  "client_error",
  "server_error",
] as const;

export type RequestReason = (typeof REQUEST_REASONS)[number];

export interface RequestResult {
  readonly status_code: number;
  readonly duration_ms: number;
  readonly outcome: RequestOutcome;
  readonly error_code?: string;
  readonly reason?: RequestReason;
}

export function classifyHTTPResult(
  statusCode: number,
  durationMs: number,
): RequestResult {
  if (
    !Number.isInteger(statusCode) ||
    statusCode < 100 ||
    statusCode > 599 ||
    !Number.isInteger(durationMs) ||
    durationMs < 0
  ) {
    throw new TypeError(
      "HTTP result requires a valid status_code and non-negative duration_ms",
    );
  }
  if (statusCode < 400) {
    return {
      status_code: statusCode,
      duration_ms: durationMs,
      outcome: "succeeded",
    };
  }
  if (statusCode === 401) {
    return {
      status_code: statusCode,
      duration_ms: durationMs,
      outcome: "blocked",
      reason: "authentication_required",
    };
  }
  if (statusCode === 403) {
    return {
      status_code: statusCode,
      duration_ms: durationMs,
      outcome: "blocked",
      reason: "permission_denied",
    };
  }
  if (statusCode === 429) {
    return {
      status_code: statusCode,
      duration_ms: durationMs,
      outcome: "blocked",
      reason: "rate_limited",
    };
  }
  return {
    status_code: statusCode,
    duration_ms: durationMs,
    outcome: "failed",
    reason: statusCode < 500 ? "client_error" : "server_error",
  };
}
